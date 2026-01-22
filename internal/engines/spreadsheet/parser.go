package spreadsheet

import (
	"github.com/xuri/excelize/v2"
)

// ExcelParser wraps excelize parser functionality
type ExcelParser struct{}

// NewExcelParser creates a new Excel parser
func NewExcelParser() *ExcelParser {
	return &ExcelParser{}
}

// Parse parses an Excel file
func (p *ExcelParser) Parse(filePath string) (*excelize.File, error) {
	return excelize.OpenFile(filePath)
}

