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
	Use:   "invoice",
	Short: "Manage invoices or generate from time for a client",
	Long: `Create, list, and manage invoices.

Examples:
  ung invoice -c skeep --pdf           Generate invoice from time + PDF
  ung invoice --client skeep --email   Generate invoice + email (auto-generates PDF)
  ung invoice --id 5 --pdf             Generate PDF for existing invoice
  ung invoice --id 5 --email           Email existing invoice`,
	RunE: runInvoiceMain,
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

var invoiceGenerateAllCmd = &cobra.Command{
	Use:   "generate-all",
	Short: "Generate invoices for all clients with unbilled time",
	Long: `Generate invoices for all clients with unbilled time tracking entries.

This command will:
1. Find all clients with unbilled time
2. Create invoices for each client based on their contract terms
3. Optionally generate PDFs and/or send emails

Examples:
  ung invoice generate-all                    Generate invoices only
  ung invoice generate-all --pdf              Generate invoices + PDFs
  ung invoice generate-all --email            Generate invoices + PDFs + emails
  ung invoice generate-all --email --email-app apple   Use Apple Mail`,
	RunE: runInvoiceGenerateAll,
}

var invoiceSendAllCmd = &cobra.Command{
	Use:   "send-all",
	Short: "Send emails for all pending invoices",
	Long: `Send emails for all pending invoices that have PDFs generated.

This command will:
1. Find all pending invoices
2. Generate PDFs if not already generated
3. Open email client for each invoice

Examples:
  ung invoice send-all                        Send all pending invoices
  ung invoice send-all --email-app outlook    Use Outlook`,
	RunE: runInvoiceSendAll,
}

var invoiceMarkCmd = &cobra.Command{
	Use:   "mark <id>",
	Short: "Mark an invoice with a new status",
	Long: `Update the status of an invoice.

Available statuses:
  - pending: Invoice not yet sent
  - sent: Invoice has been sent to client
  - paid: Payment received
  - overdue: Past due date

Examples:
  ung invoice mark 5 --status paid      Mark invoice #5 as paid
  ung invoice mark 3 --status sent      Mark invoice #3 as sent
  ung invoice mark 7 --status overdue   Mark invoice #7 as overdue`,
	Args: cobra.ExactArgs(1),
	RunE: runInvoiceMark,
}

var invoiceMarkStatus string


var (
	// Flags for invoice new command
	invoiceCompanyID   int
	invoiceClientID    int
	invoiceClientName  string
	invoiceAmount      float64
	invoiceCurrency    string
	invoiceDescription string
	invoiceDueDate     string

	// Flags for main invoice command
	invoiceFlagClient   string // --client, -c
	invoiceFlagID       int    // --id
	invoiceFlagPDF      bool   // --pdf
	invoiceFlagEmail    bool   // --email
	invoiceFlagEmailApp string // --email-app
	invoiceFlagBatch    bool   // --batch
)

