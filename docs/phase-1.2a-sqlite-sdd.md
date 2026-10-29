# SDD: Phase 1.2a - SQLite and Collections

## 1. Status and Scope

- Status: Approved for implementation
- Application: SnippetsDome, Wails v2
- Block: 1.2a - SQLite and collection management
- Related roadmap: `docs/phase-1.2-roadmap.md`

This document defines the architecture decisions and expected behavior for SQLite storage. The roadmap defines the order of work by session; detailed session tasks are not repeated here.

Phase 1.2a replaces JSON as the primary storage with a single local SQLite database containing multiple snippet collections and application preferences.

## 2. Out of Scope

- JSON import: phase 1.2b.
- JSON export: phase 1.2b.
- Automatic migration from the existing JSON file.
- Tray, autostart, and operational compatibility with Linux: phase 1.2c.
- New advanced search or tagging features.

JSON import and export will reuse this phase's data model but are not part of the SQLite bootstrap.

## 3. Initial State and Product Decisions

- A new installation creates an empty SQLite database.
- An empty initial collection is created so the application can start without content.
- The application informs users that the previous JSON file can be recovered through explicit import.
- The existing JSON file is not automatically read, modified, moved, or deleted.
- There is one SQLite database per user.
- The database contains multiple collections.
- Switching lists means changing the active collection within the same database.
- Users do not manually select the `.db` file in this phase.

Decisions finalized in `1.2a.0`:

- The database is named `snippets.db`.
- The initial collection is named `General`.
- Collection names are normalized with `strings.TrimSpace` and are unique case-insensitively.
- Collections can be renamed, but names cannot be empty or duplicate another collection.
- At least one collection must always exist; the last collection cannot be deleted.
- When the active collection is deleted, the remaining collection with the lowest `position` becomes active.
- Collection identifiers are UUID v4 strings.
- Tags are trimmed, empty tags are discarded, and duplicates are removed case-insensitively while preserving the first spelling.
- SQLite uses `journal_mode=WAL`, `busy_timeout=5000`, and at most one open connection (`SetMaxOpenConns(1)`).
- The initial schema version is `1`; the final indexes are `snippets(collection_id, position)` and `snippets(collection_id, favorite, position)`.
- Public errors retain stable, operation-oriented messages: `collection name is required`, `collection name already exists`, `collection not found`, `cannot delete the last collection`, and `snippet not found`.

## 4. Current Context

The current application has:

- preferences in `config.json` in the user's configuration directory;
- snippets in a JSON file selected by the user;
- a `SnippetService` coupled to `JSONSnippetRepository`;
- operations that read and rewrite the entire collection.

Phase 1.2a removes this coupling. The frontend and Wails layer must not know SQL statements or depend on a `.json` extension for primary storage.

## 5. Proposed Architecture

```text
Wails App
    |
    v
SnippetService
    |
    v
Repository interfaces
    |
    v
SQLite repository -> *sql.DB -> snippets.db
```

Separation rules:

- `internal/domain` contains models and domain rules.
- `internal/repository` contains SQLite and filesystem access.
- `internal/service` coordinates business operations and validation.
- `App` exposes Wails adapters and manages the lifecycle.
- The frontend only uses generated Wails methods.
- No persistence layer imports Wails packages.

## 6. SQLite Driver

Use `modernc.org/sqlite` through `database/sql`.

Reasons:

- Embedded SQLite.
- No CGO.
- Avoids requiring a C compiler or SQLite libraries installed on the system.
- Makes it easier to use the same Go code on Windows, Linux, and macOS.

The exact version must be pinned in `go.mod` during session `1.2a.1` and updated only through an explicit decision.

Do not use `github.com/mattn/go-sqlite3` in this phase because it introduces a CGO dependency into the build process.

## 7. Database Location

The database is stored in each operating system's standard data directory:

- Windows: `%APPDATA%/SnippetsDome/snippets.db`.
- Linux: `$XDG_DATA_HOME/SnippetsDome/snippets.db`.
- Linux without `XDG_DATA_HOME`: `~/.local/share/SnippetsDome/snippets.db`.

Path resolution must use `filepath.Join` and respect environment variables and system conventions. Path separators must not be hard-coded.

Cross-platform path implementation and operational validation belong to phase 1.2c. Phase 1.2a must use an isolated resolver behind an interface so it can be tested and replaced without changing the repository.

The database name must be confirmed in `1.2a.0`. This document uses `snippets.db` as the provisional name.

## 8. Data Model

The schema must support multiple lists without creating a separate database for each list.

### 8.1 Collections

```sql
CREATE TABLE collections (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0
);
```

Uniqueness is enforced by comparing `LOWER(name)` after trimming whitespace. The active collection is stored in `settings` under the `activeCollectionId` key.

### 8.2 Snippets

```sql
CREATE TABLE snippets (
    id            TEXT PRIMARY KEY,
    collection_id TEXT NOT NULL,
    title         TEXT NOT NULL,
    language      TEXT NOT NULL DEFAULT '',
    code          TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    favorite      INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
    position      INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
);
```

### 8.3 Tags

```sql
CREATE TABLE snippet_tags (
    snippet_id TEXT NOT NULL,
    tag        TEXT NOT NULL,
    PRIMARY KEY (snippet_id, tag),
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);
```

Empty tags are discarded, and duplicates are compared using `strings.ToLower` after trimming whitespace; the first non-empty spelling is preserved. This rule will also apply to future import/export operations.

### 8.4 Preferences

