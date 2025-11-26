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
	Short: "Import data from CSV or SQLite files",
	Long: `Import data from CSV or SQLite files into UNG.

Supported sources:
  - CSV files (clients, expenses, time)
  - SQLite databases (full database import)

CSV Format Requirements:
  Clients:   name,email,address,tax_id
  Expenses:  date,description,amount,category,vendor,currency
  Time:      date,client,project,hours,billable,notes

Examples:
  ung import                              Interactive import wizard
  ung import --file clients.csv --type clients
  ung import db backup.db                 Import from SQLite
  ung import db backup.db --password xyz  Import from encrypted SQLite`,
	RunE: runImport,
}

var importDBCmd = &cobra.Command{
	Use:   "db <file>",
	Short: "Import data from another UNG SQLite database",
	Long: `Import all data from another UNG SQLite database.

Supports both encrypted and non-encrypted databases.
Data is merged with existing records (duplicates are skipped).

Examples:
  ung import db backup.db                  Import from unencrypted database
  ung import db backup.db --password mykey Import from encrypted database`,
	Args: cobra.ExactArgs(1),
	RunE: runImportDB,
}

var (
	importFile     string
	importType     string
	importSkipRows int
	importDryRun   bool
	importPassword string
)

func init() {
	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "CSV file to import")
	importCmd.Flags().StringVarP(&importType, "type", "t", "", "Data type: clients, expenses, time")
	importCmd.Flags().IntVar(&importSkipRows, "skip", 1, "Number of header rows to skip")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview import without saving")

	// SQLite import flags
	importDBCmd.Flags().StringVarP(&importPassword, "password", "p", "", "Password for encrypted database")
	importDBCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview import without saving")

	importCmd.AddCommand(importDBCmd)
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
					huh.NewOption("SQLite Database (full import)", "sqlite"),
					huh.NewOption("Clients CSV (name, email, address, tax_id)", "clients"),
					huh.NewOption("Expenses CSV (date, description, amount, category)", "expenses"),
					huh.NewOption("Time Tracking CSV (date, client, project, hours)", "time"),
				).
				Value(&dataType),
		),
	)

	if err := typeForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	// Handle SQLite import separately
	if dataType == "sqlite" {
		return runImportDBInteractive()
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
		if err := db.GormDB.Where("email = ?", client.Email).First(&existing).Error; err == nil {
			fmt.Printf("  [Skip] %s (already exists)\n", client.Email)
			skipped++
			continue
		}

		if err := db.GormDB.Create(&client).Error; err != nil {
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

		if err := db.GormDB.Create(&expense).Error; err != nil {
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
	db.GormDB.Find(&clients)
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

		if err := db.GormDB.Create(&session).Error; err != nil {
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

// runImportDB imports data from another UNG SQLite database
func runImportDB(cmd *cobra.Command, args []string) error {
	dbFile := args[0]

	// Check if file exists
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return fmt.Errorf("database file not found: %s", dbFile)
	}

	fmt.Printf("\nImporting from SQLite database: %s\n", dbFile)
	if importPassword != "" {
		fmt.Println("Using password for encrypted database")
	}

	// Build connection string
	dsn := dbFile
	if importPassword != "" {
		dsn = fmt.Sprintf("%s?_pragma_key=%s", dbFile, importPassword)
	}

	// Open source database using raw SQL
	sourceDB, err := openSourceDB(dsn, importPassword)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceDB.Close()

	// Test if we can read from the database
	var count int
	if err := sourceDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		if importPassword == "" {
			return fmt.Errorf("failed to read database. If encrypted, provide password with --password flag: %w", err)
		}
		return fmt.Errorf("failed to read database. Check if password is correct: %w", err)
	}

	imported := make(map[string]int)
	skipped := make(map[string]int)

	// Import companies
	companies, skip := importCompaniesFromDB(sourceDB)
	imported["companies"] = companies
	skipped["companies"] = skip

	// Import clients
	clients, skip := importClientsFromDB(sourceDB)
	imported["clients"] = clients
	skipped["clients"] = skip

	// Import contracts
	contracts, skip := importContractsFromDB(sourceDB)
	imported["contracts"] = contracts
	skipped["contracts"] = skip

	// Import invoices
	invoices, skip := importInvoicesFromDB(sourceDB)
	imported["invoices"] = invoices
	skipped["invoices"] = skip

	// Import expenses
	expenses, skip := importExpensesFromDB(sourceDB)
	imported["expenses"] = expenses
	skipped["expenses"] = skip

	// Import tracking sessions
	sessions, skip := importSessionsFromDB(sourceDB)
	imported["time_entries"] = sessions
	skipped["time_entries"] = skip

	// Import recurring invoices
	recurring, skip := importRecurringFromDB(sourceDB)
	imported["recurring"] = recurring
	skipped["recurring"] = skip

	fmt.Println("\n✓ Import completed!")
	fmt.Println("\n  Imported:")
	for k, v := range imported {
		if v > 0 {
			fmt.Printf("    • %d %s\n", v, k)
		}
	}
	if hasSkipped(skipped) {
		fmt.Println("\n  Skipped (already exist):")
		for k, v := range skipped {
			if v > 0 {
				fmt.Printf("    • %d %s\n", v, k)
			}
		}
	}

	return nil
}

// runImportDBInteractive runs SQLite import with interactive prompts
func runImportDBInteractive() error {
	var dbPath string
	var password string
	var usePassword bool

	// Get database file path
	pathForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("SQLite Database Path").
				Description("Enter the path to the UNG database file").
				Placeholder("/path/to/ung.db").
				Value(&dbPath),
		),
	)

	if err := pathForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	// Expand ~ to home directory
	if strings.HasPrefix(dbPath, "~/") {
		dbPath = strings.Replace(dbPath, "~", os.Getenv("HOME"), 1)
	}

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file not found: %s", dbPath)
	}

	// Ask about encryption
	encryptForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Is the database encrypted?").
				Description("If the database is password-protected").
				Affirmative("Yes").
				Negative("No").
				Value(&usePassword),
		),
	)

	if err := encryptForm.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	if usePassword {
		pwForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Database Password").
					Description("Enter the encryption password").
					EchoMode(huh.EchoModePassword).
					Value(&password),
			),
		)

		if err := pwForm.Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
	}

	// Set global variables and run import
	importPassword = password
	return runImportDB(nil, []string{dbPath})
}