func init() {
	// Add subcommands
	invoiceCmd.AddCommand(invoiceNewCmd)
	invoiceCmd.AddCommand(invoiceListCmd)
	invoiceCmd.AddCommand(invoiceGenerateAllCmd)
	invoiceCmd.AddCommand(invoiceSendAllCmd)
	invoiceCmd.AddCommand(invoiceMarkCmd)

	// Mark command flags
	invoiceMarkCmd.Flags().StringVar(&invoiceMarkStatus, "status", "", "New status (pending, sent, paid, overdue)")
	invoiceMarkCmd.MarkFlagRequired("status")

	// Main invoice command flags
	invoiceCmd.Flags().StringVarP(&invoiceFlagClient, "client", "c", "", "Client name (generates invoice from tracked time)")
	invoiceCmd.Flags().IntVar(&invoiceFlagID, "id", 0, "Existing invoice ID")
	invoiceCmd.Flags().BoolVar(&invoiceFlagPDF, "pdf", false, "Generate PDF")
	invoiceCmd.Flags().BoolVar(&invoiceFlagEmail, "email", false, "Send email (auto-generates PDF)")
	invoiceCmd.Flags().StringVar(&invoiceFlagEmailApp, "email-app", "", "Email client (apple, outlook, gmail)")
	invoiceCmd.Flags().BoolVar(&invoiceFlagBatch, "batch", false, "Batch operation for multiple invoices")

	// Generate-all command flags
	invoiceGenerateAllCmd.Flags().BoolVar(&invoiceFlagPDF, "pdf", false, "Generate PDF for each invoice")
	invoiceGenerateAllCmd.Flags().BoolVar(&invoiceFlagEmail, "email", false, "Send email for each invoice (auto-generates PDF)")
	invoiceGenerateAllCmd.Flags().StringVar(&invoiceFlagEmailApp, "email-app", "", "Email client (apple, outlook, gmail)")

	// Send-all command flags
	invoiceSendAllCmd.Flags().StringVar(&invoiceFlagEmailApp, "email-app", "", "Email client (apple, outlook, gmail)")

	// New invoice flags
	invoiceNewCmd.Flags().IntVar(&invoiceCompanyID, "company", 0, "Company ID")
	invoiceNewCmd.Flags().IntVar(&invoiceClientID, "client-id", 0, "Client ID")
	invoiceNewCmd.Flags().StringVar(&invoiceClientName, "client-name", "", "Client name (partial match, min 3 chars)")
	invoiceNewCmd.Flags().Float64Var(&invoiceAmount, "price", 0, "Invoice amount (required)")
	invoiceNewCmd.Flags().StringVar(&invoiceCurrency, "currency", "USD", "Currency code")
	invoiceNewCmd.Flags().StringVar(&invoiceDescription, "description", "", "Invoice description")
	invoiceNewCmd.Flags().StringVar(&invoiceDueDate, "due", "", "Due date (YYYY-MM-DD)")
	invoiceNewCmd.MarkFlagRequired("price")
}

func runInvoiceNew(cmd *cobra.Command, args []string) error {
	var clientName string
	var resolvedClientID int

	// Resolve client: by name (preferred) or by ID
	if invoiceClientName != "" {
		// Use client name search
		clientID, fullName, err := FindClientByName(invoiceClientName)
		if err != nil {
			return fmt.Errorf("client lookup failed: %w", err)
		}
		resolvedClientID = int(clientID)
		clientName = fullName
	} else if invoiceClientID > 0 {
		// Use client ID directly
		resolvedClientID = invoiceClientID
		err := db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", invoiceClientID).Scan(&clientName)
		if err != nil {
			return fmt.Errorf("client not found: %w", err)
		}
	} else {
		// Interactive mode - show all clients and let user select
		clients, err := getClients()
		if err != nil {
			return fmt.Errorf("failed to get clients: %w", err)
		}
		if len(clients) == 0 {
			return fmt.Errorf("no clients found. Create one first with: ung client add")
		}

		clientOptions := make([]huh.Option[int], len(clients))
		for i, c := range clients {
			clientOptions[i] = huh.NewOption(fmt.Sprintf("%s (%s)", c.Name, c.Email), int(c.ID))
		}

		var selectedClientID int
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Client").
					Options(clientOptions...).
					Value(&selectedClientID),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}

		resolvedClientID = selectedClientID
		err = db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", selectedClientID).Scan(&clientName)
		if err != nil {
			return fmt.Errorf("client not found: %w", err)
		}
	}

	// Get company ID - use provided or default to first company
	if invoiceCompanyID == 0 {
		err := db.DB.QueryRow("SELECT id FROM companies LIMIT 1").Scan(&invoiceCompanyID)
		if err != nil {
			return fmt.Errorf("no company found. Create one first with: ung company add")
		}
	}

	// Use end of current month for issued date
	now := time.Now()
	issuedDate := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())

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
		invoiceID, resolvedClientID,
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


