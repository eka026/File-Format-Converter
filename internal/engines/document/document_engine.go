package document

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
	"github.com/eka026/File-Format-Converter/internal/engines/image"
)

// DocumentEngine implements IConverter for document conversions using pure Go
// Follows the same pattern as SpreadsheetEngine: Parse → HTML → PDF
type DocumentEngine struct {
	parser       *DocxParser
	htmlRenderer *HTMLRenderer
	pdfGenerator *browser.HeadlessBrowser
	workerPool   *image.WorkerPool
}

// NewDocumentEngine creates a new document conversion engine
// Uses pure Go DOCX parsing (no WASM, no CGO dependencies)
func NewDocumentEngine(pdfGenerator *browser.HeadlessBrowser) domain.IConverter {
	return &DocumentEngine{
		parser:       NewDocxParser(),
		htmlRenderer: NewHTMLRenderer(),
		pdfGenerator: pdfGenerator,
		workerPool:   image.NewWorkerPool(),
	}
}

// Convert converts a DOCX file to the specified output format
// Input and output are file paths (matches domain.IConverter interface)
func (e *DocumentEngine) Convert(ctx context.Context, input, output string) error {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Read DOCX file
	docxData, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("reading docx file: %w", err)
	}

	// Check for cancellation after reading
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Parse DOCX file
	doc, err := e.parser.Parse(docxData)
	if err != nil {
		return fmt.Errorf("parsing docx: %w", err)
	}

	// Check for cancellation after parsing
	if ctx.Err() != nil {
		return ctx.Err()
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
		// Use the provided context instead of Background()
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
func (e *DocumentEngine) Validate(ctx context.Context, file string) error {
	// Check for cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return ValidateDOCX(file)
}

// BatchConversionTask represents a single conversion task in a batch
type BatchConversionTask struct {
	InputPath  string
	OutputPath string
	Index      int
}

// BatchConversionResult represents the result of a batch conversion task
type BatchConversionResult struct {
	Index int
	Error error
}

// BatchConvert processes multiple document conversions in parallel using the worker pool
// It takes a slice of input/output path pairs and processes them concurrently
// utilizing all available CPU cores through the worker pool
func (e *DocumentEngine) BatchConvert(tasks []BatchConversionTask) []BatchConversionResult {
	if len(tasks) == 0 {
		return nil
	}

	results := make([]BatchConversionResult, len(tasks))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Submit all tasks to the worker pool
	for _, task := range tasks {
		wg.Add(1)
		task := task // Capture loop variable

		e.workerPool.Submit(func() {
			defer wg.Done()

			// Perform the conversion with background context
			// Note: For batch operations, we use background context as cancellation
			// should be handled at the batch level, not individual task level
			err := e.Convert(context.Background(), task.InputPath, task.OutputPath)

			// Store result thread-safely
			mu.Lock()
			results[task.Index] = BatchConversionResult{
				Index: task.Index,
				Error: err,
			}
			mu.Unlock()
		})
	}

	// Wait for all conversions to complete
	wg.Wait()

	return results
}


