package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"ung/api/internal/models"
)

func TestTrackingController_List(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create test client and contract
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	hourlyRate := 100.0
	contract := models.Contract{
		ContractNum:  "CON-001",
		ClientID:     client.ID,
		Name:         "Test Contract",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
	}
	db.Create(&contract)

	// Create test sessions
	sessions := []models.TrackingSession{
		{
			ClientID:    &client.ID,
			ContractID:  &contract.ID,
			ProjectName: "Project A",
			StartTime:   time.Now().Add(-2 * time.Hour),
			Billable:    true,
		},
		{
			ClientID:    &client.ID,
			ContractID:  &contract.ID,
			ProjectName: "Project B",
			StartTime:   time.Now().Add(-1 * time.Hour),
			Billable:    false,
		},
	}
	for _, session := range sessions {
		db.Create(&session)
	}

	req := httptest.NewRequest("GET", "/tracking", nil)
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.Len(t, response, 2)
}

func TestTrackingController_Get(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create test session
	session := models.TrackingSession{
		ProjectName: "Test Project",
		StartTime:   time.Now(),
		Billable:    true,
		Notes:       "Test notes",
	}
	db.Create(&session)

	req := httptest.NewRequest("GET", "/tracking/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Test Project", response.ProjectName)
	assert.True(t, response.Billable)
}

func TestTrackingController_Get_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	req := httptest.NewRequest("GET", "/tracking/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTrackingController_Start(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	payload := map[string]interface{}{
		"client_id":    client.ID,
		"project_name": "New Project",
		"billable":     true,
		"notes":        "Starting work",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/tracking/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Start(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "New Project", response.ProjectName)
	assert.True(t, response.Billable)
	assert.Nil(t, response.EndTime)
	assert.NotZero(t, response.ID)

	// Verify in database
	var dbSession models.TrackingSession
	db.First(&dbSession, response.ID)
	assert.Equal(t, "New Project", dbSession.ProjectName)
	assert.Nil(t, dbSession.EndTime)
}

