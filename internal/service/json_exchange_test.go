package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/repository"
)

func TestSnippetServiceJSONExchange(t *testing.T) {
	database, err := repository.OpenSQLite(filepath.Join(t.TempDir(), "snippets.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := repository.MigrateSQLite(database); err != nil {
		t.Fatal(err)
	}
	snippetRepository, err := repository.NewSQLiteSnippetRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	service := NewSnippetService(domain.AppConfig{}, snippetRepository, snippetRepository)

	input := `[{
		"id":"imported-id","title":"Imported","language":"Go","code":"fmt.Println()",
		"tags":["go"," Go "],"createdAt":"2026-01-01T00:00:00Z","favorite":true
	}]`
	inputPath := filepath.Join(t.TempDir(), "legacy.json")
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	collection, err := service.ImportJSON(inputPath, "Legacy")
	if err != nil {
		t.Fatalf("ImportJSON() error = %v", err)
	}
	if collection.Name != "Legacy" {
		t.Fatalf("imported collection = %#v", collection)
	}
	snippets, err := service.List()
	if err != nil || len(snippets) != 1 || snippets[0].ID != "imported-id" || !snippets[0].Favorite {
		t.Fatalf("imported snippets = %#v, error = %v", snippets, err)
	}

	outputPath := filepath.Join(t.TempDir(), "export.json")
	if err := service.ExportJSON(outputPath); err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id": "imported-id"`) || !strings.Contains(string(data), `"favorite": true`) {
		t.Errorf("exported JSON = %s", data)
	}
}

func TestSnippetServiceImportJSONRejectsInvalidInputWithoutChanges(t *testing.T) {
	service := newEmptySnippetService(t, domain.AppConfig{}, nil)
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`[{"id":"duplicate","title":"One","code":"one","createdAt":"now"},{"id":"duplicate","title":"Two","code":"two","createdAt":"now"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ImportJSON(path, "Invalid"); err == nil {
		t.Fatal("ImportJSON() error = nil, want duplicate ID error")
	}
	collections, err := service.ListCollections()
	if err != nil || len(collections) != 1 {
		t.Fatalf("collections after failed import = %#v, error = %v", collections, err)
	}
}
