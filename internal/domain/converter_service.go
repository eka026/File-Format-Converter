package domain

// NFR-01 (Data Sovereignty): All file conversion operations in this service
// are performed locally. No file data, metadata, or telemetry is transmitted
// to external servers. All processing uses local system resources only.

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ConverterService orchestrates the conversion process
// All operations are performed locally - no external data transmission
type ConverterService struct {
	engines          map[FileType]IConverter
	logger           Logger
	progressNotifier ProgressNotifier
	fileWriter       FileWriter
}

// NewConverterService creates a new converter service
func NewConverterService(
	engines map[FileType]IConverter,
	logger Logger,
	progressNotifier ProgressNotifier,
	fileWriter FileWriter,
) *ConverterService {
	return &ConverterService{
		engines:          engines,
		logger:           logger,
		progressNotifier: progressNotifier,
		fileWriter:       fileWriter,
	}
}

// Convert performs a single file conversion
func (s *ConverterService) Convert(ctx context.Context, source, target string) Result {
	startTime := time.Now()

	// Check for cancellation before starting
	if ctx.Err() != nil {
		err := ctx.Err()
		s.logger.Error("Conversion cancelled before start", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	s.logger.Info(fmt.Sprintf("Starting conversion: %s -> %s", source, target))
	s.progressNotifier.NotifyProgress(0, "Starting conversion...")

	// Validate input file
	validationResult := s.validateInput(ctx, source)
	if !validationResult.Valid {
		s.logger.Error("File validation failed", validationResult.Error)
		s.progressNotifier.NotifyError(validationResult.Error)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      validationResult.Error,
			Duration:   time.Since(startTime),
		}
	}

	// Check for cancellation after validation
	if ctx.Err() != nil {
		err := ctx.Err()
		s.logger.Error("Conversion cancelled after validation", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	// Detect file type
	fileType := s.detectFileType(source)
	if fileType == "" {
		err := fmt.Errorf("unsupported file type: %s", source)
		s.logger.Error("Unsupported file type", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	// Select appropriate engine
	engine := s.selectEngine(fileType)
	if engine == nil {
		err := fmt.Errorf("no conversion engine available for file type: %s", fileType)
		s.logger.Error("Engine not found", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	// Validate file using engine
	if err := engine.Validate(ctx, source); err != nil {
		s.logger.Error("Engine validation failed", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	// Check for cancellation before conversion
	if ctx.Err() != nil {
		err := ctx.Err()
		s.logger.Error("Conversion cancelled before engine conversion", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	// Perform conversion
	s.progressNotifier.NotifyProgress(50, "Converting file...")
	if err := engine.Convert(ctx, source, target); err != nil {
		s.logger.Error("Conversion failed", err)
		s.progressNotifier.NotifyError(err)
		return Result{
			Success:    false,
			OutputPath: "",
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	duration := time.Since(startTime)
	s.logger.Info(fmt.Sprintf("Conversion completed successfully in %v", duration))
	s.progressNotifier.NotifyProgress(100, "Conversion completed")

	result := Result{
		Success:    true,
		OutputPath: target,
		Error:      nil,
		Duration:   duration,
	}
	s.progressNotifier.NotifyComplete(result)

	return result
}

// BatchConvert performs batch file conversion
func (s *ConverterService) BatchConvert(ctx context.Context, files []string, target string) []Result {
	if len(files) == 0 {
		return nil
	}

	// Check for cancellation before starting
	if ctx.Err() != nil {
		s.logger.Error("Batch conversion cancelled before start", ctx.Err())
		return nil
	}

	s.logger.Info(fmt.Sprintf("Starting batch conversion of %d files", len(files)))
	s.progressNotifier.NotifyProgress(0, fmt.Sprintf("Starting batch conversion of %d files...", len(files)))

	results := make([]Result, len(files))
	totalFiles := len(files)

	for i, file := range files {
		// Check for cancellation before each file
		if ctx.Err() != nil {
			s.logger.Info(fmt.Sprintf("Batch conversion cancelled after %d of %d files", i, totalFiles))
			// Return partial results
			return results[:i]
		}

		// Generate output path for each file
		outputPath := s.generateOutputPath(file, target)

		// Convert single file
		result := s.Convert(ctx, file, outputPath)
		results[i] = result

		// Update progress
		progress := (i + 1) * 100 / totalFiles
		s.progressNotifier.NotifyProgress(progress, fmt.Sprintf("Converted %d of %d files", i+1, totalFiles))
	}

	s.logger.Info(fmt.Sprintf("Batch conversion completed: %d files processed", len(files)))
	return results
}

// GetSupportedFormats returns the list of supported conversion formats
func (s *ConverterService) GetSupportedFormats() []Format {
	// Note: IConverter interface doesn't have GetSupportedOutputTypes method
	// So we return the known formats based on the engines we have
	// If engines are registered, we support these formats
	if len(s.engines) == 0 {
		return nil
	}

	// Return all supported output formats
	return []Format{
		FormatPDF,
		FormatHTML,
		FormatPNG,
		FormatWEBP,
	}
}

// ValidateFile validates if a file can be converted
func (s *ConverterService) ValidateFile(ctx context.Context, file string) ValidationResult {
	// Check for cancellation
	if ctx.Err() != nil {
		return ValidationResult{
			Valid:   false,
			Message: "Validation cancelled",
			Error:   ctx.Err(),
		}
	}

	// Check if file exists
	if !s.fileWriter.Exists(file) {
		return ValidationResult{
			Valid:   false,
			Message: "File does not exist",
			Error:   fmt.Errorf("file does not exist: %s", file),
		}
	}

	// Detect file type
	fileType := s.detectFileType(file)
	if fileType == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Unsupported file type",
			Error:   fmt.Errorf("unsupported file type: %s", file),
		}
	}

	// Check if engine exists for this file type
	engine := s.selectEngine(fileType)
	if engine == nil {
		return ValidationResult{
			Valid:   false,
			Message: "No conversion engine available for this file type",
			Error:   fmt.Errorf("no conversion engine available for file type: %s", fileType),
		}
	}

	// Validate using engine
	if err := engine.Validate(ctx, file); err != nil {
		return ValidationResult{
			Valid:   false,
			Message: "File validation failed",
			Error:   err,
		}
	}

	return ValidationResult{
		Valid:   true,
		Message: "File is valid and can be converted",
		Error:   nil,
	}
}

// selectEngine selects the appropriate conversion engine for a file type
func (s *ConverterService) selectEngine(fileType FileType) IConverter {
	engine, exists := s.engines[fileType]
	if !exists {
		return nil
	}
	return engine
}

// validateInput validates the input file
func (s *ConverterService) validateInput(ctx context.Context, file string) ValidationResult {
	// Check for cancellation
	if ctx.Err() != nil {
		return ValidationResult{
			Valid:   false,
			Message: "Validation cancelled",
			Error:   ctx.Err(),
		}
	}

	// Check if file exists
	if !s.fileWriter.Exists(file) {
		return ValidationResult{
			Valid:   false,
			Message: "File does not exist",
			Error:   fmt.Errorf("file does not exist: %s", file),
		}
	}

	// Detect file type
	fileType := s.detectFileType(file)
	if fileType == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Unsupported file type",
			Error:   fmt.Errorf("unsupported file type: %s", file),
		}
	}

	// Check if engine exists for this file type
	engine := s.selectEngine(fileType)
	if engine == nil {
		return ValidationResult{
			Valid:   false,
			Message: "No conversion engine available for this file type",
			Error:   fmt.Errorf("no conversion engine available for file type: %s", fileType),
		}
	}

	return ValidationResult{
		Valid:   true,
		Message: "Input file is valid",
		Error:   nil,
	}
}

// detectFileType detects the file type from the file extension
func (s *ConverterService) detectFileType(filePath string) FileType {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".docx":
		return FileTypeDOCX
	case ".xlsx":
		return FileTypeXLSX
	case ".jpeg", ".jpg":
		return FileTypeJPEG
	case ".png":
		return FileTypePNG
	case ".webp":
		return FileTypeWEBP
	default:
		return ""
	}
}

// generateOutputPath generates an output file path based on input path and target format
func (s *ConverterService) generateOutputPath(inputPath, targetFormat string) string {
	dir := filepath.Dir(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := strings.ToLower(targetFormat)

	// Normalize JPEG format (handle both "jpg" and "jpeg")
	if ext == "jpg" {
		ext = "jpeg"
	}

	return filepath.Join(dir, baseName+"."+ext)
}
