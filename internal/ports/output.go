package ports

// FileWriter defines the output port for writing files
type FileWriter interface {
	Write(path string, data []byte) error
	WriteStream(path string, stream <-chan []byte) error
}

// Logger defines the output port for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// ProgressNotifier defines the output port for progress notifications
type ProgressNotifier interface {
	NotifyProgress(file string, percentage int)
	NotifyComplete(file string)
	NotifyError(file string, err error)
}


