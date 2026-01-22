package spreadsheet

import (
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
)

// HTMLRenderer renders Excel content as HTML
type HTMLRenderer struct{}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// Render converts an Excel file to HTML
func (r *HTMLRenderer) Render(f *excelize.File) (string, error) {
	var html bytes.Buffer
	html.WriteString("<html><head><meta charset='UTF-8'></head><body><table>")

	// Get all sheet names
	sheetList := f.GetSheetList()
	for _, sheetName := range sheetList {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return "", err
		}

		for _, row := range rows {
			html.WriteString("<tr>")
			for _, cell := range row {
				html.WriteString(fmt.Sprintf("<td>%s</td>", cell))
			}
			html.WriteString("</tr>")
		}
	}

	html.WriteString("</table></body></html>")
	return html.String(), nil
}

