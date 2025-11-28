.PHONY: build \
	install \
	test \
	clean \
	run \
	help \
	fmt \
	fmtSwift \
	fmtAll \
	release \
	releaseVsCodeExtension \
	increaseVersion \
	vscodePackage \
	macosBuild \
	macosArchive \
	macosExport \
	macosUpload \
	macosRelease

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

# macOS app variables
MACOS_PROJECT=macos-ung/ung.xcodeproj
MACOS_SCHEME=ung
MACOS_BUILD_DIR=$(BUILD_DIR)/macos
MACOS_ARCHIVE=$(MACOS_BUILD_DIR)/UNG.xcarchive
MACOS_EXPORT_DIR=$(MACOS_BUILD_DIR)/export
MACOS_APP_BUNDLE_ID=com.ung.ung

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
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "✓ Go code formatted"

# Format Swift code in macOS app
fmtSwift:
	@echo "Formatting Swift code..."
	@swift-format format --in-place --recursive macos-ung/ung
	@echo "✓ Swift code formatted"

# Format all code (Go + Swift)
fmtAll: fmt fmtSwift
	@echo "✓ All code formatted"

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

# Increase version - bump version in CLI, VSCode extension, and macOS app without committing
# Usage: make increaseVersion [v=1.2.3]  (if v is not specified, patch version is bumped)
increaseVersion:
	$(eval CURRENT_VERSION := $(shell grep 'Version   = ' cmd/version.go | sed 's/.*"\([^"]*\)".*/\1/'))
	$(eval MAJOR := $(shell echo $(CURRENT_VERSION) | cut -d. -f1))
	$(eval MINOR := $(shell echo $(CURRENT_VERSION) | cut -d. -f2))
	$(eval PATCH := $(shell echo $(CURRENT_VERSION) | cut -d. -f3))
	$(eval NEW_PATCH := $(shell echo $$(($(PATCH) + 1))))
	$(eval NEW_VERSION := $(or $(v),$(MAJOR).$(MINOR).$(NEW_PATCH)))
	$(eval CURRENT_BUILD := $(shell grep -m 1 'CURRENT_PROJECT_VERSION = ' $(MACOS_PROJECT)/project.pbxproj | sed 's/.*= \([0-9]*\);/\1/'))
	$(eval NEW_BUILD := $(shell echo $$(($(CURRENT_BUILD) + 1))))
	@echo "Bumping version from $(CURRENT_VERSION) to $(NEW_VERSION)..."
	@sed -i.bak 's/Version   = "[^"]*"/Version   = "$(NEW_VERSION)"/' cmd/version.go && rm -f cmd/version.go.bak
	@sed -i.bak 's/"version": "[^"]*"/"version": "$(NEW_VERSION)"/' vscode-ung/package.json && rm -f vscode-ung/package.json.bak
	@sed -i.bak 's/MARKETING_VERSION = [^;]*/MARKETING_VERSION = $(NEW_VERSION)/g' $(MACOS_PROJECT)/project.pbxproj && rm -f $(MACOS_PROJECT)/project.pbxproj.bak
	@sed -i.bak 's/CURRENT_PROJECT_VERSION = [^;]*/CURRENT_PROJECT_VERSION = $(NEW_BUILD)/g' $(MACOS_PROJECT)/project.pbxproj && rm -f $(MACOS_PROJECT)/project.pbxproj.bak
	@echo "✓ Updated cmd/version.go to $(NEW_VERSION)"
	@echo "✓ Updated vscode-ung/package.json to $(NEW_VERSION)"
	@echo "✓ Updated macOS app version to $(NEW_VERSION) (build $(NEW_BUILD))"

# Release - bump version, create tag, and push to trigger release
# Usage: make release [v=1.2.3]  (if v is not specified, patch version is bumped)
# Version is read from cmd/version.go (source of truth), then bumped
release:
	@git diff --quiet || (echo "Error: Working directory has uncommitted changes" && exit 1)
	@$(MAKE) increaseVersion v=$(v)
	$(eval NEW_VERSION := $(shell grep 'Version   = ' cmd/version.go | sed 's/.*"\([^"]*\)".*/\1/'))
	@git add cmd/version.go vscode-ung/package.json $(MACOS_PROJECT)/project.pbxproj
	@git commit -m "chore: bump version to $(NEW_VERSION)"
	@git tag -a v$(NEW_VERSION) -m "Release v$(NEW_VERSION)"
	@git push origin main v$(NEW_VERSION)
	@echo "✓ Version bumped, committed, tagged v$(NEW_VERSION), and pushed. Release workflow triggered."

