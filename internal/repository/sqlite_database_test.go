package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteDatabasePath(t *testing.T) {
	baseDirectory := t.TempDir()
	path, err := SQLiteDatabasePath(func() (string, error) {
		return baseDirectory, nil
	})
	if err != nil {
		t.Fatalf("SQLiteDatabasePath() error = %v", err)
	}
	want := filepath.Join(baseDirectory, "snippets.db")
	if path != want {
		t.Fatalf("SQLiteDatabasePath() = %q, want %q", path, want)
	}
}

func TestSQLiteDatabasePathReturnsResolverErrors(t *testing.T) {
	wantErr := errors.New("directory unavailable")
	_, err := SQLiteDatabasePath(func() (string, error) {
		return "", wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("SQLiteDatabasePath() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestOpenSQLiteCreatesDirectoryAndConfiguresConnection(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "data", "snippets.db")
	database, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer database.Close()

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("database file was not created: %v", err)
	}
	var journalMode string
	if err := database.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("reading journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal mode = %q, want wal", journalMode)
	}
	var busyTimeout int
	if err := database.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("reading busy timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy timeout = %d, want 5000", busyTimeout)
	}
}

func TestOpenSQLiteRejectsEmptyPath(t *testing.T) {
	if _, err := OpenSQLite(""); err == nil {
		t.Fatal("OpenSQLite() error = nil, want an error")
	}
}

func TestOpenSQLiteCanClose(t *testing.T) {
	database, err := OpenSQLite(filepath.Join(t.TempDir(), "snippets.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := database.Ping(); err == nil {
		t.Fatal("Ping() after Close() error = nil, want an error")
	}
}

func TestMigrateSQLiteCreatesSchemaAndInitialCollection(t *testing.T) {
	database, err := OpenSQLite(filepath.Join(t.TempDir(), "snippets.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer database.Close()

	if err := MigrateSQLite(database); err != nil {
		t.Fatalf("MigrateSQLite() error = %v", err)
	}

	var version int
	if err := database.QueryRow("SELECT version FROM schema_version").Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != currentSchemaVersion {
		t.Fatalf("schema version = %d, want %d", version, currentSchemaVersion)
	}
	var collectionName string
	if err := database.QueryRow("SELECT name FROM collections").Scan(&collectionName); err != nil {
		t.Fatalf("read initial collection: %v", err)
	}
	if collectionName != "General" {
		t.Fatalf("initial collection = %q, want General", collectionName)
	}
	var activeCollectionID, collectionID string
	if err := database.QueryRow("SELECT value FROM settings WHERE key = 'activeCollectionId'").Scan(&activeCollectionID); err != nil {
		t.Fatalf("read active collection: %v", err)
	}
	if err := database.QueryRow("SELECT id FROM collections WHERE name = 'General'").Scan(&collectionID); err != nil {
		t.Fatalf("read collection ID: %v", err)
	}
	if activeCollectionID != collectionID {
		t.Fatalf("active collection = %q, want %q", activeCollectionID, collectionID)
	}
}

func TestMigrateSQLiteIsIdempotent(t *testing.T) {
	database, err := OpenSQLite(filepath.Join(t.TempDir(), "snippets.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer database.Close()

	if err := MigrateSQLite(database); err != nil {
		t.Fatalf("first MigrateSQLite() error = %v", err)
	}
	if err := MigrateSQLite(database); err != nil {
		t.Fatalf("second MigrateSQLite() error = %v", err)
	}
	var versionRows, collectionCount int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&versionRows); err != nil {
		t.Fatalf("count schema versions: %v", err)
	}
	if err := database.QueryRow("SELECT COUNT(*) FROM collections").Scan(&collectionCount); err != nil {
		t.Fatalf("count collections: %v", err)
	}
	if versionRows != 1 || collectionCount != 1 {
		t.Fatalf("schema rows = %d, collections = %d, want 1 and 1", versionRows, collectionCount)
	}
}

func TestMigrateSQLiteRollsBackFailedMigration(t *testing.T) {
	database, err := OpenSQLite(filepath.Join(t.TempDir(), "snippets.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer database.Close()
	if _, err := database.Exec("CREATE TABLE schema_version (version INTEGER NOT NULL)"); err != nil {
		t.Fatalf("create schema version fixture: %v", err)
	}
	if _, err := database.Exec("INSERT INTO schema_version (version) VALUES (0)"); err != nil {
		t.Fatalf("insert schema version fixture: %v", err)
	}
	if _, err := database.Exec("CREATE TABLE collections (id TEXT PRIMARY KEY)"); err != nil {
		t.Fatalf("create invalid collections fixture: %v", err)
	}

	if err := MigrateSQLite(database); err == nil {
		t.Fatal("MigrateSQLite() error = nil, want migration failure")
	}
	var version int
	if err := database.QueryRow("SELECT version FROM schema_version").Scan(&version); err != nil {
		t.Fatalf("read schema version after rollback: %v", err)
	}
	if version != 0 {
		t.Fatalf("schema version after rollback = %d, want 0", version)
	}
	var tableCount int
	if err := database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'snippets'").Scan(&tableCount); err != nil {
		t.Fatalf("check rolled back table: %v", err)
	}
	if tableCount != 0 {
		t.Fatal("snippets table exists after failed migration")
	}
}
