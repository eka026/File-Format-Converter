.PHONY: build clean test run

# Build GUI (requires Wails)
build:
	@echo "Building GUI..."
	@wails build

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

# Run GUI in dev mode
run:
	@wails dev

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

