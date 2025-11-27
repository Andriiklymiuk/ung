package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/spf13/cobra"
)

// globalFlag indicates whether to use global config
var globalFlag bool

// dbInitialized tracks if database was successfully initialized
var dbInitialized bool

// commandsWithoutDB is a list of commands that don't require database initialization
var commandsWithoutDB = map[string]bool{
	"config":   true,
	"version":  true,
	"help":     true,
	"doctor":   true,
	"upgrade":  true,
	"update":   true,
	"docs":     true,
	"__complete": true,
	"completion": true,
}

var rootCmd = &cobra.Command{
	Use:   "ung",
	Short: "🐾 UNG — Universal Next-Gen Billing & Tracking",
	Long: `UNG is a fast, cross-platform CLI tool for managing company details,
clients, invoices, and time tracking.

Data storage (similar to VS Code settings):
- Local:  .ung/ folder in current directory (project-specific)
- Global: ~/.ung/ folder (default when no local config exists)

Use 'ung config init' to create a local workspace configuration.
Use --global flag to explicitly use global configuration.`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set global flag in config package
		config.SetForceGlobal(globalFlag)

		// Check if this command requires database
		cmdName := cmd.Name()
		if cmd.Parent() != nil && cmd.Parent().Name() != "ung" {
			// Use parent command name for subcommands
			parent := cmd.Parent()
			for parent.Parent() != nil && parent.Parent().Name() != "ung" {
				parent = parent.Parent()
			}
			cmdName = parent.Name()
		}

		// Skip DB initialization for commands that don't need it
		if commandsWithoutDB[cmdName] {
			return nil
		}

		// Try to initialize database
		if err := db.Initialize(); err != nil {
			if errors.Is(err, db.ErrNotInitialized) {
				showOnboarding()
				os.Exit(0)
			}
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		dbInitialized = true
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Close database if it was initialized
		if dbInitialized {
			db.Close()
		}
	},
}

// showOnboarding displays the welcome/onboarding message for new users
func showOnboarding() {
	fmt.Println()
	fmt.Println("🐾 Welcome to UNG — Universal Next-Gen Billing & Tracking!")
	fmt.Println()
	fmt.Println("It looks like UNG hasn't been set up yet.")
	fmt.Println()
	fmt.Println("To get started, run one of the following commands:")
	fmt.Println()
	fmt.Println("  ung config init          Create a local workspace (.ung/ in current directory)")
	fmt.Println("  ung config init --global Create a global config (~/.ung/ in home directory)")
	fmt.Println()
	fmt.Println("After initialization, you can:")
	fmt.Println("  • Set up your company:    ung company set")
	fmt.Println("  • Add a client:           ung client add")
	fmt.Println("  • Create a contract:      ung contract add")
	fmt.Println("  • Start time tracking:    ung track start")
	fmt.Println("  • Create an invoice:      ung invoice create")
	fmt.Println()
	fmt.Println("For more help, visit: https://andriiklymiuk.github.io/ung/docs/intro")
	fmt.Println()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add global flag to root command (applies to all subcommands)
	rootCmd.PersistentFlags().BoolVarP(&globalFlag, "global", "G", false, "Use global ~/.ung/ configuration instead of local")

	// Register subcommands
	rootCmd.AddCommand(companyCmd)
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(contractCmd)
	rootCmd.AddCommand(invoiceCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(recurringCmd)

	// Add update as alias for upgrade (reuse upgrade functionality)
	updateCmd := *upgradeCmd
	updateCmd.Use = "update"
	updateCmd.Aliases = []string{}
	rootCmd.AddCommand(&updateCmd)

	// Documentation generation
	rootCmd.AddCommand(docsCmd)
}
