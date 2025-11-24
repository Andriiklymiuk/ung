package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
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

func init() {
	databaseCmd.AddCommand(dbListCmd)
	databaseCmd.AddCommand(dbSwitchCmd)
	databaseCmd.AddCommand(dbCurrentCmd)
	databaseCmd.AddCommand(dbInfoCmd)
	rootCmd.AddCommand(databaseCmd)
}

func runDBList(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Known Databases\n")

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

	fmt.Println("📍 Current Database\n")
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

	fmt.Println("📈 Database Statistics\n")

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
