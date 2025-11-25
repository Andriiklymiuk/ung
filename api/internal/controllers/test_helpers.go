package controllers

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Client{},
		&models.Company{},
		&models.Contract{},
		&models.Invoice{},
		&models.InvoiceLineItem{},
		&models.InvoiceRecipient{},
		&models.Expense{},
		&models.TrackingSession{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// WithTenantDB adds a tenant database to the context
func WithTenantDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, middleware.TestTenantDBKey, db)
}

// GetTenantDBFromContext retrieves the tenant database from context (for testing)
func GetTenantDBFromContext(ctx context.Context) *gorm.DB {
	if db, ok := ctx.Value(middleware.TestTenantDBKey).(*gorm.DB); ok {
		return db
	}
	return nil
}

// DecodeStandardResponse decodes a StandardResponse and extracts the data
// This is a helper for tests since all API responses are wrapped in StandardResponse
func DecodeStandardResponse(t *testing.T, reader io.Reader, target interface{}) {
	// Read all bytes from the reader
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	var stdResp models.StandardResponse
	if err := json.Unmarshal(bodyBytes, &stdResp); err != nil {
		t.Fatalf("Failed to decode StandardResponse: %v", err)
	}

	// Then extract and decode the data field
	dataBytes, err := json.Marshal(stdResp.Data)
	if err != nil {
		t.Fatalf("Failed to marshal data field: %v", err)
	}

	if err := json.Unmarshal(dataBytes, target); err != nil {
		t.Fatalf("Failed to decode data into target: %v", err)
	}
}
