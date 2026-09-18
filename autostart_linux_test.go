//go:build linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxAutoStartManagerCreatesAndRemovesDesktopEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "autostart", linuxAutoStartFileName)
	executable := "/opt/Snippets Dome/SnippetsDome"
	manager := linuxAutoStartManager{
		filePath: func() (string, error) { return path, nil },
		executable: func() (string, error) {
			return executable, nil
		},
	}

	if err := manager.setEnabled(true); err != nil {
		t.Fatalf("setEnabled(true) error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got, want := string(data), linuxDesktopEntry(executable); got != want {
		t.Errorf("desktop entry = %q, want %q", got, want)
	}

	if err := manager.setEnabled(false); err != nil {
		t.Fatalf("setEnabled(false) error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("desktop entry after disabling error = %v, want not-exist", err)
	}
}
