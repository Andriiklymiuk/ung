package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// CompanyController handles company endpoints
type CompanyController struct{}

// NewCompanyController creates a new company controller
func NewCompanyController() *CompanyController {
	return &CompanyController{}
}

// List handles GET /api/v1/companies
func (c *CompanyController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var companies []models.Company
	if err := db.Order("created_at DESC").Find(&companies).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, companies, http.StatusOK)
}

// Get handles GET /api/v1/companies/:id
func (c *CompanyController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var company models.Company
	if err := db.First(&company, id).Error; err != nil {
		RespondError(w, "Company not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, company, http.StatusOK)
}

// Create handles POST /api/v1/companies
func (c *CompanyController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		Name                string `json:"name"`
		Email               string `json:"email"`
		Phone               string `json:"phone"`
		Address             string `json:"address"`
		RegistrationAddress string `json:"registration_address"`
		TaxID               string `json:"tax_id"`
		BankName            string `json:"bank_name"`
		BankAccount         string `json:"bank_account"`
		BankSWIFT           string `json:"bank_swift"`
		LogoPath            string `json:"logo_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		RespondError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		RespondError(w, "Email is required", http.StatusBadRequest)
		return
	}

	company := models.Company{
		Name:                req.Name,
		Email:               req.Email,
		Phone:               req.Phone,
		Address:             req.Address,
		RegistrationAddress: req.RegistrationAddress,
		TaxID:               req.TaxID,
		BankName:            req.BankName,
		BankAccount:         req.BankAccount,
		BankSWIFT:           req.BankSWIFT,
		LogoPath:            req.LogoPath,
	}

	if err := db.Create(&company).Error; err != nil {
		RespondError(w, "Failed to create company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, company, http.StatusCreated)
}

// Update handles PUT /api/v1/companies/:id
func (c *CompanyController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var company models.Company
	if err := db.First(&company, id).Error; err != nil {
		RespondError(w, "Company not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name                *string `json:"name"`
		Email               *string `json:"email"`
		Phone               *string `json:"phone"`
		Address             *string `json:"address"`
		RegistrationAddress *string `json:"registration_address"`
		TaxID               *string `json:"tax_id"`
		BankName            *string `json:"bank_name"`
		BankAccount         *string `json:"bank_account"`
		BankSWIFT           *string `json:"bank_swift"`
		LogoPath            *string `json:"logo_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != nil {
		company.Name = *req.Name
	}
	if req.Email != nil {
		company.Email = *req.Email
	}
	if req.Phone != nil {
		company.Phone = *req.Phone
	}
	if req.Address != nil {
		company.Address = *req.Address
	}
	if req.RegistrationAddress != nil {
		company.RegistrationAddress = *req.RegistrationAddress
	}
	if req.TaxID != nil {
		company.TaxID = *req.TaxID
	}
	if req.BankName != nil {
		company.BankName = *req.BankName
	}
	if req.BankAccount != nil {
		company.BankAccount = *req.BankAccount
	}
	if req.BankSWIFT != nil {
		company.BankSWIFT = *req.BankSWIFT
	}
	if req.LogoPath != nil {
		company.LogoPath = *req.LogoPath
	}

	if err := db.Save(&company).Error; err != nil {
		RespondError(w, "Failed to update company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, company, http.StatusOK)
}

// Delete handles DELETE /api/v1/companies/:id
func (c *CompanyController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var company models.Company
	if err := db.First(&company, id).Error; err != nil {
		RespondError(w, "Company not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&company).Error; err != nil {
		RespondError(w, "Failed to delete company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Company deleted successfully"}, http.StatusOK)
}
