//go:build linux || darwin

package main

import (
	_ "embed"
	"sync"

	"github.com/getlantern/systray"
)

// The PNG is supported by the Linux and macOS systray backends. macOS uses
// the regular icon here because the application icon has no template variant.
//
//go:embed build/appicon.png
var trayIcon []byte

type trayController struct {
	app   *App
	once  sync.Once
	mu    sync.Mutex
	items []*systray.MenuItem
}

func newTrayController(app *App) *trayController {
	return &trayController{app: app}
}

func (t *trayController) start() {
	t.once.Do(func() {
		go systray.Run(t.onReady, func() {})
	})
}

func (t *trayController) stop() {
	systray.Quit()
}

func (t *trayController) isSupported() bool { return true }

func (t *trayController) refresh() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, item := range t.items {
		item.Hide()
	}
	t.items = nil
	t.addSnippets()
}

func (t *trayController) addSnippets() {
	for _, snippet := range traySnippets(t.app) {
		item := systray.AddMenuItem("Copy: "+traySnippetTitle(snippet.Title), "Copy this snippet to the clipboard")
		t.items = append(t.items, item)
		go func(id string) {
			for range item.ClickedCh {
				t.app.copyTraySnippet(id)
			}
		}(snippet.ID)
	}
}

func (t *trayController) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTooltip("SnippetsDome")

	showItem := systray.AddMenuItem("Open SnippetsDome", "Show the application window")
	go func() {
		for range showItem.ClickedCh {
			t.app.showWindow()
		}
	}()

	systray.AddSeparator()
	t.addSnippets()

	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit SnippetsDome", "Close the application completely")
	go func() {
		for range quitItem.ClickedCh {
			t.app.quitFromTray()
		}
	}()
}
