package config

// Config holds application configuration
type Config struct {
	LogLevel      string
	WorkerPoolSize int
	WasmPath      string
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		LogLevel:      "info",
		WorkerPoolSize: 0, // 0 means use NumCPU
	}
}

