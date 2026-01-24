package document

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

// HTMLRenderer renders DOCX documents as HTML
type HTMLRenderer struct{}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// Render converts a DocxDocument to HTML
func (r *HTMLRenderer) Render(doc *DocxDocument) string {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body {
	font-family: 'Segoe UI', Arial, sans-serif;
	margin: 40px;
	line-height: 1.6;
	color: #333;
	max-width: 800px;
}
p {
	margin: 12px 0;
}
h1 {
	font-size: 2em;
	font-weight: bold;
	margin: 20px 0 12px 0;
	color: #1a1a1a;
}
h2 {
	font-size: 1.75em;
	font-weight: bold;
	margin: 18px 0 10px 0;
	color: #1a1a1a;
}
h3 {
	font-size: 1.5em;
	font-weight: bold;
	margin: 16px 0 8px 0;
	color: #1a1a1a;
}
h4 {
	font-size: 1.25em;
	font-weight: bold;
	margin: 14px 0 6px 0;
	color: #1a1a1a;
}
h5 {
	font-size: 1.1em;
	font-weight: bold;
	margin: 12px 0 6px 0;
	color: #1a1a1a;
}
h6 {
	font-size: 1em;
	font-weight: bold;
	margin: 10px 0 4px 0;
	color: #1a1a1a;
}
.text-left {
	text-align: left;
}
.text-center {
	text-align: center;
}
.text-right {
	text-align: right;
}
.text-justify {
	text-align: justify;
}
.bold {
	font-weight: bold;
}
.italic {
	font-style: italic;
}
.underline {
	text-decoration: underline;
}
.strike {
	text-decoration: line-through;
}
ul, ol {
	margin: 12px 0;
	padding-left: 30px;
}
li {
	margin: 4px 0;
}
table {
	border-collapse: collapse;
	width: 100%;
	margin: 12px 0;
	border: 1px solid #ddd;
}
td, th {
	border: 1px solid #ddd;
	padding: 8px;
	text-align: left;
}
th {
	background-color: #f2f2f2;
	font-weight: bold;
}
</style>
</head>
<body>
`)

	for _, elem := range doc.Elements {
		switch elem.Type {
		case ElementTypeParagraph:
			r.renderParagraph(&buf, elem.Paragraph)
		case ElementTypeList:
			r.renderList(&buf, elem.List)
		case ElementTypeTable:
			r.renderTable(&buf, elem.Table)
		}
	}

	buf.WriteString("</body></html>")
	return buf.String()
}

// renderParagraph renders a paragraph element
func (r *HTMLRenderer) renderParagraph(buf *bytes.Buffer, para *Paragraph) {
	if para == nil {
		return
	}

	// Determine tag based on heading level
	tag := "p"
	if para.HeadingLevel > 0 && para.HeadingLevel <= 6 {
		tag = fmt.Sprintf("h%d", para.HeadingLevel)
	}

	// Build alignment class
	alignClass := ""
	switch para.Alignment {
	case "center":
		alignClass = "text-center"
	case "right":
		alignClass = "text-right"
	case "justify":
		alignClass = "text-justify"
	default:
		alignClass = "text-left"
	}

	buf.WriteString("<")
	buf.WriteString(tag)
	if alignClass != "" {
		buf.WriteString(` class="`)
		buf.WriteString(alignClass)
		buf.WriteString(`"`)
	}
	buf.WriteString(">")

	// Render text runs with formatting
	if len(para.Runs) > 0 {
		for _, run := range para.Runs {
			r.renderTextRun(buf, run)
		}
	} else if para.Text != "" {
		// Fallback to plain text if no runs
		buf.WriteString(html.EscapeString(para.Text))
	}

	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteString(">\n")
}

