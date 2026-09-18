package main

import (
	"context"
	"errors"
	"log"
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
	Quit(context.Context)
}

type wailsRuntime struct{}

func (wailsRuntime) WindowHide(ctx context.Context) { runtime.WindowHide(ctx) }

func (wailsRuntime) Quit(ctx context.Context) { runtime.Quit(ctx) }

// NewApp creates a new App application struct
func NewApp() *App {
	configRepository := repository.NewJSONConfigRepository("SnippetsDome")
	config, err := configRepository.Load()
	if err != nil {
		log.Printf("unable to load snippets configuration: %v", err)
	}

	snippets := service.NewSnippetService(config, configRepository)
	app := &App{snippets: snippets, runtime: wailsRuntime{}, autoStart: newAutoStartManager()}
	app.tray = newTrayController(app)
	if config.StartAtLogin && app.autoStart.isSupported() {
		if err := app.autoStart.setEnabled(true); err != nil {
			log.Printf("unable to restore start-at-login setting: %v", err)
		}
	}
	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.snippets.SnippetsFilePath() == "" {
		return
	}
	if err := a.snippets.EnsureSnippetsFile(); err != nil {
		log.Printf("unable to create snippets file: %v", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.tray.stop()
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

// PickExistingSnippetsFile is a thin native-dialog bridge used by React.
func (a *App) PickExistingSnippetsFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Choose snippets file",
		Filters: []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}},
	})
}

// CreateSnippetsFile is a thin native-dialog bridge used by React.
func (a *App) CreateSnippetsFile() (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Create snippets file",
		DefaultFilename: "snippets.json",
		Filters:         []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}},
	})
}

func (a *App) SetSnippetsStoragePath(filePath string) (string, error) {
	return a.snippets.SetSnippetsFile(filePath)
}

func (a *App) GetSnippetsStoragePath() string {
	return a.snippets.SnippetsFilePath()
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

func (a *App) CreateSnippet(input domain.CreateSnippetInput) (domain.Snippet, error) {
	return a.snippets.CreateSnippet(input)
}

func (a *App) UpdateSnippet(snippet domain.Snippet) (domain.Snippet, error) {
	return a.snippets.UpdateSnippet(snippet)
}

func (a *App) DeleteSnippet(id string) error {
	return a.snippets.DeleteSnippet(id)
}

func (a *App) showWindow() {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
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
