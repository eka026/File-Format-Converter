package ports

import (
	"github.com/eka026/File-Format-Converter/internal/domain"
)

// IConverter defines the output port for conversion operations
type IConverter interface {
	// Convert performs the conversion from input to output format
	Convert(input []byte, outputFormat domain.Format) []byte
	
	// Validate checks if the input file is valid for this converter
	Validate(file string) domain.ValidationResult
	
	// GetSupportedInputTypes returns the input file types this converter supports
	GetSupportedInputTypes() []domain.FileType
	
	// GetSupportedOutputTypes returns the output formats this converter supports
	GetSupportedOutputTypes() []domain.Format
}

// IFileWriter defines the output port for file writing operations
type IFileWriter interface {
	// Write writes data to a file path
	Write(path string, data []byte) error
	
	// Read reads data from a file path
	Read(path string) ([]byte, error)
	
	// Exists checks if a file exists
	Exists(path string) bool
}

// ILogger defines the output port for logging operations
type ILogger interface {
	// Info logs an informational message
	Info(msg string)
	
	// Error logs an error message
	Error(msg string, err error)
	
	// Debug logs a debug message
	Debug(msg string)
}

// IWasmRuntime defines the output port for WebAssembly runtime operations
type IWasmRuntime interface {
	// Execute executes a WebAssembly module with input data
	Execute(wasm []byte, input []byte) []byte
	
	// IsSandboxed returns whether the runtime is sandboxed
	IsSandboxed() bool
}

// IPDFGenerator defines the output port for PDF generation operations
type IPDFGenerator interface {
	// GenerateFromHTML generates a PDF from HTML content
	GenerateFromHTML(html []byte) []byte
}

