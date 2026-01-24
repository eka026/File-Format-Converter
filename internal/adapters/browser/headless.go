package browser

import (
	"context"
	"io"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// HeadlessBrowser provides headless browser functionality for PDF generation
type HeadlessBrowser struct {
	browser *rod.Browser
}

// NewHeadlessBrowser creates a new headless browser adapter
func NewHeadlessBrowser() (*HeadlessBrowser, error) {
	l := launcher.New().
		Headless(true).
		Set("disable-gpu").
		Set("disable-dev-shm-usage")

	url, err := l.Launch()
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
	page, err := h.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.SetDocumentContent(htmlContent); err != nil {
		return err
	}

	// Wait for page to be ready
	page.MustWaitStable()

	reader, err := page.PDF(&proto.PagePrintToPDF{
		Landscape:       false,
		PrintBackground: true,
	})
	if err != nil {
		return err
	}

	pdf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdf, 0644)
}

// Close closes the browser
func (h *HeadlessBrowser) Close() error {
	return h.browser.Close()
}
