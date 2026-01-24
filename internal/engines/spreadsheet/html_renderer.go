package spreadsheet

import (
	"bytes"
	"fmt"
	"html"

	"github.com/xuri/excelize/v2"
)

// HTMLRenderer renders Excel content as HTML
type HTMLRenderer struct {
	parser *ExcelParser
}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{
		parser: NewExcelParser(),
	}
}

// Render converts an Excel file to HTML using the parsed workbook data
func (r *HTMLRenderer) Render(f *excelize.File) (string, error) {
	workbook, err := r.parser.ParseWorkbook(f)
	if err != nil {
		return "", fmt.Errorf("parsing workbook: %w", err)
	}

	return r.RenderWorkbook(workbook), nil
}

// RenderWorkbook renders a parsed WorkbookData to HTML
func (r *HTMLRenderer) RenderWorkbook(workbook *WorkbookData) string {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body { font-family: Arial, sans-serif; margin: 20px; }
.sheet-container { margin-bottom: 40px; }
.sheet-title { font-size: 18px; font-weight: bold; margin-bottom: 10px; color: #333; }
table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
th { background-color: #4472C4; color: white; }
tr:nth-child(even) { background-color: #f9f9f9; }
tr:hover { background-color: #f5f5f5; }
.bold { font-weight: bold; }
.italic { font-style: italic; }
.align-center { text-align: center; }
.align-right { text-align: right; }
.align-left { text-align: left; }
</style>
</head>
<body>
`)

	for _, sheet := range workbook.Sheets {
		r.renderSheet(&buf, &sheet)
	}

	buf.WriteString("</body></html>")
	return buf.String()
}

// renderSheet renders a single sheet to HTML
func (r *HTMLRenderer) renderSheet(buf *bytes.Buffer, sheet *SheetData) {
	buf.WriteString(`<div class="sheet-container">`)
	buf.WriteString(fmt.Sprintf(`<div class="sheet-title">%s</div>`, html.EscapeString(sheet.Name)))
	buf.WriteString("<table>")

	// Render colgroup for column widths
	if len(sheet.ColumnWidths) > 0 {
		buf.WriteString("<colgroup>")
		for col := 0; col < sheet.MaxColumns; col++ {
			if width, ok := sheet.ColumnWidths[col]; ok {
				buf.WriteString(fmt.Sprintf(`<col style="width: %.0fpx;">`, width))
			} else {
				buf.WriteString("<col>")
			}
		}
		buf.WriteString("</colgroup>")
	}

	for rowIdx, row := range sheet.Rows {
		// Apply row height if available
		if height, ok := sheet.RowHeights[rowIdx]; ok {
			buf.WriteString(fmt.Sprintf(`<tr style="height: %.0fpx;">`, height))
		} else {
			buf.WriteString("<tr>")
		}

		// Render cells, padding to MaxColumns for consistent structure
		for colIdx := 0; colIdx < sheet.MaxColumns; colIdx++ {
			if colIdx < len(row) {
				cell := row[colIdx]
				// Skip cells that are covered by a merge
				if cell.IsMergeCovered {
					continue
				}
				r.renderCell(buf, &cell)
			} else {
				// Pad with empty cells
				buf.WriteString("<td></td>")
			}
		}
		buf.WriteString("</tr>")
	}

	buf.WriteString("</table></div>")
}

// renderCell renders a single cell with its styling
func (r *HTMLRenderer) renderCell(buf *bytes.Buffer, cell *CellData) {
	// Build cell attributes
	var attrs bytes.Buffer

	// Handle merged cells
	if cell.IsMerged {
		if cell.MergeAcross > 0 {
			attrs.WriteString(fmt.Sprintf(` colspan="%d"`, cell.MergeAcross+1))
		}
		if cell.MergeDown > 0 {
			attrs.WriteString(fmt.Sprintf(` rowspan="%d"`, cell.MergeDown+1))
		}
	}

	// Build inline styles
	var styles bytes.Buffer

	if cell.Style.Bold {
		styles.WriteString("font-weight: bold; ")
	}
	if cell.Style.Italic {
		styles.WriteString("font-style: italic; ")
	}
	if cell.Style.FontSize > 0 && cell.Style.FontSize != 11 {
		styles.WriteString(fmt.Sprintf("font-size: %.0fpt; ", cell.Style.FontSize))
	}
	if cell.Style.FontColor != "" {
		styles.WriteString(fmt.Sprintf("color: #%s; ", cell.Style.FontColor))
	}
	if cell.Style.BackgroundColor != "" {
		styles.WriteString(fmt.Sprintf("background-color: #%s; ", cell.Style.BackgroundColor))
	}
	if cell.Style.Alignment != "" {
		styles.WriteString(fmt.Sprintf("text-align: %s; ", cell.Style.Alignment))
	}
	if cell.Style.BorderStyle != "" {
		styles.WriteString(fmt.Sprintf("border: 1px %s #000; ", cell.Style.BorderStyle))
	}

	if styles.Len() > 0 {
		attrs.WriteString(fmt.Sprintf(` style="%s"`, styles.String()))
	}

	buf.WriteString(fmt.Sprintf("<td%s>%s</td>", attrs.String(), html.EscapeString(cell.Value)))
}
