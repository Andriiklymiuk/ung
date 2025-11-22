package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/internal/repository"
	"github.com/spf13/cobra"
)

var companyCmd = &cobra.Command{
	Use:   "company",
	Short: "Manage your company details",
	Long:  "Add, list, and edit your business information",
}

var companyAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new company",
	RunE:  runCompanyAdd,
}

var companyListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all companies",
	RunE:    runCompanyList,
}

var companyEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing company",
	Args:  cobra.ExactArgs(1),
	RunE:  runCompanyEdit,
}

var (
	companyName    string
	companyEmail   string
	companyAddress string
	companyTaxID   string
)

func init() {
	// Add subcommands
	companyCmd.AddCommand(companyAddCmd)
	companyCmd.AddCommand(companyListCmd)
	companyCmd.AddCommand(companyEditCmd)

	// Add flags
	companyAddCmd.Flags().StringVar(&companyName, "name", "", "Company name (required)")
	companyAddCmd.Flags().StringVar(&companyEmail, "email", "", "Company email (required)")
	companyAddCmd.Flags().StringVar(&companyAddress, "address", "", "Company address")
	companyAddCmd.Flags().StringVar(&companyTaxID, "tax-id", "", "Tax ID")
	companyAddCmd.MarkFlagRequired("name")
	companyAddCmd.MarkFlagRequired("email")

	// Edit flags
	companyEditCmd.Flags().StringVar(&companyName, "name", "", "Company name")
	companyEditCmd.Flags().StringVar(&companyEmail, "email", "", "Company email")
	companyEditCmd.Flags().StringVar(&companyAddress, "address", "", "Company address")
	companyEditCmd.Flags().StringVar(&companyTaxID, "tax-id", "", "Tax ID")
}

func runCompanyAdd(cmd *cobra.Command, args []string) error {
	repo := repository.NewCompanyRepository()

	company := &models.Company{
		Name:    companyName,
		Email:   companyEmail,
		Address: companyAddress,
		TaxID:   companyTaxID,
	}

	if err := repo.Create(company); err != nil {
		return fmt.Errorf("failed to add company: %w", err)
	}

	fmt.Printf("✓ Company added successfully (ID: %d)\n", company.ID)
	return nil
}

func runCompanyList(cmd *cobra.Command, args []string) error {
	repo := repository.NewCompanyRepository()

	companies, err := repo.List()
	if err != nil {
		return fmt.Errorf("failed to query companies: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tADDRESS\tTAX ID\tCREATED")

	for _, c := range companies {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			c.ID, c.Name, c.Email, c.Address, c.TaxID, c.CreatedAt.Format("2006-01-02"))
	}

	w.Flush()
	return nil
}

func runCompanyEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid company ID: %w", err)
	}

	repo := repository.NewCompanyRepository()

	// Fetch existing company
	company, err := repo.GetByID(uint(id))
	if err != nil {
		return fmt.Errorf("company with ID %d not found: %w", id, err)
	}

	// Update only changed fields
	updated := false
	if cmd.Flags().Changed("name") {
		company.Name = companyName
		updated = true
	}
	if cmd.Flags().Changed("email") {
		company.Email = companyEmail
		updated = true
	}
	if cmd.Flags().Changed("address") {
		company.Address = companyAddress
		updated = true
	}
	if cmd.Flags().Changed("tax-id") {
		company.TaxID = companyTaxID
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update")
	}

	if err := repo.Update(company); err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	fmt.Printf("✓ Company %d updated successfully\n", id)
	return nil
}
