package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app, err := NewApp()
	if err != nil {
		println("Error:", err.Error())
		return
	}
	app.tray.start()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  appTitle,
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:   &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		WindowStartState:   initialWindowStartState(os.Args[1:]),
		OnStartup:          app.startup,
		OnShutdown:         app.shutdown,
		OnBeforeClose:      app.beforeClose,
		SingleInstanceLock: singleInstanceLock(app),
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
