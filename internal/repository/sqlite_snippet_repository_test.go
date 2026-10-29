package repository

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"SnippetsDome/internal/cryptography"
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

func TestSQLiteRepositoryListSnippetsReturnsEmptySliceForEmptyCollection(t *testing.T) {
	repository := newTestSQLiteRepository(t)

	snippets, err := repository.ListSnippets()
	if err != nil {
		t.Fatalf("ListSnippets() error = %v", err)
	}
	if snippets == nil {
		t.Fatal("ListSnippets() returned nil, want an empty slice")
	}
	if len(snippets) != 0 {
		t.Fatalf("ListSnippets() returned %d snippets, want 0", len(snippets))
	}
}

func TestSQLiteRepositoryDefaultCollectionPersistsAndFallsBackAfterDeletion(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	initial, err := repository.ActiveCollection()
	if err != nil {
		t.Fatalf("ActiveCollection() error = %v", err)
	}
	if got, err := repository.DefaultCollection(); err != nil || got.ID != initial.ID {
		t.Fatalf("initial DefaultCollection() = %#v, error = %v", got, err)
	}
	second, err := repository.CreateCollection("Work")
	if err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if err := repository.SetDefaultCollection(second.ID); err != nil {
		t.Fatalf("SetDefaultCollection() error = %v", err)
	}
	if err := repository.SelectCollection(initial.ID); err != nil {
		t.Fatalf("SelectCollection() error = %v", err)
	}
	if got, err := repository.DefaultCollection(); err != nil || got.ID != second.ID {
		t.Fatalf("DefaultCollection() = %#v, error = %v", got, err)
	}
	if err := repository.DeleteCollection(second.ID); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}
	if got, err := repository.DefaultCollection(); err != nil || got.ID != initial.ID {
		t.Fatalf("replacement DefaultCollection() = %#v, error = %v", got, err)
	}
}

