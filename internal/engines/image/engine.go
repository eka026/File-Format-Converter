package image

import (
	"image"
	"path/filepath"
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

	// Determine output format
	ext := filepath.Ext(output)
	
	switch ext {
	case ".webp":
		return e.webpEncoder.Encode(img, output)
	default:
		// Use imaging library for other formats
		return imaging.Save(img, output)
	}
}

// Validate checks if the input file is a valid image
func (e *ImageEngine) Validate(file string) error {
	_, err := imaging.Open(file)
	return err
}

