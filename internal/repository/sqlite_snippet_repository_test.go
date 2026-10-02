package repository

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"SnippetsDome/internal/domain"
)

func newTestSQLiteRepository(t *testing.T) *SQLiteSnippetRepository {
	t.Helper()
	database, err := OpenSQLite(filepath.Join(t.TempDir(), "snippets.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if err := MigrateSQLite(database); err != nil {
		t.Fatalf("MigrateSQLite() error = %v", err)
	}
	repository, err := NewSQLiteSnippetRepository(database)
	if err != nil {
		t.Fatalf("NewSQLiteSnippetRepository() error = %v", err)
	}
	t.Cleanup(func() { repository.Close() })
	return repository
}

func TestSQLiteRepositoryCollections(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	initial, err := repository.ActiveCollection()
	if err != nil {
		t.Fatalf("ActiveCollection() error = %v", err)
	}
	created, err := repository.CreateCollection("  Work  ")
	if err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if created.Name != "Work" {
		t.Fatalf("created collection name = %q, want Work", created.Name)
	}
	active, err := repository.ActiveCollection()
	if err != nil || active.ID != created.ID {
		t.Fatalf("active collection = %#v, error = %v", active, err)
	}
	if err := repository.SelectCollection(initial.ID); err != nil {
		t.Fatalf("SelectCollection() error = %v", err)
	}
	if err := repository.RenameCollection(created.ID, " general "); !errors.Is(err, ErrCollectionNameExists) {
		t.Fatalf("RenameCollection() error = %v, want duplicate error", err)
	}
	if err := repository.DeleteCollection(initial.ID); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}
	active, err = repository.ActiveCollection()
	if err != nil || active.ID != created.ID {
		t.Fatalf("replacement active collection = %#v, error = %v", active, err)
	}
}

func TestSQLiteRepositorySaveConfig(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	config := domain.AppConfig{
		CloseToTray:      true,
		TraySnippetLimit: 7,
		StartAtLogin:     true,
		Language:         domain.LanguageSpanish,
	}
	if err := repository.SaveConfig(config); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	rows, err := repository.database.Query("SELECT key, value FROM settings WHERE key IN ('closeToTray', 'traySnippetLimit', 'startAtLogin', 'language') ORDER BY key")
	if err != nil {
		t.Fatalf("query saved config: %v", err)
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			t.Fatalf("scan saved config: %v", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate saved config: %v", err)
	}
	want := map[string]string{
		"closeToTray":      "true",
		"traySnippetLimit": "7",
		"startAtLogin":     "true",
		"language":         domain.LanguageSpanish,
	}
	if !reflect.DeepEqual(values, want) {
		t.Errorf("saved config = %#v, want %#v", values, want)
	}
}

func TestSQLiteRepositorySnippetsIsolateCollectionsAndPersistTags(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	first, err := repository.CreateSnippet(domain.Snippet{Title: "Example", Language: " go ", Code: "fmt.Println()", Tags: []string{" Go ", "go", "Tools"}, Favorite: true})
	if err != nil {
		t.Fatalf("CreateSnippet() error = %v", err)
	}
	if !reflect.DeepEqual(first.Tags, []string{"Go", "Tools"}) {
		t.Fatalf("normalized tags = %#v", first.Tags)
	}
	listed, err := repository.ListSnippets()
	if err != nil || len(listed) != 1 {
		t.Fatalf("ListSnippets() = %#v, error = %v", listed, err)
	}
	if !listed[0].Favorite || !reflect.DeepEqual(listed[0].Tags, first.Tags) {
		t.Fatalf("listed snippet = %#v", listed[0])
	}
	updated, err := repository.UpdateSnippet(domain.Snippet{ID: first.ID, Title: "Updated", Code: "new", Tags: []string{"New"}})
	if err != nil {
		t.Fatalf("UpdateSnippet() error = %v", err)
	}
	if updated.CreatedAt != first.CreatedAt {
		t.Fatalf("CreatedAt changed from %q to %q", first.CreatedAt, updated.CreatedAt)
	}
	if err := repository.DeleteSnippet(first.ID); err != nil {
		t.Fatalf("DeleteSnippet() error = %v", err)
	}
	if err := repository.DeleteSnippet(first.ID); !errors.Is(err, ErrSnippetNotFound) {
		t.Fatalf("second DeleteSnippet() error = %v", err)
	}
}

func TestSQLiteRepositoryCollectionRules(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	if _, err := repository.CreateCollection(" "); !errors.Is(err, ErrCollectionNameRequired) {
		t.Fatalf("empty collection error = %v", err)
	}
	if _, err := repository.CreateCollection("general"); !errors.Is(err, ErrCollectionNameExists) {
		t.Fatalf("duplicate collection error = %v", err)
	}
	initial, _ := repository.ActiveCollection()
	if err := repository.DeleteCollection(initial.ID); !errors.Is(err, ErrCannotDeleteLast) {
		t.Fatalf("last collection error = %v", err)
	}
}
