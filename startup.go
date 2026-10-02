package main

import "github.com/wailsapp/wails/v2/pkg/options"

const (
	appTitle      = "SnippetsDome"
	appInstanceID = appTitle
)

func initialWindowStartState(args []string) options.WindowStartState {
	if startsMinimized(args) {
		return options.Minimised
	}
	return options.Normal
}

func singleInstanceLock(app *App) *options.SingleInstanceLock {
	return &options.SingleInstanceLock{
		UniqueId: appInstanceID,
		OnSecondInstanceLaunch: func(options.SecondInstanceData) {
			app.showWindow()
		},
	}
}
