.PHONY: build-cli build-gui build-all clean test run-cli run-gui

# Build CLI
build-cli:
	@echo "Building CLI..."
	@go build -o bin/file-format-converter-cli ./cmd/cli

# Build GUI (requires Wails)
build-gui:
	@echo "Building GUI..."
	@wails build

# Build both
build-all: build-cli build-gui

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf build/
	@rm -rf dist/

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run CLI
run-cli: build-cli
	@./bin/file-format-converter-cli

# Run GUI in dev mode
run-gui:
	@wails dev

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

