package cmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/pkg/idgen"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var invoiceFromTimeCmd = &cobra.Command{
	Use:   "from-time",
	Short: "Generate invoice from tracked time",
	Long:  "Create an invoice based on unbilled time tracking sessions",
	RunE:  runInvoiceFromTime,
}

func init() {
	invoiceCmd.AddCommand(invoiceFromTimeCmd)
}

type timeSessionGroup struct {
	ClientID     uint
	ClientName   string
	ContractID   *uint
	ContractName string
	HourlyRate   *float64
	Currency     string
	TotalHours   float64
	Sessions     []models.TrackingSession
}

func runInvoiceFromTime(cmd *cobra.Command, args []string) error {
	fmt.Println("📊 Generating invoice from tracked time...")

	// Get unbilled time sessions grouped by client/contract
	groups, err := getUnbilledTimeSessions()
	if err != nil {
		return fmt.Errorf("failed to get unbilled sessions: %w", err)
	}

	if len(groups) == 0 {
		fmt.Println("No unbilled time sessions found.")
		return nil
	}

	// Let user select which group to invoice
	var selectedGroupIdx int
	groupOptions := make([]huh.Option[int], len(groups))
	for i, g := range groups {
		label := fmt.Sprintf("%s - %.2f hours", g.ClientName, g.TotalHours)
		if g.ContractName != "" {
			label = fmt.Sprintf("%s (%s) - %.2f hours", g.ClientName, g.ContractName, g.TotalHours)
		}
		if g.HourlyRate != nil {
			label += fmt.Sprintf(" @ %.2f %s/hr = %.2f %s", *g.HourlyRate, g.Currency, g.TotalHours*(*g.HourlyRate), g.Currency)
		}
		groupOptions[i] = huh.NewOption(label, i)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select client/contract to invoice").
				Options(groupOptions...).
				Value(&selectedGroupIdx),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	selectedGroup := groups[selectedGroupIdx]

	// Show sessions to be invoiced
	fmt.Printf("\n📝 Sessions to be invoiced for %s:\n\n", selectedGroup.ClientName)
	for _, session := range selectedGroup.Sessions {
		hours := 0.0
		if session.Hours != nil {
			hours = *session.Hours
		}
		fmt.Printf("  • %s - %.2f hours - %s\n",
			session.StartTime.Format("2006-01-02"),
			hours,
			session.ProjectName)
		if session.Notes != "" {
			fmt.Printf("    Notes: %s\n", session.Notes)
		}
	}
	fmt.Printf("\nTotal: %.2f hours\n", selectedGroup.TotalHours)

	// Confirm
	var shouldProceed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create invoice for these sessions?").
				Value(&shouldProceed),
		),
	)

	if err := confirmForm.Run(); err != nil || !shouldProceed {
		fmt.Println("Invoice creation cancelled.")
		return nil
	}

	// Get company
	var companyID uint
	err = db.DB.QueryRow("SELECT id FROM companies LIMIT 1").Scan(&companyID)
	if err != nil {
		return fmt.Errorf("no company found. Please create a company first with 'ung company add'")
	}

	// Calculate amount
	var amount float64
	currency := "USD"
	if selectedGroup.HourlyRate != nil {
		amount = selectedGroup.TotalHours * (*selectedGroup.HourlyRate)
		currency = selectedGroup.Currency
	} else {
		// Prompt for amount if no hourly rate
		var amountStr string
		amountForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("No hourly rate found. Enter invoice amount:").
					Placeholder("e.g., 1500.00").
					Value(&amountStr).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("amount is required")
						}
						return nil
					}),
			),
		)
		if err := amountForm.Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		fmt.Sscanf(amountStr, "%f", &amount)
	}

	// Generate invoice number
	issuedDate := time.Now()
	invoiceNum, err := idgen.GenerateInvoiceNumber(db.GormDB, selectedGroup.ClientName, issuedDate)
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice
	dueDate := time.Now().AddDate(0, 0, 30) // 30 days from now
	description := fmt.Sprintf("Time tracking: %s", time.Now().Format("January 2006"))

	query := `
		INSERT INTO invoices (invoice_num, company_id, amount, currency, description, status, issued_date, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.DB.Exec(query,
		invoiceNum,
		companyID,
		amount,
		currency,
		description,
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
		invoiceID, selectedGroup.ClientID,
	)
	if err != nil {
		return fmt.Errorf("failed to link invoice to client: %w", err)
	}

	// Create line items for each session
	for _, session := range selectedGroup.Sessions {
		hours := 0.0
		if session.Hours != nil {
			hours = *session.Hours
		}

		rate := 0.0
		if selectedGroup.HourlyRate != nil {
			rate = *selectedGroup.HourlyRate
		} else {
			rate = amount / selectedGroup.TotalHours // Distribute amount proportionally
		}

		itemAmount := hours * rate
		itemName := fmt.Sprintf("%s - %s", session.StartTime.Format("Jan 2"), session.ProjectName)
		itemDescription := session.Notes

		_, err = db.DB.Exec(`
			INSERT INTO invoice_line_items (invoice_id, item_name, description, quantity, rate, amount)
			VALUES (?, ?, ?, ?, ?, ?)
		`, invoiceID, itemName, itemDescription, hours, rate, itemAmount)
		if err != nil {
			return fmt.Errorf("failed to create line item: %w", err)
		}

		// Mark session as billed (we'll add a billed field or invoice_id later)
		// For now, we could delete or mark as non-billable
		// Let's update notes to indicate it's billed
		newNotes := session.Notes
		if newNotes != "" {
			newNotes += " "
		}
		newNotes += fmt.Sprintf("[Invoiced: %s]", invoiceNum)

		_, err = db.DB.Exec("UPDATE tracking_sessions SET notes = ? WHERE id = ?", newNotes, session.ID)
		if err != nil {
			// Non-fatal, just log
			fmt.Printf("Warning: Could not mark session %d as invoiced\n", session.ID)
		}
	}

	fmt.Printf("\n✓ Invoice created successfully!\n")
	fmt.Printf("  Invoice Number: %s\n", invoiceNum)
	fmt.Printf("  Invoice ID: %d\n", invoiceID)
	fmt.Printf("  Amount: %.2f %s\n", amount, currency)
	fmt.Printf("  Hours: %.2f\n", selectedGroup.TotalHours)
	fmt.Printf("  Due Date: %s\n\n", dueDate.Format("2006-01-02"))

	// Ask if user wants to generate PDF
	var shouldGeneratePDF bool
	pdfForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate PDF now?").
				Value(&shouldGeneratePDF),
		),
	)

	if err := pdfForm.Run(); err == nil && shouldGeneratePDF {
		return runInvoicePDF(cmd, []string{fmt.Sprintf("%d", invoiceID)})
	}

	return nil
}

func getUnbilledTimeSessions() ([]timeSessionGroup, error) {
	// Query all billable sessions that haven't been invoiced yet
	// We check if notes contain "[Invoiced:" to see if already billed
	query := `
		SELECT
			ts.id, ts.client_id, ts.contract_id, ts.project_name, ts.start_time,
			ts.end_time, ts.duration, ts.hours, ts.notes,
			c.name as client_name,
			COALESCE(ct.name, '') as contract_name,
			ct.hourly_rate,
			COALESCE(ct.currency, 'USD') as currency
		FROM tracking_sessions ts
		LEFT JOIN clients c ON ts.client_id = c.id
		LEFT JOIN contracts ct ON ts.contract_id = ct.id
		WHERE ts.billable = 1
		  AND ts.deleted_at IS NULL
		  AND (ts.notes NOT LIKE '%[Invoiced:%' OR ts.notes IS NULL)
		ORDER BY c.id, ct.id, ts.start_time
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Group sessions by client/contract
	groupsMap := make(map[string]*timeSessionGroup)

	for rows.Next() {
		var session models.TrackingSession
		var clientName, contractName, currency string
		var clientID uint
		var contractID, duration sql.NullInt64
		var hourlyRate sql.NullFloat64
		var hours sql.NullFloat64
		var endTime sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&session.ID,
			&clientID,
			&contractID,
			&session.ProjectName,
			&session.StartTime,
			&endTime,
			&duration,
			&hours,
			&notes,
			&clientName,
			&contractName,
			&hourlyRate,
			&currency,
		)
		if err != nil {
			return nil, err
		}

		// Populate session fields
		session.ClientID = &clientID
		if contractID.Valid {
			cid := uint(contractID.Int64)
			session.ContractID = &cid
		}
		if endTime.Valid {
			session.EndTime = &endTime.Time
		}
		if duration.Valid {
			d := int(duration.Int64)
			session.Duration = &d
		}
		if hours.Valid {
			h := hours.Float64
			session.Hours = &h
		}
		if notes.Valid {
			session.Notes = notes.String
		}

		// Create group key
		groupKey := fmt.Sprintf("%d-%v", clientID, contractID.Int64)

		if _, exists := groupsMap[groupKey]; !exists {
			var rate *float64
			if hourlyRate.Valid {
				r := hourlyRate.Float64
				rate = &r
			}

			var cid *uint
			if contractID.Valid {
				c := uint(contractID.Int64)
				cid = &c
			}

			groupsMap[groupKey] = &timeSessionGroup{
				ClientID:     clientID,
				ClientName:   clientName,
				ContractID:   cid,
				ContractName: contractName,
				HourlyRate:   rate,
				Currency:     currency,
				TotalHours:   0,
				Sessions:     []models.TrackingSession{},
			}
		}

		group := groupsMap[groupKey]
		group.Sessions = append(group.Sessions, session)
		if session.Hours != nil {
			group.TotalHours += *session.Hours
		}
	}

	// Convert map to slice
	var groups []timeSessionGroup
	for _, g := range groupsMap {
		if g.TotalHours > 0 {
			groups = append(groups, *g)
		}
	}

	return groups, nil
}

// getUnbilledTimeSessionsForClient gets unbilled sessions for a specific client
func getUnbilledTimeSessionsForClient(clientID uint) ([]timeSessionGroup, error) {
	groups, err := getUnbilledTimeSessions()
	if err != nil {
		return nil, err
	}

	// Filter to only this client
	var filtered []timeSessionGroup
	for _, g := range groups {
		if g.ClientID == clientID {
			filtered = append(filtered, g)
		}
	}

	return filtered, nil
}
