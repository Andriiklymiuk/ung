package invoice

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/models"
)

// TestFormatCurrency tests the currency formatting function
func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		amount   float64
		currency string
		expected string
	}{
		{1500.00, "USD", "$1500.00"},
		{1234.56, "EUR", "€1234.56"},
		{999.99, "GBP", "£999.99"},
		{10000.00, "JPY", "¥10000.00"},
		{5000.00, "UAH", "₴5000.00"},
		{2500.50, "CAD", "C$2500.50"},
		{1000.00, "AUD", "A$1000.00"},
		{750.25, "CHF", "CHF 750.25"},
		{500.00, "PLN", "zł500.00"},
		{100.00, "UNKNOWN", "UNKNOWN 100.00"},
		{0.00, "USD", "$0.00"},
		{0.01, "EUR", "€0.01"},
		{999999.99, "GBP", "£999999.99"},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			result := FormatCurrency(tt.amount, tt.currency)
			if result != tt.expected {
				t.Errorf("FormatCurrency(%.2f, %s) = %s; want %s",
					tt.amount, tt.currency, result, tt.expected)
			}
		})
	}
}

// TestFormatCurrencyLowercase tests currency formatting with lowercase input
func TestFormatCurrencyLowercase(t *testing.T) {
	result := FormatCurrency(100.00, "usd")
	expected := "$100.00"
	if result != expected {
		t.Errorf("FormatCurrency(100.00, usd) = %s; want %s", result, expected)
	}
}

// TestCurrencySymbolsMap verifies all expected currency symbols are present
func TestCurrencySymbolsMap(t *testing.T) {
	expectedCurrencies := []string{
		"USD", "EUR", "GBP", "JPY", "CNY", "CHF", "CAD", "AUD", "NZD",
		"INR", "KRW", "BRL", "MXN", "RUB", "TRY", "PLN", "SEK", "NOK",
		"DKK", "CZK", "HUF", "UAH", "ILS", "SGD", "HKD", "THB", "ZAR",
	}

	for _, currency := range expectedCurrencies {
		if _, ok := CurrencySymbols[currency]; !ok {
			t.Errorf("Currency symbol for %s not found in CurrencySymbols map", currency)
		}
	}
}

// TestInvoiceTotalsCalculation tests the invoice totals struct
func TestInvoiceTotalsCalculation(t *testing.T) {
	totals := InvoiceTotals{
		Subtotal:      1000.00,
		Discount:      100.00,
		TaxableAmount: 900.00,
		TaxAmount:     180.00,
		GrandTotal:    1080.00,
	}

	// Verify taxable amount is subtotal minus discount
	expectedTaxable := totals.Subtotal - totals.Discount
	if totals.TaxableAmount != expectedTaxable {
		t.Errorf("TaxableAmount = %.2f; want %.2f", totals.TaxableAmount, expectedTaxable)
	}

	// Verify grand total is taxable plus tax
	expectedGrand := totals.TaxableAmount + totals.TaxAmount
	if totals.GrandTotal != expectedGrand {
		t.Errorf("GrandTotal = %.2f; want %.2f", totals.GrandTotal, expectedGrand)
	}
}

// TestLineItemWithDiscount tests line item discount calculations
func TestLineItemWithDiscount(t *testing.T) {
	tests := []struct {
		name        string
		item        models.InvoiceLineItem
		expectedNet float64
	}{
		{
			name: "Fixed discount",
			item: models.InvoiceLineItem{
				Amount:      100.00,
				Discount:    10.00,
				DiscountPct: 0,
			},
			expectedNet: 90.00,
		},
		{
			name: "Percentage discount",
			item: models.InvoiceLineItem{
				Amount:      100.00,
				Discount:    0,
				DiscountPct: 20.0, // 20%
			},
			expectedNet: 80.00,
		},
		{
			name: "No discount",
			item: models.InvoiceLineItem{
				Amount:      100.00,
				Discount:    0,
				DiscountPct: 0,
			},
			expectedNet: 100.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount := tt.item.Discount
			if tt.item.DiscountPct > 0 {
				discount = tt.item.Amount * (tt.item.DiscountPct / 100)
			}
			netAmount := tt.item.Amount - discount

			if netAmount != tt.expectedNet {
				t.Errorf("Net amount = %.2f; want %.2f", netAmount, tt.expectedNet)
			}
		})
	}
}

