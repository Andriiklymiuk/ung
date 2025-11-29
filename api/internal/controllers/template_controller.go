package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// TemplateController handles invoice template endpoints
type TemplateController struct{}

// NewTemplateController creates a new template controller
func NewTemplateController() *TemplateController {
	return &TemplateController{}
}

// List handles GET /api/v1/templates
func (c *TemplateController) List(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var templates []models.InvoiceTemplate
	if err := db.Order("created_at DESC").Find(&templates).Error; err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, templates, http.StatusOK)
}

// Get handles GET /api/v1/templates/:id
func (c *TemplateController) Get(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var template models.InvoiceTemplate
	if err := db.First(&template, id).Error; err != nil {
		RespondError(w, "Template not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, template, http.StatusOK)
}

// Create handles POST /api/v1/templates
func (c *TemplateController) Create(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
		IsDefault   bool   `json:"is_default"`
		Variables   string `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		RespondError(w, "Template name is required", http.StatusBadRequest)
		return
	}

	// If this template is set as default, unset other defaults
	if req.IsDefault {
		db.Model(&models.InvoiceTemplate{}).Where("is_default = ?", true).Update("is_default", false)
	}

	template := models.InvoiceTemplate{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		IsDefault:   req.IsDefault,
		Variables:   req.Variables,
	}

	if err := db.Create(&template).Error; err != nil {
		RespondError(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, template, http.StatusCreated)
}

// Update handles PUT /api/v1/templates/:id
func (c *TemplateController) Update(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var template models.InvoiceTemplate
	if err := db.First(&template, id).Error; err != nil {
		RespondError(w, "Template not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Content     *string `json:"content"`
		IsDefault   *bool   `json:"is_default"`
		Variables   *string `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.Content != nil {
		template.Content = *req.Content
	}
	if req.Variables != nil {
		template.Variables = *req.Variables
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			// Unset other defaults
			db.Model(&models.InvoiceTemplate{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false)
		}
		template.IsDefault = *req.IsDefault
	}

	if err := db.Save(&template).Error; err != nil {
		RespondError(w, "Failed to update template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, template, http.StatusOK)
}

// Delete handles DELETE /api/v1/templates/:id
func (c *TemplateController) Delete(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var template models.InvoiceTemplate
	if err := db.First(&template, id).Error; err != nil {
		RespondError(w, "Template not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&template).Error; err != nil {
		RespondError(w, "Failed to delete template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Template deleted successfully"}, http.StatusOK)
}

// GetDefault handles GET /api/v1/templates/default
func (c *TemplateController) GetDefault(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var template models.InvoiceTemplate
	if err := db.Where("is_default = ?", true).First(&template).Error; err != nil {
		// Return a system default template if no custom default exists
		template = models.InvoiceTemplate{
			Name:        "Default Template",
			Description: "System default invoice template",
			Content:     getDefaultTemplateContent(),
			IsDefault:   true,
			Variables:   `["company_name", "company_address", "client_name", "client_address", "invoice_number", "invoice_date", "due_date", "amount", "currency", "description", "line_items"]`,
		}
	}

	RespondJSON(w, template, http.StatusOK)
}

// Preview handles POST /api/v1/templates/:id/preview
func (c *TemplateController) Preview(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var template models.InvoiceTemplate
	if err := db.First(&template, id).Error; err != nil {
		RespondError(w, "Template not found", http.StatusNotFound)
		return
	}

	var sampleData struct {
		Variables map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&sampleData); err != nil {
		// Use default sample data
		sampleData.Variables = getDefaultSampleData()
	}

	// Simple variable substitution for preview
	preview := template.Content
	// In a real implementation, you'd use a templating engine

	response := map[string]interface{}{
		"template": template,
		"preview":  preview,
		"data":     sampleData.Variables,
	}

	RespondJSON(w, response, http.StatusOK)
}

func getDefaultTemplateContent() string {
	return `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { display: flex; justify-content: space-between; margin-bottom: 40px; }
        .company-info { text-align: left; }
        .invoice-info { text-align: right; }
        .client-info { margin-bottom: 30px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f5f5f5; }
        .total { font-weight: bold; text-align: right; font-size: 1.2em; margin-top: 20px; }
        .footer { margin-top: 40px; text-align: center; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <div class="company-info">
            <h1>{{company_name}}</h1>
            <p>{{company_address}}</p>
        </div>
        <div class="invoice-info">
            <h2>INVOICE</h2>
            <p><strong>Invoice #:</strong> {{invoice_number}}</p>
            <p><strong>Date:</strong> {{invoice_date}}</p>
            <p><strong>Due:</strong> {{due_date}}</p>
        </div>
    </div>

    <div class="client-info">
        <h3>Bill To:</h3>
        <p><strong>{{client_name}}</strong></p>
        <p>{{client_address}}</p>
    </div>

    <table>
        <thead>
            <tr>
                <th>Description</th>
                <th>Quantity</th>
                <th>Rate</th>
                <th>Amount</th>
            </tr>
        </thead>
        <tbody>
            {{line_items}}
        </tbody>
    </table>

    <div class="total">
        Total: {{currency}} {{amount}}
    </div>

    <div class="footer">
        <p>Thank you for your business!</p>
    </div>
</body>
</html>`
}

func getDefaultSampleData() map[string]interface{} {
	return map[string]interface{}{
		"company_name":    "Your Company Name",
		"company_address": "123 Business St, City, Country",
		"client_name":     "Sample Client",
		"client_address":  "456 Client Ave, City, Country",
		"invoice_number":  "INV-2024-001",
		"invoice_date":    "2024-01-15",
		"due_date":        "2024-02-14",
		"amount":          "1,500.00",
		"currency":        "USD",
		"description":     "Professional Services",
		"line_items":      "<tr><td>Consulting Services</td><td>10</td><td>150.00</td><td>1,500.00</td></tr>",
	}
}
