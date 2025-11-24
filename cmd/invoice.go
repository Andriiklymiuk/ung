package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/pkg/idgen"
	"github.com/Andriiklymiuk/ung/pkg/invoice"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var invoiceCmd = &cobra.Command{
	Use:   "invoice [client-name]",
	Short: "Manage invoices or generate from time for a client",
	Long:  "Create, list, and manage invoices. If client name provided, auto-generates invoice from tracked time.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInvoiceSimple,
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

var invoiceEmailCmd = &cobra.Command{
	Use:   "email [invoice-id]",
	Short: "Export invoice to email client",
	Args:  cobra.ExactArgs(1),
	RunE:  runInvoiceEmail,
}

var invoiceBatchEmailCmd = &cobra.Command{
	Use:   "batch-email",
	Short: "Export multiple invoices to email",
	RunE:  runInvoiceBatchEmail,
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
	invoiceCmd.AddCommand(invoiceEmailCmd)
	invoiceCmd.AddCommand(invoiceBatchEmailCmd)

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

func runInvoiceNew(cmd *cobra.Command, args []string) error {
	// Get client name for invoice number generation
	var clientName string
	err := db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", invoiceClientID).Scan(&clientName)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Use current time for issued date
	issuedDate := time.Now()

	// Generate human-readable invoice number
	invoiceNum, err := idgen.GenerateInvoiceNumber(db.GormDB, clientName, issuedDate)
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
		issuedDate,
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

	// Get company with all fields
	err = db.DB.QueryRow(`
		SELECT id, name, email, phone, address, registration_address, tax_id,
		       bank_name, bank_account, bank_swift, logo_path
		FROM companies WHERE id = ?
	`, inv.CompanyID).Scan(&company.ID, &company.Name, &company.Email, &company.Phone,
		&company.Address, &company.RegistrationAddress, &company.TaxID,
		&company.BankName, &company.BankAccount, &company.BankSWIFT, &company.LogoPath)
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

	// Get line items
	var lineItems []models.InvoiceLineItem
	rows, err := db.DB.Query(`
		SELECT id, invoice_id, item_name, description, quantity, rate, amount, created_at
		FROM invoice_line_items
		WHERE invoice_id = ?
		ORDER BY id
	`, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to fetch line items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.InvoiceLineItem
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.ItemName, &item.Description,
			&item.Quantity, &item.Rate, &item.Amount, &item.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan line item: %w", err)
		}
		lineItems = append(lineItems, item)
	}

	// If no line items, create a default one from invoice amount
	if len(lineItems) == 0 {
		lineItems = []models.InvoiceLineItem{
			{
				ItemName:    inv.Description,
				Description: "",
				Quantity:    1.0,
				Rate:        inv.Amount,
				Amount:      inv.Amount,
			},
		}
	}

	// Generate PDF
	pdfPath, err := invoice.GeneratePDF(inv, company, client, lineItems)
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

func runInvoiceEmail(cmd *cobra.Command, args []string) error {
	invoiceID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Fetch invoice details
	var inv models.Invoice
	var company models.Company

	// Get invoice
	err = db.DB.QueryRow(`
		SELECT id, invoice_num, company_id, amount, currency, description, status, issued_date, due_date, pdf_path
		FROM invoices WHERE id = ?
	`, invoiceID).Scan(
		&inv.ID, &inv.InvoiceNum, &inv.CompanyID, &inv.Amount, &inv.Currency,
		&inv.Description, &inv.Status, &inv.IssuedDate, &inv.DueDate, &inv.PDFPath,
	)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Get company info
	err = db.DB.QueryRow(`
		SELECT id, name FROM companies WHERE id = ?
	`, inv.CompanyID).Scan(&company.ID, &company.Name)
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Ensure PDF is generated
	var pdfPath string
	if inv.PDFPath == "" {
		fmt.Println("📄 Generating PDF first...")
		// Call the PDF generation logic
		if err := runInvoicePDF(cmd, args); err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
		// Re-fetch the PDF path
		err = db.DB.QueryRow("SELECT pdf_path FROM invoices WHERE id = ?", invoiceID).Scan(&pdfPath)
		if err != nil {
			return fmt.Errorf("failed to get PDF path: %w", err)
		}
	} else {
		pdfPath = inv.PDFPath
	}

	// Extract month and year from issued date
	month := inv.IssuedDate.Format("01")
	year := inv.IssuedDate.Format("2006")

	// Prepare email details
	subject := fmt.Sprintf("%s.%s %s", month, year, company.Name)
	body := fmt.Sprintf("Hi,\n\nHere is the invoice for %s %s.\n\nBest regards,\n%s",
		inv.IssuedDate.Format("January"), year, company.Name)

	// Prompt for email client selection
	var emailClient string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select email client").
				Options(
					huh.NewOption("Apple Mail", "apple"),
					huh.NewOption("Outlook", "outlook"),
					huh.NewOption("Gmail (Browser)", "gmail"),
				).
				Value(&emailClient),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("email export cancelled: %w", err)
	}

	// Export to selected email client
	switch emailClient {
	case "apple":
		return exportToAppleMail(subject, body, pdfPath)
	case "outlook":
		return exportToOutlook(subject, body, pdfPath)
	case "gmail":
		return exportToGmail(subject, body, pdfPath)
	default:
		return fmt.Errorf("unknown email client: %s", emailClient)
	}
}

// exportToAppleMail opens Apple Mail with prefilled email and attachment
func exportToAppleMail(subject, body, attachmentPath string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("Apple Mail is only available on macOS")
	}

	// Create AppleScript to compose email with attachment
	script := fmt.Sprintf(`
tell application "Mail"
	activate
	set newMessage to make new outgoing message with properties {subject:"%s", content:"%s", visible:true}
	tell newMessage
		make new attachment with properties {file name:POSIX file "%s"} at after the last paragraph
	end tell
end tell
`, escapeAppleScript(subject), escapeAppleScript(body), attachmentPath)

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open Apple Mail: %w", err)
	}

	fmt.Println("✓ Email draft created in Apple Mail with attachment")
	return nil
}

