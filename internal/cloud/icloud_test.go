package cloud

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestICloudAvailability(t *testing.T) {
	t.Run("Check iCloud availability", func(t *testing.T) {
		available := ICloudAvailable()

		// On non-macOS systems, should always be false
		if runtime.GOOS != "darwin" {
			if available {
				t.Fatal("iCloud should not be available on non-macOS systems")
			}
		}

		// On macOS, depends on whether iCloud Drive is configured
		// We can't guarantee either true or false, just that it returns without error
		t.Logf("iCloud available: %v (platform: %s)", available, runtime.GOOS)
	})
}

func TestGetICloudDrivePath(t *testing.T) {
	t.Run("Get iCloud Drive path", func(t *testing.T) {
		path := GetICloudDrivePath()

		if runtime.GOOS != "darwin" {
			if path != "" {
				t.Fatal("iCloud path should be empty on non-macOS")
			}
			return
		}

		// On macOS, if iCloud is available, path should be set
		if ICloudAvailable() {
			if path == "" {
				t.Fatal("iCloud path should not be empty when available")
			}

			// Should contain expected path component
			if !contains(path, "Library/Mobile Documents") {
				t.Fatalf("iCloud path should contain 'Library/Mobile Documents', got: %s", path)
			}
		}
	})
}

func TestGetICloudAppPath(t *testing.T) {
	t.Run("Get iCloud app path", func(t *testing.T) {
		if !ICloudAvailable() {
			_, err := GetICloudAppPath()
			if err == nil {
				t.Fatal("Expected error when iCloud is not available")
			}
			return
		}

		// If iCloud is available, should succeed
		appPath, err := GetICloudAppPath()
		if err != nil {
			t.Fatalf("Failed to get iCloud app path: %v", err)
		}

		if appPath == "" {
			t.Fatal("iCloud app path should not be empty")
		}

		// Should contain app container name
		if !contains(appPath, "ung") {
			t.Fatalf("iCloud app path should contain 'ung', got: %s", appPath)
		}

		// Directory should exist after calling GetICloudAppPath
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			t.Fatal("iCloud app directory should be created")
		}
	})
}

func TestGetICloudDatabasePath(t *testing.T) {
	t.Run("Get iCloud database path", func(t *testing.T) {
		if !ICloudAvailable() {
			t.Skip("iCloud not available")
		}

		dbPath, err := GetICloudDatabasePath("test.db")
		if err != nil {
			t.Fatalf("Failed to get iCloud database path: %v", err)
		}

		if dbPath == "" {
			t.Fatal("Database path should not be empty")
		}

		// Should end with the filename
		if filepath.Base(dbPath) != "test.db" {
			t.Fatalf("Expected filename 'test.db', got: %s", filepath.Base(dbPath))
		}

		// Should be an absolute path
		if !filepath.IsAbs(dbPath) {
			t.Fatalf("Database path should be absolute, got: %s", dbPath)
		}
	})
}

func TestGetICloudInvoicesPath(t *testing.T) {
	t.Run("Get iCloud invoices path", func(t *testing.T) {
		if !ICloudAvailable() {
			t.Skip("iCloud not available")
		}

		invoicesPath, err := GetICloudInvoicesPath()
		if err != nil {
			t.Fatalf("Failed to get iCloud invoices path: %v", err)
		}

		if invoicesPath == "" {
			t.Fatal("Invoices path should not be empty")
		}

		// Should end with 'invoices'
		if filepath.Base(invoicesPath) != "invoices" {
			t.Fatalf("Expected 'invoices' directory, got: %s", filepath.Base(invoicesPath))
		}

		// Directory should be created
		if _, err := os.Stat(invoicesPath); os.IsNotExist(err) {
			t.Fatal("Invoices directory should be created")
		}
	})
}

