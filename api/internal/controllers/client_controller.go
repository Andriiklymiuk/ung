package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// ClientController handles client endpoints
type ClientController struct{}

// NewClientController creates a new client controller
func NewClientController() *ClientController {
	return &ClientController{}
}

// List handles GET /api/v1/clients
func (c *ClientController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var clients []models.Client
	if err := db.Order("created_at DESC").Find(&clients).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, clients, http.StatusOK)
}

// Get handles GET /api/v1/clients/:id
func (c *ClientController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var client models.Client
	if err := db.First(&client, id).Error; err != nil {
		RespondError(w, "Client not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, client, http.StatusOK)
}

// Create handles POST /api/v1/clients
func (c *ClientController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Address string `json:"address"`
		TaxID   string `json:"tax_id"`
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

	client := models.Client{
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		TaxID:   req.TaxID,
	}

	if err := db.Create(&client).Error; err != nil {
		RespondError(w, "Failed to create client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, client, http.StatusCreated)
}

// Update handles PUT /api/v1/clients/:id
func (c *ClientController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var client models.Client
	if err := db.First(&client, id).Error; err != nil {
		RespondError(w, "Client not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Address string `json:"address"`
		TaxID   string `json:"tax_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != "" {
		client.Name = req.Name
	}
	if req.Email != "" {
		client.Email = req.Email
	}
	client.Address = req.Address
	client.TaxID = req.TaxID

	if err := db.Save(&client).Error; err != nil {
		RespondError(w, "Failed to update client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, client, http.StatusOK)
}

// Delete handles DELETE /api/v1/clients/:id
func (c *ClientController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var client models.Client
	if err := db.First(&client, id).Error; err != nil {
		RespondError(w, "Client not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&client).Error; err != nil {
		RespondError(w, "Failed to delete client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Client deleted successfully"}, http.StatusOK)
}
