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

// ExpenseController handles expense endpoints
type ExpenseController struct{}

// NewExpenseController creates a new expense controller
func NewExpenseController() *ExpenseController {
	return &ExpenseController{}
}

// List handles GET /api/v1/expenses
func (c *ExpenseController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var expenses []models.Expense
	if err := db.Order("date DESC").Find(&expenses).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, expenses, http.StatusOK)
}

// Get handles GET /api/v1/expenses/:id
func (c *ExpenseController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		RespondError(w, "Expense not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, expense, http.StatusOK)
}

// Create handles POST /api/v1/expenses
func (c *ExpenseController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		Description string                  `json:"description"`
		Amount      float64                 `json:"amount"`
		Currency    string                  `json:"currency"`
		Category    models.ExpenseCategory  `json:"category"`
		Date        string                  `json:"date"`
		Vendor      string                  `json:"vendor"`
		ReceiptPath string                  `json:"receipt_path"`
		Notes       string                  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Description == "" {
		RespondError(w, "Description is required", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		RespondError(w, "Amount must be greater than zero", http.StatusBadRequest)
		return
	}

	// Parse date
	expenseDate := time.Now()
	if req.Date != "" {
		if parsed, err := time.Parse("2006-01-02", req.Date); err == nil {
			expenseDate = parsed
		}
	}

	expense := models.Expense{
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Category:    req.Category,
		Date:        expenseDate,
		Vendor:      req.Vendor,
		ReceiptPath: req.ReceiptPath,
		Notes:       req.Notes,
	}

	if expense.Currency == "" {
		expense.Currency = "USD"
	}
	if expense.Category == "" {
		expense.Category = models.ExpenseCategoryOther
	}

	if err := db.Create(&expense).Error; err != nil {
		RespondError(w, "Failed to create expense: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, expense, http.StatusCreated)
}

// Update handles PUT /api/v1/expenses/:id
func (c *ExpenseController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		RespondError(w, "Expense not found", http.StatusNotFound)
		return
	}

	var req struct {
		Description *string                 `json:"description"`
		Amount      *float64                `json:"amount"`
		Currency    *string                 `json:"currency"`
		Category    *models.ExpenseCategory `json:"category"`
		Date        *string                 `json:"date"`
		Vendor      *string                 `json:"vendor"`
		ReceiptPath *string                 `json:"receipt_path"`
		Notes       *string                 `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Description != nil {
		expense.Description = *req.Description
	}
	if req.Amount != nil {
		expense.Amount = *req.Amount
	}
	if req.Currency != nil {
		expense.Currency = *req.Currency
	}
	if req.Category != nil {
		expense.Category = *req.Category
	}
	if req.Vendor != nil {
		expense.Vendor = *req.Vendor
	}
	if req.ReceiptPath != nil {
		expense.ReceiptPath = *req.ReceiptPath
	}
	if req.Notes != nil {
		expense.Notes = *req.Notes
	}
	if req.Date != nil && *req.Date != "" {
		if parsed, err := time.Parse("2006-01-02", *req.Date); err == nil {
			expense.Date = parsed
		}
	}

	if err := db.Save(&expense).Error; err != nil {
		RespondError(w, "Failed to update expense: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, expense, http.StatusOK)
}

// Delete handles DELETE /api/v1/expenses/:id
func (c *ExpenseController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		RespondError(w, "Expense not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&expense).Error; err != nil {
		RespondError(w, "Failed to delete expense: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Expense deleted successfully"}, http.StatusOK)
}
