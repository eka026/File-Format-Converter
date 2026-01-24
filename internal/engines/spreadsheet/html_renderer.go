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

	for _, row := range sheet.Rows {
		buf.WriteString("<tr>")
		for _, cell := range row {
			r.renderCell(buf, &cell)
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

	if styles.Len() > 0 {
		attrs.WriteString(fmt.Sprintf(` style="%s"`, styles.String()))
	}

	buf.WriteString(fmt.Sprintf("<td%s>%s</td>", attrs.String(), html.EscapeString(cell.Value)))
}
