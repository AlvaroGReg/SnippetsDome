package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"path/filepath"
	"sync"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/repository"
	"SnippetsDome/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	snippets      *service.SnippetService
	database      *sql.DB
	tray          appTray
	runtime       appRuntime
	autoStart     autoStartManager
	quitMu        sync.RWMutex
	quitRequested bool
}

type appTray interface {
	start()
	stop()
	isSupported() bool
}

type appRuntime interface {
	WindowHide(context.Context)
	WindowUnminimise(context.Context)
	WindowShow(context.Context)
	Quit(context.Context)
}

type wailsRuntime struct{}

func (wailsRuntime) WindowHide(ctx context.Context) { runtime.WindowHide(ctx) }

func (wailsRuntime) WindowUnminimise(ctx context.Context) { runtime.WindowUnminimise(ctx) }

func (wailsRuntime) WindowShow(ctx context.Context) { runtime.WindowShow(ctx) }

func (wailsRuntime) Quit(ctx context.Context) { runtime.Quit(ctx) }

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	databasePath, err := repository.SQLiteDatabasePath(func() (string, error) {
		return repository.DefaultDataDirectory("SnippetsDome")
	})
	if err != nil {
		return nil, err
	}
	database, err := repository.OpenSQLite(databasePath)
	if err != nil {
		return nil, err
	}
	if err := repository.MigrateSQLite(database); err != nil {
		_ = database.Close()
		return nil, err
	}
	sqliteRepository, err := repository.NewSQLiteSnippetRepository(database)
	if err != nil {
		_ = database.Close()
		return nil, err
	}
	config, err := sqliteRepository.LoadConfig()
	if err != nil {
		_ = database.Close()
		return nil, err
	}
	snippets := service.NewSnippetService(config, sqliteRepository, sqliteRepository)
	app := &App{snippets: snippets, database: database, runtime: wailsRuntime{}, autoStart: newAutoStartManager()}
	app.tray = newTrayController(app)
	if config.StartAtLogin && app.autoStart.isSupported() {
		if err := app.autoStart.setEnabled(true); err != nil {
			log.Printf("unable to restore start-at-login setting: %v", err)
		}
	}
	return app, nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.tray.stop()
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			log.Printf("unable to close SQLite database: %v", err)
		}
	}
}

// beforeClose hides the window only when the user has enabled close to tray.
// Explicit exits from the tray always continue with the application shutdown.
func (a *App) beforeClose(ctx context.Context) bool {
	a.quitMu.RLock()
	quitRequested := a.quitRequested
	a.quitMu.RUnlock()

	if quitRequested || !a.tray.isSupported() || !a.snippets.CloseToTrayEnabled() {
		return false
	}

	a.runtime.WindowHide(ctx)
	return true
}

func (a *App) GetSnippets() ([]domain.Snippet, error) {
	return a.snippets.List()
}

func (a *App) GetCloseToTrayEnabled() bool {
	return a.snippets.CloseToTrayEnabled()
}

func (a *App) SetCloseToTrayEnabled(enabled bool) error {
	return a.snippets.SetCloseToTrayEnabled(enabled)
}

func (a *App) GetStartAtLoginSupported() bool {
	return a.autoStart.isSupported()
}

func (a *App) GetStartAtLoginEnabled() bool {
	return a.snippets.StartAtLoginEnabled()
}

func (a *App) SetStartAtLoginEnabled(enabled bool) error {
	if !a.autoStart.isSupported() {
		return errors.New("start at login is not supported on this operating system")
	}

	previousEnabled := a.snippets.StartAtLoginEnabled()
	if err := a.autoStart.setEnabled(enabled); err != nil {
		return err
	}
	if err := a.snippets.SetStartAtLoginEnabled(enabled); err != nil {
		if rollbackErr := a.autoStart.setEnabled(previousEnabled); rollbackErr != nil {
			log.Printf("unable to restore start-at-login launcher after config save failure: %v", rollbackErr)
		}
		return err
	}
	return nil
}

func (a *App) GetTraySnippetLimit() int {
	return a.snippets.TraySnippetLimit()
}

func (a *App) SetTraySnippetLimit(limit int) error {
	return a.snippets.SetTraySnippetLimit(limit)
}

func (a *App) GetLanguage() string {
	return a.snippets.Language()
}

func (a *App) SetLanguage(language string) error {
	return a.snippets.SetLanguage(language)
}

func (a *App) CreateSnippet(input domain.CreateSnippetInput) (domain.Snippet, error) {
	return a.snippets.CreateSnippet(input)
}

func (a *App) UpdateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
	return a.snippets.UpdateSnippet(snippet)
}

func (a *App) DeleteSnippet(id string) error {
	return a.snippets.DeleteSnippet(id)
}

func (a *App) GetActiveCollection() (domain.Collection, error) {
	return a.snippets.ActiveCollection()
}

func (a *App) GetCollections() ([]domain.Collection, error) {
	return a.snippets.ListCollections()
}

func (a *App) CreateCollection(name string) (domain.Collection, error) {
	return a.snippets.CreateCollection(name)
}

func (a *App) RenameCollection(id, name string) error {
	return a.snippets.RenameCollection(id, name)
}

func (a *App) DeleteCollection(id string) error {
	return a.snippets.DeleteCollection(id)
}

func (a *App) SelectCollection(id string) error {
	return a.snippets.SelectCollection(id)
}

func (a *App) ImportJSON() (domain.Collection, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Import snippets JSON", Filters: []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return domain.Collection{}, err
	}
	return a.snippets.ImportJSON(path, filepath.Base(path[:len(path)-len(filepath.Ext(path))]))
}

func (a *App) ExportJSON() error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "Export snippets JSON", DefaultFilename: "snippets.json", Filters: []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return err
	}
	return a.snippets.ExportJSON(path)
}

func (a *App) showWindow() {
	if a.ctx != nil {
		a.runtime.WindowUnminimise(a.ctx)
		a.runtime.WindowShow(a.ctx)
		activateMainWindow()
	}
}

func (a *App) copySnippetToClipboard(code string) {
	if a.ctx == nil {
		return
	}
	if err := runtime.ClipboardSetText(a.ctx, code); err != nil {
		log.Printf("unable to copy snippet from tray: %v", err)
	}
}

func (a *App) quitFromTray() {
	a.quitMu.Lock()
	a.quitRequested = true
	a.quitMu.Unlock()

	if a.ctx != nil {
		a.runtime.Quit(a.ctx)
	}
}
