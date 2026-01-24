package browser

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// HeadlessBrowser provides headless browser functionality for PDF generation
type HeadlessBrowser struct {
	browser *rod.Browser
}

// findChromeExecutable finds Chrome/Chromium executable on the system
func findChromeExecutable() (string, error) {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Common Chrome installation paths on Windows
		chromePaths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			os.Getenv("LOCALAPPDATA") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("PROGRAMFILES") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("PROGRAMFILES(X86)") + `\Google\Chrome\Application\chrome.exe`,
		}
		paths = chromePaths

		// Also try Edge Chromium
		edgePaths := []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			os.Getenv("PROGRAMFILES(X86)") + `\Microsoft\Edge\Application\msedge.exe`,
		}
		paths = append(paths, edgePaths...)

	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}

	case "linux":
		paths = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	// Check each path
	for _, path := range paths {
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			// Verify it's executable
			if runtime.GOOS != "windows" {
				if info.Mode().Perm()&0111 != 0 {
					return path, nil
				}
			} else {
				return path, nil
			}
		}
	}

	// Try to find via command
	if runtime.GOOS != "windows" {
		if path, err := exec.LookPath("google-chrome"); err == nil {
			return path, nil
		}
		if path, err := exec.LookPath("chromium"); err == nil {
			return path, nil
		}
		if path, err := exec.LookPath("chromium-browser"); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("Chrome/Chromium not found in common locations")
}

// NewHeadlessBrowser creates a new headless browser adapter
// NFR-01 Compliance: This function only uses locally installed browsers.
// It will NOT download Chromium or any other browser binaries from the internet.
// If no local browser is found, it returns an error to maintain data sovereignty.
func NewHeadlessBrowser() (*HeadlessBrowser, error) {
	// NFR-01 (Data Sovereignty): Only use locally installed browsers.
	// Do not allow auto-download of browser binaries from external servers.
	chromePath, err := findChromeExecutable()
	if err != nil {
		return nil, fmt.Errorf("no local browser found: %w\n\n"+
			"Please install Chrome, Chromium, or Edge locally. "+
			"The application does not download browsers from the internet to maintain data sovereignty. "+
			"You can install Chrome from: https://www.google.com/chrome/", err)
	}

	// Use only the locally found Chrome/Chromium installation
	l := launcher.New().Bin(chromePath)

	// Configure launcher
	l = l.
		Headless(true).
		Set("disable-gpu").
		Set("disable-dev-shm-usage").
		Set("no-sandbox") // May be needed in some environments

	url, err := l.Launch()
	if err != nil {
		// Provide helpful error message
		return nil, fmt.Errorf("failed to launch browser: %w\n\n"+
			"Please ensure Chrome, Chromium, or Edge is installed locally. "+
			"The application only uses locally installed browsers to maintain data sovereignty. "+
			"You can install Chrome from: https://www.google.com/chrome/", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	return &HeadlessBrowser{browser: browser}, nil
}

// GeneratePDFFromHTML generates a PDF from HTML content and writes it to a file
func (h *HeadlessBrowser) GeneratePDFFromHTML(ctx context.Context, htmlContent string, outputPath string) error {
	pdf, err := h.GeneratePDFFromHTMLBytes(ctx, htmlContent)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdf, 0644)
}

// GeneratePDFFromHTMLBytes generates a PDF from HTML content and returns the PDF bytes
func (h *HeadlessBrowser) GeneratePDFFromHTMLBytes(ctx context.Context, htmlContent string) ([]byte, error) {
	page, err := h.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, err
	}
	defer page.Close()

	if err := page.SetDocumentContent(htmlContent); err != nil {
		return nil, err
	}

	// Wait for page to be ready
	page.MustWaitStable()

	reader, err := page.PDF(&proto.PagePrintToPDF{
		Landscape:       false,
		PrintBackground: true,
	})
	if err != nil {
		return nil, err
	}

	pdf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

// Close closes the browser
func (h *HeadlessBrowser) Close() error {
	return h.browser.Close()
}
