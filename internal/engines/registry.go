package engines

import (
	"github.com/eka026/File-Format-Converter/internal/domain"
)

// Registry manages conversion engines
type Registry struct {
	engines map[string]domain.IConverter
}

// NewRegistry creates a new engine registry
func NewRegistry(
	spreadsheetEngine domain.IConverter,
	imageEngine domain.IConverter,
	documentEngine domain.IConverter,
) *Registry {
	return &Registry{
		engines: map[string]domain.IConverter{
			"spreadsheet": spreadsheetEngine,
			"image":       imageEngine,
			"document":    documentEngine,
		},
	}
}

// GetEngine returns a conversion engine by type
func (r *Registry) GetEngine(engineType string) (domain.IConverter, bool) {
	engine, exists := r.engines[engineType]
	return engine, exists
}

