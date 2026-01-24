package spreadsheet

import (
	"context"
	"fmt"
	"sync"

	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/engines/image"
)

// SpreadsheetEngine implements IConverter for spreadsheet conversions (Excel → PDF)
type SpreadsheetEngine struct {
	parser       *ExcelParser
	htmlRenderer *HTMLRenderer
	pdfGenerator *PDFGenerator
	workerPool   *image.WorkerPool
}

// NewSpreadsheetEngine creates a new spreadsheet conversion engine
func NewSpreadsheetEngine(htmlRenderer *HTMLRenderer, pdfGenerator *PDFGenerator) domain.IConverter {
	return &SpreadsheetEngine{
		parser:       NewExcelParser(),
		htmlRenderer: htmlRenderer,
		pdfGenerator: pdfGenerator,
		workerPool:   image.NewWorkerPool(),
	}
}

// Convert converts an Excel file to PDF
func (e *SpreadsheetEngine) Convert(ctx context.Context, input, output string) error {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Parse Excel file using excelize via our parser
	f, err := e.parser.Parse(input)
	if err != nil {
		return fmt.Errorf("parsing excel file: %w", err)
	}
	defer f.Close()

	// Check for cancellation after parsing
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Convert to HTML (internally parses workbook data)
	htmlContent, err := e.htmlRenderer.Render(f)
	if err != nil {
		return fmt.Errorf("rendering html: %w", err)
	}

	// Check for cancellation after rendering
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Generate PDF from HTML
	if err := e.pdfGenerator.Generate(ctx, htmlContent, output); err != nil {
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
func (e *SpreadsheetEngine) Validate(ctx context.Context, file string) error {
	// Check for cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
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

// BatchConvert processes multiple spreadsheet conversions in parallel using the worker pool
// It takes a slice of input/output path pairs and processes them concurrently
// utilizing all available CPU cores through the worker pool
func (e *SpreadsheetEngine) BatchConvert(tasks []BatchConversionTask) []BatchConversionResult {
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

