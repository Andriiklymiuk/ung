package controllers

import (
	"net/http"
	"strings"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// SearchController handles global search endpoints
type SearchController struct{}

// NewSearchController creates a new search controller
func NewSearchController() *SearchController {
	return &SearchController{}
}

// SearchResult represents a unified search result
type SearchResult struct {
	Type        string      `json:"type"` // client, invoice, contract, expense, tracking
	ID          uint        `json:"id"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Counts  map[string]int `json:"counts"`
}

// Search handles GET /api/v1/search?q=query
func (c *SearchController) Search(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		RespondError(w, "Search query is required", http.StatusBadRequest)
		return
	}

	// Optional type filter
	typeFilter := r.URL.Query().Get("type")
	limit := 10

	response := SearchResponse{
		Query:   query,
		Results: []SearchResult{},
		Counts:  make(map[string]int),
	}

	searchPattern := "%" + strings.ToLower(query) + "%"

	// Search Clients
	if typeFilter == "" || typeFilter == "client" {
		var clients []models.Client
		db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern).
			Limit(limit).Find(&clients)

		response.Counts["clients"] = len(clients)
		for _, client := range clients {
			response.Results = append(response.Results, SearchResult{
				Type:        "client",
				ID:          client.ID,
				Title:       client.Name,
				Subtitle:    client.Email,
				Description: client.Address,
				Data:        client,
			})
		}
	}

	// Search Invoices
	if typeFilter == "" || typeFilter == "invoice" {
		var invoices []models.Invoice
		db.Where("LOWER(invoice_num) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern).
			Limit(limit).Find(&invoices)

		response.Counts["invoices"] = len(invoices)
		for _, inv := range invoices {
			response.Results = append(response.Results, SearchResult{
				Type:        "invoice",
				ID:          inv.ID,
				Title:       inv.InvoiceNum,
				Subtitle:    string(inv.Status),
				Description: inv.Description,
				Data:        inv,
			})
		}
	}

	// Search Contracts
	if typeFilter == "" || typeFilter == "contract" {
		var contracts []models.Contract
		db.Preload("Client").
			Where("LOWER(name) LIKE ? OR LOWER(contract_num) LIKE ? OR LOWER(notes) LIKE ?",
				searchPattern, searchPattern, searchPattern).
			Limit(limit).Find(&contracts)

		response.Counts["contracts"] = len(contracts)
		for _, contract := range contracts {
			clientName := ""
			if contract.Client.Name != "" {
				clientName = contract.Client.Name
			}
			response.Results = append(response.Results, SearchResult{
				Type:        "contract",
				ID:          contract.ID,
				Title:       contract.Name,
				Subtitle:    clientName,
				Description: string(contract.ContractType),
				Data:        contract,
			})
		}
	}

	// Search Expenses
	if typeFilter == "" || typeFilter == "expense" {
		var expenses []models.Expense
		db.Where("LOWER(description) LIKE ? OR LOWER(vendor) LIKE ? OR LOWER(notes) LIKE ?",
			searchPattern, searchPattern, searchPattern).
			Limit(limit).Find(&expenses)

		response.Counts["expenses"] = len(expenses)
		for _, expense := range expenses {
			response.Results = append(response.Results, SearchResult{
				Type:        "expense",
				ID:          expense.ID,
				Title:       expense.Description,
				Subtitle:    expense.Vendor,
				Description: string(expense.Category),
				Data:        expense,
			})
		}
	}

	// Search Companies
	if typeFilter == "" || typeFilter == "company" {
		var companies []models.Company
		db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern).
			Limit(limit).Find(&companies)

		response.Counts["companies"] = len(companies)
		for _, company := range companies {
			response.Results = append(response.Results, SearchResult{
				Type:        "company",
				ID:          company.ID,
				Title:       company.Name,
				Subtitle:    company.Email,
				Description: company.Address,
				Data:        company,
			})
		}
	}

	// Search Tracking Sessions
	if typeFilter == "" || typeFilter == "tracking" {
		var sessions []models.TrackingSession
		db.Preload("Client").Preload("Contract").
			Where("LOWER(project_name) LIKE ? OR LOWER(notes) LIKE ?", searchPattern, searchPattern).
			Limit(limit).Find(&sessions)

		response.Counts["tracking"] = len(sessions)
		for _, session := range sessions {
			subtitle := ""
			if session.Client != nil {
				subtitle = session.Client.Name
			}
			response.Results = append(response.Results, SearchResult{
				Type:        "tracking",
				ID:          session.ID,
				Title:       session.ProjectName,
				Subtitle:    subtitle,
				Description: session.Notes,
				Data:        session,
			})
		}
	}

	RespondJSON(w, response, http.StatusOK)
}

// SearchInvoices handles GET /api/v1/search/invoices
func (c *SearchController) SearchInvoices(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	status := r.URL.Query().Get("status")
	clientID := r.URL.Query().Get("client_id")

	dbQuery := db.Preload("Company").Order("created_at DESC")

	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(invoice_num) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	if clientID != "" {
		// Find invoices where client is a recipient
		var recipientInvoiceIDs []uint
		db.Model(&models.InvoiceRecipient{}).Where("client_id = ?", clientID).Pluck("invoice_id", &recipientInvoiceIDs)
		if len(recipientInvoiceIDs) > 0 {
			dbQuery = dbQuery.Where("id IN ?", recipientInvoiceIDs)
		} else {
			dbQuery = dbQuery.Where("1 = 0") // No results
		}
	}

	var invoices []models.Invoice
	dbQuery.Find(&invoices)

	RespondJSON(w, invoices, http.StatusOK)
}

// SearchClients handles GET /api/v1/search/clients
func (c *SearchController) SearchClients(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	query := strings.TrimSpace(r.URL.Query().Get("q"))

	dbQuery := db.Order("name ASC")

	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(address) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	var clients []models.Client
	dbQuery.Find(&clients)

	RespondJSON(w, clients, http.StatusOK)
}

// SearchContracts handles GET /api/v1/search/contracts
func (c *SearchController) SearchContracts(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	clientID := r.URL.Query().Get("client_id")
	active := r.URL.Query().Get("active")

	dbQuery := db.Preload("Client").Order("created_at DESC")

	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(contract_num) LIKE ?", searchPattern, searchPattern)
	}

	if clientID != "" {
		dbQuery = dbQuery.Where("client_id = ?", clientID)
	}

	if active != "" {
		dbQuery = dbQuery.Where("active = ?", active == "true")
	}

	var contracts []models.Contract
	dbQuery.Find(&contracts)

	RespondJSON(w, contracts, http.StatusOK)
}