releaseVsCodeExtension:
	cd vscode-ung && npm run compile
	cd vscode-ung && vsce package patch
	cd vscode-ung && git add .
	cd vscode-ung && git commit -m "vscode: release new version"
	cd vscode-ung && git push
	cd vscode-ung && vsce publish

# Package VSCode extension locally (compile and create .vsix)
vscodePackage:
	cd vscode-ung && bun compile && vsce package

# macOS App Build and Release Commands

# Build macOS app in Release configuration
macosBuild:
	@echo "Building macOS app in Release configuration..."
	@xcodebuild -project $(MACOS_PROJECT) \
		-scheme $(MACOS_SCHEME) \
		-configuration Release \
		build
	@echo "✓ macOS app build complete"

# Archive macOS app for distribution
macosArchive:
	@echo "Archiving macOS app..."
	@mkdir -p $(MACOS_BUILD_DIR)
	@xcodebuild -project $(MACOS_PROJECT) \
		-scheme $(MACOS_SCHEME) \
		-configuration Release \
		-archivePath $(MACOS_ARCHIVE) \
		archive
	@echo "✓ Archive created at $(MACOS_ARCHIVE)"

# Export archive for App Store distribution
macosExport:
	@echo "Exporting macOS app for App Store..."
	@mkdir -p $(MACOS_EXPORT_DIR)
	@xcodebuild -exportArchive \
		-archivePath $(MACOS_ARCHIVE) \
		-exportOptionsPlist macos-ung/ExportOptions.plist \
		-exportPath $(MACOS_EXPORT_DIR)
	@echo "✓ App exported to $(MACOS_EXPORT_DIR)"

# Upload to App Store Connect (requires valid credentials)
# Set APP_STORE_CONNECT_API_KEY_PATH, API_KEY_ID, and API_ISSUER_ID environment variables
macosUpload:
	@echo "Uploading macOS app to App Store Connect..."
	@xcrun altool --upload-app \
		-f $(MACOS_EXPORT_DIR)/UNG.pkg \
		-t macos \
		--apiKey $(API_KEY_ID) \
		--apiIssuer $(API_ISSUER_ID)
	@echo "✓ App uploaded to App Store Connect"

# Complete release process: build, archive, export, and upload
macosRelease: macosArchive macosExport macosUpload
	@echo "✓ macOS app release process complete!"

# Help
help:
	@echo "UNG Makefile targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  test       - Run tests"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Build and run the application"
	@echo "  tidy       - Tidy Go modules"
	@echo "  fmt        - Format Go code"
	@echo "  fmtSwift   - Format Swift code in macOS app"
	@echo "  fmtAll     - Format all code (Go + Swift)"
	@echo "  lint       - Run linter"
	@echo "  dev        - Build with race detector"
	@echo "  build-all  - Cross-compile for all platforms"
	@echo "  release    - Create and push a release tag (auto-bumps patch, or: make release v=1.2.3)"
	@echo "  releaseVsCodeExtension - Compile, package, commit, push and publish VS Code extension"
	@echo "  increaseVersion - Bump version in CLI, VSCode extension, and macOS app (auto-bumps patch, or: make increaseVersion v=1.2.3)"
	@echo "  vscodePackage - Compile and package VSCode extension locally"
	@echo ""
	@echo "macOS App targets:"
	@echo "  macosBuild   - Build macOS app in Release configuration"
	@echo "  macosArchive - Create .xcarchive for distribution"
	@echo "  macosExport  - Export archive for App Store (requires ExportOptions.plist)"
	@echo "  macosUpload  - Upload to App Store Connect (requires API key)"
	@echo "  macosRelease - Complete release: archive, export, and upload"
	@echo "  help         - Show this help message"
