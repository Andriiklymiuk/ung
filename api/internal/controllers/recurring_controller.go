package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// RecurringController handles recurring invoice endpoints
type RecurringController struct{}

// NewRecurringController creates a new recurring invoice controller
func NewRecurringController() *RecurringController {
	return &RecurringController{}
}

// List handles GET /api/v1/recurring
func (c *RecurringController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var recurring []models.RecurringInvoice
	if err := db.Preload("Client").Preload("Company").Order("created_at DESC").Find(&recurring).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, recurring, http.StatusOK)
}

// Get handles GET /api/v1/recurring/:id
func (c *RecurringController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var recurring models.RecurringInvoice
	if err := db.Preload("Client").Preload("Company").First(&recurring, id).Error; err != nil {
		RespondError(w, "Recurring invoice not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, recurring, http.StatusOK)
}

// Create handles POST /api/v1/recurring
func (c *RecurringController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		ClientID      uint                      `json:"client_id"`
		CompanyID     uint                      `json:"company_id"`
		Amount        float64                   `json:"amount"`
		Currency      string                    `json:"currency"`
		Description   string                    `json:"description"`
		Frequency     models.RecurringFrequency `json:"frequency"`
		DayOfMonth    int                       `json:"day_of_month"`
		DayOfWeek     int                       `json:"day_of_week"`
		InvoicePrefix string                    `json:"invoice_prefix"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ClientID == 0 {
		RespondError(w, "Client ID is required", http.StatusBadRequest)
		return
	}
	if req.CompanyID == 0 {
		RespondError(w, "Company ID is required", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		RespondError(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	// Calculate next run date based on frequency
	nextRunDate := calculateNextRunDate(req.Frequency, req.DayOfMonth, req.DayOfWeek)

	recurring := models.RecurringInvoice{
		ClientID:      req.ClientID,
		CompanyID:     req.CompanyID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Description:   req.Description,
		Frequency:     req.Frequency,
		DayOfMonth:    req.DayOfMonth,
		DayOfWeek:     req.DayOfWeek,
		NextRunDate:   nextRunDate,
		Active:        true,
		InvoicePrefix: req.InvoicePrefix,
	}

	if recurring.Currency == "" {
		recurring.Currency = "USD"
	}
	if recurring.InvoicePrefix == "" {
		recurring.InvoicePrefix = "REC"
	}

	if err := db.Create(&recurring).Error; err != nil {
		RespondError(w, "Failed to create recurring invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Preload relations for response
	db.Preload("Client").Preload("Company").First(&recurring, recurring.ID)

	RespondJSON(w, recurring, http.StatusCreated)
}

// Update handles PUT /api/v1/recurring/:id
func (c *RecurringController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var recurring models.RecurringInvoice
	if err := db.First(&recurring, id).Error; err != nil {
		RespondError(w, "Recurring invoice not found", http.StatusNotFound)
		return
	}

	var req struct {
		Amount        *float64                   `json:"amount"`
		Currency      *string                    `json:"currency"`
		Description   *string                    `json:"description"`
		Frequency     *models.RecurringFrequency `json:"frequency"`
		DayOfMonth    *int                       `json:"day_of_month"`
		DayOfWeek     *int                       `json:"day_of_week"`
		InvoicePrefix *string                    `json:"invoice_prefix"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount != nil {
		recurring.Amount = *req.Amount
	}
	if req.Currency != nil {
		recurring.Currency = *req.Currency
	}
	if req.Description != nil {
		recurring.Description = *req.Description
	}
	if req.InvoicePrefix != nil {
		recurring.InvoicePrefix = *req.InvoicePrefix
	}

	// Recalculate next run date if frequency changed
	if req.Frequency != nil || req.DayOfMonth != nil || req.DayOfWeek != nil {
		freq := recurring.Frequency
		dayOfMonth := recurring.DayOfMonth
		dayOfWeek := recurring.DayOfWeek

		if req.Frequency != nil {
			freq = *req.Frequency
			recurring.Frequency = freq
		}
		if req.DayOfMonth != nil {
			dayOfMonth = *req.DayOfMonth
			recurring.DayOfMonth = dayOfMonth
		}
		if req.DayOfWeek != nil {
			dayOfWeek = *req.DayOfWeek
			recurring.DayOfWeek = dayOfWeek
		}

		recurring.NextRunDate = calculateNextRunDate(freq, dayOfMonth, dayOfWeek)
	}

	if err := db.Save(&recurring).Error; err != nil {
		RespondError(w, "Failed to update recurring invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db.Preload("Client").Preload("Company").First(&recurring, recurring.ID)
	RespondJSON(w, recurring, http.StatusOK)
}

// Delete handles DELETE /api/v1/recurring/:id
func (c *RecurringController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var recurring models.RecurringInvoice
	if err := db.First(&recurring, id).Error; err != nil {
		RespondError(w, "Recurring invoice not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&recurring).Error; err != nil {
		RespondError(w, "Failed to delete recurring invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Recurring invoice deleted successfully"}, http.StatusOK)
}

// Pause handles POST /api/v1/recurring/:id/pause
func (c *RecurringController) Pause(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var recurring models.RecurringInvoice
	if err := db.First(&recurring, id).Error; err != nil {
		RespondError(w, "Recurring invoice not found", http.StatusNotFound)
		return
	}

	recurring.Active = false
	if err := db.Save(&recurring).Error; err != nil {
		RespondError(w, "Failed to pause recurring invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, recurring, http.StatusOK)
}

// Resume handles POST /api/v1/recurring/:id/resume
func (c *RecurringController) Resume(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var recurring models.RecurringInvoice
	if err := db.First(&recurring, id).Error; err != nil {
		RespondError(w, "Recurring invoice not found", http.StatusNotFound)
		return
	}

	recurring.Active = true
	// Recalculate next run date when resuming
	recurring.NextRunDate = calculateNextRunDate(recurring.Frequency, recurring.DayOfMonth, recurring.DayOfWeek)

	if err := db.Save(&recurring).Error; err != nil {
		RespondError(w, "Failed to resume recurring invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, recurring, http.StatusOK)
}

// Generate handles POST /api/v1/recurring/generate - generates all due recurring invoices
func (c *RecurringController) Generate(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	now := time.Now()
	var dueRecurring []models.RecurringInvoice
	if err := db.Where("active = ? AND next_run_date <= ?", true, now).Find(&dueRecurring).Error; err != nil {
		RespondError(w, "Failed to fetch due recurring invoices: "+err.Error(), http.StatusInternalServerError)
		return
	}

	generated := 0
	var createdInvoices []models.Invoice

	for _, rec := range dueRecurring {
		// Generate invoice number
		invoiceNum := fmt.Sprintf("%s-%d-%03d", rec.InvoicePrefix, time.Now().Year(), rec.TotalGenerated+1)

		invoice := models.Invoice{
			InvoiceNum:  invoiceNum,
			CompanyID:   rec.CompanyID,
			Amount:      rec.Amount,
			Currency:    rec.Currency,
			Description: rec.Description,
			Status:      models.StatusPending,
			IssuedDate:  now,
			DueDate:     now.AddDate(0, 0, 30),
		}

		if err := db.Create(&invoice).Error; err != nil {
			continue
		}

		// Create invoice recipient
		db.Create(&models.InvoiceRecipient{
			InvoiceID: invoice.ID,
			ClientID:  rec.ClientID,
		})

		// Update recurring invoice
		rec.LastRunDate = &now
		rec.NextRunDate = calculateNextRunDate(rec.Frequency, rec.DayOfMonth, rec.DayOfWeek)
		rec.TotalGenerated++
		db.Save(&rec)

		createdInvoices = append(createdInvoices, invoice)
		generated++
	}

	response := map[string]interface{}{
		"generated": generated,
		"invoices":  createdInvoices,
	}

	RespondJSON(w, response, http.StatusOK)
}

// GenerateSingle handles POST /api/v1/recurring/:id/generate - generates a single invoice from recurring template
func (c *RecurringController) GenerateSingle(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var rec models.RecurringInvoice
	if err := db.First(&rec, id).Error; err != nil {
		RespondError(w, "Recurring invoice not found", http.StatusNotFound)
		return
	}

	now := time.Now()

	// Generate invoice number
	invoiceNum := fmt.Sprintf("%s-%d-%03d", rec.InvoicePrefix, time.Now().Year(), rec.TotalGenerated+1)

	invoice := models.Invoice{
		InvoiceNum:  invoiceNum,
		CompanyID:   rec.CompanyID,
		Amount:      rec.Amount,
		Currency:    rec.Currency,
		Description: rec.Description,
		Status:      models.StatusPending,
		IssuedDate:  now,
		DueDate:     now.AddDate(0, 0, 30),
	}

	if err := db.Create(&invoice).Error; err != nil {
		RespondError(w, "Failed to create invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create invoice recipient
	db.Create(&models.InvoiceRecipient{
		InvoiceID: invoice.ID,
		ClientID:  rec.ClientID,
	})

	// Update recurring invoice
	rec.LastRunDate = &now
	rec.NextRunDate = calculateNextRunDate(rec.Frequency, rec.DayOfMonth, rec.DayOfWeek)
	rec.TotalGenerated++
	db.Save(&rec)

	RespondJSON(w, invoice, http.StatusCreated)
}

func calculateNextRunDate(frequency models.RecurringFrequency, dayOfMonth, dayOfWeek int) time.Time {
	now := time.Now()

	switch frequency {
	case models.FrequencyWeekly:
		daysUntilNext := (dayOfWeek - int(now.Weekday()) + 7) % 7
		if daysUntilNext == 0 {
			daysUntilNext = 7
		}
		return now.AddDate(0, 0, daysUntilNext)

	case models.FrequencyBiweekly:
		daysUntilNext := (dayOfWeek - int(now.Weekday()) + 7) % 7
		if daysUntilNext == 0 {
			daysUntilNext = 14
		}
		return now.AddDate(0, 0, daysUntilNext)

	case models.FrequencyMonthly:
		nextMonth := time.Date(now.Year(), now.Month()+1, dayOfMonth, 0, 0, 0, 0, now.Location())
		if dayOfMonth > 28 {
			// Handle months with fewer days
			lastDay := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
			if dayOfMonth > lastDay {
				nextMonth = time.Date(nextMonth.Year(), nextMonth.Month(), lastDay, 0, 0, 0, 0, now.Location())
			}
		}
		return nextMonth

	case models.FrequencyQuarterly:
		currentQuarter := (int(now.Month()) - 1) / 3
		nextQuarterStart := time.Date(now.Year(), time.Month((currentQuarter+1)*3+1), dayOfMonth, 0, 0, 0, 0, now.Location())
		if nextQuarterStart.Before(now) || nextQuarterStart.Equal(now) {
			nextQuarterStart = nextQuarterStart.AddDate(0, 3, 0)
		}
		return nextQuarterStart

	case models.FrequencyYearly:
		nextYear := time.Date(now.Year()+1, time.January, dayOfMonth, 0, 0, 0, 0, now.Location())
		return nextYear

	default:
		return now.AddDate(0, 1, 0)
	}
}
