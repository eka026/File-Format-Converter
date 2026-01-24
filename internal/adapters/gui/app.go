package gui

// NFR-01 (Data Sovereignty): This GUI adapter handles all file operations locally.
// No file data, metadata, or telemetry is transmitted to external servers.
// All file I/O operations use the local filesystem only.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
	"github.com/eka026/File-Format-Converter/internal/adapters/filesystem"
	"github.com/eka026/File-Format-Converter/internal/adapters/logger"
	"github.com/eka026/File-Format-Converter/internal/adapters/progress"
	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/engines/document"
	"github.com/eka026/File-Format-Converter/internal/engines/image"
	"github.com/eka026/File-Format-Converter/internal/engines/spreadsheet"
)

const (
	// tempFileCleanupDelay is the delay before cleaning up temporary files
	// to ensure file operations are complete
	tempFileCleanupDelay = 5 * time.Second
)

// ConversionResult represents the result of a file conversion
type ConversionResult struct {
	Success    bool   `json:"success"`
	OutputPath string `json:"outputPath,omitempty"`
	Error      string `json:"error,omitempty"`
}

// App represents the GUI application adapter
type App struct {
	ctx               context.Context
	converterService  *domain.ConverterService
	spreadsheetEngine domain.IConverter
	documentEngine    domain.IConverter
	imageEngine       domain.IConverter
	headlessBrowser   *browser.HeadlessBrowser
	logger            domain.Logger
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	return &App{
		logger: logger.NewDomainLoggerAdapter(logger.LogLevelInfo),
	}
}

// OnStartup is called when the application starts
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Initialize headless browser
	browser, err := browser.NewHeadlessBrowser()
	if err != nil {
		// Log error but don't fail startup - browser will be initialized lazily
		a.logger.Error("Could not initialize headless browser", err)
	} else {
		a.headlessBrowser = browser
	}

	// Initialize engines
	if a.headlessBrowser != nil {
		if err := a.initializeEngines(); err != nil {
			a.logger.Error("Could not initialize engines", err)
		}
	}

	// Initialize ConverterService with engines
	if err := a.initializeConverterService(); err != nil {
		a.logger.Error("Could not initialize converter service", err)
	}
}

// OnDomReady is called when the DOM is ready
func (a *App) OnDomReady(ctx context.Context) {
	// Initialize frontend components
}

// OnShutdown is called when the application shuts down
func (a *App) OnShutdown(ctx context.Context) {
	// Cleanup resources
	if a.headlessBrowser != nil {
		a.headlessBrowser.Close()
	}

	// Close ImageEngine worker pool to prevent goroutine leaks
	if imgEngine, ok := a.imageEngine.(*image.ImageEngine); ok {
		imgEngine.Close()
	}

	// Clean up all temp files on shutdown
	if err := a.CleanupTempFiles(); err != nil {
		a.logger.Error("Failed to cleanup temp files", err)
	}
}

// ConvertFile handles file conversion from the GUI
// Returns the output file path on success
// If outputPath is empty, shows a save dialog to let user choose location
func (a *App) ConvertFile(sourcePath, targetFormat string) ConversionResult {
	return a.ConvertFileWithPath(sourcePath, targetFormat, "")
}

