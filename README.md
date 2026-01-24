# File Format Converter - Local-First Open Source File Converter

A hexagonal architecture-based file conversion application built with Go, supporting multiple file format conversions.

## Architecture

This project follows hexagonal (ports and adapters) architecture principles:

- **Driving Adapters**: GUI (Wails)
- **Domain Core**: Pure Go business logic
- **Driven Adapters**: Filesystem, Headless browser

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

## Data Sovereignty

**This application is designed with data sovereignty as a core principle.** All file processing occurs entirely locally on your machine. The application:

- ✅ **Does NOT transmit** any file data to external servers
- ✅ **Does NOT send** metadata or telemetry to external services
- ✅ **Does NOT require** internet connection for file conversion
- ✅ **Processes all files** locally using only local system resources

All conversion operations are performed using local libraries and system tools. No data leaves your device during the conversion process.

## Single Binary Distribution

**This application is distributed as a single executable file with no external runtime dependencies.**

- ✅ **Single Executable**: The application builds to a single `.exe` file (Windows) with all dependencies statically linked
- ✅ **No External Runtimes**: No Python, JVM, Node.js, or other runtime environments required
- ✅ **Embedded Assets**: All frontend assets (HTML, CSS, JavaScript) are embedded in the binary using Go's `embed` package
- ✅ **Statically Linked**: All Go dependencies are compiled into the binary

**Note on PDF Generation**: PDF conversion (for Word and Excel documents) requires Chrome, Chromium, or Microsoft Edge to be installed on the system. This is a system-level dependency for the headless browser functionality, not a runtime dependency. The application will detect and use the installed browser automatically.

**Windows WebView2**: On Windows, the GUI uses WebView2, which is typically pre-installed on Windows 10/11. On older systems, it may need to be installed separately from Microsoft.

## Building

### Prerequisites

Before building, ensure you have:
- **Go** (version 1.22 or later) - [Download](https://golang.org/dl/)
- **Wails CLI** (v2.x) - Install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Node.js** (for frontend development only, not required for the final executable)

### Build the Executable

**Option 1: Using Makefile (Recommended)**
```bash
# Build the single executable (includes vulnerability scan)
make build

# The executable will be created in: build/bin/file-format-converter.exe
```

**Option 2: Using Wails directly**
```bash
# Build with optimizations (smaller binary, no debug symbols)
wails build -platform windows/amd64 -ldflags "-s -w" -trimpath

# Or simple build
wails build -platform windows/amd64

# The executable will be created in: build/bin/file-format-converter.exe
```

**Option 3: Build for specific platform**
```bash
# Build for Windows 64-bit (recommended)
wails build -platform windows/amd64

# Build for Windows 32-bit (requires 32-bit C toolchain)
wails build -platform windows/386
```

### Build Requirements


### Other Commands

```bash
# Run in development mode (with hot reload)
make run
# or
wails dev

# Run tests
make test
# or
go test ./...

# Clean build artifacts
make clean

# Install/update dependencies
make deps
```

### Build Output

After building, you'll find:
- **Executable**: `build/bin/file-format-converter.exe` (Windows)
- **Distribution folder**: `build/bin/` contains the single executable file

The executable is **self-contained** - you can copy just the `.exe` file to any Windows machine and run it (no installation required, except WebView2 on older Windows systems).

## License

MIT

