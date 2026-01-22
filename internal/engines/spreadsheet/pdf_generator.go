package spreadsheet

import (
	"context"
	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
)

// PDFGenerator generates PDF from HTML content
type PDFGenerator struct {
	browser *browser.HeadlessBrowser
}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator(browser *browser.HeadlessBrowser) *PDFGenerator {
	return &PDFGenerator{
		browser: browser,
	}
}

// Generate generates a PDF file from HTML content
func (g *PDFGenerator) Generate(htmlContent, outputPath string) error {
	ctx := context.Background()
	return g.browser.GeneratePDFFromHTML(ctx, htmlContent, outputPath)
}

