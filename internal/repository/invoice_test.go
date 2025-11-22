package repository

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestInvoiceRepository_Create(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	invoiceRepo := NewInvoiceRepository()

	company := &models.Company{
		Name:  "Test Company",
		Email: "company@example.com",
	}
	companyRepo.Create(company)

	invoice := &models.Invoice{
		InvoiceNum:  "INV-001",
		CompanyID:   company.ID,
		Amount:      1000.00,
		Currency:    "USD",
		Description: "Test Invoice",
		Status:      models.StatusPending,
		IssuedDate:  time.Now(),
		DueDate:     time.Now().AddDate(0, 0, 30),
	}

	err := invoiceRepo.Create(invoice)
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}

	if invoice.ID == 0 {
		t.Error("expected invoice ID to be set")
	}
}

func TestInvoiceRepository_GetByMonth(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	invoiceRepo := NewInvoiceRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	// Create invoices in different months
	now := time.Now()
	thisMonth := &models.Invoice{
		InvoiceNum: "INV-001",
		CompanyID:  company.ID,
		Amount:     1000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: now,
		DueDate:    now.AddDate(0, 0, 30),
	}
	invoiceRepo.Create(thisMonth)

	lastMonth := &models.Invoice{
		InvoiceNum: "INV-002",
		CompanyID:  company.ID,
		Amount:     2000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: now.AddDate(0, -1, 0),
		DueDate:    now.AddDate(0, -1, 30),
	}
	invoiceRepo.Create(lastMonth)

	// Get invoices for this month
	result, err := invoiceRepo.GetByMonth(now.Year(), int(now.Month()))
	if err != nil {
		t.Fatalf("failed to get invoices by month: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 invoice for this month, got %d", len(result))
	}

	if result[0].InvoiceNum != "INV-001" {
		t.Errorf("expected invoice INV-001, got %s", result[0].InvoiceNum)
	}
}

func TestInvoiceRepository_GetByStatus(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	invoiceRepo := NewInvoiceRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	now := time.Now()

	// Create pending invoice
	invoiceRepo.Create(&models.Invoice{
		InvoiceNum: "INV-001",
		CompanyID:  company.ID,
		Amount:     1000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: now,
		DueDate:    now.AddDate(0, 0, 30),
	})

	// Create paid invoice
	invoiceRepo.Create(&models.Invoice{
		InvoiceNum: "INV-002",
		CompanyID:  company.ID,
		Amount:     2000.00,
		Currency:   "USD",
		Status:     models.StatusPaid,
		IssuedDate: now,
		DueDate:    now.AddDate(0, 0, 30),
	})

	result, err := invoiceRepo.GetByStatus(models.StatusPending)
	if err != nil {
		t.Fatalf("failed to get invoices by status: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 pending invoice, got %d", len(result))
	}

	if result[0].Status != models.StatusPending {
		t.Errorf("expected status pending, got %s", result[0].Status)
	}
}

func TestInvoiceRepository_CountByInvoiceNumPattern(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	invoiceRepo := NewInvoiceRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	now := time.Now()

	// Create invoices with pattern INV-2024-%
	invoiceRepo.Create(&models.Invoice{
		InvoiceNum: "INV-2024-001",
		CompanyID:  company.ID,
		Amount:     1000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: now,
		DueDate:    now.AddDate(0, 0, 30),
	})

	invoiceRepo.Create(&models.Invoice{
		InvoiceNum: "INV-2024-002",
		CompanyID:  company.ID,
		Amount:     2000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: now,
		DueDate:    now.AddDate(0, 0, 30),
	})

	count, err := invoiceRepo.CountByInvoiceNumPattern("INV-2024-%")
	if err != nil {
		t.Fatalf("failed to count invoices: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 invoices with pattern, got %d", count)
	}
}

