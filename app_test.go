package main

import (
	"context"
	"testing"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/service"
)

type wailsRuntimeDouble struct {
	hideCalls int
	quitCalls int
}

func (r *wailsRuntimeDouble) WindowHide(context.Context) { r.hideCalls++ }

func (r *wailsRuntimeDouble) Quit(context.Context) { r.quitCalls++ }

type trayDouble struct{ supported bool }

func (t trayDouble) start()            {}
func (t trayDouble) stop()             {}
func (t trayDouble) isSupported() bool { return t.supported }

func TestAppCloseToTrayLifecycle(t *testing.T) {
	runtime := &wailsRuntimeDouble{}
	app := &App{
		ctx:      context.Background(),
		snippets: service.NewSnippetService(domain.AppConfig{CloseToTray: true}, nil, nil),
		tray:     trayDouble{supported: true},
		runtime:  runtime,
	}

	if intercepted := app.beforeClose(app.ctx); !intercepted {
		t.Fatal("beforeClose() = false, want close intercepted")
	}
	if runtime.hideCalls != 1 {
		t.Fatalf("WindowHide calls = %d, want 1", runtime.hideCalls)
	}

	app.quitFromTray()
	if runtime.quitCalls != 1 {
		t.Fatalf("Quit calls = %d, want 1", runtime.quitCalls)
	}
	if intercepted := app.beforeClose(app.ctx); intercepted {
		t.Fatal("beforeClose() = true after tray exit, want shutdown to continue")
	}
	if runtime.hideCalls != 1 {
		t.Fatalf("WindowHide calls = %d after tray exit, want 1", runtime.hideCalls)
	}
}
