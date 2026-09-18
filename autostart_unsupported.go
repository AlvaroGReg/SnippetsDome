//go:build !windows && !linux && !darwin

package main

import "errors"

type unsupportedAutoStartManager struct{}

func newAutoStartManager() autoStartManager { return unsupportedAutoStartManager{} }

func (unsupportedAutoStartManager) isSupported() bool { return false }

func (unsupportedAutoStartManager) setEnabled(bool) error {
	return errors.New("start at login is not supported on this operating system")
}
