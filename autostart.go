package main

// autoStartManager creates and removes the operating-system entry that starts
// the current application for the signed-in user.
type autoStartManager interface {
	setEnabled(bool) error
	isSupported() bool
}
