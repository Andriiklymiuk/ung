package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/pkg/invoice"
	"github.com/spf13/cobra"
)

var invoiceCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Manage invoices",
	Long:  "Create, list, and manage invoices",
}

var invoiceNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new invoice",
	RunE:  runInvoiceNew,
}

var invoiceListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all invoices",
	RunE:    runInvoiceList,
}

var invoicePDFCmd = &cobra.Command{
	Use:   "pdf [invoice-id]",
	Short: "Generate PDF for an invoice",
	Args:  cobra.ExactArgs(1),
	RunE:  runInvoicePDF,
}

var (
	invoiceCompanyID   int
	invoiceClientID    int
	invoiceAmount      float64
	invoiceCurrency    string
	invoiceDescription string
	invoiceDueDate     string
)

func init() {
	// Add subcommands
	invoiceCmd.AddCommand(invoiceNewCmd)
	invoiceCmd.AddCommand(invoiceListCmd)
	invoiceCmd.AddCommand(invoicePDFCmd)

	// New invoice flags
	invoiceNewCmd.Flags().IntVar(&invoiceCompanyID, "company", 0, "Company ID (required)")
	invoiceNewCmd.Flags().IntVar(&invoiceClientID, "client", 0, "Client ID (required)")
	invoiceNewCmd.Flags().Float64Var(&invoiceAmount, "price", 0, "Invoice amount (required)")
	invoiceNewCmd.Flags().StringVar(&invoiceCurrency, "currency", "USD", "Currency code")
	invoiceNewCmd.Flags().StringVar(&invoiceDescription, "description", "", "Invoice description")
	invoiceNewCmd.Flags().StringVar(&invoiceDueDate, "due", "", "Due date (YYYY-MM-DD)")
	invoiceNewCmd.MarkFlagRequired("company")
	invoiceNewCmd.MarkFlagRequired("client")
	invoiceNewCmd.MarkFlagRequired("price")
}

func generateInvoiceNumber() (string, error) {
	year := time.Now().Year()
	query := `SELECT COUNT(*) FROM invoices WHERE invoice_num LIKE ?`
	var count int
	err := db.DB.QueryRow(query, fmt.Sprintf("INV-%d-%%", year)).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%d-%03d", year, count+1), nil
}

func runInvoiceNew(cmd *cobra.Command, args []string) error {
	// Generate invoice number
	invoiceNum, err := generateInvoiceNumber()
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Parse due date
	var dueDate time.Time
	if invoiceDueDate != "" {
		dueDate, err = time.Parse("2006-01-02", invoiceDueDate)
		if err != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
		}
	} else {
		// Default: 30 days from now
		dueDate = time.Now().AddDate(0, 0, 30)
	}

	// Insert invoice
	query := `
		INSERT INTO invoices (invoice_num, company_id, amount, currency, description, status, issued_date, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.DB.Exec(query,
		invoiceNum,
		invoiceCompanyID,
		invoiceAmount,
		invoiceCurrency,
		invoiceDescription,
		models.StatusPending,
		time.Now(),
		dueDate,
	)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	invoiceID, _ := result.LastInsertId()

	// Link invoice to client
	_, err = db.DB.Exec(
		"INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)",
		invoiceID, invoiceClientID,
	)
	if err != nil {
		return fmt.Errorf("failed to link invoice to client: %w", err)
	}

	fmt.Printf("✓ Invoice created successfully\n")
	fmt.Printf("  Invoice Number: %s\n", invoiceNum)
	fmt.Printf("  Invoice ID: %d\n", invoiceID)
	fmt.Printf("  Amount: %.2f %s\n", invoiceAmount, invoiceCurrency)
	fmt.Printf("  Due Date: %s\n", dueDate.Format("2006-01-02"))
	return nil
}

func runInvoiceList(cmd *cobra.Command, args []string) error {
	query := `
		SELECT i.id, i.invoice_num, i.amount, i.currency, i.status, i.issued_date, i.due_date
		FROM invoices i
		ORDER BY i.id DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNUMBER\tAMOUNT\tSTATUS\tISSUED\tDUE")

	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.Amount, &inv.Currency,
			&inv.Status, &inv.IssuedDate, &inv.DueDate); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		fmt.Fprintf(w, "%d\t%s\t%.2f %s\t%s\t%s\t%s\n",
			inv.ID,
			inv.InvoiceNum,
			inv.Amount,
			inv.Currency,
			inv.Status,
			inv.IssuedDate.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
		)
	}

	w.Flush()
	return nil
}

func runInvoicePDF(cmd *cobra.Command, args []string) error {
	invoiceID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Fetch invoice details
	var inv models.Invoice
	var company models.Company
	var client models.Client

	// Get invoice
	err = db.DB.QueryRow(`
		SELECT id, invoice_num, company_id, amount, currency, description, status, issued_date, due_date
		FROM invoices WHERE id = ?
	`, invoiceID).Scan(
		&inv.ID, &inv.InvoiceNum, &inv.CompanyID, &inv.Amount, &inv.Currency,
		&inv.Description, &inv.Status, &inv.IssuedDate, &inv.DueDate,
	)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Get company
	err = db.DB.QueryRow(`
		SELECT id, name, email, address, tax_id
		FROM companies WHERE id = ?
	`, inv.CompanyID).Scan(&company.ID, &company.Name, &company.Email, &company.Address, &company.TaxID)
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Get client
	err = db.DB.QueryRow(`
		SELECT c.id, c.name, c.email, c.address, c.tax_id
		FROM clients c
		JOIN invoice_recipients ir ON c.id = ir.client_id
		WHERE ir.invoice_id = ?
	`, invoiceID).Scan(&client.ID, &client.Name, &client.Email, &client.Address, &client.TaxID)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Generate PDF
	pdfPath, err := invoice.GeneratePDF(inv, company, client)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Update invoice with PDF path
	_, err = db.DB.Exec("UPDATE invoices SET pdf_path = ? WHERE id = ?", pdfPath, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	fmt.Printf("✓ PDF generated successfully: %s\n", pdfPath)
	return nil
}
