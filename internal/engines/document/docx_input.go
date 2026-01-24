package document

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readDOCX reads a DOCX file and returns its content
func readDOCX(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// validateDOCX validates that a file is a valid DOCX file
func validateDOCX(filePath string) error {
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	// Check if it's a regular file (not a directory)
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".docx" {
		return fmt.Errorf("invalid file extension: expected .docx, got %s", ext)
	}

	// Validate DOCX file structure (DOCX files are ZIP archives)
	// Open the file as a ZIP archive to verify it's a valid DOCX
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("invalid DOCX file structure: %w", err)
	}
	defer reader.Close()

	// Check for required DOCX structure files
	// A valid DOCX must contain at least [Content_Types].xml
	hasContentTypes := false
	for _, file := range reader.File {
		if file.Name == "[Content_Types].xml" {
			hasContentTypes = true
			break
		}
	}

	if !hasContentTypes {
		return fmt.Errorf("invalid DOCX file: missing required [Content_Types].xml")
	}

	return nil
}
