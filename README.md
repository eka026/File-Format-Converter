# File Format Converter - Local-First Open Source File Converter

A hexagonal architecture-based file conversion application built with Go, supporting multiple file format conversions.

## Architecture

This project follows hexagonal (ports and adapters) architecture principles:

- **Driving Adapters**: GUI (Wails)
- **Domain Core**: Pure Go business logic
- **Driven Adapters**: Filesystem, WebAssembly runtime, Headless browser

## Project Structure

```
.
├── main.go                # Application entry point (Wails GUI)
├── internal/
│   ├── domain/            # Core business logic
│   ├── ports/             # Input and output ports
│   ├── adapters/          # Driving and driven adapters
│   └── engines/           # Conversion engines
├── web/                   # Wails frontend (HTML/CSS/JS)
├── frontend/              # Wails-generated TypeScript bindings
├── wasm/                  # WebAssembly modules
└── docs/                  # Documentation

```

## Features

- Spreadsheet conversion (Excel → PDF)
- Image format conversion
- Document conversion (Word → PDF)
- Batch processing
- Progress tracking

## Building

```bash
# Build GUI application
wails build

# Run in development mode
wails dev

# Or use Makefile
make build    # Build
make run      # Run in dev mode
make test     # Run tests
make clean    # Clean build artifacts
```

## License

MIT

