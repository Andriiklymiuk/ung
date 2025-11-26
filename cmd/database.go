package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var databaseCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"db"},
	Short:   "Manage multiple databases",
	Long: `Manage multiple databases for different projects or clients.

Examples:
  ung db list              # List all known databases
  ung db switch ./work.db  # Switch to work database
  ung db current           # Show current database
  ung db info              # Show database statistics`,
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known databases",
	Long:  `List all databases found in workspace configs and recent usage.`,
	Run:   runDBList,
}

var dbSwitchCmd = &cobra.Command{
	Use:   "switch <database-path>",
	Short: "Switch to a different database",
	Long: `Switch to a different database by creating or updating the local workspace config.

The database path can be:
  - Relative: ./mydb.db or data/billing.db
  - Absolute: /full/path/to/db.db
  - Tilde: ~/.ung/production.db

Examples:
  ung db switch ./client-a.db
  ung db switch ~/.ung/production.db`,
	Args: cobra.ExactArgs(1),
	Run:  runDBSwitch,
}

var dbCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current database",
	Long:  `Display the currently active database path and statistics.`,
	Run:   runDBCurrent,
}

var dbInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show database statistics",
	Long:  `Display statistics about the current database including record counts.`,
	Run:   runDBInfo,
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database (delete all data)",
	Long: `⚠️  DANGER: Reset the database by deleting ALL data.

This command will permanently delete:
  • All companies
  • All clients
  • All contracts
  • All invoices
  • All tracking sessions
  • All expenses
  • All recurring invoices

This action CANNOT be undone! Make sure to backup first with 'ung sync backup'.`,
	Run: runDBReset,
}

func init() {
	databaseCmd.AddCommand(dbListCmd)
	databaseCmd.AddCommand(dbSwitchCmd)
	databaseCmd.AddCommand(dbCurrentCmd)
	databaseCmd.AddCommand(dbInfoCmd)
	databaseCmd.AddCommand(dbResetCmd)
	rootCmd.AddCommand(databaseCmd)
}

func runDBList(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Known Databases")

	// Get current database
	currentDB := db.GetDBPath()

	// Find databases in common locations
	databases := findDatabases()

	if len(databases) == 0 {
		fmt.Println("No databases found.")
		fmt.Println("\n💡 Create a workspace database:")
		fmt.Println("   ung config init --local")
		return
	}

	for _, dbPath := range databases {
		isCurrent := dbPath == currentDB
		icon := "⚪"
		suffix := ""

		if isCurrent {
			icon = "✅"
			suffix = " (current)"
		}

		// Show relative path if in current directory
		displayPath := dbPath
		if rel, err := filepath.Rel(".", dbPath); err == nil && !strings.HasPrefix(rel, "..") {
			displayPath = "./" + rel
		}

		// Get database stats
		stats := getDBStats(dbPath)

		fmt.Printf("%s %s%s\n", icon, displayPath, suffix)
		if stats != "" {
			fmt.Printf("   %s\n", stats)
		}
	}
}

func runDBSwitch(cmd *cobra.Command, args []string) {
	newDBPath := args[0]

	// Expand path
	newDBPath = expandPathForUser(newDBPath)

	fmt.Printf("🔄 Switching to database: %s\n", newDBPath)

	// Check if database file exists
	if _, err := os.Stat(newDBPath); os.IsNotExist(err) {
		fmt.Print("\n⚠️  Database file doesn't exist. Create it? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled")
			return
		}

		// Create directory if needed
		dir := filepath.Dir(newDBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory: %v\n", err)
			return
		}
	}

	// Create or update local config
	cfg, _ := config.Load()
	cfg.DatabasePath = newDBPath

	// Determine invoices directory relative to database
	dbDir := filepath.Dir(newDBPath)
	cfg.InvoicesDir = filepath.Join(dbDir, "invoices")

	if err := config.Save(cfg, false); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}

	fmt.Printf("✅ Switched to: %s\n", newDBPath)
	fmt.Printf("   Invoices:   %s\n", cfg.InvoicesDir)
	fmt.Println("\n💡 Run 'ung doctor' to verify the setup")
}

