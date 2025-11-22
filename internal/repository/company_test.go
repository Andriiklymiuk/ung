package repository

import (
	"testing"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	var err error
	// Create a new in-memory database for each test
	db.GormDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to setup test database: %v", err)
	}

	// Run migrations
	err = db.GormDB.AutoMigrate(
		&models.Company{},
		&models.Client{},
		&models.Contract{},
		&models.Invoice{},
		&models.InvoiceLineItem{},
		&models.TrackingSession{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Create invoice_recipients table manually (join table)
	db.GormDB.Exec(`
		CREATE TABLE IF NOT EXISTS invoice_recipients (
			invoice_id INTEGER NOT NULL,
			client_id INTEGER NOT NULL,
			PRIMARY KEY (invoice_id, client_id)
		)
	`)
}

func TestCompanyRepository_Create(t *testing.T) {
	setupTestDB(t)
	repo := NewCompanyRepository()

	company := &models.Company{
		Name:    "Test Company",
		Email:   "test@example.com",
		Address: "123 Test St",
		TaxID:   "12345",
	}

	err := repo.Create(company)
	if err != nil {
		t.Fatalf("failed to create company: %v", err)
	}

	if company.ID == 0 {
		t.Error("expected company ID to be set")
	}
}

func TestCompanyRepository_GetByID(t *testing.T) {
	setupTestDB(t)
	repo := NewCompanyRepository()

	// Create test company
	company := &models.Company{
		Name:  "Test Company",
		Email: "test@example.com",
	}
	repo.Create(company)

	// Retrieve by ID
	retrieved, err := repo.GetByID(company.ID)
	if err != nil {
		t.Fatalf("failed to get company: %v", err)
	}

	if retrieved.Name != company.Name {
		t.Errorf("expected name %s, got %s", company.Name, retrieved.Name)
	}
}

func TestCompanyRepository_List(t *testing.T) {
	setupTestDB(t)
	repo := NewCompanyRepository()

	// Create test companies
	companies := []*models.Company{
		{Name: "Company 1", Email: "c1@example.com"},
		{Name: "Company 2", Email: "c2@example.com"},
		{Name: "Company 3", Email: "c3@example.com"},
	}

	for _, c := range companies {
		repo.Create(c)
	}

	// List all
	result, err := repo.List()
	if err != nil {
		t.Fatalf("failed to list companies: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 companies, got %d", len(result))
	}
}

func TestCompanyRepository_Update(t *testing.T) {
	setupTestDB(t)
	repo := NewCompanyRepository()

	// Create test company
	company := &models.Company{
		Name:  "Original Name",
		Email: "test@example.com",
	}
	repo.Create(company)

	// Update
	company.Name = "Updated Name"
	err := repo.Update(company)
	if err != nil {
		t.Fatalf("failed to update company: %v", err)
	}

	// Verify
	updated, _ := repo.GetByID(company.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", updated.Name)
	}
}

func TestCompanyRepository_Delete(t *testing.T) {
	setupTestDB(t)
	repo := NewCompanyRepository()

	// Create test company
	company := &models.Company{
		Name:  "Test Company",
		Email: "test@example.com",
	}
	repo.Create(company)

	// Delete
	err := repo.Delete(company.ID)
	if err != nil {
		t.Fatalf("failed to delete company: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(company.ID)
	if err == nil {
		t.Error("expected error when getting deleted company")
	}
}

func TestCompanyRepository_Count(t *testing.T) {
	setupTestDB(t)
	repo := NewCompanyRepository()

	// Initially empty
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("failed to count companies: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	// Add companies
	repo.Create(&models.Company{Name: "Company 1", Email: "c1@example.com"})
	repo.Create(&models.Company{Name: "Company 2", Email: "c2@example.com"})

	count, err = repo.Count()
	if err != nil {
		t.Fatalf("failed to count companies: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}
