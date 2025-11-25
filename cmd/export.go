package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data for accounting software",
	Long: `Export invoices, expenses, and time tracking data for accounting.

Supported formats:
  - CSV (universal, works with any spreadsheet/accounting software)
  - QuickBooks (IIF format)
  - JSON (for custom integrations)

Examples:
  ung export                       Interactive export wizard
  ung export --format csv          Export all to CSV
  ung export --invoices --year 2024    Export 2024 invoices
  ung export --expenses --quarter Q4   Export Q4 expenses`,
	RunE: runExport,
}

var (
	exportFormat   string
	exportInvoices bool
	exportExpenses bool
	exportTime     bool
	exportYear     int
	exportQuarter  string
	exportOutput   string
)

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "", "Export format: csv, quickbooks, json")
	exportCmd.Flags().BoolVar(&exportInvoices, "invoices", false, "Export invoices")
	exportCmd.Flags().BoolVar(&exportExpenses, "expenses", false, "Export expenses")
	exportCmd.Flags().BoolVar(&exportTime, "time", false, "Export time tracking")
	exportCmd.Flags().IntVarP(&exportYear, "year", "y", 0, "Filter by year")
	exportCmd.Flags().StringVarP(&exportQuarter, "quarter", "q", "", "Filter by quarter (Q1, Q2, Q3, Q4)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output directory")

	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	// Interactive mode if no flags specified
	if exportFormat == "" && !exportInvoices && !exportExpenses && !exportTime {
		return runExportInteractive()
	}

	// Default to CSV if format not specified
	if exportFormat == "" {
		exportFormat = "csv"
	}

	// Default to all if nothing selected
	if !exportInvoices && !exportExpenses && !exportTime {
		exportInvoices = true
		exportExpenses = true
		exportTime = true
	}

	// Default output directory
	if exportOutput == "" {
		exportOutput = filepath.Join(os.Getenv("HOME"), ".ung", "exports")
	}

	// Create output directory
	if err := os.MkdirAll(exportOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02")
	exported := []string{}

	if exportInvoices {
		path, err := exportInvoicesData(exportFormat, timestamp)
		if err != nil {
			return fmt.Errorf("failed to export invoices: %w", err)
		}
		exported = append(exported, path)
	}

	if exportExpenses {
		path, err := exportExpensesData(exportFormat, timestamp)
		if err != nil {
			return fmt.Errorf("failed to export expenses: %w", err)
		}
		exported = append(exported, path)
	}

	if exportTime {
		path, err := exportTimeData(exportFormat, timestamp)
		if err != nil {
			return fmt.Errorf("failed to export time tracking: %w", err)
		}
		exported = append(exported, path)
	}

	fmt.Println("\n✓ Export complete!")
	fmt.Println("Files created:")
	for _, p := range exported {
		fmt.Printf("  • %s\n", p)
	}

	return nil
}

func runExportInteractive() error {
	var format string
	var dataTypes []string
	var year int

	// Select format
	formatForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Export Format").
				Description("Choose format for your accounting software").
				Options(
					huh.NewOption("CSV (Universal)", "csv"),
					huh.NewOption("QuickBooks (IIF)", "quickbooks"),
					huh.NewOption("JSON (Custom)", "json"),
				).
				Value(&format),
		),
	)

	if err := formatForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	// Select data types
	dataForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("What to Export").
				Options(
					huh.NewOption("Invoices", "invoices"),
					huh.NewOption("Expenses", "expenses"),
					huh.NewOption("Time Tracking", "time"),
				).
				Value(&dataTypes),
		),
	)

	if err := dataForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	if len(dataTypes) == 0 {
		return fmt.Errorf("no data selected for export")
	}

	// Select year
	currentYear := time.Now().Year()
	yearForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Year").
				Options(
					huh.NewOption(fmt.Sprintf("%d (current)", currentYear), currentYear),
					huh.NewOption(fmt.Sprintf("%d", currentYear-1), currentYear-1),
					huh.NewOption(fmt.Sprintf("%d", currentYear-2), currentYear-2),
					huh.NewOption("All years", 0),
				).
				Value(&year),
		),
	)

	if err := yearForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	// Set variables and run
	exportFormat = format
	exportYear = year
	for _, dt := range dataTypes {
		switch dt {
		case "invoices":
			exportInvoices = true
		case "expenses":
			exportExpenses = true
		case "time":
			exportTime = true
		}
	}

	exportOutput = filepath.Join(os.Getenv("HOME"), ".ung", "exports")

	return runExport(nil, nil)
}

