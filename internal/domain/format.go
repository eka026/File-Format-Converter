package domain

// Format represents output file formats
type Format string

const (
	FormatPDF  Format = "PDF"
	FormatHTML Format = "HTML"
	FormatWEBP Format = "WEBP"
	FormatPNG  Format = "PNG"
)

// FileType represents input file types
type FileType string

const (
	FileTypeXLSX FileType = "XLSX"
	FileTypeDOCX FileType = "DOCX"
	FileTypeJPEG FileType = "JPEG"
	FileTypePNG  FileType = "PNG"
	FileTypeWEBP FileType = "WEBP"
)

