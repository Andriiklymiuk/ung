package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Andriiklymiuk/ung/internal/db"
)

// setupTestDB initializes a test database for unit tests
func setupTestDB(t *testing.T) {
	// Close any existing database connection
	if db.DB != nil {
		db.Close()
	}

	// Set test database path
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Create .ung directory with a proper config file to mark it as initialized
	ungDir := filepath.Join(tempHome, ".ung")
	if err := os.MkdirAll(ungDir, 0755); err != nil {
		t.Fatalf("Failed to create .ung directory: %v", err)
	}

	// Create a config file with required paths
	configContent := `database_path: ` + filepath.Join(ungDir, "ung.db") + `
invoices_dir: ` + filepath.Join(ungDir, "invoices") + `
`
	configPath := filepath.Join(ungDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Initialize test database
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Clean up any existing test data
	db.DB.Exec("DELETE FROM tracking_sessions")
	db.DB.Exec("DELETE FROM contracts")
	db.DB.Exec("DELETE FROM clients")
}
