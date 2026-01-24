package gui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/engines/document"
	"github.com/eka026/File-Format-Converter/internal/engines/image"
	"github.com/eka026/File-Format-Converter/internal/engines/spreadsheet"
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
	spreadsheetEngine *spreadsheet.SpreadsheetEngine
	documentEngine    domain.IConverter
	imageEngine       domain.IConverter
	headlessBrowser   *browser.HeadlessBrowser
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the application starts
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Initialize headless browser
	browser, err := browser.NewHeadlessBrowser()
	if err != nil {
		// Log error but don't fail startup - browser will be initialized lazily
		fmt.Printf("Warning: Could not initialize headless browser: %v\n", err)
	} else {
		a.headlessBrowser = browser
	}

	// Initialize engines
	if a.headlessBrowser != nil {
		if err := a.initializeEngines(); err != nil {
			fmt.Printf("Warning: Could not initialize engines: %v\n", err)
		}
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

	// Clean up all temp files on shutdown
	if err := a.CleanupTempFiles(); err != nil {
		fmt.Printf("Warning: Failed to cleanup temp files: %v\n", err)
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
	// Detect file type and validate
	fileType := a.detectFileType(sourcePath)

	// Validate .docx files (FR-05 requirement)
	if fileType == domain.FileTypeDOCX {
		// Validate DOCX file structure - this satisfies FR-05
		if err := a.validateDOCXFile(sourcePath); err != nil {
			return ConversionResult{
				Success: false,
				Error:   fmt.Sprintf("Invalid .docx file: %v", err),
			}
		}

		// Initialize document engine if not already initialized
		if a.documentEngine == nil {
			if err := a.initializeDocumentEngine(); err != nil {
				return ConversionResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to initialize document engine: %v", err),
				}
			}
		}
	} else if fileType == domain.FileTypeXLSX {
		// Handle XLSX files with spreadsheet engine
		if a.spreadsheetEngine == nil {
			if err := a.initializeSpreadsheetEngine(); err != nil {
				return ConversionResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to initialize spreadsheet engine: %v", err),
				}
			}
		}
	} else if fileType == domain.FileTypeJPEG || fileType == domain.FileTypePNG {
		// Validate image files (FR-08 requirement)
		if err := a.validateImageFile(sourcePath, fileType); err != nil {
			return ConversionResult{
				Success: false,
				Error:   fmt.Sprintf("Invalid image file: %v", err),
			}
		}

		// Initialize image engine if not already initialized
		if a.imageEngine == nil {
			if err := a.initializeImageEngine(); err != nil {
				return ConversionResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to initialize image engine: %v", err),
				}
			}
		}
	} else {
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported file type: %s", fileType),
		}
	}

	// Select appropriate engine for conversion
	var engine domain.IConverter
	switch fileType {
	case domain.FileTypeDOCX:
		engine = a.documentEngine
	case domain.FileTypeXLSX:
		engine = a.spreadsheetEngine
	case domain.FileTypeJPEG, domain.FileTypePNG:
		engine = a.imageEngine
	default:
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported file type: %s", fileType),
		}
	}

	if engine == nil {
		return ConversionResult{
			Success: false,
			Error:   "Conversion engine not available",
		}
	}

	// Track if we're using a temp file
	var tempFilePath string
	var isTempFile bool

	// If no output path provided, use default location (skip dialog to avoid WebSocket issues)
	if outputPath == "" {
		// Use default location: same directory as source file, or Downloads folder if source is in temp
		outputPath = a.generateOutputPath(sourcePath, targetFormat)

		// If source is in temp directory, save to user's Downloads folder instead
		tempDir := filepath.Join(os.TempDir(), "file-format-converter")
		if strings.HasPrefix(sourcePath, tempDir) {
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
	} else {
		// Check if output path is in temp directory
		tempDir := filepath.Join(os.TempDir(), "file-format-converter")
		if strings.HasPrefix(outputPath, tempDir) {
			isTempFile = true
			tempFilePath = outputPath
		}
	}

	// Perform conversion to temp location first (for cleanup safety)
	tempOutputPath := filepath.Join(os.TempDir(), "file-format-converter", filepath.Base(outputPath))
	if err := os.MkdirAll(filepath.Dir(tempOutputPath), 0755); err != nil {
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp directory: %v", err),
		}
	}

	// Convert to temp location using the selected engine
	err := engine.Convert(sourcePath, tempOutputPath)
	if err != nil {
		// Clean up temp file on error
		os.Remove(tempOutputPath)
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Conversion failed: %v", err),
		}
	}

	// Copy from temp to final location
	if err := a.copyFile(tempOutputPath, outputPath); err != nil {
		// Clean up temp file on error
		os.Remove(tempOutputPath)
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to save file: %v", err),
		}
	}

	// Clean up temp file after successful copy
	os.Remove(tempOutputPath)

	// If the final output is also in temp, mark it for cleanup
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

	// Detect file types for all files
	fileTypes := make([]domain.FileType, len(files))
	for i, file := range files {
		fileTypes[i] = a.detectFileType(file)
	}

	// Check if all files are of the same type
	firstType := fileTypes[0]
	allSameType := true
	for _, fileType := range fileTypes {
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
					return a.batchConvertImagesParallel(imgEngine, files, targetFormat, fileTypes)
				}
			}
		case domain.FileTypeDOCX:
			// All files are documents - use parallel document conversion
			if a.documentEngine != nil {
				if docEngine, ok := a.documentEngine.(*document.DocumentEngine); ok {
					return a.batchConvertDocumentsParallel(docEngine, files, targetFormat, fileTypes)
				}
			}
		case domain.FileTypeXLSX:
			// All files are spreadsheets - use parallel spreadsheet conversion
			if a.spreadsheetEngine != nil {
				return a.batchConvertSpreadsheetsParallel(a.spreadsheetEngine, files, targetFormat, fileTypes)
			}
		}
	}

	// Fall back to sequential processing for mixed file types or if engines are not available
	results := make([]ConversionResult, len(files))
	for i, file := range files {
		results[i] = a.ConvertFile(file, targetFormat)
	}

	return results
}

