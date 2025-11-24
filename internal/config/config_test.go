package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := getDefaultConfig()

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
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ung.yaml")

	configContent := `language: "uk"
invoice:
  invoice_label: "РАХУНОК"
  from_label: "Від"
  bill_to_label: "Кому"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Reset currentConfig to force reload
	currentConfig = nil

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Language != "uk" {
		t.Errorf("expected language 'uk', got '%s'", cfg.Language)
	}

	if cfg.Invoice.InvoiceLabel != "РАХУНОК" {
		t.Errorf("expected invoice label 'РАХУНОК', got '%s'", cfg.Invoice.InvoiceLabel)
	}

	if cfg.Invoice.FromLabel != "Від" {
		t.Errorf("expected from label 'Від', got '%s'", cfg.Invoice.FromLabel)
	}
}
