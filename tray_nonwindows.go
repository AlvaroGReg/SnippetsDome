//go:build !windows && !linux && !darwin

package main

// trayController keeps unsupported operating-system builds functional without
// claiming that a notification-area integration is available.
type trayController struct{}

func newTrayController(app *App) *trayController {
	return &trayController{}
}

func (t *trayController) start() {}

func (t *trayController) stop() {}

func (t *trayController) isSupported() bool { return false }