func exportInvoicesData(format, timestamp string) (string, error) {
	query := `
		SELECT i.id, i.invoice_num, c.name as client_name, i.amount, i.currency,
		       i.status, i.issued_date, i.due_date, i.description
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
	`
	args := []interface{}{}

	if exportYear > 0 {
		query += " WHERE strftime('%Y', i.issued_date) = ?"
		args = append(args, fmt.Sprintf("%d", exportYear))
	}

	query += " ORDER BY i.issued_date DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type invoiceRow struct {
		ID          int
		InvoiceNum  string
		ClientName  string
		Amount      float64
		Currency    string
		Status      string
		IssuedDate  time.Time
		DueDate     time.Time
		Description string
	}

	var invoices []invoiceRow
	for rows.Next() {
		var inv invoiceRow
		if err := rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.ClientName, &inv.Amount,
			&inv.Currency, &inv.Status, &inv.IssuedDate, &inv.DueDate, &inv.Description); err != nil {
			continue
		}
		invoices = append(invoices, inv)
	}

	filename := filepath.Join(exportOutput, fmt.Sprintf("invoices_%s.%s", timestamp, format))

	switch format {
	case "csv":
		return exportInvoicesCSV(filename, invoices)
	case "quickbooks":
		return exportInvoicesIIF(filename, invoices)
	case "json":
		return exportInvoicesJSON(filename, invoices)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func exportInvoicesCSV(filename string, invoices []interface{}) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	writer.Write([]string{
		"Invoice Number", "Client", "Amount", "Currency", "Status",
		"Issue Date", "Due Date", "Description",
	})

	// Query and write data
	query := `
		SELECT i.invoice_num, c.name, i.amount, i.currency, i.status,
		       i.issued_date, i.due_date, i.description
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
	`
	args := []interface{}{}
	if exportYear > 0 {
		query += " WHERE strftime('%Y', i.issued_date) = ?"
		args = append(args, fmt.Sprintf("%d", exportYear))
	}
	query += " ORDER BY i.issued_date DESC"

	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var num, client, currency, status, desc string
		var amount float64
		var issued, due time.Time

		rows.Scan(&num, &client, &amount, &currency, &status, &issued, &due, &desc)

		writer.Write([]string{
			num, client, fmt.Sprintf("%.2f", amount), currency, status,
			issued.Format("2006-01-02"), due.Format("2006-01-02"), desc,
		})
		count++
	}

	fmt.Printf("  ✓ Exported %d invoices\n", count)
	return filename, nil
}

func exportInvoicesIIF(filename string, invoices []interface{}) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// QuickBooks IIF format
	fmt.Fprintln(file, "!TRNS\tTRNSTYPE\tDATE\tACCNT\tNAME\tAMOUNT\tDOCNUM\tMEMO")
	fmt.Fprintln(file, "!SPL\tTRNSTYPE\tDATE\tACCNT\tAMOUNT\tMEMO")
	fmt.Fprintln(file, "!ENDTRNS")

	query := `
		SELECT i.invoice_num, c.name, i.amount, i.issued_date, i.description
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
		WHERE i.status = 'paid'
	`
	args := []interface{}{}
	if exportYear > 0 {
		query += " AND strftime('%Y', i.issued_date) = ?"
		args = append(args, fmt.Sprintf("%d", exportYear))
	}

	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var num, client, desc string
		var amount float64
		var issued time.Time

		rows.Scan(&num, &client, &amount, &issued, &desc)

		// TRNS line (main transaction)
		fmt.Fprintf(file, "TRNS\tINVOICE\t%s\tAccounts Receivable\t%s\t%.2f\t%s\t%s\n",
			issued.Format("01/02/2006"), client, amount, num, desc)
		// SPL line (split)
		fmt.Fprintf(file, "SPL\tINVOICE\t%s\tSales\t-%.2f\t%s\n",
			issued.Format("01/02/2006"), amount, desc)
		fmt.Fprintln(file, "ENDTRNS")
		count++
	}

	fmt.Printf("  ✓ Exported %d invoices (QuickBooks IIF)\n", count)
	return filename, nil
}

