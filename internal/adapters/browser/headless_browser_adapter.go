package browser

import (
	"context"

	"github.com/eka026/File-Format-Converter/internal/ports"
)

// HeadlessBrowserAdapter is the headless browser driven adapter that implements IPDFGenerator
type HeadlessBrowserAdapter struct {
	headlessBrowser *HeadlessBrowser
}

// NewHeadlessBrowserAdapter creates a new headless browser adapter
func NewHeadlessBrowserAdapter() ports.IPDFGenerator {
	return &HeadlessBrowserAdapter{}
}

// GenerateFromHTML generates a PDF from HTML content and returns PDF bytes
func (a *HeadlessBrowserAdapter) GenerateFromHTML(html []byte) []byte {
	ctx := context.Background()

	// Lazy initialization of headless browser
	if a.headlessBrowser == nil {
		browser, err := NewHeadlessBrowser()
		if err != nil {
			return nil
		}
		a.headlessBrowser = browser
	}

	// Convert HTML bytes to string
	htmlContent := string(html)

	// Generate PDF using headless browser
	pdfBytes, err := a.headlessBrowser.GeneratePDFFromHTMLBytes(ctx, htmlContent)
	if err != nil {
		return nil
	}

	return pdfBytes
}

// Close closes the underlying headless browser
func (a *HeadlessBrowserAdapter) Close() error {
	if a.headlessBrowser != nil {
		return a.headlessBrowser.Close()
	}
	return nil
}
