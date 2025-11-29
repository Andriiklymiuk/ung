package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// ExportController handles data export/import endpoints
type ExportController struct{}

// NewExportController creates a new export controller
func NewExportController() *ExportController {
	return &ExportController{}
}

// ExportInvoicesCSV handles GET /api/v1/export/invoices/csv
func (c *ExportController) ExportInvoicesCSV(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Parse query params for filtering
	status := r.URL.Query().Get("status")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	query := db.Preload("Company").Order("issued_date DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("issued_date >= ?", date)
		}
	}
	if endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("issued_date <= ?", date)
		}
	}

	var invoices []models.Invoice
	query.Find(&invoices)

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoices_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"Invoice Number", "Company", "Amount", "Currency", "Status",
		"Description", "Issued Date", "Due Date", "Created At",
	})

	// Write data
	for _, inv := range invoices {
		companyName := ""
		if inv.Company.Name != "" {
			companyName = inv.Company.Name
		}

		writer.Write([]string{
			inv.InvoiceNum,
			companyName,
			fmt.Sprintf("%.2f", inv.Amount),
			inv.Currency,
			string(inv.Status),
			inv.Description,
			inv.IssuedDate.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
			inv.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

// ExportExpensesCSV handles GET /api/v1/export/expenses/csv
func (c *ExportController) ExportExpensesCSV(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	category := r.URL.Query().Get("category")

	query := db.Order("date DESC")

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("date >= ?", date)
		}
	}
	if endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("date <= ?", date)
		}
	}

	var expenses []models.Expense
	query.Find(&expenses)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=expenses_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{
		"Date", "Description", "Amount", "Currency", "Category", "Vendor", "Notes",
	})

	for _, exp := range expenses {
		writer.Write([]string{
			exp.Date.Format("2006-01-02"),
			exp.Description,
			fmt.Sprintf("%.2f", exp.Amount),
			exp.Currency,
			string(exp.Category),
			exp.Vendor,
			exp.Notes,
		})
	}
}

// ExportTrackingCSV handles GET /api/v1/export/tracking/csv
func (c *ExportController) ExportTrackingCSV(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	contractID := r.URL.Query().Get("contract_id")

	query := db.Preload("Client").Preload("Contract").Order("start_time DESC")

	if contractID != "" {
		query = query.Where("contract_id = ?", contractID)
	}
	if startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("start_time >= ?", date)
		}
	}
	if endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			endOfDay := date.Add(24 * time.Hour)
			query = query.Where("start_time < ?", endOfDay)
		}
	}

	var sessions []models.TrackingSession
	query.Find(&sessions)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=time_tracking_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{
		"Date", "Client", "Contract", "Project", "Start Time", "End Time",
		"Hours", "Billable", "Notes",
	})

	for _, session := range sessions {
		clientName := ""
		if session.Client != nil {
			clientName = session.Client.Name
		}
		contractName := ""
		if session.Contract != nil {
			contractName = session.Contract.Name
		}
		endTime := ""
		if session.EndTime != nil {
			endTime = session.EndTime.Format("15:04")
		}
		hours := ""
		if session.Hours != nil {
			hours = fmt.Sprintf("%.2f", *session.Hours)
		}

		writer.Write([]string{
			session.StartTime.Format("2006-01-02"),
			clientName,
			contractName,
			session.ProjectName,
			session.StartTime.Format("15:04"),
			endTime,
			hours,
			fmt.Sprintf("%t", session.Billable),
			session.Notes,
		})
	}
}

// ExportClientsCSV handles GET /api/v1/export/clients/csv
func (c *ExportController) ExportClientsCSV(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var clients []models.Client
	db.Order("name ASC").Find(&clients)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=clients_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{
		"Name", "Email", "Address", "Tax ID", "Created At",
	})

	for _, client := range clients {
		writer.Write([]string{
			client.Name,
			client.Email,
			client.Address,
			client.TaxID,
			client.CreatedAt.Format("2006-01-02"),
		})
	}
}

// ExportAllJSON handles GET /api/v1/export/all
func (c *ExportController) ExportAllJSON(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	export := make(map[string]interface{})

	var clients []models.Client
	db.Find(&clients)
	export["clients"] = clients

	var companies []models.Company
	db.Find(&companies)
	export["companies"] = companies

	var contracts []models.Contract
	db.Find(&contracts)
	export["contracts"] = contracts

	var invoices []models.Invoice
	db.Find(&invoices)
	export["invoices"] = invoices

	var expenses []models.Expense
	db.Find(&expenses)
	export["expenses"] = expenses

	var tracking []models.TrackingSession
	db.Find(&tracking)
	export["tracking"] = tracking

	var goals []models.IncomeGoal
	db.Find(&goals)
	export["goals"] = goals

	var settings models.UserSettings
	if err := db.First(&settings).Error; err == nil {
		export["settings"] = settings
	}

	export["exported_at"] = time.Now().Format(time.RFC3339)
	export["version"] = "1.0"

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=ung_backup_%s.json", time.Now().Format("2006-01-02")))

	json.NewEncoder(w).Encode(export)
}