func TestSQLiteRepositoryListSnippetsReturnsEmptyTagsSlice(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	if _, err := repository.CreateSnippet(domain.Snippet{Title: "No tags", Language: "go", Code: "fmt.Println()"}); err != nil {
		t.Fatalf("CreateSnippet() error = %v", err)
	}

	snippets, err := repository.ListSnippets()
	if err != nil {
		t.Fatalf("ListSnippets() error = %v", err)
	}
	if len(snippets) != 1 {
		t.Fatalf("ListSnippets() returned %d snippets, want 1", len(snippets))
	}
	if snippets[0].Tags == nil {
		t.Fatal("ListSnippets() returned nil tags, want an empty slice")
	}
	if len(snippets[0].Tags) != 0 {
		t.Fatalf("ListSnippets() returned tags %#v, want an empty slice", snippets[0].Tags)
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

func TestSQLiteRepositoryEncryptDecryptCollectionPreservesOutOfScopeFields(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	collection, err := repository.ActiveCollection()
	if err != nil {
		t.Fatalf("ActiveCollection() error = %v", err)
	}
	snippet, err := repository.CreateSnippet(domain.Snippet{Title: "Title", Language: "go", Code: "secret code", Tags: []string{"Go", "Tools"}, Favorite: true})
	if err != nil {
		t.Fatalf("CreateSnippet() error = %v", err)
	}
	var created, collectionPosition, snippetPosition int
	if err := repository.database.QueryRow("SELECT position FROM collections WHERE id = ?", collection.ID).Scan(&collectionPosition); err != nil {
		t.Fatal(err)
	}
	if err := repository.database.QueryRow("SELECT position FROM snippets WHERE id = ?", snippet.ID).Scan(&snippetPosition); err != nil {
		t.Fatal(err)
	}
	var createdAt string
	if err := repository.database.QueryRow("SELECT created_at FROM snippets WHERE id = ?", snippet.ID).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}
	key := make([]byte, cryptography.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	salt, err := cryptography.NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	parameters := cryptography.DefaultParameters()
	if err := repository.EncryptCollection(collection.ID, key, salt, parameters, "v1:check"); err != nil {
		t.Fatalf("EncryptCollection() error = %v", err)
	}
	var encryptedName, encryptedTitle, encryptedLanguage, encryptedCode string
	if err := repository.database.QueryRow("SELECT name,position FROM collections WHERE id = ?", collection.ID).Scan(&encryptedName, &created); err != nil {
		t.Fatal(err)
	}
	if err := repository.database.QueryRow("SELECT title,language,code,created_at,favorite,position FROM snippets WHERE id = ?", snippet.ID).Scan(&encryptedTitle, &encryptedLanguage, &encryptedCode, &createdAt, &created, &snippetPosition); err != nil {
		t.Fatal(err)
	}
	if encryptedName == collection.Name || encryptedTitle == snippet.Title || encryptedLanguage == snippet.Language || encryptedCode == snippet.Code {
		t.Fatal("plaintext fields remained after encryption")
	}
	if created != 1 || createdAt != snippet.CreatedAt || snippetPosition != 0 || collectionPosition != 0 {
		t.Fatalf("out-of-scope fields changed: favorite=%d createdAt=%q snippetPosition=%d collectionPosition=%d", created, createdAt, snippetPosition, collectionPosition)
	}
	var encryptedTagCount int
	if err := repository.database.QueryRow("SELECT COUNT(*) FROM snippet_tags WHERE snippet_id = ? AND tag LIKE 'v1:%'", snippet.ID).Scan(&encryptedTagCount); err != nil || encryptedTagCount != 2 {
		t.Fatalf("encrypted tag count = %d, error = %v", encryptedTagCount, err)
	}
	var encrypted int
	if err := repository.database.QueryRow("SELECT encrypted FROM collections WHERE id = ?", collection.ID).Scan(&encrypted); err != nil || encrypted != 1 {
		t.Fatalf("encrypted state = %d, error = %v", encrypted, err)
	}
	if err := repository.DecryptCollection(collection.ID, key); err != nil {
		t.Fatalf("DecryptCollection() error = %v", err)
	}
	got, err := repository.ListCollectionSnippets(collection.ID)
	if err != nil || len(got) != 1 {
		t.Fatalf("ListCollectionSnippets() = %#v, error = %v", got, err)
	}
	if got[0].Title != snippet.Title || got[0].Language != snippet.Language || got[0].Code != snippet.Code || !reflect.DeepEqual(got[0].Tags, snippet.Tags) || !got[0].Favorite {
		t.Fatalf("decrypted snippet = %#v", got[0])
	}
	var finalName string
	var finalEncrypted int
	if err := repository.database.QueryRow("SELECT name,encrypted FROM collections WHERE id = ?", collection.ID).Scan(&finalName, &finalEncrypted); err != nil {
		t.Fatal(err)
	}
	if finalName != collection.Name || finalEncrypted != 0 {
		t.Fatalf("decrypted collection name=%q encrypted=%d", finalName, finalEncrypted)
	}
}

func TestSQLiteRepositoryDecryptCollectionRollsBackOnCorruptValue(t *testing.T) {
	repository := newTestSQLiteRepository(t)
	collection, _ := repository.ActiveCollection()
	first, _ := repository.CreateSnippet(domain.Snippet{Title: "First", Code: "one"})
	second, _ := repository.CreateSnippet(domain.Snippet{Title: "Second", Code: "two"})
	key := make([]byte, cryptography.KeySize)
	salt, _ := cryptography.NewSalt()
	if err := repository.EncryptCollection(collection.ID, key, salt, cryptography.DefaultParameters(), "v1:check"); err != nil {
		t.Fatalf("EncryptCollection() error = %v", err)
	}
	if _, err := repository.database.Exec("UPDATE snippets SET code = 'invalid' WHERE id = ?", second.ID); err != nil {
		t.Fatal(err)
	}
	beforeName, beforeTitle, beforeCode := "", "", ""
	if err := repository.database.QueryRow("SELECT name FROM collections WHERE id = ?", collection.ID).Scan(&beforeName); err != nil {
		t.Fatal(err)
	}
	if err := repository.database.QueryRow("SELECT title,code FROM snippets WHERE id = ?", first.ID).Scan(&beforeTitle, &beforeCode); err != nil {
		t.Fatal(err)
	}
	if err := repository.DecryptCollection(collection.ID, key); err == nil {
		t.Fatal("DecryptCollection() error = nil, want corrupt ciphertext error")
	}
	var afterName, afterTitle, afterCode string
	var encrypted int
	if err := repository.database.QueryRow("SELECT name,encrypted FROM collections WHERE id = ?", collection.ID).Scan(&afterName, &encrypted); err != nil {
		t.Fatal(err)
	}
	if err := repository.database.QueryRow("SELECT title,code FROM snippets WHERE id = ?", first.ID).Scan(&afterTitle, &afterCode); err != nil {
		t.Fatal(err)
	}
	if beforeName != afterName || beforeTitle != afterTitle || beforeCode != afterCode || encrypted != 1 {
		t.Fatal("failed decryption left a partial transformation")
	}
}
