package document

import (
	"context"
	"os"
	"github.com/openconvert/file-converter/internal/adapters/wasm"
)

// WasmBridge bridges between the document engine and the WebAssembly runtime
type WasmBridge struct {
	runtime *wasm.WazeroRuntime
}

// NewWasmBridge creates a new WebAssembly bridge
func NewWasmBridge(runtime *wasm.WazeroRuntime) *WasmBridge {
	return &WasmBridge{
		runtime: runtime,
	}
}

// Convert converts document data using pandoc.wasm
func (b *WasmBridge) Convert(docxData []byte, outputPath string) error {
	ctx := context.Background()
	
	// Execute pandoc conversion via WebAssembly
	result, err := b.runtime.Execute(ctx, "convert", docxData)
	if err != nil {
		return err
	}

	// Write result to output file
	return os.WriteFile(outputPath, result, 0644)
}

