package service

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"SnippetsDome/internal/cryptography"
	"SnippetsDome/internal/domain"
)

var (
	ErrInvalidCollectionPassword = errors.New("collection password must contain between 1 and 10 characters")
	ErrCollectionLocked          = errors.New("collection is locked")
	ErrCollectionAuthentication  = errors.New("authentication failed")
)

type SnippetService struct {
	mu                   sync.RWMutex
	repository           snippetRepository
	config               domain.AppConfig
	configRepository     configRepository
	activeKey            []byte
	unlockedCollectionID string
}

type configRepository interface {
	SaveConfig(domain.AppConfig) error
}

type snippetRepository interface {
	ListSnippets() ([]domain.Snippet, error)
	CreateSnippet(domain.Snippet) (domain.Snippet, error)
	UpdateSnippet(domain.Snippet) (domain.Snippet, error)
	DeleteSnippet(string) error
}

type collectionRepository interface {
	ActiveCollection() (domain.Collection, error)
	DefaultCollection() (domain.Collection, error)
	ListCollections() ([]domain.Collection, error)
	CreateCollection(string) (domain.Collection, error)
	RenameCollection(string, string) error
	DeleteCollection(string) error
	SelectCollection(string) error
	SetDefaultCollection(string) error
	RenameEncryptedCollection(string, string, string) error
	EncryptCollection(string, []byte, []byte, cryptography.Parameters, string) error
	DecryptCollection(string, []byte) error
	CollectionEncryptionMetadata(string) ([]byte, string, string, error)
}

func validCollectionPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return utf8.ValidString(password) && length >= 1 && length <= 10
}

func (s *SnippetService) activeCollection() (domain.Collection, error) {
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	return repository.ActiveCollection()
}

func (s *SnippetService) requireUnlockedActiveCollection() error {
	collection, err := s.activeCollection()
	if err != nil {
		return err
	}
	if collection.Encrypted && (len(s.activeKey) == 0 || s.unlockedCollectionID != collection.ID) {
		return ErrCollectionLocked
	}
	return nil
}

func (s *SnippetService) EncryptCollection(id, password, confirmation string) error {
	if !validCollectionPassword(password) || !validCollectionPassword(confirmation) || password != confirmation {
		return ErrInvalidCollectionPassword
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	salt, err := cryptography.NewSalt()
	if err != nil {
		return err
	}
	parameters := cryptography.DefaultParameters()
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return err
	}
	defer cryptography.Clear(key)
	check, err := cryptography.Encrypt(key, "collection-key-check-v1", cryptography.CollectionAAD(id, id, "collection.key-check"))
	if err != nil {
		return err
	}
	if err := repository.EncryptCollection(id, key, salt, parameters, check); err != nil {
		return err
	}
	collection, err := repository.ActiveCollection()
	if err == nil && collection.ID == id {
		s.clearKey()
		s.activeKey = append([]byte(nil), key...)
		s.unlockedCollectionID = id
	}
	return nil
}

func (s *SnippetService) UnlockCollection(id, password string) (domain.Collection, error) {
	if !validCollectionPassword(password) {
		return domain.Collection{}, ErrInvalidCollectionPassword
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	s.clearKey()
	if err := repository.SelectCollection(id); err != nil {
		return domain.Collection{}, err
	}
	salt, encodedParameters, check, err := repository.CollectionEncryptionMetadata(id)
	if err != nil {
		return domain.Collection{}, err
	}
	parameters, err := cryptography.UnmarshalParameters(encodedParameters)
	if err != nil {
		return domain.Collection{}, ErrCollectionAuthentication
	}
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return domain.Collection{}, ErrCollectionAuthentication
	}
	verified, err := cryptography.Decrypt(key, check, cryptography.CollectionAAD(id, id, "collection.key-check"))
	if err != nil || verified != "collection-key-check-v1" {
		cryptography.Clear(key)
		return domain.Collection{}, ErrCollectionAuthentication
	}
	s.activeKey = key
	s.unlockedCollectionID = id
	collection, err := repository.ActiveCollection()
	if err != nil {
		s.clearKey()
		return domain.Collection{}, err
	}
	collection.Name, err = cryptography.Decrypt(key, collection.Name, cryptography.CollectionAAD(id, id, "collection.name"))
	if err != nil {
		s.clearKey()
		return domain.Collection{}, ErrCollectionAuthentication
	}
	return collection, nil
}

func (s *SnippetService) clearKey() {
	cryptography.Clear(s.activeKey)
	s.activeKey = nil
	s.unlockedCollectionID = ""
}

func (s *SnippetService) LockActiveCollection() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearKey()
	return nil
}

func (s *SnippetService) DecryptCollection(id, password string) error {
	if !validCollectionPassword(password) {
		return ErrInvalidCollectionPassword
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	salt, encodedParameters, check, err := repository.CollectionEncryptionMetadata(id)
	if err != nil {
		return err
	}
	parameters, err := cryptography.UnmarshalParameters(encodedParameters)
	if err != nil {
		return ErrCollectionAuthentication
	}
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return ErrCollectionAuthentication
	}
	defer cryptography.Clear(key)
	verified, err := cryptography.Decrypt(key, check, cryptography.CollectionAAD(id, id, "collection.key-check"))
	if err != nil || verified != "collection-key-check-v1" {
		return ErrCollectionAuthentication
	}
	if err := repository.DecryptCollection(id, key); err != nil {
		return err
	}
	if s.unlockedCollectionID == id {
		s.clearKey()
	}
	return nil
}

