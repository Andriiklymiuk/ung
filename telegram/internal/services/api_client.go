package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// APIClient handles communication with UNG API
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Invoice represents an invoice from the API
type Invoice struct {
	ID         uint    `json:"id"`
	InvoiceNum string  `json:"invoice_num"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	DueDate    string  `json:"due_date"`
}

// Client represents a client from the API
type Client struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
}

// Company represents a company from the API
type Company struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
}

// Contract represents a contract from the API
type Contract struct {
	ID       uint    `json:"id"`
	ClientID uint    `json:"client_id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Rate     float64 `json:"rate"`
	Status   string  `json:"status"`
}

// Expense represents an expense from the API
type Expense struct {
	ID          uint    `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Vendor      string  `json:"vendor"`
	Date        string  `json:"date"`
}

// TrackingSession represents a time tracking session from the API
type TrackingSession struct {
	ID        uint    `json:"id"`
	ProjectID uint    `json:"project_id"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Duration  float64 `json:"duration"`
	Notes     string  `json:"notes"`
	Active    bool    `json:"active"`
}

// ClientCreateRequest represents a client creation request
type ClientCreateRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address,omitempty"`
	TaxID   string `json:"tax_id,omitempty"`
}

// CompanyCreateRequest represents a company creation request
type CompanyCreateRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
	TaxID   string `json:"tax_id,omitempty"`
}

// ContractCreateRequest represents a contract creation request
type ContractCreateRequest struct {
	ClientID uint    `json:"client_id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Rate     float64 `json:"rate"`
}

// ExpenseCreateRequest represents an expense creation request
type ExpenseCreateRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Vendor      string  `json:"vendor,omitempty"`
}

// APIResponse represents standard API response
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// ListInvoices fetches invoices for a user
func (c *APIClient) ListInvoices(token string) ([]Invoice, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/invoices", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var invoices []Invoice
	if err := json.Unmarshal(apiResp.Data, &invoices); err != nil {
		return nil, err
	}

	return invoices, nil
}

// ListClients fetches clients for a user
func (c *APIClient) ListClients(token string) ([]Client, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/clients", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var clients []Client
	if err := json.Unmarshal(apiResp.Data, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

// CreateInvoice creates a new invoice
func (c *APIClient) CreateInvoice(token string, invoice map[string]interface{}) (*Invoice, error) {
	jsonData, err := json.Marshal(invoice)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/invoices", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var createdInvoice Invoice
	if err := json.Unmarshal(apiResp.Data, &createdInvoice); err != nil {
		return nil, err
	}

	return &createdInvoice, nil
}

// Login authenticates a user
func (c *APIClient) Login(email, password string) (string, error) {
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed: %s", string(body))
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", err
	}

	return apiResp.Data.AccessToken, nil
}

// CreateClient creates a new client
func (c *APIClient) CreateClient(token string, req ClientCreateRequest) (*Client, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/clients", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var createdClient Client
	if err := json.Unmarshal(apiResp.Data, &createdClient); err != nil {
		return nil, err
	}

	return &createdClient, nil
}

// ListCompanies fetches companies for a user
func (c *APIClient) ListCompanies(token string) ([]Company, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/companies", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var companies []Company
	if err := json.Unmarshal(apiResp.Data, &companies); err != nil {
		return nil, err
	}

	return companies, nil
}

// CreateCompany creates a new company
func (c *APIClient) CreateCompany(token string, req CompanyCreateRequest) (*Company, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/companies", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var createdCompany Company
	if err := json.Unmarshal(apiResp.Data, &createdCompany); err != nil {
		return nil, err
	}

	return &createdCompany, nil
}

// ListContracts fetches contracts for a user
func (c *APIClient) ListContracts(token string) ([]Contract, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/contracts", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var contracts []Contract
	if err := json.Unmarshal(apiResp.Data, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

// CreateContract creates a new contract
func (c *APIClient) CreateContract(token string, req ContractCreateRequest) (*Contract, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/contracts", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var createdContract Contract
	if err := json.Unmarshal(apiResp.Data, &createdContract); err != nil {
		return nil, err
	}

	return &createdContract, nil
}

// ListExpenses fetches expenses for a user
func (c *APIClient) ListExpenses(token string) ([]Expense, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/expenses", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var expenses []Expense
	if err := json.Unmarshal(apiResp.Data, &expenses); err != nil {
		return nil, err
	}

	return expenses, nil
}

// CreateExpense creates a new expense
func (c *APIClient) CreateExpense(token string, req ExpenseCreateRequest) (*Expense, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/expenses", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var createdExpense Expense
	if err := json.Unmarshal(apiResp.Data, &createdExpense); err != nil {
		return nil, err
	}

	return &createdExpense, nil
}

// ListTracking fetches time tracking sessions for a user
func (c *APIClient) ListTracking(token string) ([]TrackingSession, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/tracking", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var sessions []TrackingSession
	if err := json.Unmarshal(apiResp.Data, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// StartTracking starts a new time tracking session
func (c *APIClient) StartTracking(token string, projectID uint, notes string) (*TrackingSession, error) {
	reqData := map[string]interface{}{
		"project_id": projectID,
		"notes":      notes,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/tracking/start", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var session TrackingSession
	if err := json.Unmarshal(apiResp.Data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// StopTracking stops the active time tracking session
func (c *APIClient) StopTracking(token string) (*TrackingSession, error) {
	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/tracking/stop", nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var session TrackingSession
	if err := json.Unmarshal(apiResp.Data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// TrackingCreateRequest represents manual time entry request
type TrackingCreateRequest struct {
	ContractID  uint    `json:"contract_id"`
	ClientID    uint    `json:"client_id"`
	ProjectName string  `json:"project_name"`
	Hours       float64 `json:"hours"`
	Notes       string  `json:"notes"`
	Billable    bool    `json:"billable"`
}

// CreateTracking creates a manual time tracking entry
func (c *APIClient) CreateTracking(token string, req TrackingCreateRequest) (*TrackingSession, error) {
	// Calculate start and end times based on hours
	now := time.Now()
	startTime := now.Add(-time.Duration(req.Hours * float64(time.Hour)))

	payload := map[string]interface{}{
		"contract_id":  req.ContractID,
		"client_id":    req.ClientID,
		"project_name": req.ProjectName,
		"start_time":   startTime.Format(time.RFC3339),
		"end_time":     now.Format(time.RFC3339),
		"hours":        req.Hours,
		"billable":     req.Billable,
		"notes":        req.Notes,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/tracking", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var createdSession TrackingSession
	if err := json.Unmarshal(apiResp.Data, &createdSession); err != nil {
		return nil, err
	}

	return &createdSession, nil
}

// RevenueProjection represents revenue data from dashboard
type RevenueProjection struct {
	TotalMonthlyRevenue    float64              `json:"total_monthly_revenue"`
	Currency               string               `json:"currency"`
	ActiveContracts        int                  `json:"active_contracts"`
	HourlyContractsRevenue float64              `json:"hourly_contracts_revenue"`
	RetainerRevenue        float64              `json:"retainer_revenue"`
	ProjectedHours         float64              `json:"projected_hours"`
	AverageHourlyRate      float64              `json:"average_hourly_rate"`
	ContractBreakdown      []ContractProjection `json:"contract_breakdown"`
}

type ContractProjection struct {
	ContractName   string  `json:"contract_name"`
	ClientName     string  `json:"client_name"`
	ContractType   string  `json:"contract_type"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
	Currency       string  `json:"currency"`
}

