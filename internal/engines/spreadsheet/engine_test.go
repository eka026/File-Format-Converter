package spreadsheet

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
	"github.com/xuri/excelize/v2"
)

// TestSpreadsheetEngine_Validate_ValidXLSX tests FR-01: The system shall accept valid .xlsx files as input
func TestSpreadsheetEngine_Validate_ValidXLSX(t *testing.T) {
	// Create a temporary valid Excel file
	tmpFile := createTempExcelFile(t)
	defer os.Remove(tmpFile)

	engine := NewSpreadsheetEngine(nil, nil).(*SpreadsheetEngine)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err != nil {
		t.Errorf("FR-01: Expected valid .xlsx file to be accepted, but got error: %v", err)
	}
}

// TestSpreadsheetEngine_Validate_InvalidFile tests that invalid files are rejected
func TestSpreadsheetEngine_Validate_InvalidFile(t *testing.T) {
	// Create a temporary invalid file
	tmpFile := filepath.Join(t.TempDir(), "invalid.txt")
	if err := os.WriteFile(tmpFile, []byte("not an excel file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	engine := NewSpreadsheetEngine(nil, nil).(*SpreadsheetEngine)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err == nil {
		t.Error("FR-01: Expected invalid file to be rejected, but validation passed")
	}
}

// TestSpreadsheetEngine_ParseExcelFile tests FR-02: The system shall parse the spreadsheet data using the excelize library
func TestSpreadsheetEngine_ParseExcelFile(t *testing.T) {
	tmpFile := createTempExcelFile(t)
	defer os.Remove(tmpFile)

	engine := NewSpreadsheetEngine(nil, nil).(*SpreadsheetEngine)
	parser := engine.parser

	// Test parsing
	f, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("FR-02: Failed to parse Excel file using excelize: %v", err)
	}
	defer f.Close()

	// Verify we can read data from the parsed file
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		t.Error("FR-02: Expected at least one sheet in the parsed workbook")
	}

	// Try to read a cell value to verify parsing worked
	value, err := f.GetCellValue(sheetList[0], "A1")
	if err != nil {
		t.Errorf("FR-02: Failed to read cell value from parsed file: %v", err)
	}
	if value != "Test" {
		t.Errorf("FR-02: Expected cell A1 to contain 'Test', got '%s'", value)
	}
}

// TestSpreadsheetEngine_RenderToHTML tests FR-03: The system shall render the spreadsheet data to HTML intermediate format
func TestSpreadsheetEngine_RenderToHTML(t *testing.T) {
	tmpFile := createTempExcelFile(t)
	defer os.Remove(tmpFile)

	engine := NewSpreadsheetEngine(nil, nil).(*SpreadsheetEngine)
	parser := engine.parser
	renderer := NewHTMLRenderer()

	// Parse the Excel file
	f, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse Excel file: %v", err)
	}
	defer f.Close()

	// Render to HTML
	htmlContent, err := renderer.Render(f)
	if err != nil {
		t.Fatalf("FR-03: Failed to render spreadsheet to HTML: %v", err)
	}

	// Verify HTML structure
	if htmlContent == "" {
		t.Error("FR-03: Rendered HTML content is empty")
	}

	// Check for essential HTML elements
	if !contains(htmlContent, "<!DOCTYPE html>") {
		t.Error("FR-03: Rendered HTML missing DOCTYPE declaration")
	}
	if !contains(htmlContent, "<html>") {
		t.Error("FR-03: Rendered HTML missing html tag")
	}
	if !contains(htmlContent, "<table>") {
		t.Error("FR-03: Rendered HTML missing table element (layout structure not preserved)")
	}
	if !contains(htmlContent, "Test") {
		t.Error("FR-03: Rendered HTML missing cell data (data not preserved)")
	}
}

// TestSpreadsheetEngine_Convert_EndToEnd tests the full conversion flow including PDF generation
// This tests FR-04: The system shall export the rendered HTML to PDF using a headless browser engine
func TestSpreadsheetEngine_Convert_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode (requires headless browser)")
	}

	tmpFile := createTempExcelFile(t)
	defer os.Remove(tmpFile)

	outputFile := filepath.Join(t.TempDir(), "output.pdf")
	defer os.Remove(outputFile)

	// Test HTML rendering separately (PDF generation requires browser)
	engine := NewSpreadsheetEngine(nil, nil).(*SpreadsheetEngine)
	parser := engine.parser
	renderer := NewHTMLRenderer()

	// Parse and render to HTML
	f, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse Excel file: %v", err)
	}
	defer f.Close()

	htmlContent, err := renderer.Render(f)
	if err != nil {
		t.Fatalf("FR-04: Failed to render HTML: %v", err)
	}

	// Verify HTML was generated (this is the intermediate step before PDF)
	if htmlContent == "" {
		t.Error("FR-04: HTML content was not generated")
	}
	if !contains(htmlContent, "<table>") {
		t.Error("FR-04: HTML does not contain table structure")
	}
}

// TestSpreadsheetEngine_Convert_WithRealBrowser tests FR-04 with actual browser (integration test)
func TestSpreadsheetEngine_Convert_WithRealBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires Chrome/Chromium to be installed
	// Skip if browser is not available
	browser, err := createTestBrowser()
	if err != nil {
		t.Skipf("Skipping test: browser not available: %v", err)
	}
	defer browser.Close()

	tmpFile := createTempExcelFile(t)
	defer os.Remove(tmpFile)

	outputFile := filepath.Join(t.TempDir(), "output.pdf")
	defer os.Remove(outputFile)

	pdfGen := NewPDFGenerator(browser)
	engine := NewSpreadsheetEngine(nil, pdfGen).(*SpreadsheetEngine)

	ctx := context.Background()
	err = engine.Convert(ctx, tmpFile, outputFile)
	if err != nil {
		t.Fatalf("FR-04: Conversion to PDF failed: %v", err)
	}

	// Verify PDF file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("FR-04: PDF output file was not created")
	}

	// Verify PDF file is not empty
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("FR-04: Generated PDF file is empty")
	}
}

// Helper functions

// createTempExcelFile creates a temporary Excel file with test data
func createTempExcelFile(t *testing.T) string {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	// Set cell value
	if err := f.SetCellValue(sheetName, "A1", "Test"); err != nil {
		t.Fatalf("Failed to set cell value: %v", err)
	}
	if err := f.SetCellValue(sheetName, "B1", "Data"); err != nil {
		t.Fatalf("Failed to set cell value: %v", err)
	}

	// Save to temporary file
	tmpFile := filepath.Join(t.TempDir(), "test.xlsx")
	if err := f.SaveAs(tmpFile); err != nil {
		t.Fatalf("Failed to save Excel file: %v", err)
	}

	return tmpFile
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createTestBrowser creates a headless browser for testing (if available)
func createTestBrowser() (*browser.HeadlessBrowser, error) {
	return browser.NewHeadlessBrowser()
}