func TestIsSyncedToiCloud(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Local file not in iCloud", func(t *testing.T) {
		localFile := filepath.Join(tmpDir, "local.db")
		if err := os.WriteFile(localFile, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create local file: %v", err)
		}

		if IsSyncedToiCloud(localFile) {
			t.Fatal("Local file should not be detected as iCloud-synced")
		}
	})

	t.Run("iCloud file detection", func(t *testing.T) {
		if !ICloudAvailable() {
			t.Skip("iCloud not available")
		}

		appPath, err := GetICloudAppPath()
		if err != nil {
			t.Skip("Cannot get iCloud app path")
		}

		iCloudFile := filepath.Join(appPath, "test.db")
		if err := os.WriteFile(iCloudFile, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create iCloud file: %v", err)
		}
		defer os.Remove(iCloudFile)

		if !IsSyncedToiCloud(iCloudFile) {
			t.Fatal("iCloud file should be detected as iCloud-synced")
		}
	})
}

func TestGetSyncStatus(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Local file sync status", func(t *testing.T) {
		localFile := filepath.Join(tmpDir, "local.db")
		if err := os.WriteFile(localFile, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create local file: %v", err)
		}

		status := GetSyncStatus(localFile)
		if status != "not-in-icloud" {
			t.Fatalf("Expected 'not-in-icloud', got: %s", status)
		}
	})

	t.Run("Nonexistent file sync status", func(t *testing.T) {
		nonexistent := filepath.Join(tmpDir, "nonexistent.db")
		status := GetSyncStatus(nonexistent)

		// Should be either "not-in-icloud" or "downloading"
		if status != "not-in-icloud" && status != "downloading" {
			t.Fatalf("Unexpected status for nonexistent file: %s", status)
		}
	})

	t.Run("iCloud file sync status", func(t *testing.T) {
		if !ICloudAvailable() {
			t.Skip("iCloud not available")
		}

		appPath, err := GetICloudAppPath()
		if err != nil {
			t.Skip("Cannot get iCloud app path")
		}

		iCloudFile := filepath.Join(appPath, "test-sync.db")
		if err := os.WriteFile(iCloudFile, []byte("test content"), 0600); err != nil {
			t.Fatalf("Failed to create iCloud file: %v", err)
		}
		defer os.Remove(iCloudFile)

		status := GetSyncStatus(iCloudFile)

		// Should be one of the valid states
		validStatuses := map[string]bool{
			"synced":         true,
			"uploading":      true,
			"downloading":    true,
			"not-in-icloud":  true,
			"error":          true,
		}

		if !validStatuses[status] {
			t.Fatalf("Unexpected sync status: %s", status)
		}

		t.Logf("iCloud file sync status: %s", status)
	})
}

func TestGetOptimalPath(t *testing.T) {
	t.Run("Prefer local storage", func(t *testing.T) {
		dbPath, invoicesPath, err := GetOptimalPath(false)
		if err != nil {
			t.Fatalf("Failed to get local path: %v", err)
		}

		if dbPath == "" || invoicesPath == "" {
			t.Fatal("Paths should not be empty")
		}

		// Should not be in iCloud
		if IsSyncedToiCloud(dbPath) {
			t.Fatal("Local path should not be in iCloud")
		}
	})

	t.Run("Prefer iCloud storage", func(t *testing.T) {
		if !ICloudAvailable() {
			// Should fall back to local
			dbPath, invoicesPath, err := GetOptimalPath(true)
			if err != nil {
				t.Fatalf("Failed to get fallback path: %v", err)
			}

			if dbPath == "" || invoicesPath == "" {
				t.Fatal("Fallback paths should not be empty")
			}

			return
		}

		// iCloud is available
		dbPath, invoicesPath, err := GetOptimalPath(true)
		if err != nil {
			t.Fatalf("Failed to get iCloud path: %v", err)
		}

		if dbPath == "" || invoicesPath == "" {
			t.Fatal("iCloud paths should not be empty")
		}

		// Should be in iCloud
		if !IsSyncedToiCloud(dbPath) {
			t.Fatal("Optimal path with iCloud preference should be in iCloud")
		}
	})
}

func TestIsIOSDevice(t *testing.T) {
	t.Run("iOS device detection", func(t *testing.T) {
		isIOS := IsIOSDevice()

		// Can only be true on darwin/arm64
		if isIOS && !(runtime.GOOS == "darwin" && runtime.GOARCH == "arm64") {
			t.Fatal("IsIOSDevice should only return true on darwin/arm64")
		}

		t.Logf("iOS device: %v (GOOS=%s, GOARCH=%s)", isIOS, runtime.GOOS, runtime.GOARCH)
	})
}

// Helper function
func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
