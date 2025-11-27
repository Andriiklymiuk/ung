package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DatabasePath string           `yaml:"database_path"`
	InvoicesDir  string           `yaml:"invoices_dir"`
	ContractsDir string           `yaml:"contracts_dir,omitempty"` // Path to contracts directory
	Language     string           `yaml:"language"`                // e.g., "en", "uk", "de"
	Invoice      InvoiceConfig    `yaml:"invoice"`
	PDF          PDFConfig        `yaml:"pdf"`
	Templates    TemplateConfig   `yaml:"templates"`
	Email        EmailConfig      `yaml:"email"`
	Security     SecurityConfig   `yaml:"security"`
}

// ConfigSource indicates where the config was loaded from
type ConfigSource int

const (
	SourceDefault ConfigSource = iota
	SourceLocal                // .ung/config.yaml in current directory
	SourceGlobal               // ~/.ung/config.yaml
)

// LocalUngDir is the name of the local ung directory
const LocalUngDir = ".ung"

// configSource tracks where the current config was loaded from
var configSource ConfigSource

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

// forceGlobal forces using global config when true
var forceGlobal bool

// SetForceGlobal sets whether to force using global configuration
func SetForceGlobal(force bool) {
	forceGlobal = force
	// Reset cached config so it will be reloaded
	currentConfig = nil
}

// GetLocalUngDir returns the path to the local .ung directory
func GetLocalUngDir() string {
	return LocalUngDir
}

// GetGlobalUngDir returns the path to the global .ung directory
func GetGlobalUngDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ung")
}

// GetConfigSource returns where the current config was loaded from
func GetConfigSource() ConfigSource {
	return configSource
}

// GetConfigSourceString returns a human-readable string for the config source
func GetConfigSourceString() string {
	switch configSource {
	case SourceLocal:
		return "local (.ung/config.yaml)"
	case SourceGlobal:
		return "global (~/.ung/config.yaml)"
	default:
		return "default"
	}
}

// Reload clears the cached config and reloads it
func Reload() (*Config, error) {
	currentConfig = nil
	return Load()
}