// batchConvertImagesParallel performs parallel batch conversion of images using worker pools
func (a *App) batchConvertImagesParallel(
	imgEngine *image.ImageEngine,
	files []string,
	targetFormat string,
	fileTypes []domain.FileType,
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
	fileTypes []domain.FileType,
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
	fileTypes []domain.FileType,
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
		spreadsheetEngine = a.spreadsheetEngine
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
	return []string{"pdf", "html"}
}

// OpenFile opens a file in the default system application
func (a *App) OpenFile(filePath string) error {
	// Clean and normalize the path
	normalizedPath := filepath.Clean(filePath)

	// Remove any leading/trailing whitespace
	normalizedPath = strings.TrimSpace(normalizedPath)

	// Remove any leading backslash that might cause issues
	normalizedPath = strings.TrimPrefix(normalizedPath, "\\")

	// Verify file exists before trying to open
	if _, err := os.Stat(normalizedPath); err != nil {
		return fmt.Errorf("file does not exist: %s (error: %w)", normalizedPath, err)
	}

	// On Windows, use start command
	// Use the path directly without extra quotes - exec.Command handles escaping
	cmd := exec.Command("cmd", "/c", "start", "", normalizedPath)
	if err := cmd.Run(); err != nil {
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

// GetSavePath shows a save dialog and returns the selected path
// Returns empty string if cancelled or on error (WebSocket connection issues)
// NOTE: Currently disabled due to WebSocket connection issues - uses default location instead
func (a *App) GetSavePath(defaultFileName string) string {
	// Skip dialog to avoid WebSocket crashes - return empty to use default location
	// The dialog can be re-enabled once Wails context connection is stable
	return ""

	/* Disabled dialog code - causes WebSocket crashes
	ctx := a.getContext()
	if ctx == nil {
		return ""
	}

	savePath, err := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		Title:           "Save PDF As",
		DefaultFilename: defaultFileName,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PDF Files (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})

	if err != nil {
		fmt.Printf("Warning: Save dialog failed: %v\n", err)
		return ""
	}

	return savePath
	*/
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
		fmt.Printf("Warning: Could not initialize document engine: %v\n", err)
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

	// Extract the concrete type
	if se, ok := engine.(*spreadsheet.SpreadsheetEngine); ok {
		a.spreadsheetEngine = se
	} else {
		return fmt.Errorf("failed to cast engine to SpreadsheetEngine")
	}

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
func (a *App) validateDOCXFile(filePath string) error {
	// Import document package validation function
	// We'll use a simple validation approach that doesn't require the full engine
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".docx" {
		return fmt.Errorf("file does not have .docx extension")
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Basic structure validation - check if it's a valid ZIP (DOCX files are ZIP archives)
	// We'll do a minimal check by trying to read it as a ZIP
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// Check ZIP file signature (first 4 bytes should be "PK\x03\x04")
	signature := make([]byte, 4)
	if _, err := file.Read(signature); err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	if signature[0] != 'P' || signature[1] != 'K' || signature[2] != 0x03 || signature[3] != 0x04 {
		return fmt.Errorf("invalid DOCX file: not a valid ZIP archive")
	}

	return nil
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
	if !strings.HasPrefix(filePath, tempDir) {
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
		time.Sleep(5 * time.Second)
		a.DeleteTempFile(filePath)
	}()
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
