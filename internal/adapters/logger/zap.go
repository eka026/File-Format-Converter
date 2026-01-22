package logger

import (
	"go.uber.org/zap"
	"github.com/openconvert/file-converter/internal/ports"
)

// ZapLogger implements the Logger port using zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new zap logger adapter
func NewZapLogger() (ports.Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger: zapLogger}, nil
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, convertFields(fields)...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, err error, fields ...interface{}) {
	allFields := append([]zap.Field{zap.Error(err)}, convertFields(fields)...)
	l.logger.Error(msg, allFields...)
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, convertFields(fields)...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, convertFields(fields)...)
}

func convertFields(fields []interface{}) []zap.Field {
	// Simple conversion - can be enhanced
	return []zap.Field{}
}

