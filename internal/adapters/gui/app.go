package gui

import (
	"context"
)

// App represents the GUI application adapter
type App struct {
	ctx context.Context
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the application starts
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// OnDomReady is called when the DOM is ready
func (a *App) OnDomReady(ctx context.Context) {
	// Initialize frontend components
}

// OnShutdown is called when the application shuts down
func (a *App) OnShutdown(ctx context.Context) {
	// Cleanup resources
}

// ConvertFile handles file conversion from the GUI
func (a *App) ConvertFile(source, target string) error {
	// Implementation will connect to domain core
	return nil
}

// BatchConvertFiles handles batch file conversion from the GUI
func (a *App) BatchConvertFiles(files []string, targetFormat string) error {
	// Implementation will connect to domain core
	return nil
}

// GetSupportedFormats returns supported formats for the GUI
func (a *App) GetSupportedFormats() []string {
	// Implementation will connect to domain core
	return nil
}

