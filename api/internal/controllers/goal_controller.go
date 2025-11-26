package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// GoalController handles income goal endpoints
type GoalController struct{}

// NewGoalController creates a new goal controller
func NewGoalController() *GoalController {
	return &GoalController{}
}

// List handles GET /api/v1/goals
func (c *GoalController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Optional filter by period
	period := r.URL.Query().Get("period")

	var goals []models.IncomeGoal
	query := db.Order("year DESC, period, month DESC, quarter DESC")
	if period != "" {
		query = query.Where("period = ?", period)
	}

	if err := query.Find(&goals).Error; err != nil {
		RespondError(w, "Failed to list goals: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, goals, http.StatusOK)
}

// Get handles GET /api/v1/goals/:id
func (c *GoalController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var goal models.IncomeGoal
	if err := db.First(&goal, id).Error; err != nil {
		RespondError(w, "Goal not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, goal, http.StatusOK)
}

// Create handles POST /api/v1/goals
func (c *GoalController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		Amount      float64 `json:"amount"`
		Period      string  `json:"period"`      // monthly, quarterly, yearly
		Year        *int    `json:"year"`        // defaults to current year
		Month       *int    `json:"month"`       // for monthly goals (1-12)
		Quarter     *int    `json:"quarter"`     // for quarterly goals (1-4)
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		RespondError(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	if req.Period != "monthly" && req.Period != "quarterly" && req.Period != "yearly" {
		RespondError(w, "Period must be monthly, quarterly, or yearly", http.StatusBadRequest)
		return
	}

	now := time.Now()
	year := now.Year()
	if req.Year != nil {
		year = *req.Year
	}

	month := 0
	if req.Period == "monthly" {
		if req.Month != nil {
			month = *req.Month
		} else {
			month = int(now.Month())
		}
		if month < 1 || month > 12 {
			RespondError(w, "Month must be between 1 and 12", http.StatusBadRequest)
			return
		}
	}

	quarter := 0
	if req.Period == "quarterly" {
		if req.Quarter != nil {
			quarter = *req.Quarter
		} else {
			quarter = (int(now.Month())-1)/3 + 1
		}
		if quarter < 1 || quarter > 4 {
			RespondError(w, "Quarter must be between 1 and 4", http.StatusBadRequest)
			return
		}
	}

	// Check for existing goal
	var existing models.IncomeGoal
	query := db.Where("period = ? AND year = ?", req.Period, year)
	if req.Period == "monthly" {
		query = query.Where("month = ?", month)
	} else if req.Period == "quarterly" {
		query = query.Where("quarter = ?", quarter)
	}

	if query.First(&existing).Error == nil {
		// Update existing
		existing.Amount = req.Amount
		existing.Description = req.Description
		if err := db.Save(&existing).Error; err != nil {
			RespondError(w, "Failed to update goal: "+err.Error(), http.StatusInternalServerError)
			return
		}
		RespondJSON(w, existing, http.StatusOK)
		return
	}

	// Create new
	goal := models.IncomeGoal{
		Amount:      req.Amount,
		Period:      req.Period,
		Year:        year,
		Month:       month,
		Quarter:     quarter,
		Description: req.Description,
	}

	if err := db.Create(&goal).Error; err != nil {
		RespondError(w, "Failed to create goal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, goal, http.StatusCreated)
}

// Update handles PUT /api/v1/goals/:id
func (c *GoalController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var goal models.IncomeGoal
	if err := db.First(&goal, id).Error; err != nil {
		RespondError(w, "Goal not found", http.StatusNotFound)
		return
	}

	var req struct {
		Amount      *float64 `json:"amount"`
		Description *string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			RespondError(w, "Amount must be positive", http.StatusBadRequest)
			return
		}
		goal.Amount = *req.Amount
	}
	if req.Description != nil {
		goal.Description = *req.Description
	}

	if err := db.Save(&goal).Error; err != nil {
		RespondError(w, "Failed to update goal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, goal, http.StatusOK)
}

// Delete handles DELETE /api/v1/goals/:id
func (c *GoalController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var goal models.IncomeGoal
	if err := db.First(&goal, id).Error; err != nil {
		RespondError(w, "Goal not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&goal).Error; err != nil {
		RespondError(w, "Failed to delete goal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Goal deleted successfully"}, http.StatusOK)
}

// GoalProgress represents progress toward a goal
type GoalProgress struct {
	Goal         models.IncomeGoal `json:"goal"`
	Actual       float64           `json:"actual"`
	Progress     float64           `json:"progress_percent"`
	Remaining    float64           `json:"remaining"`
	DaysLeft     int               `json:"days_left"`
	DailyNeeded  float64           `json:"daily_needed"`
	IsAchieved   bool              `json:"is_achieved"`
	PeriodStart  time.Time         `json:"period_start"`
	PeriodEnd    time.Time         `json:"period_end"`
}

// Status handles GET /api/v1/goals/status
func (c *GoalController) Status(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Optional filter by period
	period := r.URL.Query().Get("period")

	var goals []models.IncomeGoal
	query := db.Order("year DESC, period, month DESC, quarter DESC")
	if period != "" {
		query = query.Where("period = ?", period)
	}

	if err := query.Find(&goals).Error; err != nil {
		RespondError(w, "Failed to list goals: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var progressList []GoalProgress

	for _, goal := range goals {
		startDate, endDate := getGoalDateRange(goal)
		actual := getIncomeForPeriod(db, startDate, endDate)

		progress := GoalProgress{
			Goal:        goal,
			Actual:      actual,
			PeriodStart: startDate,
			PeriodEnd:   endDate,
		}

		if goal.Amount > 0 {
			progress.Progress = (actual / goal.Amount) * 100
			if progress.Progress > 100 {
				progress.Progress = 100
			}
		}

		progress.Remaining = goal.Amount - actual
		if progress.Remaining < 0 {
			progress.Remaining = 0
			progress.IsAchieved = true
		}

		progress.DaysLeft = getDaysRemaining(endDate)
		if progress.DaysLeft > 0 && progress.Remaining > 0 {
			progress.DailyNeeded = progress.Remaining / float64(progress.DaysLeft)
		}

		progressList = append(progressList, progress)
	}

	RespondJSON(w, progressList, http.StatusOK)
}

func getGoalDateRange(g models.IncomeGoal) (time.Time, time.Time) {
	loc := time.Now().Location()

	switch g.Period {
	case "monthly":
		start := time.Date(g.Year, time.Month(g.Month), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0)
		return start, end

	case "quarterly":
		startMonth := (g.Quarter-1)*3 + 1
		start := time.Date(g.Year, time.Month(startMonth), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 3, 0)
		return start, end

	case "yearly":
		start := time.Date(g.Year, 1, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(1, 0, 0)
		return start, end
	}

	return time.Time{}, time.Time{}
}

func getIncomeForPeriod(db *gorm.DB, start, end time.Time) float64 {
	var result struct {
		Total float64
	}
	db.Raw(`
		SELECT COALESCE(SUM(amount), 0) as total
		FROM invoices
		WHERE status = ?
		  AND updated_at >= ?
		  AND updated_at < ?
	`, models.StatusPaid, start, end).Scan(&result)

	return result.Total
}

func getDaysRemaining(end time.Time) int {
	now := time.Now()
	if now.After(end) {
		return 0
	}
	return int(end.Sub(now).Hours() / 24)
}
