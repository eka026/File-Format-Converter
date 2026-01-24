package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// DocxDocument represents a parsed DOCX document
type DocxDocument struct {
	Elements []DocumentElement
}

// DocumentElement represents any element in the document (paragraph, list, table)
type DocumentElement struct {
	Type      ElementType
	Paragraph *Paragraph
	List      *List
	Table     *Table
}

// ElementType represents the type of document element
type ElementType int

const (
	ElementTypeParagraph ElementType = iota
	ElementTypeList
	ElementTypeTable
)

// Paragraph represents a paragraph in the document
type Paragraph struct {
	Text      string
	Runs      []TextRun
	Style     string
	HeadingLevel int // 0 = normal paragraph, 1-6 = heading levels
	Alignment string // left, center, right, justify
}

// TextRun represents a formatted text run within a paragraph
type TextRun struct {
	Text        string
	IsBold      bool
	IsItalic    bool
	IsUnderline bool
	IsStrike    bool
	FontSize    float64
	FontColor   string
}

// List represents a list (ordered or unordered)
type List struct {
	Items      []ListItem
	IsOrdered  bool
	Level      int // nesting level
}

// ListItem represents a single list item
type ListItem struct {
	Text      string
	Runs      []TextRun
	SubItems  []ListItem // for nested lists
}

// Table represents a table in the document
type Table struct {
	Rows []TableRow
}

// TableRow represents a row in a table
type TableRow struct {
	Cells []TableCell
}

// TableCell represents a cell in a table
type TableCell struct {
	Text      string
	Runs      []TextRun
	ColSpan   int
	RowSpan   int
}

// DocxParser parses DOCX files
type DocxParser struct{}

// NewDocxParser creates a new DOCX parser
func NewDocxParser() *DocxParser {
	return &DocxParser{}
}

// Parse parses a DOCX file from bytes
func (p *DocxParser) Parse(data []byte) (*DocxDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening docx as zip: %w", err)
	}

	// Find and read the main document XML
	var docXML []byte
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("opening document.xml: %w", err)
			}

			docXML, err = io.ReadAll(rc)
			rc.Close() // Close immediately after reading
			if err != nil {
				return nil, fmt.Errorf("reading document.xml: %w", err)
			}
			break
		}
	}

	if docXML == nil {
		return nil, fmt.Errorf("document.xml not found in docx file")
	}

	return p.parseDocumentXML(docXML)
}

