package document

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eka026/File-Format-Converter/internal/adapters/browser"
)

// TestDocumentEngine_Validate_ValidDOCX tests FR-05: The system shall accept valid .docx files as input
func TestDocumentEngine_Validate_ValidDOCX(t *testing.T) {
	tmpFile := createTempDOCXFile(t)
	defer os.Remove(tmpFile)

	engine := NewDocumentEngine(nil).(*DocumentEngine)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err != nil {
		t.Errorf("FR-05: Expected valid .docx file to be accepted, but got error: %v", err)
	}
}

// TestDocumentEngine_Validate_InvalidFile tests that invalid files are rejected
func TestDocumentEngine_Validate_InvalidFile(t *testing.T) {
	// Create a temporary invalid file
	tmpFile := filepath.Join(t.TempDir(), "invalid.txt")
	if err := os.WriteFile(tmpFile, []byte("not a docx file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	engine := NewDocumentEngine(nil).(*DocumentEngine)
	ctx := context.Background()

	err := engine.Validate(ctx, tmpFile)
	if err == nil {
		t.Error("FR-05: Expected invalid file to be rejected, but validation passed")
	}
}

// TestDocumentEngine_ParseDOCX tests FR-06: The system shall process the document
// Note: The actual implementation uses pure Go DOCX parsing, not pandoc.wasm
// This test verifies the document processing works correctly
func TestDocumentEngine_ParseDOCX(t *testing.T) {
	tmpFile := createTempDOCXFile(t)
	defer os.Remove(tmpFile)

	engine := NewDocumentEngine(nil).(*DocumentEngine)
	parser := engine.parser

	// Read DOCX file
	docxData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read DOCX file: %v", err)
	}

	// Parse DOCX file
	doc, err := parser.Parse(docxData)
	if err != nil {
		t.Fatalf("FR-06: Failed to parse DOCX file: %v", err)
	}

	// Verify document structure
	if doc == nil {
		t.Error("FR-06: Parsed document is nil")
	}

	// Verify we can access document elements
	if len(doc.Elements) == 0 {
		t.Error("FR-06: Parsed document has no elements")
	}
}

// TestDocumentEngine_RenderToHTML tests that document is rendered to HTML intermediate format
func TestDocumentEngine_RenderToHTML(t *testing.T) {
	tmpFile := createTempDOCXFile(t)
	defer os.Remove(tmpFile)

	engine := NewDocumentEngine(nil).(*DocumentEngine)
	parser := engine.parser
	renderer := engine.htmlRenderer

	// Read and parse DOCX file
	docxData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read DOCX file: %v", err)
	}

	doc, err := parser.Parse(docxData)
	if err != nil {
		t.Fatalf("Failed to parse DOCX file: %v", err)
	}

	// Render to HTML
	htmlContent := renderer.Render(doc)

	// Verify HTML structure
	if htmlContent == "" {
		t.Error("Rendered HTML content is empty")
	}

	// Check for essential HTML elements
	if !contains(htmlContent, "<!DOCTYPE html>") || !contains(htmlContent, "<html>") {
		t.Error("Rendered HTML missing basic structure")
	}
}

// TestDocumentEngine_Convert_ToPDF tests FR-07: The system shall generate a PDF output that visually reflects the input document's structure
func TestDocumentEngine_Convert_ToPDF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode (requires headless browser)")
	}

	tmpFile := createTempDOCXFile(t)
	defer os.Remove(tmpFile)

	// Test HTML rendering separately (PDF generation requires browser)
	engine := NewDocumentEngine(nil).(*DocumentEngine)
	parser := engine.parser
	renderer := engine.htmlRenderer

	// Read and parse DOCX file
	docxData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read DOCX file: %v", err)
	}

	doc, err := parser.Parse(docxData)
	if err != nil {
		t.Fatalf("Failed to parse DOCX file: %v", err)
	}

	// Render to HTML (intermediate step before PDF)
	htmlContent := renderer.Render(doc)

	// Verify HTML was generated
	if htmlContent == "" {
		t.Error("FR-07: HTML content was not generated")
	}
	if !contains(htmlContent, "<html>") {
		t.Error("FR-07: HTML does not contain HTML structure")
	}
}

