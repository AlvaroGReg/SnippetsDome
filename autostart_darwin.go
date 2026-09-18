//go:build darwin

package main

import (
	"bytes"
	"os"
	"path/filepath"
)

const macOSAutoStartFileName = "com.snippetsdome.app.plist"

type macOSAutoStartManager struct {
	filePath   func() (string, error)
	executable func() (string, error)
}

func newAutoStartManager() autoStartManager { return macOSAutoStartManager{} }

func (macOSAutoStartManager) isSupported() bool { return true }

func (manager macOSAutoStartManager) setEnabled(enabled bool) error {
	filePath := manager.filePath
	if filePath == nil {
		filePath = macOSAutoStartFilePath
	}
	executablePath := manager.executable
	if executablePath == nil {
		executablePath = os.Executable
	}

	path, err := filePath()
	if err != nil {
		return err
	}
	if !enabled {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	executable, err := executablePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, macOSLaunchAgent(executable), 0o644)
}

func macOSAutoStartFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", macOSAutoStartFileName), nil
}

func macOSLaunchAgent(executable string) []byte {
	var escaped bytes.Buffer
	for _, r := range executable {
		switch r {
		case '&':
			escaped.WriteString("&amp;")
		case '<':
			escaped.WriteString("&lt;")
		case '>':
			escaped.WriteString("&gt;")
		case '"':
			escaped.WriteString("&quot;")
		case '\'':
			escaped.WriteString("&apos;")
		default:
			escaped.WriteRune(r)
		}
	}

	return []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n<plist version=\"1.0\">\n<dict>\n  <key>Label</key>\n  <string>com.snippetsdome.app</string>\n  <key>ProgramArguments</key>\n  <array>\n    <string>" + escaped.String() + "</string>\n  </array>\n  <key>RunAtLoad</key>\n  <true/>\n</dict>\n</plist>\n")
}
