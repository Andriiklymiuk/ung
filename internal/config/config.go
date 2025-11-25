package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DatabasePath string           `yaml:"database_path"`
	InvoicesDir  string           `yaml:"invoices_dir"`
	Language     string           `yaml:"language"`      // e.g., "en", "uk", "de"
	Invoice      InvoiceConfig    `yaml:"invoice"`
	PDF          PDFConfig        `yaml:"pdf"`
	Templates    TemplateConfig   `yaml:"templates"`
	Email        EmailConfig      `yaml:"email"`
	Security     SecurityConfig   `yaml:"security"`
}

// PDFConfig represents PDF generation configuration
type PDFConfig struct {
	// Color theme (RGB values)
	PrimaryColor   ColorRGB `yaml:"primary_color"`   // Header/accent color (default: orange #E87722)
	SecondaryColor ColorRGB `yaml:"secondary_color"` // Secondary accent color
	TextColor      ColorRGB `yaml:"text_color"`      // Main text color

	// Watermark settings
	ShowWatermark bool   `yaml:"show_watermark"` // Show status watermark (PAID, DRAFT, etc.)
	WatermarkText string `yaml:"watermark_text"` // Custom watermark text (overrides status)

	// Layout options
	ShowLogo       bool `yaml:"show_logo"`        // Show company logo
	ShowQRCode     bool `yaml:"show_qr_code"`     // Show payment QR code
	ShowPageNumber bool `yaml:"show_page_number"` // Show page numbers on multi-page documents
	ShowTaxBreakdown bool `yaml:"show_tax_breakdown"` // Show VAT/tax breakdown

	// Tax settings
	TaxRate     float64 `yaml:"tax_rate"`      // Tax/VAT rate (e.g., 0.20 for 20%)
	TaxLabel    string  `yaml:"tax_label"`     // e.g., "VAT", "GST", "Tax"
	TaxInclusive bool   `yaml:"tax_inclusive"` // Whether prices include tax

	// Additional labels
	SubtotalLabel   string `yaml:"subtotal_label"`
	DiscountLabel   string `yaml:"discount_label"`
	TaxAmountLabel  string `yaml:"tax_amount_label"`
	BalanceDueLabel string `yaml:"balance_due_label"`
	PaidLabel       string `yaml:"paid_label"`
	DraftLabel      string `yaml:"draft_label"`
	OverdueLabel    string `yaml:"overdue_label"`
}

// ColorRGB represents an RGB color
type ColorRGB struct {
	R int `yaml:"r"`
	G int `yaml:"g"`
	B int `yaml:"b"`
}

// TemplateConfig represents template paths for PDF generation
type TemplateConfig struct {
	InvoiceHTML  string `yaml:"invoice_html"`  // Path to invoice HTML template
	ContractHTML string `yaml:"contract_html"` // Path to contract HTML template
}

// InvoiceConfig represents invoice-specific configuration
type InvoiceConfig struct {
	Terms            string `yaml:"terms"`             // Terms & Conditions
	PaymentNote      string `yaml:"payment_note"`      // Payment instruction
	NotesLabel       string `yaml:"notes_label"`       // "Notes" label in chosen language
	TermsLabel       string `yaml:"terms_label"`       // "Terms & Conditions" label
	InvoiceLabel     string `yaml:"invoice_label"`     // "INVOICE" title
	FromLabel        string `yaml:"from_label"`        // "From" label
	BillToLabel      string `yaml:"bill_to_label"`     // "Bill To" label
	DescriptionLabel string `yaml:"description_label"` // "Description" label
	TotalLabel       string `yaml:"total_label"`       // "Total" label
	ItemLabel        string `yaml:"item_label"`        // "Item" column header
	QuantityLabel    string `yaml:"quantity_label"`    // "Quantity" column header
	RateLabel        string `yaml:"rate_label"`        // "Rate" column header
	AmountLabel      string `yaml:"amount_label"`      // "Amount" column header
}

// EmailConfig represents email/SMTP configuration
type EmailConfig struct {
	SMTPHost  string `yaml:"smtp_host"`  // SMTP server host
	SMTPPort  int    `yaml:"smtp_port"`  // SMTP server port
	Username  string `yaml:"username"`   // SMTP username
	Password  string `yaml:"password"`   // SMTP password (or app password)
	FromEmail string `yaml:"from_email"` // Sender email address
	FromName  string `yaml:"from_name"`  // Sender name
	UseTLS    bool   `yaml:"use_tls"`    // Use TLS encryption
}

// SecurityConfig represents database security configuration
type SecurityConfig struct {
	EncryptDatabase bool `yaml:"encrypt_database"` // Whether to encrypt database at rest
}

