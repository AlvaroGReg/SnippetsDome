package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"SnippetsDome/internal/cryptography"
	"SnippetsDome/internal/domain"
)

var (
	ErrImportJSONEmpty           = errors.New("JSON import must contain an array")
	ErrImportJSONField           = errors.New("JSON snippet is missing a required field")
	ErrImportJSONDuplicateID     = errors.New("JSON import contains duplicate snippet IDs")
	ErrImportJSONIDExists        = errors.New("a snippet ID already exists")
	ErrEncryptedCollectionExport = errors.New("encrypted collections cannot be exported")
	ErrEncryptedJSONPassword     = errors.New("encrypted JSON password is incorrect or missing")
	ErrInvalidEncryptedJSON      = errors.New("invalid encrypted JSON file")
)

const encryptedJSONFormat = "SnippetsDome-encrypted-json-v1"

type encryptedJSON struct {
	Format     string                  `json:"format"`
	Salt       string                  `json:"salt"`
	Parameters cryptography.Parameters `json:"parameters"`
	Data       string                  `json:"data"`
}

type jsonExchangeRepository interface {
	ImportCollection(domain.Collection, []domain.Snippet) error
	ListSnippets() ([]domain.Snippet, error)
	ListCollectionSnippets(string) ([]domain.Snippet, error)
	ActiveCollection() (domain.Collection, error)
}

// ImportJSON reads the legacy array format and creates one SQLite collection.
func (s *SnippetService) ImportJSON(path, collectionName string) (domain.Collection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Collection{}, fmt.Errorf("read JSON file: %w", err)
	}
	var snippets []domain.Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return domain.Collection{}, fmt.Errorf("decode JSON file: %w", err)
	}
	return s.importSnippets(path, collectionName, snippets)
}

// ImportJSONWithPassword imports either a legacy JSON array or an encrypted JSON exchange file.
func (s *SnippetService) ImportJSONWithPassword(path, collectionName, password string) (domain.Collection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Collection{}, fmt.Errorf("read JSON file: %w", err)
	}
	var envelope encryptedJSON
	if strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		if err := json.Unmarshal(data, &envelope); err != nil {
			return domain.Collection{}, fmt.Errorf("decode JSON file: %w", err)
		}
	}
	if envelope.Format == encryptedJSONFormat {
		plain, err := decryptJSONEnvelope(envelope, password)
		if err != nil {
			return domain.Collection{}, err
		}
		data = plain
	}
	var snippets []domain.Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return domain.Collection{}, fmt.Errorf("decode JSON file: %w", err)
	}
	return s.importSnippets(path, collectionName, snippets)
}

// IsEncryptedJSONFile reports whether a JSON exchange file uses the encrypted format.
func IsEncryptedJSONFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read JSON file: %w", err)
	}
	var envelope encryptedJSON
	if err := json.Unmarshal(data, &envelope); err != nil {
		// Legacy exports are JSON arrays rather than exchange objects.
		if strings.HasPrefix(strings.TrimSpace(string(data)), "[") {
			return false, nil
		}
		return false, fmt.Errorf("decode JSON file: %w", err)
	}
	return envelope.Format == encryptedJSONFormat, nil
}

