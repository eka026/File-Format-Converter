package image

import (
	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/ports"
)

// ImageEngineLegacy implements ports.IConverter for image format conversions (legacy/unused)
// NOTE: This is a legacy implementation that is not currently used.
// The active implementation is in engine.go which implements domain.IConverter
type ImageEngineLegacy struct {
	workerCount int
	workerPool  *WorkerPool
}

// NewImageEngineLegacy creates a new image conversion engine (legacy/unused)
func NewImageEngineLegacy(
	workerCount int,
	workerPool *WorkerPool,
) ports.IConverter {
	return &ImageEngineLegacy{
		workerCount: workerCount,
		workerPool:  workerPool,
	}
}

// Convert performs the conversion from input to output format
func (e *ImageEngineLegacy) Convert(input []byte, outputFormat domain.Format) []byte {
	// Implementation will be added
	return nil
}

// Validate checks if the input file is valid for this converter
func (e *ImageEngineLegacy) Validate(file string) domain.ValidationResult {
	// Implementation will be added
	return domain.ValidationResult{}
}

// GetSupportedInputTypes returns the input file types this converter supports
func (e *ImageEngineLegacy) GetSupportedInputTypes() []domain.FileType {
	// Implementation will be added
	return nil
}

// GetSupportedOutputTypes returns the output formats this converter supports
func (e *ImageEngineLegacy) GetSupportedOutputTypes() []domain.Format {
	// Implementation will be added
	return nil
}

// BatchProcess processes multiple images in batch
func (e *ImageEngineLegacy) BatchProcess(images [][]byte) [][]byte {
	// Implementation will be added
	return nil
}

// initWorkerPool initializes the worker pool
func (e *ImageEngineLegacy) initWorkerPool() *WorkerPool {
	// Implementation will be added
	return nil
}