func runDBCurrent(cmd *cobra.Command, args []string) {
	currentDB := db.GetDBPath()
	invoicesDir := db.GetInvoicesDir()

	fmt.Println("📍 Current Database")
	fmt.Printf("Database:  %s\n", currentDB)
	fmt.Printf("Invoices:  %s\n", invoicesDir)

	// Check if file exists
	if _, err := os.Stat(currentDB); os.IsNotExist(err) {
		fmt.Println("\n⚠️  Database file doesn't exist yet (will be created on first use)")
		return
	}

	// Show stats
	stats := getDBStats(currentDB)
	if stats != "" {
		fmt.Printf("\n%s\n", stats)
	}

	// Show config source
	configSource := "default"
	if _, err := os.Stat(".ung.yaml"); err == nil {
		configSource = ".ung.yaml (local workspace)"
	} else {
		home, _ := os.UserHomeDir()
		globalConfig := filepath.Join(home, ".ung", "config.yaml")
		if _, err := os.Stat(globalConfig); err == nil {
			configSource = globalConfig + " (global)"
		}
	}
	fmt.Printf("\nConfig:    %s\n", configSource)
}

func runDBInfo(cmd *cobra.Command, args []string) {
	currentDB := db.GetDBPath()

	if _, err := os.Stat(currentDB); os.IsNotExist(err) {
		fmt.Println("❌ Database doesn't exist yet")
		return
	}

	fmt.Println("📈 Database Statistics")

	// Get detailed stats
	if db.DB == nil {
		fmt.Println("⚠️  Database not initialized")
		return
	}

	var companyCount, clientCount, contractCount, invoiceCount, sessionCount, expenseCount int

	db.DB.QueryRow("SELECT COUNT(*) FROM companies").Scan(&companyCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM clients").Scan(&clientCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM contracts").Scan(&contractCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM invoices").Scan(&invoiceCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE deleted_at IS NULL").Scan(&sessionCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM expenses").Scan(&expenseCount)

	fmt.Printf("Companies:         %d\n", companyCount)
	fmt.Printf("Clients:           %d\n", clientCount)
	fmt.Printf("Contracts:         %d\n", contractCount)
	fmt.Printf("Invoices:          %d\n", invoiceCount)
	fmt.Printf("Tracking Sessions: %d\n", sessionCount)
	fmt.Printf("Expenses:          %d\n", expenseCount)

	// Get file size
	if info, err := os.Stat(currentDB); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("\nFile Size:         %.2f MB\n", sizeMB)
	}

	// Get active contracts
	var activeContracts int
	db.DB.QueryRow("SELECT COUNT(*) FROM contracts WHERE active = 1").Scan(&activeContracts)
	fmt.Printf("Active Contracts:  %d\n", activeContracts)

	// Get unbilled sessions
	var unbilledSessions int
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM tracking_sessions
		WHERE billable = 1
		  AND deleted_at IS NULL
		  AND (notes NOT LIKE '%[Invoiced:%' OR notes IS NULL)
	`).Scan(&unbilledSessions)
	fmt.Printf("Unbilled Sessions: %d\n", unbilledSessions)
}

func findDatabases() []string {
	var databases []string
	seen := make(map[string]bool)

	// 1. Current config database
	currentDB := db.GetDBPath()
	if currentDB != "" {
		databases = append(databases, currentDB)
		seen[currentDB] = true
	}

	// 2. Local workspace database
	if _, err := os.Stat(".ung.yaml"); err == nil {
		// There's a local config, it's already loaded as currentDB
	}

	// 3. Look for *.db files in current directory
	files, _ := filepath.Glob("*.db")
	for _, file := range files {
		absPath, _ := filepath.Abs(file)
		if !seen[absPath] {
			databases = append(databases, absPath)
			seen[absPath] = true
		}
	}

	// 4. Look in common subdirectories
	commonDirs := []string{"data", "db", "database", "."}
	for _, dir := range commonDirs {
		pattern := filepath.Join(dir, "*.db")
		files, _ := filepath.Glob(pattern)
		for _, file := range files {
			absPath, _ := filepath.Abs(file)
			if !seen[absPath] {
				databases = append(databases, absPath)
				seen[absPath] = true
			}
		}
	}

	return databases
}

func getDBStats(dbPath string) string {
	// Quick stats without opening DB
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "not created yet"
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return ""
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	if sizeMB < 0.01 {
		return fmt.Sprintf("%.0f KB", float64(info.Size())/1024)
	}
	return fmt.Sprintf("%.2f MB", sizeMB)
}

func expandPathForUser(path string) string {
	if len(path) == 0 {
		return path
	}

	// Expand ~ to home directory
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert relative paths to absolute
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		return absPath
	}

	return path
}

func runDBReset(cmd *cobra.Command, args []string) {
	currentDB := db.GetDBPath()

	if _, err := os.Stat(currentDB); os.IsNotExist(err) {
		fmt.Println("❌ Database doesn't exist yet. Nothing to reset.")
		return
	}

	// Show current stats
	fmt.Println("⚠️  DATABASE RESET WARNING")
	fmt.Println("==========================")
	fmt.Printf("\nDatabase: %s\n", currentDB)

	if db.DB == nil {
		fmt.Println("⚠️  Database not initialized")
		return
	}

	// Get current record counts
	var companyCount, clientCount, contractCount, invoiceCount, sessionCount, expenseCount, recurringCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM companies").Scan(&companyCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM clients").Scan(&clientCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM contracts").Scan(&contractCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM invoices").Scan(&invoiceCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions").Scan(&sessionCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM expenses").Scan(&expenseCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM recurring_invoices").Scan(&recurringCount)

	fmt.Println("\n⚠️  The following data will be PERMANENTLY DELETED:")
	fmt.Printf("   • %d companies\n", companyCount)
	fmt.Printf("   • %d clients\n", clientCount)
	fmt.Printf("   • %d contracts\n", contractCount)
	fmt.Printf("   • %d invoices\n", invoiceCount)
	fmt.Printf("   • %d tracking sessions\n", sessionCount)
	fmt.Printf("   • %d expenses\n", expenseCount)
	fmt.Printf("   • %d recurring invoices\n", recurringCount)

	totalRecords := companyCount + clientCount + contractCount + invoiceCount + sessionCount + expenseCount + recurringCount
	if totalRecords == 0 {
		fmt.Println("\nDatabase is already empty. Nothing to reset.")
		return
	}

	fmt.Println("\n💡 Tip: Run 'ung sync backup' first to create a backup!")

	// First confirmation
	var confirm1 bool
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to reset the database?").
				Description("This action CANNOT be undone!").
				Affirmative("Yes, I understand").
				Negative("Cancel").
				Value(&confirm1),
		),
	)

	if err := form1.Run(); err != nil || !confirm1 {
		fmt.Println("\n✅ Reset cancelled. Your data is safe.")
		return
	}

	// Second confirmation - require typing "RESET"
	var confirmText string
	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Type RESET to confirm").
				Description("This is your last chance to cancel!").
				Placeholder("Type RESET").
				Value(&confirmText).
				Validate(func(s string) error {
					if s != "RESET" {
						return fmt.Errorf("please type RESET exactly to confirm")
					}
					return nil
				}),
		),
	)

	if err := form2.Run(); err != nil {
		fmt.Println("\n✅ Reset cancelled. Your data is safe.")
		return
	}

	if confirmText != "RESET" {
		fmt.Println("\n✅ Reset cancelled. Your data is safe.")
		return
	}

	// Perform the reset
	fmt.Println("\n🔄 Resetting database...")

	// Delete all data in order (respect foreign keys)
	tables := []string{
		"invoice_line_items",
		"invoice_recipients",
		"invoices",
		"tracking_sessions",
		"expenses",
		"recurring_invoices",
		"contracts",
		"clients",
		"companies",
	}

	for _, table := range tables {
		_, err := db.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to clear %s: %v\n", table, err)
		}
	}

	// Reset SQLite auto-increment counters
	db.DB.Exec("DELETE FROM sqlite_sequence")

	fmt.Println("\n✅ Database has been reset successfully!")
	fmt.Println("   All data has been permanently deleted.")
	fmt.Println("\n💡 Run 'ung create' to set up your data again.")
}
