package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from CSV files",
	Long: `Import data from CSV files into UNG.

Supported data types:
  - clients     Import client list
  - expenses    Import expense records
  - time        Import time tracking entries

CSV Format Requirements:
  Clients:   name,email,address,tax_id
  Expenses:  date,description,amount,category,vendor,currency
  Time:      date,client,project,hours,billable,notes

Examples:
  ung import                         Interactive import wizard
  ung import --file clients.csv --type clients
  ung import --file expenses.csv --type expenses`,
	RunE: runImport,
}

var (
	importFile     string
	importType     string
	importSkipRows int
	importDryRun   bool
)

func init() {
	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "CSV file to import")
	importCmd.Flags().StringVarP(&importType, "type", "t", "", "Data type: clients, expenses, time")
	importCmd.Flags().IntVar(&importSkipRows, "skip", 1, "Number of header rows to skip")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview import without saving")

	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	// Interactive mode if no file specified
	if importFile == "" {
		return runImportInteractive()
	}

	if importType == "" {
		return fmt.Errorf("please specify data type with --type (clients, expenses, time)")
	}

	return runImportFile(importFile, importType)
}

func runImportInteractive() error {
	var dataType string
	var filePath string

	// Select data type
	typeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What data do you want to import?").
				Options(
					huh.NewOption("Clients (name, email, address, tax_id)", "clients"),
					huh.NewOption("Expenses (date, description, amount, category)", "expenses"),
					huh.NewOption("Time Tracking (date, client, project, hours)", "time"),
				).
				Value(&dataType),
		),
	)

	if err := typeForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	// Get file path
	fileForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("CSV File Path").
				Description("Enter the path to your CSV file").
				Placeholder("/path/to/data.csv").
				Value(&filePath),
		),
	)

	if err := fileForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		filePath = strings.Replace(filePath, "~", os.Getenv("HOME"), 1)
	}

	importType = dataType
	return runImportFile(filePath, dataType)
}

func runImportFile(filePath, dataType string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Skip header rows
	for i := 0; i < importSkipRows; i++ {
		reader.Read()
	}

	switch dataType {
	case "clients":
		return importClients(reader)
	case "expenses":
		return importExpenses(reader)
	case "time":
		return importTimeEntries(reader)
	default:
		return fmt.Errorf("unknown data type: %s", dataType)
	}
}

func importClients(reader *csv.Reader) error {
	fmt.Println("\nImporting clients...")
	fmt.Println("Expected format: name,email,address,tax_id")
	fmt.Println()

	imported := 0
	skipped := 0
	errors := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors++
			continue
		}

		if len(record) < 2 {
			skipped++
			continue
		}

		client := models.Client{
			Name:  strings.TrimSpace(record[0]),
			Email: strings.TrimSpace(record[1]),
		}

		if len(record) > 2 {
			client.Address = strings.TrimSpace(record[2])
		}
		if len(record) > 3 {
			client.TaxID = strings.TrimSpace(record[3])
		}

		if client.Name == "" || client.Email == "" {
			skipped++
			continue
		}

		if importDryRun {
			fmt.Printf("  [Preview] %s <%s>\n", client.Name, client.Email)
			imported++
			continue
		}

		// Check if client exists
		var existing models.Client
		if err := db.DB.Where("email = ?", client.Email).First(&existing).Error; err == nil {
			fmt.Printf("  [Skip] %s (already exists)\n", client.Email)
			skipped++
			continue
		}

		if err := db.DB.Create(&client).Error; err != nil {
			fmt.Printf("  [Error] %s: %v\n", client.Name, err)
			errors++
			continue
		}

		fmt.Printf("  [OK] %s\n", client.Name)
		imported++
	}

	fmt.Println()
	if importDryRun {
		fmt.Println("Dry run completed - no data was saved")
	}
	fmt.Printf("✓ Imported: %d, Skipped: %d, Errors: %d\n", imported, skipped, errors)
	return nil
}

