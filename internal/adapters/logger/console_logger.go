package logger

import (
	"github.com/eka026/File-Format-Converter/internal/ports"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelError
)

// ConsoleLogger is the console logging driven adapter
type ConsoleLogger struct {
	level LogLevel
}

// NewConsoleLogger creates a new console logger adapter
func NewConsoleLogger(level LogLevel) ports.ILogger {
	return &ConsoleLogger{
		level: level,
	}
}

// Info logs an informational message
func (l *ConsoleLogger) Info(msg string) {
	// Implementation will be added
}

// Error logs an error message
func (l *ConsoleLogger) Error(msg string, err error) {
	// Implementation will be added
}

// Debug logs a debug message
func (l *ConsoleLogger) Debug(msg string) {
	// Implementation will be added
}

