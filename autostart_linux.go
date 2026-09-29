//go:build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const linuxAutoStartFileName = "snippetsdome.desktop"

type linuxAutoStartManager struct {
	filePath   func() (string, error)
	executable func() (string, error)
}

func newAutoStartManager() autoStartManager { return linuxAutoStartManager{} }

func (linuxAutoStartManager) isSupported() bool { return true }

func (manager linuxAutoStartManager) setEnabled(enabled bool) error {
	filePath := manager.filePath
	if filePath == nil {
		filePath = linuxAutoStartFilePath
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
	return os.WriteFile(path, []byte(linuxDesktopEntry(executable)), 0o644)
}

func linuxAutoStartFilePath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "autostart", linuxAutoStartFileName), nil
}

func linuxDesktopEntry(executable string) string {
	return fmt.Sprintf("[Desktop Entry]\nType=Application\nName=SnippetsDome\nExec=%s %s\nTerminal=false\nX-GNOME-Autostart-enabled=true\n", quoteDesktopExecArgument(executable), startMinimizedArgument)
}

func quoteDesktopExecArgument(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "`", "\\`", "$", "\\$")
	return "\"" + replacer.Replace(value) + "\""
}