// Load loads configuration with priority:
// 1. Local .ung/config.yaml (if not --global and local .ung/ exists)
// 2. Global ~/.ung/config.yaml
// 3. Default config (uses global ~/.ung/ paths unless local .ung/ already exists)
func Load() (*Config, error) {
	if currentConfig != nil {
		return currentConfig, nil
	}

	// If not forcing global, try local config first (only if .ung directory exists)
	if !forceGlobal {
		// Check if local .ung directory exists
		if _, err := os.Stat(LocalUngDir); err == nil {
			localConfigPath := filepath.Join(LocalUngDir, "config.yaml")
			if _, err := os.Stat(localConfigPath); err == nil {
				cfg, err := loadFromFile(localConfigPath)
				if err == nil {
					configSource = SourceLocal
					currentConfig = cfg
					return currentConfig, nil
				}
			}
			// Local .ung/ exists but no config - use local defaults
			currentConfig = getDefaultConfig(false)
			configSource = SourceLocal
			return currentConfig, nil
		}
	}

	// Try global config
	home, err := os.UserHomeDir()
	if err != nil {
		currentConfig = getDefaultConfig(true) // Use global paths
		configSource = SourceDefault
		return currentConfig, nil
	}

	globalConfig := filepath.Join(home, ".ung", "config.yaml")
	if _, err := os.Stat(globalConfig); err == nil {
		cfg, err := loadFromFile(globalConfig)
		if err == nil {
			configSource = SourceGlobal
			currentConfig = cfg
			return currentConfig, nil
		}
	}

	// Return default config - use global paths (don't create local .ung/ automatically)
	currentConfig = getDefaultConfig(true)
	configSource = SourceGlobal
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
// If useGlobal is true, uses global ~/.ung/ paths, otherwise uses local .ung/ paths
func getDefaultConfig(useGlobal bool) *Config {
	var basePath string

	if useGlobal {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		basePath = filepath.Join(home, ".ung")
	} else {
		// Use local .ung directory
		basePath = LocalUngDir
	}

	return &Config{
		DatabasePath: filepath.Join(basePath, "ung.db"),
		InvoicesDir:  filepath.Join(basePath, "invoices"),
		ContractsDir: filepath.Join(basePath, "contracts"),
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
			PrimaryColor:     ColorRGB{R: 232, G: 119, B: 34}, // Orange #E87722
			SecondaryColor:   ColorRGB{R: 80, G: 80, B: 80},   // Gray
			TextColor:        ColorRGB{R: 60, G: 60, B: 60},   // Dark gray
			ShowWatermark:    true,
			ShowLogo:         true,
			ShowQRCode:       false, // Disabled by default until library is added
			ShowPageNumber:   true,
			ShowTaxBreakdown: false,
			TaxRate:          0.0,
			TaxLabel:         "VAT",
			TaxInclusive:     false,
			SubtotalLabel:    "Subtotal",
			DiscountLabel:    "Discount",
			TaxAmountLabel:   "VAT",
			BalanceDueLabel:  "Balance Due",
			PaidLabel:        "PAID",
			DraftLabel:       "DRAFT",
			OverdueLabel:     "OVERDUE",
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
// If global is true, saves to ~/.ung/config.yaml
// If global is false, saves to .ung/config.yaml (local directory)
func Save(cfg *Config, global bool) error {
	var path string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".ung", "config.yaml")
	} else {
		// Use new local structure: .ung/config.yaml
		path = filepath.Join(LocalUngDir, "config.yaml")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
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

// InitLocalDirectory creates the local .ung directory structure with default config
func InitLocalDirectory() error {
	// Create main .ung directory
	if err := os.MkdirAll(LocalUngDir, 0755); err != nil {
		return fmt.Errorf("failed to create .ung directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(LocalUngDir, "invoices"),
		filepath.Join(LocalUngDir, "contracts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Add .ung to .gitignore if not already present
	addToGitignore(".ung")

	return nil
}

// addToGitignore adds an entry to .gitignore if it doesn't already exist
func addToGitignore(entry string) {
	gitignorePath := ".gitignore"

	// Read existing .gitignore
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return // Can't read file, skip
	}

	// Check if entry already exists
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == entry || strings.TrimSpace(line) == entry+"/" {
			return // Already present
		}
	}

	// Append entry to .gitignore
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Can't open file, skip
	}
	defer f.Close()

	// Add newline before entry if file doesn't end with one
	prefix := ""
	if len(content) > 0 && content[len(content)-1] != '\n' {
		prefix = "\n"
	}

	f.WriteString(prefix + entry + "/\n")
}

// InitGlobalDirectory creates the global ~/.ung directory structure with default config
func InitGlobalDirectory() error {
	globalDir := GetGlobalUngDir()
	if globalDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Create main .ung directory
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		return fmt.Errorf("failed to create global .ung directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(globalDir, "invoices"),
		filepath.Join(globalDir, "contracts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// MigrateLocalToGlobal copies the local config to global
func MigrateLocalToGlobal() error {
	// Load current local config
	localConfigPath := filepath.Join(LocalUngDir, "config.yaml")
	cfg, err := loadFromFile(localConfigPath)
	if err != nil {
		return fmt.Errorf("no local config found to migrate: %w", err)
	}

	// Update paths to use global directory
	globalDir := GetGlobalUngDir()
	cfg.DatabasePath = filepath.Join(globalDir, "ung.db")
	cfg.InvoicesDir = filepath.Join(globalDir, "invoices")
	cfg.ContractsDir = filepath.Join(globalDir, "contracts")

	// Ensure global directory exists
	if err := InitGlobalDirectory(); err != nil {
		return err
	}

	// Save to global
	return Save(cfg, true)
}

// MigrateGlobalToLocal copies the global config to local
func MigrateGlobalToLocal() error {
	// Load global config
	globalConfigPath := filepath.Join(GetGlobalUngDir(), "config.yaml")
	cfg, err := loadFromFile(globalConfigPath)
	if err != nil {
		return fmt.Errorf("no global config found to migrate: %w", err)
	}

	// Update paths to use local directory
	cfg.DatabasePath = filepath.Join(LocalUngDir, "ung.db")
	cfg.InvoicesDir = filepath.Join(LocalUngDir, "invoices")
	cfg.ContractsDir = filepath.Join(LocalUngDir, "contracts")

	// Ensure local directory exists
	if err := InitLocalDirectory(); err != nil {
		return err
	}

	// Save to local
	return Save(cfg, false)
}

// GetContractsDir returns the configured contracts directory
func GetContractsDir() string {
	cfg, _ := Load()
	if cfg.ContractsDir != "" {
		return expandPath(cfg.ContractsDir)
	}
	// Fallback: derive from invoices dir
	return filepath.Join(filepath.Dir(cfg.InvoicesDir), "contracts")
}

// IsUsingLocalConfig returns true if the current config is from local .ung directory
func IsUsingLocalConfig() bool {
	return configSource == SourceLocal
}

// GetActiveConfigPath returns the path to the currently active config file
func GetActiveConfigPath() string {
	switch configSource {
	case SourceLocal:
		absPath, _ := filepath.Abs(filepath.Join(LocalUngDir, "config.yaml"))
		return absPath
	case SourceGlobal:
		return filepath.Join(GetGlobalUngDir(), "config.yaml")
	default:
		return "(using defaults)"
	}
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

// IsInitialized checks if UNG has been initialized (either local or global .ung directory exists with content)
func IsInitialized() bool {
	// Check local .ung directory first
	if isDirectoryInitialized(LocalUngDir) {
		return true
	}

	// Check global ~/.ung directory
	globalDir := GetGlobalUngDir()
	if globalDir != "" && isDirectoryInitialized(globalDir) {
		return true
	}

	return false
}

// isDirectoryInitialized checks if a .ung directory is properly initialized
// A directory is considered initialized if it contains either a database file or config file
func isDirectoryInitialized(dir string) bool {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check for database file
	dbPath := filepath.Join(dir, "ung.db")
	if _, err := os.Stat(dbPath); err == nil {
		return true
	}

	// Check for encrypted database file
	encryptedDBPath := filepath.Join(dir, "ung.db.encrypted")
	if _, err := os.Stat(encryptedDBPath); err == nil {
		return true
	}

	// Check for config file
	configPath := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	return false
}

// GetInitializedDir returns the path to the initialized .ung directory, or empty string if not initialized
func GetInitializedDir() string {
	// Check local first
	if isDirectoryInitialized(LocalUngDir) {
		absPath, _ := filepath.Abs(LocalUngDir)
		return absPath
	}

	// Check global
	globalDir := GetGlobalUngDir()
	if globalDir != "" && isDirectoryInitialized(globalDir) {
		return globalDir
	}

	return ""
}
