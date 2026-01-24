package spreadsheet

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// CellData represents parsed cell information
type CellData struct {
	Value          string
	Row            int
	Col            int
	Style          CellStyle
	IsMerged       bool
	MergeAcross    int  // number of columns this cell spans
	MergeDown      int  // number of rows this cell spans
	IsMergeCovered bool // true if this cell is covered by another cell's merge (should be skipped)
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
	Name         string
	Rows         [][]CellData
	MaxColumns   int
	MaxRows      int
	ColumnWidths map[int]float64 // column index -> width in pixels
	RowHeights   map[int]float64 // row index -> height in pixels
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
	mergeResult := p.buildMergeMap(mergedCells)

	// Extract column widths
	sheet.ColumnWidths = make(map[int]float64)
	cols, err := f.GetCols(sheetName)
	if err == nil {
		for colIdx := range cols {
			colName, _ := excelize.ColumnNumberToName(colIdx + 1)
			width, err := f.GetColWidth(sheetName, colName)
			if err == nil && width > 0 {
				// Convert Excel width units to approximate pixels (1 unit ≈ 7 pixels)
				sheet.ColumnWidths[colIdx] = width * 7
			}
		}
	}

	// Extract row heights
	sheet.RowHeights = make(map[int]float64)

	for rowIdx, row := range rows {
		// Get row height
		height, err := f.GetRowHeight(sheetName, rowIdx+1)
		if err == nil && height > 0 {
			sheet.RowHeights[rowIdx] = height
		}

		rowData := make([]CellData, len(row))
		for colIdx, cellValue := range row {
			cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)

			cellData := CellData{
				Value: cellValue,
				Row:   rowIdx,
				Col:   colIdx,
			}

			// Check if this cell is covered by another merge (should be skipped)
			if mergeResult.coveredCells[cellRef] {
				cellData.IsMergeCovered = true
			}

			// Check if this cell starts a merge
			if merge, ok := mergeResult.startCells[cellRef]; ok {
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

// mergeMapResult contains both start cells and covered cells
type mergeMapResult struct {
	startCells   map[string]*mergeInfo // cells that start a merge
	coveredCells map[string]bool       // cells covered by a merge (not the start)
}

// buildMergeMap creates a lookup map for merged cell ranges
func (p *ExcelParser) buildMergeMap(mergedCells []excelize.MergeCell) *mergeMapResult {
	result := &mergeMapResult{
		startCells:   make(map[string]*mergeInfo),
		coveredCells: make(map[string]bool),
	}

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

		result.startCells[startCell] = info

		// Mark all covered cells (except the start cell)
		for row := startRow; row <= endRow; row++ {
			for col := startCol; col <= endCol; col++ {
				cellRef, _ := excelize.CoordinatesToCellName(col, row)
				if cellRef != startCell {
					result.coveredCells[cellRef] = true
				}
			}
		}
	}

	return result
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

	// Extract border style (check if any border is defined)
	if styleInfo.Border != nil && len(styleInfo.Border) > 0 {
		for _, border := range styleInfo.Border {
			if border.Style > 0 {
				style.BorderStyle = mapBorderStyle(border.Style)
				break
			}
		}
	}

	return style
}

// mapBorderStyle converts Excel border style ID to CSS border style
func mapBorderStyle(styleID int) string {
	switch styleID {
	case 1: // thin
		return "solid"
	case 2: // medium
		return "solid"
	case 3: // dashed
		return "dashed"
	case 4: // dotted
		return "dotted"
	case 5: // thick
		return "solid"
	case 6: // double
		return "double"
	default:
		return "solid"
	}
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