func (s *SnippetService) importSnippets(path, collectionName string, snippets []domain.Snippet) (domain.Collection, error) {
	if snippets == nil {
		return domain.Collection{}, ErrImportJSONEmpty
	}
	if strings.TrimSpace(collectionName) == "" {
		collectionName = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	collectionName = strings.TrimSpace(collectionName)
	if collectionName == "" {
		return domain.Collection{}, errors.New("import collection name is required")
	}
	seen := make(map[string]struct{}, len(snippets))
	for _, snippet := range snippets {
		if snippet.ID == "" || strings.TrimSpace(snippet.Title) == "" || strings.TrimSpace(snippet.Code) == "" || strings.TrimSpace(snippet.CreatedAt) == "" {
			return domain.Collection{}, ErrImportJSONField
		}
		if _, exists := seen[snippet.ID]; exists {
			return domain.Collection{}, ErrImportJSONDuplicateID
		}
		seen[snippet.ID] = struct{}{}
	}
	// Snippet IDs are globally unique in the database, so each import is a new copy.
	importedSnippets := make([]domain.Snippet, len(snippets))
	copy(importedSnippets, snippets)
	for i := range importedSnippets {
		importedSnippets[i].ID = uuid.NewString()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(jsonExchangeRepository)
	if !ok {
		return domain.Collection{}, errors.New("JSON exchange is not supported")
	}
	collection := domain.Collection{Name: collectionName}
	s.clearKey()
	if err := repository.ImportCollection(collection, importedSnippets); err != nil {
		if errors.Is(err, domain.ErrSnippetIDExists) {
			return domain.Collection{}, ErrImportJSONIDExists
		}
		return domain.Collection{}, err
	}
	return repository.ActiveCollection()
}

// ExportJSON writes only the active collection using the stable legacy array format.
func (s *SnippetService) ExportJSON(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.requireExportableActiveCollection(); err != nil {
		return err
	}
	repository, ok := s.repository.(jsonExchangeRepository)
	if !ok {
		return errors.New("JSON exchange is not supported")
	}
	snippets, err := repository.ListSnippets()
	if err != nil {
		return err
	}
	return writeSnippetsJSON(path, snippets)
}

func (s *SnippetService) ExportCollectionJSON(path, collectionID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	collections, err := repository.ListCollections()
	if err != nil {
		return err
	}
	found := false
	for _, collection := range collections {
		if collection.ID == collectionID {
			found = true
			if collection.Encrypted {
				return ErrEncryptedCollectionExport
			}
			break
		}
	}
	if !found {
		return errors.New("collection not found")
	}
	exchangeRepository, ok := s.repository.(jsonExchangeRepository)
	if !ok {
		return errors.New("JSON exchange is not supported")
	}
	snippets, err := exchangeRepository.ListCollectionSnippets(collectionID)
	if err != nil {
		return err
	}
	return writeSnippetsJSON(path, snippets)
}

// ExportEncryptedCollectionJSON verifies the list password and encrypts an export using password.
func (s *SnippetService) ExportEncryptedCollectionJSON(path, collectionID, collectionPassword, password string) error {
	if collectionPassword == "" || password == "" {
		return ErrEncryptedJSONPassword
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	collections, err := repository.ListCollections()
	if err != nil {
		return err
	}
	found, encrypted := false, false
	for _, collection := range collections {
		if collection.ID == collectionID {
			found, encrypted = true, collection.Encrypted
			break
		}
	}
	if !found {
		return errors.New("collection not found")
	}
	if !encrypted {
		return errors.New("collection is not encrypted")
	}
	salt, encodedParameters, check, err := repository.CollectionEncryptionMetadata(collectionID)
	if err != nil {
		return err
	}
	parameters, err := cryptography.UnmarshalParameters(encodedParameters)
	if err != nil {
		return ErrCollectionAuthentication
	}
	collectionKey, err := cryptography.DeriveKey(collectionPassword, salt, parameters)
	if err != nil {
		return ErrCollectionAuthentication
	}
	defer cryptography.Clear(collectionKey)
	verified, err := cryptography.Decrypt(collectionKey, check, cryptography.CollectionAAD(collectionID, collectionID, "collection.key-check"))
	if err != nil || verified != "collection-key-check-v1" {
		return ErrCollectionAuthentication
	}
	exchangeRepository, ok := s.repository.(jsonExchangeRepository)
	if !ok {
		return errors.New("JSON exchange is not supported")
	}
	snippets, err := exchangeRepository.ListCollectionSnippets(collectionID)
	if err != nil {
		return err
	}
	for i := range snippets {
		if err := transformSnippet(collectionID, collectionKey, &snippets[i], false); err != nil {
			return ErrCollectionAuthentication
		}
	}
	plain, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON file: %w", err)
	}
	salt, err = cryptography.NewSalt()
	if err != nil {
		return err
	}
	parameters = cryptography.DefaultParameters()
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return err
	}
	defer cryptography.Clear(key)
	sealed, err := cryptography.Encrypt(key, string(plain), []byte(encryptedJSONFormat))
	if err != nil {
		return err
	}
	envelope := encryptedJSON{Format: encryptedJSONFormat, Salt: base64.RawStdEncoding.EncodeToString(salt), Parameters: parameters, Data: sealed}
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("encode encrypted JSON file: %w", err)
	}
	return writeJSONFile(path, append(data, '\n'))
}

func decryptJSONEnvelope(envelope encryptedJSON, password string) ([]byte, error) {
	if password == "" || envelope.Format != encryptedJSONFormat || envelope.Data == "" {
		return nil, ErrEncryptedJSONPassword
	}
	salt, err := base64.RawStdEncoding.DecodeString(envelope.Salt)
	if err != nil || len(salt) != cryptography.SaltSize || envelope.Parameters.Validate() != nil {
		return nil, ErrInvalidEncryptedJSON
	}
	key, err := cryptography.DeriveKey(password, salt, envelope.Parameters)
	if err != nil {
		return nil, ErrInvalidEncryptedJSON
	}
	defer cryptography.Clear(key)
	plain, err := cryptography.Decrypt(key, envelope.Data, []byte(encryptedJSONFormat))
	if err != nil {
		return nil, ErrEncryptedJSONPassword
	}
	return []byte(plain), nil
}

func (s *SnippetService) requireExportableActiveCollection() error {
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return nil
	}
	collection, err := repository.ActiveCollection()
	if err != nil {
		return err
	}
	if collection.Encrypted {
		return ErrEncryptedCollectionExport
	}
	return nil
}

func writeSnippetsJSON(path string, snippets []domain.Snippet) error {
	data, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON file: %w", err)
	}
	data = append(data, '\n')
	return writeJSONFile(path, data)
}

func writeJSONFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write JSON file: %w", err)
	}
	return nil
}