var currentConfig *Config

// Load loads configuration from local (.ung.yaml) or global (~/.ung/config.yaml)
func Load() (*Config, error) {
	if currentConfig != nil {
		return currentConfig, nil
	}

	// Try local config first
	localConfig := ".ung.yaml"
	if _, err := os.Stat(localConfig); err == nil {
		cfg, err := loadFromFile(localConfig)
		if err == nil {
			currentConfig = cfg
			return currentConfig, nil
		}
	}

	// Try global config
	home, err := os.UserHomeDir()
	if err != nil {
		return getDefaultConfig(), nil
	}

	globalConfig := filepath.Join(home, ".ung", "config.yaml")
	if _, err := os.Stat(globalConfig); err == nil {
		cfg, err := loadFromFile(globalConfig)
		if err == nil {
			currentConfig = cfg
			return currentConfig, nil
		}
	}

	// Return default config
	currentConfig = getDefaultConfig()
	return currentConfig, nil
}

// loadFromFile loads config from a YAML file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand paths
	if cfg.DatabasePath != "" {
		cfg.DatabasePath = expandPath(cfg.DatabasePath)
	}
	if cfg.InvoicesDir != "" {
		cfg.InvoicesDir = expandPath(cfg.InvoicesDir)
	}

	return &cfg, nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return &Config{
		DatabasePath: filepath.Join(home, ".ung", "ung.db"),
		InvoicesDir:  filepath.Join(home, ".ung", "invoices"),
		Language:     "en",
		Invoice: InvoiceConfig{
			Terms:            "Please make the payment by the due date.",
			PaymentNote:      "Payment is due within the specified term.",
			NotesLabel:       "Notes",
			TermsLabel:       "Terms & Conditions",
			InvoiceLabel:     "INVOICE",
			FromLabel:        "From",
			BillToLabel:      "Bill To",
			DescriptionLabel: "Description",
			TotalLabel:       "Total",
			ItemLabel:        "Item",
			QuantityLabel:    "Quantity",
			RateLabel:        "Rate",
			AmountLabel:      "Amount",
		},
		PDF: PDFConfig{
			PrimaryColor:   ColorRGB{R: 232, G: 119, B: 34},  // Orange #E87722
			SecondaryColor: ColorRGB{R: 80, G: 80, B: 80},    // Gray
			TextColor:      ColorRGB{R: 60, G: 60, B: 60},    // Dark gray
			ShowWatermark:  true,
			ShowLogo:       true,
			ShowQRCode:     false, // Disabled by default until library is added
			ShowPageNumber: true,
			ShowTaxBreakdown: false,
			TaxRate:        0.0,
			TaxLabel:       "VAT",
			TaxInclusive:   false,
			SubtotalLabel:   "Subtotal",
			DiscountLabel:   "Discount",
			TaxAmountLabel:  "VAT",
			BalanceDueLabel: "Balance Due",
			PaidLabel:       "PAID",
			DraftLabel:      "DRAFT",
			OverdueLabel:    "OVERDUE",
		},
	}
}

// GetDefaultPDFConfig returns the default PDF configuration
func GetDefaultPDFConfig() PDFConfig {
	return PDFConfig{
		PrimaryColor:   ColorRGB{R: 232, G: 119, B: 34},
		SecondaryColor: ColorRGB{R: 80, G: 80, B: 80},
		TextColor:      ColorRGB{R: 60, G: 60, B: 60},
		ShowWatermark:  true,
		ShowLogo:       true,
		ShowQRCode:     false,
		ShowPageNumber: true,
		ShowTaxBreakdown: false,
		TaxRate:        0.0,
		TaxLabel:       "VAT",
		TaxInclusive:   false,
		SubtotalLabel:   "Subtotal",
		DiscountLabel:   "Discount",
		TaxAmountLabel:  "VAT",
		BalanceDueLabel: "Balance Due",
		PaidLabel:       "PAID",
		DraftLabel:      "DRAFT",
		OverdueLabel:    "OVERDUE",
	}
}

// expandPath expands ~ to home directory and converts relative paths to absolute
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}

	// Expand ~ to home directory
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert relative paths to absolute
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		return absPath
	}

	return path
}

// Save saves the current configuration to a file
func Save(cfg *Config, global bool) error {
	var path string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".ung", "config.yaml")
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
	} else {
		path = ".ung.yaml"
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDatabasePath returns the configured database path
func GetDatabasePath() string {
	cfg, _ := Load()
	return cfg.DatabasePath
}

// GetInvoicesDir returns the configured invoices directory
func GetInvoicesDir() string {
	cfg, _ := Load()
	return cfg.InvoicesDir
}