// ImportClientsCSV handles POST /api/v1/import/clients/csv
func (c *ExportController) ImportClientsCSV(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		RespondError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		RespondError(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		RespondError(w, "Failed to read CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		RespondError(w, "CSV file is empty or has no data rows", http.StatusBadRequest)
		return
	}

	// Skip header row
	imported := 0
	errors := []string{}

	for i, record := range records[1:] {
		if len(record) < 2 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		client := models.Client{
			Name:  strings.TrimSpace(record[0]),
			Email: strings.TrimSpace(record[1]),
		}

		if len(record) > 2 {
			client.Address = strings.TrimSpace(record[2])
		}
		if len(record) > 3 {
			client.TaxID = strings.TrimSpace(record[3])
		}

		if client.Name == "" || client.Email == "" {
			errors = append(errors, fmt.Sprintf("Row %d: name and email are required", i+2))
			continue
		}

		// Check for duplicate email
		var existing models.Client
		if err := db.Where("email = ?", client.Email).First(&existing).Error; err == nil {
			errors = append(errors, fmt.Sprintf("Row %d: client with email %s already exists", i+2, client.Email))
			continue
		}

		if err := db.Create(&client).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+2, err.Error()))
			continue
		}

		imported++
	}

	response := map[string]interface{}{
		"imported": imported,
		"errors":   errors,
		"total":    len(records) - 1,
	}

	RespondJSON(w, response, http.StatusOK)
}

// ImportExpensesCSV handles POST /api/v1/import/expenses/csv
func (c *ExportController) ImportExpensesCSV(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		RespondError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		RespondError(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		RespondError(w, "Failed to read CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		RespondError(w, "CSV file is empty or has no data rows", http.StatusBadRequest)
		return
	}

	imported := 0
	errors := []string{}

	for i, record := range records[1:] {
		if len(record) < 4 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		date, err := time.Parse("2006-01-02", strings.TrimSpace(record[0]))
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: invalid date format", i+2))
			continue
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: invalid amount", i+2))
			continue
		}

		expense := models.Expense{
			Date:        date,
			Description: strings.TrimSpace(record[1]),
			Amount:      amount,
			Currency:    "USD",
			Category:    models.ExpenseCategory(strings.TrimSpace(record[3])),
		}

		if len(record) > 4 {
			expense.Vendor = strings.TrimSpace(record[4])
		}
		if len(record) > 5 {
			expense.Notes = strings.TrimSpace(record[5])
		}

		if err := db.Create(&expense).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+2, err.Error()))
			continue
		}

		imported++
	}

	response := map[string]interface{}{
		"imported": imported,
		"errors":   errors,
		"total":    len(records) - 1,
	}

	RespondJSON(w, response, http.StatusOK)
}

// ImportAllJSON handles POST /api/v1/import/all
func (c *ExportController) ImportAllJSON(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var importData struct {
		Clients   []models.Client          `json:"clients"`
		Companies []models.Company         `json:"companies"`
		Contracts []models.Contract        `json:"contracts"`
		Invoices  []models.Invoice         `json:"invoices"`
		Expenses  []models.Expense         `json:"expenses"`
		Tracking  []models.TrackingSession `json:"tracking"`
		Goals     []models.IncomeGoal      `json:"goals"`
		Settings  *models.UserSettings     `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
		RespondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	stats := map[string]int{
		"clients":   0,
		"companies": 0,
		"contracts": 0,
		"invoices":  0,
		"expenses":  0,
		"tracking":  0,
		"goals":     0,
	}

	// Import in order (to maintain foreign key relationships)
	for _, company := range importData.Companies {
		company.ID = 0 // Reset ID for new insert
		if err := db.Create(&company).Error; err == nil {
			stats["companies"]++
		}
	}

	for _, client := range importData.Clients {
		client.ID = 0
		if err := db.Create(&client).Error; err == nil {
			stats["clients"]++
		}
	}

	for _, contract := range importData.Contracts {
		contract.ID = 0
		if err := db.Create(&contract).Error; err == nil {
			stats["contracts"]++
		}
	}

	for _, invoice := range importData.Invoices {
		invoice.ID = 0
		if err := db.Create(&invoice).Error; err == nil {
			stats["invoices"]++
		}
	}

	for _, expense := range importData.Expenses {
		expense.ID = 0
		if err := db.Create(&expense).Error; err == nil {
			stats["expenses"]++
		}
	}

	for _, session := range importData.Tracking {
		session.ID = 0
		if err := db.Create(&session).Error; err == nil {
			stats["tracking"]++
		}
	}

	for _, goal := range importData.Goals {
		goal.ID = 0
		if err := db.Create(&goal).Error; err == nil {
			stats["goals"]++
		}
	}

	if importData.Settings != nil {
		importData.Settings.ID = 0
		db.Create(importData.Settings)
	}

	RespondJSON(w, stats, http.StatusOK)
}
