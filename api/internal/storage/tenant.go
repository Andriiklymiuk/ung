package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TenantManager manages tenant-specific databases with S3 storage and encryption
type TenantManager struct {
	storage     *S3Storage
	cache       map[string]*TenantDB
	cacheMutex  sync.RWMutex
	localDir    string
	autoSync    bool
	syncInterval time.Duration
}

// TenantDB represents a tenant's database connection
type TenantDB struct {
	TenantID     string
	DB           *sql.DB
	LocalPath    string
	LastSync     time.Time
	LastAccess   time.Time
	Password     string
	mutex        sync.RWMutex
}

// TenantManagerConfig holds configuration for tenant manager
type TenantManagerConfig struct {
	S3Config       *S3Config
	LocalDir       string
	AutoSync       bool
	SyncInterval   time.Duration
	CacheSize      int
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(cfg *TenantManagerConfig) (*TenantManager, error) {
	// Create S3 storage
	storage, err := NewS3Storage(cfg.S3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 storage: %w", err)
	}

	// Ensure local directory exists
	if err := os.MkdirAll(cfg.LocalDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local directory: %w", err)
	}

	tm := &TenantManager{
		storage:      storage,
		cache:        make(map[string]*TenantDB),
		localDir:     cfg.LocalDir,
		autoSync:     cfg.AutoSync,
		syncInterval: cfg.SyncInterval,
	}

	// Start auto-sync if enabled
	if tm.autoSync {
		go tm.autoSyncWorker()
	}

	return tm, nil
}

// GetTenantDB gets or creates a tenant database connection
func (tm *TenantManager) GetTenantDB(ctx context.Context, tenantID string, password string) (*TenantDB, error) {
	// Check cache first
	tm.cacheMutex.RLock()
	if tdb, exists := tm.cache[tenantID]; exists {
		tdb.LastAccess = time.Now()
		tm.cacheMutex.RUnlock()
		return tdb, nil
	}
	tm.cacheMutex.RUnlock()

	// Not in cache, load from S3
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if tdb, exists := tm.cache[tenantID]; exists {
		tdb.LastAccess = time.Now()
		return tdb, nil
	}

	// Download and decrypt database
	localPath := tm.getTenantLocalPath(tenantID)
	encryptedPath := localPath + ".encrypted"

	// Check if exists in S3
	exists, err := tm.storage.TenantDBExists(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant existence: %w", err)
	}

	if exists {
		// Download encrypted database
		if err := tm.storage.DownloadTenantDB(ctx, tenantID, encryptedPath); err != nil {
			return nil, fmt.Errorf("failed to download database: %w", err)
		}

		// Decrypt database
		if err := DecryptFile(encryptedPath, localPath, password); err != nil {
			os.Remove(encryptedPath)
			return nil, fmt.Errorf("failed to decrypt database: %w", err)
		}

		// Clean up encrypted file
		os.Remove(encryptedPath)
	} else {
		// Create new database
		if err := tm.createNewTenantDB(localPath); err != nil {
			return nil, fmt.Errorf("failed to create new database: %w", err)
		}
	}

	// Open database connection
	db, err := sql.Open("sqlite3", localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tenant DB entry
	tdb := &TenantDB{
		TenantID:   tenantID,
		DB:         db,
		LocalPath:  localPath,
		LastSync:   time.Now(),
		LastAccess: time.Now(),
		Password:   password,
	}

	// Add to cache
	tm.cache[tenantID] = tdb

	return tdb, nil
}

// SyncTenantDB syncs a tenant database to S3
func (tm *TenantManager) SyncTenantDB(ctx context.Context, tenantID string) error {
	tm.cacheMutex.RLock()
	tdb, exists := tm.cache[tenantID]
	tm.cacheMutex.RUnlock()

	if !exists {
		return fmt.Errorf("tenant %s not loaded", tenantID)
	}

	tdb.mutex.Lock()
	defer tdb.mutex.Unlock()

	// Close database temporarily for encryption
	if tdb.DB != nil {
		tdb.DB.Close()
	}

	// Encrypt database
	encryptedPath := tdb.LocalPath + ".encrypted"
	if err := EncryptFile(tdb.LocalPath, encryptedPath, tdb.Password); err != nil {
		// Reopen database
		tdb.DB, _ = sql.Open("sqlite3", tdb.LocalPath)
		return fmt.Errorf("failed to encrypt database: %w", err)
	}

	// Upload to S3
	if err := tm.storage.UploadTenantDB(ctx, tenantID, encryptedPath); err != nil {
		os.Remove(encryptedPath)
		tdb.DB, _ = sql.Open("sqlite3", tdb.LocalPath)
		return fmt.Errorf("failed to upload database: %w", err)
	}

	// Clean up encrypted file
	os.Remove(encryptedPath)

	// Reopen database
	var err error
	tdb.DB, err = sql.Open("sqlite3", tdb.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to reopen database: %w", err)
	}

	tdb.LastSync = time.Now()

	return nil
}

// CloseTenantDB closes a tenant database and syncs to S3
func (tm *TenantManager) CloseTenantDB(ctx context.Context, tenantID string) error {
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()

	tdb, exists := tm.cache[tenantID]
	if !exists {
		return nil // Already closed
	}

	// Sync to S3
	if err := tm.SyncTenantDB(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to sync database: %w", err)
	}

	// Close database
	if tdb.DB != nil {
		tdb.DB.Close()
	}

	// Remove from cache
	delete(tm.cache, tenantID)

	// Clean up local file
	os.Remove(tdb.LocalPath)

	return nil
}

// CloseAll closes all tenant databases and syncs them
func (tm *TenantManager) CloseAll(ctx context.Context) error {
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()

	var errors []error

	for tenantID, tdb := range tm.cache {
		// Sync to S3
		if err := tm.SyncTenantDB(ctx, tenantID); err != nil {
			errors = append(errors, fmt.Errorf("failed to sync %s: %w", tenantID, err))
			continue
		}

		// Close database
		if tdb.DB != nil {
			tdb.DB.Close()
		}

		// Clean up local file
		os.Remove(tdb.LocalPath)
	}

	// Clear cache
	tm.cache = make(map[string]*TenantDB)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// autoSyncWorker periodically syncs all cached databases
func (tm *TenantManager) autoSyncWorker() {
	ticker := time.NewTicker(tm.syncInterval)
	defer ticker.Stop()

	for range ticker.C {
		tm.cacheMutex.RLock()
		tenants := make([]string, 0, len(tm.cache))
		for tenantID := range tm.cache {
			tenants = append(tenants, tenantID)
		}
		tm.cacheMutex.RUnlock()

		ctx := context.Background()
		for _, tenantID := range tenants {
			if err := tm.SyncTenantDB(ctx, tenantID); err != nil {
				// Log error (in production, use proper logging)
				fmt.Printf("auto-sync error for tenant %s: %v\n", tenantID, err)
			}
		}
	}
}

// getTenantLocalPath returns the local path for a tenant database
func (tm *TenantManager) getTenantLocalPath(tenantID string) string {
	return filepath.Join(tm.localDir, fmt.Sprintf("%s.db", tenantID))
}

// createNewTenantDB creates a new empty tenant database with schema
func (tm *TenantManager) createNewTenantDB(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run schema creation (same as CLI schema)
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

	CREATE INDEX IF NOT EXISTS idx_contracts_client ON contracts(client_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_company ON invoices(company_id);
	CREATE INDEX IF NOT EXISTS idx_tracking_sessions_client ON tracking_sessions(client_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}
