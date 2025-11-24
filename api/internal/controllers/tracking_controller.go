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

// TrackingController handles time tracking endpoints
type TrackingController struct{}

// NewTrackingController creates a new tracking controller
func NewTrackingController() *TrackingController {
	return &TrackingController{}
}

// List handles GET /api/v1/tracking
func (c *TrackingController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var sessions []models.TrackingSession
	if err := db.Preload("Client").Preload("Contract").Order("start_time DESC").Find(&sessions).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, sessions, http.StatusOK)
}

// Get handles GET /api/v1/tracking/:id
func (c *TrackingController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.TrackingSession
	if err := db.Preload("Client").Preload("Contract").First(&session, id).Error; err != nil {
		RespondError(w, "Tracking session not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, session, http.StatusOK)
}

// Start handles POST /api/v1/tracking/start
func (c *TrackingController) Start(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		ClientID    *uint  `json:"client_id"`
		ContractID  *uint  `json:"contract_id"`
		ProjectName string `json:"project_name"`
		Billable    bool   `json:"billable"`
		Notes       string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check for active session
	var activeSession models.TrackingSession
	if err := db.Where("end_time IS NULL").First(&activeSession).Error; err == nil {
		RespondError(w, "There is already an active tracking session. Please stop it first.", http.StatusConflict)
		return
	}

	session := models.TrackingSession{
		ClientID:    req.ClientID,
		ContractID:  req.ContractID,
		ProjectName: req.ProjectName,
		StartTime:   time.Now(),
		Billable:    req.Billable,
		Notes:       req.Notes,
	}

	if err := db.Create(&session).Error; err != nil {
		RespondError(w, "Failed to start tracking: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, session, http.StatusCreated)
}

// Stop handles POST /api/v1/tracking/:id/stop
func (c *TrackingController) Stop(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.TrackingSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Tracking session not found", http.StatusNotFound)
		return
	}

	if session.EndTime != nil {
		RespondError(w, "Session already stopped", http.StatusBadRequest)
		return
	}

	endTime := time.Now()
	duration := int(endTime.Sub(session.StartTime).Seconds())
	hours := float64(duration) / 3600.0

	session.EndTime = &endTime
	session.Duration = &duration
	session.Hours = &hours

	if err := db.Save(&session).Error; err != nil {
		RespondError(w, "Failed to stop tracking: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, session, http.StatusOK)
}

// Create handles POST /api/v1/tracking (manual entry)
func (c *TrackingController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		ClientID    *uint   `json:"client_id"`
		ContractID  *uint   `json:"contract_id"`
		ProjectName string  `json:"project_name"`
		StartTime   string  `json:"start_time"`
		EndTime     string  `json:"end_time"`
		Hours       float64 `json:"hours"`
		Billable    bool    `json:"billable"`
		Notes       string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		RespondError(w, "Invalid start_time format. Use RFC3339 format.", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		RespondError(w, "Invalid end_time format. Use RFC3339 format.", http.StatusBadRequest)
		return
	}

	if endTime.Before(startTime) {
		RespondError(w, "End time must be after start time", http.StatusBadRequest)
		return
	}

	duration := int(endTime.Sub(startTime).Seconds())
	hours := float64(duration) / 3600.0

	// If hours provided, use that instead
	if req.Hours > 0 {
		hours = req.Hours
		duration = int(hours * 3600)
	}

	session := models.TrackingSession{
		ClientID:    req.ClientID,
		ContractID:  req.ContractID,
		ProjectName: req.ProjectName,
		StartTime:   startTime,
		EndTime:     &endTime,
		Duration:    &duration,
		Hours:       &hours,
		Billable:    req.Billable,
		Notes:       req.Notes,
	}

	if err := db.Create(&session).Error; err != nil {
		RespondError(w, "Failed to create tracking session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, session, http.StatusCreated)
}

// Update handles PUT /api/v1/tracking/:id
func (c *TrackingController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.TrackingSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Tracking session not found", http.StatusNotFound)
		return
	}

	var req struct {
		ProjectName *string  `json:"project_name"`
		Billable    *bool    `json:"billable"`
		Notes       *string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.ProjectName != nil {
		session.ProjectName = *req.ProjectName
	}
	if req.Billable != nil {
		session.Billable = *req.Billable
	}
	if req.Notes != nil {
		session.Notes = *req.Notes
	}

	if err := db.Save(&session).Error; err != nil {
		RespondError(w, "Failed to update tracking session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, session, http.StatusOK)
}

// Delete handles DELETE /api/v1/tracking/:id
func (c *TrackingController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.TrackingSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Tracking session not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&session).Error; err != nil {
		RespondError(w, "Failed to delete tracking session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Tracking session deleted successfully"}, http.StatusOK)
}

// Active handles GET /api/v1/tracking/active
func (c *TrackingController) Active(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var session models.TrackingSession
	if err := db.Preload("Client").Preload("Contract").Where("end_time IS NULL").First(&session).Error; err != nil {
		RespondJSON(w, nil, http.StatusOK)
		return
	}

	RespondJSON(w, session, http.StatusOK)
}
