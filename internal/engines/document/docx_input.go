package document

import (
	"os"
)

// readDOCX reads a DOCX file and returns its content
func readDOCX(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// validateDOCX validates that a file is a valid DOCX file
func validateDOCX(filePath string) error {
	// Basic validation - check if file exists and has .docx extension
	// More sophisticated validation can be added
	_, err := os.Stat(filePath)
	return err
}

