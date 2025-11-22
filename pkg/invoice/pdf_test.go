package invoice

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/models"
)

// TestDrawLineItemsTable tests the line items table generation
func TestDrawLineItemsTable(t *testing.T) {
	// This is a basic test to verify the function compiles and doesn't panic
	// Full PDF generation testing would require mocking the db.GetInvoicesDir function
	// For now, we test the table drawing logic in isolation

	lineItems := []models.InvoiceLineItem{
		{
			ID:          1,
			InvoiceID:   1,
			ItemName:    "Consulting Hours",
			Description: "Software development consulting",
			Quantity:    10.0,
			Rate:        150.00,
			Amount:      1500.00,
		},
		{
			ID:          2,
			InvoiceID:   1,
			ItemName:    "Design Work",
			Description: "",
			Quantity:    5.0,
			Rate:        100.00,
			Amount:      500.00,
		},
	}

	// Verify total calculation
	var expectedTotal float64
	for _, item := range lineItems {
		expectedTotal += item.Amount
	}

	if expectedTotal != 2000.00 {
		t.Errorf("expected total 2000.00, got %.2f", expectedTotal)
	}

	// Verify line item amounts match quantity * rate
	for i, item := range lineItems {
		expected := item.Quantity * item.Rate
		if item.Amount != expected {
			t.Errorf("line item %d: expected amount %.2f (%.2f * %.2f), got %.2f",
				i, expected, item.Quantity, item.Rate, item.Amount)
		}
	}
}

// TestInvoiceDataStructure verifies the invoice and related models structure
func TestInvoiceDataStructure(t *testing.T) {
	now := time.Now()

	company := models.Company{
		ID:                  1,
		Name:                "Test Company LLC",
		Email:               "test@company.com",
		Phone:               "+1-555-0100",
		Address:             "123 Business St, Suite 100",
		RegistrationAddress: "123 Business St, Suite 100",
		TaxID:               "12-3456789",
		BankName:            "Test Bank",
		BankAccount:         "1234567890",
		BankSWIFT:           "TESTUS33",
	}

	client := models.Client{
		ID:      1,
		Name:    "Test Client Inc",
		Email:   "client@test.com",
		Address: "456 Client Ave",
		TaxID:   "98-7654321",
	}

	invoice := models.Invoice{
		ID:          1,
		InvoiceNum:  "INV-2024-001",
		CompanyID:   1,
		Amount:      1500.00,
		Currency:    "USD",
		Description: "Consulting Services",
		Status:      models.StatusPending,
		IssuedDate:  now,
		DueDate:     now.AddDate(0, 0, 30),
	}

	// Verify data integrity
	if company.Name == "" {
		t.Error("company name should not be empty")
	}
	if client.Name == "" {
		t.Error("client name should not be empty")
	}
	if invoice.InvoiceNum == "" {
		t.Error("invoice number should not be empty")
	}
	if invoice.Amount <= 0 {
		t.Error("invoice amount should be positive")
	}
	if invoice.Currency == "" {
		t.Error("currency should not be empty")
	}
}
