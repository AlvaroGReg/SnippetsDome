package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

const sqliteDatabaseFileName = "snippets.db"

// DataDirectoryResolver returns the application data directory.
type DataDirectoryResolver func() (string, error)

// DefaultDataDirectory follows the platform's conventional per-user data
// directory without creating it.
func DefaultDataDirectory(applicationName string) (string, error) {
	if applicationName == "" {
		return "", errors.New("application name is required")
	}

	var baseDirectory string
	var err error
	switch runtime.GOOS {
	case "windows":
		baseDirectory, err = os.UserConfigDir()
	case "darwin":
		var home string
		home, err = os.UserHomeDir()
		if err == nil {
			baseDirectory = filepath.Join(home, "Library", "Application Support")
		}
	default:
		baseDirectory = os.Getenv("XDG_DATA_HOME")
		if baseDirectory == "" {
			var home string
			home, err = os.UserHomeDir()
			if err == nil {
				baseDirectory = filepath.Join(home, ".local", "share")
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("resolve application data directory: %w", err)
	}
	return filepath.Join(baseDirectory, applicationName), nil
}

// SQLiteDatabasePath resolves the database path using an injectable data
// directory resolver, which keeps filesystem and platform concerns testable.
func SQLiteDatabasePath(resolveDataDirectory DataDirectoryResolver) (string, error) {
	if resolveDataDirectory == nil {
		return "", errors.New("data directory resolver is required")
	}
	directory, err := resolveDataDirectory()
	if err != nil {
		return "", fmt.Errorf("resolve SQLite directory: %w", err)
	}
	if directory == "" {
		return "", errors.New("resolved SQLite directory is empty")
	}
	return filepath.Join(directory, sqliteDatabaseFileName), nil
}

// OpenSQLite creates the database directory and opens a single-connection
// SQLite database configured for the application's local persistence.
func OpenSQLite(databasePath string) (*sql.DB, error) {
	if databasePath == "" {
		return nil, errors.New("SQLite database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create SQLite directory: %w", err)
	}

	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if _, err := database.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		database.Close()
		return nil, fmt.Errorf("configure SQLite busy timeout: %w", err)
	}
	if _, err := database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		database.Close()
		return nil, fmt.Errorf("configure SQLite foreign keys: %w", err)
	}
	if _, err := database.Exec("PRAGMA journal_mode = WAL"); err != nil {
		database.Close()
		return nil, fmt.Errorf("configure SQLite journal mode: %w", err)
	}
	return database, nil
}
