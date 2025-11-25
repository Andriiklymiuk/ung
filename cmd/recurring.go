package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/pkg/idgen"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var recurringCmd = &cobra.Command{
	Use:     "recurring",
	Aliases: []string{"rec"},
	Short:   "Manage recurring invoices",
	Long: `Set up automatic invoice generation for retainer clients.

Recurring invoices are templates that automatically generate invoices
on a schedule (weekly, monthly, quarterly, or yearly).

Examples:
  ung recurring add                    Create a new recurring invoice
  ung recurring ls                     List all recurring invoices
  ung recurring generate               Generate all due invoices now
  ung recurring pause 1                Pause recurring invoice #1
  ung recurring resume 1               Resume recurring invoice #1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var recurringAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new recurring invoice",
	RunE:  runRecurringAdd,
}

var recurringListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all recurring invoices",
	RunE:    runRecurringList,
}

var recurringGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate all due recurring invoices",
	Long: `Generate invoices for all active recurring invoices that are due.

This command checks all active recurring invoices and generates
actual invoices for any that have passed their next generation date.

Examples:
  ung recurring generate               Generate due invoices
  ung recurring generate --all         Generate for all (ignore dates)
  ung recurring generate --dry-run     Preview what would be generated`,
	RunE: runRecurringGenerate,
}

var recurringPauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause a recurring invoice",
	Args:  cobra.ExactArgs(1),
	RunE:  runRecurringPause,
}

var recurringResumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a paused recurring invoice",
	Args:  cobra.ExactArgs(1),
	RunE:  runRecurringResume,
}

var recurringDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a recurring invoice",
	Args:  cobra.ExactArgs(1),
	RunE:  runRecurringDelete,
}

var (
	recurringGenerateAll    bool
	recurringGenerateDryRun bool
)

func init() {
	recurringCmd.AddCommand(recurringAddCmd)
	recurringCmd.AddCommand(recurringListCmd)
	recurringCmd.AddCommand(recurringGenerateCmd)
	recurringCmd.AddCommand(recurringPauseCmd)
	recurringCmd.AddCommand(recurringResumeCmd)
	recurringCmd.AddCommand(recurringDeleteCmd)

	recurringGenerateCmd.Flags().BoolVar(&recurringGenerateAll, "all", false, "Generate for all recurring invoices (ignore dates)")
	recurringGenerateCmd.Flags().BoolVar(&recurringGenerateDryRun, "dry-run", false, "Preview what would be generated without creating invoices")
}

func runRecurringAdd(cmd *cobra.Command, args []string) error {
	// Get clients
	clients, err := getClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %w", err)
	}
	if len(clients) == 0 {
		return fmt.Errorf("no clients found. Create one first with: ung client add")
	}

	// Build client options
	clientOptions := make([]huh.Option[int], len(clients))
	for i, c := range clients {
		clientOptions[i] = huh.NewOption(fmt.Sprintf("%s (%s)", c.Name, c.Email), int(c.ID))
	}

	var selectedClientID int
	var amount float64
	var currency string
	var description string
	var frequency string
	var dayOfMonth int
	var autoPDF bool
	var autoSend bool

	// Form for basic info
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select Client").
				Description("Which client is this for?").
				Options(clientOptions...).
				Value(&selectedClientID),

			huh.NewInput().
				Title("Amount").
				Description("Invoice amount").
				Placeholder("e.g., 5000").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("amount is required")
					}
					_, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return fmt.Errorf("invalid amount")
					}
					return nil
				}).
				Value((*string)(&[]string{""}[0])).
				CharLimit(20),

			huh.NewSelect[string]().
				Title("Currency").
				Options(
					huh.NewOption("USD", "USD"),
					huh.NewOption("EUR", "EUR"),
					huh.NewOption("GBP", "GBP"),
					huh.NewOption("CHF", "CHF"),
					huh.NewOption("PLN", "PLN"),
				).
				Value(&currency),

			huh.NewInput().
				Title("Description").
				Description("What's this for?").
				Placeholder("e.g., Monthly retainer").
				Value(&description),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Frequency").
				Description("How often to generate?").
				Options(
					huh.NewOption("Monthly", "monthly"),
					huh.NewOption("Weekly", "weekly"),
					huh.NewOption("Bi-weekly (every 2 weeks)", "biweekly"),
					huh.NewOption("Quarterly", "quarterly"),
					huh.NewOption("Yearly", "yearly"),
				).
				Value(&frequency),

			huh.NewSelect[int]().
				Title("Day of Month").
				Description("Generate on which day?").
				Options(
					huh.NewOption("1st", 1),
					huh.NewOption("5th", 5),
					huh.NewOption("10th", 10),
					huh.NewOption("15th", 15),
					huh.NewOption("20th", 20),
					huh.NewOption("25th", 25),
					huh.NewOption("Last day (28th)", 28),
				).
				Value(&dayOfMonth),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Auto-generate PDF?").
				Description("Automatically create PDF when invoice is generated").
				Value(&autoPDF),

			huh.NewConfirm().
				Title("Auto-send email?").
				Description("Automatically open email client when invoice is generated").
				Value(&autoSend),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}

	// Get amount from a separate input (huh doesn't handle float well)
	var amountStr string
	amountForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter amount").
				Value(&amountStr).
				Validate(func(s string) error {
					_, err := strconv.ParseFloat(s, 64)
					return err
				}),
		),
	)
	if err := amountForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}
	amount, _ = strconv.ParseFloat(amountStr, 64)

	// Calculate next generation date
	nextDate := calculateNextGenerationDate(models.RecurringFrequency(frequency), dayOfMonth, 0)

	// Create recurring invoice
	recurring := models.RecurringInvoice{
		ClientID:           uint(selectedClientID),
		Amount:             amount,
		Currency:           currency,
		Description:        description,
		Frequency:          models.RecurringFrequency(frequency),
		DayOfMonth:         dayOfMonth,
		NextGenerationDate: nextDate,
		Active:             true,
		AutoPDF:            autoPDF,
		AutoSend:           autoSend,
	}

	result := db.GormDB.Create(&recurring)
	if result.Error != nil {
		return fmt.Errorf("failed to create recurring invoice: %w", result.Error)
	}

	// Get client name for display
	var clientName string
	db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", selectedClientID).Scan(&clientName)

	fmt.Printf("\n✓ Recurring invoice created!\n")
	fmt.Printf("  Client: %s\n", clientName)
	fmt.Printf("  Amount: %.2f %s\n", amount, currency)
	fmt.Printf("  Frequency: %s\n", frequency)
	fmt.Printf("  Next generation: %s\n", nextDate.Format("2006-01-02"))
	fmt.Printf("  Auto PDF: %v\n", autoPDF)
	fmt.Printf("  Auto email: %v\n", autoSend)

	return nil
}

func runRecurringList(cmd *cobra.Command, args []string) error {
	query := `
		SELECT r.id, c.name, r.amount, r.currency, r.frequency,
		       r.next_generation_date, r.active, r.generated_count,
		       r.description
		FROM recurring_invoices r
		JOIN clients c ON r.client_id = c.id
		ORDER BY r.next_generation_date ASC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query recurring invoices: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCLIENT\tAMOUNT\tFREQUENCY\tNEXT DATE\tSTATUS\tGENERATED\tDESCRIPTION")

	count := 0
	for rows.Next() {
		var id int
		var clientName, currency, frequency, description string
		var amount float64
		var nextDate time.Time
		var active bool
		var generatedCount int

		if err := rows.Scan(&id, &clientName, &amount, &currency, &frequency,
			&nextDate, &active, &generatedCount, &description); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		status := "✓ Active"
		if !active {
			status = "⏸ Paused"
		}

		// Truncate description if too long
		if len(description) > 20 {
			description = description[:17] + "..."
		}

		fmt.Fprintf(w, "%d\t%s\t%.2f %s\t%s\t%s\t%s\t%d\t%s\n",
			id,
			clientName,
			amount,
			currency,
			frequency,
			nextDate.Format("2006-01-02"),
			status,
			generatedCount,
			description,
		)
		count++
	}

	w.Flush()

	if count == 0 {
		fmt.Println("No recurring invoices found. Create one with: ung recurring add")
	}

	return nil
}

