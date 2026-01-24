package spreadsheet

import (
	"fmt"

	"github.com/eka026/File-Format-Converter/internal/domain"
)

// SpreadsheetEngine implements IConverter for spreadsheet conversions (Excel → PDF)
type SpreadsheetEngine struct {
	parser       *ExcelParser
	htmlRenderer *HTMLRenderer
	pdfGenerator *PDFGenerator
}

// NewSpreadsheetEngine creates a new spreadsheet conversion engine
func NewSpreadsheetEngine(htmlRenderer *HTMLRenderer, pdfGenerator *PDFGenerator) domain.IConverter {
	return &SpreadsheetEngine{
		parser:       NewExcelParser(),
		htmlRenderer: htmlRenderer,
		pdfGenerator: pdfGenerator,
	}
}

// Convert converts an Excel file to PDF
func (e *SpreadsheetEngine) Convert(input, output string) error {
	// Parse Excel file using excelize via our parser
	f, err := e.parser.Parse(input)
	if err != nil {
		return fmt.Errorf("parsing excel file: %w", err)
	}
	defer f.Close()

	// Convert to HTML (internally parses workbook data)
	htmlContent, err := e.htmlRenderer.Render(f)
	if err != nil {
		return fmt.Errorf("rendering html: %w", err)
	}

	// Generate PDF from HTML
	if err := e.pdfGenerator.Generate(htmlContent, output); err != nil {
		return fmt.Errorf("generating pdf: %w", err)
	}

	return nil
}

// ConvertBytes converts Excel data from bytes to HTML string
func (e *SpreadsheetEngine) ConvertBytes(data []byte) (string, error) {
	f, err := e.parser.ParseFromBytes(data)
	if err != nil {
		return "", fmt.Errorf("parsing excel data: %w", err)
	}
	defer f.Close()

	htmlContent, err := e.htmlRenderer.Render(f)
	if err != nil {
		return "", fmt.Errorf("rendering html: %w", err)
	}

	return htmlContent, nil
}

// Validate checks if the input file is a valid Excel file
func (e *SpreadsheetEngine) Validate(file string) error {
	f, err := e.parser.Parse(file)
	if err != nil {
		return fmt.Errorf("invalid excel file: %w", err)
	}
	f.Close()
	return nil
}

// ValidateBytes checks if the input bytes represent a valid Excel file
func (e *SpreadsheetEngine) ValidateBytes(data []byte) error {
	f, err := e.parser.ParseFromBytes(data)
	if err != nil {
		return fmt.Errorf("invalid excel data: %w", err)
	}
	f.Close()
	return nil
}

