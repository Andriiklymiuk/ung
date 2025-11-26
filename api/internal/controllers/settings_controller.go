package controllers

import (
	"encoding/json"
	"net/http"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// SettingsController handles user settings endpoints
type SettingsController struct{}

// NewSettingsController creates a new settings controller
func NewSettingsController() *SettingsController {
	return &SettingsController{}
}

// Get handles GET /api/v1/settings
func (c *SettingsController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var settings models.UserSettings
	if err := db.First(&settings).Error; err != nil {
		// Create default settings if not found
		settings = models.UserSettings{
			HoursPerWeek:      40,
			WeeksPerYear:      48,
			DefaultTaxPercent: 25,
			DefaultMargin:     20,
			AnnualExpenses:    0,
			DefaultCurrency:   "USD",
		}
		if err := db.Create(&settings).Error; err != nil {
			RespondError(w, "Failed to create default settings: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	RespondJSON(w, settings, http.StatusOK)
}

// Update handles PUT /api/v1/settings
func (c *SettingsController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		HoursPerWeek      *float64 `json:"hours_per_week"`
		WeeksPerYear      *int     `json:"weeks_per_year"`
		DefaultTaxPercent *float64 `json:"default_tax_percent"`
		DefaultMargin     *float64 `json:"default_margin"`
		AnnualExpenses    *float64 `json:"annual_expenses"`
		DefaultCurrency   *string  `json:"default_currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get or create settings
	var settings models.UserSettings
	if err := db.First(&settings).Error; err != nil {
		settings = models.UserSettings{
			HoursPerWeek:      40,
			WeeksPerYear:      48,
			DefaultTaxPercent: 25,
			DefaultMargin:     20,
			AnnualExpenses:    0,
			DefaultCurrency:   "USD",
		}
		db.Create(&settings)
	}

	// Update fields
	if req.HoursPerWeek != nil {
		if *req.HoursPerWeek <= 0 || *req.HoursPerWeek > 168 {
			RespondError(w, "hours_per_week must be between 1 and 168", http.StatusBadRequest)
			return
		}
		settings.HoursPerWeek = *req.HoursPerWeek
	}
	if req.WeeksPerYear != nil {
		if *req.WeeksPerYear <= 0 || *req.WeeksPerYear > 52 {
			RespondError(w, "weeks_per_year must be between 1 and 52", http.StatusBadRequest)
			return
		}
		settings.WeeksPerYear = *req.WeeksPerYear
	}
	if req.DefaultTaxPercent != nil {
		if *req.DefaultTaxPercent < 0 || *req.DefaultTaxPercent > 100 {
			RespondError(w, "default_tax_percent must be between 0 and 100", http.StatusBadRequest)
			return
		}
		settings.DefaultTaxPercent = *req.DefaultTaxPercent
	}
	if req.DefaultMargin != nil {
		if *req.DefaultMargin < 0 || *req.DefaultMargin > 100 {
			RespondError(w, "default_margin must be between 0 and 100", http.StatusBadRequest)
			return
		}
		settings.DefaultMargin = *req.DefaultMargin
	}
	if req.AnnualExpenses != nil {
		if *req.AnnualExpenses < 0 {
			RespondError(w, "annual_expenses must be non-negative", http.StatusBadRequest)
			return
		}
		settings.AnnualExpenses = *req.AnnualExpenses
	}
	if req.DefaultCurrency != nil {
		settings.DefaultCurrency = *req.DefaultCurrency
	}

	if err := db.Save(&settings).Error; err != nil {
		RespondError(w, "Failed to update settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, settings, http.StatusOK)
}

// GetWorkingHours handles GET /api/v1/settings/working-hours
// Returns calculated working hours per month based on settings
func (c *SettingsController) GetWorkingHours(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var settings models.UserSettings
	if err := db.First(&settings).Error; err != nil {
		// Use defaults
		settings = models.UserSettings{
			HoursPerWeek: 40,
			WeeksPerYear: 48,
		}
	}

	hoursPerMonth := (settings.HoursPerWeek * float64(settings.WeeksPerYear)) / 12

	response := map[string]interface{}{
		"hours_per_week":  settings.HoursPerWeek,
		"weeks_per_year":  settings.WeeksPerYear,
		"hours_per_month": hoursPerMonth,
		"hours_per_year":  settings.HoursPerWeek * float64(settings.WeeksPerYear),
	}

	RespondJSON(w, response, http.StatusOK)
}
