package main

// NFR-01 (Data Sovereignty): This application does not transmit any file data,
// metadata, or telemetry to external servers. All processing occurs locally.
//
// NFR-04 (Single Binary): This application is distributed as a single executable file.
// The Go binary is statically linked and embeds all frontend assets. No external runtimes
// (Python, JVM, Node.js) are required. Note: PDF generation requires Chrome/Chromium/Edge
// to be installed on the system, which is a system-level dependency, not a runtime dependency.

import (
	"embed"

	"github.com/eka026/File-Format-Converter/internal/adapters/gui"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:web
var assets embed.FS

func main() {
	app := gui.NewApp()

	err := wails.Run(&options.App{
		Title:  "File Format Converter",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnDomReady:       app.OnDomReady,
		OnShutdown:       app.OnShutdown,
		Bind:             []interface{}{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
