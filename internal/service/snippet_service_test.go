package service

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/repository"
)

func TestSnippetServiceSetSnippetsFile(t *testing.T) {
	previousFilePath := filepath.Join(t.TempDir(), "previous.json")
	newFilePath := filepath.Join(t.TempDir(), "selected.json")

	t.Run("creates the selected file and persists its path", func(t *testing.T) {
		configStore := &recordingConfigRepository{}
		service := NewSnippetService(domain.AppConfig{SnippetsFilePath: previousFilePath}, configStore)

		selectedFilePath, err := service.SetSnippetsFile(newFilePath)
		if err != nil {
			t.Fatalf("SetSnippetsFile() error = %v", err)
		}
		if selectedFilePath != newFilePath || service.SnippetsFilePath() != newFilePath {
			t.Errorf("selected path = %q, service path = %q, want %q", selectedFilePath, service.SnippetsFilePath(), newFilePath)
		}
		if err := service.EnsureSnippetsFile(); err != nil {
			t.Errorf("selected snippets file was not created: %v", err)
		}
	})

	t.Run("restores the previous path when persisting the selection fails", func(t *testing.T) {
		configStore := &recordingConfigRepository{err: errors.New("save config failed")}
		service := NewSnippetService(domain.AppConfig{SnippetsFilePath: previousFilePath}, configStore)

		if _, err := service.SetSnippetsFile(newFilePath); err == nil {
			t.Fatal("SetSnippetsFile() error = nil, want config save error")
		}
		if got := service.SnippetsFilePath(); got != previousFilePath {
			t.Errorf("service path after rollback = %q, want %q", got, previousFilePath)
		}
	})
}

func TestSnippetServiceSetCloseToTrayEnabled(t *testing.T) {
	t.Run("persists the enabled preference", func(t *testing.T) {
		configStore := &recordingConfigRepository{}
		service := NewSnippetService(domain.AppConfig{}, configStore)

		if err := service.SetCloseToTrayEnabled(true); err != nil {
			t.Fatalf("SetCloseToTrayEnabled() error = %v", err)
		}
		if !service.CloseToTrayEnabled() {
			t.Error("CloseToTrayEnabled() = false, want true")
		}
		if !configStore.config.CloseToTray {
			t.Error("saved config does not enable CloseToTray")
		}
	})

	t.Run("restores the previous preference when saving fails", func(t *testing.T) {
		configStore := &recordingConfigRepository{err: errors.New("save config failed")}
		service := NewSnippetService(domain.AppConfig{CloseToTray: false}, configStore)

		if err := service.SetCloseToTrayEnabled(true); err == nil {
			t.Fatal("SetCloseToTrayEnabled() error = nil, want config save error")
		}
		if service.CloseToTrayEnabled() {
			t.Error("CloseToTrayEnabled() = true after failed save, want false")
		}
	})
}

func TestSnippetServiceSetStartAtLoginEnabled(t *testing.T) {
	t.Run("persists the enabled preference", func(t *testing.T) {
		configStore := &recordingConfigRepository{}
		service := NewSnippetService(domain.AppConfig{}, configStore)

		if err := service.SetStartAtLoginEnabled(true); err != nil {
			t.Fatalf("SetStartAtLoginEnabled() error = %v", err)
		}
		if !service.StartAtLoginEnabled() {
			t.Error("StartAtLoginEnabled() = false, want true")
		}
		if !configStore.config.StartAtLogin {
			t.Error("saved config does not enable StartAtLogin")
		}
	})

	t.Run("restores the previous preference when saving fails", func(t *testing.T) {
		configStore := &recordingConfigRepository{err: errors.New("save config failed")}
		service := NewSnippetService(domain.AppConfig{StartAtLogin: false}, configStore)

		if err := service.SetStartAtLoginEnabled(true); err == nil {
			t.Fatal("SetStartAtLoginEnabled() error = nil, want config save error")
		}
		if service.StartAtLoginEnabled() {
			t.Error("StartAtLoginEnabled() = true after failed save, want false")
		}
	})
}

func TestSnippetServiceSetTraySnippetLimit(t *testing.T) {
	// TODO: describe persistence, validation, and rollback behavior for the tray snippet limit.
	t.Skip("TODO: implement tests for SetTraySnippetLimit")
}

