package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ung",
	Short: "🐾 UNG — Universal Next-Gen Billing & Tracking",
	Long: `UNG is a fast, cross-platform CLI tool for managing company details,
clients, invoices, and time tracking.

All data is stored locally in ~/.ung/ung.db`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Register subcommands
	rootCmd.AddCommand(companyCmd)
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(contractCmd)
	rootCmd.AddCommand(invoiceCmd)
	rootCmd.AddCommand(trackCmd)

	// Add update as alias for upgrade (reuse upgrade functionality)
	updateCmd := *upgradeCmd
	updateCmd.Use = "update"
	updateCmd.Aliases = []string{}
	rootCmd.AddCommand(&updateCmd)
}
