package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ung/api/internal/models"
)

// InitAPIDatabase initializes the main API database
func InitAPIDatabase(dbPath string) (*gorm.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate API models
	if err := db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.APIKey{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// InitUserDatabase creates or opens a user's database
func InitUserDatabase(userDBPath string) (*gorm.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(userDBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(userDBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open user database: %w", err)
	}

	// Note: User database schema is managed by CLI migrations
	// The API just opens and uses it

	return db, nil
}

// CreateUserDatabase creates a new user database with migrations
func CreateUserDatabase(userID uint, userDataDir string) (string, error) {
	userDir := filepath.Join(userDataDir, fmt.Sprintf("user_%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	dbPath := filepath.Join(userDir, "ung.db")

	// Initialize database
	db, err := InitUserDatabase(dbPath)
	if err != nil {
		return "", err
	}

	// Run migrations (reuse CLI migrations)
	// For now, we'll use AutoMigrate with UNG models
	// In production, this should run actual migration files
	if err := runUserMigrations(db); err != nil {
		return "", fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbPath, nil
}

// runUserMigrations runs migrations on user database
func runUserMigrations(db *gorm.DB) error {
	// Import UNG CLI models and run migrations
	// For now, this is a placeholder
	// In production, you'd exec the actual migration SQL files

	// This would be replaced with proper migration runner
	// that executes ../migrations/*.up.sql files

	return nil
}
