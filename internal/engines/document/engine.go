package document

import (
	"github.com/openconvert/file-converter/internal/domain"
	"github.com/openconvert/file-converter/internal/adapters/wasm"
)

// DocumentEngine implements IConverter for document conversions (Word → PDF)
type DocumentEngine struct {
	wasmBridge *WasmBridge
}

// NewDocumentEngine creates a new document conversion engine
func NewDocumentEngine(wasmBridge *WasmBridge) domain.IConverter {
	return &DocumentEngine{
		wasmBridge: wasmBridge,
	}
}

// Convert converts a Word document to PDF using pandoc.wasm
func (e *DocumentEngine) Convert(input, output string) error {
	// Read DOCX file
	docxData, err := readDOCX(input)
	if err != nil {
		return err
	}

	// Convert using pandoc.wasm via WasmBridge
	return e.wasmBridge.Convert(docxData, output)
}

// Validate checks if the input file is a valid Word document
func (e *DocumentEngine) Validate(file string) error {
	// Basic validation - check file extension and structure
	return validateDOCX(file)
}

