# Snippets Dome

<img src="build/appicon.png" alt="Snippets Dome icon" width="128">

A desktop application for saving, finding, and reusing code snippets without relying on an account or cloud service. It is built with Wails: Go handles the application logic and disk access, while React and TypeScript provide the user interface.

Snippets are stored locally in a SQLite database, so data remains private and persists after the application is closed. A single database can contain multiple snippet collections.

> A cross-platform desktop application built with Wails, Go, React, and TypeScript to manage code snippets locally with search, tags, and offline persistence.

## Screenshots

### Main view

![Snippet Dome main view](docs/screenshots/main-view.png)

The main screen brings together search, collection selection, the create-snippet action, the results list, and the theme controls.

### Create or edit a snippet

![Snippet editor](docs/screenshots/snippet-editor.png)

### Search and copy

![Searching and copying code](docs/screenshots/search-and-copy.png)

## Features

- Create, edit, and delete snippets.
- Store a title, language, code block, and comma-separated tags.
- Case-insensitive search across titles, languages, tags, and code content.
- Copy code to the clipboard from each snippet.
- Mark snippets as favorites so they stay at the top of the list.
- Confirmation before deletion and user-facing operation errors.
- Loading and empty-list states.
- Store snippets in one SQLite database with multiple collections.
- Create collections and switch the active collection without moving files.
- Keep snippets, tags, favorites, ordering, and application preferences across launches.
- Show how to recover an older JSON list through the explicit import flow planned for phase 1.2b.
- Switch between light and dark themes.
- Optional close-to-tray behavior on supported platforms, configured from the Settings dialog.
- Cross-platform system-tray menu on Windows, Linux, and macOS to open the app, copy snippets from the active collection, or quit completely.
- Start automatically with the operating system when supported by the platform integration.

## Technologies

