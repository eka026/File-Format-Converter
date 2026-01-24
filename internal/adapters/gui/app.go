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

	// Initialize spreadsheet engine
	if a.headlessBrowser != nil {
		if err := a.initializeEngine(); err != nil {
			fmt.Printf("Warning: Could not initialize spreadsheet engine: %v\n", err)
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
	if a.spreadsheetEngine == nil {
		// Try to initialize if not already done
		if err := a.initializeEngine(); err != nil {
			return ConversionResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to initialize conversion engine: %v", err),
			}
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

	// Convert to temp location
	err := a.spreadsheetEngine.Convert(sourcePath, tempOutputPath)
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

// initializeEngine initializes the conversion engine if not already done
func (a *App) initializeEngine() error {
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
