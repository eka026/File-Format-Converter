package domain

// IConverter defines the contract for conversion operations
type IConverter interface {
	// Convert performs the conversion from input to output
	Convert(input, output string) error
	
	// Validate checks if the input file is valid for this converter
	Validate(file string) error
}

// Format represents a file format
type Format struct {
	Extension string
	Name      string
	MimeType  string
}

