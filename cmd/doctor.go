package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the health of your ung installation",
	Long: `Run diagnostics on your ung installation to identify and fix common issues.

This command will check:
- Database connection and schema
- Configuration files
- Required directories and permissions
- Data integrity`,
	Run: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type CheckResult struct {
	Name    string
	Status  string // "ok", "warning", "error"
	Message string
}

func runDoctor(cmd *cobra.Command, args []string) {
	fmt.Println("🏥 Running ung health checks...\n")

	checks := []CheckResult{}

	// Check 1: Database file exists
	checks = append(checks, checkDatabaseExists())

	// Check 2: Database connection
	checks = append(checks, checkDatabaseConnection())

	// Check 3: Database schema
	checks = append(checks, checkDatabaseSchema())

	// Check 4: Configuration
	checks = append(checks, checkConfiguration())

	// Check 5: Directories
	checks = append(checks, checkDirectories())

	// Check 6: Permissions
	checks = append(checks, checkPermissions())

	// Check 7: Data integrity
	checks = append(checks, checkDataIntegrity())

	// Print results
	okCount := 0
	warningCount := 0
	errorCount := 0

	for _, check := range checks {
		var icon string
		switch check.Status {
		case "ok":
			icon = "✅"
			okCount++
		case "warning":
			icon = "⚠️ "
			warningCount++
		case "error":
			icon = "❌"
			errorCount++
		}

		fmt.Printf("%s %s\n", icon, check.Name)
		if check.Message != "" {
			fmt.Printf("   %s\n", check.Message)
		}
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("═", 50))
	fmt.Printf("Summary: %d passed, %d warnings, %d errors\n", okCount, warningCount, errorCount)

	if errorCount == 0 && warningCount == 0 {
		fmt.Println("\n🎉 Everything looks good! Your ung installation is healthy.")
	} else if errorCount > 0 {
		fmt.Println("\n⚠️  Some issues were found. Please address the errors above.")
	} else {
		fmt.Println("\n✓ Your installation is functional, but some warnings were found.")
	}
}

func checkDatabaseExists() CheckResult {
	dbPath := db.GetDBPath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return CheckResult{
			Name:    "Database file",
			Status:  "warning",
			Message: fmt.Sprintf("Database not found at %s (will be created on first use)", dbPath),
		}
	}

	return CheckResult{
		Name:    "Database file",
		Status:  "ok",
		Message: fmt.Sprintf("Found at %s", dbPath),
	}
}

func checkDatabaseConnection() CheckResult {
	if db.DB == nil {
		return CheckResult{
			Name:    "Database connection",
			Status:  "error",
			Message: "Database not initialized",
		}
	}

	if err := db.DB.Ping(); err != nil {
		return CheckResult{
			Name:    "Database connection",
			Status:  "error",
			Message: fmt.Sprintf("Cannot connect: %v", err),
		}
	}

	return CheckResult{
		Name:   "Database connection",
		Status: "ok",
	}
}

func checkDatabaseSchema() CheckResult {
	if db.DB == nil {
		return CheckResult{
			Name:    "Database schema",
			Status:  "error",
			Message: "Database not initialized",
		}
	}

	// Check for required tables
	requiredTables := []string{
		"companies",
		"clients",
		"contracts",
		"invoices",
		"invoice_line_items",
		"invoice_recipients",
		"tracking_sessions",
		"expenses",
	}

	missingTables := []string{}
	for _, table := range requiredTables {
		var exists bool
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		err := db.DB.QueryRow(query, table).Scan(&exists)
		if err == sql.ErrNoRows {
			missingTables = append(missingTables, table)
		}
	}

	if len(missingTables) > 0 {
		return CheckResult{
			Name:    "Database schema",
			Status:  "error",
			Message: fmt.Sprintf("Missing tables: %v", missingTables),
		}
	}

	return CheckResult{
		Name:    "Database schema",
		Status:  "ok",
		Message: "All required tables present",
	}
}

func checkConfiguration() CheckResult {
	cfg, err := config.Load()
	if err != nil {
		return CheckResult{
			Name:    "Configuration",
			Status:  "error",
			Message: fmt.Sprintf("Failed to load config: %v", err),
		}
	}

	if cfg.DatabasePath == "" {
		return CheckResult{
			Name:    "Configuration",
			Status:  "error",
			Message: "Database path not configured",
		}
	}

	return CheckResult{
		Name:    "Configuration",
		Status:  "ok",
		Message: fmt.Sprintf("Home directory: %s", filepath.Dir(cfg.DatabasePath)),
	}
}

func checkDirectories() CheckResult {
	ungDir := filepath.Dir(db.GetDBPath())
	invoicesDir := db.GetInvoicesDir()

	dirsToCheck := []string{ungDir, invoicesDir}
	missingDirs := []string{}

	for _, dir := range dirsToCheck {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			missingDirs = append(missingDirs, dir)
		}
	}

	if len(missingDirs) > 0 {
		return CheckResult{
			Name:    "Required directories",
			Status:  "warning",
			Message: fmt.Sprintf("Missing: %v (will be created automatically)", missingDirs),
		}
	}

	return CheckResult{
		Name:    "Required directories",
		Status:  "ok",
		Message: fmt.Sprintf("~/.ung and ~/.ung/invoices exist"),
	}
}

func checkPermissions() CheckResult {
	ungDir := filepath.Dir(db.GetDBPath())

	// Check if we can write to the ung directory
	testFile := filepath.Join(ungDir, ".test_write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return CheckResult{
			Name:    "Directory permissions",
			Status:  "error",
			Message: fmt.Sprintf("Cannot write to %s: %v", ungDir, err),
		}
	}
	os.Remove(testFile)

	return CheckResult{
		Name:   "Directory permissions",
		Status: "ok",
	}
}

func checkDataIntegrity() CheckResult {
	if db.DB == nil {
		return CheckResult{
			Name:    "Data integrity",
			Status:  "error",
			Message: "Database not initialized",
		}
	}

	// Count records in main tables
	var companyCount, clientCount, contractCount, invoiceCount int

	db.DB.QueryRow("SELECT COUNT(*) FROM companies").Scan(&companyCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM clients").Scan(&clientCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM contracts").Scan(&contractCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM invoices").Scan(&invoiceCount)

	message := fmt.Sprintf("%d companies, %d clients, %d contracts, %d invoices",
		companyCount, clientCount, contractCount, invoiceCount)

	// Check for orphaned records
	var orphanedContracts int
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM contracts
		WHERE client_id NOT IN (SELECT id FROM clients)
	`).Scan(&orphanedContracts)

	if orphanedContracts > 0 {
		return CheckResult{
			Name:    "Data integrity",
			Status:  "warning",
			Message: fmt.Sprintf("%s, %d orphaned contracts", message, orphanedContracts),
		}
	}

	return CheckResult{
		Name:    "Data integrity",
		Status:  "ok",
		Message: message,
	}
}