func TestInvoiceLineItemRepository_Create(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	invoiceRepo := NewInvoiceRepository()
	lineItemRepo := NewInvoiceLineItemRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	invoice := &models.Invoice{
		InvoiceNum: "INV-001",
		CompanyID:  company.ID,
		Amount:     1000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: time.Now(),
		DueDate:    time.Now().AddDate(0, 0, 30),
	}
	invoiceRepo.Create(invoice)

	lineItem := &models.InvoiceLineItem{
		InvoiceID:   invoice.ID,
		ItemName:    "Consulting",
		Description: "Development work",
		Quantity:    10.0,
		Rate:        100.00,
		Amount:      1000.00,
	}

	err := lineItemRepo.Create(lineItem)
	if err != nil {
		t.Fatalf("failed to create line item: %v", err)
	}

	if lineItem.ID == 0 {
		t.Error("expected line item ID to be set")
	}
}

func TestInvoiceLineItemRepository_GetByInvoiceID(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	invoiceRepo := NewInvoiceRepository()
	lineItemRepo := NewInvoiceLineItemRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	invoice := &models.Invoice{
		InvoiceNum: "INV-001",
		CompanyID:  company.ID,
		Amount:     1500.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: time.Now(),
		DueDate:    time.Now().AddDate(0, 0, 30),
	}
	invoiceRepo.Create(invoice)

	lineItemRepo.Create(&models.InvoiceLineItem{
		InvoiceID: invoice.ID,
		ItemName:  "Item 1",
		Quantity:  10.0,
		Rate:      100.00,
		Amount:    1000.00,
	})

	lineItemRepo.Create(&models.InvoiceLineItem{
		InvoiceID: invoice.ID,
		ItemName:  "Item 2",
		Quantity:  5.0,
		Rate:      100.00,
		Amount:    500.00,
	})

	items, err := lineItemRepo.GetByInvoiceID(invoice.ID)
	if err != nil {
		t.Fatalf("failed to get line items: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 line items, got %d", len(items))
	}
}

func TestInvoiceRecipientRepository_Create(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	clientRepo := NewClientRepository()
	invoiceRepo := NewInvoiceRepository()
	recipientRepo := NewInvoiceRecipientRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	client := &models.Client{Name: "Test Client", Email: "client@example.com"}
	clientRepo.Create(client)

	invoice := &models.Invoice{
		InvoiceNum: "INV-001",
		CompanyID:  company.ID,
		Amount:     1000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: time.Now(),
		DueDate:    time.Now().AddDate(0, 0, 30),
	}
	invoiceRepo.Create(invoice)

	err := recipientRepo.Create(invoice.ID, client.ID)
	if err != nil {
		t.Fatalf("failed to create invoice recipient: %v", err)
	}
}

func TestInvoiceRecipientRepository_GetClientByInvoiceID(t *testing.T) {
	setupTestDB(t)
	companyRepo := NewCompanyRepository()
	clientRepo := NewClientRepository()
	invoiceRepo := NewInvoiceRepository()
	recipientRepo := NewInvoiceRecipientRepository()

	company := &models.Company{Name: "Test Company", Email: "company@example.com"}
	companyRepo.Create(company)

	client := &models.Client{Name: "Test Client", Email: "client@example.com"}
	clientRepo.Create(client)

	invoice := &models.Invoice{
		InvoiceNum: "INV-001",
		CompanyID:  company.ID,
		Amount:     1000.00,
		Currency:   "USD",
		Status:     models.StatusPending,
		IssuedDate: time.Now(),
		DueDate:    time.Now().AddDate(0, 0, 30),
	}
	invoiceRepo.Create(invoice)

	recipientRepo.Create(invoice.ID, client.ID)

	retrievedClient, err := recipientRepo.GetClientByInvoiceID(invoice.ID)
	if err != nil {
		t.Fatalf("failed to get client by invoice ID: %v", err)
	}

	if retrievedClient.ID != client.ID {
		t.Errorf("expected client ID %d, got %d", client.ID, retrievedClient.ID)
	}
}