func hasSkipped(m map[string]int) bool {
	for _, v := range m {
		if v > 0 {
			return true
		}
	}
	return false
}

// openSourceDB opens a source SQLite database
func openSourceDB(dsn, password string) (*db.SQLiteDB, error) {
	return db.OpenSQLite(dsn, password)
}

func importCompaniesFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT name, email, phone, address, registration_address, tax_id, bank_name, bank_account, bank_swift FROM companies`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var c models.Company
		if err := rows.Scan(&c.Name, &c.Email, &c.Phone, &c.Address, &c.RegistrationAddress, &c.TaxID, &c.BankName, &c.BankAccount, &c.BankSWIFT); err != nil {
			continue
		}

		// Check if exists
		var existing models.Company
		if err := db.GormDB.Where("email = ?", c.Email).First(&existing).Error; err == nil {
			skipped++
			continue
		}

		if !importDryRun {
			if err := db.GormDB.Create(&c).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}

func importClientsFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT name, email, address, tax_id FROM clients`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.Name, &c.Email, &c.Address, &c.TaxID); err != nil {
			continue
		}

		var existing models.Client
		if err := db.GormDB.Where("email = ?", c.Email).First(&existing).Error; err == nil {
			skipped++
			continue
		}

		if !importDryRun {
			if err := db.GormDB.Create(&c).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}

func importContractsFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT contract_num, client_id, name, contract_type, hourly_rate, fixed_price, currency, start_date, end_date, active, notes FROM contracts`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var c models.Contract
		if err := rows.Scan(&c.ContractNum, &c.ClientID, &c.Name, &c.ContractType, &c.HourlyRate, &c.FixedPrice, &c.Currency, &c.StartDate, &c.EndDate, &c.Active, &c.Notes); err != nil {
			continue
		}

		var existing models.Contract
		if err := db.GormDB.Where("contract_num = ?", c.ContractNum).First(&existing).Error; err == nil {
			skipped++
			continue
		}

		if !importDryRun {
			if err := db.GormDB.Create(&c).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}

func importInvoicesFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT invoice_num, company_id, amount, currency, description, status, issued_date, due_date FROM invoices`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.InvoiceNum, &inv.CompanyID, &inv.Amount, &inv.Currency, &inv.Description, &inv.Status, &inv.IssuedDate, &inv.DueDate); err != nil {
			continue
		}

		var existing models.Invoice
		if err := db.GormDB.Where("invoice_num = ?", inv.InvoiceNum).First(&existing).Error; err == nil {
			skipped++
			continue
		}

		if !importDryRun {
			if err := db.GormDB.Create(&inv).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}

func importExpensesFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT description, amount, currency, category, date, vendor, notes FROM expenses`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.Description, &e.Amount, &e.Currency, &e.Category, &e.Date, &e.Vendor, &e.Notes); err != nil {
			continue
		}

		// For expenses, we import all (no unique key to check)
		if !importDryRun {
			if err := db.GormDB.Create(&e).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}

func importSessionsFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT client_id, contract_id, project_name, start_time, end_time, duration, hours, billable, notes FROM tracking_sessions WHERE deleted_at IS NULL`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var s models.TrackingSession
		if err := rows.Scan(&s.ClientID, &s.ContractID, &s.ProjectName, &s.StartTime, &s.EndTime, &s.Duration, &s.Hours, &s.Billable, &s.Notes); err != nil {
			continue
		}

		// Import all sessions (time entries are typically unique)
		if !importDryRun {
			if err := db.GormDB.Create(&s).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}

func importRecurringFromDB(sourceDB *db.SQLiteDB) (int, int) {
	rows, err := sourceDB.Query(`SELECT client_id, contract_id, amount, currency, description, frequency, day_of_month, day_of_week, next_generation_date, active FROM recurring_invoices`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	imported := 0
	skipped := 0

	for rows.Next() {
		var r models.RecurringInvoice
		if err := rows.Scan(&r.ClientID, &r.ContractID, &r.Amount, &r.Currency, &r.Description, &r.Frequency, &r.DayOfMonth, &r.DayOfWeek, &r.NextGenerationDate, &r.Active); err != nil {
			continue
		}

		// Check for similar recurring invoice
		var existing models.RecurringInvoice
		if err := db.GormDB.Where("client_id = ? AND amount = ? AND frequency = ?", r.ClientID, r.Amount, r.Frequency).First(&existing).Error; err == nil {
			skipped++
			continue
		}

		if !importDryRun {
			if err := db.GormDB.Create(&r).Error; err == nil {
				imported++
			}
		} else {
			imported++
		}
	}

	return imported, skipped
}
