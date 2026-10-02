package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"SnippetsDome/internal/domain"
)

var (
	ErrImportJSONEmpty       = errors.New("JSON import must contain an array")
	ErrImportJSONField       = errors.New("JSON snippet is missing a required field")
	ErrImportJSONDuplicateID = errors.New("JSON import contains duplicate snippet IDs")
	ErrImportJSONIDExists    = errors.New("a snippet ID already exists")
)

type jsonExchangeRepository interface {
	ImportCollection(domain.Collection, []domain.Snippet) error
	ListSnippets() ([]domain.Snippet, error)
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

	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(jsonExchangeRepository)
	if !ok {
		return domain.Collection{}, errors.New("JSON exchange is not supported")
	}
	collection := domain.Collection{Name: collectionName}
	if err := repository.ImportCollection(collection, snippets); err != nil {
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
	repository, ok := s.repository.(jsonExchangeRepository)
	if !ok {
		return errors.New("JSON exchange is not supported")
	}
	snippets, err := repository.ListSnippets()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON file: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write JSON file: %w", err)
	}
	return nil
}
