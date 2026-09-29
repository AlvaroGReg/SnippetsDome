package main

import "github.com/wailsapp/wails/v2/pkg/options"

func initialWindowStartState(args []string) options.WindowStartState {
	if startsMinimized(args) {
		return options.Minimised
	}
	return options.Normal
}
