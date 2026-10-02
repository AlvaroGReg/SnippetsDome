package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const currentSchemaVersion = 1

// MigrateSQLite applies all pending schema migrations atomically.
func MigrateSQLite(database *sql.DB) error {
	transaction, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin SQLite migration: %w", err)
	}
	defer transaction.Rollback()

	if _, err := transaction.Exec(`
        CREATE TABLE IF NOT EXISTS schema_version (
            version INTEGER NOT NULL
        )
    `); err != nil {
		return fmt.Errorf("create schema version table: %w", err)
	}

	version, err := schemaVersion(transaction)
	if err != nil {
		return err
	}
	if version > currentSchemaVersion {
		return fmt.Errorf("unsupported SQLite schema version %d", version)
	}
	if version < 1 {
		if err := migrateSQLiteV1(transaction); err != nil {
			return fmt.Errorf("apply SQLite schema version 1: %w", err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit SQLite migration: %w", err)
	}
	return nil
}

func schemaVersion(transaction *sql.Tx) (int, error) {
	var version int
	err := transaction.QueryRow("SELECT version FROM schema_version LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read SQLite schema version: %w", err)
	}
	return version, nil
}

func migrateSQLiteV1(transaction *sql.Tx) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS collections (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            created_at TEXT NOT NULL,
            position INTEGER NOT NULL DEFAULT 0
        )`,
		`CREATE UNIQUE INDEX IF NOT EXISTS collections_name_unique ON collections (LOWER(TRIM(name)))`,
		`CREATE INDEX IF NOT EXISTS collections_position_index ON collections (position)`,
		`CREATE TABLE IF NOT EXISTS snippets (
            id TEXT PRIMARY KEY,
            collection_id TEXT NOT NULL,
            title TEXT NOT NULL,
            language TEXT NOT NULL DEFAULT '',
            code TEXT NOT NULL,
            created_at TEXT NOT NULL,
            favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
            position INTEGER NOT NULL DEFAULT 0,
            FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
        )`,
		`CREATE INDEX IF NOT EXISTS snippets_collection_order_index ON snippets (collection_id, favorite DESC, position)`,
		`CREATE TABLE IF NOT EXISTS snippet_tags (
            snippet_id TEXT NOT NULL,
            tag TEXT NOT NULL,
            PRIMARY KEY (snippet_id, tag),
            FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
        )`,
		`CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL
        )`,
	}
	for _, statement := range statements {
		if _, err := transaction.Exec(statement); err != nil {
			return err
		}
	}

	var collectionCount int
	if err := transaction.QueryRow("SELECT COUNT(*) FROM collections").Scan(&collectionCount); err != nil {
		return fmt.Errorf("count initial collections: %w", err)
	}
	if collectionCount == 0 {
		collectionID := uuid.NewString()
		if _, err := transaction.Exec(
			"INSERT INTO collections (id, name, created_at, position) VALUES (?, ?, ?, 0)",
			collectionID, "General", time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			return fmt.Errorf("create initial collection: %w", err)
		}
		if _, err := transaction.Exec(
			"INSERT INTO settings (key, value) VALUES ('activeCollectionId', ?)", collectionID,
		); err != nil {
			return fmt.Errorf("store active collection: %w", err)
		}
	}

	if _, err := transaction.Exec("INSERT INTO schema_version (version) VALUES (1)"); err != nil {
		return fmt.Errorf("store SQLite schema version: %w", err)
	}
	return nil
}