// renderTextRun renders a formatted text run
func (r *HTMLRenderer) renderTextRun(buf *bytes.Buffer, run TextRun) {
	if run.Text == "" {
		return
	}

	// Build inline styles
	var styles []string
	if run.FontSize > 0 {
		styles = append(styles, fmt.Sprintf("font-size: %.1fpt", run.FontSize))
	}
	if run.FontColor != "" {
		// Convert hex color if needed (DOCX uses hex without #)
		color := run.FontColor
		if len(color) == 6 {
			color = "#" + color
		}
		styles = append(styles, fmt.Sprintf("color: %s", color))
	}

	// Build classes
	var classes []string
	if run.IsBold {
		classes = append(classes, "bold")
	}
	if run.IsItalic {
		classes = append(classes, "italic")
	}
	if run.IsUnderline {
		classes = append(classes, "underline")
	}
	if run.IsStrike {
		classes = append(classes, "strike")
	}

	// Determine if we need a span
	needsSpan := len(styles) > 0 || len(classes) > 0

	if needsSpan {
		buf.WriteString("<span")
		if len(classes) > 0 {
			buf.WriteString(` class="`)
			for i, class := range classes {
				if i > 0 {
					buf.WriteString(" ")
				}
				buf.WriteString(class)
			}
			buf.WriteString(`"`)
		}
		if len(styles) > 0 {
			buf.WriteString(` style="`)
			for i, style := range styles {
				if i > 0 {
					buf.WriteString("; ")
				}
				buf.WriteString(style)
			}
			buf.WriteString(`"`)
		}
		buf.WriteString(">")
	}

	// Escape and write text, preserving line breaks
	text := html.EscapeString(run.Text)
	text = strings.ReplaceAll(text, "\n", "<br>")
	buf.WriteString(text)

	if needsSpan {
		buf.WriteString("</span>")
	}
}

// renderList renders a list element
func (r *HTMLRenderer) renderList(buf *bytes.Buffer, list *List) {
	if list == nil || len(list.Items) == 0 {
		return
	}

	tag := "ul"
	if list.IsOrdered {
		tag = "ol"
	}

	buf.WriteString("<")
	buf.WriteString(tag)
	buf.WriteString(">\n")

	for _, item := range list.Items {
		r.renderListItem(buf, item)
	}

	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteString(">\n")
}

// renderListItem renders a list item
func (r *HTMLRenderer) renderListItem(buf *bytes.Buffer, item ListItem) {
	buf.WriteString("<li>")

	if len(item.Runs) > 0 {
		for _, run := range item.Runs {
			r.renderTextRun(buf, run)
		}
	} else if item.Text != "" {
		buf.WriteString(html.EscapeString(item.Text))
	}

	// Render nested sub-items
	if len(item.SubItems) > 0 {
		buf.WriteString("<ul>\n")
		for _, subItem := range item.SubItems {
			r.renderListItem(buf, subItem)
		}
		buf.WriteString("</ul>\n")
	}

	buf.WriteString("</li>\n")
}

// renderTable renders a table element
func (r *HTMLRenderer) renderTable(buf *bytes.Buffer, table *Table) {
	if table == nil || len(table.Rows) == 0 {
		return
	}

	buf.WriteString("<table>\n")

	for i, row := range table.Rows {
		buf.WriteString("<tr>\n")
		for _, cell := range row.Cells {
			tag := "td"
			// First row could be header (simple heuristic)
			if i == 0 {
				tag = "th"
			}

			buf.WriteString("<")
			buf.WriteString(tag)
			if cell.ColSpan > 1 {
				buf.WriteString(fmt.Sprintf(` colspan="%d"`, cell.ColSpan))
			}
			if cell.RowSpan > 1 {
				buf.WriteString(fmt.Sprintf(` rowspan="%d"`, cell.RowSpan))
			}
			buf.WriteString(">")

			if len(cell.Runs) > 0 {
				for _, run := range cell.Runs {
					r.renderTextRun(buf, run)
				}
			} else if cell.Text != "" {
				buf.WriteString(html.EscapeString(cell.Text))
			}

			buf.WriteString("</")
			buf.WriteString(tag)
			buf.WriteString(">\n")
		}
		buf.WriteString("</tr>\n")
	}

	buf.WriteString("</table>\n")
}

