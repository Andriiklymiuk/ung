package db

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestGetInvoicesDir(t *testing.T) {
	dir := GetInvoicesDir()
	if dir == "" {
		t.Error("expected invoices directory, got empty string")
	}

	// Should contain .ung in the path
	if !contains(dir, ".ung") {
		t.Errorf("expected path to contain '.ung', got: %s", dir)
	}
}

func TestGetDBPath(t *testing.T) {
	path := GetDBPath()
	if path == "" {
		t.Error("expected database path, got empty string")
	}

	// Should end with .db
	if filepath.Ext(path) != ".db" {
		t.Errorf("expected path to end with '.db', got: %s", path)
	}
}

func TestInlineSchema(t *testing.T) {
	// Create a temporary database
	tmpDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer tmpDB.Close()

	// Save and restore global DB
	oldDB := DB
	DB = tmpDB
	defer func() { DB = oldDB }()

	// Run inline schema
	if err := runInlineSchema(); err != nil {
		t.Fatalf("failed to run inline schema: %v", err)
	}

	// Verify core tables exist (inline schema may not have all tables from migrations)
	tables := []string{
		"companies",
		"clients",
		"invoices",
		"invoice_recipients",
		"tracking_sessions",
	}

	for _, table := range tables {
		var name string
		err := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table '%s' not found: %v", table, err)
		}
	}
}

func contains(s, substr string) bool {
	return filepath.Base(filepath.Dir(s)) == substr || filepath.Base(s) == substr
}
