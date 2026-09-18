//go:build windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	windowsAutoStartRegistryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	windowsAutoStartValueName    = "SnippetsDome"
)

type windowsAutoStartManager struct {
	registryPath string
	valueName    string
	registry     windowsRegistryAPI
	executable   func() (string, error)
}

type windowsRegistryKey interface {
	Close() error
	DeleteValue(string) error
	SetStringValue(string, string) error
}

type windowsRegistryAPI interface {
	OpenKey(registry.Key, string, uint32) (windowsRegistryKey, error)
	CreateKey(registry.Key, string, uint32) (windowsRegistryKey, bool, error)
}

type systemWindowsRegistryAPI struct{}

func (systemWindowsRegistryAPI) OpenKey(root registry.Key, path string, access uint32) (windowsRegistryKey, error) {
	return registry.OpenKey(root, path, access)
}

func (systemWindowsRegistryAPI) CreateKey(root registry.Key, path string, access uint32) (windowsRegistryKey, bool, error) {
	return registry.CreateKey(root, path, access)
}

func newAutoStartManager() autoStartManager { return windowsAutoStartManager{} }

func (windowsAutoStartManager) isSupported() bool { return true }

func (manager windowsAutoStartManager) setEnabled(enabled bool) error {
	registryPath := manager.registryPath
	if registryPath == "" {
		registryPath = windowsAutoStartRegistryPath
	}
	valueName := manager.valueName
	if valueName == "" {
		valueName = windowsAutoStartValueName
	}
	api := manager.registry
	if api == nil {
		api = systemWindowsRegistryAPI{}
	}
	executablePath := manager.executable
	if executablePath == nil {
		executablePath = os.Executable
	}
	return setWindowsAutoStartEnabled(api, registryPath, valueName, enabled, executablePath)
}

func setWindowsAutoStartEnabled(api windowsRegistryAPI, registryPath, valueName string, enabled bool, executablePath func() (string, error)) error {
	if !enabled {
		key, err := api.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
		if err == registry.ErrNotExist {
			return nil
		}
		if err != nil {
			return err
		}
		defer key.Close()

		err = key.DeleteValue(valueName)
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}

	executable, err := executablePath()
	if err != nil {
		return err
	}
	key, _, err := api.CreateKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(valueName, windowsAutoStartCommand(executable))
}

func windowsAutoStartCommand(executable string) string {
	return fmt.Sprintf("\"%s\"", executable)
}
