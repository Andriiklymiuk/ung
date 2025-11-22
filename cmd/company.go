package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
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
	query := `
		INSERT INTO companies (name, email, address, tax_id)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.DB.Exec(query, companyName, companyEmail, companyAddress, companyTaxID)
	if err != nil {
		return fmt.Errorf("failed to add company: %w", err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("✓ Company added successfully (ID: %d)\n", id)
	return nil
}

func runCompanyList(cmd *cobra.Command, args []string) error {
	query := `SELECT id, name, email, address, tax_id, created_at FROM companies ORDER BY id`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query companies: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tADDRESS\tTAX ID\tCREATED")

	for rows.Next() {
		var c models.Company
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Address, &c.TaxID, &c.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
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

	updates := make(map[string]interface{})
	if cmd.Flags().Changed("name") {
		updates["name"] = companyName
	}
	if cmd.Flags().Changed("email") {
		updates["email"] = companyEmail
	}
	if cmd.Flags().Changed("address") {
		updates["address"] = companyAddress
	}
	if cmd.Flags().Changed("tax-id") {
		updates["tax_id"] = companyTaxID
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE companies SET "
	args_list := []interface{}{}
	i := 0
	for key, val := range updates {
		if i > 0 {
			query += ", "
		}
		query += key + " = ?"
		args_list = append(args_list, val)
		i++
	}
	query += ", updated_at = ? WHERE id = ?"
	args_list = append(args_list, time.Now(), id)

	result, err := db.DB.Exec(query, args_list...)
	if err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("company with ID %d not found", id)
	}

	fmt.Printf("✓ Company %d updated successfully\n", id)
	return nil
}