// exportToOutlook opens Outlook with prefilled email
func exportToOutlook(subject, body, attachmentPath string) error {
	// For Outlook, we'll use mailto: link and inform user to manually attach
	// Outlook doesn't support attachments via mailto: protocol
	mailtoURL := fmt.Sprintf("mailto:?subject=%s&body=%s",
		url.QueryEscape(subject),
		url.QueryEscape(body))

	if err := openURL(mailtoURL); err != nil {
		return fmt.Errorf("failed to open Outlook: %w", err)
	}

	fmt.Println("✓ Email draft created in Outlook")
	fmt.Printf("📎 Please manually attach the PDF: %s\n", attachmentPath)
	return nil
}

// exportToGmail opens Gmail in the browser with prefilled email
func exportToGmail(subject, body, attachmentPath string) error {
	// Gmail compose URL format
	gmailURL := fmt.Sprintf("https://mail.google.com/mail/?view=cm&fs=1&su=%s&body=%s",
		url.QueryEscape(subject),
		url.QueryEscape(body))

	if err := openURL(gmailURL); err != nil {
		return fmt.Errorf("failed to open Gmail: %w", err)
	}

	fmt.Println("✓ Gmail compose opened in browser")
	fmt.Printf("📎 Please manually attach the PDF: %s\n", attachmentPath)
	return nil
}

