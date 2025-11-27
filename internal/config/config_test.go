package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	// Test local default config
	cfg := getDefaultConfig(false)

	if cfg == nil {
		t.Fatal("expected config, got nil")
	}

	if cfg.Language != "en" {
		t.Errorf("expected default language 'en', got '%s'", cfg.Language)
	}

	if cfg.Invoice.InvoiceLabel != "INVOICE" {
		t.Errorf("expected invoice label 'INVOICE', got '%s'", cfg.Invoice.InvoiceLabel)
	}

	if cfg.Invoice.FromLabel != "From" {
		t.Errorf("expected from label 'From', got '%s'", cfg.Invoice.FromLabel)
	}

	if cfg.Invoice.BillToLabel != "Bill To" {
		t.Errorf("expected bill to label 'Bill To', got '%s'", cfg.Invoice.BillToLabel)
	}

	// Check local paths
	if cfg.DatabasePath != ".ung/ung.db" {
		t.Errorf("expected local database path '.ung/ung.db', got '%s'", cfg.DatabasePath)
	}

	// Test global default config
	cfgGlobal := getDefaultConfig(true)

	if cfgGlobal == nil {
		t.Fatal("expected global config, got nil")
	}

	// Global paths should contain home directory
	home, _ := os.UserHomeDir()
	expectedGlobalDB := filepath.Join(home, ".ung", "ung.db")
	if cfgGlobal.DatabasePath != expectedGlobalDB {
		t.Errorf("expected global database path '%s', got '%s'", expectedGlobalDB, cfgGlobal.DatabasePath)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantHome      bool
		wantAbsolute  bool
		wantUnchanged bool
	}{
		{
			name:     "home directory expansion",
			input:    "~/test",
			wantHome: true,
		},
		{
			name:          "absolute path no expansion",
			input:         "/tmp/test",
			wantUnchanged: true,
		},
		{
			name:         "relative path to absolute",
			input:        "test/path",
			wantAbsolute: true,
		},
		{
			name:         "dot relative path",
			input:        "./test.db",
			wantAbsolute: true,
		},
		{
			name:          "empty string",
			input:         "",
			wantUnchanged: true,
		},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)

			if tt.wantHome {
				expected := filepath.Join(home, tt.input[1:])
				if result != expected {
					t.Errorf("expected %s, got %s", expected, result)
				}
			} else if tt.wantUnchanged {
				if result != tt.input {
					t.Errorf("expected %s, got %s", tt.input, result)
				}
			} else if tt.wantAbsolute {
				if !filepath.IsAbs(result) {
					t.Errorf("expected absolute path, got %s", result)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create new .ung/config.yaml structure
	ungDir := filepath.Join(tmpDir, ".ung")
	if err := os.MkdirAll(ungDir, 0755); err != nil {
		t.Fatalf("failed to create .ung directory: %v", err)
	}

	configPath := filepath.Join(ungDir, "config.yaml")
	configContent := `language: "de"
database_path: ".ung/test.db"
invoices_dir: ".ung/invoices"
contracts_dir: ".ung/contracts"
invoice:
  invoice_label: "RECHNUNG"
  from_label: "Von"
  bill_to_label: "Rechnung an"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Reset state
	currentConfig = nil
	forceGlobal = false

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Language != "de" {
		t.Errorf("expected language 'de', got '%s'", cfg.Language)
	}

	if cfg.Invoice.InvoiceLabel != "RECHNUNG" {
		t.Errorf("expected invoice label 'RECHNUNG', got '%s'", cfg.Invoice.InvoiceLabel)
	}

	// Test that it's recognized as local config
	if configSource != SourceLocal {
		t.Errorf("expected config source to be SourceLocal, got %d", configSource)
	}

	// Test ContractsDir is set
	if cfg.ContractsDir != ".ung/contracts" {
		t.Errorf("expected contracts dir '.ung/contracts', got '%s'", cfg.ContractsDir)
	}
}

func TestForceGlobal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create local config
	ungDir := filepath.Join(tmpDir, ".ung")
	if err := os.MkdirAll(ungDir, 0755); err != nil {
		t.Fatalf("failed to create .ung directory: %v", err)
	}

	localConfigPath := filepath.Join(ungDir, "config.yaml")
	localConfigContent := `language: "fr"
invoice:
  invoice_label: "FACTURE"
`
	if err := os.WriteFile(localConfigPath, []byte(localConfigContent), 0644); err != nil {
		t.Fatalf("failed to write local config: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Reset state and force global
	currentConfig = nil
	SetForceGlobal(true)
	defer SetForceGlobal(false)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Should not be French (local), should be default or global
	if cfg.Language == "fr" {
		t.Errorf("expected global config to be used, but local config was loaded")
	}

	// Config source should not be local
	if configSource == SourceLocal {
		t.Errorf("expected non-local config source, got %d", configSource)
	}
}

func TestLocalUngDirExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Should not exist initially
	if LocalUngDirExists() {
		t.Error("expected local .ung dir to not exist")
	}

	// Create .ung directory
	if err := os.MkdirAll(LocalUngDir, 0755); err != nil {
		t.Fatalf("failed to create .ung directory: %v", err)
	}

	// Should exist now
	if !LocalUngDirExists() {
		t.Error("expected local .ung dir to exist")
	}
}
