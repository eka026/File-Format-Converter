# WebAssembly Modules

This directory contains WebAssembly modules used by the application.

## pandoc.wasm

The `pandoc.wasm` file should be placed in this directory. It will be embedded into the Go binary using `//go:embed` directives.

To obtain pandoc.wasm:
1. Download from the official Pandoc WebAssembly releases
2. Place the file in this directory
3. The file will be automatically embedded during build

