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
	filename := filepath.Join(exportOutput, fmt.Sprintf("invoices_%s.%s", timestamp, format))

	switch format {
	case "csv":
		return exportInvoicesCSV(filename)
	case "quickbooks":
		return exportInvoicesIIF(filename)
	case "json":
		return exportInvoicesJSON(filename)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func exportInvoicesCSV(filename string) (string, error) {
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

	// Get invoices
	var invoices []models.Invoice
	query := db.GormDB.Order("issued_date DESC")
	if exportYear > 0 {
		startOfYear := time.Date(exportYear, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(exportYear+1, 1, 1, 0, 0, 0, 0, time.UTC)
		query = query.Where("issued_date >= ? AND issued_date < ?", startOfYear, endOfYear)
	}
	query.Find(&invoices)

	// Get clients for mapping
	var recipients []models.InvoiceRecipient
	db.GormDB.Find(&recipients)
	var clients []models.Client
	db.GormDB.Find(&clients)

	clientMap := make(map[uint]string)
	for _, c := range clients {
		clientMap[c.ID] = c.Name
	}
	invoiceClient := make(map[uint]string)
	for _, r := range recipients {
		invoiceClient[r.InvoiceID] = clientMap[r.ClientID]
	}

	count := 0
	for _, inv := range invoices {
		clientName := invoiceClient[inv.ID]
		writer.Write([]string{
			inv.InvoiceNum, clientName, fmt.Sprintf("%.2f", inv.Amount), inv.Currency, string(inv.Status),
			inv.IssuedDate.Format("2006-01-02"), inv.DueDate.Format("2006-01-02"), inv.Description,
		})
		count++
	}

	fmt.Printf("  ✓ Exported %d invoices\n", count)
	return filename, nil
}

func exportInvoicesIIF(filename string) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// QuickBooks IIF format
	fmt.Fprintln(file, "!TRNS\tTRNSTYPE\tDATE\tACCNT\tNAME\tAMOUNT\tDOCNUM\tMEMO")
	fmt.Fprintln(file, "!SPL\tTRNSTYPE\tDATE\tACCNT\tAMOUNT\tMEMO")
	fmt.Fprintln(file, "!ENDTRNS")

	// Get paid invoices
	var invoices []models.Invoice
	query := db.GormDB.Where("status = ?", models.StatusPaid)
	if exportYear > 0 {
		startOfYear := time.Date(exportYear, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(exportYear+1, 1, 1, 0, 0, 0, 0, time.UTC)
		query = query.Where("issued_date >= ? AND issued_date < ?", startOfYear, endOfYear)
	}
	query.Find(&invoices)

	// Get clients for mapping
	var recipients []models.InvoiceRecipient
	db.GormDB.Find(&recipients)
	var clients []models.Client
	db.GormDB.Find(&clients)

	clientMap := make(map[uint]string)
	for _, c := range clients {
		clientMap[c.ID] = c.Name
	}
	invoiceClient := make(map[uint]string)
	for _, r := range recipients {
		invoiceClient[r.InvoiceID] = clientMap[r.ClientID]
	}

	count := 0
	for _, inv := range invoices {
		clientName := invoiceClient[inv.ID]

		// TRNS line (main transaction)
		fmt.Fprintf(file, "TRNS\tINVOICE\t%s\tAccounts Receivable\t%s\t%.2f\t%s\t%s\n",
			inv.IssuedDate.Format("01/02/2006"), clientName, inv.Amount, inv.InvoiceNum, inv.Description)
		// SPL line (split)
		fmt.Fprintf(file, "SPL\tINVOICE\t%s\tSales\t-%.2f\t%s\n",
			inv.IssuedDate.Format("01/02/2006"), inv.Amount, inv.Description)
		fmt.Fprintln(file, "ENDTRNS")
		count++
	}

	fmt.Printf("  ✓ Exported %d invoices (QuickBooks IIF)\n", count)
	return filename, nil
}

func exportInvoicesJSON(filename string) (string, error) {
	// Get invoices
	var invoices []models.Invoice
	query := db.GormDB.Order("issued_date DESC")
	if exportYear > 0 {
		startOfYear := time.Date(exportYear, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(exportYear+1, 1, 1, 0, 0, 0, 0, time.UTC)
		query = query.Where("issued_date >= ? AND issued_date < ?", startOfYear, endOfYear)
	}
	query.Find(&invoices)

	// Get clients for mapping
	var recipients []models.InvoiceRecipient
	db.GormDB.Find(&recipients)
	var clients []models.Client
	db.GormDB.Find(&clients)

	clientMap := make(map[uint]string)
	for _, c := range clients {
		clientMap[c.ID] = c.Name
	}
	invoiceClient := make(map[uint]string)
	for _, r := range recipients {
		invoiceClient[r.InvoiceID] = clientMap[r.ClientID]
	}

	type jsonInvoice struct {
		ID          uint      `json:"id"`
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
	for _, inv := range invoices {
		data = append(data, jsonInvoice{
			ID:          inv.ID,
			InvoiceNum:  inv.InvoiceNum,
			Client:      invoiceClient[inv.ID],
			Amount:      inv.Amount,
			Currency:    inv.Currency,
			Status:      string(inv.Status),
			IssuedDate:  inv.IssuedDate,
			DueDate:     inv.DueDate,
			Description: inv.Description,
		})
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

	// Get expenses
	var expenses []models.Expense
	query := db.GormDB.Order("date DESC")
	if exportYear > 0 {
		startOfYear := time.Date(exportYear, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(exportYear+1, 1, 1, 0, 0, 0, 0, time.UTC)
		query = query.Where("date >= ? AND date < ?", startOfYear, endOfYear)
	}
	query.Find(&expenses)

	switch format {
	case "csv":
		return exportExpensesCSV(filename, expenses)
	case "json":
		return exportExpensesJSON(filename, expenses)
	default:
		// For QuickBooks, use CSV as fallback
		filename = filepath.Join(exportOutput, fmt.Sprintf("expenses_%s.csv", timestamp))
		return exportExpensesCSV(filename, expenses)
	}
}

func exportExpensesCSV(filename string, expenses []models.Expense) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Date", "Description", "Category", "Amount", "Currency", "Vendor", "Notes"})

	count := 0
	for _, exp := range expenses {
		writer.Write([]string{
			exp.Date.Format("2006-01-02"), exp.Description, string(exp.Category),
			fmt.Sprintf("%.2f", exp.Amount), exp.Currency, exp.Vendor, exp.Notes,
		})
		count++
	}

	fmt.Printf("  ✓ Exported %d expenses\n", count)
	return filename, nil
}

func exportExpensesJSON(filename string, expenses []models.Expense) (string, error) {
	type jsonExpense struct {
		ID          uint      `json:"id"`
		Description string    `json:"description"`
		Amount      float64   `json:"amount"`
		Currency    string    `json:"currency"`
		Category    string    `json:"category"`
		Date        time.Time `json:"date"`
		Vendor      string    `json:"vendor"`
		Notes       string    `json:"notes"`
	}

	var data []jsonExpense
	for _, exp := range expenses {
		data = append(data, jsonExpense{
			ID:          exp.ID,
			Description: exp.Description,
			Amount:      exp.Amount,
			Currency:    exp.Currency,
			Category:    string(exp.Category),
			Date:        exp.Date,
			Vendor:      exp.Vendor,
			Notes:       exp.Notes,
		})
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

	// Get tracking sessions with client preloaded
	var sessions []models.TrackingSession
	query := db.GormDB.Where("deleted_at IS NULL").Preload("Client").Order("start_time DESC")
	if exportYear > 0 {
		startOfYear := time.Date(exportYear, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(exportYear+1, 1, 1, 0, 0, 0, 0, time.UTC)
		query = query.Where("start_time >= ? AND start_time < ?", startOfYear, endOfYear)
	}
	query.Find(&sessions)

	switch format {
	case "csv":
		return exportTimeCSV(filename, sessions)
	case "json":
		return exportTimeJSON(filename, sessions)
	default:
		filename = filepath.Join(exportOutput, fmt.Sprintf("time_tracking_%s.csv", timestamp))
		return exportTimeCSV(filename, sessions)
	}
}

func exportTimeCSV(filename string, sessions []models.TrackingSession) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Date", "Client", "Project", "Hours", "Billable", "Notes"})

	count := 0
	totalHours := 0.0
	for _, s := range sessions {
		clientName := ""
		if s.Client != nil {
			clientName = s.Client.Name
		}

		h := 0.0
		if s.Hours != nil {
			h = *s.Hours
			totalHours += h
		}

		billableStr := "No"
		if s.Billable {
			billableStr = "Yes"
		}

		writer.Write([]string{
			s.StartTime.Format("2006-01-02"), clientName, s.ProjectName,
			fmt.Sprintf("%.2f", h), billableStr, s.Notes,
		})
		count++
	}

	fmt.Printf("  ✓ Exported %d time entries (%.1f hours total)\n", count, totalHours)
	return filename, nil
}

func exportTimeJSON(filename string, sessions []models.TrackingSession) (string, error) {
	type jsonTime struct {
		ID        uint       `json:"id"`
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
	for _, s := range sessions {
		clientName := ""
		if s.Client != nil {
			clientName = s.Client.Name
		}

		hours := 0.0
		if s.Hours != nil {
			hours = *s.Hours
			totalHours += hours
		}

		data = append(data, jsonTime{
			ID:        s.ID,
			Client:    clientName,
			Project:   s.ProjectName,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Hours:     hours,
			Billable:  s.Billable,
			Notes:     s.Notes,
		})
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
