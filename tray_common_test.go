package main

import (
	"path/filepath"
	"testing"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/repository"
	"SnippetsDome/internal/service"
)

func TestTraySnippetsEncryptedCollectionLockState(t *testing.T) {
	database, err := repository.OpenSQLite(filepath.Join(t.TempDir(), "tray.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := repository.MigrateSQLite(database); err != nil {
		t.Fatal(err)
	}
	repo, err := repository.NewSQLiteSnippetRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateSnippet(domain.Snippet{ID: "tray-secret", Title: "Secret title", Code: "secret()", CreatedAt: "2026-01-01T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	snippets := service.NewSnippetService(domain.AppConfig{}, repo, repo)
	collection, err := snippets.ActiveCollection()
	if err != nil {
		t.Fatal(err)
	}
	if err := snippets.EncryptCollection(collection.ID, "0042", "0042"); err != nil {
		t.Fatal(err)
	}
	app := &App{snippets: snippets}
	if got := traySnippets(app); len(got) != 1 || got[0].Title != "Secret title" {
		t.Fatalf("traySnippets() while unlocked = %#v, want decrypted snippet", got)
	}
	if err := snippets.LockActiveCollection(); err != nil {
		t.Fatal(err)
	}
	if got := traySnippets(app); len(got) != 0 {
		t.Fatalf("traySnippets() while locked = %#v, want no snippets", got)
	}
	if _, err := snippets.UnlockCollection(collection.ID, "0042"); err != nil {
		t.Fatal(err)
	}
	if got := traySnippets(app); len(got) != 1 || got[0].Code != "secret()" {
		t.Fatalf("traySnippets() after unlock = %#v, want decrypted snippet", got)
	}
}

func TestTraySnippetTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "empty title", title: "  ", want: "Untitled snippet"},
		{name: "trimmed title", title: "  Go example  ", want: "Go example"},
		{name: "long title", title: "12345678901234567890123456789012345678901234567890", want: "12345678901234567890123456789012345678901234567…"},
		{name: "long unicode title", title: "こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは世界", want: "こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは世界こんにちは…"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := traySnippetTitle(test.title); got != test.want {
				t.Errorf("traySnippetTitle(%q) = %q, want %q", test.title, got, test.want)
			}
		})
	}
}
