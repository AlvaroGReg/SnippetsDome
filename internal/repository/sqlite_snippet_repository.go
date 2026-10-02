package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"SnippetsDome/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrCollectionNameRequired = errors.New("collection name is required")
	ErrCollectionNameExists   = errors.New("collection name already exists")
	ErrCollectionNotFound     = errors.New("collection not found")
	ErrCannotDeleteLast       = errors.New("cannot delete the last collection")
	ErrSnippetNotFound        = errors.New("snippet not found")
)

// SQLiteSnippetRepository stores collections and snippets in one SQLite database.
type SQLiteSnippetRepository struct {
	database *sql.DB
}

func NewSQLiteSnippetRepository(database *sql.DB) (*SQLiteSnippetRepository, error) {
	if database == nil {
		return nil, errors.New("SQLite database is required")
	}
	return &SQLiteSnippetRepository{database: database}, nil
}

func (r *SQLiteSnippetRepository) Close() error { return r.database.Close() }

func (r *SQLiteSnippetRepository) LoadConfig() (domain.AppConfig, error) {
	config := domain.AppConfig{TraySnippetLimit: domain.DefaultTraySnippetLimit}
	rows, err := r.database.Query("SELECT key, value FROM settings WHERE key IN ('closeToTray', 'traySnippetLimit', 'startAtLogin', 'language')")
	if err != nil {
		return config, fmt.Errorf("load config: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return config, fmt.Errorf("scan config: %w", err)
		}
		switch key {
		case "closeToTray":
			config.CloseToTray = value == "true"
		case "traySnippetLimit":
			if _, err := fmt.Sscanf(value, "%d", &config.TraySnippetLimit); err != nil {
				config.TraySnippetLimit = domain.DefaultTraySnippetLimit
			}
		case "startAtLogin":
			config.StartAtLogin = value == "true"
		case "language":
			config.Language = value
		}
	}
	if err := rows.Err(); err != nil {
		return config, fmt.Errorf("load config: %w", err)
	}
	return config, nil
}

