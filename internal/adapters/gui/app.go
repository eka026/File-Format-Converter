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

		// For now, return success for validation
		// Conversion will be implemented separately
		return ConversionResult{
			Success: false,
			Error:   "DOCX file validation passed, but conversion is not yet implemented",
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
				outputPath = filepath.Join(downloadsDir, baseName+"."+strings.ToLower(targetFormat))
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
func (a *App) BatchConvertFiles(files []string, targetFormat string) []ConversionResult {
	results := make([]ConversionResult, len(files))

	for i, file := range files {
		results[i] = a.ConvertFile(file, targetFormat)
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
// Note: This is currently disabled due to conflicting NewDocumentEngine implementations
// in the document package. Validation is handled separately in validateDOCXFile.
func (a *App) initializeDocumentEngine() error {
	// TODO: Resolve conflict between document/engine.go and document/document_engine.go
	// Both have NewDocumentEngine with different signatures
	// For now, document engine initialization is skipped
	// Validation is handled by validateDOCXFile which satisfies FR-05

	return fmt.Errorf("document engine initialization not yet implemented due to package conflicts")
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

// generateOutputPath generates an output file path based on input path and target format
func (a *App) generateOutputPath(inputPath, targetFormat string) string {
	dir := filepath.Dir(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := strings.ToLower(targetFormat)

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
