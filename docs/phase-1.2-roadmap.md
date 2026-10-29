# Agentic Session Roadmap: Version 1.2

## Overall Goal

Version 1.2 will replace JSON as the primary storage system with SQLite, add JSON import and export, and complete operational compatibility with Linux.

Implementation is divided into three independent, ordered blocks:

1. **1.2a - SQLite and collection management**
2. **1.2b - JSON import and export**
3. **1.2c - Linux compatibility**

Each session should keep its scope small, run its tests, and document any decisions or risks discovered.

## Progress

- [x] `1.2a.0` - Finalize the design and approve the SDD.
- [x] `1.2a.1` - Resolve paths and bootstrap SQLite.
- [x] `1.2a.2` - SQLite schema and migrations.
- [x] `1.2a.3` - SQLite repository.
- [x] `1.2a.4` - Service and configuration.
- [x] `1.2a.5` - Wails and frontend integration.
- [x] `1.2b` - JSON import and export.
- [x] `1.2c` - Linux compatibility.

## Product Decisions

- SQLite is the application's primary storage.
- The application uses a single SQLite database per user.
- The database is stored in the operating system's standard data directory.
- The database contains multiple snippet collections.
- Switching lists means changing the active collection, not manually moving files.
- The **Manage lists** window allows users to create, rename, and delete lists, set the default list, and import/export JSON.
- The last remaining list cannot be deleted.
- Existing JSON data is not migrated automatically.
- A new installation starts with an empty collection.
- The application informs users that the previous list can be recovered by importing its JSON file.
- JSON import and export are explicit features of version 1.2b.

## Shared Session Rules

- Review the repository state before editing.
- Do not revert changes made by others.
- Keep SQL and persistence logic outside Wails and the frontend.
- Prefer small, verifiable changes.
- Add or update tests alongside each behavior change.
- Run `go test ./...` after each session that changes Go code.
- Run `npm test` and `npm run build` after each session that changes the frontend.
- Do not use `wails build` or `wails dev` for automated verification.
- Do not create commits automatically.

## Overall Dependencies

```text
1.2a.0 -> 1.2a.1 -> 1.2a.2 -> 1.2a.3 -> 1.2a.4 -> 1.2a.5
                                                        |
                                                        v
1.2b.0 -> 1.2b.1 -> 1.2b.2 -> 1.2b.3
                                  |
                                  v
1.2c.0 -> 1.2c.1 -> 1.2c.2 -> 1.2c.3
```

# 1.2a - SQLite and Collections

## 1.2a.0 - Finalize the Design

### Goal

Finalize the SDD before implementing SQLite persistence.

### Scope

- Confirm a single SQLite database with multiple collections.
- Define the schema for collections, snippets, tags, and preferences.
- Define the data location for each operating system.
- Confirm that JSON is not migrated automatically.
- Define the repository contract.
- Define the empty state and initial collection.

### Deliverables

- Updated and approved SDD.
- Finalized data schema.
- Clear acceptance criteria.

### Verification

Documentation review. No code changes required.

## 1.2a.1 - Path Resolution and SQLite Bootstrap

### Goal

Create the minimum infrastructure required to locate and open the database.

### Scope

- Resolve the SnippetsDome data directory.
- Create the directory if it does not exist.
- Open and close the SQLite connection.
- Configure basic pragmas.
- Return clear errors for unusable paths.

### Out of Scope

- Collections.
- Snippets.
- JSON import and export.
- Operational compatibility with Linux.

### Verification

```bash
go test ./...
```

Use `t.TempDir()` to isolate filesystem tests.

## 1.2a.2 - SQLite Schema and Migrations

### Goal

Implement the versioned database schema.

### Scope

- Schema version table.
- Collections table.
- Snippets table.
- Tags table.
- Preferences table.
- Foreign keys and indexes.
- Idempotent migrations.

### Verification

- Create a new database.
- Reopen an existing database.
- Confirm that a migration is not repeated.
- Verify rollback when a migration fails.

```bash
go test ./...
```

## 1.2a.3 - SQLite Repository

### Goal

Implement data access without Wails dependencies.

### Scope

- Create, list, select, and delete collections.
- Snippet CRUD.
- Persist tags in a related table.
- Persist favorites and stable ordering.
- Run compound operations in transactions.
- Isolate snippets by collection.

### Verification

- Full persistence of all fields.
- Multiple independent collections.
- Switching the active collection.
- Cascade deletion of tags.
- Rollback of partial operations.

```bash
go test ./...
```

## 1.2a.4 - Service and Configuration

### Goal

Remove the service's dependency on the JSON repository.

### Scope

- Replace `JSONSnippetRepository` with a repository interface.
- Persist preferences in SQLite.
- Preserve current validation.
- Preserve favorite ordering.
- Preserve preference rollback.
- Add the concept of an active collection.
- Adapt existing service tests.

### Verification

Cover at least:

- Create a snippet.
- Edit a snippet.
- Delete a snippet.
- List snippets.
- Mark favorites.
- Switch collections.
- Persist preferences.
- Handle persistence errors.

```bash
go test ./...
```

## 1.2a.5 - Wails and Frontend Integration