// openURL opens a URL in the default browser or email client
func openURL(urlStr string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", urlStr)
	case "linux":
		cmd = exec.Command("xdg-open", urlStr)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", urlStr)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}

// escapeAppleScript escapes special characters for AppleScript strings
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func runInvoiceBatchEmail(cmd *cobra.Command, args []string) error {
	// Prompt for export option
	var exportOption string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What invoices would you like to export?").
				Options(
					huh.NewOption("Export for latest month", "latest"),
					huh.NewOption("Select specific month/year", "specific"),
					huh.NewOption("Export all pending invoices", "all_pending"),
				).
				Value(&exportOption),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("export cancelled: %w", err)
	}

	var invoices []models.Invoice

	switch exportOption {
	case "latest":
		// Get invoices from the latest month
		query := `
			SELECT id, invoice_num, company_id, amount, currency, description, status, issued_date, due_date, pdf_path
			FROM invoices
			WHERE strftime('%Y-%m', issued_date) = strftime('%Y-%m', 'now')
			ORDER BY issued_date DESC
		`
		rows, err := db.DB.Query(query)
		if err != nil {
			return fmt.Errorf("failed to fetch invoices: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var inv models.Invoice
			if err := rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.CompanyID, &inv.Amount, &inv.Currency,
				&inv.Description, &inv.Status, &inv.IssuedDate, &inv.DueDate, &inv.PDFPath); err != nil {
				return fmt.Errorf("failed to scan invoice: %w", err)
			}
			invoices = append(invoices, inv)
		}

	case "specific":
		// Prompt for month and year
		var monthStr, yearStr string
		monthForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Month (1-12)").
					Placeholder("e.g., 10").
					Value(&monthStr).
					Validate(func(s string) error {
						month, err := strconv.Atoi(s)
						if err != nil || month < 1 || month > 12 {
							return fmt.Errorf("please enter a valid month (1-12)")
						}
						return nil
					}),
				huh.NewInput().
					Title("Year").
					Placeholder("e.g., 2024").
					Value(&yearStr).
					Validate(func(s string) error {
						year, err := strconv.Atoi(s)
						if err != nil || year < 2000 {
							return fmt.Errorf("please enter a valid year")
						}
						return nil
					}),
			),
		)

		if err := monthForm.Run(); err != nil {
			return fmt.Errorf("export cancelled: %w", err)
		}

		// Pad month to 2 digits
		month, _ := strconv.Atoi(monthStr)
		yearMonth := fmt.Sprintf("%s-%02d", yearStr, month)

		query := `
			SELECT id, invoice_num, company_id, amount, currency, description, status, issued_date, due_date, pdf_path
			FROM invoices
			WHERE strftime('%Y-%m', issued_date) = ?
			ORDER BY issued_date DESC
		`
		rows, err := db.DB.Query(query, yearMonth)
		if err != nil {
			return fmt.Errorf("failed to fetch invoices: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var inv models.Invoice
			if err := rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.CompanyID, &inv.Amount, &inv.Currency,
				&inv.Description, &inv.Status, &inv.IssuedDate, &inv.DueDate, &inv.PDFPath); err != nil {
				return fmt.Errorf("failed to scan invoice: %w", err)
			}
			invoices = append(invoices, inv)
		}

	case "all_pending":
		// Get all pending invoices
		query := `
			SELECT id, invoice_num, company_id, amount, currency, description, status, issued_date, due_date, pdf_path
			FROM invoices
			WHERE status = ?
			ORDER BY issued_date DESC
		`
		rows, err := db.DB.Query(query, models.StatusPending)
		if err != nil {
			return fmt.Errorf("failed to fetch invoices: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var inv models.Invoice
			if err := rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.CompanyID, &inv.Amount, &inv.Currency,
				&inv.Description, &inv.Status, &inv.IssuedDate, &inv.DueDate, &inv.PDFPath); err != nil {
				return fmt.Errorf("failed to scan invoice: %w", err)
			}
			invoices = append(invoices, inv)
		}
	}

	if len(invoices) == 0 {
		fmt.Println("No invoices found matching the criteria.")
		return nil
	}

	fmt.Printf("\nFound %d invoice(s) to export:\n", len(invoices))
	for _, inv := range invoices {
		fmt.Printf("  - %s (%s) - %.2f %s\n", inv.InvoiceNum, inv.IssuedDate.Format("2006-01-02"), inv.Amount, inv.Currency)
	}
	fmt.Println()

	// Confirm export
	var shouldProceed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Proceed with email export?").
				Value(&shouldProceed),
		),
	)

	if err := confirmForm.Run(); err != nil || !shouldProceed {
		fmt.Println("Export cancelled.")
		return nil
	}

	// Export each invoice
	for i, inv := range invoices {
		fmt.Printf("\n[%d/%d] Processing %s...\n", i+1, len(invoices), inv.InvoiceNum)

		// Convert invoice ID to string for runInvoiceEmail
		idStr := strconv.Itoa(int(inv.ID))
		if err := runInvoiceEmail(cmd, []string{idStr}); err != nil {
			fmt.Printf("  ❌ Failed to export %s: %v\n", inv.InvoiceNum, err)
			continue
		}
	}

	fmt.Println("\n✓ Batch export completed!")
	return nil
}

