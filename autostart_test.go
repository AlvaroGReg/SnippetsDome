package main

import (
	"errors"
	"reflect"
	"testing"

	"SnippetsDome/internal/domain"
	"SnippetsDome/internal/service"
)

type autoStartManagerDouble struct {
	supported bool
	enabled   bool
	err       error
	calls     []bool
}

func (m *autoStartManagerDouble) isSupported() bool { return m.supported }

func (m *autoStartManagerDouble) setEnabled(enabled bool) error {
	m.calls = append(m.calls, enabled)
	if m.err != nil {
		return m.err
	}
	m.enabled = enabled
	return nil
}

type configRepositoryDouble struct {
	config domain.AppConfig
	err    error
}

func (r *configRepositoryDouble) SaveConfig(config domain.AppConfig) error {
	r.config = config
	return r.err
}

func TestAppSetStartAtLoginEnabled(t *testing.T) {
	t.Run("registers the launcher before persisting the preference", func(t *testing.T) {
		launcher := &autoStartManagerDouble{supported: true}
		configStore := &configRepositoryDouble{}
		app := &App{
			snippets:  service.NewSnippetService(domain.AppConfig{}, configStore),
			autoStart: launcher,
		}

		if err := app.SetStartAtLoginEnabled(true); err != nil {
			t.Fatalf("SetStartAtLoginEnabled() error = %v", err)
		}
		if !launcher.enabled || !app.GetStartAtLoginEnabled() || !configStore.config.StartAtLogin {
			t.Error("start-at-login preference was not enabled in both the launcher and configuration")
		}
	})

	t.Run("does not persist when launcher registration fails", func(t *testing.T) {
		launcher := &autoStartManagerDouble{supported: true, err: errors.New("registration failed")}
		app := &App{
			snippets:  service.NewSnippetService(domain.AppConfig{}, &configRepositoryDouble{}),
			autoStart: launcher,
		}

		if err := app.SetStartAtLoginEnabled(true); err == nil {
			t.Fatal("SetStartAtLoginEnabled() error = nil, want launcher error")
		}
		if app.GetStartAtLoginEnabled() {
			t.Error("StartAtLoginEnabled() = true after launcher failure, want false")
		}
	})

	t.Run("restores the launcher when configuration persistence fails", func(t *testing.T) {
		launcher := &autoStartManagerDouble{supported: true}
		app := &App{
			snippets:  service.NewSnippetService(domain.AppConfig{}, &configRepositoryDouble{err: errors.New("save config failed")}),
			autoStart: launcher,
		}

		if err := app.SetStartAtLoginEnabled(true); err == nil {
			t.Fatal("SetStartAtLoginEnabled() error = nil, want config save error")
		}
		if app.GetStartAtLoginEnabled() {
			t.Error("StartAtLoginEnabled() = true after config save failure, want false")
		}
		if launcher.enabled {
			t.Error("launcher remained enabled after config save failure")
		}
		if got, want := launcher.calls, []bool{true, false}; !reflect.DeepEqual(got, want) {
			t.Errorf("launcher calls = %v, want %v", got, want)
		}
	})

	t.Run("rejects unsupported operating systems", func(t *testing.T) {
		launcher := &autoStartManagerDouble{}
		app := &App{
			snippets:  service.NewSnippetService(domain.AppConfig{}, nil),
			autoStart: launcher,
		}

		if err := app.SetStartAtLoginEnabled(true); err == nil {
			t.Fatal("SetStartAtLoginEnabled() error = nil, want unsupported error")
		}
	})
}
