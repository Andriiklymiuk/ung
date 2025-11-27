package cmd

import (
	"fmt"
	"os"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/spf13/cobra"
)

// globalFlag indicates whether to use global config
var globalFlag bool

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set global flag in config package
		config.SetForceGlobal(globalFlag)
	},
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