func TestTrackingController_Start_ConflictWithActiveSession(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create an active session (no end time)
	activeSession := models.TrackingSession{
		ProjectName: "Active Project",
		StartTime:   time.Now(),
		Billable:    true,
	}
	db.Create(&activeSession)

	payload := map[string]interface{}{
		"project_name": "New Project",
		"billable":     true,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/tracking/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Start(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestTrackingController_Stop(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create an active session
	startTime := time.Now().Add(-2 * time.Hour)
	session := models.TrackingSession{
		ProjectName: "Test Project",
		StartTime:   startTime,
		Billable:    true,
	}
	db.Create(&session)

	req := httptest.NewRequest("POST", "/tracking/1/stop", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Stop(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.NotNil(t, response.EndTime)
	assert.NotNil(t, response.Duration)
	assert.NotNil(t, response.Hours)
	assert.Greater(t, *response.Hours, 0.0)

	// Verify in database
	var dbSession models.TrackingSession
	db.First(&dbSession, 1)
	assert.NotNil(t, dbSession.EndTime)
	assert.NotNil(t, dbSession.Duration)
}

func TestTrackingController_Stop_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	req := httptest.NewRequest("POST", "/tracking/999/stop", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Stop(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTrackingController_Stop_AlreadyStopped(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create a stopped session
	endTime := time.Now()
	duration := 3600
	hours := 1.0
	session := models.TrackingSession{
		ProjectName: "Test Project",
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     &endTime,
		Duration:    &duration,
		Hours:       &hours,
		Billable:    true,
	}
	db.Create(&session)

	req := httptest.NewRequest("POST", "/tracking/1/stop", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Stop(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTrackingController_Create_ManualEntry(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create test client
	client := models.Client{Name: "Test Client", Email: "client@test.com"}
	db.Create(&client)

	startTime := time.Now().Add(-3 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	payload := map[string]interface{}{
		"client_id":    client.ID,
		"project_name": "Manual Entry",
		"start_time":   startTime,
		"end_time":     endTime,
		"billable":     true,
		"notes":        "Manual time entry",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/tracking", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Manual Entry", response.ProjectName)
	assert.NotNil(t, response.EndTime)
	assert.NotNil(t, response.Duration)
	assert.NotNil(t, response.Hours)
	assert.Greater(t, *response.Hours, 1.5) // Should be around 2 hours
}

func TestTrackingController_Create_ValidationError(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "Invalid start_time format",
			payload: map[string]interface{}{
				"project_name": "Test",
				"start_time":   "invalid-time",
				"end_time":     time.Now().Format(time.RFC3339),
			},
		},
		{
			name: "Invalid end_time format",
			payload: map[string]interface{}{
				"project_name": "Test",
				"start_time":   time.Now().Format(time.RFC3339),
				"end_time":     "invalid-time",
			},
		},
		{
			name: "End time before start time",
			payload: map[string]interface{}{
				"project_name": "Test",
				"start_time":   time.Now().Format(time.RFC3339),
				"end_time":     time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/tracking", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := WithTenantDB(req.Context(), db)
			req = req.WithContext(ctx)

			controller.Create(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestTrackingController_Update(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create session
	session := models.TrackingSession{
		ProjectName: "Old Project",
		StartTime:   time.Now(),
		Billable:    false,
		Notes:       "Old notes",
	}
	db.Create(&session)

	payload := map[string]interface{}{
		"project_name": "Updated Project",
		"billable":     true,
		"notes":        "Updated notes",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/tracking/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Updated Project", response.ProjectName)
	assert.True(t, response.Billable)
	assert.Equal(t, "Updated notes", response.Notes)

	// Verify in database
	var dbSession models.TrackingSession
	db.First(&dbSession, 1)
	assert.Equal(t, "Updated Project", dbSession.ProjectName)
}

func TestTrackingController_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	payload := map[string]string{"project_name": "Updated"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/tracking/999", bytes.NewBuffer(body))
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

func TestTrackingController_Delete(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create session
	session := models.TrackingSession{
		ProjectName: "To Delete",
		StartTime:   time.Now(),
		Billable:    true,
	}
	db.Create(&session)

	req := httptest.NewRequest("DELETE", "/tracking/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion (soft delete)
	var session2 models.TrackingSession
	err := db.First(&session2, 1).Error
	assert.Error(t, err) // Should be soft deleted
}

func TestTrackingController_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	req := httptest.NewRequest("DELETE", "/tracking/999", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = WithTenantDB(ctx, db)
	req = req.WithContext(ctx)

	controller.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTrackingController_Active_WithActiveSession(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create an active session (no end time)
	session := models.TrackingSession{
		ProjectName: "Active Project",
		StartTime:   time.Now(),
		Billable:    true,
	}
	db.Create(&session)

	req := httptest.NewRequest("GET", "/tracking/active", nil)
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Active(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TrackingSession
	DecodeStandardResponse(t, w.Body, &response)
	assert.Equal(t, "Active Project", response.ProjectName)
	assert.Nil(t, response.EndTime)
}

func TestTrackingController_Active_NoActiveSession(t *testing.T) {
	db := SetupTestDB(t)
	controller := NewTrackingController()

	// Create a stopped session
	endTime := time.Now()
	duration := 3600
	hours := 1.0
	session := models.TrackingSession{
		ProjectName: "Stopped Project",
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     &endTime,
		Duration:    &duration,
		Hours:       &hours,
		Billable:    true,
	}
	db.Create(&session)

	req := httptest.NewRequest("GET", "/tracking/active", nil)
	w := httptest.NewRecorder()

	ctx := WithTenantDB(req.Context(), db)
	req = req.WithContext(ctx)

	controller.Active(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should return null/nil when no active session
	var response interface{}
	DecodeStandardResponse(t, w.Body, &response)
	assert.Nil(t, response)
}