### Goal

Expose SQLite and collections to the complete application.

### Backend Scope

- Initialize SQLite from `NewApp`.
- Close the connection during `shutdown`.
- Expose Wails methods for collections.
- Remove the option to select a JSON file as the primary storage.
- Propagate errors to the frontend.

### Frontend Scope

- Display the active collection.
- Switch collections.
- Create a collection.
- Rename and delete collections while always preserving at least one.
- Set the default collection.
- Display the initial empty state.
- Handle JSON import and export from the list management window.
- Remove validations that require a `.json` extension for primary storage.

### Verification

Backend:

```bash
go test ./...
```

Frontend:

```bash
npm test
npm run build
```

# 1.2b - JSON Import and Export

## 1.2b.0 - JSON Exchange Contract

### Goal

Define import and export behavior before implementing dialogs and filesystem operations.

### Scope

- Preserve compatibility with the current JSON format.
- Import a collection from a JSON file.
- Export a specific collection.
- Decide how to handle duplicate IDs.
- Validate required fields.
- Preserve IDs, dates, favorites, languages, code, and tags.
- Define messages for invalid or incomplete JSON.

### Deliverables

- Documented JSON contract.
- Defined edge cases.
- Prepared behavior tests.

## 1.2b.1 - JSON Importer

### Goal

Import a JSON file as a new SQLite collection.

### Scope

- Read a selected file.
- Validate its structure.
- Create a new collection.
- Insert snippets and tags in a transaction.
- Select the imported collection on success.
- Leave the state unchanged if import fails.

### Verification

- Valid JSON.
- Empty JSON.
- Invalid JSON.
- Missing required fields.
- Empty and multiple tags.
- Favorites and dates.
- Full rollback on error.

```bash
go test ./...
```

## 1.2b.2 - JSON Exporter

### Goal

Export the active collection to a JSON file.

### Scope

- Export only the selected collection.
- Use the native save dialog.
- Maintain a stable JSON format.
- Do not modify internal data.
- Handle write errors.

### Verification

- Export a collection containing data.
- Export an empty collection.
- Preserve all fields exactly.
- Handle an unwritable destination.

```bash
go test ./...
```

## 1.2b.3 - Import and Export UX

### Goal

Connect import and export flows to React.

### Scope

- Import button.
- Export button.
- Loading indicators.
- Success and error messages.
- Reload the list after importing.
- Automatically select the imported collection.
- Interaction tests.

### Verification

```bash
npm test
npm run build
```

# 1.2c - Linux Compatibility

## 1.2c.0 - Cross-Platform Path Resolution

### Goal

Validate and complete data locations according to each system's conventions.

### Scope

- Windows: `%APPDATA%/SnippetsDome`.
- Linux: `$XDG_DATA_HOME/SnippetsDome`.
- Linux without `XDG_DATA_HOME`: `~/.local/share/SnippetsDome`.
- Create directories.
- Use portable path separators.
- Handle permissions.
- Avoid hard-coded paths.

### Verification

- Path resolution tests.
- Tests conditioned on `runtime.GOOS`.
- Environment variable tests.

```bash
go test ./...
```

## 1.2c.1 - Linux Autostart

### Goal

Complete automatic startup on Linux.

### Scope

- Linux desktop entry.
- Enable and disable autostart.
- Start minimized when supported.
- Clean up disabled configurations.
- Return clear errors when the platform does not allow it.

### Verification

- Tests for generated file contents.
- Tests for platform-specific paths.
- Tests for unsupported platforms.

```bash
go test ./...
```

## 1.2c.2 - Cross-Platform System Tray

### Goal

Complete system tray behavior on Windows and Linux with the same functional menu.

### Scope

- Show and hide the window.
- Display snippets from the active collection.
- Copy snippets from the tray.
- Explicitly quit from the tray.
- Preserve close-to-tray behavior.
- Keep a no-tray fallback for operating systems unsupported by the dependency.

### Verification

- Tests for controller selection by platform.
- Tests for quit behavior.
- Fallback tests.

```bash
go test ./...
```

## 1.2c.3 - Version 1.2 Integration and Completion

### Goal

Validate version 1.2 end to end.

### Minimum Cases

- Start with an empty database.
- Create a collection.
- Switch between collections.
- Create, edit, and delete snippets.
- Import JSON.
- Export JSON.
- Persist preferences.
- Autostart.
- System tray.
- Permission errors.
- Restart the application without data loss.

### Final Verification

From the repository root:

```bash
go test ./...
```

From `frontend/`:

```bash
npm test
npm run build
```

Interactive verification on each operating system will be performed manually. `wails build` and `wails dev` will not be used as substitutes for automated tests.

## Version 1.2 Completion Criteria

- SQLite is the only primary data source.
- The application supports multiple collections.
- A new installation starts empty and explains how to import JSON.
- JSON import and export work without data loss.
- Linux uses an appropriate data location.
- Autostart and the system tray have platform-specific behavior defined.
- Go tests pass with `go test ./...`.
- Frontend tests and the frontend build pass.
- Storage and usage documentation is up to date.
- No references remain that describe JSON as the primary storage.
