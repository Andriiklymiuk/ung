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
1. Local workspace: .ung/config.yaml (in current directory)
2. Legacy local: .ung.yaml (deprecated, for backwards compatibility)
3. Global: ~/.ung/config.yaml

Use local workspace config for project-specific databases.
Use global config for your default settings.

Similar to VS Code, when you have a local .ung folder, it will be used
instead of the global configuration for database, invoices, and contracts.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize workspace configuration",
	Long: `Create a local .ung/ directory structure in the current directory.

This creates:
  .ung/
  ├── config.yaml    # Configuration file
  ├── ung.db         # SQLite database (created on first use)
  ├── invoices/      # Generated invoice PDFs
  └── contracts/     # Generated contract PDFs

This is useful for project-specific databases where you want to keep
your billing data separate from the global database.

Use --global to initialize the global ~/.ung/ directory instead.`,
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

var configMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate configuration between local and global",
	Long: `Migrate configuration settings between local and global.

Use --to-global to copy local config settings to global.
Use --to-local to copy global config settings to local.

Note: This copies configuration settings only, not data files.
Use 'ung sync' or 'ung database' commands to manage data.`,
	Run: runConfigMigrate,
}

var (
	configInitGlobal   bool
	configInitICloud   bool
	configMigrateToGlobal bool
	configMigrateToLocal  bool
)

func init() {
	configInitCmd.Flags().BoolVarP(&configInitGlobal, "global", "g", false, "Initialize global config (~/.ung/) instead of local")
	configInitCmd.Flags().BoolVar(&configInitICloud, "icloud", false, "Use iCloud Drive for storage (macOS only)")

	configMigrateCmd.Flags().BoolVar(&configMigrateToGlobal, "to-global", false, "Copy local config to global")
	configMigrateCmd.Flags().BoolVar(&configMigrateToLocal, "to-local", false, "Copy global config to local")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configMigrateCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) {
	var configPath string
	var dbPath string
	var invoicesDir string
	var contractsDir string
	var ungDir string

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
		contractsDir = filepath.Join(filepath.Dir(iCloudInvoices), "contracts")

		if !configInitGlobal {
			ungDir = config.LocalUngDir
			configPath = filepath.Join(ungDir, "config.yaml")
			fmt.Println("☁️  Creating local workspace configuration with iCloud storage...")
		} else {
			ungDir = config.GetGlobalUngDir()
			configPath = filepath.Join(ungDir, "config.yaml")
			fmt.Println("☁️  Creating global configuration with iCloud storage...")
		}

		fmt.Println("\n✨ iCloud Drive sync enabled!")
		fmt.Println("   Your data will sync across all your Apple devices")
	} else if !configInitGlobal {
		// Local workspace config - use new .ung/ directory structure
		ungDir = config.LocalUngDir
		configPath = filepath.Join(ungDir, "config.yaml")
		dbPath = filepath.Join(ungDir, "ung.db")
		invoicesDir = filepath.Join(ungDir, "invoices")
		contractsDir = filepath.Join(ungDir, "contracts")

		fmt.Println("🔧 Creating local workspace configuration...")
		fmt.Println("\n💡 This will create a project-specific .ung/ directory.")
	} else {
		// Global config
		ungDir = config.GetGlobalUngDir()
		if ungDir == "" {
			fmt.Println("❌ Failed to get home directory")
			return
		}
		configPath = filepath.Join(ungDir, "config.yaml")
		dbPath = filepath.Join(ungDir, "ung.db")
		invoicesDir = filepath.Join(ungDir, "invoices")
		contractsDir = filepath.Join(ungDir, "contracts")

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

	// Create the directory structure
	dirs := []string{ungDir, invoicesDir, contractsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory %s: %v\n", dir, err)
			return
		}
	}

	// Create config
	cfg := &config.Config{
		DatabasePath: dbPath,
		InvoicesDir:  invoicesDir,
		ContractsDir: contractsDir,
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

	if err := config.Save(cfg, configInitGlobal); err != nil {
		fmt.Printf("❌ Failed to save configuration: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Configuration created: %s\n", configPath)
	fmt.Println("\n📁 Directory structure:")
	fmt.Printf("   %s/\n", ungDir)
	fmt.Println("   ├── config.yaml")
	fmt.Println("   ├── ung.db (created on first use)")
	fmt.Println("   ├── invoices/")
	fmt.Println("   └── contracts/")
	fmt.Println("\n📝 Configuration details:")
	fmt.Printf("   Database:  %s\n", dbPath)
	fmt.Printf("   Invoices:  %s\n", invoicesDir)
	fmt.Printf("   Contracts: %s\n", contractsDir)
	fmt.Printf("   Language:  en\n")

	// Show iCloud status
	if cloud.IsSyncedToiCloud(dbPath) {
		fmt.Println("\n☁️  iCloud Sync: Enabled")
		fmt.Println("   Status:     " + cloud.GetSyncStatus(dbPath))
		fmt.Println("   Device:     Data syncs across all Apple devices")
		fmt.Println("   iOS App:    Can access this database")
	}

	if !configInitGlobal {
		fmt.Println("\n💡 Tips:")
		if !configInitICloud {
			fmt.Println("   - Add .ung/ to .gitignore to exclude billing data from git")
		}
		fmt.Println("   - This project will now use its own database")
		fmt.Println("   - Run 'ung doctor' to verify the setup")
		fmt.Println("   - Use 'ung config show' to see active configuration")
		if configInitICloud {
			fmt.Println("   - Data automatically syncs via iCloud Drive")
			fmt.Println("   - Use 'ung cloud status' to check sync status")
		}
	}
}

func runConfigMigrate(cmd *cobra.Command, args []string) {
	if !configMigrateToGlobal && !configMigrateToLocal {
		fmt.Println("❌ Please specify migration direction:")
		fmt.Println("   --to-global   Copy local config to global")
		fmt.Println("   --to-local    Copy global config to local")
		return
	}

	if configMigrateToGlobal && configMigrateToLocal {
		fmt.Println("❌ Cannot specify both --to-global and --to-local")
		return
	}

	if configMigrateToGlobal {
		fmt.Println("🔄 Migrating local configuration to global...")
		if err := config.MigrateLocalToGlobal(); err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)
			return
		}
		fmt.Println("✅ Local configuration migrated to global!")
		fmt.Printf("   Global config: %s\n", filepath.Join(config.GetGlobalUngDir(), "config.yaml"))
	} else {
		fmt.Println("🔄 Migrating global configuration to local...")
		if err := config.MigrateGlobalToLocal(); err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)
			return
		}
		fmt.Println("✅ Global configuration migrated to local!")
		fmt.Printf("   Local config: %s\n", filepath.Join(config.LocalUngDir, "config.yaml"))
	}

	fmt.Println("\n💡 Note: Only configuration settings were migrated.")
	fmt.Println("   Data files (database, PDFs) remain in their original locations.")
	fmt.Println("   Use 'ung sync' or 'ung database' commands to manage data.")
}

