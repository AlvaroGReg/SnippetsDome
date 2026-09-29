package main

const startMinimizedArgument = "--start-minimized"

// autoStartManager creates and removes the operating-system entry that starts
// the current application for the signed-in user.
type autoStartManager interface {
	setEnabled(bool) error
	isSupported() bool
}

func startsMinimized(args []string) bool {
	for _, arg := range args {
		if arg == startMinimizedArgument {
			return true
		}
	}
	return false
}
