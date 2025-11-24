package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Andriiklymiuk/ung/internal/cloud"
	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ung configuration",
	Long: `Manage ung configuration files.

Configuration priority (highest to lowest):
1. Local workspace: .ung.yaml (in current directory)
2. Global: ~/.ung/config.yaml

Use local workspace config for project-specific databases.
Use global config for your default settings.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize workspace configuration",
	Long: `Create a local .ung.yaml configuration file in the current directory.

This is useful for project-specific databases where you want to keep
your billing data separate from the global database.

Example workspace config:
  database_path: ./ung.db
  invoices_dir: ./invoices
  language: en`,
	Run: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Show the current configuration and where it's loaded from.`,
	Run:   runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file paths",
	Long:  `Display the paths to config files (local and global).`,
	Run:   runConfigPath,
}

var (
	configInitLocal  bool
	configInitICloud bool
)

func init() {
	configInitCmd.Flags().BoolVarP(&configInitLocal, "local", "l", true, "Create local workspace config (.ung.yaml)")
	configInitCmd.Flags().BoolVarP(&configInitLocal, "global", "g", false, "Create global config (~/.ung/config.yaml)")
	configInitCmd.Flags().BoolVar(&configInitICloud, "icloud", false, "Use iCloud Drive for storage (macOS only)")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) {
	var configPath string
	var dbPath string
	var invoicesDir string

	// Check for iCloud flag
	if configInitICloud {
		if !cloud.ICloudAvailable() {
			fmt.Println("❌ iCloud Drive is not available")
			fmt.Println("   Make sure you're on macOS and iCloud Drive is enabled")
			return
		}

		iCloudDB, iCloudInvoices, err := cloud.GetOptimalPath(true)
		if err != nil {
			fmt.Printf("❌ Failed to get iCloud paths: %v\n", err)
			return
		}

		dbPath = iCloudDB
		invoicesDir = iCloudInvoices

		if configInitLocal {
			configPath = ".ung.yaml"
			fmt.Println("☁️  Creating local workspace configuration with iCloud storage...")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("❌ Failed to get home directory: %v\n", err)
				return
			}
			configPath = filepath.Join(home, ".ung", "config.yaml")
			fmt.Println("☁️  Creating global configuration with iCloud storage...")
		}

		fmt.Println("\n✨ iCloud Drive sync enabled!")
		fmt.Println("   Your data will sync across all your Apple devices")
	} else if configInitLocal {
		// Local workspace config
		configPath = ".ung.yaml"
		dbPath = "./ung.db"
		invoicesDir = "./invoices"

		fmt.Println("🔧 Creating local workspace configuration...")
		fmt.Println("\n💡 This will create a project-specific database.")
	} else {
		// Global config
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("❌ Failed to get home directory: %v\n", err)
			return
		}
		configPath = filepath.Join(home, ".ung", "config.yaml")
		dbPath = filepath.Join(home, ".ung", "ung.db")
		invoicesDir = filepath.Join(home, ".ung", "invoices")

		fmt.Println("🔧 Creating global configuration...")
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Configuration file already exists: %s\n", configPath)
		fmt.Print("\n❓ Overwrite? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled")
			return
		}
	}

	// Create config
	cfg := &config.Config{
		DatabasePath: dbPath,
		InvoicesDir:  invoicesDir,
		Language:     "en",
		Invoice: config.InvoiceConfig{
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
	}

	if err := config.Save(cfg, !configInitLocal); err != nil {
		fmt.Printf("❌ Failed to save configuration: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Configuration created: %s\n", configPath)
	fmt.Println("\n📝 Configuration details:")
	fmt.Printf("   Database:  %s\n", dbPath)
	fmt.Printf("   Invoices:  %s\n", invoicesDir)
	fmt.Printf("   Language:  en\n")

	// Show iCloud status
	if cloud.IsSyncedToiCloud(dbPath) {
		fmt.Println("\n☁️  iCloud Sync: Enabled")
		fmt.Println("   Status:     " + cloud.GetSyncStatus(dbPath))
		fmt.Println("   Device:     Data syncs across all Apple devices")
		fmt.Println("   iOS App:    Can access this database")
	}

	if configInitLocal {
		fmt.Println("\n💡 Tips:")
		if !configInitICloud {
			fmt.Println("   - Add ung.db and invoices/ to .gitignore")
		}
		fmt.Println("   - This project will now use its own database")
		fmt.Println("   - Run 'ung doctor' to verify the setup")
		if configInitICloud {
			fmt.Println("   - Data automatically syncs via iCloud Drive")
			fmt.Println("   - Use 'ung cloud status' to check sync status")
		}
	}
}

func runConfigShow(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		return
	}

	// Determine which config file is being used
	configSource := "default (no config file)"
	if _, err := os.Stat(".ung.yaml"); err == nil {
		configSource = ".ung.yaml (local workspace)"
	} else {
		home, _ := os.UserHomeDir()
		globalConfig := filepath.Join(home, ".ung", "config.yaml")
		if _, err := os.Stat(globalConfig); err == nil {
			configSource = globalConfig + " (global)"
		}
	}

	fmt.Printf("📋 Current Configuration\n\n")
	fmt.Printf("Source: %s\n\n", configSource)
	fmt.Printf("Database Path:  %s\n", cfg.DatabasePath)
	fmt.Printf("Invoices Dir:   %s\n", cfg.InvoicesDir)
	fmt.Printf("Language:       %s\n\n", cfg.Language)

	fmt.Println("Invoice Labels:")
	fmt.Printf("  Invoice:      %s\n", cfg.Invoice.InvoiceLabel)
	fmt.Printf("  From:         %s\n", cfg.Invoice.FromLabel)
	fmt.Printf("  Bill To:      %s\n", cfg.Invoice.BillToLabel)
	fmt.Printf("  Description:  %s\n", cfg.Invoice.DescriptionLabel)
	fmt.Printf("  Total:        %s\n", cfg.Invoice.TotalLabel)
}

func runConfigPath(cmd *cobra.Command, args []string) {
	home, _ := os.UserHomeDir()
	globalConfig := filepath.Join(home, ".ung", "config.yaml")
	localConfig := ".ung.yaml"

	fmt.Println("📂 Configuration File Paths\n")

	// Check local config
	if _, err := os.Stat(localConfig); err == nil {
		fmt.Printf("✅ Local:  %s (active)\n", localConfig)
	} else {
		fmt.Printf("⚪ Local:  %s (not found)\n", localConfig)
	}

	// Check global config
	if _, err := os.Stat(globalConfig); err == nil {
		if _, err := os.Stat(localConfig); err == nil {
			fmt.Printf("⚪ Global: %s (overridden by local)\n", globalConfig)
		} else {
			fmt.Printf("✅ Global: %s (active)\n", globalConfig)
		}
	} else {
		fmt.Printf("⚪ Global: %s (not found)\n", globalConfig)
	}

	fmt.Println("\n💡 Priority: Local workspace (.ung.yaml) > Global (~/.ung/config.yaml)")
}
