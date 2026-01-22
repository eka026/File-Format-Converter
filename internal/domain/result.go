package domain

import "time"

// Result represents the result of a conversion operation
type Result struct {
	Success    bool
	OutputPath string
	Error      error
	Duration   time.Duration
}

// ValidationResult represents the result of file validation
type ValidationResult struct {
	Valid   bool
	Message string
	Error   error
}

