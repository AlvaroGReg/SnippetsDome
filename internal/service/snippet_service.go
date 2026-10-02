package service

import (
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"SnippetsDome/internal/domain"
)

type SnippetService struct {
	mu               sync.RWMutex
	repository       snippetRepository
	config           domain.AppConfig
	configRepository configRepository
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
	ListCollections() ([]domain.Collection, error)
	CreateCollection(string) (domain.Collection, error)
	RenameCollection(string, string) error
	DeleteCollection(string) error
	SelectCollection(string) error
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, err := s.repository.ListSnippets()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Favorite && !result[j].Favorite
	})
	log.Printf("snippet_service::List() => result:: %+v", result)

	return result, nil
}

// CREATE
func (s *SnippetService) CreateSnippet(input domain.CreateSnippetInput) (domain.Snippet, error) {
	log.Printf("snippet_service::CreateSnippet(input)::input %+v", input)

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

	log.Printf("snippet_service::CreateSnippet(input) => res::res %+v", snippet)

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.repository.CreateSnippet(snippet)
}

// DELETE
func (s *SnippetService) DeleteSnippet(id string) error {
	log.Printf("snippet_service::DeleteSnippet(id)::id %+v", id)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repository.DeleteSnippet(id)
}

// UPDATE
func (s *SnippetService) UpdateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
	log.Printf("snippet_service::UpdateSnippet(snippet)::newSnippet %+v", snippet)
	// normalize and validate snippet item
	snippet.Title = strings.TrimSpace(snippet.Title)
	snippet.Language = strings.TrimSpace(snippet.Language)
	snippet.Code = strings.TrimSpace(snippet.Code)

	if err := CheckValidSnippet(snippet); err != nil {
		return domain.Snippet{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repository.UpdateSnippet(snippet)
}

func (s *SnippetService) ActiveCollection() (domain.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	return repository.ActiveCollection()
}

func (s *SnippetService) ListCollections() ([]domain.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return nil, errors.New("collection operations are not supported")
	}
	return repository.ListCollections()
}

func (s *SnippetService) CreateCollection(name string) (domain.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	return repository.RenameCollection(id, name)
}

func (s *SnippetService) DeleteCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.DeleteCollection(id)
}

func (s *SnippetService) SelectCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.SelectCollection(id)
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