// ConvertFileWithPath handles file conversion with a specific output path
// If outputPath is empty, shows a save dialog
func (a *App) ConvertFileWithPath(sourcePath, targetFormat, outputPath string) ConversionResult {
	// Ensure ConverterService is initialized
	if a.converterService == nil {
		if err := a.initializeConverterService(); err != nil {
			return ConversionResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to initialize converter service: %v", err),
			}
		}
	}

	// Track if output is in temp directory for cleanup scheduling
	var isTempFile bool
	var tempFilePath string

	// If no output path provided, use default location (skip dialog to avoid WebSocket issues)
	if outputPath == "" {
		// Use default location: same directory as source file, or Downloads folder if source is in temp
		outputPath = a.generateOutputPath(sourcePath, targetFormat)

		// If source is in temp directory, save to user's Downloads folder instead
		tempDir := filepath.Join(os.TempDir(), "file-format-converter")
		cleanSourcePath := filepath.Clean(sourcePath)
		cleanTempDir := filepath.Clean(tempDir)
		if strings.HasPrefix(cleanSourcePath, cleanTempDir) {
			// Get user's Downloads folder
			homeDir, err := os.UserHomeDir()
			if err == nil {
				downloadsDir := filepath.Join(homeDir, "Downloads")
				baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
				formatExt := strings.ToLower(targetFormat)
				// Normalize JPEG format (handle both "jpg" and "jpeg")
				if formatExt == "jpg" {
					formatExt = "jpeg"
				}
				outputPath = filepath.Join(downloadsDir, baseName+"."+formatExt)
			}
		}
	}

	// Check if output path is in temp directory
	tempDir := filepath.Join(os.TempDir(), "file-format-converter")
	cleanOutputPath := filepath.Clean(outputPath)
	cleanTempDir := filepath.Clean(tempDir)
	if strings.HasPrefix(cleanOutputPath, cleanTempDir) {
		isTempFile = true
		tempFilePath = outputPath
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create output directory: %v", err),
		}
	}

	// Convert directly to final destination (no temp file copy needed)
	ctx := a.getContext()
	result := a.converterService.Convert(ctx, sourcePath, outputPath)
	if !result.Success {
		// Clean up output file on error if it was created
		os.Remove(outputPath)
		errorMsg := "Conversion failed"
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		return ConversionResult{
			Success: false,
			Error:   errorMsg,
		}
	}

	// If the final output is in temp, mark it for cleanup
	if isTempFile {
		// Schedule cleanup of temp file (could be done on app shutdown or after a delay)
		go a.scheduleTempFileCleanup(tempFilePath)
	}

	return ConversionResult{
		Success:    true,
		OutputPath: outputPath,
	}
}

// BatchConvertFiles handles batch file conversion from the GUI
// Uses parallel processing with worker pools to utilize all CPU cores for images, documents, and spreadsheets
func (a *App) BatchConvertFiles(files []string, targetFormat string) []ConversionResult {
	if len(files) == 0 {
		return nil
	}

	// Detect file types and check if all are the same type with short-circuiting
	// Detect first file type
	firstType := a.detectFileType(files[0])

	// Check remaining files, short-circuiting on first mismatch
	allSameType := true
	for i := 1; i < len(files); i++ {
		fileType := a.detectFileType(files[i])
		if fileType != firstType {
			allSameType = false
			break
		}
	}

	// Use parallel processing for homogeneous batches
	if allSameType {
		switch firstType {
		case domain.FileTypeJPEG, domain.FileTypePNG, domain.FileTypeWEBP:
			// All files are images - use parallel image conversion
			if a.imageEngine != nil {
				if imgEngine, ok := a.imageEngine.(*image.ImageEngine); ok {
					return a.batchConvertImagesParallel(imgEngine, files, targetFormat)
				} else {
					a.logger.Error("Image engine type assertion failed, falling back to sequential processing", fmt.Errorf("expected *image.ImageEngine, got %T", a.imageEngine))
				}
			}
		case domain.FileTypeDOCX:
			// All files are documents - use parallel document conversion
			if a.documentEngine != nil {
				if docEngine, ok := a.documentEngine.(*document.DocumentEngine); ok {
					return a.batchConvertDocumentsParallel(docEngine, files, targetFormat)
				} else {
					a.logger.Error("Document engine type assertion failed, falling back to sequential processing", fmt.Errorf("expected *document.DocumentEngine, got %T", a.documentEngine))
				}
			}
		case domain.FileTypeXLSX:
			// All files are spreadsheets - use parallel spreadsheet conversion
			if a.spreadsheetEngine != nil {
				if se, ok := a.spreadsheetEngine.(*spreadsheet.SpreadsheetEngine); ok {
					return a.batchConvertSpreadsheetsParallel(se, files, targetFormat)
				} else {
					a.logger.Error("Spreadsheet engine type assertion failed, falling back to sequential processing", fmt.Errorf("expected *spreadsheet.SpreadsheetEngine, got %T", a.spreadsheetEngine))
				}
			}
		}
	}

	// Fall back to using ConverterService for mixed file types or if engines are not available
	if a.converterService == nil {
		if err := a.initializeConverterService(); err != nil {
			// Last resort: sequential processing without service
			results := make([]ConversionResult, len(files))
			for i, file := range files {
				results[i] = a.ConvertFile(file, targetFormat)
			}
			return results
		}
	}

	// Use ConverterService for batch conversion
	ctx := a.getContext()
	domainResults := a.converterService.BatchConvert(ctx, files, targetFormat)
	results := make([]ConversionResult, len(domainResults))
	for i, domainResult := range domainResults {
		if domainResult.Success {
			results[i] = ConversionResult{
				Success:    true,
				OutputPath: domainResult.OutputPath,
			}
		} else {
			errorMsg := "Conversion failed"
			if domainResult.Error != nil {
				errorMsg = domainResult.Error.Error()
			}
			results[i] = ConversionResult{
				Success: false,
				Error:   errorMsg,
			}
		}
	}

	return results
}

