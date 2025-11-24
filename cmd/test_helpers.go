package cmd

import (
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
	t.Setenv("HOME", t.TempDir())

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