func runConfigShow(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		return
	}

	fmt.Printf("📋 Current Configuration\n\n")
	fmt.Printf("Source: %s\n", config.GetConfigSourceString())
	fmt.Printf("Path:   %s\n\n", config.GetActiveConfigPath())

	fmt.Println("📁 Storage Paths:")
	fmt.Printf("   Database:   %s\n", cfg.DatabasePath)
	fmt.Printf("   Invoices:   %s\n", cfg.InvoicesDir)
	fmt.Printf("   Contracts:  %s\n", config.GetContractsDir())
	fmt.Printf("   Language:   %s\n\n", cfg.Language)

	fmt.Println("📝 Invoice Labels:")
	fmt.Printf("   Invoice:      %s\n", cfg.Invoice.InvoiceLabel)
	fmt.Printf("   From:         %s\n", cfg.Invoice.FromLabel)
	fmt.Printf("   Bill To:      %s\n", cfg.Invoice.BillToLabel)
	fmt.Printf("   Description:  %s\n", cfg.Invoice.DescriptionLabel)
	fmt.Printf("   Total:        %s\n", cfg.Invoice.TotalLabel)

	// Show if using local or global
	if config.IsUsingLocalConfig() {
		fmt.Println("\n💡 Using local workspace configuration.")
		fmt.Println("   Use --global flag to use global config instead.")
	} else if config.GetConfigSource() == config.SourceGlobal {
		fmt.Println("\n💡 Using global configuration.")
		fmt.Println("   Run 'ung config init' to create a local workspace config.")
	}
}

func runConfigPath(cmd *cobra.Command, args []string) {
	globalDir := config.GetGlobalUngDir()
	globalConfig := filepath.Join(globalDir, "config.yaml")
	localConfig := filepath.Join(config.LocalUngDir, "config.yaml")
	legacyLocalConfig := ".ung.yaml"

	fmt.Println("📂 Configuration File Paths")
	fmt.Println()

	// Determine active config
	activeSource := config.GetConfigSource()

	// Check new local config (.ung/config.yaml)
	localAbsPath, _ := filepath.Abs(localConfig)
	if _, err := os.Stat(localConfig); err == nil {
		if activeSource == config.SourceLocal {
			fmt.Printf("✅ Local:        %s (active)\n", localAbsPath)
		} else {
			fmt.Printf("⚪ Local:        %s (exists)\n", localAbsPath)
		}
	} else {
		fmt.Printf("⚪ Local:        %s (not found)\n", localAbsPath)
	}

	// Check legacy local config (.ung.yaml)
	legacyAbsPath, _ := filepath.Abs(legacyLocalConfig)
	if _, err := os.Stat(legacyLocalConfig); err == nil {
		if activeSource == config.SourceLocalLegacy {
			fmt.Printf("✅ Local legacy: %s (active)\n", legacyAbsPath)
		} else {
			fmt.Printf("⚠️  Local legacy: %s (deprecated, consider migrating)\n", legacyAbsPath)
		}
	}

	// Check global config
	if _, err := os.Stat(globalConfig); err == nil {
		if activeSource == config.SourceGlobal {
			fmt.Printf("✅ Global:       %s (active)\n", globalConfig)
		} else {
			fmt.Printf("⚪ Global:       %s (overridden by local)\n", globalConfig)
		}
	} else {
		fmt.Printf("⚪ Global:       %s (not found)\n", globalConfig)
	}

	fmt.Println("\n💡 Priority (highest to lowest):")
	fmt.Println("   1. Local:  .ung/config.yaml")
	fmt.Println("   2. Legacy: .ung.yaml (deprecated)")
	fmt.Println("   3. Global: ~/.ung/config.yaml")

	if activeSource == config.SourceDefault {
		fmt.Println("\n⚠️  No config file found, using defaults.")
		fmt.Println("   Run 'ung config init' to create a local config.")
		fmt.Println("   Run 'ung config init --global' to create a global config.")
	}
}
