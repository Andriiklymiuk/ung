---
id: installation
title: Installation
---

# Installation

## Homebrew (Recommended for macOS/Linux)

```bash
brew tap Andriiklymiuk/tools
brew install ung
```

### Update

```bash
brew upgrade ung
```

Or use the built-in update command:
```bash
ung update
```

## Go Install

If you have Go 1.22+ installed:

```bash
go install github.com/Andriiklymiuk/ung@latest
```

## Download Binary

Download the latest release from [GitHub Releases](https://github.com/Andriiklymiuk/ung/releases/latest).

### macOS (Apple Silicon)
```bash
curl -L https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_Darwin_arm64.tar.gz | tar xz
sudo mv ung /usr/local/bin/
```

### macOS (Intel)
```bash
curl -L https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_Darwin_x86_64.tar.gz | tar xz
sudo mv ung /usr/local/bin/
```

### Linux (amd64)
```bash
curl -L https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_Linux_x86_64.tar.gz | tar xz
sudo mv ung /usr/local/bin/
```

### Linux (arm64)
```bash
curl -L https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_Linux_arm64.tar.gz | tar xz
sudo mv ung /usr/local/bin/
```

## Verify Installation

```bash
ung version
```

## VSCode Extension

UNG also has a VSCode extension for GUI access:

1. Open VSCode
2. Go to Extensions (Cmd+Shift+X)
3. Search for "UNG"
4. Install the extension

The extension requires the CLI to be installed first.
