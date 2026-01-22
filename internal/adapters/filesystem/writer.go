package filesystem

import (
	"os"
	"github.com/eka026/File-Format-Converter/internal/ports"
)

// FileSystemWriter implements the FileWriter port
type FileSystemWriter struct{}

// NewFileSystemWriter creates a new filesystem writer adapter
func NewFileSystemWriter() ports.FileWriter {
	return &FileSystemWriter{}
}

// Write writes data to a file
func (w *FileSystemWriter) Write(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// WriteStream writes a stream of data chunks to a file
func (w *FileSystemWriter) WriteStream(path string, stream <-chan []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for chunk := range stream {
		if _, err := file.Write(chunk); err != nil {
			return err
		}
	}
	return nil
}

