package image

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/eka026/File-Format-Converter/internal/domain"
)

// ImageEngine implements IConverter for image format conversions
type ImageEngine struct {
	workerPool *WorkerPool
	webpEncoder *WebPEncoder
}

// NewImageEngine creates a new image conversion engine
func NewImageEngine(workerPool *WorkerPool, webpEncoder *WebPEncoder) domain.IConverter {
	return &ImageEngine{
		workerPool:  workerPool,
		webpEncoder: webpEncoder,
	}
}

// Convert converts an image from one format to another
func (e *ImageEngine) Convert(input, output string) error {
	// Load image using imaging library
	img, err := imaging.Open(input)
	if err != nil {
		return err
	}

	// Determine output format from file extension
	ext := strings.ToLower(filepath.Ext(output))
	
	switch ext {
	case ".webp":
		return e.webpEncoder.Encode(img, output)
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
func (e *ImageEngine) Validate(file string) error {
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

			// Perform the conversion
			err := e.Convert(task.InputPath, task.OutputPath)

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