// runInvoiceSimple handles the simple case: ung invoice <client-name>
func runInvoiceSimple(cmd *cobra.Command, args []string) error {
	// If no arguments and no subcommand, show help
	if len(args) == 0 {
		return cmd.Help()
	}

	// Get client name from argument
	clientName := args[0]

	// Find client by name
	var clientID uint
	err := db.DB.QueryRow(`
		SELECT id FROM clients
		WHERE LOWER(name) LIKE LOWER(?)
		LIMIT 1
	`, "%"+clientName+"%").Scan(&clientID)

	if err != nil {
		return fmt.Errorf("client '%s' not found. Create client first with: ung client create", clientName)
	}

	fmt.Printf("📊 Generating invoice from tracked time for %s...\n\n", clientName)

	// Get unbilled time sessions for this client
	groups, err := getUnbilledTimeSessionsForClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to get unbilled sessions: %w", err)
	}

	if len(groups) == 0 {
		return fmt.Errorf("no unbilled time found for %s. Track time first with: ung track log --client %s --hours <hours>", clientName, clientName)
	}

	// Auto-select the first (and likely only) group for this client
	selectedGroup := groups[0]

	// Show what will be invoiced
	fmt.Printf("Sessions to invoice:\n")
	for _, session := range selectedGroup.Sessions {
		hours := 0.0
		if session.Hours != nil {
			hours = *session.Hours
		}
		fmt.Printf("  • %s - %.2f hours", session.StartTime.Format("2006-01-02"), hours)
		if session.ProjectName != "" {
			fmt.Printf(" - %s", session.ProjectName)
		}
		fmt.Println()
	}
	fmt.Printf("\nTotal: %.2f hours", selectedGroup.TotalHours)
	if selectedGroup.HourlyRate != nil {
		fmt.Printf(" @ %.2f %s/hr = %.2f %s", *selectedGroup.HourlyRate, selectedGroup.Currency, selectedGroup.TotalHours*(*selectedGroup.HourlyRate), selectedGroup.Currency)
	}
	fmt.Println("\n")

	// Get company ID (should be 1 typically)
	var companyID uint
	err = db.DB.QueryRow("SELECT id FROM companies LIMIT 1").Scan(&companyID)
	if err != nil {
		return fmt.Errorf("no company found. Create one first with: ung company create")
	}

	// Calculate amount
	amount := 0.0
	currency := selectedGroup.Currency
	if selectedGroup.HourlyRate != nil {
		amount = selectedGroup.TotalHours * (*selectedGroup.HourlyRate)
	}

	if amount == 0 {
		return fmt.Errorf("cannot calculate invoice amount (no hourly rate set). Please set a rate on the contract.")
	}

	// Generate invoice number
	var fullClientName string
	db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", clientID).Scan(&fullClientName)
	invoiceNum, err := idgen.GenerateInvoiceNumber(db.GormDB, fullClientName, time.Now())
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice
	dueDate := time.Now().AddDate(0, 0, 15) // 15 days from now
	issuedDate := time.Now()

	result, err := db.DB.Exec(`
		INSERT INTO invoices (invoice_num, company_id, amount, currency, description, status, issued_date, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, invoiceNum, companyID, amount, currency, "Time-based services", models.StatusPending, issuedDate, dueDate)

	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	invoiceID, _ := result.LastInsertId()

	// Link to client
	db.DB.Exec("INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)", invoiceID, clientID)

	// Create line items and mark sessions as invoiced
	for _, session := range selectedGroup.Sessions {
		hours := 0.0
		if session.Hours != nil {
			hours = *session.Hours
		}

		rate := 0.0
		if selectedGroup.HourlyRate != nil {
			rate = *selectedGroup.HourlyRate
		}

		itemAmount := hours * rate
		itemName := session.ProjectName
		if itemName == "" {
			itemName = "Development work"
		}
		itemName = fmt.Sprintf("%s - %s", session.StartTime.Format("Jan 2"), itemName)

		db.DB.Exec(`
			INSERT INTO invoice_line_items (invoice_id, item_name, description, quantity, rate, amount)
			VALUES (?, ?, ?, ?, ?, ?)
		`, invoiceID, itemName, session.Notes, hours, rate, itemAmount)

		// Mark as invoiced
		newNotes := session.Notes
		if newNotes != "" {
			newNotes += " "
		}
		newNotes += fmt.Sprintf("[Invoiced: %s]", invoiceNum)
		db.DB.Exec("UPDATE tracking_sessions SET notes = ? WHERE id = ?", newNotes, session.ID)
	}

	fmt.Printf("✓ Invoice created: %s\n", invoiceNum)
	fmt.Printf("  Amount: %.2f %s\n", amount, currency)
	fmt.Printf("  Due: %s\n\n", dueDate.Format("2006-01-02"))

	// Auto-generate PDF
	fmt.Println("Generating PDF...")
	err = runInvoicePDF(cmd, []string{fmt.Sprintf("%d", invoiceID)})
	if err != nil {
		fmt.Printf("Warning: Could not generate PDF: %v\n", err)
	}

	return nil
}
