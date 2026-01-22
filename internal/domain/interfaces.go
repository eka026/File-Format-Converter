package domain

// Logger defines the logging interface needed by the domain
type Logger interface {
	Info(msg string)
	Error(msg string, err error)
	Debug(msg string)
}

// FileWriter defines the file writing interface needed by the domain
type FileWriter interface {
	Write(path string, data []byte) error
	Read(path string) ([]byte, error)
	Exists(path string) bool
}

// ProgressNotifier defines the progress notification interface needed by the domain
type ProgressNotifier interface {
	NotifyProgress(pct int, msg string)
	NotifyComplete(result Result)
	NotifyError(err error)
}

