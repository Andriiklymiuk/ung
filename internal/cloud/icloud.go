package cloud

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// iCloud Drive paths for different platforms
	iCloudDriveMacOS = "Library/Mobile Documents/com~apple~CloudDocs"
	iCloudAppContainer = "ung" // App-specific container in iCloud Drive
)

// ICloudAvailable checks if iCloud Drive is available on this system
func ICloudAvailable() bool {
	if runtime.GOOS != "darwin" {
		return false // iCloud only available on macOS/iOS
	}

	path := GetICloudDrivePath()
	if path == "" {
		return false
	}

	// Check if path exists and is accessible
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// GetICloudDrivePath returns the path to iCloud Drive root
// Returns empty string if not available
func GetICloudDrivePath() string {
	if runtime.GOOS != "darwin" {
		return ""
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	iCloudPath := filepath.Join(home, iCloudDriveMacOS)

	// Check if path exists
	if _, err := os.Stat(iCloudPath); err != nil {
		return ""
	}

	return iCloudPath
}

// GetICloudAppPath returns the path to the app-specific container in iCloud Drive
// Creates the directory if it doesn't exist
func GetICloudAppPath() (string, error) {
	iCloudRoot := GetICloudDrivePath()
	if iCloudRoot == "" {
		return "", fmt.Errorf("iCloud Drive not available")
	}

	appPath := filepath.Join(iCloudRoot, iCloudAppContainer)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create iCloud app directory: %w", err)
	}

	return appPath, nil
}

// GetICloudDatabasePath returns the full path for a database file in iCloud
func GetICloudDatabasePath(filename string) (string, error) {
	appPath, err := GetICloudAppPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(appPath, filename), nil
}

// GetICloudInvoicesPath returns the path for invoices in iCloud
func GetICloudInvoicesPath() (string, error) {
	appPath, err := GetICloudAppPath()
	if err != nil {
		return "", err
	}

	invoicesPath := filepath.Join(appPath, "invoices")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(invoicesPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create invoices directory: %w", err)
	}

	return invoicesPath, nil
}

// IsSyncedToiCloud checks if a file is in an iCloud-synced location
func IsSyncedToiCloud(path string) bool {
	iCloudPath := GetICloudDrivePath()
	if iCloudPath == "" {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check if path is within iCloud Drive
	rel, err := filepath.Rel(iCloudPath, absPath)
	if err != nil {
		return false
	}

	// If relative path doesn't start with "..", it's inside iCloud
	return len(rel) > 0 && rel[0] != '.' && rel[:2] != ".."
}

// GetSyncStatus returns the iCloud sync status of a file
// Returns one of: "synced", "downloading", "uploading", "not-in-icloud", "error"
func GetSyncStatus(path string) string {
	if !IsSyncedToiCloud(path) {
		return "not-in-icloud"
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "downloading" // File might be evicted and downloading
		}
		return "error"
	}

	// Check for .icloud placeholder files (macOS evicts files not recently used)
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	placeholder := filepath.Join(dir, "."+base+".icloud")

	if _, err := os.Stat(placeholder); err == nil {
		return "downloading" // File is evicted, placeholder exists
	}

	// If file exists and no placeholder, assume synced
	// Note: There's no reliable way to check upload/download progress without
	// private APIs, so we assume synced if file is present
	if info.Size() > 0 {
		return "synced"
	}

	return "uploading"
}

// IsIOSDevice checks if running on iOS (for future iOS app)
func IsIOSDevice() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

// GetOptimalPath returns the optimal database path based on platform and availability
// Prefers iCloud if available on macOS, local otherwise
func GetOptimalPath(preferICloud bool) (dbPath string, invoicesPath string, err error) {
	if preferICloud && ICloudAvailable() {
		dbPath, err = GetICloudDatabasePath("ung.db")
		if err != nil {
			return "", "", err
		}

		invoicesPath, err = GetICloudInvoicesPath()
		if err != nil {
			return "", "", err
		}

		return dbPath, invoicesPath, nil
	}

	// Fallback to local path
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	dbPath = filepath.Join(home, ".ung", "ung.db")
	invoicesPath = filepath.Join(home, ".ung", "invoices")

	return dbPath, invoicesPath, nil
}
