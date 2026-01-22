# File Format Converter - Local-First Open Source File Converter

A hexagonal architecture-based file conversion application built with Go, supporting multiple file format conversions.

## Architecture

This project follows hexagonal (ports and adapters) architecture principles:

- **Driving Adapters**: GUI (Wails) and CLI (Cobra)
- **Domain Core**: Pure Go business logic
- **Driven Adapters**: Filesystem, WebAssembly runtime, Headless browser

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── cli/               # CLI adapter (Cobra)
│   └── gui/               # GUI adapter (Wails)
├── internal/
│   ├── domain/            # Core business logic
│   ├── ports/             # Input and output ports
│   ├── adapters/          # Driving and driven adapters
│   └── engines/           # Conversion engines
├── web/                   # Wails frontend (HTML/CSS/JS)
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
# CLI
go build -o bin/file-format-converter-cli ./cmd/cli

# GUI
wails build
```

## License

MIT