// SaveConfig persists preferences in the same database as snippets.
func (r *SQLiteSnippetRepository) SaveConfig(config domain.AppConfig) error {
	values := map[string]string{
		"closeToTray":      fmt.Sprintf("%t", config.CloseToTray),
		"traySnippetLimit": fmt.Sprintf("%d", config.TraySnippetLimit),
		"startAtLogin":     fmt.Sprintf("%t", config.StartAtLogin),
		"language":         config.Language,
	}
	tx, err := r.database.Begin()
	if err != nil {
		return fmt.Errorf("begin save config: %w", err)
	}
	defer tx.Rollback()
	for key, value := range values {
		if _, err := tx.Exec("INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value); err != nil {
			return fmt.Errorf("save config %s: %w", key, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit config: %w", err)
	}
	return nil
}

func (r *SQLiteSnippetRepository) ActiveCollection() (domain.Collection, error) {
	var collection domain.Collection
	err := r.database.QueryRow(`
		SELECT c.id, c.name, c.created_at
		FROM collections c JOIN settings s ON s.value = c.id
		WHERE s.key = 'activeCollectionId'`).Scan(&collection.ID, &collection.Name, &collection.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Collection{}, ErrCollectionNotFound
	}
	if err != nil {
		return domain.Collection{}, fmt.Errorf("read active collection: %w", err)
	}
	return collection, nil
}

func (r *SQLiteSnippetRepository) ListCollections() ([]domain.Collection, error) {
	rows, err := r.database.Query("SELECT id, name, created_at FROM collections ORDER BY position, id")
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()
	collections := make([]domain.Collection, 0)
	for rows.Next() {
		var collection domain.Collection
		if err := rows.Scan(&collection.ID, &collection.Name, &collection.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		collections = append(collections, collection)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	return collections, nil
}

// CreateCollection creates a collection and makes it active atomically.
func (r *SQLiteSnippetRepository) CreateCollection(name string) (domain.Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Collection{}, ErrCollectionNameRequired
	}
	tx, err := r.database.Begin()
	if err != nil {
		return domain.Collection{}, fmt.Errorf("begin create collection: %w", err)
	}
	defer tx.Rollback()
	if err := ensureCollectionNameAvailable(tx, name, ""); err != nil {
		return domain.Collection{}, err
	}
	var position int
	if err := tx.QueryRow("SELECT COALESCE(MAX(position), -1) + 1 FROM collections").Scan(&position); err != nil {
		return domain.Collection{}, fmt.Errorf("read collection position: %w", err)
	}
	collection := domain.Collection{ID: uuid.NewString(), Name: name, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if _, err := tx.Exec("INSERT INTO collections (id, name, created_at, position) VALUES (?, ?, ?, ?)", collection.ID, collection.Name, collection.CreatedAt, position); err != nil {
		return domain.Collection{}, fmt.Errorf("create collection: %w", err)
	}
	if _, err := tx.Exec("INSERT INTO settings (key, value) VALUES ('activeCollectionId', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", collection.ID); err != nil {
		return domain.Collection{}, fmt.Errorf("select created collection: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Collection{}, fmt.Errorf("commit create collection: %w", err)
	}
	return collection, nil
}

func (r *SQLiteSnippetRepository) RenameCollection(id, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrCollectionNameRequired
	}
	tx, err := r.database.Begin()
	if err != nil {
		return fmt.Errorf("begin rename collection: %w", err)
	}
	defer tx.Rollback()
	if err := ensureCollectionNameAvailable(tx, name, id); err != nil {
		return err
	}
	result, err := tx.Exec("UPDATE collections SET name = ? WHERE id = ?", name, id)
	if err != nil {
		return fmt.Errorf("rename collection: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return ErrCollectionNotFound
	}
	return tx.Commit()
}

func (r *SQLiteSnippetRepository) DeleteCollection(id string) error {
	tx, err := r.database.Begin()
	if err != nil {
		return fmt.Errorf("begin delete collection: %w", err)
	}
	defer tx.Rollback()
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM collections").Scan(&count); err != nil {
		return fmt.Errorf("count collections: %w", err)
	}
	if count <= 1 {
		return ErrCannotDeleteLast
	}
	var exists bool
	if err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE id = ?)", id).Scan(&exists); err != nil {
		return fmt.Errorf("find collection: %w", err)
	}
	if !exists {
		return ErrCollectionNotFound
	}
	var active string
	_ = tx.QueryRow("SELECT value FROM settings WHERE key = 'activeCollectionId'").Scan(&active)
	if _, err := tx.Exec("DELETE FROM collections WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	if active == id {
		if _, err := tx.Exec("UPDATE settings SET value = (SELECT id FROM collections ORDER BY position, id LIMIT 1) WHERE key = 'activeCollectionId'"); err != nil {
			return fmt.Errorf("select replacement collection: %w", err)
		}
	}
	return tx.Commit()
}

func (r *SQLiteSnippetRepository) SelectCollection(id string) error {
	result, err := r.database.Exec("UPDATE settings SET value = ? WHERE key = 'activeCollectionId' AND EXISTS (SELECT 1 FROM collections WHERE id = ?)", id, id)
	if err != nil {
		return fmt.Errorf("select collection: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return ErrCollectionNotFound
	}
	return nil
}

func (r *SQLiteSnippetRepository) ListSnippets() ([]domain.Snippet, error) {
	collection, err := r.ActiveCollection()
	if err != nil {
		return nil, err
	}
	rows, err := r.database.Query("SELECT id, title, language, code, created_at, favorite FROM snippets WHERE collection_id = ? ORDER BY favorite DESC, position, id", collection.ID)
	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}
	defer rows.Close()
	var snippets []domain.Snippet
	for rows.Next() {
		var snippet domain.Snippet
		var favorite int
		if err := rows.Scan(&snippet.ID, &snippet.Title, &snippet.Language, &snippet.Code, &snippet.CreatedAt, &favorite); err != nil {
			return nil, fmt.Errorf("scan snippet: %w", err)
		}
		snippet.Favorite = favorite != 0
		snippets = append(snippets, snippet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close snippets: %w", err)
	}
	for index := range snippets {
		snippets[index].Tags, err = r.tagsForSnippet(snippets[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return snippets, nil
}

func (r *SQLiteSnippetRepository) CreateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
	return r.saveSnippet(snippet, false)
}

func (r *SQLiteSnippetRepository) UpdateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
	return r.saveSnippet(snippet, true)
}

func (r *SQLiteSnippetRepository) saveSnippet(snippet domain.Snippet, update bool) (domain.Snippet, error) {
	collection, err := r.ActiveCollection()
	if err != nil {
		return domain.Snippet{}, err
	}
	tx, err := r.database.Begin()
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("begin save snippet: %w", err)
	}
	defer tx.Rollback()
	snippet.Title = strings.TrimSpace(snippet.Title)
	snippet.Language = strings.TrimSpace(snippet.Language)
	if snippet.ID == "" {
		snippet.ID = uuid.NewString()
	}
	if update {
		var existingCreatedAt string
		if err := tx.QueryRow("SELECT created_at FROM snippets WHERE id = ? AND collection_id = ?", snippet.ID, collection.ID).Scan(&existingCreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.Snippet{}, ErrSnippetNotFound
			}
			return domain.Snippet{}, fmt.Errorf("read snippet: %w", err)
		}
		snippet.CreatedAt = existingCreatedAt
		result, updateErr := tx.Exec("UPDATE snippets SET title = ?, language = ?, code = ?, favorite = ? WHERE id = ? AND collection_id = ?", snippet.Title, snippet.Language, snippet.Code, boolInt(snippet.Favorite), snippet.ID, collection.ID)
		if updateErr != nil {
			return domain.Snippet{}, fmt.Errorf("update snippet: %w", updateErr)
		}
		if count, _ := result.RowsAffected(); count == 0 {
			return domain.Snippet{}, ErrSnippetNotFound
		}
	} else {
		if snippet.CreatedAt == "" {
			snippet.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		}
		var position int
		if err := tx.QueryRow("SELECT COALESCE(MAX(position), -1) + 1 FROM snippets WHERE collection_id = ?", collection.ID).Scan(&position); err != nil {
			return domain.Snippet{}, fmt.Errorf("read snippet position: %w", err)
		}
		if _, err := tx.Exec("INSERT INTO snippets (id, collection_id, title, language, code, created_at, favorite, position) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", snippet.ID, collection.ID, snippet.Title, snippet.Language, snippet.Code, snippet.CreatedAt, boolInt(snippet.Favorite), position); err != nil {
			return domain.Snippet{}, fmt.Errorf("create snippet: %w", err)
		}
	}
	if _, err := tx.Exec("DELETE FROM snippet_tags WHERE snippet_id = ?", snippet.ID); err != nil {
		return domain.Snippet{}, fmt.Errorf("replace snippet tags: %w", err)
	}
	for _, tag := range normalizeTags(snippet.Tags) {
		if _, err := tx.Exec("INSERT INTO snippet_tags (snippet_id, tag) VALUES (?, ?)", snippet.ID, tag); err != nil {
			return domain.Snippet{}, fmt.Errorf("save snippet tag: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Snippet{}, fmt.Errorf("commit snippet: %w", err)
	}
	snippet.Tags = normalizeTags(snippet.Tags)
	return snippet, nil
}

func (r *SQLiteSnippetRepository) DeleteSnippet(id string) error {
	collection, err := r.ActiveCollection()
	if err != nil {
		return err
	}
	result, err := r.database.Exec("DELETE FROM snippets WHERE id = ? AND collection_id = ?", id, collection.ID)
	if err != nil {
		return fmt.Errorf("delete snippet: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return ErrSnippetNotFound
	}
	return nil
}

func (r *SQLiteSnippetRepository) tagsForSnippet(id string) ([]string, error) {
	rows, err := r.database.Query("SELECT tag FROM snippet_tags WHERE snippet_id = ? ORDER BY rowid", id)
	if err != nil {
		return nil, fmt.Errorf("list snippet tags: %w", err)
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func ensureCollectionNameAvailable(tx *sql.Tx, name, excludedID string) error {
	var exists bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE LOWER(TRIM(name)) = LOWER(TRIM(?)) AND id <> ?)", name, excludedID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check collection name: %w", err)
	}
	if exists {
		return ErrCollectionNameExists
	}
	return nil
}

func normalizeTags(tags []string) []string {
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

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