// runInvoiceMain handles the main invoice command with flags
func runInvoiceMain(cmd *cobra.Command, args []string) error {
	// If --batch flag is set, handle batch operations
	if invoiceFlagBatch {
		return runBatchInvoice(cmd)
	}

	// If --client is provided, generate invoice from tracked time
	if invoiceFlagClient != "" {
		invoiceID, err := generateInvoiceFromTime(invoiceFlagClient)
		if err != nil {
			return err
		}
		// Use the newly created invoice ID for PDF/email
		invoiceFlagID = int(invoiceID)
	}

	// If no --client and no --id, show help
	if invoiceFlagClient == "" && invoiceFlagID == 0 {
		return cmd.Help()
	}

	// Handle --pdf flag (or auto-generate for --email)
	if invoiceFlagPDF || invoiceFlagEmail {
		if err := generateInvoicePDFByID(invoiceFlagID); err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
	}

	// Handle --email flag
	if invoiceFlagEmail {
		if err := emailInvoiceByID(invoiceFlagID, invoiceFlagEmailApp); err != nil {
			return fmt.Errorf("failed to email invoice: %w", err)
		}
	}

	return nil
}

// generateInvoiceFromTime creates an invoice from tracked time for a client
func generateInvoiceFromTime(clientName string) (int64, error) {
	// Find client (handles multiple matches with interactive selection)
	clientID, fullClientName, err := FindClientByName(clientName)
	if err != nil {
		return 0, fmt.Errorf("%w. Create client first with: ung client create", err)
	}

	fmt.Printf("📊 Generating invoice from tracked time for %s...\n\n", fullClientName)

	// Get unbilled time sessions for this client
	groups, err := getUnbilledTimeSessionsForClient(clientID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unbilled sessions: %w", err)
	}

	if len(groups) == 0 {
		return 0, fmt.Errorf("no unbilled time found for %s. Track time first with: ung track log --client %s --hours <hours>", fullClientName, fullClientName)
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
	if selectedGroup.ContractType == "hourly" && selectedGroup.HourlyRate != nil {
		fmt.Printf(" @ %.2f %s/hr = %.2f %s", *selectedGroup.HourlyRate, selectedGroup.Currency, selectedGroup.TotalHours*(*selectedGroup.HourlyRate), selectedGroup.Currency)
	} else if selectedGroup.ContractType == "fixed_price" && selectedGroup.FixedPrice != nil {
		fmt.Printf(" [Fixed price: %.2f %s]", *selectedGroup.FixedPrice, selectedGroup.Currency)
	}
	fmt.Println("\n")

	// Get company ID (should be 1 typically)
	var companyID uint
	err = db.DB.QueryRow("SELECT id FROM companies LIMIT 1").Scan(&companyID)
	if err != nil {
		return 0, fmt.Errorf("no company found. Create one first with: ung company create")
	}

	// Calculate amount based on contract type
	amount := 0.0
	currency := selectedGroup.Currency
	if selectedGroup.ContractType == "fixed_price" && selectedGroup.FixedPrice != nil {
		amount = *selectedGroup.FixedPrice
	} else if selectedGroup.ContractType == "hourly" && selectedGroup.HourlyRate != nil {
		amount = selectedGroup.TotalHours * (*selectedGroup.HourlyRate)
	}

	if amount == 0 {
		return 0, fmt.Errorf("cannot calculate invoice amount (no rate set). Please set a rate on the contract.")
	}

	// Generate invoice number
	invoiceNum, err := idgen.GenerateInvoiceNumber(db.GormDB, fullClientName, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice - use end of current month for issued date
	now := time.Now()
	issuedDate := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())
	dueDate := issuedDate.AddDate(0, 1, 0) // 1 month from issued date

	result, err := db.DB.Exec(`
		INSERT INTO invoices (invoice_num, company_id, amount, currency, description, status, issued_date, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, invoiceNum, companyID, amount, currency, "Time-based services", models.StatusPending, issuedDate, dueDate)

	if err != nil {
		return 0, fmt.Errorf("failed to create invoice: %w", err)
	}

	invoiceID, _ := result.LastInsertId()

	// Link to client
	db.DB.Exec("INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)", invoiceID, clientID)

	// Create line items based on contract type
	if selectedGroup.ContractType == "fixed_price" {
		// For fixed price contracts, create a single line item
		monthName := time.Now().Format("January 2006")
		itemName := fmt.Sprintf("Software services in %s", monthName)
		itemDescription := fmt.Sprintf("Fixed price contract work (%.2f hours tracked)", selectedGroup.TotalHours)

		db.DB.Exec(`
			INSERT INTO invoice_line_items (invoice_id, item_name, description, quantity, rate, amount)
			VALUES (?, ?, ?, ?, ?, ?)
		`, invoiceID, itemName, itemDescription, 1, amount, amount)

		// Mark all sessions as invoiced
		for _, session := range selectedGroup.Sessions {
			newNotes := session.Notes
			if newNotes != "" {
				newNotes += " "
			}
			newNotes += fmt.Sprintf("[Invoiced: %s]", invoiceNum)
			db.DB.Exec("UPDATE tracking_sessions SET notes = ? WHERE id = ?", newNotes, session.ID)
		}
	} else {
		// For hourly contracts, create line items for each session
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
	}

	fmt.Printf("✓ Invoice created: %s\n", invoiceNum)
	fmt.Printf("  Amount: %.2f %s\n", amount, currency)
	fmt.Printf("  Due: %s\n\n", dueDate.Format("2006-01-02"))

	return invoiceID, nil
}

// generateInvoicePDFByID generates PDF for an invoice by ID
func generateInvoicePDFByID(invoiceID int) error {
	// Fetch invoice details
	var inv models.Invoice
	var company models.Company
	var client models.Client

	// Get invoice
	err := db.DB.QueryRow(`
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

// emailInvoiceByID sends an invoice email by ID
func emailInvoiceByID(invoiceID int, emailApp string) error {
	// Fetch invoice details
	var inv models.Invoice
	var company models.Company

	// Get invoice
	err := db.DB.QueryRow(`
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

	// Ensure PDF exists
	pdfPath := inv.PDFPath
	if pdfPath == "" {
		return fmt.Errorf("PDF not generated. Run with --pdf flag first")
	}

	// Extract month and year from issued date
	month := inv.IssuedDate.Format("01")
	year := inv.IssuedDate.Format("2006")

	// Prepare email details
	subject := fmt.Sprintf("%s.%s %s", month, year, company.Name)
	body := fmt.Sprintf("Hi,\n\nHere is the invoice for %s %s.\n\nBest regards,\n%s",
		inv.IssuedDate.Format("January"), year, company.Name)

	// Determine email client
	emailClient := emailApp

	if emailClient == "" {
		// Interactive mode - prompt for selection
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
	} else {
		// Validate provided email client
		if emailClient != "apple" && emailClient != "outlook" && emailClient != "gmail" {
			return fmt.Errorf("invalid email client: %s (valid: apple, outlook, gmail)", emailClient)
		}
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

// runBatchInvoice handles batch invoice operations
func runBatchInvoice(cmd *cobra.Command) error {
	// Prompt for export option
	var exportOption string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What invoices would you like to process?").
				Options(
					huh.NewOption("Process for latest month", "latest"),
					huh.NewOption("Select specific month/year", "specific"),
					huh.NewOption("Process all pending invoices", "all_pending"),
				).
				Value(&exportOption),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	var invoices []models.Invoice

	switch exportOption {
	case "latest":
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
			return fmt.Errorf("cancelled: %w", err)
		}

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

	fmt.Printf("\nFound %d invoice(s):\n", len(invoices))
	for _, inv := range invoices {
		fmt.Printf("  - %s (%s) - %.2f %s\n", inv.InvoiceNum, inv.IssuedDate.Format("2006-01-02"), inv.Amount, inv.Currency)
	}
	fmt.Println()

	// Confirm
	var shouldProceed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Proceed?").
				Value(&shouldProceed),
		),
	)

	if err := confirmForm.Run(); err != nil || !shouldProceed {
		fmt.Println("Cancelled.")
		return nil
	}

	// Process each invoice
	for i, inv := range invoices {
		fmt.Printf("\n[%d/%d] Processing %s...\n", i+1, len(invoices), inv.InvoiceNum)

		// Generate PDF if --pdf or --email
		if invoiceFlagPDF || invoiceFlagEmail {
			if err := generateInvoicePDFByID(int(inv.ID)); err != nil {
				fmt.Printf("  ❌ Failed to generate PDF for %s: %v\n", inv.InvoiceNum, err)
				continue
			}
		}

		// Email if --email
		if invoiceFlagEmail {
			if err := emailInvoiceByID(int(inv.ID), invoiceFlagEmailApp); err != nil {
				fmt.Printf("  ❌ Failed to email %s: %v\n", inv.InvoiceNum, err)
				continue
			}
		}
	}

	fmt.Println("\n✓ Batch processing completed!")
	return nil
}

// runInvoiceGenerateAll generates invoices for all clients with unbilled time
func runInvoiceGenerateAll(cmd *cobra.Command, args []string) error {
	// Get all unbilled time sessions grouped by client
	groups, err := getUnbilledTimeSessions()
	if err != nil {
		return fmt.Errorf("failed to get unbilled sessions: %w", err)
	}

	if len(groups) == 0 {
		fmt.Println("No unbilled time found for any clients.")
		return nil
	}

	// Show summary
	fmt.Println("📊 Clients with unbilled time:\n")
	totalAmount := 0.0
	for i, group := range groups {
		amount := 0.0
		if group.ContractType == "fixed_price" && group.FixedPrice != nil {
			amount = *group.FixedPrice
		} else if group.HourlyRate != nil {
			amount = group.TotalHours * (*group.HourlyRate)
		}
		totalAmount += amount
		fmt.Printf("  %d. %s - %.2f hours", i+1, group.ClientName, group.TotalHours)
		if amount > 0 {
			fmt.Printf(" = %.2f %s", amount, group.Currency)
		}
		fmt.Println()
	}
	fmt.Printf("\nTotal: %.2f (across all currencies)\n", totalAmount)
	fmt.Printf("Will create %d invoice(s)\n\n", len(groups))

	// Confirm
	var shouldProceed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate invoices for all clients?").
				Value(&shouldProceed),
		),
	)

	if err := confirmForm.Run(); err != nil || !shouldProceed {
		fmt.Println("Cancelled.")
		return nil
	}

	// Generate invoice for each client
	generatedInvoices := []struct {
		ID         int64
		InvoiceNum string
		ClientName string
	}{}

	for i, group := range groups {
		fmt.Printf("\n[%d/%d] Generating invoice for %s...\n", i+1, len(groups), group.ClientName)

		// Use the existing generateInvoiceFromTime function
		invoiceID, err := generateInvoiceFromTime(group.ClientName)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Get invoice number
		var invoiceNum string
		db.DB.QueryRow("SELECT invoice_num FROM invoices WHERE id = ?", invoiceID).Scan(&invoiceNum)

		generatedInvoices = append(generatedInvoices, struct {
			ID         int64
			InvoiceNum string
			ClientName string
		}{invoiceID, invoiceNum, group.ClientName})

		fmt.Printf("  ✓ Created %s\n", invoiceNum)

		// Generate PDF if --pdf or --email
		if invoiceFlagPDF || invoiceFlagEmail {
			if err := generateInvoicePDFByID(int(invoiceID)); err != nil {
				fmt.Printf("  ❌ Failed to generate PDF: %v\n", err)
				continue
			}
			fmt.Printf("  ✓ PDF generated\n")
		}

		// Email if --email
		if invoiceFlagEmail {
			if err := emailInvoiceByID(int(invoiceID), invoiceFlagEmailApp); err != nil {
				fmt.Printf("  ❌ Failed to email: %v\n", err)
				continue
			}
			fmt.Printf("  ✓ Email prepared\n")
		}
	}

	fmt.Printf("\n✓ Generated %d invoice(s)!\n", len(generatedInvoices))
	if len(generatedInvoices) > 0 {
		fmt.Println("\nSummary:")
		for _, inv := range generatedInvoices {
			fmt.Printf("  • %s for %s\n", inv.InvoiceNum, inv.ClientName)
		}
	}

	return nil
}

// runInvoiceSendAll sends emails for all pending invoices
func runInvoiceSendAll(cmd *cobra.Command, args []string) error {
	// Get all pending invoices
	query := `
		SELECT i.id, i.invoice_num, i.amount, i.currency, i.status, i.pdf_path,
		       c.name as client_name
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
		WHERE i.status = ?
		ORDER BY i.issued_date DESC
	`

	rows, err := db.DB.Query(query, models.StatusPending)
	if err != nil {
		return fmt.Errorf("failed to fetch invoices: %w", err)
	}
	defer rows.Close()

	type pendingInvoice struct {
		ID         int64
		InvoiceNum string
		Amount     float64
		Currency   string
		Status     string
		PDFPath    *string
		ClientName string
	}

	var invoices []pendingInvoice
	for rows.Next() {
		var inv pendingInvoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.Amount, &inv.Currency,
			&inv.Status, &inv.PDFPath, &inv.ClientName); err != nil {
			return fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}

	if len(invoices) == 0 {
		fmt.Println("No pending invoices found.")
		return nil
	}

	fmt.Printf("📧 Found %d pending invoice(s):\n\n", len(invoices))
	for i, inv := range invoices {
		fmt.Printf("  %d. %s - %s - %.2f %s\n", i+1, inv.InvoiceNum, inv.ClientName, inv.Amount, inv.Currency)
	}
	fmt.Println()

	// Confirm
	var shouldProceed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Send emails for all pending invoices?").
				Value(&shouldProceed),
		),
	)

	if err := confirmForm.Run(); err != nil || !shouldProceed {
		fmt.Println("Cancelled.")
		return nil
	}

	// Process each invoice
	successCount := 0
	for i, inv := range invoices {
		fmt.Printf("\n[%d/%d] Processing %s (%s)...\n", i+1, len(invoices), inv.InvoiceNum, inv.ClientName)

		// Generate PDF if not exists
		if err := generateInvoicePDFByID(int(inv.ID)); err != nil {
			fmt.Printf("  ❌ Failed to generate PDF: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ PDF ready\n")

		// Send email
		if err := emailInvoiceByID(int(inv.ID), invoiceFlagEmailApp); err != nil {
			fmt.Printf("  ❌ Failed to email: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ Email prepared\n")
		successCount++
	}

	fmt.Printf("\n✓ Processed %d/%d invoice(s)!\n", successCount, len(invoices))
	return nil
}

// runInvoiceMark updates the status of an invoice
func runInvoiceMark(cmd *cobra.Command, args []string) error {
	// Parse invoice ID from args
	invoiceID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %s", args[0])
	}

	// Validate status
	validStatuses := map[string]models.InvoiceStatus{
		"pending": models.StatusPending,
		"sent":    models.StatusSent,
		"paid":    models.StatusPaid,
		"overdue": models.StatusOverdue,
	}

	newStatus, ok := validStatuses[invoiceMarkStatus]
	if !ok {
		return fmt.Errorf("invalid status: %s (valid: pending, sent, paid, overdue)", invoiceMarkStatus)
	}

	// Check if invoice exists
	var currentStatus string
	var invoiceNum string
	err = db.DB.QueryRow("SELECT invoice_num, status FROM invoices WHERE id = ?", invoiceID).Scan(&invoiceNum, &currentStatus)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Update status
	_, err = db.DB.Exec("UPDATE invoices SET status = ? WHERE id = ?", newStatus, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	fmt.Printf("✓ Invoice %s status updated: %s → %s\n", invoiceNum, currentStatus, newStatus)
	return nil
}
