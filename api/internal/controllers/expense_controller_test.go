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

func TestExpenseController_List(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	// Create test expenses
	expenses := []models.Expense{
		{
			Description: "Expense A",
			Amount:      100.50,
			Currency:    "USD",
			Category:    models.ExpenseCategorySoftware,
			Vendor:      "Vendor A",
		},
		{
			Description: "Expense B",
			Amount:      250.75,
			Currency:    "USD",
			Category:    models.ExpenseCategoryHardware,
			Vendor:      "Vendor B",
		},
	}
	for _, expense := range expenses {
		db.Create(&expense)
	}

	req := httptest.NewRequest("GET", "/expenses", nil)
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Expense
	DecodeStandardResponse(t, w.Body, &response)
	assert.Len(t, response, 2)
}

func TestExpenseController_Get(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	// Create test expense
	expense := models.Expense{
		Description: "Test Expense",
		Amount:      500.00,
		Currency:    "USD",
		Category:    models.ExpenseCategorySoftware,
		Vendor:      "Adobe",
		Notes:       "Monthly subscription",
	}
	db.Create(&expense)

	req := httptest.NewRequest("GET", "/expenses/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Expense
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Test Expense", response.Description)
	assert.Equal(t, 500.00, response.Amount)
	assert.Equal(t, "Adobe", response.Vendor)
}

func TestExpenseController_Get_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	req := httptest.NewRequest("GET", "/expenses/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestExpenseController_Create(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	payload := map[string]interface{}{
		"description":  "New Expense",
		"amount":       150.00,
		"currency":     "USD",
		"category":     "software",
		"date":         "2024-01-15",
		"vendor":       "Microsoft",
		"receipt_path": "/receipts/receipt1.pdf",
		"notes":        "Office 365 subscription",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Expense
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "New Expense", response.Description)
	assert.Equal(t, 150.00, response.Amount)
	assert.Equal(t, "Microsoft", response.Vendor)
	assert.NotZero(t, response.ID)

	// Verify in database
	var dbExpense models.Expense
	db.First(&dbExpense, response.ID)
	assert.Equal(t, "New Expense", dbExpense.Description)
}

func TestExpenseController_Create_ValidationError(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "Missing description",
			payload: map[string]interface{}{
				"amount":   100.00,
				"category": "software",
			},
		},
		{
			name: "Invalid amount (zero)",
			payload: map[string]interface{}{
				"description": "Test",
				"amount":      0,
				"category":    "software",
			},
		},
		{
			name: "Invalid amount (negative)",
			payload: map[string]interface{}{
				"description": "Test",
				"amount":      -50.00,
				"category":    "software",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/expenses", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := WithTenantDB(req.Context(), db)
			req = req.WithContext(ctx)

			controller.Create(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestExpenseController_Create_DefaultValues(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	// Test that currency defaults to USD and category defaults to "other"
	payload := map[string]interface{}{
		"description": "Test Expense",
		"amount":      100.00,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Expense
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "USD", response.Currency)
	assert.Equal(t, models.ExpenseCategoryOther, response.Category)
}

func TestExpenseController_Update(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	// Create expense
	expense := models.Expense{
		Description: "Old Description",
		Amount:      100.00,
		Currency:    "USD",
		Category:    models.ExpenseCategorySoftware,
	}
	db.Create(&expense)

	payload := map[string]interface{}{
		"description": "Updated Description",
		"amount":      200.00,
		"vendor":      "New Vendor",
		"notes":       "Updated notes",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/expenses/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Expense
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Updated Description", response.Description)
	assert.Equal(t, 200.00, response.Amount)
	assert.Equal(t, "New Vendor", response.Vendor)

	// Verify in database
	var dbExpense models.Expense
	db.First(&dbExpense, 1)
	assert.Equal(t, "Updated Description", dbExpense.Description)
}

func TestExpenseController_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	payload := map[string]interface{}{"description": "Updated"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/expenses/999", bytes.NewBuffer(body))
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

func TestExpenseController_Delete(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	// Create expense
	expense := models.Expense{
		Description: "To Delete",
		Amount:      50.00,
		Currency:    "USD",
		Category:    models.ExpenseCategoryOther,
	}
	db.Create(&expense)

	req := httptest.NewRequest("DELETE", "/expenses/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	var count int64
	db.Model(&models.Expense{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestExpenseController_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewExpenseController()

	req := httptest.NewRequest("DELETE", "/expenses/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
