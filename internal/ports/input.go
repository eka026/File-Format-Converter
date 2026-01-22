package ports

import (
	"context"
	"github.com/eka026/File-Format-Converter/internal/domain"
)

// ConvertPort defines the input port for single file conversion
type ConvertPort interface {
	Convert(ctx context.Context, source, target string) error
}

// BatchConvertPort defines the input port for batch file conversion
type BatchConvertPort interface {
	BatchConvert(ctx context.Context, files []string, targetFormat string) error
}

// GetSupportedFormatsPort defines the input port for querying supported formats
type GetSupportedFormatsPort interface {
	GetSupportedFormats() []domain.Format
}