func NewSnippetService(config domain.AppConfig, snippetRepository snippetRepository, configRepository configRepository) *SnippetService {
	if config.TraySnippetLimit < 1 {
		config.TraySnippetLimit = domain.DefaultTraySnippetLimit
	}
	if config.Language == "system" {
		// Migrate the preference used by the previous system-language option.
		config.Language = ""
	}
	if !validLanguage(config.Language) && config.Language != "" {
		config.Language = domain.LanguageEnglish
	}

	return &SnippetService{
		repository:       snippetRepository,
		config:           config,
		configRepository: configRepository,
	}
}

func validLanguage(language string) bool {
	return language == domain.LanguageEnglish || language == domain.LanguageSpanish
}

func (s *SnippetService) Language() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Language
}

func (s *SnippetService) SetLanguage(language string) error {
	if !validLanguage(language) {
		return errors.New("unsupported language")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	previousConfig := s.config
	s.config.Language = language
	if s.configRepository == nil {
		return nil
	}
	if err := s.configRepository.SaveConfig(s.config); err != nil {
		s.config = previousConfig
		return err
	}
	return nil
}

// CloseToTrayEnabled reports whether closing the main window should keep the
// application available from the system tray.
func (s *SnippetService) CloseToTrayEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.CloseToTray
}

// SetCloseToTrayEnabled persists the close-to-tray preference. The previous
// in-memory value is retained when persistence fails.
func (s *SnippetService) SetCloseToTrayEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	previousConfig := s.config
	s.config.CloseToTray = enabled
	if s.configRepository == nil {
		return nil
	}
	if err := s.configRepository.SaveConfig(s.config); err != nil {
		s.config = previousConfig
		return err
	}
	return nil
}

// StartAtLoginEnabled reports whether the application should register itself
// to launch when the user signs in.
func (s *SnippetService) StartAtLoginEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.StartAtLogin
}

// SetStartAtLoginEnabled persists the start-at-login preference. Registering
// the platform-specific launcher is handled by the application layer.
func (s *SnippetService) SetStartAtLoginEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	previousConfig := s.config
	s.config.StartAtLogin = enabled
	if s.configRepository == nil {
		return nil
	}
	if err := s.configRepository.SaveConfig(s.config); err != nil {
		s.config = previousConfig
		return err
	}
	return nil
}

// TraySnippetLimit reports the maximum number of snippets shown in the tray menu.
func (s *SnippetService) TraySnippetLimit() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.TraySnippetLimit
}

// SetTraySnippetLimit persists the maximum number of snippets shown in the tray menu.
func (s *SnippetService) SetTraySnippetLimit(limit int) error {
	if limit < 1 {
		return errors.New("tray snippet limit must be greater than zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	previousConfig := s.config
	s.config.TraySnippetLimit = limit
	if s.configRepository == nil {
		return nil
	}
	if err := s.configRepository.SaveConfig(s.config); err != nil {
		s.config = previousConfig
		return err
	}
	return nil
}

// GET
func (s *SnippetService) List() ([]domain.Snippet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireUnlockedActiveCollection(); err != nil {
		return nil, err
	}

	result, err := s.repository.ListSnippets()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Favorite && !result[j].Favorite
	})
	if len(s.activeKey) > 0 {
		collection, err := s.activeCollection()
		if err != nil {
			return nil, err
		}
		for i := range result {
			if err := transformSnippet(collection.ID, s.activeKey, &result[i], false); err != nil {
				return nil, ErrCollectionAuthentication
			}
		}
	}

	return result, nil
}

