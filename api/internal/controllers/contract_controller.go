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

// ContractController handles contract endpoints
type ContractController struct{}

// NewContractController creates a new contract controller
func NewContractController() *ContractController {
	return &ContractController{}
}

// List handles GET /api/v1/contracts
func (c *ContractController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var contracts []models.Contract
	if err := db.Preload("Client").Order("created_at DESC").Find(&contracts).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, contracts, http.StatusOK)
}

// Get handles GET /api/v1/contracts/:id
func (c *ContractController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var contract models.Contract
	if err := db.Preload("Client").First(&contract, id).Error; err != nil {
		RespondError(w, "Contract not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, contract, http.StatusOK)
}

// Create handles POST /api/v1/contracts
func (c *ContractController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		ContractNum  string               `json:"contract_num"`
		ClientID     uint                 `json:"client_id"`
		Name         string               `json:"name"`
		ContractType models.ContractType  `json:"contract_type"`
		HourlyRate   *float64             `json:"hourly_rate"`
		FixedPrice   *float64             `json:"fixed_price"`
		Currency     string               `json:"currency"`
		StartDate    string               `json:"start_date"`
		EndDate      *string              `json:"end_date"`
		Active       bool                 `json:"active"`
		Notes        string               `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.ContractNum == "" {
		RespondError(w, "Contract number is required", http.StatusBadRequest)
		return
	}
	if req.ClientID == 0 {
		RespondError(w, "Client ID is required", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		RespondError(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Parse dates
	startDate := time.Now()
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = parsed
		}
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		if parsed, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			endDate = &parsed
		}
	}

	contract := models.Contract{
		ContractNum:  req.ContractNum,
		ClientID:     req.ClientID,
		Name:         req.Name,
		ContractType: req.ContractType,
		HourlyRate:   req.HourlyRate,
		FixedPrice:   req.FixedPrice,
		Currency:     req.Currency,
		StartDate:    startDate,
		EndDate:      endDate,
		Active:       req.Active,
		Notes:        req.Notes,
	}

	if contract.Currency == "" {
		contract.Currency = "USD"
	}

	if err := db.Create(&contract).Error; err != nil {
		RespondError(w, "Failed to create contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, contract, http.StatusCreated)
}

// Update handles PUT /api/v1/contracts/:id
func (c *ContractController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var contract models.Contract
	if err := db.First(&contract, id).Error; err != nil {
		RespondError(w, "Contract not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name         *string              `json:"name"`
		ContractType *models.ContractType `json:"contract_type"`
		HourlyRate   *float64             `json:"hourly_rate"`
		FixedPrice   *float64             `json:"fixed_price"`
		Currency     *string              `json:"currency"`
		EndDate      *string              `json:"end_date"`
		Active       *bool                `json:"active"`
		Notes        *string              `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != nil {
		contract.Name = *req.Name
	}
	if req.ContractType != nil {
		contract.ContractType = *req.ContractType
	}
	if req.HourlyRate != nil {
		contract.HourlyRate = req.HourlyRate
	}
	if req.FixedPrice != nil {
		contract.FixedPrice = req.FixedPrice
	}
	if req.Currency != nil {
		contract.Currency = *req.Currency
	}
	if req.Active != nil {
		contract.Active = *req.Active
	}
	if req.Notes != nil {
		contract.Notes = *req.Notes
	}
	if req.EndDate != nil && *req.EndDate != "" {
		if parsed, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			contract.EndDate = &parsed
		}
	}

	if err := db.Save(&contract).Error; err != nil {
		RespondError(w, "Failed to update contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, contract, http.StatusOK)
}

// Delete handles DELETE /api/v1/contracts/:id
func (c *ContractController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var contract models.Contract
	if err := db.First(&contract, id).Error; err != nil {
		RespondError(w, "Contract not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&contract).Error; err != nil {
		RespondError(w, "Failed to delete contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Contract deleted successfully"}, http.StatusOK)
}