// batchConvertImagesParallel performs parallel batch conversion of images using worker pools
func (a *App) batchConvertImagesParallel(
	imgEngine *image.ImageEngine,
	files []string,
	targetFormat string,
) []ConversionResult {
	// Initialize image engine if needed
	if imgEngine == nil {
		if err := a.initializeImageEngine(); err != nil {
			// Fall back to sequential if initialization fails
			results := make([]ConversionResult, len(files))
			for i, file := range files {
				results[i] = a.ConvertFile(file, targetFormat)
			}
			return results
		}
		// Re-cast after initialization
		if ie, ok := a.imageEngine.(*image.ImageEngine); ok {
			imgEngine = ie
		}
	}

	// Prepare batch conversion tasks
	tasks := make([]image.BatchConversionTask, len(files))
	for i, file := range files {
		// Generate output path for each file
		outputPath := a.generateOutputPath(file, targetFormat)

		tasks[i] = image.BatchConversionTask{
			InputPath:  file,
			OutputPath: outputPath,
			Index:      i,
		}
	}

	// Perform parallel batch conversion using worker pool
	batchResults := imgEngine.BatchConvert(tasks)

	// Convert batch results to ConversionResult format
	results := make([]ConversionResult, len(batchResults))
	for i, batchResult := range batchResults {
		if batchResult.Error != nil {
			results[i] = ConversionResult{
				Success: false,
				Error:   batchResult.Error.Error(),
			}
		} else {
			results[i] = ConversionResult{
				Success:    true,
				OutputPath: tasks[i].OutputPath,
			}
		}
	}

	return results
}

// batchConvertDocumentsParallel performs parallel batch conversion of documents using worker pools
func (a *App) batchConvertDocumentsParallel(
	docEngine *document.DocumentEngine,
	files []string,
	targetFormat string,
) []ConversionResult {
	// Initialize document engine if needed
	if docEngine == nil {
		if err := a.initializeDocumentEngine(); err != nil {
			// Fall back to sequential if initialization fails
			results := make([]ConversionResult, len(files))
			for i, file := range files {
				results[i] = a.ConvertFile(file, targetFormat)
			}
			return results
		}
		// Re-cast after initialization
		if de, ok := a.documentEngine.(*document.DocumentEngine); ok {
			docEngine = de
		}
	}

	// Prepare batch conversion tasks
	tasks := make([]document.BatchConversionTask, len(files))
	for i, file := range files {
		// Generate output path for each file
		outputPath := a.generateOutputPath(file, targetFormat)

		tasks[i] = document.BatchConversionTask{
			InputPath:  file,
			OutputPath: outputPath,
			Index:      i,
		}
	}

	// Perform parallel batch conversion using worker pool
	batchResults := docEngine.BatchConvert(tasks)

	// Convert batch results to ConversionResult format
	results := make([]ConversionResult, len(batchResults))
	for i, batchResult := range batchResults {
		if batchResult.Error != nil {
			results[i] = ConversionResult{
				Success: false,
				Error:   batchResult.Error.Error(),
			}
		} else {
			results[i] = ConversionResult{
				Success:    true,
				OutputPath: tasks[i].OutputPath,
			}
		}
	}

	return results
}

