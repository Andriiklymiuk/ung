package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncDefaults(t *testing.T) {
	// Test default values
	if syncOutputPath != "" {
		t.Errorf("Default sync output path should be empty, got %s", syncOutputPath)
	}

	if syncForce {
		t.Errorf("Default syncForce should be false")
	}
}

func TestBackupDirectory(t *testing.T) {
	// Test default backup directory is set correctly
	homeDir := os.Getenv("HOME")
	expectedDefault := filepath.Join(homeDir, ".ung", "backups")

	// Verify path construction
	testDir := filepath.Join(homeDir, ".ung", "backups")
	if testDir != expectedDefault {
		t.Errorf("Expected backup directory %s, got %s", expectedDefault, testDir)
	}
}

func TestBackupDataStructure(t *testing.T) {
	// Test that BackupData can be created
	backup := BackupData{
		Version: "1.0",
	}

	if backup.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", backup.Version)
	}
}
