package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *sql.DB
var GormDB *gorm.DB

// GetDBPath returns the path to the database file
func GetDBPath() string {
	return config.GetDatabasePath()
}

// GetInvoicesDir returns the path to the invoices directory
func GetInvoicesDir() string {
	return config.GetInvoicesDir()
}

// Initialize opens the database and runs migrations
func Initialize() error {
	dbPath := GetDBPath()

	// Create .ung directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create invoices directory
	invoicesDir := GetInvoicesDir()
	if err := os.MkdirAll(invoicesDir, 0755); err != nil {
		return fmt.Errorf("failed to create invoices directory: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := runMigrations(dbPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize GORM
	GormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize GORM: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// runMigrations runs database migrations
func runMigrations(dbPath string) error {
	migrationsDir := "migrations"

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return runInlineSchema()
	}

	// Create migrations table to track applied migrations
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return runInlineSchema()
	}

	// Execute each .up.sql file in order
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process .up.sql files
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".up.sql") {
			continue
		}

		version := fileName[:len(fileName)-7] // Remove .up.sql

		// Check if already applied
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err == nil && count > 0 {
			continue // Already applied
		}

		// Read and execute migration
		migrationPath := filepath.Join(migrationsDir, file.Name())
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		_, err = DB.Exec(string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}

		// Record migration
		_, err = DB.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// runInlineSchema is a fallback for when migration files aren't available
func runInlineSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS companies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		address TEXT,
		tax_id TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		address TEXT,
		tax_id TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS invoices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invoice_num TEXT UNIQUE NOT NULL,
		company_id INTEGER NOT NULL,
		amount REAL NOT NULL,
		currency TEXT DEFAULT 'USD',
		description TEXT,
		status TEXT DEFAULT 'pending',
		issued_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		due_date TIMESTAMP,
		pdf_path TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (company_id) REFERENCES companies(id)
	);

	CREATE TABLE IF NOT EXISTS invoice_recipients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invoice_id INTEGER NOT NULL,
		client_id INTEGER NOT NULL,
		FOREIGN KEY (invoice_id) REFERENCES invoices(id),
		FOREIGN KEY (client_id) REFERENCES clients(id)
	);

	CREATE TABLE IF NOT EXISTS tracking_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id INTEGER,
		project_name TEXT,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP,
		duration INTEGER,
		billable BOOLEAN DEFAULT 1,
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (client_id) REFERENCES clients(id)
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_company ON invoices(company_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_recipients_invoice ON invoice_recipients(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_recipients_client ON invoice_recipients(client_id);
	CREATE INDEX IF NOT EXISTS idx_tracking_sessions_client ON tracking_sessions(client_id);
	`

	_, err := DB.Exec(schema)
	return err
}