// batchConvertSpreadsheetsParallel performs parallel batch conversion of spreadsheets using worker pools
func (a *App) batchConvertSpreadsheetsParallel(
	spreadsheetEngine *spreadsheet.SpreadsheetEngine,
	files []string,
	targetFormat string,
) []ConversionResult {
	// Initialize spreadsheet engine if needed
	if spreadsheetEngine == nil {
		if err := a.initializeSpreadsheetEngine(); err != nil {
			// Fall back to sequential if initialization fails
			results := make([]ConversionResult, len(files))
			for i, file := range files {
				results[i] = a.ConvertFile(file, targetFormat)
			}
			return results
		}
		// Re-cast after initialization
		if se, ok := a.spreadsheetEngine.(*spreadsheet.SpreadsheetEngine); ok {
			spreadsheetEngine = se
		} else {
			// Fall back to sequential if type assertion fails
			results := make([]ConversionResult, len(files))
			for i, file := range files {
				results[i] = a.ConvertFile(file, targetFormat)
			}
			return results
		}
	}

	// Prepare batch conversion tasks
	tasks := make([]spreadsheet.BatchConversionTask, len(files))
	for i, file := range files {
		// Generate output path for each file
		outputPath := a.generateOutputPath(file, targetFormat)

		tasks[i] = spreadsheet.BatchConversionTask{
			InputPath:  file,
			OutputPath: outputPath,
			Index:      i,
		}
	}

	// Perform parallel batch conversion using worker pool
	batchResults := spreadsheetEngine.BatchConvert(tasks)

	// Convert batch results to ConversionResult format
	results := make([]ConversionResult, len(batchResults))
	for i, batchResult := range batchResults {
		if batchResult.Error != nil {
			results[i] = ConversionResult{
				Success: false,
				Error:   batchResult.Error.Error(),
			}
		} else {
			results[i] = ConversionResult{
				Success:    true,
				OutputPath: tasks[i].OutputPath,
			}
		}
	}

	return results
}

// GetSupportedFormats returns supported formats for the GUI
func (a *App) GetSupportedFormats() []string {
	if a.converterService == nil {
		// Fallback to default formats if service not initialized
		return []string{"pdf", "html"}
	}

	formats := a.converterService.GetSupportedFormats()
	result := make([]string, len(formats))
	for i, format := range formats {
		result[i] = strings.ToLower(string(format))
	}
	return result
}

