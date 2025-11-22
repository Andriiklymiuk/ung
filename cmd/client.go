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

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage clients",
	Long:  "Add, list, and edit client information",
}

var clientAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new client",
	RunE:  runClientAdd,
}

var clientListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all clients",
	RunE:    runClientList,
}

var clientEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing client",
	Args:  cobra.ExactArgs(1),
	RunE:  runClientEdit,
}

var (
	clientName    string
	clientEmail   string
	clientAddress string
	clientTaxID   string
)

func init() {
	// Add subcommands
	clientCmd.AddCommand(clientAddCmd)
	clientCmd.AddCommand(clientListCmd)
	clientCmd.AddCommand(clientEditCmd)

	// Add flags
	clientAddCmd.Flags().StringVar(&clientName, "name", "", "Client name (required)")
	clientAddCmd.Flags().StringVar(&clientEmail, "email", "", "Client email (required)")
	clientAddCmd.Flags().StringVar(&clientAddress, "address", "", "Client address")
	clientAddCmd.Flags().StringVar(&clientTaxID, "tax-id", "", "Tax ID")
	clientAddCmd.MarkFlagRequired("name")
	clientAddCmd.MarkFlagRequired("email")

	// Edit flags
	clientEditCmd.Flags().StringVar(&clientName, "name", "", "Client name")
	clientEditCmd.Flags().StringVar(&clientEmail, "email", "", "Client email")
	clientEditCmd.Flags().StringVar(&clientAddress, "address", "", "Client address")
	clientEditCmd.Flags().StringVar(&clientTaxID, "tax-id", "", "Tax ID")
}

func runClientAdd(cmd *cobra.Command, args []string) error {
	query := `
		INSERT INTO clients (name, email, address, tax_id)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.DB.Exec(query, clientName, clientEmail, clientAddress, clientTaxID)
	if err != nil {
		return fmt.Errorf("failed to add client: %w", err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("✓ Client added successfully (ID: %d)\n", id)
	return nil
}

func runClientList(cmd *cobra.Command, args []string) error {
	query := `SELECT id, name, email, address, tax_id, created_at FROM clients ORDER BY id`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tADDRESS\tTAX ID\tCREATED")

	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Address, &c.TaxID, &c.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			c.ID, c.Name, c.Email, c.Address, c.TaxID, c.CreatedAt.Format("2006-01-02"))
	}

	w.Flush()
	return nil
}

func runClientEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid client ID: %w", err)
	}

	updates := make(map[string]interface{})
	if cmd.Flags().Changed("name") {
		updates["name"] = clientName
	}
	if cmd.Flags().Changed("email") {
		updates["email"] = clientEmail
	}
	if cmd.Flags().Changed("address") {
		updates["address"] = clientAddress
	}
	if cmd.Flags().Changed("tax-id") {
		updates["tax_id"] = clientTaxID
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE clients SET "
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
		return fmt.Errorf("failed to update client: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("client with ID %d not found", id)
	}

	fmt.Printf("✓ Client %d updated successfully\n", id)
	return nil
}
