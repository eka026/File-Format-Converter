package domain

import "context"

// IConverter defines the contract for conversion operations
type IConverter interface {
	// Convert performs the conversion from input to output
	Convert(ctx context.Context, input, output string) error

	// Validate checks if the input file is valid for this converter
	Validate(ctx context.Context, file string) error
}
