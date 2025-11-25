package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"ung/api/internal/models"
)

func TestInvoiceController_List(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	// Create test invoices
	invoices := []models.Invoice{
		{
			InvoiceNum:  "INV-001",
			CompanyID:   company.ID,
			Amount:      1000.00,
			Currency:    "USD",
			Description: "Invoice A",
			Status:      models.StatusPending,
		},
		{
			InvoiceNum:  "INV-002",
			CompanyID:   company.ID,
			Amount:      2000.00,
			Currency:    "USD",
			Description: "Invoice B",
			Status:      models.StatusSent,
		},
	}
	for _, invoice := range invoices {
		db.Create(&invoice)
	}

	req := httptest.NewRequest("GET", "/invoices", nil)
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Invoice
	DecodeStandardResponse(t, w.Body, &response)
	assert.Len(t, response, 2)
	assert.Equal(t, "Invoice B", response[0].Description) // Ordered by created_at DESC
}

func TestInvoiceController_Get(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	// Create test invoice
	invoice := models.Invoice{
		InvoiceNum:  "INV-TEST",
		CompanyID:   company.ID,
		Amount:      1500.00,
		Currency:    "USD",
		Description: "Test Invoice",
		Status:      models.StatusPending,
	}
	db.Create(&invoice)

	// Create line items
	lineItem := models.InvoiceLineItem{
		InvoiceID:   invoice.ID,
		ItemName:    "Service A",
		Description: "Consulting services",
		Quantity:    10,
		Rate:        100.00,
		Amount:      1000.00,
	}
	db.Create(&lineItem)

	// Create recipient
	recipient := models.InvoiceRecipient{
		InvoiceID: invoice.ID,
		ClientID:  client.ID,
	}
	db.Create(&recipient)

	req := httptest.NewRequest("GET", "/invoices/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	DecodeStandardResponse(t, w.Body, &response)
	assert.NotNil(t, response["invoice"])
	assert.NotNil(t, response["line_items"])
	assert.NotNil(t, response["recipients"])

	// Verify line items count
	lineItems := response["line_items"].([]interface{})
	assert.Len(t, lineItems, 1)

	// Verify recipients count
	recipients := response["recipients"].([]interface{})
	assert.Len(t, recipients, 1)
}

func TestInvoiceController_Get_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	req := httptest.NewRequest("GET", "/invoices/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInvoiceController_Create(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	payload := map[string]interface{}{
		"invoice_num": "INV-NEW",
		"company_id":  company.ID,
		"amount":      1500.00,
		"currency":    "USD",
		"description": "New Invoice",
		"status":      "pending",
		"issued_date": "2024-01-01",
		"due_date":    "2024-01-31",
		"line_items": []map[string]interface{}{
			{
				"item_name":   "Service A",
				"description": "Consulting",
				"quantity":    10.0,
				"rate":        100.00,
				"amount":      1000.00,
			},
			{
				"item_name":   "Service B",
				"description": "Development",
				"quantity":    5.0,
				"rate":        100.00,
				"amount":      500.00,
			},
		},
		"client_ids": []uint{client.ID},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/invoices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Invoice
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "INV-NEW", response.InvoiceNum)
	assert.Equal(t, 1500.00, response.Amount)
	assert.NotZero(t, response.ID)

	// Verify in database
	var dbInvoice models.Invoice
	db.First(&dbInvoice, response.ID)
	assert.Equal(t, "New Invoice", dbInvoice.Description)

	// Verify line items were created
	var lineItems []models.InvoiceLineItem
	db.Where("invoice_id = ?", response.ID).Find(&lineItems)
	assert.Len(t, lineItems, 2)

	// Verify recipients were created
	var recipients []models.InvoiceRecipient
	db.Where("invoice_id = ?", response.ID).Find(&recipients)
	assert.Len(t, recipients, 1)
}