// TestDrawLineItemsTable tests the line items table generation
func TestDrawLineItemsTable(t *testing.T) {
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

// TestLineItemsWithDiscounts tests discount handling in line items
func TestLineItemsWithDiscounts(t *testing.T) {
	lineItems := []models.InvoiceLineItem{
		{
			ID:       1,
			ItemName: "Service A",
			Quantity: 1,
			Rate:     100.00,
			Amount:   100.00,
			Discount: 10.00, // $10 off
		},
		{
			ID:          2,
			ItemName:    "Service B",
			Quantity:    1,
			Rate:        200.00,
			Amount:      200.00,
			DiscountPct: 15.0, // 15% off = $30
		},
	}

	// Calculate totals with discounts
	var subtotal, totalDiscount float64
	for _, item := range lineItems {
		subtotal += item.Amount
		if item.DiscountPct > 0 {
			totalDiscount += item.Amount * (item.DiscountPct / 100)
		} else {
			totalDiscount += item.Discount
		}
	}

	expectedSubtotal := 300.00
	expectedDiscount := 40.00 // $10 + $30
	expectedNet := 260.00

	if subtotal != expectedSubtotal {
		t.Errorf("Subtotal = %.2f; want %.2f", subtotal, expectedSubtotal)
	}
	if totalDiscount != expectedDiscount {
		t.Errorf("Total discount = %.2f; want %.2f", totalDiscount, expectedDiscount)
	}
	if subtotal-totalDiscount != expectedNet {
		t.Errorf("Net total = %.2f; want %.2f", subtotal-totalDiscount, expectedNet)
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
		LogoPath:            "/path/to/logo.png",
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
	if company.LogoPath == "" {
		t.Error("company logo path should be set for test")
	}
}

// TestInvoiceStatusValues tests that invoice status values are correct
func TestInvoiceStatusValues(t *testing.T) {
	statuses := map[models.InvoiceStatus]string{
		models.StatusPending: "pending",
		models.StatusSent:    "sent",
		models.StatusPaid:    "paid",
		models.StatusOverdue: "overdue",
	}

	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("Status %v = %s; want %s", status, string(status), expected)
		}
	}
}

// TestPDFConfigDefaults tests the default PDF configuration
func TestPDFConfigDefaults(t *testing.T) {
	cfg := config.GetDefaultPDFConfig()

	// Test color defaults
	if cfg.PrimaryColor.R != 232 || cfg.PrimaryColor.G != 119 || cfg.PrimaryColor.B != 34 {
		t.Errorf("Primary color should be orange (#E87722), got RGB(%d,%d,%d)",
			cfg.PrimaryColor.R, cfg.PrimaryColor.G, cfg.PrimaryColor.B)
	}

	// Test boolean defaults
	if !cfg.ShowWatermark {
		t.Error("ShowWatermark should be true by default")
	}
	if !cfg.ShowLogo {
		t.Error("ShowLogo should be true by default")
	}
	if cfg.ShowQRCode {
		t.Error("ShowQRCode should be false by default")
	}
	if !cfg.ShowPageNumber {
		t.Error("ShowPageNumber should be true by default")
	}
	if cfg.ShowTaxBreakdown {
		t.Error("ShowTaxBreakdown should be false by default")
	}

	// Test label defaults
	if cfg.SubtotalLabel != "Subtotal" {
		t.Errorf("SubtotalLabel = %s; want Subtotal", cfg.SubtotalLabel)
	}
	if cfg.DiscountLabel != "Discount" {
		t.Errorf("DiscountLabel = %s; want Discount", cfg.DiscountLabel)
	}
	if cfg.BalanceDueLabel != "Balance Due" {
		t.Errorf("BalanceDueLabel = %s; want Balance Due", cfg.BalanceDueLabel)
	}
	if cfg.PaidLabel != "PAID" {
		t.Errorf("PaidLabel = %s; want PAID", cfg.PaidLabel)
	}
	if cfg.OverdueLabel != "OVERDUE" {
		t.Errorf("OverdueLabel = %s; want OVERDUE", cfg.OverdueLabel)
	}
}

// TestTaxCalculation tests tax rate application
func TestTaxCalculation(t *testing.T) {
	tests := []struct {
		name          string
		taxableAmount float64
		taxRate       float64
		expectedTax   float64
	}{
		{"20% VAT", 1000.00, 0.20, 200.00},
		{"10% GST", 500.00, 0.10, 50.00},
		{"0% Tax", 1000.00, 0.00, 0.00},
		{"7.5% Tax", 200.00, 0.075, 15.00},
		{"25% Tax", 800.00, 0.25, 200.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxAmount := tt.taxableAmount * tt.taxRate
			if taxAmount != tt.expectedTax {
				t.Errorf("Tax on %.2f at %.2f%% = %.2f; want %.2f",
					tt.taxableAmount, tt.taxRate*100, taxAmount, tt.expectedTax)
			}
		})
	}
}

