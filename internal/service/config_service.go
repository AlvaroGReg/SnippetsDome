package service

import (
	"errors"

	"SnippetsDome/internal/domain"
)

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
