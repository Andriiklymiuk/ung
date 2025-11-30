package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
)

func TestInvoiceCRUD(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create company first
	companyResult, err := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Invoice Test Company', 'invoicetest@test.com')")
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}
	companyID, _ := companyResult.LastInsertId()

	// Test Create - using correct column names from schema
	issueDate := time.Now()
	dueDate := issueDate.AddDate(0, 0, 30)

	result, err := db.DB.Exec(`
		INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
		VALUES (?, 'INV-TEST-001', 1500.00, 'USD', 'pending', ?, ?)
	`, companyID, issueDate, dueDate)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	id, _ := result.LastInsertId()
	if id < 1 {
		t.Errorf("Expected invoice ID >= 1, got %d", id)
	}

	// Test Read
	var invoiceNum, currency, status string
	var amount float64
	err = db.DB.QueryRow(`
		SELECT invoice_num, amount, currency, status FROM invoices WHERE id = ?
	`, id).Scan(&invoiceNum, &amount, &currency, &status)
	if err != nil {
		t.Fatalf("Failed to query invoice: %v", err)
	}

	if invoiceNum != "INV-TEST-001" {
		t.Errorf("Expected invoice_num 'INV-TEST-001', got '%s'", invoiceNum)
	}
	if amount != 1500.00 {
		t.Errorf("Expected amount 1500.00, got %f", amount)
	}

	// Test Delete
	_, err = db.DB.Exec("DELETE FROM invoices WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to delete invoice: %v", err)
	}

	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM invoices WHERE id = ?", id).Scan(&count)
	if count != 0 {
		t.Errorf("Expected invoice to be deleted")
	}
}

func TestInvoiceStatuses(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Status Test Co', 'statustest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	statuses := []string{"pending", "sent", "paid", "overdue"}

	for i, status := range statuses {
		result, err := db.DB.Exec(`
			INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
			VALUES (?, ?, 100.00, 'USD', ?, ?, ?)
		`, companyID, "INV-STATUS-"+string(rune('0'+i)), status, time.Now(), time.Now().AddDate(0, 0, 30))
		if err != nil {
			t.Errorf("Failed to create invoice with status '%s': %v", status, err)
			continue
		}

		id, _ := result.LastInsertId()

		var storedStatus string
		err = db.DB.QueryRow("SELECT status FROM invoices WHERE id = ?", id).Scan(&storedStatus)
		if err != nil {
			t.Errorf("Failed to query invoice: %v", err)
			continue
		}

		if storedStatus != status {
			t.Errorf("Expected status '%s', got '%s'", status, storedStatus)
		}
	}
}

func TestInvoiceStatusUpdate(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Update Test Co', 'updatetest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	// Create pending invoice
	result, err := db.DB.Exec(`
		INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
		VALUES (?, 'INV-UPDATE-001', 500.00, 'USD', 'pending', ?, ?)
	`, companyID, time.Now(), time.Now().AddDate(0, 0, 30))
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	id, _ := result.LastInsertId()

	// Update to paid
	_, err = db.DB.Exec("UPDATE invoices SET status = 'paid' WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to update invoice status: %v", err)
	}

	// Verify
	var status string
	err = db.DB.QueryRow("SELECT status FROM invoices WHERE id = ?", id).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to query invoice: %v", err)
	}

	if status != "paid" {
		t.Errorf("Expected status 'paid', got '%s'", status)
	}
}

func TestInvoiceWithLineItems(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('LineItem Test Co', 'lineitemtest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	// Create invoice
	invoiceResult, err := db.DB.Exec(`
		INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
		VALUES (?, 'INV-ITEMS-001', 1000.00, 'USD', 'pending', ?, ?)
	`, companyID, time.Now(), time.Now().AddDate(0, 0, 30))
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	invoiceID, _ := invoiceResult.LastInsertId()

	// Add line items using correct column names
	lineItems := []struct {
		itemName    string
		description string
		quantity    float64
		rate        float64
	}{
		{"Web Development", "Frontend work", 10, 50.00},
		{"Design Work", "UI/UX design", 5, 75.00},
		{"Consulting", "Technical consulting", 2, 125.00},
	}

	for _, item := range lineItems {
		_, err := db.DB.Exec(`
			INSERT INTO invoice_line_items (invoice_id, item_name, description, quantity, rate, amount)
			VALUES (?, ?, ?, ?, ?, ?)
		`, invoiceID, item.itemName, item.description, item.quantity, item.rate, item.quantity*item.rate)
		if err != nil {
			t.Errorf("Failed to add line item '%s': %v", item.itemName, err)
		}
	}

	// Verify line items count
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM invoice_line_items WHERE invoice_id = ?", invoiceID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count line items: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 line items, got %d", count)
	}

	// Calculate total from line items
	var total float64
	err = db.DB.QueryRow("SELECT SUM(amount) FROM invoice_line_items WHERE invoice_id = ?", invoiceID).Scan(&total)
	if err != nil {
		t.Fatalf("Failed to sum line items: %v", err)
	}

	expected := (10 * 50) + (5 * 75) + (2 * 125) // 500 + 375 + 250 = 1125
	if total != float64(expected) {
		t.Errorf("Expected total %d, got %f", expected, total)
	}
}

func TestInvoiceCurrencies(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Currency Test Co', 'currencytest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	currencies := []string{"USD", "EUR", "GBP", "CAD", "AUD", "UAH", "PLN"}

	for i, currency := range currencies {
		result, err := db.DB.Exec(`
			INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
			VALUES (?, ?, 100.00, ?, 'pending', ?, ?)
		`, companyID, "INV-CUR-"+currency+"-"+string(rune('0'+i)), currency, time.Now(), time.Now().AddDate(0, 0, 30))
		if err != nil {
			t.Errorf("Failed to create invoice with currency '%s': %v", currency, err)
			continue
		}

		id, _ := result.LastInsertId()

		var storedCurrency string
		err = db.DB.QueryRow("SELECT currency FROM invoices WHERE id = ?", id).Scan(&storedCurrency)
		if err != nil {
			t.Errorf("Failed to query invoice: %v", err)
			continue
		}

		if storedCurrency != currency {
			t.Errorf("Expected currency '%s', got '%s'", currency, storedCurrency)
		}
	}
}

func TestInvoiceRecipients(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Setup
	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Recipient Test Co', 'recipienttest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	clientResult1, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES ('Recipient Client A', 'reca@test.com')")
	clientID1, _ := clientResult1.LastInsertId()

	clientResult2, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES ('Recipient Client B', 'recb@test.com')")
	clientID2, _ := clientResult2.LastInsertId()

	// Create invoice
	invoiceResult, err := db.DB.Exec(`
		INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
		VALUES (?, 'INV-REC-001', 500.00, 'USD', 'pending', ?, ?)
	`, companyID, time.Now(), time.Now().AddDate(0, 0, 30))
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	invoiceID, _ := invoiceResult.LastInsertId()

	// Add recipients
	_, err = db.DB.Exec("INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)", invoiceID, clientID1)
	if err != nil {
		t.Fatalf("Failed to add recipient 1: %v", err)
	}
	_, err = db.DB.Exec("INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)", invoiceID, clientID2)
	if err != nil {
		t.Fatalf("Failed to add recipient 2: %v", err)
	}

	// Query recipients
	var recipientCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM invoice_recipients WHERE invoice_id = ?", invoiceID).Scan(&recipientCount)
	if err != nil {
		t.Fatalf("Failed to count recipients: %v", err)
	}

	if recipientCount != 2 {
		t.Errorf("Expected 2 recipients, got %d", recipientCount)
	}
}
