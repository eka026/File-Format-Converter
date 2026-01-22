package progress

import (
	"github.com/openconvert/file-converter/internal/ports"
)

// ProgressNotifierAdapter implements the ProgressNotifier port
type ProgressNotifierAdapter struct {
	callbacks map[string]func(int)
}

// NewProgressNotifier creates a new progress notifier adapter
func NewProgressNotifier() ports.ProgressNotifier {
	return &ProgressNotifierAdapter{
		callbacks: make(map[string]func(int)),
	}
}

// NotifyProgress notifies about conversion progress
func (p *ProgressNotifierAdapter) NotifyProgress(file string, percentage int) {
	if callback, exists := p.callbacks[file]; exists {
		callback(percentage)
	}
}

// NotifyComplete notifies that conversion is complete
func (p *ProgressNotifierAdapter) NotifyComplete(file string) {
	p.NotifyProgress(file, 100)
}

// NotifyError notifies about conversion errors
func (p *ProgressNotifierAdapter) NotifyError(file string, err error) {
	// Implementation for error notification
}

// RegisterCallback registers a callback for progress updates
func (p *ProgressNotifierAdapter) RegisterCallback(file string, callback func(int)) {
	p.callbacks[file] = callback
}