type recordingConfigRepository struct {
	config domain.AppConfig
	err    error
}

func (r *recordingConfigRepository) SaveConfig(config domain.AppConfig) error {
	r.config = config
	return r.err
}

func newSnippetServiceWithSnippets(t *testing.T, snippets []domain.Snippet) *SnippetService {
	t.Helper()

	// Create an isolated JSON repository for each service test.
	filePath := filepath.Join(t.TempDir(), "snippets.json")
	repo := repository.NewJSONSnippetRepository()
	repo.SetFilePath(filePath)
	if err := repo.EnsureFile(); err != nil {
		t.Fatalf("EnsureFile() error = %v", err)
	}
	if err := repo.SaveList(snippets); err != nil {
		t.Fatalf("SaveList() error = %v", err)
	}

	return NewSnippetService(domain.AppConfig{SnippetsFilePath: filePath}, nil)
}

func TestCheckValidSnippet(t *testing.T) {
	tests := []struct {
		name    string
		snippet domain.Snippet
		wantErr bool
	}{
		{
			name: "accepts a snippet with title and code",
			snippet: domain.Snippet{
				Title: "Print greeting",
				Code:  "fmt.Println(\"Hello\")",
			},
		},
		{
			name: "rejects a snippet without title",
			snippet: domain.Snippet{
				Code: "fmt.Println(\"Hello\")",
			},
			wantErr: true,
		},
		{
			name: "rejects a snippet without code",
			snippet: domain.Snippet{
				Title: "Print greeting",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckValidSnippet(tt.snippet)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckValidSnippet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnippetServiceCreateSnippet(t *testing.T) {
	existingSnippet := domain.Snippet{ID: "existing", Title: "Existing", Code: "existing()", CreatedAt: "2026-01-01T00:00:00Z"}
	tests := []struct {
		name            string
		input           domain.CreateSnippetInput
		initialSnippets []domain.Snippet
		wantErr         bool
		wantTitle       string
		wantLanguage    string
		wantCode        string
		wantTags        []string
	}{
		{
			name: "creates a normalized snippet and persists it",
			input: domain.CreateSnippetInput{
				Title: "  Print greeting  ", Language: "  go  ", Code: "  fmt.Println(\"Hello\")  ", Tags: []string{"go"},
			},
			initialSnippets: []domain.Snippet{existingSnippet},
			wantTitle:       "Print greeting",
			wantLanguage:    "go",
			wantCode:        "fmt.Println(\"Hello\")",
			wantTags:        []string{"go"},
		},
		{
			name:            "empty title returns an error without changing the file",
			input:           domain.CreateSnippetInput{Title: "  ", Code: "code"},
			initialSnippets: []domain.Snippet{existingSnippet},
			wantErr:         true,
		},
		{
			name:            "empty code returns an error without changing the file",
			input:           domain.CreateSnippetInput{Title: "Title", Code: "  "},
			initialSnippets: []domain.Snippet{existingSnippet},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare an isolated service with a known persisted list.
			service := newSnippetServiceWithSnippets(t, tt.initialSnippets)
			// Run the method under test.
			created, err := service.CreateSnippet(tt.input)
			// Check whether the returned error matches this case's expectation.
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateSnippet() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Read the file through the service to verify persistence.
			persisted, listErr := service.List()
			if listErr != nil {
				t.Fatalf("List() error = %v", listErr)
			}
			if tt.wantErr {
				// Invalid input must not modify the existing snippets.
				if !reflect.DeepEqual(persisted, tt.initialSnippets) {
					t.Errorf("persisted snippets = %#v, want %#v", persisted, tt.initialSnippets)
				}
				return
			}

			// A successful creation must generate an ID and an RFC3339 timestamp.
			if created.ID == "" {
				t.Error("CreateSnippet() returned an empty ID")
			}
			if _, err := time.Parse(time.RFC3339, created.CreatedAt); err != nil {
				t.Errorf("CreatedAt = %q, want RFC3339 timestamp: %v", created.CreatedAt, err)
			}
			// Check that text fields were normalized and tags were preserved.
			if created.Title != tt.wantTitle || created.Language != tt.wantLanguage || created.Code != tt.wantCode || !reflect.DeepEqual(created.Tags, tt.wantTags) {
				t.Errorf("CreateSnippet() = %#v, want title=%q language=%q code=%q tags=%#v", created, tt.wantTitle, tt.wantLanguage, tt.wantCode, tt.wantTags)
			}
			// The new snippet must be appended to the persisted list.
			wantPersisted := append(append([]domain.Snippet{}, tt.initialSnippets...), created)
			if !reflect.DeepEqual(persisted, wantPersisted) {
				t.Errorf("persisted snippets = %#v, want %#v", persisted, wantPersisted)
			}
		})
	}
}

func TestSnippetServiceUpdateSnippet(t *testing.T) {
	existingSnippet := domain.Snippet{ID: "existing", Title: "Old title", Language: "go", Code: "old()", Tags: []string{"old"}, CreatedAt: "2026-01-01T00:00:00Z"}
	tests := []struct {
		name      string
		snippet   domain.Snippet
		wantErr   bool
		wantTitle string
	}{
		{
			name:      "updates a snippet and keeps its creation time",
			snippet:   domain.Snippet{ID: "existing", Title: "  New title  ", Language: "  rust  ", Code: "  new()  ", Tags: []string{"new"}},
			wantTitle: "New title",
		},
		{
			name:    "unknown ID returns an error without changing the file",
			snippet: domain.Snippet{ID: "missing", Title: "Title", Code: "code"},
			wantErr: true,
		},
		{
			name:    "invalid snippet returns an error without changing the file",
			snippet: domain.Snippet{ID: "existing", Title: "  ", Code: "code"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialSnippets := []domain.Snippet{existingSnippet}
			// Prepare an isolated service with the snippet that may be updated.
			service := newSnippetServiceWithSnippets(t, initialSnippets)
			// Run the method under test.
			updated, err := service.UpdateSnippet(tt.snippet)
			// Check whether the returned error matches this case's expectation.
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateSnippet() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Read the list again to verify the state persisted on disk.
			persisted, listErr := service.List()
			if listErr != nil {
				t.Fatalf("List() error = %v", listErr)
			}
			if tt.wantErr {
				// Failed updates must preserve the original file contents.
				if !reflect.DeepEqual(persisted, initialSnippets) {
					t.Errorf("persisted snippets = %#v, want %#v", persisted, initialSnippets)
				}
				return
			}

			// Updated fields must be normalized while CreatedAt is kept from the stored snippet.
			if updated.Title != tt.wantTitle || updated.Language != "rust" || updated.Code != "new()" || !reflect.DeepEqual(updated.Tags, []string{"new"}) || updated.CreatedAt != existingSnippet.CreatedAt {
				t.Errorf("UpdateSnippet() = %#v, want normalized fields and CreatedAt %q", updated, existingSnippet.CreatedAt)
			}
			// The returned snippet must match the value written to disk.
			if !reflect.DeepEqual(persisted, []domain.Snippet{updated}) {
				t.Errorf("persisted snippets = %#v, want %#v", persisted, []domain.Snippet{updated})
			}
		})
	}
}

// DELETE
func TestSnippetServiceDeleteSnippet(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "snippets.json")
	repo := repository.NewJSONSnippetRepository()
	repo.SetFilePath(filePath)

	if err := repo.EnsureFile(); err != nil {
		t.Fatalf("EnsureFile() error = %v", err)
	}

	snippets := []domain.Snippet{
		{ID: "keep", Title: "Keep this", Code: "keep()"},
		{ID: "delete", Title: "Delete this", Code: "delete()"},
	}
	if err := repo.SaveList(snippets); err != nil {
		t.Fatalf("SaveList() error = %v", err)
	}

	service := NewSnippetService(domain.AppConfig{SnippetsFilePath: filePath}, nil)

	if err := service.DeleteSnippet("delete"); err != nil {
		t.Fatalf("DeleteSnippet() error = %v", err)
	}

	persistedSnippets, err := service.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(persistedSnippets) != 1 {
		t.Fatalf("List() returned %d snippets, want 1", len(persistedSnippets))
	}
	if persistedSnippets[0].ID != "keep" {
		t.Errorf("remaining snippet ID = %q, want %q", persistedSnippets[0].ID, "keep")
	}
}
