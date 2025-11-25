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
var isEncrypted bool
var encryptedPath string
var decryptedPath string

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

	// Check if encryption is enabled
	cfg, _ := config.Load()
	encryptedDBPath := dbPath + ".encrypted"

	// Handle encrypted database
	if cfg.Security.EncryptDatabase || fileExists(encryptedDBPath) {
		isEncrypted = true
		encryptedPath = encryptedDBPath
		decryptedPath = dbPath + ".decrypted"

		// If encrypted file exists, decrypt it
		if fileExists(encryptedDBPath) {
			password, err := GetDatabasePassword()
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			if err := DecryptDatabase(encryptedDBPath, decryptedPath, password); err != nil {
				return fmt.Errorf("failed to decrypt database: %w", err)
			}

			dbPath = decryptedPath
		} else {
			// First time encryption - use regular path, will encrypt on Close
			decryptedPath = dbPath
		}
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

// Close closes the database connection and re-encrypts if needed
func Close() error {
	if DB != nil {
		if err := DB.Close(); err != nil {
			return err
		}
	}

	// Re-encrypt database if encryption is enabled
	if isEncrypted && decryptedPath != "" && encryptedPath != "" {
		password, err := GetDatabasePassword()
		if err != nil {
			return fmt.Errorf("failed to get password for encryption: %w", err)
		}

		// Encrypt the database
		if err := EncryptDatabase(decryptedPath, encryptedPath, password); err != nil {
			return fmt.Errorf("failed to encrypt database: %w", err)
		}

		// Clean up decrypted file
		os.Remove(decryptedPath)

		// Clear password cache for security
		ClearPasswordCache()
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
		phone TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
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

	CREATE TABLE IF NOT EXISTS contracts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		contract_num TEXT UNIQUE NOT NULL,
		client_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		contract_type TEXT NOT NULL,
		hourly_rate REAL,
		fixed_price REAL,
		currency TEXT DEFAULT 'USD',
		start_date TIMESTAMP NOT NULL,
		end_date TIMESTAMP,
		active BOOLEAN DEFAULT 1,
		notes TEXT,
		pdf_path TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (client_id) REFERENCES clients(id)
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
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (company_id) REFERENCES companies(id)
	);

	CREATE TABLE IF NOT EXISTS invoice_line_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invoice_id INTEGER NOT NULL,
		item_name TEXT NOT NULL,
		description TEXT,
		quantity REAL NOT NULL,
		rate REAL NOT NULL,
		amount REAL NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (invoice_id) REFERENCES invoices(id)
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
		contract_id INTEGER,
		project_name TEXT,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP,
		duration INTEGER,
		hours REAL,
		billable BOOLEAN DEFAULT 1,
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP,
		FOREIGN KEY (client_id) REFERENCES clients(id),
		FOREIGN KEY (contract_id) REFERENCES contracts(id)
	);

	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL,
		amount REAL NOT NULL,
		currency TEXT DEFAULT 'USD',
		category TEXT NOT NULL,
		date TIMESTAMP NOT NULL,
		vendor TEXT,
		notes TEXT,
		receipt_path TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS recurring_invoices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id INTEGER NOT NULL,
		contract_id INTEGER,
		amount REAL NOT NULL,
		currency TEXT DEFAULT 'USD',
		description TEXT,
		frequency TEXT NOT NULL,
		day_of_month INTEGER DEFAULT 1,
		day_of_week INTEGER DEFAULT 1,
		next_generation_date TIMESTAMP NOT NULL,
		last_generated_date TIMESTAMP,
		last_invoice_id INTEGER,
		active BOOLEAN DEFAULT 1,
		auto_send BOOLEAN DEFAULT 0,
		auto_pdf BOOLEAN DEFAULT 1,
		email_app TEXT,
		generated_count INTEGER DEFAULT 0,
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (client_id) REFERENCES clients(id),
		FOREIGN KEY (contract_id) REFERENCES contracts(id),
		FOREIGN KEY (last_invoice_id) REFERENCES invoices(id)
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_company ON invoices(company_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_recipients_invoice ON invoice_recipients(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_recipients_client ON invoice_recipients(client_id);
	CREATE INDEX IF NOT EXISTS idx_tracking_sessions_client ON tracking_sessions(client_id);
	CREATE INDEX IF NOT EXISTS idx_tracking_sessions_contract ON tracking_sessions(contract_id);
	CREATE INDEX IF NOT EXISTS idx_contracts_client ON contracts(client_id);
	CREATE INDEX IF NOT EXISTS idx_contracts_active ON contracts(active);
	CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice ON invoice_line_items(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_recurring_invoices_client ON recurring_invoices(client_id);
	CREATE INDEX IF NOT EXISTS idx_recurring_invoices_active ON recurring_invoices(active);
	CREATE INDEX IF NOT EXISTS idx_recurring_invoices_next_date ON recurring_invoices(next_generation_date);
	`

	_, err := DB.Exec(schema)
	return err
}

// SQLiteDB wraps a raw SQL database connection for import operations
type SQLiteDB struct {
	*sql.DB
}

// OpenSQLite opens a SQLite database for reading (used for imports)
func OpenSQLite(path, password string) (*SQLiteDB, error) {
	dsn := path
	if password != "" {
		// SQLCipher format for encrypted databases
		dsn = fmt.Sprintf("file:%s?_pragma_key=%s&_pragma_cipher_page_size=4096", path, password)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &SQLiteDB{db}, nil
}