```sql
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

Current preferences will be preserved conceptually:

- `closeToTray` as a `0`/`1` boolean;
- `traySnippetLimit` as a decimal integer;
- `startAtLogin` as a `0`/`1` boolean;
- `language` as text;
- `activeCollectionId` as text.

`snippetsFilePath` will not be stored because the JSON file is no longer active storage and there will be no automatic migration.

### 8.5 Schema Version

```sql
CREATE TABLE schema_version (
    version INTEGER NOT NULL
);
```

Table creation and each subsequent change will use versioned, idempotent migrations. The initial schema must create the empty initial collection and store the schema version in a single transaction.

Final names, constraints, and indexes are fixed by the decisions in `1.2a.0` before implementing `1.2a.2`.

## 9. Repository Contract

The service must not depend on a concrete implementation. At minimum, the contract must cover:

- schema initialization and migration;
- closing the database;
- reading and writing preferences;
- creating, listing, renaming, and deleting collections;
- reading the contents of the active collection;
- inserting, updating, and deleting snippets;
- transactional replacement of tags;
- selecting the active collection.

Compound operations must be atomic. In particular:

- creating or updating a snippet and its tags;
- deleting a collection and its snippets;
- creating a collection and making it active;
- importing data in phase 1.2b.

The concrete repository API is defined once in `1.2a.0` and implemented in `1.2a.3`.

## 10. Concurrency and Lifecycle

- The application uses a single `*sql.DB` instance.
- `database/sql` manages the connection pool.
- The service protects in-memory state transitions with its existing mutex.
- Related writes are protected using SQLite transactions.
- SQLite policy is fixed to `journal_mode=WAL`, `busy_timeout=5000`, and at most one open connection.
- Startup opens SQLite and runs migrations before serving frontend calls.
- `shutdown` closes the connection and logs any close errors.
- JSON files are not created or initialized during startup.

## 11. Functional Compatibility

The phase must preserve existing snippet behavior:

- Generated and persisted IDs.
- `CreatedAt` is preserved during updates.
- Title and code are required.
- Current text field normalization.
- Persisted favorites.
- Persisted tags.
- Favorites first in the list.
- Language, tray, tray limit, and autostart preferences.

Snippet operations apply to the active collection. If no valid active collection exists, the service must return a controlled error or select the initial collection according to the decision in `1.2a.0`.

## 12. Frontend and Wails API

The Wails interface must expose operations to:

- get the active collection;
- list collections;
- create a collection;
- rename a collection;
- delete a collection according to the approved rules;
- select a collection;
- list and modify snippets in the active collection.

The frontend must:

- display the active collection;
- allow switching collections;
- display the initial empty state;
- inform users about future or available JSON import, depending on session 1.2b;
- stop treating a `.json` path as primary storage.

Generated bindings must not be edited manually. Regenerate them through the normal Wails workflow when needed.

## 13. Required Tests

### Bootstrap and Schema

- creates the directory and database at a temporary path;
- creates the initial schema;
- creates the initial collection;
- reopens the same database without duplicating tables or the initial collection;
- applies migrations only once;
- rolls back a failed migration.

### Collections

- creates a collection;
- lists collections in stable order;
- selects a collection;
- preserves the active collection after restart;
- applies the naming rules;
- applies the deletion rules;
- keeps snippets isolated between collections.

### Snippets and Tags

- persists all fields;
- creates, updates, and deletes snippets;
- preserves `CreatedAt` when updating;
- persists zero, one, and multiple tags;
- deletes tags when deleting a snippet;
- preserves favorites and position;
- rolls back partial operations.

### Preferences

- persists and retrieves each preference;
- persists the active collection;
- rolls back in-memory state when a write fails.

### Application

- initializes the repository once;
- closes the database during `shutdown`;
- does not attempt to read or migrate the old JSON file;
- displays an empty list in a new installation.

Required verification for 1.2a:

```bash
go test ./...
```

Frontend tests and the frontend build run in `1.2a.5`:

```bash
npm test
npm run build
```

## 14. Phase 1.2a Acceptance Criteria

1. A new installation creates an SQLite database and an empty collection.
2. The database is located in the system's standard data directory.
3. SQLite is the only primary data source.
4. The application supports multiple collections in one database.
5. Switching collections does not require moving or selecting `.db` files.
6. Snippets, tags, favorites, and preferences survive a restart.
7. The existing JSON file is never migrated automatically.
8. The interface informs users that JSON can be recovered through import.
9. The service does not depend on `JSONSnippetRepository`.
10. The persistence layer contains no Wails imports.
11. The SQLite driver does not require CGO.
12. `go test ./...` passes before the phase is handed over.

## 15. Decisions Finalized in Session 1.2a.0

These decisions must not be revisited in later sessions unless a technical blocker arises:

- final database name: `snippets.db`;
- initial collection name: `General`;
- unique names, case-insensitively, after trimming whitespace;
- renaming is allowed with a non-empty, non-duplicate name;
- the active collection can be deleted if another collection remains;
- after deleting the active collection, select the remaining collection with the lowest `position`;
- at least one collection must always exist;
- collection IDs are UUID v4;
- names and tags are trimmed; empty and case-insensitive duplicate tags are removed;
- duplicate tags are resolved by preserving the first spelling;
- `journal_mode=WAL`, `busy_timeout=5000`, and at most one open connection;
- initial schema version `1`, with indexes on collection/position and collection/favorite/position;
- stable public messages: `collection name is required`, `collection name already exists`, `collection not found`, `cannot delete the last collection`, and `snippet not found`.

Once finalized, these decisions must be recorded in this document and not duplicated in the documents for sessions `1.2a.1` through `1.2a.5`.
