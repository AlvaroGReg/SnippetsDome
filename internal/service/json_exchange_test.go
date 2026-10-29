package service

import (
	"errors"
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

	if _, err := service.CreateCollection("Other"); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	collectionOutputPath := filepath.Join(t.TempDir(), "legacy-export.json")
	if err := service.ExportCollectionJSON(collectionOutputPath, collection.ID); err != nil {
		t.Fatalf("ExportCollectionJSON() error = %v", err)
	}
	collectionData, err := os.ReadFile(collectionOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(collectionData), `"id": "imported-id"`) {
		t.Errorf("collection export = %s, want imported snippet", collectionData)
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

func TestSnippetServiceEncryptedCollectionExportIsRejected(t *testing.T) {
	service := newEmptySnippetService(t, domain.AppConfig{}, nil)
	active, err := service.ActiveCollection()
	if err != nil {
		t.Fatal(err)
	}
	if err := service.EncryptCollection(active.ID, "0042", "0042"); err != nil {
		t.Fatal(err)
	}
	for name, export := range map[string]func(string) error{
		"active collection":    func(path string) error { return service.ExportJSON(path) },
		"specified collection": func(path string) error { return service.ExportCollectionJSON(path, active.ID) },
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "export.json")
			if err := export(path); !errors.Is(err, ErrEncryptedCollectionExport) {
				t.Fatalf("export error = %v, want ErrEncryptedCollectionExport", err)
			}
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("export file stat error = %v, want file not created", err)
			}
		})
	}
}

func TestSnippetServiceImportJSONRemainsUnencrypted(t *testing.T) {
	service := newEmptySnippetService(t, domain.AppConfig{}, nil)
	active, err := service.ActiveCollection()
	if err != nil {
		t.Fatal(err)
	}
	if err := service.EncryptCollection(active.ID, "0042", "0042"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "legacy.json")
	data := `[{"id":"plain-import","title":"Imported","language":"Go","code":"plain()","createdAt":"2026-01-01T00:00:00Z"}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	imported, err := service.ImportJSON(path, "Imported")
	if err != nil {
		t.Fatalf("ImportJSON() error = %v", err)
	}
	if imported.Encrypted {
		t.Fatalf("imported collection = %#v, want unencrypted", imported)
	}
	snippets, err := service.List()
	if err != nil || len(snippets) != 1 || snippets[0].Title != "Imported" {
		t.Fatalf("imported snippets = %#v, error = %v", snippets, err)
	}
}
