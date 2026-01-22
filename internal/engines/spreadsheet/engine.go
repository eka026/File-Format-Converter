package spreadsheet

import (
	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/xuri/excelize/v2"
)

// SpreadsheetEngine implements IConverter for spreadsheet conversions (Excel → PDF)
type SpreadsheetEngine struct {
	htmlRenderer *HTMLRenderer
	pdfGenerator *PDFGenerator
}

// NewSpreadsheetEngine creates a new spreadsheet conversion engine
func NewSpreadsheetEngine(htmlRenderer *HTMLRenderer, pdfGenerator *PDFGenerator) domain.IConverter {
	return &SpreadsheetEngine{
		htmlRenderer: htmlRenderer,
		pdfGenerator: pdfGenerator,
	}
}

// Convert converts an Excel file to PDF
func (e *SpreadsheetEngine) Convert(input, output string) error {
	// Parse Excel file using excelize
	f, err := excelize.OpenFile(input)
	if err != nil {
		return err
	}
	defer f.Close()

	// Convert to HTML
	htmlContent, err := e.htmlRenderer.Render(f)
	if err != nil {
		return err
	}

	// Generate PDF from HTML
	return e.pdfGenerator.Generate(htmlContent, output)
}

// Validate checks if the input file is a valid Excel file
func (e *SpreadsheetEngine) Validate(file string) error {
	_, err := excelize.OpenFile(file)
	return err
}

