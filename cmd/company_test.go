package cmd

import (
	"testing"

	"github.com/Andriiklymiuk/ung/internal/db"
)

func TestCompanyCRUD(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Test Create - using only columns that exist in the schema
	result, err := db.DB.Exec(`
		INSERT INTO companies (name, email, phone, address, tax_id)
		VALUES ('Test Company', 'company@test.com', '+1234567890', '123 Business St', 'TAX123456')
	`)
	if err != nil {
		t.Fatalf("Failed to add company: %v", err)
	}

	id, _ := result.LastInsertId()
	if id < 1 {
		t.Errorf("Expected company ID >= 1, got %d", id)
	}

	// Test Read
	var name, email string
	var phone, address, taxID *string
	err = db.DB.QueryRow(`
		SELECT name, email, phone, address, tax_id FROM companies WHERE id = ?
	`, id).Scan(&name, &email, &phone, &address, &taxID)
	if err != nil {
		t.Fatalf("Failed to query company: %v", err)
	}

	if name != "Test Company" {
		t.Errorf("Expected name 'Test Company', got '%s'", name)
	}
	if email != "company@test.com" {
		t.Errorf("Expected email 'company@test.com', got '%s'", email)
	}

	// Test Update
	_, err = db.DB.Exec("UPDATE companies SET name = 'Updated Corp', email = 'updated@corp.com' WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to update company: %v", err)
	}

	err = db.DB.QueryRow("SELECT name, email FROM companies WHERE id = ?", id).Scan(&name, &email)
	if err != nil {
		t.Fatalf("Failed to query updated company: %v", err)
	}
	if name != "Updated Corp" {
		t.Errorf("Expected name 'Updated Corp', got '%s'", name)
	}

	// Test Delete
	_, err = db.DB.Exec("DELETE FROM companies WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to delete company: %v", err)
	}

	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM companies WHERE id = ?", id).Scan(&count)
	if count != 0 {
		t.Errorf("Expected company to be deleted")
	}
}

func TestCompanyMultiple(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get initial count
	var initialCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM companies").Scan(&initialCount)

	// Add companies
	for i := 0; i < 3; i++ {
		_, err := db.DB.Exec("INSERT INTO companies (name, email) VALUES (?, ?)",
			"MultiTest Company "+string(rune('A'+i)),
			"multitest"+string(rune('a'+i))+"@test.com")
		if err != nil {
			t.Fatalf("Failed to add company: %v", err)
		}
	}

	// Verify count increased
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM companies").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count companies: %v", err)
	}
	if count != initialCount+3 {
		t.Errorf("Expected %d companies, got %d", initialCount+3, count)
	}
}

func TestCompanyRequiredFields(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Name and email are required
	result, err := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Minimal Company', 'minimal@test.com')")
	if err != nil {
		t.Fatalf("Failed to add company with minimal fields: %v", err)
	}

	id, _ := result.LastInsertId()

	// Verify
	var name, email string
	err = db.DB.QueryRow("SELECT name, email FROM companies WHERE id = ?", id).Scan(&name, &email)
	if err != nil {
		t.Fatalf("Failed to query company: %v", err)
	}

	if name != "Minimal Company" {
		t.Errorf("Expected name 'Minimal Company', got '%s'", name)
	}
}

func TestCompanyOptionalFields(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Insert with all optional fields
	result, err := db.DB.Exec(`
		INSERT INTO companies (name, email, phone, address, tax_id)
		VALUES ('Full Company', 'full@test.com', '+1-555-1234', '789 Full Ave', 'FULL-TAX-123')
	`)
	if err != nil {
		t.Fatalf("Failed to add company: %v", err)
	}

	id, _ := result.LastInsertId()

	var phone, address, taxID *string
	err = db.DB.QueryRow("SELECT phone, address, tax_id FROM companies WHERE id = ?", id).
		Scan(&phone, &address, &taxID)
	if err != nil {
		t.Fatalf("Failed to query company: %v", err)
	}

	if phone == nil || *phone != "+1-555-1234" {
		t.Errorf("Expected phone '+1-555-1234', got %v", phone)
	}
	if address == nil || *address != "789 Full Ave" {
		t.Errorf("Expected address '789 Full Ave', got %v", address)
	}
}
