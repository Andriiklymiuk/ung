package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
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

	type Invoice struct {
		ID         uint    `json:"id"`
		InvoiceNum string  `json:"invoice_num"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		Status     string  `json:"status"`
	}

	var invoices []Invoice
	if err := db.Table("invoices").Find(&invoices).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, invoices, http.StatusOK)
}

// Get handles GET /api/v1/invoices/:id
func (c *InvoiceController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	type Invoice struct {
		ID         uint    `json:"id"`
		InvoiceNum string  `json:"invoice_num"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		Status     string  `json:"status"`
	}

	var invoice Invoice
	if err := db.Table("invoices").First(&invoice, id).Error; err != nil {
		RespondError(w, "Invoice not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, invoice, http.StatusOK)
}

// Other CRUD methods would follow similar pattern...
