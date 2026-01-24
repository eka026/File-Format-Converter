package logger

import (
	"fmt"
	"os"

	"github.com/eka026/File-Format-Converter/internal/domain"
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

// DomainLoggerAdapter adapts ConsoleLogger to domain.Logger interface
type DomainLoggerAdapter struct {
	*ConsoleLogger
}

// NewDomainLoggerAdapter creates a domain Logger adapter
func NewDomainLoggerAdapter(level LogLevel) domain.Logger {
	return &DomainLoggerAdapter{
		ConsoleLogger: &ConsoleLogger{level: level},
	}
}

// Info logs an informational message
func (l *ConsoleLogger) Info(msg string) {
	if l.level <= LogLevelInfo {
		fmt.Fprintf(os.Stdout, "[INFO] %s\n", msg)
	}
}

// Error logs an error message
func (l *ConsoleLogger) Error(msg string, err error) {
	if l.level <= LogLevelError {
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %s: %v\n", msg, err)
		} else {
			fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
		}
	}
}

// Debug logs a debug message
func (l *ConsoleLogger) Debug(msg string) {
	if l.level <= LogLevelDebug {
		fmt.Fprintf(os.Stdout, "[DEBUG] %s\n", msg)
	}
}
