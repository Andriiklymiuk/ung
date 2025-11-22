package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Interactive creation wizard",
	Long:  "Launch an interactive wizard to create companies, clients, contracts, invoices, or track time",
	RunE:  runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	fmt.Println("🎯 UNG Creation Wizard")
	fmt.Println("What would you like to create?\n")

	var selection string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose what to create").
				Options(
					huh.NewOption("Company - Your business information", "company"),
					huh.NewOption("Client - Customer or client details", "client"),
					huh.NewOption("Contract - Work agreement with a client", "contract"),
					huh.NewOption("Invoice - Bill for services rendered", "invoice"),
					huh.NewOption("Track Time - Log hours worked", "track"),
				).
				Value(&selection),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("creation cancelled: %w", err)
	}

	fmt.Printf("\n✨ Creating %s...\n\n", selection)

	// Dispatch to appropriate creation function
	switch selection {
	case "company":
		return runCompanyAdd(cmd, args)
	case "client":
		return runClientAdd(cmd, args)
	case "contract":
		return runContractAdd(cmd, args)
	case "invoice":
		return runInvoiceNew(cmd, args)
	case "track":
		return runTrackLog(cmd, args)
	default:
		return fmt.Errorf("unknown selection: %s", selection)
	}
}
