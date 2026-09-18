//go:build windows

package main

import (
	"errors"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestWindowsAutoStartCommand(t *testing.T) {
	got := windowsAutoStartCommand(`C:\Program Files\SnippetsDome\SnippetsDome.exe`)
	want := `"C:\Program Files\SnippetsDome\SnippetsDome.exe"`
	if got != want {
		t.Errorf("windowsAutoStartCommand() = %q, want %q", got, want)
	}
}

func TestWindowsAutoStartManagerCreatesAndRemovesRegistryValue(t *testing.T) {
	registryPath := `Software\SnippetsDomeTest`
	valueName := "SnippetsDomeTest"
	api := &fakeWindowsRegistryAPI{keys: make(map[string]*fakeWindowsRegistryKey)}
	manager := windowsAutoStartManager{
		registryPath: registryPath,
		valueName:    valueName,
		registry:     api,
		executable: func() (string, error) {
			return "test.exe", nil
		},
	}
	if err := manager.setEnabled(true); err != nil {
		t.Fatalf("setEnabled(true) error = %v", err)
	}
	if got, want := api.keys[registryPath].values[valueName], windowsAutoStartCommand("test.exe"); got != want {
		t.Errorf("registry command = %q, want %q", got, want)
	}

	if err := manager.setEnabled(false); err != nil {
		t.Fatalf("setEnabled(false) error = %v", err)
	}
	if _, ok := api.keys[registryPath].values[valueName]; ok {
		t.Error("registry value still exists after disabling")
	}
}

type fakeWindowsRegistryAPI struct {
	keys map[string]*fakeWindowsRegistryKey
}

func (api *fakeWindowsRegistryAPI) OpenKey(_ registry.Key, path string, _ uint32) (windowsRegistryKey, error) {
	key, ok := api.keys[path]
	if !ok {
		return nil, registry.ErrNotExist
	}
	return key, nil
}

func (api *fakeWindowsRegistryAPI) CreateKey(_ registry.Key, path string, _ uint32) (windowsRegistryKey, bool, error) {
	if key, ok := api.keys[path]; ok {
		return key, true, nil
	}
	key := &fakeWindowsRegistryKey{values: make(map[string]string)}
	api.keys[path] = key
	return key, false, nil
}

type fakeWindowsRegistryKey struct {
	values map[string]string
}

func (key *fakeWindowsRegistryKey) Close() error { return nil }

func (key *fakeWindowsRegistryKey) DeleteValue(name string) error {
	if _, ok := key.values[name]; !ok {
		return registry.ErrNotExist
	}
	delete(key.values, name)
	return nil
}

func (key *fakeWindowsRegistryKey) SetStringValue(name, value string) error {
	if value == "" {
		return errors.New("registry value cannot be empty")
	}
	key.values[name] = value
	return nil
}
