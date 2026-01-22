package gui

import (
	"github.com/eka026/File-Format-Converter/internal/ports"
)

// GUIAdapter is the GUI driving adapter
type GUIAdapter struct {
	service  ports.IConversionService
	notifier ports.IProgressNotifier
}

// NewGUIAdapter creates a new GUI adapter
func NewGUIAdapter(service ports.IConversionService, notifier ports.IProgressNotifier) *GUIAdapter {
	return &GUIAdapter{
		service:  service,
		notifier: notifier,
	}
}

// HandleDrop handles file drop events
func (a *GUIAdapter) HandleDrop(files []string) {
	// Implementation will be added
}

// HandleConvertClick handles convert button click events
func (a *GUIAdapter) HandleConvertClick() {
	// Implementation will be added
}

// UpdateProgress updates the progress display
func (a *GUIAdapter) UpdateProgress(pct int) {
	// Implementation will be added
}

// ShowError shows an error message
func (a *GUIAdapter) ShowError(msg string) {
	// Implementation will be added
}

