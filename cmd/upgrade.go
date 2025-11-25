package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade ung to the latest version",
	Long: `Check for the latest release on GitHub and upgrade ung to the newest version.

This command will:
1. Check GitHub for the latest release
2. Compare with your current version
3. Download and install the new version if available`,
	Run: runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Checking for updates...")

	// Detect installation method
	installMethod := detectInstallMethod()

	// Get latest release info from GitHub
	release, err := getLatestRelease()
	if err != nil {
		fmt.Printf("❌ Failed to check for updates: %v\n", err)
		fmt.Println("\n💡 You can manually check for updates at:")
		fmt.Println("   https://github.com/Andriiklymiuk/ung/releases")
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	fmt.Printf("\nCurrent version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", latestVersion)

	if installMethod != "" {
		fmt.Printf("Installation:    %s\n", installMethod)
	}

	// Compare versions
	if currentVersion == latestVersion {
		fmt.Println("\n✅ You are already running the latest version!")
		return
	}

	if latestVersion == "" {
		fmt.Println("\n⚠️  Could not determine latest version")
		return
	}

	fmt.Printf("\n🎉 New version available: %s\n", release.Name)
	fmt.Printf("📝 Release page: %s\n", release.HTMLURL)

	// Auto-upgrade based on installation method
	switch installMethod {
	case "homebrew":
		fmt.Println("\n📦 Upgrading via Homebrew...")
		fmt.Println("   Running: brew update && brew upgrade ung")
		cmd := exec.Command("sh", "-c", "brew update && brew upgrade ung")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\n❌ Homebrew upgrade failed: %v\n", err)
			fmt.Println("Try running manually: brew update && brew upgrade ung")
			return
		}
		fmt.Println("\n✅ Successfully upgraded via Homebrew!")
		return
	case "go install":
		fmt.Println("\n📦 Upgrading via go install...")
		fmt.Println("   Running: go install github.com/Andriiklymiuk/ung@latest")
		cmd := exec.Command("go", "install", "github.com/Andriiklymiuk/ung@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\n❌ go install failed: %v\n", err)
			fmt.Println("Try running manually: go install github.com/Andriiklymiuk/ung@latest")
			return
		}
		fmt.Println("\n✅ Successfully upgraded via go install!")
		return
	}

	// For direct binary installation, proceed with download
	fmt.Println("\n📦 Direct binary installation detected.")

	// Find the appropriate asset for current OS/Arch
	assetURL := findAssetForPlatform(release, runtime.GOOS, runtime.GOARCH)
	if assetURL == "" {
		fmt.Println("\n⚠️  No pre-built binary found for your platform")
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("\n💡 You can build from source:")
		fmt.Println("   git clone https://github.com/Andriiklymiuk/ung")
		fmt.Println("   cd ung")
		fmt.Println("   make build")
		return
	}

	// Ask for confirmation
	fmt.Print("\n❓ Would you like to upgrade? (y/N): ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		fmt.Println("Upgrade cancelled")
		return
	}

	// Download and install
	fmt.Println("\n⬇️  Downloading new version...")
	if err := downloadAndInstall(assetURL); err != nil {
		fmt.Printf("❌ Failed to upgrade: %v\n", err)
		return
	}

	fmt.Println("\n✅ Successfully upgraded to version", latestVersion)
	fmt.Println("🔄 Please restart ung to use the new version")
}

func detectInstallMethod() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return ""
	}

	// Check for Homebrew installation
	if strings.Contains(exePath, "/homebrew/") || strings.Contains(exePath, "/Homebrew/") ||
	   strings.Contains(exePath, "/Cellar/") || strings.Contains(exePath, "/opt/homebrew/") {
		return "homebrew"
	}

	// Check for go install (usually in $GOPATH/bin or $HOME/go/bin)
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		home, _ := os.UserHomeDir()
		goPath = filepath.Join(home, "go")
	}
	if strings.HasPrefix(exePath, filepath.Join(goPath, "bin")) {
		return "go install"
	}

	return "binary"
}

func getLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/Andriiklymiuk/ung/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func findAssetForPlatform(release *GitHubRelease, goos, goarch string) string {
	// Common naming patterns for binaries
	patterns := []string{
		fmt.Sprintf("ung_%s_%s", goos, goarch),
		fmt.Sprintf("ung-%s-%s", goos, goarch),
		fmt.Sprintf("ung_%s-%s", goos, goarch),
	}

	// Special case for Windows
	if goos == "windows" {
		for i, p := range patterns {
			patterns = append(patterns, p+".exe")
			patterns[i] = p
		}
	}

	for _, asset := range release.Assets {
		assetName := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			if strings.Contains(assetName, strings.ToLower(pattern)) {
				return asset.BrowserDownloadURL
			}
		}
	}

	return ""
}

func downloadAndInstall(url string) error {
	// Download the new binary
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "ung-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write downloaded content to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Make it executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Backup current binary
	backupPath := exePath + ".backup"
	if err := copyFile(exePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Replace the current binary
	if err := os.Rename(tmpPath, exePath); err != nil {
		// Try to restore backup on failure
		os.Rename(backupPath, exePath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

func isGoInstalled() bool {
	_, err := exec.LookPath("go")
	return err == nil
}