// OpenFile opens a file in the default system application
func (a *App) OpenFile(filePath string) error {
	// Clean and normalize the path
	normalizedPath := filepath.Clean(filePath)

	// Remove any leading/trailing whitespace
	normalizedPath = strings.TrimSpace(normalizedPath)

	// Remove any leading backslash that might cause issues
	normalizedPath = strings.TrimPrefix(normalizedPath, "\\")

	// Validate path doesn't contain shell metacharacters that could cause command injection
	if err := a.validatePathForShell(normalizedPath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Verify file exists before trying to open
	if _, err := os.Stat(normalizedPath); err != nil {
		return fmt.Errorf("file does not exist: %s (error: %w)", normalizedPath, err)
	}

	// Use os.StartProcess for safer execution (avoids shell interpretation)
	// On Windows, use rundll32.exe to open file with default application
	// This is safer than exec.Command as it doesn't involve shell interpretation
	absPath, err := filepath.Abs(normalizedPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Use rundll32.exe which is safer than cmd.exe /c start
	// This avoids shell metacharacter interpretation
	_, err = os.StartProcess(
		"rundll32.exe",
		[]string{"rundll32.exe", "shell32.dll,ShellExec_RunDLL", absPath},
		&os.ProcAttr{},
	)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}

// GetFileInfo returns information about a file
func (a *App) GetFileInfo(filePath string) map[string]interface{} {
	info, err := os.Stat(filePath)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"path":    filePath,
		"size":    info.Size(),
		"modTime": info.ModTime().Unix(),
		"exists":  true,
	}
}

// SaveFileFromBytes saves file data to a temporary location and returns the file path
func (a *App) SaveFileFromBytes(fileName string, fileData []byte) (string, error) {
	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "file-format-converter")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create temp file path
	tempFilePath := filepath.Join(tempDir, fileName)

	// Write file data
	if err := os.WriteFile(tempFilePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tempFilePath, nil
}

// initializeEngines initializes all conversion engines
func (a *App) initializeEngines() error {
	if err := a.initializeSpreadsheetEngine(); err != nil {
		return fmt.Errorf("failed to initialize spreadsheet engine: %w", err)
	}
	if err := a.initializeDocumentEngine(); err != nil {
		// Document engine initialization is optional - log warning but don't fail
		a.logger.Error("Could not initialize document engine", err)
	}
	return nil
}

// initializeSpreadsheetEngine initializes the spreadsheet conversion engine
func (a *App) initializeSpreadsheetEngine() error {
	if a.headlessBrowser == nil {
		browser, err := browser.NewHeadlessBrowser()
		if err != nil {
			return fmt.Errorf("failed to create headless browser: %w", err)
		}
		a.headlessBrowser = browser
	}

	htmlRenderer := spreadsheet.NewHTMLRenderer()
	pdfGenerator := spreadsheet.NewPDFGenerator(a.headlessBrowser)
	engine := spreadsheet.NewSpreadsheetEngine(htmlRenderer, pdfGenerator)

	a.spreadsheetEngine = engine

	return nil
}

// initializeDocumentEngine initializes the document conversion engine
// Uses pure Go DOCX parsing + HTML rendering + headless browser for PDF generation
func (a *App) initializeDocumentEngine() error {
	if a.headlessBrowser == nil {
		browser, err := browser.NewHeadlessBrowser()
		if err != nil {
			return fmt.Errorf("failed to create headless browser: %w", err)
		}
		a.headlessBrowser = browser
	}

	// Create document engine with pure Go approach (no WASM)
	// Uses the same headless browser as spreadsheet engine
	engine := document.NewDocumentEngine(a.headlessBrowser)
	a.documentEngine = engine

	return nil
}

// initializeImageEngine initializes the image conversion engine
func (a *App) initializeImageEngine() error {
	// Create worker pool and WebP encoder for image engine
	// For basic JPEG/PNG support, these can be nil, but we'll create them for completeness
	workerPool := image.NewWorkerPool()
	webpEncoder := image.NewWebPEncoder()

	engine := image.NewImageEngine(workerPool, webpEncoder)
	a.imageEngine = engine

	return nil
}

// initializeConverterService initializes the ConverterService with all engines
func (a *App) initializeConverterService() error {
	// Create adapters for domain interfaces
	domainLogger := logger.NewDomainLoggerAdapter(logger.LogLevelInfo)
	domainProgressNotifier := progress.NewDomainProgressNotifierAdapter()
	domainFileWriter := filesystem.NewDomainFileWriterAdapter("")

	// Build engines map
	engines := make(map[domain.FileType]domain.IConverter)

	// Add engines if they're initialized
	if a.spreadsheetEngine != nil {
		engines[domain.FileTypeXLSX] = a.spreadsheetEngine
	}
	if a.documentEngine != nil {
		engines[domain.FileTypeDOCX] = a.documentEngine
	}
	if a.imageEngine != nil {
		engines[domain.FileTypeJPEG] = a.imageEngine
		engines[domain.FileTypePNG] = a.imageEngine
		engines[domain.FileTypeWEBP] = a.imageEngine
	}

	// Create ConverterService
	a.converterService = domain.NewConverterService(
		engines,
		domainLogger,
		domainProgressNotifier,
		domainFileWriter,
	)

	return nil
}

// detectFileType detects the file type from the file extension
func (a *App) detectFileType(filePath string) domain.FileType {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".docx":
		return domain.FileTypeDOCX
	case ".xlsx":
		return domain.FileTypeXLSX
	case ".jpeg", ".jpg":
		return domain.FileTypeJPEG
	case ".png":
		return domain.FileTypePNG
	case ".webp":
		return domain.FileTypeWEBP
	default:
		return ""
	}
}

// validateDOCXFile validates a .docx file (FR-05 requirement)
// Uses the consolidated validation function from the document package
func (a *App) validateDOCXFile(filePath string) error {
	return document.ValidateDOCX(filePath)
}

