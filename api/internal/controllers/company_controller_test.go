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

func TestCompanyController_List(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	// Create test companies
	companies := []models.Company{
		{Name: "Company A", Email: "a@company.com", Phone: "123-456-7890", TaxID: "TAX1"},
		{Name: "Company B", Email: "b@company.com", Phone: "098-765-4321", TaxID: "TAX2"},
	}
	for _, company := range companies {
		db.Create(&company)
	}

	// Create request
	req := httptest.NewRequest("GET", "/companies", nil)
	w := httptest.NewRecorder()

	// Add tenant DB to context
	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	// Call handler
	controller.List(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var responseCompanies []models.Company
	DecodeStandardResponse(t, w.Body, &responseCompanies)

	assert.Len(t, responseCompanies, 2)
	assert.Equal(t, "Company B", responseCompanies[0].Name) // Ordered by created_at DESC
}

func TestCompanyController_Get(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	// Create test company
	company := models.Company{
		Name:                "Test Company",
		Email:               "test@company.com",
		Phone:               "123-456-7890",
		Address:             "123 Test St",
		RegistrationAddress: "456 Reg St",
		TaxID:               "TAX123",
		BankName:            "Test Bank",
		BankAccount:         "ACC123",
		BankSWIFT:           "SWIFT123",
		LogoPath:            "/logo.png",
	}
	db.Create(&company)

	// Create request
	req := httptest.NewRequest("GET", "/companies/1", nil)
	w := httptest.NewRecorder()

	// Add URL params and DB to context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	// Call handler
	controller.Get(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Company
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Test Company", response.Name)
	assert.Equal(t, "test@company.com", response.Email)
	assert.Equal(t, "Test Bank", response.BankName)
}

func TestCompanyController_Get_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	req := httptest.NewRequest("GET", "/companies/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCompanyController_Create(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	payload := map[string]string{
		"name":                 "New Company",
		"email":                "new@company.com",
		"phone":                "555-1234",
		"address":              "789 New St",
		"registration_address": "101 Reg Ave",
		"tax_id":               "TAX789",
		"bank_name":            "New Bank",
		"bank_account":         "ACC789",
		"bank_swift":           "SWIFT789",
		"logo_path":            "/newlogo.png",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/companies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Company
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "New Company", response.Name)
	assert.Equal(t, "new@company.com", response.Email)
	assert.Equal(t, "New Bank", response.BankName)
	assert.NotZero(t, response.ID)

	// Verify in database
	var dbCompany models.Company
	db.First(&dbCompany, response.ID)
	assert.Equal(t, "New Company", dbCompany.Name)
}

func TestCompanyController_Create_ValidationError(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	tests := []struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "Missing name",
			payload: map[string]string{"email": "test@company.com"},
		},
		{
			name:    "Missing email",
			payload: map[string]string{"name": "Test Company"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/companies", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := WithTenantDB(req.Context(), db)
			req = req.WithContext(ctx)

			controller.Create(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestCompanyController_Update(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	// Create company
	company := models.Company{
		Name:  "Old Name",
		Email: "old@company.com",
	}
	db.Create(&company)

	payload := map[string]string{
		"name":       "Updated Name",
		"email":      "updated@company.com",
		"bank_name":  "Updated Bank",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/companies/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Company
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "updated@company.com", response.Email)
	assert.Equal(t, "Updated Bank", response.BankName)

	// Verify in database
	var dbCompany models.Company
	db.First(&dbCompany, 1)
	assert.Equal(t, "Updated Name", dbCompany.Name)
}

func TestCompanyController_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	payload := map[string]string{"name": "Updated"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/companies/999", bytes.NewBuffer(body))
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

func TestCompanyController_Delete(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	// Create company
	company := models.Company{
		Name:  "To Delete",
		Email: "delete@company.com",
	}
	db.Create(&company)

	req := httptest.NewRequest("DELETE", "/companies/1", nil)
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
	db.Model(&models.Company{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCompanyController_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewCompanyController()

	req := httptest.NewRequest("DELETE", "/companies/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
