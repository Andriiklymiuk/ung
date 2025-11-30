package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
)

func TestExpenseCRUD(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Test Create
	result, err := db.DB.Exec(`
		INSERT INTO expenses (description, amount, currency, category, date, vendor, notes)
		VALUES ('Software License', 99.99, 'USD', 'software', ?, 'Adobe', 'Creative Cloud subscription')
	`, time.Now())
	if err != nil {
		t.Fatalf("Failed to add expense: %v", err)
	}

	id, _ := result.LastInsertId()
	if id < 1 {
		t.Errorf("Expected expense ID >= 1, got %d", id)
	}

	// Test Read
	var description, currency, category string
	var amount float64
	var vendor, notes *string
	err = db.DB.QueryRow(`
		SELECT description, amount, currency, category, vendor, notes
		FROM expenses WHERE id = ?
	`, id).Scan(&description, &amount, &currency, &category, &vendor, &notes)
	if err != nil {
		t.Fatalf("Failed to query expense: %v", err)
	}

	if description != "Software License" {
		t.Errorf("Expected description 'Software License', got '%s'", description)
	}
	if amount != 99.99 {
		t.Errorf("Expected amount 99.99, got %f", amount)
	}

	// Test Delete
	_, err = db.DB.Exec("DELETE FROM expenses WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to delete expense: %v", err)
	}

	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM expenses WHERE id = ?", id).Scan(&count)
	if count != 0 {
		t.Errorf("Expected expense to be deleted")
	}
}

func TestExpenseCategories(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	categories := []string{
		"software",
		"hardware",
		"travel",
		"meals",
		"office_supplies",
		"utilities",
		"marketing",
		"other",
	}

	for _, cat := range categories {
		result, err := db.DB.Exec(`
			INSERT INTO expenses (description, amount, currency, category, date)
			VALUES (?, 50.00, 'USD', ?, ?)
		`, cat+" expense", cat, time.Now())
		if err != nil {
			t.Errorf("Failed to insert expense with category '%s': %v", cat, err)
			continue
		}

		id, _ := result.LastInsertId()

		// Verify category
		var storedCategory string
		err = db.DB.QueryRow("SELECT category FROM expenses WHERE id = ?", id).Scan(&storedCategory)
		if err != nil {
			t.Errorf("Failed to query expense: %v", err)
			continue
		}
		if storedCategory != cat {
			t.Errorf("Expected category '%s', got '%s'", cat, storedCategory)
		}
	}
}

func TestExpenseReportByCategory(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline totals
	var baseSoftware, baseHardware float64
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = 'software'").Scan(&baseSoftware)
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = 'hardware'").Scan(&baseHardware)

	// Add expenses in different categories
	testData := []struct {
		category string
		amount   float64
	}{
		{"software", 100.00},
		{"software", 50.00},
		{"hardware", 200.00},
	}

	for _, td := range testData {
		_, err := db.DB.Exec(`
			INSERT INTO expenses (description, amount, currency, category, date)
			VALUES ('Test', ?, 'USD', ?, ?)
		`, td.amount, td.category, time.Now())
		if err != nil {
			t.Fatalf("Failed to insert expense: %v", err)
		}
	}

	// Query category totals
	var softwareTotal, hardwareTotal float64
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = 'software'").Scan(&softwareTotal)
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = 'hardware'").Scan(&hardwareTotal)

	// Verify increases
	expectedSoftware := baseSoftware + 150.00
	expectedHardware := baseHardware + 200.00

	if softwareTotal != expectedSoftware {
		t.Errorf("Expected software total %f, got %f", expectedSoftware, softwareTotal)
	}
	if hardwareTotal != expectedHardware {
		t.Errorf("Expected hardware total %f, got %f", expectedHardware, hardwareTotal)
	}
}

func TestExpenseAmountFormats(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	amounts := []float64{
		0.01,
		99.99,
		1000.50,
		12345.67,
	}

	for _, amt := range amounts {
		result, err := db.DB.Exec(`
			INSERT INTO expenses (description, amount, currency, category, date)
			VALUES ('Amount Test', ?, 'USD', 'other', ?)
		`, amt, time.Now())
		if err != nil {
			t.Errorf("Failed to insert amount %f: %v", amt, err)
			continue
		}

		id, _ := result.LastInsertId()

		var storedAmount float64
		err = db.DB.QueryRow("SELECT amount FROM expenses WHERE id = ?", id).Scan(&storedAmount)
		if err != nil {
			t.Errorf("Failed to query expense: %v", err)
			continue
		}

		if storedAmount != amt {
			t.Errorf("Expected amount %f, got %f", amt, storedAmount)
		}
	}
}

func TestExpenseWithVendor(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	vendors := []string{
		"Adobe",
		"Microsoft",
		"Amazon Web Services",
	}

	for _, vendor := range vendors {
		result, err := db.DB.Exec(`
			INSERT INTO expenses (description, amount, currency, category, date, vendor)
			VALUES ('Vendor Test', 50.00, 'USD', 'software', ?, ?)
		`, time.Now(), vendor)
		if err != nil {
			t.Errorf("Failed to insert expense with vendor '%s': %v", vendor, err)
			continue
		}

		id, _ := result.LastInsertId()

		var storedVendor *string
		err = db.DB.QueryRow("SELECT vendor FROM expenses WHERE id = ?", id).Scan(&storedVendor)
		if err != nil {
			t.Errorf("Failed to query expense: %v", err)
			continue
		}

		if storedVendor == nil || *storedVendor != vendor {
			t.Errorf("Expected vendor '%s', got %v", vendor, storedVendor)
		}
	}
}