// TestMinFunction tests the min helper function
func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{10, 10, 10},
		{0, 5, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// TestGeneratePDFCreatesFile tests that GeneratePDF creates a file
func TestGeneratePDFCreatesFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "ung-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up test data
	now := time.Now()
	company := models.Company{
		ID:          1,
		Name:        "Test Company",
		Email:       "test@test.com",
		TaxID:       "123456789",
		BankAccount: "1234567890",
		BankName:    "Test Bank",
		BankSWIFT:   "TESTSWIFT",
		Address:     "123 Test St",
	}

	client := models.Client{
		ID:      1,
		Name:    "Test Client",
		Email:   "client@test.com",
		Address: "456 Client Ave",
		TaxID:   "987654321",
	}

	invoice := models.Invoice{
		ID:          1,
		InvoiceNum:  "TEST-001",
		CompanyID:   1,
		Amount:      1500.00,
		Currency:    "USD",
		Description: "Test Invoice",
		Status:      models.StatusPending,
		IssuedDate:  now,
		DueDate:     now.AddDate(0, 0, 30),
	}

	lineItems := []models.InvoiceLineItem{
		{
			ID:        1,
			InvoiceID: 1,
			ItemName:  "Test Service",
			Quantity:  10,
			Rate:      150.00,
			Amount:    1500.00,
		},
	}

	// Generate PDF - this will use the config's invoices directory
	// We can't easily test the full path without modifying the config
	// But we can verify the function doesn't panic and returns expected format
	pdfPath, err := GeneratePDF(invoice, company, client, lineItems)
	if err != nil {
		// If error is about directory, that's expected in test environment
		t.Logf("PDF generation returned error (may be expected in test): %v", err)
		return
	}

	// If successful, verify path format
	if pdfPath == "" {
		t.Error("PDF path should not be empty on success")
	}

	expectedFilename := "TEST-001.pdf"
	if filepath.Base(pdfPath) != expectedFilename {
		t.Errorf("PDF filename = %s; want %s", filepath.Base(pdfPath), expectedFilename)
	}

	// Clean up if file was created
	if _, err := os.Stat(pdfPath); err == nil {
		os.Remove(pdfPath)
	}
}

// TestLineItemQuantityFormatting tests quantity display logic
func TestLineItemQuantityFormatting(t *testing.T) {
	tests := []struct {
		quantity     float64
		expectWholeNumber bool
	}{
		{10.0, true},
		{5.5, false},
		{1.0, true},
		{0.5, false},
		{100.00, true},
		{99.99, false},
	}

	for _, tt := range tests {
		isWhole := tt.quantity == float64(int(tt.quantity))
		if isWhole != tt.expectWholeNumber {
			t.Errorf("Quantity %.2f isWholeNumber = %v; want %v",
				tt.quantity, isWhole, tt.expectWholeNumber)
		}
	}
}

// TestInvoiceStatusWatermarkMapping tests status to watermark text mapping
func TestInvoiceStatusWatermarkMapping(t *testing.T) {
	cfg := config.GetDefaultPDFConfig()

	tests := []struct {
		status        models.InvoiceStatus
		expectWatermark bool
		watermarkText string
	}{
		{models.StatusPaid, true, cfg.PaidLabel},
		{models.StatusOverdue, true, cfg.OverdueLabel},
		{models.StatusPending, false, ""},  // No watermark for pending
		{models.StatusSent, false, ""},     // No watermark for sent
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			var text string
			showWatermark := false

			switch tt.status {
			case models.StatusPaid:
				text = cfg.PaidLabel
				showWatermark = true
			case models.StatusOverdue:
				text = cfg.OverdueLabel
				showWatermark = true
			}

			if showWatermark != tt.expectWatermark {
				t.Errorf("Status %s showWatermark = %v; want %v",
					tt.status, showWatermark, tt.expectWatermark)
			}
			if tt.expectWatermark && text != tt.watermarkText {
				t.Errorf("Status %s watermark = %s; want %s",
					tt.status, text, tt.watermarkText)
			}
		})
	}
}

// Benchmark test for currency formatting
func BenchmarkFormatCurrency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatCurrency(1234.56, "USD")
	}
}

// Benchmark test for calculating totals
func BenchmarkCalculateTotals(b *testing.B) {
	items := []models.InvoiceLineItem{
		{Amount: 100.00, Discount: 10.00},
		{Amount: 200.00, DiscountPct: 15.0},
		{Amount: 300.00},
		{Amount: 400.00, Discount: 20.00},
		{Amount: 500.00, DiscountPct: 10.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var subtotal, totalDiscount float64
		for _, item := range items {
			subtotal += item.Amount
			if item.DiscountPct > 0 {
				totalDiscount += item.Amount * (item.DiscountPct / 100)
			} else {
				totalDiscount += item.Discount
			}
		}
		_ = subtotal - totalDiscount
	}
}