func exportInvoicesJSON(filename string, invoices []interface{}) (string, error) {
	query := `
		SELECT i.id, i.invoice_num, c.name, i.amount, i.currency, i.status,
		       i.issued_date, i.due_date, i.description
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
	`
	args := []interface{}{}
	if exportYear > 0 {
		query += " WHERE strftime('%Y', i.issued_date) = ?"
		args = append(args, fmt.Sprintf("%d", exportYear))
	}

	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	type jsonInvoice struct {
		ID          int       `json:"id"`
		InvoiceNum  string    `json:"invoice_number"`
		Client      string    `json:"client"`
		Amount      float64   `json:"amount"`
		Currency    string    `json:"currency"`
		Status      string    `json:"status"`
		IssuedDate  time.Time `json:"issued_date"`
		DueDate     time.Time `json:"due_date"`
		Description string    `json:"description"`
	}

	var data []jsonInvoice
	for rows.Next() {
		var inv jsonInvoice
		rows.Scan(&inv.ID, &inv.InvoiceNum, &inv.Client, &inv.Amount,
			&inv.Currency, &inv.Status, &inv.IssuedDate, &inv.DueDate, &inv.Description)
		data = append(data, inv)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(map[string]interface{}{
		"exported_at": time.Now(),
		"type":        "invoices",
		"count":       len(data),
		"data":        data,
	})

	fmt.Printf("  ✓ Exported %d invoices (JSON)\n", len(data))
	return filename, nil
}

func exportExpensesData(format, timestamp string) (string, error) {
	filename := filepath.Join(exportOutput, fmt.Sprintf("expenses_%s.%s", timestamp, format))

	query := `SELECT id, description, amount, currency, category, date, vendor, notes FROM expenses`
	args := []interface{}{}
	if exportYear > 0 {
		query += " WHERE strftime('%Y', date) = ?"
		args = append(args, fmt.Sprintf("%d", exportYear))
	}
	query += " ORDER BY date DESC"

	switch format {
	case "csv":
		return exportExpensesCSV(filename, query, args)
	case "json":
		return exportExpensesJSON(filename, query, args)
	default:
		// For QuickBooks, use CSV as fallback
		filename = filepath.Join(exportOutput, fmt.Sprintf("expenses_%s.csv", timestamp))
		return exportExpensesCSV(filename, query, args)
	}
}

func exportExpensesCSV(filename, query string, args []interface{}) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Date", "Description", "Category", "Amount", "Currency", "Vendor", "Notes"})

	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var desc, currency, category, vendor, notes string
		var amount float64
		var date time.Time

		rows.Scan(&id, &desc, &amount, &currency, &category, &date, &vendor, &notes)

		writer.Write([]string{
			date.Format("2006-01-02"), desc, category,
			fmt.Sprintf("%.2f", amount), currency, vendor, notes,
		})
		count++
	}

	fmt.Printf("  ✓ Exported %d expenses\n", count)
	return filename, nil
}