// validateImageFile validates a JPEG or PNG image file (FR-08 requirement)
func (a *App) validateImageFile(filePath string, fileType domain.FileType) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Validate extension matches file type
	switch fileType {
	case domain.FileTypeJPEG:
		if ext != ".jpeg" && ext != ".jpg" {
			return fmt.Errorf("file does not have .jpeg or .jpg extension")
		}
	case domain.FileTypePNG:
		if ext != ".png" {
			return fmt.Errorf("file does not have .png extension")
		}
	default:
		return fmt.Errorf("unsupported image file type: %s", fileType)
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Validate image file by checking file signature (magic bytes)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// Read first few bytes to check file signature
	signature := make([]byte, 8)
	if _, err := file.Read(signature); err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	// Validate JPEG signature: FF D8 FF
	if fileType == domain.FileTypeJPEG {
		if signature[0] != 0xFF || signature[1] != 0xD8 || signature[2] != 0xFF {
			return fmt.Errorf("invalid JPEG file: incorrect file signature")
		}
	}

	// Validate PNG signature: 89 50 4E 47 0D 0A 1A 0A
	if fileType == domain.FileTypePNG {
		pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		for i := 0; i < 8; i++ {
			if signature[i] != pngSignature[i] {
				return fmt.Errorf("invalid PNG file: incorrect file signature")
			}
		}
	}

	return nil
}

// generateOutputPath generates an output file path based on input path and target format
func (a *App) generateOutputPath(inputPath, targetFormat string) string {
	dir := filepath.Dir(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := strings.ToLower(targetFormat)

	// Normalize JPEG format (handle both "jpg" and "jpeg")
	if ext == "jpg" {
		ext = "jpeg"
	}

	return filepath.Join(dir, baseName+"."+ext)
}

// copyFile copies a file from source to destination
func (a *App) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// DeleteTempFile deletes a temporary file (for cleanup)
func (a *App) DeleteTempFile(filePath string) error {
	// Only delete files in temp directory for safety
	tempDir := filepath.Join(os.TempDir(), "file-format-converter")
	cleanFilePath := filepath.Clean(filePath)
	cleanTempDir := filepath.Clean(tempDir)
	if !strings.HasPrefix(cleanFilePath, cleanTempDir) {
		return fmt.Errorf("file is not in temp directory, refusing to delete")
	}
	return os.Remove(filePath)
}

// CleanupTempInputFile deletes a temporary input file (alias for DeleteTempFile for frontend compatibility)
func (a *App) CleanupTempInputFile(filePath string) error {
	return a.DeleteTempFile(filePath)
}

// CleanupTempFiles removes all files in the temp directory
func (a *App) CleanupTempFiles() error {
	tempDir := filepath.Join(os.TempDir(), "file-format-converter")
	return os.RemoveAll(tempDir)
}

// scheduleTempFileCleanup schedules cleanup of a temp file after a delay
func (a *App) scheduleTempFileCleanup(filePath string) {
	// Clean up after 1 hour (or could be done on app shutdown)
	// For now, we'll clean up immediately if it's a temp file
	// In production, you might want to schedule this
	go func() {
		// Small delay to ensure file operations are complete
		time.Sleep(tempFileCleanupDelay)
		a.DeleteTempFile(filePath)
	}()
}

// validatePathForShell validates that a file path doesn't contain shell metacharacters
// that could be used for command injection attacks
func (a *App) validatePathForShell(path string) error {
	// Windows cmd.exe metacharacters that could be dangerous
	dangerousChars := []string{
		"&",  // command separator
		"|",  // pipe
		"&&", // conditional execution
		"||", // conditional execution
		";",  // command separator
		"<",  // input redirection
		">",  // output redirection
		"^",  // escape character
		"%",  // variable expansion (in some contexts)
		"`",  // command substitution (PowerShell)
		"$",  // variable expansion (PowerShell)
		"(",  // command grouping
		")",  // command grouping
	}

	// Check for dangerous character sequences
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains potentially dangerous character: %s", char)
		}
	}

	// Additional validation: ensure path doesn't start with command prefixes
	lowerPath := strings.ToLower(path)
	if strings.HasPrefix(lowerPath, "cmd") || strings.HasPrefix(lowerPath, "powershell") {
		return fmt.Errorf("path cannot start with command interpreter name")
	}

	return nil
}

// getContext returns the current window context
func (a *App) getContext() context.Context {
	// Use stored context from OnStartup
	if a.ctx != nil {
		return a.ctx
	}

	// Fallback to background context (dialog might not work, but won't crash)
	return context.Background()
}
