package domain

import (
	"context"
	"github.com/eka026/File-Format-Converter/internal/ports"
)

// ConverterService is the orchestrator that manages the conversion process
type ConverterService struct {
	converter    IConverter
	fileWriter   ports.FileWriter
	logger       ports.Logger
	progressNotifier ports.ProgressNotifier
}

// NewConverterService creates a new converter service
func NewConverterService(
	converter IConverter,
	fileWriter ports.FileWriter,
	logger ports.Logger,
	progressNotifier ports.ProgressNotifier,
) *ConverterService {
	return &ConverterService{
		converter:        converter,
		fileWriter:       fileWriter,
		logger:           logger,
		progressNotifier: progressNotifier,
	}
}

// Convert performs a single file conversion
func (s *ConverterService) Convert(ctx context.Context, source, target string) error {
	// Implementation will be added
	return nil
}

// BatchConvert performs batch file conversion
func (s *ConverterService) BatchConvert(ctx context.Context, files []string, targetFormat string) error {
	// Implementation will be added
	return nil
}

// GetSupportedFormats returns the list of supported conversion formats
func (s *ConverterService) GetSupportedFormats() []Format {
	// Implementation will be added
	return nil
}