func exportExpensesJSON(filename, query string, args []interface{}) (string, error) {
	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	type jsonExpense struct {
		ID          int       `json:"id"`
		Description string    `json:"description"`
		Amount      float64   `json:"amount"`
		Currency    string    `json:"currency"`
		Category    string    `json:"category"`
		Date        time.Time `json:"date"`
		Vendor      string    `json:"vendor"`
		Notes       string    `json:"notes"`
	}

	var data []jsonExpense
	for rows.Next() {
		var exp jsonExpense
		rows.Scan(&exp.ID, &exp.Description, &exp.Amount, &exp.Currency,
			&exp.Category, &exp.Date, &exp.Vendor, &exp.Notes)
		data = append(data, exp)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(map[string]interface{}{
		"exported_at": time.Now(),
		"type":        "expenses",
		"count":       len(data),
		"data":        data,
	})

	fmt.Printf("  ✓ Exported %d expenses (JSON)\n", len(data))
	return filename, nil
}

func exportTimeData(format, timestamp string) (string, error) {
	filename := filepath.Join(exportOutput, fmt.Sprintf("time_tracking_%s.%s", timestamp, format))

	query := `
		SELECT t.id, c.name as client, t.project_name, t.start_time, t.end_time,
		       t.hours, t.billable, t.notes
		FROM tracking_sessions t
		LEFT JOIN clients c ON t.client_id = c.id
		WHERE t.deleted_at IS NULL
	`
	args := []interface{}{}
	if exportYear > 0 {
		query += " AND strftime('%Y', t.start_time) = ?"
		args = append(args, fmt.Sprintf("%d", exportYear))
	}
	query += " ORDER BY t.start_time DESC"

	switch format {
	case "csv":
		return exportTimeCSV(filename, query, args)
	case "json":
		return exportTimeJSON(filename, query, args)
	default:
		filename = filepath.Join(exportOutput, fmt.Sprintf("time_tracking_%s.csv", timestamp))
		return exportTimeCSV(filename, query, args)
	}
}

func exportTimeCSV(filename, query string, args []interface{}) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Date", "Client", "Project", "Hours", "Billable", "Notes"})

	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	count := 0
	totalHours := 0.0
	for rows.Next() {
		var id int
		var client, project, notes string
		var startTime time.Time
		var endTime *time.Time
		var hours *float64
		var billable bool

		rows.Scan(&id, &client, &project, &startTime, &endTime, &hours, &billable, &notes)

		h := 0.0
		if hours != nil {
			h = *hours
			totalHours += h
		}

		billableStr := "No"
		if billable {
			billableStr = "Yes"
		}

		writer.Write([]string{
			startTime.Format("2006-01-02"), client, project,
			fmt.Sprintf("%.2f", h), billableStr, notes,
		})
		count++
	}

	fmt.Printf("  ✓ Exported %d time entries (%.1f hours total)\n", count, totalHours)
	return filename, nil
}

func exportTimeJSON(filename, query string, args []interface{}) (string, error) {
	rows, _ := db.DB.Query(query, args...)
	defer rows.Close()

	type jsonTime struct {
		ID        int        `json:"id"`
		Client    string     `json:"client"`
		Project   string     `json:"project"`
		StartTime time.Time  `json:"start_time"`
		EndTime   *time.Time `json:"end_time"`
		Hours     float64    `json:"hours"`
		Billable  bool       `json:"billable"`
		Notes     string     `json:"notes"`
	}

	var data []jsonTime
	totalHours := 0.0
	for rows.Next() {
		var t jsonTime
		var hours *float64
		rows.Scan(&t.ID, &t.Client, &t.Project, &t.StartTime, &t.EndTime, &hours, &t.Billable, &t.Notes)
		if hours != nil {
			t.Hours = *hours
			totalHours += t.Hours
		}
		data = append(data, t)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(map[string]interface{}{
		"exported_at":  time.Now(),
		"type":         "time_tracking",
		"count":        len(data),
		"total_hours":  totalHours,
		"data":         data,
	})

	fmt.Printf("  ✓ Exported %d time entries (JSON)\n", len(data))
	return filename, nil
}

// Silence unused variable warning
var _ = models.StatusPaid
