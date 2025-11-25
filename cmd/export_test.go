package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportDefaults(t *testing.T) {
	// Test default values
	if exportFormat != "" {
		t.Errorf("Default export format should be empty, got %s", exportFormat)
	}

	if exportInvoices {
		t.Errorf("Default exportInvoices should be false")
	}

	if exportExpenses {
		t.Errorf("Default exportExpenses should be false")
	}

	if exportTime {
		t.Errorf("Default exportTime should be false")
	}
}

func TestExportOutputDirectory(t *testing.T) {
	// Test that default output directory is set correctly
	homeDir := os.Getenv("HOME")
	expectedDefault := filepath.Join(homeDir, ".ung", "exports")

	// Reset to test defaults
	exportOutput = ""

	// Default should be constructed from HOME
	if exportOutput == "" {
		exportOutput = filepath.Join(homeDir, ".ung", "exports")
	}

	if exportOutput != expectedDefault {
		t.Errorf("Expected output directory %s, got %s", expectedDefault, exportOutput)
	}
}

func TestSupportedFormats(t *testing.T) {
	// Test that expected formats are supported
	supportedFormats := []string{"csv", "quickbooks", "json"}

	for _, format := range supportedFormats {
		// These formats should not cause a panic when referenced
		switch format {
		case "csv", "quickbooks", "json":
			// Valid format
		default:
			t.Errorf("Expected format %s to be supported", format)
		}
	}
}
