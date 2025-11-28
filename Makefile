.PHONY: build install test clean run help release release-vs-code-extension increase_version

# Build variables
BINARY_NAME=ung
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR=dist
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")
LDFLAGS=-ldflags "-s -w \
	-X 'github.com/Andriiklymiuk/ung/cmd.Version=$(VERSION)' \
	-X 'github.com/Andriiklymiuk/ung/cmd.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/Andriiklymiuk/ung/cmd.BuildDate=$(BUILD_DATE)'"

# Build the binary
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✓ Build complete: ./$(BINARY_NAME)"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	@go install $(LDFLAGS)
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
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Cross-compilation complete"

# Increase version - bump version in both CLI and VSCode extension without committing
# Usage: make increase_version [v=1.2.3]  (if v is not specified, patch version is bumped)
increase_version:
	$(eval CURRENT_VERSION := $(shell grep 'Version   = ' cmd/version.go | sed 's/.*"\([^"]*\)".*/\1/'))
	$(eval MAJOR := $(shell echo $(CURRENT_VERSION) | cut -d. -f1))
	$(eval MINOR := $(shell echo $(CURRENT_VERSION) | cut -d. -f2))
	$(eval PATCH := $(shell echo $(CURRENT_VERSION) | cut -d. -f3))
	$(eval NEW_PATCH := $(shell echo $$(($(PATCH) + 1))))
	$(eval NEW_VERSION := $(or $(v),$(MAJOR).$(MINOR).$(NEW_PATCH)))
	@echo "Bumping version from $(CURRENT_VERSION) to $(NEW_VERSION)..."
	@sed -i.bak 's/Version   = "[^"]*"/Version   = "$(NEW_VERSION)"/' cmd/version.go && rm -f cmd/version.go.bak
	@sed -i.bak 's/"version": "[^"]*"/"version": "$(NEW_VERSION)"/' vscode-ung/package.json && rm -f vscode-ung/package.json.bak
	@echo "✓ Updated cmd/version.go to $(NEW_VERSION)"
	@echo "✓ Updated vscode-ung/package.json to $(NEW_VERSION)"

# Release - bump version, create tag, and push to trigger release
# Usage: make release [v=1.2.3]  (if v is not specified, patch version is bumped)
# Version is read from cmd/version.go (source of truth), then bumped
release:
	@git diff --quiet || (echo "Error: Working directory has uncommitted changes" && exit 1)
	@$(MAKE) increase_version v=$(v)
	$(eval NEW_VERSION := $(shell grep 'Version   = ' cmd/version.go | sed 's/.*"\([^"]*\)".*/\1/'))
	@git add cmd/version.go vscode-ung/package.json
	@git commit -m "chore: bump version to $(NEW_VERSION)"
	@git tag -a v$(NEW_VERSION) -m "Release v$(NEW_VERSION)"
	@git push origin main v$(NEW_VERSION)
	@echo "✓ Version bumped, committed, tagged v$(NEW_VERSION), and pushed. Release workflow triggered."

release-vs-code-extension:
	cd vscode-ung && npm run compile
	cd vscode-ung && vsce package patch
	cd vscode-ung && git add .
	cd vscode-ung && git commit -m "vscode: release new version"
	cd vscode-ung && git push
	cd vscode-ung && vsce publish

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
	@echo "  release    - Create and push a release tag (auto-bumps patch, or: make release v=1.2.3)"
	@echo "  release-vs-code-extension - Compile, package, commit, push and publish VS Code extension"
	@echo "  increase_version - Bump version in CLI and VSCode extension (auto-bumps patch, or: make increase_version v=1.2.3)"
	@echo "  help       - Show this help message"
