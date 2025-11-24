package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// InvoiceController handles invoice endpoints
type InvoiceController struct{}

// NewInvoiceController creates a new invoice controller
func NewInvoiceController() *InvoiceController {
	return &InvoiceController{}
}

// List handles GET /api/v1/invoices
func (c *InvoiceController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var invoices []models.Invoice
	if err := db.Preload("Company").Order("created_at DESC").Find(&invoices).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, invoices, http.StatusOK)
}

// Get handles GET /api/v1/invoices/:id
func (c *InvoiceController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var invoice models.Invoice
	if err := db.Preload("Company").First(&invoice, id).Error; err != nil {
		RespondError(w, "Invoice not found", http.StatusNotFound)
		return
	}

	// Load line items
	var lineItems []models.InvoiceLineItem
	db.Where("invoice_id = ?", invoice.ID).Find(&lineItems)

	// Load recipients
	var recipients []models.InvoiceRecipient
	db.Where("invoice_id = ?", invoice.ID).Find(&recipients)

	response := map[string]interface{}{
		"invoice":    invoice,
		"line_items": lineItems,
		"recipients": recipients,
	}

	RespondJSON(w, response, http.StatusOK)
}

// Create handles POST /api/v1/invoices
func (c *InvoiceController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		InvoiceNum  string                      `json:"invoice_num"`
		CompanyID   uint                        `json:"company_id"`
		Amount      float64                     `json:"amount"`
		Currency    string                      `json:"currency"`
		Description string                      `json:"description"`
		Status      models.InvoiceStatus        `json:"status"`
		IssuedDate  string                      `json:"issued_date"`
		DueDate     string                      `json:"due_date"`
		LineItems   []models.InvoiceLineItem    `json:"line_items"`
		ClientIDs   []uint                      `json:"client_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.InvoiceNum == "" {
		RespondError(w, "Invoice number is required", http.StatusBadRequest)
		return
	}
	if req.CompanyID == 0 {
		RespondError(w, "Company ID is required", http.StatusBadRequest)
		return
	}

	// Parse dates
	issuedDate := time.Now()
	if req.IssuedDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.IssuedDate); err == nil {
			issuedDate = parsed
		}
	}

	dueDate := issuedDate.AddDate(0, 0, 30) // Default 30 days
	if req.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			dueDate = parsed
		}
	}

	invoice := models.Invoice{
		InvoiceNum:  req.InvoiceNum,
		CompanyID:   req.CompanyID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Status:      req.Status,
		IssuedDate:  issuedDate,
		DueDate:     dueDate,
	}

	if invoice.Currency == "" {
		invoice.Currency = "USD"
	}
	if invoice.Status == "" {
		invoice.Status = models.StatusPending
	}

	// Create invoice
	if err := db.Create(&invoice).Error; err != nil {
		RespondError(w, "Failed to create invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create line items
	for _, item := range req.LineItems {
		item.InvoiceID = invoice.ID
		db.Create(&item)
	}

	// Create recipients
	for _, clientID := range req.ClientIDs {
		db.Create(&models.InvoiceRecipient{
			InvoiceID: invoice.ID,
			ClientID:  clientID,
		})
	}

	RespondJSON(w, invoice, http.StatusCreated)
}

// Update handles PUT /api/v1/invoices/:id
func (c *InvoiceController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var invoice models.Invoice
	if err := db.First(&invoice, id).Error; err != nil {
		RespondError(w, "Invoice not found", http.StatusNotFound)
		return
	}

	var req struct {
		Amount      *float64              `json:"amount"`
		Currency    *string               `json:"currency"`
		Description *string               `json:"description"`
		Status      *models.InvoiceStatus `json:"status"`
		DueDate     *string               `json:"due_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Amount != nil {
		invoice.Amount = *req.Amount
	}
	if req.Currency != nil {
		invoice.Currency = *req.Currency
	}
	if req.Description != nil {
		invoice.Description = *req.Description
	}
	if req.Status != nil {
		invoice.Status = *req.Status
	}
	if req.DueDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			invoice.DueDate = parsed
		}
	}

	if err := db.Save(&invoice).Error; err != nil {
		RespondError(w, "Failed to update invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, invoice, http.StatusOK)
}

// Delete handles DELETE /api/v1/invoices/:id
func (c *InvoiceController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var invoice models.Invoice
	if err := db.First(&invoice, id).Error; err != nil {
		RespondError(w, "Invoice not found", http.StatusNotFound)
		return
	}

	// Delete related records
	db.Where("invoice_id = ?", invoice.ID).Delete(&models.InvoiceLineItem{})
	db.Where("invoice_id = ?", invoice.ID).Delete(&models.InvoiceRecipient{})

	// Delete invoice
	if err := db.Delete(&invoice).Error; err != nil {
		RespondError(w, "Failed to delete invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Invoice deleted successfully"}, http.StatusOK)
}

// UpdateStatus handles PATCH /api/v1/invoices/:id/status
func (c *InvoiceController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var invoice models.Invoice
	if err := db.First(&invoice, id).Error; err != nil {
		RespondError(w, "Invoice not found", http.StatusNotFound)
		return
	}

	var req struct {
		Status models.InvoiceStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	invoice.Status = req.Status

	if err := db.Save(&invoice).Error; err != nil {
		RespondError(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, invoice, http.StatusOK)
}
