package ports

import (
	"github.com/eka026/File-Format-Converter/internal/domain"
)

// IConversionService defines the input port for conversion operations
type IConversionService interface {
	// Convert performs a single file conversion
	Convert(source, target string) domain.Result
	
	// BatchConvert performs batch file conversion
	BatchConvert(files []string, target string) []domain.Result
	
	// GetSupportedFormats returns the list of supported conversion formats
	GetSupportedFormats() []domain.Format
	
	// ValidateFile validates if a file can be converted
	ValidateFile(file string) domain.ValidationResult
}

// IProgressNotifier defines the input port for progress notifications
type IProgressNotifier interface {
	// NotifyProgress notifies about conversion progress
	NotifyProgress(pct int, msg string)
	
	// NotifyComplete notifies about conversion completion
	NotifyComplete(result domain.Result)
	
	// NotifyError notifies about conversion errors
	NotifyError(err error)
}

