package progress

import (
	"fmt"

	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/ports"
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

// DomainProgressNotifierAdapter adapts to domain.ProgressNotifier interface
type DomainProgressNotifierAdapter struct {
	*ProgressNotifierAdapter
}

// NewDomainProgressNotifierAdapter creates a domain ProgressNotifier adapter
func NewDomainProgressNotifierAdapter() domain.ProgressNotifier {
	return &DomainProgressNotifierAdapter{
		ProgressNotifierAdapter: &ProgressNotifierAdapter{
			callbacks: make(map[string]func(int)),
		},
	}
}

// NotifyProgress notifies about conversion progress (ports interface)
func (p *ProgressNotifierAdapter) NotifyProgress(file string, percentage int) {
	if callback, exists := p.callbacks[file]; exists {
		callback(percentage)
	}
}

// NotifyProgress notifies about conversion progress (domain interface)
func (p *DomainProgressNotifierAdapter) NotifyProgress(pct int, msg string) {
	// Simple console output for now - can be enhanced with callbacks
	fmt.Printf("[PROGRESS] %d%%: %s\n", pct, msg)
}

// NotifyComplete notifies that conversion is complete (ports interface)
func (p *ProgressNotifierAdapter) NotifyComplete(file string) {
	p.NotifyProgress(file, 100)
}

// NotifyComplete notifies that conversion is complete (domain interface)
func (p *DomainProgressNotifierAdapter) NotifyComplete(result domain.Result) {
	if result.Success {
		fmt.Printf("[COMPLETE] Conversion successful: %s (took %v)\n", result.OutputPath, result.Duration)
	} else {
		fmt.Printf("[COMPLETE] Conversion failed: %v\n", result.Error)
	}
}

// NotifyError notifies about conversion errors (ports interface)
func (p *ProgressNotifierAdapter) NotifyError(file string, err error) {
	// Implementation for error notification
}

// NotifyError notifies about conversion errors (domain interface)
func (p *DomainProgressNotifierAdapter) NotifyError(err error) {
	fmt.Printf("[ERROR] Conversion error: %v\n", err)
}

// RegisterCallback registers a callback for progress updates
func (p *ProgressNotifierAdapter) RegisterCallback(file string, callback func(int)) {
	p.callbacks[file] = callback
}

