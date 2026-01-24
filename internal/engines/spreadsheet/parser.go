package spreadsheet

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// CellData represents parsed cell information
type CellData struct {
	Value       string
	Row         int
	Col         int
	Style       CellStyle
	IsMerged    bool
	MergeAcross int // number of columns this cell spans
	MergeDown   int // number of rows this cell spans
}

// CellStyle represents cell styling information
type CellStyle struct {
	Bold          bool
	Italic        bool
	FontSize      float64
	FontColor     string
	BackgroundColor string
	Alignment     string
	BorderStyle   string
}

// SheetData represents a parsed worksheet
type SheetData struct {
	Name       string
	Rows       [][]CellData
	MaxColumns int
	MaxRows    int
}

// WorkbookData represents a fully parsed Excel workbook
type WorkbookData struct {
	Sheets []SheetData
}

// ExcelParser wraps excelize parser functionality
type ExcelParser struct{}

// NewExcelParser creates a new Excel parser
func NewExcelParser() *ExcelParser {
	return &ExcelParser{}
}

// Parse parses an Excel file from a file path
func (p *ExcelParser) Parse(filePath string) (*excelize.File, error) {
	return excelize.OpenFile(filePath)
}

// ParseFromBytes parses an Excel file from byte data (for in-memory processing)
func (p *ExcelParser) ParseFromBytes(data []byte) (*excelize.File, error) {
	return excelize.OpenReader(bytes.NewReader(data))
}

// ParseWorkbook extracts all data from an Excel file into structured format
func (p *ExcelParser) ParseWorkbook(f *excelize.File) (*WorkbookData, error) {
	workbook := &WorkbookData{
		Sheets: make([]SheetData, 0),
	}

	sheetList := f.GetSheetList()
	for _, sheetName := range sheetList {
		sheetData, err := p.parseSheet(f, sheetName)
		if err != nil {
			return nil, fmt.Errorf("parsing sheet %s: %w", sheetName, err)
		}
		workbook.Sheets = append(workbook.Sheets, *sheetData)
	}

	return workbook, nil
}

// parseSheet extracts data from a single worksheet
func (p *ExcelParser) parseSheet(f *excelize.File, sheetName string) (*SheetData, error) {
	sheet := &SheetData{
		Name: sheetName,
		Rows: make([][]CellData, 0),
	}

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("getting rows: %w", err)
	}

	// Get merged cells for this sheet
	mergedCells, err := f.GetMergeCells(sheetName)
	if err != nil {
		return nil, fmt.Errorf("getting merged cells: %w", err)
	}

	// Build a map of merged cell ranges
	mergeMap := p.buildMergeMap(mergedCells)

	for rowIdx, row := range rows {
		rowData := make([]CellData, len(row))
		for colIdx, cellValue := range row {
			cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)

			cellData := CellData{
				Value: cellValue,
				Row:   rowIdx,
				Col:   colIdx,
			}

			// Check if this cell is part of a merge
			if merge, ok := mergeMap[cellRef]; ok {
				cellData.IsMerged = true
				cellData.MergeAcross = merge.colSpan - 1
				cellData.MergeDown = merge.rowSpan - 1
			}

			// Extract cell style
			styleID, err := f.GetCellStyle(sheetName, cellRef)
			if err == nil && styleID > 0 {
				cellData.Style = p.extractCellStyle(f, styleID)
			}

			rowData[colIdx] = cellData
		}
		sheet.Rows = append(sheet.Rows, rowData)

		if len(row) > sheet.MaxColumns {
			sheet.MaxColumns = len(row)
		}
	}

	sheet.MaxRows = len(rows)
	return sheet, nil
}

// mergeInfo holds information about a merged cell range
type mergeInfo struct {
	startCell string
	endCell   string
	colSpan   int
	rowSpan   int
	value     string
}

// buildMergeMap creates a lookup map for merged cell ranges
func (p *ExcelParser) buildMergeMap(mergedCells []excelize.MergeCell) map[string]*mergeInfo {
	mergeMap := make(map[string]*mergeInfo)

	for _, mc := range mergedCells {
		startCell := mc.GetStartAxis()
		endCell := mc.GetEndAxis()

		startCol, startRow, _ := excelize.CellNameToCoordinates(startCell)
		endCol, endRow, _ := excelize.CellNameToCoordinates(endCell)

		info := &mergeInfo{
			startCell: startCell,
			endCell:   endCell,
			colSpan:   endCol - startCol + 1,
			rowSpan:   endRow - startRow + 1,
			value:     mc.GetCellValue(),
		}

		mergeMap[startCell] = info
	}

	return mergeMap
}

// extractCellStyle extracts styling information from a cell
func (p *ExcelParser) extractCellStyle(f *excelize.File, styleID int) CellStyle {
	style := CellStyle{
		FontSize: 11, // default
	}

	styleInfo, err := f.GetStyle(styleID)
	if err != nil || styleInfo == nil {
		return style
	}

	// Extract font properties
	if styleInfo.Font != nil {
		style.Bold = styleInfo.Font.Bold
		style.Italic = styleInfo.Font.Italic
		if styleInfo.Font.Size > 0 {
			style.FontSize = styleInfo.Font.Size
		}
		if styleInfo.Font.Color != "" {
			style.FontColor = styleInfo.Font.Color
		}
	}

	// Extract fill/background color
	if styleInfo.Fill.Color != nil && len(styleInfo.Fill.Color) > 0 {
		style.BackgroundColor = styleInfo.Fill.Color[0]
	}

	// Extract alignment
	if styleInfo.Alignment != nil {
		style.Alignment = styleInfo.Alignment.Horizontal
	}

	return style
}

// GetCellValue retrieves a specific cell value with error handling
func (p *ExcelParser) GetCellValue(f *excelize.File, sheetName, cellRef string) (string, error) {
	value, err := f.GetCellValue(sheetName, cellRef)
	if err != nil {
		return "", fmt.Errorf("getting cell %s value: %w", cellRef, err)
	}
	return value, nil
}

// GetSheetNames returns all sheet names in the workbook
func (p *ExcelParser) GetSheetNames(f *excelize.File) []string {
	return f.GetSheetList()
}

// GetSheetDimension returns the used range of a sheet
func (p *ExcelParser) GetSheetDimension(f *excelize.File, sheetName string) (string, error) {
	dim, err := f.GetSheetDimension(sheetName)
	if err != nil {
		return "", fmt.Errorf("getting sheet dimension: %w", err)
	}
	return dim, nil
}