// CREATE
func (s *SnippetService) CreateSnippet(input domain.CreateSnippetInput) (domain.Snippet, error) {

	snippet := domain.Snippet{
		ID:        uuid.NewString(),
		Title:     strings.TrimSpace(input.Title),
		Language:  strings.TrimSpace(input.Language),
		Code:      strings.TrimSpace(input.Code),
		Tags:      input.Tags,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := CheckValidSnippet(snippet); err != nil {
		return domain.Snippet{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireUnlockedActiveCollection(); err != nil {
		return domain.Snippet{}, err
	}
	return s.saveSnippet(snippet, false)
}

// DELETE
func (s *SnippetService) DeleteSnippet(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireUnlockedActiveCollection(); err != nil {
		return err
	}
	return s.repository.DeleteSnippet(id)
}

// UPDATE
func (s *SnippetService) UpdateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
	// normalize and validate snippet item
	snippet.Title = strings.TrimSpace(snippet.Title)
	snippet.Language = strings.TrimSpace(snippet.Language)
	snippet.Code = strings.TrimSpace(snippet.Code)

	if err := CheckValidSnippet(snippet); err != nil {
		return domain.Snippet{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireUnlockedActiveCollection(); err != nil {
		return domain.Snippet{}, err
	}
	return s.saveSnippet(snippet, true)
}

func (s *SnippetService) saveSnippet(snippet domain.Snippet, update bool) (domain.Snippet, error) {
	collection, err := s.activeCollection()
	if err != nil {
		return domain.Snippet{}, err
	}
	snippet.Tags = normalizeServiceTags(snippet.Tags)
	if len(s.activeKey) > 0 {
		if err := transformSnippet(collection.ID, s.activeKey, &snippet, true); err != nil {
			return domain.Snippet{}, err
		}
	}
	var saved domain.Snippet
	if update {
		saved, err = s.repository.UpdateSnippet(snippet)
	} else {
		saved, err = s.repository.CreateSnippet(snippet)
	}
	if err != nil {
		return domain.Snippet{}, err
	}
	if len(s.activeKey) > 0 {
		if err := transformSnippet(collection.ID, s.activeKey, &saved, false); err != nil {
			return domain.Snippet{}, ErrCollectionAuthentication
		}
	}
	return saved, nil
}

func normalizeServiceTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		key := strings.ToLower(tag)
		if tag != "" {
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				result = append(result, tag)
			}
		}
	}
	return result
}

func transformSnippet(collectionID string, key []byte, snippet *domain.Snippet, encrypt bool) error {
	fields := []struct {
		value *string
		name  string
	}{{&snippet.Title, "snippet.title"}, {&snippet.Language, "snippet.language"}, {&snippet.Code, "snippet.code"}}
	for _, field := range fields {
		var result string
		var err error
		aad := cryptography.CollectionAAD(collectionID, snippet.ID, field.name)
		if encrypt {
			result, err = cryptography.Encrypt(key, *field.value, aad)
		} else {
			result, err = cryptography.Decrypt(key, *field.value, aad)
		}
		if err != nil {
			return err
		}
		*field.value = result
	}
	for i, tag := range snippet.Tags {
		var err error
		aad := cryptography.CollectionAAD(collectionID, snippet.ID, "snippet.tag")
		if encrypt {
			snippet.Tags[i], err = cryptography.Encrypt(key, tag, aad)
		} else {
			snippet.Tags[i], err = cryptography.Decrypt(key, tag, aad)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SnippetService) ActiveCollection() (domain.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	collection, err := repository.ActiveCollection()
	if err == nil && collection.Encrypted {
		collection.Locked = s.unlockedCollectionID != collection.ID
		if collection.Locked {
			collection.Name = lockedCollectionName(collection.ID)
		}
	}
	return collection, err
}

func lockedCollectionName(id string) string {
	if len(id) > 8 {
		id = id[:8]
	}
	return "Locked collection (" + id + ")"
}

func (s *SnippetService) DefaultCollection() (domain.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	collection, err := repository.DefaultCollection()
	if err == nil && collection.Encrypted {
		collection.Locked = true
		collection.Name = lockedCollectionName(collection.ID)
	}
	return collection, err
}

func (s *SnippetService) ListCollections() ([]domain.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return nil, errors.New("collection operations are not supported")
	}
	collections, err := repository.ListCollections()
	if err == nil {
		for i := range collections {
			if collections[i].Encrypted {
				collections[i].Locked = s.unlockedCollectionID != collections[i].ID
				if collections[i].Locked {
					collections[i].Name = lockedCollectionName(collections[i].ID)
				} else {
					collections[i].Name, err = cryptography.Decrypt(s.activeKey, collections[i].Name, cryptography.CollectionAAD(collections[i].ID, collections[i].ID, "collection.name"))
					if err != nil {
						return nil, ErrCollectionAuthentication
					}
				}
			}
		}
	}
	return collections, err
}

func (s *SnippetService) CreateCollection(name string) (domain.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearKey()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	return repository.CreateCollection(name)
}

func (s *SnippetService) RenameCollection(id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("collection name is required")
	}
	collections, err := repository.ListCollections()
	if err != nil {
		return err
	}
	var encrypted bool
	found := false
	for _, collection := range collections {
		if collection.ID == id {
			encrypted = collection.Encrypted
			found = true
			break
		}
	}
	if !found {
		return errors.New("collection not found")
	}
	if encrypted {
		if len(s.activeKey) == 0 || s.unlockedCollectionID != id {
			return ErrCollectionLocked
		}
		sealedName, err := cryptography.Encrypt(s.activeKey, name, cryptography.CollectionAAD(id, id, "collection.name"))
		if err != nil {
			return err
		}
		return repository.RenameEncryptedCollection(id, sealedName, name)
	}
	return repository.RenameCollection(id, name)
}

func (s *SnippetService) DeleteCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unlockedCollectionID == id {
		s.clearKey()
	}
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.DeleteCollection(id)
}

func (s *SnippetService) SelectCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearKey()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.SelectCollection(id)
}

func (s *SnippetService) SetDefaultCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.SetDefaultCollection(id)
}

// UTILS
func CheckValidSnippet(snippet domain.Snippet) error {
	if snippet.Title == "" {
		return errors.New("title is required")
	}

	if snippet.Code == "" {
		return errors.New("code is required")
	}

	return nil
}
