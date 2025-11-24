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

// ClientCreateRequest represents a client creation request
type ClientCreateRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address,omitempty"`
	TaxID   string `json:"tax_id,omitempty"`
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
