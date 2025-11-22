.PHONY: build install test clean run help

# Build variables
BINARY_NAME=ung
VERSION?=$(shell cat VERSION 2>/dev/null || echo "0.1.0")
BUILD_DIR=dist
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")

# Build the binary
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) .
	@echo "✓ Build complete: ./$(BINARY_NAME)"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	@go install -ldflags="-s -w -X main.version=$(VERSION)"
	@echo "✓ Installation complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"

# Run the application
run: build
	@./$(BINARY_NAME)

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✓ Dependencies tidied"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "golangci-lint not installed, skipping..."

# Get version
getActionVersion:
	@echo "VERSION=$(VERSION)" >> $$GITHUB_ENV

# Development build with race detector
dev:
	@echo "Building for development..."
	@go build -race -o $(BINARY_NAME) .
	@echo "✓ Development build complete"

# Cross-compile for multiple platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Cross-compilation complete"

# Help
help:
	@echo "UNG Makefile targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  test       - Run tests"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Build and run the application"
	@echo "  tidy       - Tidy Go modules"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  dev        - Build with race detector"
	@echo "  build-all  - Cross-compile for all platforms"
	@echo "  help       - Show this help message"
