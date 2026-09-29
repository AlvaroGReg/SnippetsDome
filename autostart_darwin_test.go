//go:build darwin

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacOSAutoStartManagerCreatesAndRemovesLaunchAgent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "LaunchAgents", macOSAutoStartFileName)
	executable := `/Applications/Snippets & Dome/SnippetsDome.app/Contents/MacOS/SnippetsDome`
	manager := macOSAutoStartManager{
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
	content := string(data)
	if !strings.Contains(content, "<string>/Applications/Snippets &amp; Dome/SnippetsDome.app/Contents/MacOS/SnippetsDome</string>") {
		t.Errorf("launch agent does not contain the escaped executable path: %q", content)
	}
	if !strings.Contains(content, "<string>--start-minimized</string>") {
		t.Errorf("launch agent does not request a minimized startup: %q", content)
	}

	if err := manager.setEnabled(false); err != nil {
		t.Fatalf("setEnabled(false) error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("launch agent after disabling error = %v, want not-exist", err)
	}
}
