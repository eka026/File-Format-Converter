package filesystem

import (
	"os"
	"path/filepath"

	"github.com/eka026/File-Format-Converter/internal/domain"
	"github.com/eka026/File-Format-Converter/internal/ports"
)

// FilesystemAdapter is the filesystem driven adapter
type FilesystemAdapter struct {
	basePath string
}

// NewFilesystemAdapter creates a new filesystem adapter
func NewFilesystemAdapter(basePath string) ports.IFileWriter {
	return &FilesystemAdapter{
		basePath: basePath,
	}
}

// DomainFileWriterAdapter adapts FilesystemAdapter to domain.FileWriter interface
type DomainFileWriterAdapter struct {
	*FilesystemAdapter
}

// NewDomainFileWriterAdapter creates a domain FileWriter adapter
func NewDomainFileWriterAdapter(basePath string) domain.FileWriter {
	return &DomainFileWriterAdapter{
		FilesystemAdapter: &FilesystemAdapter{basePath: basePath},
	}
}

// Write writes data to a file path
func (a *FilesystemAdapter) Write(path string, data []byte) error {
	fullPath := a.resolvePath(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

// Read reads data from a file path
func (a *FilesystemAdapter) Read(path string) ([]byte, error) {
	fullPath := a.resolvePath(path)
	return os.ReadFile(fullPath)
}

// Exists checks if a file exists
func (a *FilesystemAdapter) Exists(path string) bool {
	fullPath := a.resolvePath(path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// resolvePath resolves the full path
func (a *FilesystemAdapter) resolvePath(path string) string {
	if a.basePath == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(a.basePath, path)
}
