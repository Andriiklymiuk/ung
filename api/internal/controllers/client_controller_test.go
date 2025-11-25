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

func TestClientController_List(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	// Create test clients
	clients := []models.Client{
		{Name: "Client A", Email: "a@example.com", Address: "123 St", TaxID: "TAX1"},
		{Name: "Client B", Email: "b@example.com", Address: "456 Ave", TaxID: "TAX2"},
	}
	for _, client := range clients {
		db.Create(&client)
	}

	// Create request
	req := httptest.NewRequest("GET", "/clients", nil)
	w := httptest.NewRecorder()

	// Add tenant DB to context
	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	// Call handler
	controller.List(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Client
	DecodeStandardResponse(t, w.Body, &response)
	assert.Len(t, response, 2)
	assert.Equal(t, "Client B", response[0].Name) // Ordered by created_at DESC
}

func TestClientController_Get(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	// Create test client
	client := models.Client{
		Name:    "Test Client",
		Email:   "test@example.com",
		Address: "123 Test St",
		TaxID:   "TAX123",
	}
	db.Create(&client)

	// Create request
	req := httptest.NewRequest("GET", "/clients/1", nil)
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

	var response models.Client
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Test Client", response.Name)
	assert.Equal(t, "test@example.com", response.Email)
}

func TestClientController_Get_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	req := httptest.NewRequest("GET", "/clients/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClientController_Create(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	payload := map[string]string{
		"name":    "New Client",
		"email":   "new@example.com",
		"address": "789 New St",
		"tax_id":  "TAX789",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Client
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "New Client", response.Name)
	assert.Equal(t, "new@example.com", response.Email)
	assert.NotZero(t, response.ID)

	// Verify in database
	var dbClient models.Client
	db.First(&dbClient, response.ID)
	assert.Equal(t, "New Client", dbClient.Name)
}

func TestClientController_Create_ValidationError(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	tests := []struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "Missing name",
			payload: map[string]string{"email": "test@example.com"},
		},
		{
			name:    "Missing email",
			payload: map[string]string{"name": "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/clients", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := WithTenantDB(req.Context(), db)
			req = req.WithContext(ctx)
			controller.Create(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestClientController_Update(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	// Create client
	client := models.Client{Name: "Old Name", Email: "old@example.com"}
	db.Create(&client)

	payload := map[string]string{
		"name":  "Updated Name",
		"email": "updated@example.com",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/clients/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Client
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "updated@example.com", response.Email)
}

func TestClientController_Delete(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewClientController()

	// Create client
	client := models.Client{Name: "To Delete", Email: "delete@example.com"}
	db.Create(&client)

	req := httptest.NewRequest("DELETE", "/clients/1", nil)
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
	db.Model(&models.Client{}).Count(&count)
	assert.Equal(t, int64(0), count)
}
