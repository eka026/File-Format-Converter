.PHONY: build clean test run govulncheck

# Build GUI (requires Wails)
# NFR-04 (Single Binary): Builds a single executable with embedded assets
# All Go dependencies are statically linked. No external runtimes required.
build: govulncheck
	@echo "Building GUI (single binary, statically linked)..."
	@wails build -platform windows/amd64 -ldflags "-s -w" -trimpath

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

