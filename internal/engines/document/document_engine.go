package document

import (
	"context"
	"fmt"
	"os"

	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
)

// DocumentEngine implements IConverter for document conversions using pure Go
// Follows the same pattern as SpreadsheetEngine: Parse → HTML → PDF
type DocumentEngine struct {
	parser       *DocxParser
	htmlRenderer *HTMLRenderer
	pdfGenerator *browser.HeadlessBrowser
}

// NewDocumentEngine creates a new document conversion engine
// Uses pure Go DOCX parsing (no WASM, no CGO dependencies)
func NewDocumentEngine(pdfGenerator *browser.HeadlessBrowser) domain.IConverter {
	return &DocumentEngine{
		parser:       NewDocxParser(),
		htmlRenderer: NewHTMLRenderer(),
		pdfGenerator: pdfGenerator,
	}
}

// Convert converts a DOCX file to the specified output format
// Input and output are file paths (matches domain.IConverter interface)
func (e *DocumentEngine) Convert(input, output string) error {
	// Read DOCX file
	docxData, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("reading docx file: %w", err)
	}

	// Parse DOCX file
	doc, err := e.parser.Parse(docxData)
	if err != nil {
		return fmt.Errorf("parsing docx: %w", err)
	}

	// Convert to HTML
	htmlContent := e.htmlRenderer.Render(doc)

	// Determine output format from file extension
	outputExt := getFileExtension(output)
	if outputExt == ".html" || outputExt == ".htm" {
		// Write HTML directly
		return os.WriteFile(output, []byte(htmlContent), 0644)
	}

	// For PDF, use headless browser
	if outputExt == ".pdf" {
		if e.pdfGenerator == nil {
			return fmt.Errorf("pdf generator not available")
		}
		ctx := context.Background()
		return e.pdfGenerator.GeneratePDFFromHTML(ctx, htmlContent, output)
	}

	return fmt.Errorf("unsupported output format: %s", outputExt)
}

// getFileExtension extracts file extension in lowercase
func getFileExtension(filename string) string {
	ext := filename
	if len(filename) > 0 {
		for i := len(filename) - 1; i >= 0; i-- {
			if filename[i] == '.' {
				ext = filename[i:]
				break
			}
			if filename[i] == '/' || filename[i] == '\\' {
				break
			}
		}
	}
	// Convert to lowercase for comparison
	result := ""
	for _, r := range ext {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

// Validate checks if the input file is valid for this converter
func (e *DocumentEngine) Validate(file string) error {
	return validateDOCX(file)
}