- [Wails v2](https://wails.io/): desktop packaging and Go-to-frontend communication.
- [Go](https://go.dev/): domain, validation, services, and local persistence.
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite): SQLite driver without CGO.
- [React](https://react.dev/) and [TypeScript](https://www.typescriptlang.org/): user interface.
- [Vite](https://vite.dev/): frontend development environment and build tool.
- [Fluent UI React v9](https://react.fluentui.dev/): accessible components, icons, and themes.
- JSON: an explicit import/export format for phase 1.2b, not the primary storage.

## Architecture

There is no HTTP API or remote server. Wails generates bindings that allow React to call public methods on `App`, which delegate to Go application logic.

```text
React + TypeScript (frontend/src)
          │ generated Wails bindings
          ▼
app.go (application API)
          ▼
internal/service (use cases and validation)
          ▼
internal/repository (SQLite and filesystem access)
          ▼
one SQLite database: snippets.db
```

- `internal/domain/` defines the data handled by the application (`Snippet`, `Collection`, and configuration).
- `internal/service/` generates identifiers and timestamps, validates snippets, and coordinates collection and snippet operations.
- `internal/repository/` owns SQLite bootstrap, migrations, persistence, and platform data paths.
- `app.go` exposes the Wails API for collections, snippets, preferences, tray behavior, and autostart.
- `frontend/wailsjs/` contains generated code: use it from the frontend, but do not edit it manually.

## Requirements

- Go (the required version is defined in [go.mod](go.mod)).
- Node.js and npm.
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation/).

On Ubuntu 24.04 or WSL, native Wails development also requires `build-essential`, `pkg-config`, `libgtk-3-dev`, and `libwebkit2gtk-4.1-dev`.

## Run in development mode

Install frontend dependencies the first time:

```bash
cd frontend
npm install
cd ..
```

Then start the application with hot reload:

```bash
wails dev
```

On Ubuntu 24.04 / WSL, use the matching WebKitGTK tag:

```bash
wails dev -tags webkit2_41
```

At first launch, SnippetsDome creates an empty `General` collection in its SQLite database. The previous JSON list is not migrated automatically; it will be recovered through the explicit import flow from phase 1.2b.

## Build

First validate the frontend and backend:

```bash
cd frontend
npm run build
cd ..
go test ./...
```

Generate a distributable build with:

```bash
wails build
```

On Ubuntu 24.04 / WSL:

```bash
wails build -tags webkit2_41
```

Wails writes the generated binary to `build/bin/`. Its final format depends on the operating system used for the build.

## Local data

The application stores snippets, collections, and preferences in one SQLite database. The database is created in the standard per-user data directory:

| Platform | Location |
| --- | --- | --- |
| Windows | `%APPDATA%/SnippetsDome/snippets.db` |
| Linux | `$XDG_DATA_HOME/SnippetsDome/snippets.db`, or `~/.local/share/SnippetsDome/snippets.db` when `XDG_DATA_HOME` is not set |
| macOS | `~/Library/Application Support/SnippetsDome/snippets.db` |

Each snippet contains an identifier, title, language, code, tags, creation date, favorite status, and collection membership. JSON import and export are explicit phase 1.2b operations and do not replace SQLite as the primary store.

## Tests

The project includes focused Go tests for SQLite bootstrap, migrations, repositories, services, and application behavior, plus frontend tests with Vitest and Testing Library. The frontend suite covers search input behavior, rendering and actions in the snippets list (including copying), and tag normalization in the snippet editor.

Run the Go tests from the repository root:

```bash
go test ./...
```

Run the frontend suite from the `frontend` directory:

```bash
cd frontend
npm install
npm test
```

For interactive development, keep Vitest running in watch mode:

```bash
npm run test:watch
```

Before handing frontend changes over, also verify the production build:

```bash
npm run build
```

## Changelog

### 1.2 (in progress)

- **1.2a:** Replaced the JSON primary store with one SQLite database containing multiple collections, snippets, tags, and preferences. New installations start with an empty `General` collection.
- **1.2b:** Adds explicit JSON import and export so older lists can be recovered without automatic migration.
- **1.2c:** Completes Linux and macOS data paths, autostart, and system-tray behavior while preserving the existing Windows integration.
- The previous JSON file is never read, moved, modified, or deleted automatically.

### 1.1.0

- Replaced folder-based storage with a user-selected JSON file before the SQLite transition in version 1.2.
- Added the Settings dialog, Windows close-to-tray flow, Windows notification-area menu, and automatic startup when supported.

### 1.0.0

- Delivered the initial offline snippet manager with creation, editing, deletion, searching, copying, tags, and persisted local storage.
- Added focused Go and frontend test coverage, visual polish, screenshots, and a distributable desktop build.

## Roadmap

Version 1.2 is being delivered in three blocks:

- [X] **1.2a - SQLite and collections:** SQLite bootstrap, schema, migrations, repository, preferences, collection API, and frontend integration.
- [X] **1.2b - JSON import/export:** import an existing JSON list as a collection and export a selected collection without changing SQLite data.
- [ ] **1.2c - Linux and macOS compatibility:** standard data paths, autostart, system tray, and platform-specific fallbacks.
- [X] Launch SnippetsDome automatically when the operating system starts, minimized when supported on implemented platforms.
- [ ] Quick, combinable tag filters in a collapsible sidebar.
- [X] Favorites, predictably sorted and accessible without opening the editor.
- [X] Cross-platform system-tray integration for quick access to five snippets.
- [ ] Syntax highlighting, a `Ctrl/Cmd + K` shortcut, and snippet-to-file export.
- [X] Localization in English and Spanish, selecting system by default, configurable from Settings.
- [ ] Configurable system command to save selected text as a snippet with a generic title.
- [ ] Configurable command for pasting a snippet without interacting with the application.
- [ ] Priority system: assign a number to an asset that makes it appear 1st or last.

## License

No license has been defined for this repository yet.
