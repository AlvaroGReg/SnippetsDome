package service

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"SnippetsDome/internal/domain"
)

// List returns snippets in the active collection, with favorites first.
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

// CreateSnippet validates and persists a new snippet.
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

// DeleteSnippet removes a snippet from the active collection.
func (s *SnippetService) DeleteSnippet(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireUnlockedActiveCollection(); err != nil {
		return err
	}
	return s.repository.DeleteSnippet(id)
}

// UpdateSnippet validates and persists changes to a snippet.
func (s *SnippetService) UpdateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
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

// CheckValidSnippet validates the required snippet fields.
func CheckValidSnippet(snippet domain.Snippet) error {
	if snippet.Title == "" {
		return errors.New("title is required")
	}
	if snippet.Code == "" {
		return errors.New("code is required")
	}
	return nil
}