func importExpenses(reader *csv.Reader) error {
	fmt.Println("\nImporting expenses...")
	fmt.Println("Expected format: date,description,amount,category,vendor,currency")
	fmt.Println()

	imported := 0
	skipped := 0
	errors := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors++
			continue
		}

		if len(record) < 3 {
			skipped++
			continue
		}

		// Parse date (try common formats)
		date, err := parseDate(record[0])
		if err != nil {
			fmt.Printf("  [Skip] Invalid date: %s\n", record[0])
			skipped++
			continue
		}

		// Parse amount
		amount, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		if err != nil {
			fmt.Printf("  [Skip] Invalid amount: %s\n", record[2])
			skipped++
			continue
		}

		expense := models.Expense{
			Date:        date,
			Description: strings.TrimSpace(record[1]),
			Amount:      amount,
			Currency:    "USD",
			Category:    models.ExpenseCategoryOther,
		}

		// Optional fields
		if len(record) > 3 {
			expense.Category = parseCategory(record[3])
		}
		if len(record) > 4 {
			expense.Vendor = strings.TrimSpace(record[4])
		}
		if len(record) > 5 {
			expense.Currency = strings.TrimSpace(record[5])
		}

		if importDryRun {
			fmt.Printf("  [Preview] %s: %.2f %s - %s\n", date.Format("2006-01-02"), amount, expense.Currency, expense.Description)
			imported++
			continue
		}

		if err := db.DB.Create(&expense).Error; err != nil {
			fmt.Printf("  [Error] %s: %v\n", expense.Description, err)
			errors++
			continue
		}

		fmt.Printf("  [OK] %s: %.2f\n", expense.Description, amount)
		imported++
	}

	fmt.Println()
	if importDryRun {
		fmt.Println("Dry run completed - no data was saved")
	}
	fmt.Printf("✓ Imported: %d, Skipped: %d, Errors: %d\n", imported, skipped, errors)
	return nil
}

func importTimeEntries(reader *csv.Reader) error {
	fmt.Println("\nImporting time entries...")
	fmt.Println("Expected format: date,client,project,hours,billable,notes")
	fmt.Println()

	// Get client map for lookups
	var clients []models.Client
	db.DB.Find(&clients)
	clientMap := make(map[string]uint)
	for _, c := range clients {
		clientMap[strings.ToLower(c.Name)] = c.ID
	}

	imported := 0
	skipped := 0
	errors := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors++
			continue
		}

		if len(record) < 4 {
			skipped++
			continue
		}

		// Parse date
		date, err := parseDate(record[0])
		if err != nil {
			fmt.Printf("  [Skip] Invalid date: %s\n", record[0])
			skipped++
			continue
		}

		// Parse hours
		hours, err := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		if err != nil {
			fmt.Printf("  [Skip] Invalid hours: %s\n", record[3])
			skipped++
			continue
		}

		clientName := strings.TrimSpace(record[1])
		projectName := strings.TrimSpace(record[2])

		session := models.TrackingSession{
			StartTime:   date,
			ProjectName: projectName,
			Hours:       &hours,
			Billable:    true,
		}

		// Try to find client
		if clientID, ok := clientMap[strings.ToLower(clientName)]; ok {
			session.ClientID = &clientID
		}

		// Calculate end time and duration
		endTime := date.Add(time.Duration(hours * float64(time.Hour)))
		session.EndTime = &endTime
		duration := int(hours * 3600)
		session.Duration = &duration

		// Optional billable flag
		if len(record) > 4 {
			billable := strings.ToLower(strings.TrimSpace(record[4]))
			session.Billable = billable == "true" || billable == "yes" || billable == "1"
		}

		// Optional notes
		if len(record) > 5 {
			session.Notes = strings.TrimSpace(record[5])
		}

		if importDryRun {
			fmt.Printf("  [Preview] %s: %.2fh for %s - %s\n", date.Format("2006-01-02"), hours, clientName, projectName)
			imported++
			continue
		}

		if err := db.DB.Create(&session).Error; err != nil {
			fmt.Printf("  [Error] %v\n", err)
			errors++
			continue
		}

		fmt.Printf("  [OK] %.2fh - %s\n", hours, projectName)
		imported++
	}

	fmt.Println()
	if importDryRun {
		fmt.Println("Dry run completed - no data was saved")
	}
	fmt.Printf("✓ Imported: %d, Skipped: %d, Errors: %d\n", imported, skipped, errors)
	return nil
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006/01/02",
		"Jan 2, 2006",
		"January 2, 2006",
		"02-Jan-2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func parseCategory(s string) models.ExpenseCategory {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.Contains(s, "software") || strings.Contains(s, "subscription"):
		return models.ExpenseCategorySoftware
	case strings.Contains(s, "hardware") || strings.Contains(s, "equipment"):
		return models.ExpenseCategoryHardware
	case strings.Contains(s, "travel"):
		return models.ExpenseCategoryTravel
	case strings.Contains(s, "meal") || strings.Contains(s, "food"):
		return models.ExpenseCategoryMeals
	case strings.Contains(s, "office") || strings.Contains(s, "supplies"):
		return models.ExpenseCategoryOfficeSupplies
	case strings.Contains(s, "utilit"):
		return models.ExpenseCategoryUtilities
	case strings.Contains(s, "marketing") || strings.Contains(s, "advertis"):
		return models.ExpenseCategoryMarketing
	default:
		return models.ExpenseCategoryOther
	}
}