func runRecurringGenerate(cmd *cobra.Command, args []string) error {
	now := time.Now()

	// Get due recurring invoices
	query := `
		SELECT r.id, r.client_id, c.name, r.amount, r.currency, r.frequency,
		       r.day_of_month, r.next_generation_date, r.description,
		       r.auto_pdf, r.auto_send, r.email_app, r.generated_count
		FROM recurring_invoices r
		JOIN clients c ON r.client_id = c.id
		WHERE r.active = 1
	`

	if !recurringGenerateAll {
		query += " AND r.next_generation_date <= ?"
	}

	query += " ORDER BY r.next_generation_date ASC"

	var rows interface{ Close() error }
	var err error

	if recurringGenerateAll {
		rows, err = db.DB.Query(query)
	} else {
		rows, err = db.DB.Query(query, now)
	}
	if err != nil {
		return fmt.Errorf("failed to query recurring invoices: %w", err)
	}
	defer rows.Close()

	type dueInvoice struct {
		ID             int
		ClientID       int
		ClientName     string
		Amount         float64
		Currency       string
		Frequency      string
		DayOfMonth     int
		NextDate       time.Time
		Description    string
		AutoPDF        bool
		AutoSend       bool
		EmailApp       string
		GeneratedCount int
	}

	var dueInvoices []dueInvoice
	scanner := rows.(*interface{})
	_ = scanner // Type assertion placeholder

	// Re-query with proper scanning
	if recurringGenerateAll {
		rows2, _ := db.DB.Query(query)
		defer rows2.Close()
		for rows2.Next() {
			var inv dueInvoice
			rows2.Scan(&inv.ID, &inv.ClientID, &inv.ClientName, &inv.Amount, &inv.Currency,
				&inv.Frequency, &inv.DayOfMonth, &inv.NextDate, &inv.Description,
				&inv.AutoPDF, &inv.AutoSend, &inv.EmailApp, &inv.GeneratedCount)
			dueInvoices = append(dueInvoices, inv)
		}
	} else {
		rows2, _ := db.DB.Query(query, now)
		defer rows2.Close()
		for rows2.Next() {
			var inv dueInvoice
			rows2.Scan(&inv.ID, &inv.ClientID, &inv.ClientName, &inv.Amount, &inv.Currency,
				&inv.Frequency, &inv.DayOfMonth, &inv.NextDate, &inv.Description,
				&inv.AutoPDF, &inv.AutoSend, &inv.EmailApp, &inv.GeneratedCount)
			dueInvoices = append(dueInvoices, inv)
		}
	}

	if len(dueInvoices) == 0 {
		fmt.Println("No recurring invoices are due for generation.")
		return nil
	}

	fmt.Printf("📋 Found %d recurring invoice(s) to generate:\n\n", len(dueInvoices))
	for i, inv := range dueInvoices {
		fmt.Printf("  %d. %s - %.2f %s (%s)\n", i+1, inv.ClientName, inv.Amount, inv.Currency, inv.Frequency)
	}

	if recurringGenerateDryRun {
		fmt.Println("\n(Dry run - no invoices created)")
		return nil
	}

	// Confirm
	var shouldProceed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate these invoices?").
				Value(&shouldProceed),
		),
	)

	if err := confirmForm.Run(); err != nil || !shouldProceed {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get company ID
	var companyID uint
	if err := db.DB.QueryRow("SELECT id FROM companies LIMIT 1").Scan(&companyID); err != nil {
		return fmt.Errorf("no company found. Create one first with: ung company add")
	}

	// Generate invoices
	generated := 0
	for _, inv := range dueInvoices {
		fmt.Printf("\n[%d/%d] Generating for %s...\n", generated+1, len(dueInvoices), inv.ClientName)

		// Generate invoice number
		invoiceNum, err := idgen.GenerateInvoiceNumber(db.GormDB, inv.ClientName, now)
		if err != nil {
			fmt.Printf("  ❌ Failed to generate invoice number: %v\n", err)
			continue
		}

		// Calculate dates
		issuedDate := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()) // End of month
		dueDate := issuedDate.AddDate(0, 1, 0)                                             // 1 month from issued

		// Create invoice
		result, err := db.DB.Exec(`
			INSERT INTO invoices (invoice_num, company_id, amount, currency, description, status, issued_date, due_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, invoiceNum, companyID, inv.Amount, inv.Currency, inv.Description, models.StatusPending, issuedDate, dueDate)

		if err != nil {
			fmt.Printf("  ❌ Failed to create invoice: %v\n", err)
			continue
		}

		invoiceID, _ := result.LastInsertId()

		// Link to client
		db.DB.Exec("INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)", invoiceID, inv.ClientID)

		// Create line item
		db.DB.Exec(`
			INSERT INTO invoice_line_items (invoice_id, item_name, description, quantity, rate, amount)
			VALUES (?, ?, ?, 1, ?, ?)
		`, invoiceID, inv.Description, fmt.Sprintf("Recurring invoice - %s", inv.Frequency), inv.Amount, inv.Amount)

		fmt.Printf("  ✓ Created %s\n", invoiceNum)

		// Generate PDF if auto_pdf
		if inv.AutoPDF {
			if err := generateInvoicePDFByID(int(invoiceID)); err != nil {
				fmt.Printf("  ⚠ PDF generation failed: %v\n", err)
			} else {
				fmt.Printf("  ✓ PDF generated\n")
			}
		}

		// Send email if auto_send
		if inv.AutoSend {
			emailApp := inv.EmailApp
			if emailApp == "" {
				emailApp = "gmail"
			}
			if err := emailInvoiceByID(int(invoiceID), emailApp); err != nil {
				fmt.Printf("  ⚠ Email failed: %v\n", err)
			} else {
				fmt.Printf("  ✓ Email prepared\n")
			}
		}

		// Update recurring invoice
		nextDate := calculateNextGenerationDate(models.RecurringFrequency(inv.Frequency), inv.DayOfMonth, 0)
		db.DB.Exec(`
			UPDATE recurring_invoices
			SET last_generated_date = ?, last_invoice_id = ?, next_generation_date = ?,
			    generated_count = generated_count + 1, updated_at = ?
			WHERE id = ?
		`, now, invoiceID, nextDate, now, inv.ID)

		generated++
	}

	fmt.Printf("\n✓ Generated %d invoice(s)!\n", generated)
	return nil
}

func runRecurringPause(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %s", args[0])
	}

	result, err := db.DB.Exec("UPDATE recurring_invoices SET active = 0, updated_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to pause recurring invoice: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("recurring invoice #%d not found", id)
	}

	fmt.Printf("✓ Recurring invoice #%d paused\n", id)
	return nil
}

func runRecurringResume(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %s", args[0])
	}

	// Get current frequency and day to recalculate next date
	var frequency string
	var dayOfMonth int
	err = db.DB.QueryRow("SELECT frequency, day_of_month FROM recurring_invoices WHERE id = ?", id).Scan(&frequency, &dayOfMonth)
	if err != nil {
		return fmt.Errorf("recurring invoice #%d not found", id)
	}

	nextDate := calculateNextGenerationDate(models.RecurringFrequency(frequency), dayOfMonth, 0)

	_, err = db.DB.Exec(`
		UPDATE recurring_invoices
		SET active = 1, next_generation_date = ?, updated_at = ?
		WHERE id = ?
	`, nextDate, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to resume recurring invoice: %w", err)
	}

	fmt.Printf("✓ Recurring invoice #%d resumed (next generation: %s)\n", id, nextDate.Format("2006-01-02"))
	return nil
}

func runRecurringDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %s", args[0])
	}

	// Confirm deletion
	var shouldDelete bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete recurring invoice #%d?", id)).
				Description("This will not delete any invoices already generated").
				Value(&shouldDelete),
		),
	)

	if err := form.Run(); err != nil || !shouldDelete {
		fmt.Println("Cancelled.")
		return nil
	}

	result, err := db.DB.Exec("DELETE FROM recurring_invoices WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete recurring invoice: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("recurring invoice #%d not found", id)
	}

	fmt.Printf("✓ Recurring invoice #%d deleted\n", id)
	return nil
}

// calculateNextGenerationDate calculates the next date based on frequency
func calculateNextGenerationDate(frequency models.RecurringFrequency, dayOfMonth int, dayOfWeek int) time.Time {
	now := time.Now()

	switch frequency {
	case models.FrequencyWeekly:
		// Next week, same day
		daysUntil := (7 - int(now.Weekday()) + dayOfWeek) % 7
		if daysUntil == 0 {
			daysUntil = 7
		}
		return now.AddDate(0, 0, daysUntil)

	case models.FrequencyBiweekly:
		// Two weeks from now
		daysUntil := (7 - int(now.Weekday()) + dayOfWeek) % 7
		if daysUntil == 0 {
			daysUntil = 14
		} else {
			daysUntil += 7
		}
		return now.AddDate(0, 0, daysUntil)

	case models.FrequencyMonthly:
		// Next month on specified day
		nextMonth := now.AddDate(0, 1, 0)
		day := dayOfMonth
		if day > 28 {
			day = 28
		}
		return time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, now.Location())

	case models.FrequencyQuarterly:
		// 3 months from now
		nextQuarter := now.AddDate(0, 3, 0)
		day := dayOfMonth
		if day > 28 {
			day = 28
		}
		return time.Date(nextQuarter.Year(), nextQuarter.Month(), day, 0, 0, 0, 0, now.Location())

	case models.FrequencyYearly:
		// Same month next year
		nextYear := now.AddDate(1, 0, 0)
		day := dayOfMonth
		if day > 28 {
			day = 28
		}
		return time.Date(nextYear.Year(), nextYear.Month(), day, 0, 0, 0, 0, now.Location())

	default:
		return now.AddDate(0, 1, 0) // Default to monthly
	}
}
