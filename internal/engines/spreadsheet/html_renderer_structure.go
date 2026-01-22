package spreadsheet

import "html/template"

// HTMLRenderer renders data structures as HTML
type HTMLRenderer struct {
	template *template.Template
}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// Render renders data to HTML string
func (r *HTMLRenderer) Render(data interface{}) string {
	// Implementation will be added
	return ""
}

// RenderTable renders a table structure to HTML string
func (r *HTMLRenderer) RenderTable(table interface{}) string {
	// Implementation will be added
	return ""
}