func TestInvoiceController_Create_ValidationError(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "Missing invoice number",
			payload: map[string]interface{}{
				"company_id": company.ID,
				"amount":     1000.00,
			},
		},
		{
			name: "Missing company ID",
			payload: map[string]interface{}{
				"invoice_num": "INV-001",
				"amount":      1000.00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/invoices", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := WithTenantDB(req.Context(), db)
			req = req.WithContext(ctx)

			controller.Create(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestInvoiceController_Create_DefaultValues(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	payload := map[string]interface{}{
		"invoice_num": "INV-DEFAULT",
		"company_id":  company.ID,
		"amount":      1000.00,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/invoices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Invoice
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "USD", response.Currency)      // Default currency
	assert.Equal(t, models.StatusPending, response.Status) // Default status
}

func TestInvoiceController_Update(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	// Create invoice
	invoice := models.Invoice{
		InvoiceNum:  "INV-001",
		CompanyID:   company.ID,
		Amount:      1000.00,
		Currency:    "USD",
		Description: "Old Description",
		Status:      models.StatusPending,
	}
	db.Create(&invoice)

	payload := map[string]interface{}{
		"amount":      1500.00,
		"description": "Updated Description",
		"status":      "sent",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/invoices/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Invoice
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, 1500.00, response.Amount)
	assert.Equal(t, "Updated Description", response.Description)
	assert.Equal(t, models.StatusSent, response.Status)

	// Verify in database
	var dbInvoice models.Invoice
	db.First(&dbInvoice, 1)
	assert.Equal(t, "Updated Description", dbInvoice.Description)
}

func TestInvoiceController_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	payload := map[string]interface{}{"amount": 1000.00}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/invoices/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInvoiceController_Delete(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	// Create invoice
	invoice := models.Invoice{
		InvoiceNum:  "INV-001",
		CompanyID:   company.ID,
		Amount:      1000.00,
		Currency:    "USD",
		Status:      models.StatusPending,
	}
	db.Create(&invoice)

	// Create line items
	lineItem := models.InvoiceLineItem{
		InvoiceID: invoice.ID,
		ItemName:  "Service A",
		Quantity:  10,
		Rate:      100.00,
		Amount:    1000.00,
	}
	db.Create(&lineItem)

	// Create recipient
	recipient := models.InvoiceRecipient{
		InvoiceID: invoice.ID,
		ClientID:  client.ID,
	}
	db.Create(&recipient)

	req := httptest.NewRequest("DELETE", "/invoices/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify invoice deletion
	var invoiceCount int64
	db.Model(&models.Invoice{}).Count(&invoiceCount)
	assert.Equal(t, int64(0), invoiceCount)

	// Verify line items deletion
	var lineItemCount int64
	db.Model(&models.InvoiceLineItem{}).Where("invoice_id = ?", invoice.ID).Count(&lineItemCount)
	assert.Equal(t, int64(0), lineItemCount)

	// Verify recipients deletion
	var recipientCount int64
	db.Model(&models.InvoiceRecipient{}).Where("invoice_id = ?", invoice.ID).Count(&recipientCount)
	assert.Equal(t, int64(0), recipientCount)
}

func TestInvoiceController_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	req := httptest.NewRequest("DELETE", "/invoices/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInvoiceController_UpdateStatus(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	// Create test company
	company := models.Company{Name: "Test Company", Email: "company@test.com"}
	db.Create(&company)

	// Create invoice
	invoice := models.Invoice{
		InvoiceNum:  "INV-001",
		CompanyID:   company.ID,
		Amount:      1000.00,
		Currency:    "USD",
		Status:      models.StatusPending,
	}
	db.Create(&invoice)

	payload := map[string]string{
		"status": "paid",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/invoices/1/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.UpdateStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Invoice
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, models.StatusPaid, response.Status)

	// Verify in database
	var dbInvoice models.Invoice
	db.First(&dbInvoice, 1)
	assert.Equal(t, models.StatusPaid, dbInvoice.Status)
}

func TestInvoiceController_UpdateStatus_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewInvoiceController()

	payload := map[string]string{"status": "paid"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/invoices/999/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.UpdateStatus(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
