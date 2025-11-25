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

func TestContractController_List(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	// Create test client first
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	// Create test contracts
	hourlyRate := 100.0
	fixedPrice := 5000.0
	contracts := []models.Contract{
		{
			ContractNum:  "CON-001",
			ClientID:     client.ID,
			Name:         "Contract A",
			ContractType: models.ContractTypeHourly,
			HourlyRate:   &hourlyRate,
			Currency:     "USD",
			Active:       true,
		},
		{
			ContractNum:  "CON-002",
			ClientID:     client.ID,
			Name:         "Contract B",
			ContractType: models.ContractTypeFixedPrice,
			FixedPrice:   &fixedPrice,
			Currency:     "USD",
			Active:       true,
		},
	}
	for _, contract := range contracts {
		db.Create(&contract)
	}

	// Create request
	req := httptest.NewRequest("GET", "/contracts", nil)
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Contract
	DecodeStandardResponse(t, w.Body, &response)
	assert.Len(t, response, 2)
	assert.Equal(t, "Contract B", response[0].Name) // Ordered by created_at DESC
}

func TestContractController_Get(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	// Create test contract
	hourlyRate := 150.0
	contract := models.Contract{
		ContractNum:  "CON-TEST",
		ClientID:     client.ID,
		Name:         "Test Contract",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		Active:       true,
		Notes:        "Test notes",
	}
	db.Create(&contract)

	req := httptest.NewRequest("GET", "/contracts/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Contract
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Test Contract", response.Name)
	assert.Equal(t, "CON-TEST", response.ContractNum)
	assert.Equal(t, 150.0, *response.HourlyRate)
}

func TestContractController_Get_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	req := httptest.NewRequest("GET", "/contracts/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestContractController_Create(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	hourlyRate := 125.0
	payload := map[string]interface{}{
		"contract_num":  "CON-NEW",
		"client_id":     client.ID,
		"name":          "New Contract",
		"contract_type": "hourly",
		"hourly_rate":   hourlyRate,
		"currency":      "USD",
		"start_date":    "2024-01-01",
		"active":        true,
		"notes":         "New contract notes",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/contracts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Contract
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "New Contract", response.Name)
	assert.Equal(t, "CON-NEW", response.ContractNum)
	assert.Equal(t, hourlyRate, *response.HourlyRate)
	assert.NotZero(t, response.ID)

	// Verify in database
	var dbContract models.Contract
	db.First(&dbContract, response.ID)
	assert.Equal(t, "New Contract", dbContract.Name)
}

func TestContractController_Create_ValidationError(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "Missing contract number",
			payload: map[string]interface{}{
				"client_id": client.ID,
				"name":      "Test",
			},
		},
		{
			name: "Missing client ID",
			payload: map[string]interface{}{
				"contract_num": "CON-001",
				"name":         "Test",
			},
		},
		{
			name: "Missing name",
			payload: map[string]interface{}{
				"contract_num": "CON-001",
				"client_id":    client.ID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/contracts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := WithTenantDB(req.Context(), db)
			req = req.WithContext(ctx)

			controller.Create(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestContractController_Update(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	// Create contract
	hourlyRate := 100.0
	contract := models.Contract{
		ContractNum:  "CON-001",
		ClientID:     client.ID,
		Name:         "Old Name",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		Active:       true,
	}
	db.Create(&contract)

	newRate := 150.0
	payload := map[string]interface{}{
		"name":        "Updated Name",
		"hourly_rate": newRate,
		"notes":       "Updated notes",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/contracts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Contract
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, newRate, *response.HourlyRate)
	assert.Equal(t, "Updated notes", response.Notes)

	// Verify in database
	var dbContract models.Contract
	db.First(&dbContract, 1)
	assert.Equal(t, "Updated Name", dbContract.Name)
}

func TestContractController_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	payload := map[string]string{"name": "Updated"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/contracts/999", bytes.NewBuffer(body))
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

func TestContractController_Delete(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	// Create contract
	hourlyRate := 100.0
	contract := models.Contract{
		ContractNum:  "CON-001",
		ClientID:     client.ID,
		Name:         "To Delete",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
	}
	db.Create(&contract)

	req := httptest.NewRequest("DELETE", "/contracts/1", nil)
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
	db.Model(&models.Contract{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestContractController_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewContractController()

	req := httptest.NewRequest("DELETE", "/contracts/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
