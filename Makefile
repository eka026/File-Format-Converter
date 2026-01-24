.PHONY: build clean test run govulncheck

# Build GUI (requires Wails)
build: govulncheck
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
	@wails dev -assetdir ./web -reloaddirs ./web

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run vulnerability scanning (NFR-03)
govulncheck:
	@echo "Running vulnerability scan..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@$(shell go env GOPATH)/bin/govulncheck.exe ./... 2>NUL || $(shell go env GOPATH)/bin/govulncheck ./...
	@echo "Vulnerability scan passed."

