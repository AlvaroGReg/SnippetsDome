package service

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/repository"
)

type SnippetService struct {
	mu               sync.RWMutex
	repository       *repository.JSONSnippetRepository
	config           domain.AppConfig
	configRepository configRepository
}

type configRepository interface {
	SaveConfig(domain.AppConfig) error
}

func NewSnippetService(config domain.AppConfig, configRepository configRepository) *SnippetService {
	if config.TraySnippetLimit < 1 {
		config.TraySnippetLimit = domain.DefaultTraySnippetLimit
	}

	snippetRepository := repository.NewJSONSnippetRepository()
	snippetRepository.SetFilePath(config.SnippetsFilePath)

	return &SnippetService{
		repository:       snippetRepository,
		config:           config,
		configRepository: configRepository,
	}
}

func (s *SnippetService) SnippetsFilePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repository.FilePath()
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

func (s *SnippetService) EnsureSnippetsFile() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repository.EnsureFile()
}

// SetSnippetsFile changes the snippets file and persists the selection. If
// either operation fails, it restores the previous path.
func (s *SnippetService) SetSnippetsFile(filePath string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if filePath == "" {
		return "", errors.New("no snippets file selected")
	}

	previousFilePath := s.repository.FilePath()
	s.repository.SetFilePath(filePath)
	if err := s.repository.EnsureFile(); err != nil {
		s.repository.SetFilePath(previousFilePath)
		return "", err
	}

	previousConfig := s.config
	s.config.SnippetsFilePath = filePath
	if s.configRepository != nil {
		if err := s.configRepository.SaveConfig(s.config); err != nil {
			s.config = previousConfig
			s.repository.SetFilePath(previousFilePath)
			return "", err
		}
	}

	return filePath, nil
}

// GET
func (s *SnippetService) List() ([]domain.Snippet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, err := s.repository.List()
	if err != nil {
		return nil, err
	}
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

	snippets, err := s.repository.List()
	if err != nil {
		return domain.Snippet{}, err
	}
	snippets = append(snippets, snippet)
	s.repository.SaveList(snippets)

	return snippet, nil
}

// DELETE
func (s *SnippetService) DeleteSnippet(id string) error {
	log.Printf("snippet_service::DeleteSnippet(id)::id %+v", id)
	snippets, err := s.List()
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i, value := range snippets {
		if value.ID == id {
			log.Printf("snippet_service::DeleteSnippet(id)::snippetToDelete %+v", value)
			snippets = append(snippets[:i], snippets[i+1:]...)
			s.repository.SaveList(snippets)
			return nil
		}
	}
	return errors.New("unable to delete the snippet")
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
	// get and check list
	snippets, err := s.List()
	if err != nil {
		return domain.Snippet{}, err
	}
	// rewrite item, then save
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, value := range snippets {
		if value.ID == snippet.ID {
			snippet.CreatedAt = value.CreatedAt
			snippets[i] = snippet
			s.repository.SaveList(snippets)
			return snippet, nil
		}
	}

	return domain.Snippet{}, errors.New("snippet not found")
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