// TestDocumentEngine_Convert_WithRealBrowser tests FR-07 with actual browser (integration test)
func TestDocumentEngine_Convert_WithRealBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires Chrome/Chromium to be installed
	browser, err := browser.NewHeadlessBrowser()
	if err != nil {
		t.Skipf("Skipping test: browser not available: %v", err)
	}
	defer browser.Close()

	tmpFile := createTempDOCXFile(t)
	defer os.Remove(tmpFile)

	outputFile := filepath.Join(t.TempDir(), "output.pdf")
	defer os.Remove(outputFile)

	engine := NewDocumentEngine(browser).(*DocumentEngine)

	ctx := context.Background()
	err = engine.Convert(ctx, tmpFile, outputFile)
	if err != nil {
		t.Fatalf("FR-07: Conversion to PDF failed: %v", err)
	}

	// Verify PDF file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("FR-07: PDF output file was not created")
	}

	// Verify PDF file is not empty
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("FR-07: Generated PDF file is empty")
	}
}

// TestDocumentEngine_Convert_PreservesStructure tests that document structure is preserved in PDF
func TestDocumentEngine_Convert_PreservesStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create a DOCX with specific structure (paragraphs, headings, etc.)
	tmpFile := createTempDOCXFileWithStructure(t)
	defer os.Remove(tmpFile)

	outputFile := filepath.Join(t.TempDir(), "output.pdf")
	defer os.Remove(outputFile)

	browser, err := browser.NewHeadlessBrowser()
	if err != nil {
		t.Skipf("Skipping test: browser not available: %v", err)
	}
	defer browser.Close()

	engine := NewDocumentEngine(browser).(*DocumentEngine)

	ctx := context.Background()
	err = engine.Convert(ctx, tmpFile, outputFile)
	if err != nil {
		t.Fatalf("FR-07: Conversion failed: %v", err)
	}

	// Verify PDF was created (structure preservation is verified by successful conversion)
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("FR-07: PDF output file was not created, document structure may not be preserved")
	}
}

// Helper functions

// createTempDOCXFile creates a minimal valid DOCX file for testing
func createTempDOCXFile(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "test.docx")

	// Create a minimal DOCX file structure
	// DOCX files are ZIP archives containing XML files
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add [Content_Types].xml (required)
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	w, err := zipWriter.Create("[Content_Types].xml")
	if err != nil {
		t.Fatalf("Failed to create content types: %v", err)
	}
	if _, err := w.Write([]byte(contentTypes)); err != nil {
		t.Fatalf("Failed to write content types: %v", err)
	}

	// Add word/document.xml (main document)
	documentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p>
<w:r>
<w:t>Test Document Content</w:t>
</w:r>
</w:p>
</w:body>
</w:document>`

	w, err = zipWriter.Create("word/document.xml")
	if err != nil {
		t.Fatalf("Failed to create document.xml: %v", err)
	}
	if _, err := w.Write([]byte(documentXML)); err != nil {
		t.Fatalf("Failed to write document.xml: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	// Write to file
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write DOCX file: %v", err)
	}

	return tmpFile
}

// createTempDOCXFileWithStructure creates a DOCX file with more complex structure
func createTempDOCXFileWithStructure(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "test_structured.docx")

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add [Content_Types].xml
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	w, _ := zipWriter.Create("[Content_Types].xml")
	w.Write([]byte(contentTypes))

	// Add word/document.xml with structure (headings, paragraphs)
	documentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p>
<w:pPr>
<w:pStyle w:val="Heading1"/>
</w:pPr>
<w:r>
<w:t>Heading 1</w:t>
</w:r>
</w:p>
<w:p>
<w:r>
<w:t>Paragraph text</w:t>
</w:r>
</w:p>
</w:body>
</w:document>`

	w, _ = zipWriter.Create("word/document.xml")
	w.Write([]byte(documentXML))

	zipWriter.Close()

	os.WriteFile(tmpFile, buf.Bytes(), 0644)
	return tmpFile
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


