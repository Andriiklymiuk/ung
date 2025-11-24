package models

import (
	"testing"
	"time"
)

func TestContractType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant ContractType
		expected string
	}{
		{"Hourly", ContractTypeHourly, "hourly"},
		{"Fixed Price", ContractTypeFixedPrice, "fixed_price"},
		{"Retainer", ContractTypeRetainer, "retainer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.constant))
			}
		})
	}
}

func TestInvoiceStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant InvoiceStatus
		expected string
	}{
		{"Pending", StatusPending, "pending"},
		{"Sent", StatusSent, "sent"},
		{"Paid", StatusPaid, "paid"},
		{"Overdue", StatusOverdue, "overdue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.constant))
			}
		})
	}
}

func TestExpenseCategory_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant ExpenseCategory
		expected string
	}{
		{"Software", ExpenseCategorySoftware, "software"},
		{"Hardware", ExpenseCategoryHardware, "hardware"},
		{"Travel", ExpenseCategoryTravel, "travel"},
		{"Meals", ExpenseCategoryMeals, "meals"},
		{"Office Supplies", ExpenseCategoryOfficeSupplies, "office_supplies"},
		{"Utilities", ExpenseCategoryUtilities, "utilities"},
		{"Marketing", ExpenseCategoryMarketing, "marketing"},
		{"Other", ExpenseCategoryOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.constant))
			}
		})
	}
}

func TestCompany_Initialization(t *testing.T) {
	company := Company{
		Name:    "Test Company",
		Email:   "test@company.com",
		Phone:   "123-456-7890",
		Address: "123 Test St",
	}

	if company.Name != "Test Company" {
		t.Errorf("Expected Name 'Test Company', got %s", company.Name)
	}

	if company.Email != "test@company.com" {
		t.Errorf("Expected Email 'test@company.com', got %s", company.Email)
	}
}

func TestClient_Initialization(t *testing.T) {
	client := Client{
		Name:    "Test Client",
		Email:   "client@example.com",
		Address: "456 Client Ave",
		TaxID:   "TAX123",
	}

	if client.Name != "Test Client" {
		t.Errorf("Expected Name 'Test Client', got %s", client.Name)
	}

	if client.Email != "client@example.com" {
		t.Errorf("Expected Email 'client@example.com', got %s", client.Email)
	}

	if client.TaxID != "TAX123" {
		t.Errorf("Expected TaxID 'TAX123', got %s", client.TaxID)
	}
}

func TestContract_HourlyRate(t *testing.T) {
	rate := 150.0
	contract := Contract{
		ContractNum:  "contract.test.jan.2025",
		ClientID:     1,
		Name:         "Test Contract",
		ContractType: ContractTypeHourly,
		HourlyRate:   &rate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       true,
	}

	if contract.ContractType != ContractTypeHourly {
		t.Errorf("Expected ContractType hourly, got %s", contract.ContractType)
	}

	if contract.HourlyRate == nil {
		t.Fatal("HourlyRate should not be nil")
	}

	if *contract.HourlyRate != 150.0 {
		t.Errorf("Expected HourlyRate 150.0, got %f", *contract.HourlyRate)
	}
}

func TestContract_FixedPrice(t *testing.T) {
	price := 5000.0
	contract := Contract{
		ContractNum:  "contract.test.feb.2025",
		ClientID:     1,
		Name:         "Fixed Price Project",
		ContractType: ContractTypeFixedPrice,
		FixedPrice:   &price,
		Currency:     "EUR",
		StartDate:    time.Now(),
		Active:       true,
	}

	if contract.ContractType != ContractTypeFixedPrice {
		t.Errorf("Expected ContractType fixed_price, got %s", contract.ContractType)
	}

	if contract.FixedPrice == nil {
		t.Fatal("FixedPrice should not be nil")
	}

	if *contract.FixedPrice != 5000.0 {
		t.Errorf("Expected FixedPrice 5000.0, got %f", *contract.FixedPrice)
	}
}

