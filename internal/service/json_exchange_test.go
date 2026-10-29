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
	if err != nil || len(snippets) != 1 || snippets[0].ID == "imported-id" || !snippets[0].Favorite {
		t.Fatalf("imported snippets = %#v, error = %v", snippets, err)
	}
	firstImportedID := snippets[0].ID
	if snippets[0].Title != "Imported" || snippets[0].Code != "fmt.Println()" {
		t.Fatalf("imported snippet content = %#v", snippets[0])
	}

	if _, err := service.ImportJSON(inputPath, "Legacy copy"); err != nil {
		t.Fatalf("import same JSON a second time: %v", err)
	}
	secondImport, err := service.List()
	if err != nil || len(secondImport) != 1 || secondImport[0].ID == firstImportedID || secondImport[0].ID == "imported-id" {
		t.Fatalf("second imported snippets = %#v, error = %v", secondImport, err)
	}
	if secondImport[0].Title != "Imported" || secondImport[0].Code != "fmt.Println()" || !secondImport[0].Favorite {
		t.Fatalf("second imported snippet content = %#v", secondImport[0])
	}

	outputPath := filepath.Join(t.TempDir(), "export.json")
	if err := service.ExportJSON(outputPath); err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id": "`+secondImport[0].ID+`"`) || !strings.Contains(string(data), `"favorite": true`) {
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
	if !strings.Contains(string(collectionData), `"id": "`+firstImportedID+`"`) {
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

func TestSnippetServiceEncryptedCollectionJSONRoundTrip(t *testing.T) {
	service := newEmptySnippetService(t, domain.AppConfig{}, nil)
	active, err := service.ActiveCollection()
	if err != nil {
		t.Fatal(err)
	}
	if err := service.EncryptCollection(active.ID, "0042", "0042"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateSnippet(domain.CreateSnippetInput{Title: "Private", Language: "Go", Code: "secret code", Tags: []string{"private"}}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "encrypted.json")
	if err := service.ExportEncryptedCollectionJSON(path, active.ID, "0042", "new export password"); err != nil {
		t.Fatalf("ExportEncryptedCollectionJSON() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := IsEncryptedJSONFile(path)
	if err != nil || !encrypted {
		t.Fatalf("IsEncryptedJSONFile() = %t, error = %v; want encrypted file", encrypted, err)
	}
	if strings.Contains(string(data), "secret code") {
		t.Fatalf("encrypted export contains plaintext: %s", data)
	}
	if err := service.ExportEncryptedCollectionJSON(filepath.Join(t.TempDir(), "bad.json"), active.ID, "9999", "password"); !errors.Is(err, ErrCollectionAuthentication) {
		t.Fatalf("wrong collection password error = %v", err)
	}

	importService := newEmptySnippetService(t, domain.AppConfig{}, nil)
	if _, err := importService.ImportJSONWithPassword(path, "Imported", "wrong password"); !errors.Is(err, ErrEncryptedJSONPassword) {
		t.Fatalf("wrong export password error = %v", err)
	}
	imported, err := importService.ImportJSONWithPassword(path, "Imported", "new export password")
	if err != nil {
		t.Fatalf("ImportJSONWithPassword() error = %v", err)
	}
	if imported.Encrypted {
		t.Fatalf("imported collection = %#v, want unencrypted collection", imported)
	}
	snippets, err := importService.List()
	if err != nil || len(snippets) != 1 || snippets[0].Code != "secret code" || len(snippets[0].Tags) != 1 || snippets[0].Tags[0] != "private" {
		t.Fatalf("imported snippets = %#v, error = %v", snippets, err)
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
	encrypted, err := IsEncryptedJSONFile(path)
	if err != nil || encrypted {
		t.Fatalf("IsEncryptedJSONFile() = %t, error = %v; want legacy JSON", encrypted, err)
	}
	imported, err := service.ImportJSONWithPassword(path, "Imported", "")
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
