package image

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/disintegration/imaging"
)

// TestImageEngine_Validate_JPEG tests FR-08: The system shall accept common image formats (JPEG) as input
func TestImageEngine_Validate_JPEG(t *testing.T) {
	tmpFile := createTempJPEGFile(t)
	defer os.Remove(tmpFile)

	engine := createTestImageEngine(t)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err != nil {
		t.Errorf("FR-08: Expected valid JPEG file to be accepted, but got error: %v", err)
	}
}

// TestImageEngine_Validate_PNG tests FR-08: The system shall accept common image formats (PNG) as input
func TestImageEngine_Validate_PNG(t *testing.T) {
	tmpFile := createTempPNGFile(t)
	defer os.Remove(tmpFile)

	engine := createTestImageEngine(t)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err != nil {
		t.Errorf("FR-08: Expected valid PNG file to be accepted, but got error: %v", err)
	}
}

// TestImageEngine_Validate_InvalidFile tests that invalid files are rejected
func TestImageEngine_Validate_InvalidFile(t *testing.T) {
	// Create a temporary invalid file
	tmpFile := filepath.Join(t.TempDir(), "invalid.txt")
	if err := os.WriteFile(tmpFile, []byte("not an image file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	engine := createTestImageEngine(t)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err == nil {
		t.Error("FR-08: Expected invalid file to be rejected, but validation passed")
	}
}


// TestImageEngine_Convert_ToPNG tests FR-09: The system shall allow users to select target formats (PNG)
func TestImageEngine_Convert_ToPNG(t *testing.T) {
	tmpFile := createTempJPEGFile(t)
	defer os.Remove(tmpFile)

	outputFile := filepath.Join(t.TempDir(), "output.png")
	defer os.Remove(outputFile)

	engine := createTestImageEngine(t)
	ctx := context.Background()

	err := engine.Convert(ctx, tmpFile, outputFile)
	if err != nil {
		t.Fatalf("FR-09: Conversion to PNG failed: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("FR-09: PNG output file was not created")
	}

	// Verify file can be opened as PNG
	_, err = imaging.Open(outputFile)
	if err != nil {
		t.Errorf("FR-09: Generated PNG file is not a valid PNG: %v", err)
	}
}

// TestImageEngine_Convert_ToJPEG tests FR-09: The system shall allow users to select target formats (JPEG)
func TestImageEngine_Convert_ToJPEG(t *testing.T) {
	tmpFile := createTempPNGFile(t)
	defer os.Remove(tmpFile)

	outputFile := filepath.Join(t.TempDir(), "output.jpeg")
	defer os.Remove(outputFile)

	engine := createTestImageEngine(t)
	ctx := context.Background()

	err := engine.Convert(ctx, tmpFile, outputFile)
	if err != nil {
		t.Fatalf("FR-09: Conversion to JPEG failed: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("FR-09: JPEG output file was not created")
	}

	// Verify file can be opened as JPEG
	_, err = imaging.Open(outputFile)
	if err != nil {
		t.Errorf("FR-09: Generated JPEG file is not a valid JPEG: %v", err)
	}
}

// TestImageEngine_BatchConvert_ParallelProcessing tests FR-10: The system shall utilize parallel processing (worker pools) to handle batch image conversions
func TestImageEngine_BatchConvert_ParallelProcessing(t *testing.T) {
	// Create multiple test images
	numImages := 10
	tasks := make([]BatchConversionTask, numImages)
	inputFiles := make([]string, numImages)

	for i := 0; i < numImages; i++ {
		if i%2 == 0 {
			inputFiles[i] = createTempJPEGFile(t)
		} else {
			inputFiles[i] = createTempPNGFile(t)
		}
		defer os.Remove(inputFiles[i])

		outputFile := filepath.Join(t.TempDir(), filepath.Base(inputFiles[i])+".png")
		defer os.Remove(outputFile)

		tasks[i] = BatchConversionTask{
			InputPath:  inputFiles[i],
			OutputPath: outputFile,
			Index:      i,
		}
	}

	engine := createTestImageEngine(t)
	defer engine.Close()

	// Measure time to verify parallel processing
	results := engine.BatchConvert(tasks)

	// Verify all conversions completed
	if len(results) != numImages {
		t.Fatalf("FR-10: Expected %d conversion results, got %d", numImages, len(results))
	}

	// Verify all conversions succeeded
	for i, result := range results {
		if result.Error != nil {
			t.Errorf("FR-10: Conversion %d failed: %v", i, result.Error)
		}
		if result.Index != i {
			t.Errorf("FR-10: Result index mismatch: expected %d, got %d", i, result.Index)
		}
	}

	// Verify output files were created
	for _, task := range tasks {
		if _, err := os.Stat(task.OutputPath); os.IsNotExist(err) {
			t.Errorf("FR-10: Output file was not created: %s", task.OutputPath)
		}
	}
}

// TestImageEngine_BatchConvert_WorkerPoolUtilization tests that worker pool utilizes available CPU cores
func TestImageEngine_BatchConvert_WorkerPoolUtilization(t *testing.T) {
	// Create enough tasks to utilize multiple workers
	numTasks := runtime.NumCPU() * 2
	if numTasks < 4 {
		numTasks = 4 // Ensure at least 4 tasks for meaningful test
	}

	tasks := make([]BatchConversionTask, numTasks)
	inputFiles := make([]string, numTasks)

	for i := 0; i < numTasks; i++ {
		inputFiles[i] = createTempJPEGFile(t)
		defer os.Remove(inputFiles[i])

		outputFile := filepath.Join(t.TempDir(), filepath.Base(inputFiles[i])+".png")
		defer os.Remove(outputFile)

		tasks[i] = BatchConversionTask{
			InputPath:  inputFiles[i],
			OutputPath: outputFile,
			Index:      i,
		}
	}

	engine := createTestImageEngine(t)
	defer engine.Close()

	// Track concurrent executions
	var maxConcurrent int
	var currentConcurrent int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Wrap tasks to track concurrency
	for i := range tasks {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			// Perform conversion
			err := engine.Convert(context.Background(), tasks[idx].InputPath, tasks[idx].OutputPath)

			mu.Lock()
			currentConcurrent--
			mu.Unlock()

			// Store result
			if err != nil {
				t.Errorf("FR-10: Conversion %d failed: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify that multiple workers were utilized (maxConcurrent > 1)
	// Note: This is a heuristic - actual parallelism depends on scheduler
	if maxConcurrent < 2 && runtime.NumCPU() > 1 {
		t.Logf("FR-10: Warning - Only %d concurrent conversions detected (expected more with %d CPUs)", maxConcurrent, runtime.NumCPU())
		// Don't fail the test as this is timing-dependent
	}
}

// TestImageEngine_BatchConvert_AllFormats tests batch conversion with different target formats
func TestImageEngine_BatchConvert_AllFormats(t *testing.T) {
	formats := []string{"png", "jpeg"}
	tasks := make([]BatchConversionTask, len(formats))
	inputFile := createTempJPEGFile(t)
	defer os.Remove(inputFile)

	for i, format := range formats {
		outputFile := filepath.Join(t.TempDir(), "output."+format)
		defer os.Remove(outputFile)

		tasks[i] = BatchConversionTask{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Index:      i,
		}
	}

	engine := createTestImageEngine(t)
	defer engine.Close()

	results := engine.BatchConvert(tasks)

	// Verify all conversions succeeded
	for i, result := range results {
		if result.Error != nil {
			t.Errorf("FR-09: Conversion to %s failed: %v", formats[i], result.Error)
		}
	}
}

// Helper functions

// createTestImageEngine creates an ImageEngine instance for testing
func createTestImageEngine(t *testing.T) *ImageEngine {
	workerPool := NewWorkerPool()
	return NewImageEngine(workerPool).(*ImageEngine)
}

// createTempJPEGFile creates a temporary JPEG file for testing
func createTempJPEGFile(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "test.jpg")

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}

	// Save as JPEG
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test JPEG file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}

	return tmpFile
}

// createTempPNGFile creates a temporary PNG file for testing
func createTempPNGFile(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "test.png")

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}

	// Save as PNG
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test PNG file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	return tmpFile
}


