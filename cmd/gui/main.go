package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/openconvert/file-converter/internal/adapters/gui"
)

//go:embed all:../../web
var assets embed.FS

func main() {
	app := gui.NewApp()

	err := wails.Run(&options.App{
		Title:  "OpenConvert",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnDomReady:       app.OnDomReady,
		OnShutdown:       app.OnShutdown,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

