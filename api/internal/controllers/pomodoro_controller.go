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

// PomodoroController handles pomodoro timer endpoints
type PomodoroController struct{}

// NewPomodoroController creates a new pomodoro controller
func NewPomodoroController() *PomodoroController {
	return &PomodoroController{}
}

// List handles GET /api/v1/pomodoro
func (c *PomodoroController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Optional query params
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	dateStr := r.URL.Query().Get("date")
	var sessions []models.PomodoroSession

	query := db.Preload("Client").Preload("Contract").Order("start_time DESC").Limit(limit)

	if dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)
			query = query.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
		}
	}

	if err := query.Find(&sessions).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, sessions, http.StatusOK)
}

// Get handles GET /api/v1/pomodoro/:id
func (c *PomodoroController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.PomodoroSession
	if err := db.Preload("Client").Preload("Contract").First(&session, id).Error; err != nil {
		RespondError(w, "Pomodoro session not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, session, http.StatusOK)
}

// Start handles POST /api/v1/pomodoro/start
func (c *PomodoroController) Start(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Check for active session
	var activeSession models.PomodoroSession
	if err := db.Where("end_time IS NULL").First(&activeSession).Error; err == nil {
		RespondError(w, "There is already an active pomodoro session", http.StatusConflict)
		return
	}

	var req struct {
		ContractID  *uint  `json:"contract_id"`
		ClientID    *uint  `json:"client_id"`
		ProjectName string `json:"project_name"`
		Duration    int    `json:"duration"`
		BreakTime   int    `json:"break_time"`
		Notes       string `json:"notes"`
		SessionType string `json:"session_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := models.PomodoroSession{
		ContractID:  req.ContractID,
		ClientID:    req.ClientID,
		ProjectName: req.ProjectName,
		Duration:    req.Duration,
		BreakTime:   req.BreakTime,
		StartTime:   time.Now(),
		Notes:       req.Notes,
		SessionType: req.SessionType,
		Completed:   false,
	}

	if session.Duration <= 0 {
		session.Duration = 25 // Default 25 minutes
	}
	if session.BreakTime <= 0 {
		session.BreakTime = 5 // Default 5 minutes
	}
	if session.SessionType == "" {
		session.SessionType = "work"
	}

	if err := db.Create(&session).Error; err != nil {
		RespondError(w, "Failed to start pomodoro session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db.Preload("Client").Preload("Contract").First(&session, session.ID)
	RespondJSON(w, session, http.StatusCreated)
}

// Stop handles POST /api/v1/pomodoro/:id/stop
func (c *PomodoroController) Stop(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.PomodoroSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Pomodoro session not found", http.StatusNotFound)
		return
	}

	if session.EndTime != nil {
		RespondError(w, "Session is already stopped", http.StatusBadRequest)
		return
	}

	now := time.Now()
	session.EndTime = &now

	// Check if completed (ran for at least 80% of duration)
	elapsed := now.Sub(session.StartTime)
	expectedDuration := time.Duration(session.Duration) * time.Minute
	if elapsed >= time.Duration(float64(expectedDuration)*0.8) {
		session.Completed = true
	}

	if err := db.Save(&session).Error; err != nil {
		RespondError(w, "Failed to stop pomodoro session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db.Preload("Client").Preload("Contract").First(&session, session.ID)
	RespondJSON(w, session, http.StatusOK)
}

// Complete handles POST /api/v1/pomodoro/:id/complete
func (c *PomodoroController) Complete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.PomodoroSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Pomodoro session not found", http.StatusNotFound)
		return
	}

	now := time.Now()
	session.EndTime = &now
	session.Completed = true

	if err := db.Save(&session).Error; err != nil {
		RespondError(w, "Failed to complete pomodoro session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db.Preload("Client").Preload("Contract").First(&session, session.ID)
	RespondJSON(w, session, http.StatusOK)
}

// Active handles GET /api/v1/pomodoro/active
func (c *PomodoroController) Active(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var session models.PomodoroSession
	if err := db.Preload("Client").Preload("Contract").Where("end_time IS NULL").First(&session).Error; err != nil {
		RespondJSON(w, nil, http.StatusOK)
		return
	}

	// Calculate time remaining
	elapsed := time.Since(session.StartTime)
	totalDuration := time.Duration(session.Duration) * time.Minute
	remaining := totalDuration - elapsed

	response := map[string]interface{}{
		"session":          session,
		"elapsed_seconds":  int(elapsed.Seconds()),
		"remaining_seconds": int(remaining.Seconds()),
		"progress_percent": float64(elapsed) / float64(totalDuration) * 100,
	}

	RespondJSON(w, response, http.StatusOK)
}

// Delete handles DELETE /api/v1/pomodoro/:id
func (c *PomodoroController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var session models.PomodoroSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Pomodoro session not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&session).Error; err != nil {
		RespondError(w, "Failed to delete pomodoro session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Pomodoro session deleted successfully"}, http.StatusOK)
}

// Stats handles GET /api/v1/pomodoro/stats
func (c *PomodoroController) Stats(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	stats := map[string]interface{}{}

	// Today's completed pomodoros
	var todayCount int64
	db.Model(&models.PomodoroSession{}).
		Where("start_time >= ? AND completed = ?", todayStart, true).
		Count(&todayCount)
	stats["today_completed"] = todayCount

	// Today's total minutes
	var todaySessions []models.PomodoroSession
	db.Where("start_time >= ? AND completed = ?", todayStart, true).Find(&todaySessions)
	var todayMinutes int
	for _, s := range todaySessions {
		todayMinutes += s.Duration
	}
	stats["today_minutes"] = todayMinutes

	// This week's completed
	var weekCount int64
	db.Model(&models.PomodoroSession{}).
		Where("start_time >= ? AND completed = ?", weekStart, true).
		Count(&weekCount)
	stats["week_completed"] = weekCount

	// This month's completed
	var monthCount int64
	db.Model(&models.PomodoroSession{}).
		Where("start_time >= ? AND completed = ?", monthStart, true).
		Count(&monthCount)
	stats["month_completed"] = monthCount

	// All time completed
	var totalCount int64
	db.Model(&models.PomodoroSession{}).Where("completed = ?", true).Count(&totalCount)
	stats["total_completed"] = totalCount

	// Current streak (consecutive days with at least one completed pomodoro)
	streak := 0
	checkDate := todayStart
	for {
		var count int64
		nextDay := checkDate.Add(24 * time.Hour)
		db.Model(&models.PomodoroSession{}).
			Where("start_time >= ? AND start_time < ? AND completed = ?", checkDate, nextDay, true).
			Count(&count)

		if count > 0 {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else if checkDate.Before(todayStart) {
			// If checking previous days and no pomodoros, streak ends
			break
		} else {
			// Today might not have any yet, check yesterday
			checkDate = checkDate.AddDate(0, 0, -1)
		}

		// Safety limit
		if streak > 365 {
			break
		}
	}
	stats["current_streak"] = streak

	// Average pomodoros per day (last 30 days)
	thirtyDaysAgo := todayStart.AddDate(0, 0, -30)
	var last30Count int64
	db.Model(&models.PomodoroSession{}).
		Where("start_time >= ? AND completed = ?", thirtyDaysAgo, true).
		Count(&last30Count)
	stats["avg_daily_30d"] = float64(last30Count) / 30.0

	RespondJSON(w, stats, http.StatusOK)
}