// parseDocumentXML parses the Word document XML structure
func (p *DocxParser) parseDocumentXML(xmlData []byte) (*DocxDocument, error) {
	var doc DocxDocument

	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	var currentParagraph *Paragraph
	var currentTable *Table
	var currentTableRow *TableRow
	var currentTableCell *TableCell
	var currentRun TextRun
	var currentText strings.Builder
	
	inParagraph := false
	inTable := false
	inTableRow := false
	inTableCell := false
	inRun := false
	inText := false
	inRunProperties := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing XML: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "p": // Paragraph
				inParagraph = true
				currentParagraph = &Paragraph{
					Runs:        []TextRun{},
					Alignment:   "left",
					HeadingLevel: 0,
				}

			case "pPr": // Paragraph properties
				// This will be processed when we encounter child elements

			case "pStyle": // Paragraph style (child of pPr)
				if currentParagraph != nil {
					// Extract style name from val attribute
					for _, attr := range se.Attr {
						if attr.Name.Local == "val" {
							styleName := attr.Value
							currentParagraph.Style = styleName
							// Check if it's a heading style
							if strings.HasPrefix(styleName, "Heading") {
								// Extract heading level (e.g., "Heading1" -> 1, "Heading 1" -> 1)
								styleName = strings.ReplaceAll(styleName, " ", "")
								if len(styleName) > 7 {
									level := int(styleName[7] - '0')
									if level >= 1 && level <= 6 {
										currentParagraph.HeadingLevel = level
									}
								}
							} else if strings.HasPrefix(styleName, "heading") {
								// Handle lowercase "heading"
								styleName = strings.ReplaceAll(styleName, " ", "")
								if len(styleName) > 7 {
									level := int(styleName[7] - '0')
									if level >= 1 && level <= 6 {
										currentParagraph.HeadingLevel = level
									}
								}
							}
							break
						}
					}
				}

			case "jc": // Justification/alignment (child of pPr)
				if currentParagraph != nil {
					// Extract alignment from val attribute
					for _, attr := range se.Attr {
						if attr.Name.Local == "val" {
							switch attr.Value {
							case "center":
								currentParagraph.Alignment = "center"
							case "right":
								currentParagraph.Alignment = "right"
							case "both":
								currentParagraph.Alignment = "justify"
							default:
								currentParagraph.Alignment = "left"
							}
							break
						}
					}
				}

			case "tbl": // Table
				inTable = true
				currentTable = &Table{Rows: []TableRow{}}

			case "tr": // Table row
				if inTable {
					inTableRow = true
					currentTableRow = &TableRow{Cells: []TableCell{}}
				}

			case "tc": // Table cell
				if inTableRow {
					inTableCell = true
					currentTableCell = &TableCell{
						Runs:   []TextRun{},
						ColSpan: 1,
						RowSpan: 1,
					}
				}

			case "r": // Run (text with formatting)
				inRun = true
				currentRun = TextRun{}
				currentText.Reset()

			case "rPr": // Run properties (formatting)
				inRunProperties = true

			case "b": // Bold property
				if inRunProperties {
					currentRun.IsBold = true
				}

			case "i": // Italic property
				if inRunProperties {
					currentRun.IsItalic = true
				}

			case "u": // Underline property
				if inRunProperties {
					currentRun.IsUnderline = true
				}

			case "strike": // Strikethrough property
				if inRunProperties {
					currentRun.IsStrike = true
				}

			case "sz": // Font size
				if inRunProperties {
					for _, attr := range se.Attr {
						if attr.Name.Local == "val" {
							// Font size is in half-points, convert to points
							if size, err := parseInt(attr.Value); err == nil {
								currentRun.FontSize = float64(size) / 2.0
							}
						}
					}
				}

			case "color": // Font color
				if inRunProperties {
					for _, attr := range se.Attr {
						if attr.Name.Local == "val" {
							currentRun.FontColor = attr.Value
						}
					}
				}

			case "t": // Text node
				inText = true

			case "br": // Line break
				if inRun {
					currentText.WriteString("\n")
				}
			}

		case xml.EndElement:
			switch se.Name.Local {
			case "p": // End of paragraph
				if inParagraph && currentParagraph != nil {
					// Collect all text from runs
					var paraText strings.Builder
					for _, run := range currentParagraph.Runs {
						paraText.WriteString(run.Text)
					}
					currentParagraph.Text = strings.TrimSpace(paraText.String())
					
					if currentParagraph.Text != "" || len(currentParagraph.Runs) > 0 {
						doc.Elements = append(doc.Elements, DocumentElement{
							Type:      ElementTypeParagraph,
							Paragraph: currentParagraph,
						})
					}
					inParagraph = false
					currentParagraph = nil
				}

			case "r": // End of run
				if inRun {
					currentRun.Text = currentText.String()
					if inParagraph && currentParagraph != nil {
						currentParagraph.Runs = append(currentParagraph.Runs, currentRun)
					} else if inTableCell && currentTableCell != nil {
						currentTableCell.Runs = append(currentTableCell.Runs, currentRun)
					}
					inRun = false
					currentRun = TextRun{}
					currentText.Reset()
				}
				inRunProperties = false

			case "rPr": // End of run properties
				inRunProperties = false

			case "t": // End of text node
				inText = false

			case "tc": // End of table cell
				if inTableCell && currentTableCell != nil {
					// Collect text from runs
					var cellText strings.Builder
					for _, run := range currentTableCell.Runs {
						cellText.WriteString(run.Text)
					}
					currentTableCell.Text = strings.TrimSpace(cellText.String())
					
					if currentTableRow != nil {
						currentTableRow.Cells = append(currentTableRow.Cells, *currentTableCell)
					}
					inTableCell = false
					currentTableCell = nil
				}

			case "tr": // End of table row
				if inTableRow && currentTableRow != nil {
					if currentTable != nil {
						currentTable.Rows = append(currentTable.Rows, *currentTableRow)
					}
					inTableRow = false
					currentTableRow = nil
				}

			case "tbl": // End of table
				if inTable && currentTable != nil && len(currentTable.Rows) > 0 {
					doc.Elements = append(doc.Elements, DocumentElement{
						Type:  ElementTypeTable,
						Table: currentTable,
					})
					inTable = false
					currentTable = nil
				}
			}

		case xml.CharData:
			if inText && inRun {
				currentText.Write(se)
			}
		}
	}

	return &doc, nil
}

// parseInt parses an integer string, handling common formats
func parseInt(s string) (int, error) {
	// Remove any non-numeric prefix/suffix if needed
	s = strings.TrimSpace(s)
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

