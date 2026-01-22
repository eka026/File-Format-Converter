package browser

import (
	"context"
	"os"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// HeadlessBrowser provides headless browser functionality for PDF generation
type HeadlessBrowser struct {
	browser *rod.Browser
}

// NewHeadlessBrowser creates a new headless browser adapter
func NewHeadlessBrowser() (*HeadlessBrowser, error) {
	launcher := launcher.New().
		Headless(true).
		Set("disable-gpu").
		Set("disable-dev-shm-usage")

	url, err := launcher.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, err
	}

	return &HeadlessBrowser{browser: browser}, nil
}

// GeneratePDFFromHTML generates a PDF from HTML content
func (h *HeadlessBrowser) GeneratePDFFromHTML(ctx context.Context, htmlContent string, outputPath string) error {
	page, err := h.browser.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.SetContent(htmlContent); err != nil {
		return err
	}

	pdf, err := page.PDF(&rod.PDFOptions{
		Landscape: false,
		PrintBackground: true,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdf, 0644)
}

// Close closes the browser
func (h *HeadlessBrowser) Close() error {
	return h.browser.Close()
}