// GetDashboard fetches revenue projection dashboard
func (c *APIClient) GetDashboard(token string) (*RevenueProjection, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/dashboard/revenue", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var projection RevenueProjection
	if err := json.Unmarshal(apiResp.Data, &projection); err != nil {
		return nil, err
	}

	return &projection, nil
}

// RecurringInvoice represents a recurring invoice from the API
type RecurringInvoice struct {
	ID             uint    `json:"id"`
	ClientID       uint    `json:"client_id"`
	CompanyID      uint    `json:"company_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Description    string  `json:"description"`
	Frequency      string  `json:"frequency"`
	DayOfMonth     int     `json:"day_of_month"`
	NextRunDate    string  `json:"next_run_date"`
	Active         bool    `json:"active"`
	TotalGenerated int     `json:"total_generated"`
}

// ListRecurring fetches recurring invoices
func (c *APIClient) ListRecurring(token string) ([]RecurringInvoice, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/recurring", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var recurring []RecurringInvoice
	if err := json.Unmarshal(body, &recurring); err != nil {
		return nil, err
	}

	return recurring, nil
}

// WeeklyReport represents weekly report data
type WeeklyReport struct {
	WeekStart    string  `json:"week_start"`
	WeekEnd      string  `json:"week_end"`
	TotalHours   float64 `json:"total_hours"`
	TotalRevenue float64 `json:"total_revenue"`
	Sessions     int     `json:"sessions"`
}

// GetWeeklyReport fetches weekly report
func (c *APIClient) GetWeeklyReport(token string) (*WeeklyReport, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/reports/weekly", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var report WeeklyReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// MonthlyReport represents monthly report data
type MonthlyReport struct {
	Month         string  `json:"month"`
	Year          int     `json:"year"`
	TotalHours    float64 `json:"total_hours"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	Profit        float64 `json:"profit"`
	InvoicesSent  int     `json:"invoices_sent"`
	InvoicesPaid  int     `json:"invoices_paid"`
}

