package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
