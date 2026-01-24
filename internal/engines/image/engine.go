package image

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/eka026/File-Format-Converter/internal/domain"
)

// ImageEngine implements IConverter for image format conversions
type ImageEngine struct {
	workerPool *WorkerPool
}

// NewImageEngine creates a new image conversion engine
func NewImageEngine(workerPool *WorkerPool) domain.IConverter {
	return &ImageEngine{
		workerPool: workerPool,
	}
}

// Convert converts an image from one format to another
func (e *ImageEngine) Convert(ctx context.Context, input, output string) error {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Load image using imaging library
	img, err := imaging.Open(input)
	if err != nil {
		return err
	}

	// Check for cancellation after loading
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Determine output format from file extension
	ext := strings.ToLower(filepath.Ext(output))
	
	switch ext {
	case ".jpeg", ".jpg":
		// JPEG format - imaging.Save will handle this automatically
		return imaging.Save(img, output)
	case ".png":
		// PNG format - imaging.Save will handle this automatically
		return imaging.Save(img, output)
	default:
		// Use imaging library for other formats (auto-detects from extension)
		return imaging.Save(img, output)
	}
}

// Validate checks if the input file is a valid image
func (e *ImageEngine) Validate(ctx context.Context, file string) error {
	// Check for cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
	_, err := imaging.Open(file)
	return err
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

// BatchConvert processes multiple image conversions in parallel using the worker pool
// It takes a slice of input/output path pairs and processes them concurrently
// utilizing all available CPU cores through the worker pool
func (e *ImageEngine) BatchConvert(tasks []BatchConversionTask) []BatchConversionResult {
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

// Close closes the worker pool to prevent goroutine leaks
func (e *ImageEngine) Close() {
	if e.workerPool != nil {
		e.workerPool.Close()
	}
}