// GetMonthlyReport fetches monthly report
func (c *APIClient) GetMonthlyReport(token string) (*MonthlyReport, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/reports/monthly", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var report MonthlyReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// OverdueReport represents overdue invoices report
type OverdueReport struct {
	TotalOverdue float64 `json:"total_overdue"`
	Count        int     `json:"count"`
}

// GetOverdueReport fetches overdue invoices report
func (c *APIClient) GetOverdueReport(token string) (*OverdueReport, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/reports/overdue", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var report OverdueReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// UnpaidReport represents unpaid invoices report
type UnpaidReport struct {
	TotalUnpaid float64 `json:"total_unpaid"`
	Pending     float64 `json:"pending"`
	Overdue     float64 `json:"overdue"`
	Count       int     `json:"count"`
}

// GetUnpaidReport fetches unpaid invoices report
func (c *APIClient) GetUnpaidReport(token string) (*UnpaidReport, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/reports/unpaid", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var report UnpaidReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// PomodoroStats represents pomodoro statistics
type PomodoroStats struct {
	TodayCompleted int64   `json:"today_completed"`
	TodayMinutes   int     `json:"today_minutes"`
	WeekCompleted  int64   `json:"week_completed"`
	MonthCompleted int64   `json:"month_completed"`
	TotalCompleted int64   `json:"total_completed"`
	CurrentStreak  int     `json:"current_streak"`
	AvgDaily30d    float64 `json:"avg_daily_30d"`
}

// StartPomodoro starts a pomodoro session
func (c *APIClient) StartPomodoro(token string, duration int, projectName string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"duration":     duration,
		"project_name": projectName,
		"session_type": "work",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/pomodoro/start", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetActivePomodoro gets the active pomodoro session
func (c *APIClient) GetActivePomodoro(token string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/pomodoro/active", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetPomodoroStats gets pomodoro statistics
func (c *APIClient) GetPomodoroStats(token string) (*PomodoroStats, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/pomodoro/stats", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var stats PomodoroStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// IncomeGoal represents an income goal
type IncomeGoal struct {
	ID          uint    `json:"id"`
	Amount      float64 `json:"amount"`
	Period      string  `json:"period"`
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	Description string  `json:"description"`
}

// GoalStatus represents goal progress
type GoalStatus struct {
	Goal      IncomeGoal `json:"goal"`
	Current   float64    `json:"current"`
	Progress  float64    `json:"progress_percent"`
	Remaining float64    `json:"remaining"`
}

// GetGoalStatus fetches goal status/progress
func (c *APIClient) GetGoalStatus(token string) (*GoalStatus, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/goals/status", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var status GoalStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// RateAnalysis represents rate analysis data
type RateAnalysis struct {
	AverageRate   float64 `json:"average_rate"`
	EffectiveRate float64 `json:"effective_rate"`
	TotalHours    float64 `json:"total_hours"`
	TotalRevenue  float64 `json:"total_revenue"`
	SuggestedRate float64 `json:"suggested_rate"`
}

// GetRateAnalysis fetches rate analysis
func (c *APIClient) GetRateAnalysis(token string) (*RateAnalysis, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/rate/analyze", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var analysis RateAnalysis
	if err := json.Unmarshal(body, &analysis); err != nil {
		return nil, err
	}

	return &analysis, nil
}

// UserSettings represents user settings
type UserSettings struct {
	ID                uint    `json:"id"`
	HoursPerWeek      float64 `json:"hours_per_week"`
	WeeksPerYear      int     `json:"weeks_per_year"`
	DefaultTaxPercent float64 `json:"default_tax_percent"`
	DefaultMargin     float64 `json:"default_margin"`
	AnnualExpenses    float64 `json:"annual_expenses"`
	DefaultCurrency   string  `json:"default_currency"`
}

// GetSettings fetches user settings
func (c *APIClient) GetSettings(token string) (*UserSettings, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/settings", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var settings UserSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SearchResult represents a search result
type SearchResult struct {
	Type        string `json:"type"`
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
}

// SearchResponse represents search response
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Counts  map[string]int `json:"counts"`
}

// Search performs a global search
func (c *APIClient) Search(token string, query string) (*SearchResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/search?q="+query, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

// =====================================
// Job Hunter API
// =====================================

// HunterProfile represents a hunter profile
type HunterProfile struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Title      string   `json:"title"`
	Bio        string   `json:"bio"`
	Skills     []string `json:"skills"`
	Experience int      `json:"experience"`
	Rate       float64  `json:"rate"`
	Currency   string   `json:"currency"`
	Location   string   `json:"location"`
	Remote     bool     `json:"remote"`
}

// HunterJob represents a scraped job
type HunterJob struct {
	ID          uint    `json:"id"`
	Source      string  `json:"source"`
	SourceURL   string  `json:"source_url"`
	Title       string  `json:"title"`
	Company     string  `json:"company"`
	Description string  `json:"description"`
	Skills      []string `json:"skills"`
	RateMin     float64 `json:"rate_min"`
	RateMax     float64 `json:"rate_max"`
	RateType    string  `json:"rate_type"`
	Currency    string  `json:"currency"`
	Remote      bool    `json:"remote"`
	Location    string  `json:"location"`
	MatchScore  float64 `json:"match_score"`
	PostedAt    string  `json:"posted_at"`
}

// HunterApplication represents a job application
type HunterApplication struct {
	ID        uint   `json:"id"`
	JobID     uint   `json:"job_id"`
	Proposal  string `json:"proposal"`
	Status    string `json:"status"`
	AppliedAt string `json:"applied_at"`
}

// HunterStats represents hunter statistics
type HunterStats struct {
	TotalJobs         int            `json:"total_jobs"`
	TotalApplications int            `json:"total_applications"`
	StatusCounts      map[string]int `json:"status_counts"`
	TopSkills         []string       `json:"top_skills"`
	AverageMatchScore float64        `json:"average_match_score"`
}

// HuntResult represents the result of a hunt operation
type HuntResult struct {
	Jobs       []HunterJob `json:"jobs"`
	NewCount   int         `json:"new_count"`
	TotalCount int         `json:"total_count"`
}

// GetHunterProfile gets the user's hunter profile
func (c *APIClient) GetHunterProfile(token string) (*HunterProfile, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/hunter/profile", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No profile yet
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var profile HunterProfile
	if err := json.Unmarshal(apiResp.Data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// UpdateHunterProfile updates the user's hunter profile
func (c *APIClient) UpdateHunterProfile(token string, profile *HunterProfile) (*HunterProfile, error) {
	jsonData, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/v1/hunter/profile", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var updatedProfile HunterProfile
	if err := json.Unmarshal(apiResp.Data, &updatedProfile); err != nil {
		return nil, err
	}

	return &updatedProfile, nil
}

// ImportProfileFromPDF imports a hunter profile from a PDF CV file
func (c *APIClient) ImportProfileFromPDF(token string, pdfData []byte, filename string) (*HunterProfile, error) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	part, err := writer.CreateFormFile("cv", filename)
	if err != nil {
		return nil, err
	}
	part.Write(pdfData)
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/hunter/profile/import", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Longer timeout for PDF processing
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, err
	}

	var profile HunterProfile
	if err := json.Unmarshal(apiResp.Data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// HuntJobs triggers a job hunt/scrape
func (c *APIClient) HuntJobs(token string) (*HuntResult, error) {
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/hunter/hunt", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Longer timeout for scraping
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var result HuntResult
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetHunterJobs gets list of scraped jobs
func (c *APIClient) GetHunterJobs(token string, limit int) ([]HunterJob, error) {
	url := fmt.Sprintf("%s/api/v1/hunter/jobs?limit=%d", c.baseURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var jobs []HunterJob
	if err := json.Unmarshal(apiResp.Data, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetHunterStats gets hunter statistics
func (c *APIClient) GetHunterStats(token string) (*HunterStats, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/hunter/stats", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var stats HunterStats
	if err := json.Unmarshal(apiResp.Data, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// CreateHunterApplication creates a job application
func (c *APIClient) CreateHunterApplication(token string, jobID uint, generateProposal bool) (*HunterApplication, error) {
	data := map[string]interface{}{
		"job_id":            jobID,
		"generate_proposal": generateProposal,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/hunter/applications", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var application HunterApplication
	if err := json.Unmarshal(apiResp.Data, &application); err != nil {
		return nil, err
	}

	return &application, nil
}

// GetHunterApplications gets list of applications
func (c *APIClient) GetHunterApplications(token string) ([]HunterApplication, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/hunter/applications", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var applications []HunterApplication
	if err := json.Unmarshal(apiResp.Data, &applications); err != nil {
		return nil, err
	}

	return applications, nil
}