func TestInvoice_Initialization(t *testing.T) {
	now := time.Now()
	dueDate := now.AddDate(0, 0, 30)

	invoice := Invoice{
		InvoiceNum:  "inv.test.20250124",
		CompanyID:   1,
		Amount:      1500.50,
		Currency:    "USD",
		Description: "Web development services",
		Status:      StatusPending,
		IssuedDate:  now,
		DueDate:     dueDate,
	}

	if invoice.InvoiceNum != "inv.test.20250124" {
		t.Errorf("Expected InvoiceNum 'inv.test.20250124', got %s", invoice.InvoiceNum)
	}

	if invoice.Amount != 1500.50 {
		t.Errorf("Expected Amount 1500.50, got %f", invoice.Amount)
	}

	if invoice.Status != StatusPending {
		t.Errorf("Expected Status pending, got %s", invoice.Status)
	}

	if invoice.Currency != "USD" {
		t.Errorf("Expected Currency USD, got %s", invoice.Currency)
	}
}

func TestInvoiceLineItem_Calculation(t *testing.T) {
	lineItem := InvoiceLineItem{
		InvoiceID:   1,
		ItemName:    "Consulting",
		Description: "Consulting hours",
		Quantity:    10,
		Rate:        150.0,
		Amount:      1500.0,
	}

	expectedTotal := lineItem.Quantity * lineItem.Rate
	if expectedTotal != 1500.0 {
		t.Errorf("Expected total 1500.0, got %f", expectedTotal)
	}

	if lineItem.Amount != expectedTotal {
		t.Errorf("Amount should match Quantity * Rate")
	}
}

func TestExpense_Initialization(t *testing.T) {
	expense := Expense{
		Description: "Adobe Creative Cloud",
		Amount:      52.99,
		Currency:    "USD",
		Category:    ExpenseCategorySoftware,
		Date:        time.Now(),
		Vendor:      "Adobe",
		Notes:       "Monthly subscription",
	}

	if expense.Category != ExpenseCategorySoftware {
		t.Errorf("Expected Category software, got %s", expense.Category)
	}

	if expense.Amount != 52.99 {
		t.Errorf("Expected Amount 52.99, got %f", expense.Amount)
	}

	if expense.Description != "Adobe Creative Cloud" {
		t.Errorf("Expected Description 'Adobe Creative Cloud', got %s", expense.Description)
	}
}

func TestTrackingSession_Billable(t *testing.T) {
	clientID := uint(1)
	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	session := TrackingSession{
		ClientID:    &clientID,
		ProjectName: "Development work",
		StartTime:   startTime,
		EndTime:     &endTime,
		Billable:    true,
	}

	if !session.Billable {
		t.Error("Session should be billable")
	}

	if session.EndTime != nil {
		duration := session.EndTime.Sub(session.StartTime)
		hours := duration.Hours()

		if hours < 1.9 || hours > 2.1 {
			t.Errorf("Expected duration ~2 hours, got %f", hours)
		}
	}
}

func TestTrackingSession_NonBillable(t *testing.T) {
	clientID := uint(1)
	endTime := time.Now()

	session := TrackingSession{
		ClientID:    &clientID,
		ProjectName: "Internal meeting",
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     &endTime,
		Billable:    false,
	}

	if session.Billable {
		t.Error("Session should not be billable")
	}
}

func TestInvoiceRecipient_LinkingInvoiceToClient(t *testing.T) {
	recipient := InvoiceRecipient{
		InvoiceID: 1,
		ClientID:  2,
	}

	if recipient.InvoiceID != 1 {
		t.Errorf("Expected InvoiceID 1, got %d", recipient.InvoiceID)
	}

	if recipient.ClientID != 2 {
		t.Errorf("Expected ClientID 2, got %d", recipient.ClientID)
	}
}
