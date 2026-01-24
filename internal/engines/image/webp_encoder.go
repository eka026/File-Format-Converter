package image

import (
	"context"
	"image"
	"os"

	"github.com/chai2010/webp"
)

// WebPEncoder encodes images to WebP format using native encoder
type WebPEncoder struct{}

// NewWebPEncoder creates a new WebP encoder
func NewWebPEncoder() *WebPEncoder {
	return &WebPEncoder{}
}

// Encode encodes an image to WebP format
func (e *WebPEncoder) Encode(ctx context.Context, img image.Image, outputPath string) error {
	// Check for cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check for cancellation after file creation
	if ctx.Err() != nil {
		file.Close()
		os.Remove(outputPath)
		return ctx.Err()
	}

	// Encode image to WebP format with lossy encoding (quality 80)
	// Quality ranges from 0 (lowest quality, smallest size) to 100 (highest quality, largest size)
	options := &webp.Options{
		Lossless: false,
		Quality:  80,
	}

	if err := webp.Encode(file, img, options); err != nil {
		// If encoding fails, remove the empty/corrupt file
		os.Remove(outputPath)
		return err
	}

	return nil
}
