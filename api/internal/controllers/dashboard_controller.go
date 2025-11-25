package controllers

import (
	"net/http"
	"time"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// DashboardController handles dashboard and analytics endpoints
type DashboardController struct{}

// NewDashboardController creates a new dashboard controller
func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// RevenueProjection represents projected monthly revenue
type RevenueProjection struct {
	TotalMonthlyRevenue    float64              `json:"total_monthly_revenue"`
	Currency               string               `json:"currency"`
	ActiveContracts        int                  `json:"active_contracts"`
	HourlyContractsRevenue float64              `json:"hourly_contracts_revenue"`
	RetainerRevenue        float64              `json:"retainer_revenue"`
	ProjectedHours         float64              `json:"projected_hours"`
	AverageHourlyRate      float64              `json:"average_hourly_rate"`
	ContractBreakdown      []ContractProjection `json:"contract_breakdown"`
	LastMonthActual        float64              `json:"last_month_actual"`
	CurrentMonthActual     float64              `json:"current_month_actual"`
}

// ContractProjection represents revenue projection for a single contract
type ContractProjection struct {
	ContractID      uint    `json:"contract_id"`
	ContractName    string  `json:"contract_name"`
	ClientName      string  `json:"client_name"`
	ContractType    string  `json:"contract_type"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`
	HourlyRate      float64 `json:"hourly_rate,omitempty"`
	EstimatedHours  float64 `json:"estimated_hours,omitempty"`
	FixedAmount     float64 `json:"fixed_amount,omitempty"`
	Currency        string  `json:"currency"`
	Active          bool    `json:"active"`
}

// GetRevenue handles GET /api/v1/dashboard/revenue
func (c *DashboardController) GetRevenue(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Fetch active contracts with client info
	var contracts []models.Contract
	if err := db.Preload("Client").Where("active = ?", true).Find(&contracts).Error; err != nil {
		RespondError(w, "Failed to fetch contracts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	projection := RevenueProjection{
		Currency:          "USD",
		ActiveContracts:   len(contracts),
		ContractBreakdown: make([]ContractProjection, 0),
	}

	// Calculate average hours worked per month for hourly contracts
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var totalBillableHours float64
	var hourlyContractCount int

	for _, contract := range contracts {
		cp := ContractProjection{
			ContractID:   contract.ID,
			ContractName: contract.Name,
			ClientName:   contract.Client.Name,
			ContractType: string(contract.ContractType),
			Currency:     contract.Currency,
			Active:       contract.Active,
		}

		switch contract.ContractType {
		case models.ContractTypeHourly:
			if contract.HourlyRate != nil && *contract.HourlyRate > 0 {
				// Get tracking sessions for this contract in the last 30 days
				var sessions []models.TrackingSession
				db.Where("contract_id = ? AND start_time >= ? AND billable = ?",
					contract.ID, thirtyDaysAgo, true).Find(&sessions)

				var totalHours float64
				for _, session := range sessions {
					if session.Hours != nil {
						totalHours += *session.Hours
					}
				}

				// Project to monthly revenue
				cp.HourlyRate = *contract.HourlyRate
				cp.EstimatedHours = totalHours // Last 30 days
				cp.MonthlyRevenue = totalHours * (*contract.HourlyRate)

				projection.HourlyContractsRevenue += cp.MonthlyRevenue
				projection.ProjectedHours += totalHours
				totalBillableHours += totalHours
				hourlyContractCount++
			}

		case models.ContractTypeRetainer:
			// Retainer is monthly recurring
			if contract.FixedPrice != nil {
				cp.FixedAmount = *contract.FixedPrice
				cp.MonthlyRevenue = *contract.FixedPrice
				projection.RetainerRevenue += cp.MonthlyRevenue
			}

		case models.ContractTypeFixedPrice:
			// Fixed price is one-time, but show prorated if contract has duration
			if contract.FixedPrice != nil {
				cp.FixedAmount = *contract.FixedPrice

				// If contract has end date, prorate over duration
				if contract.EndDate != nil {
					duration := contract.EndDate.Sub(contract.StartDate)
					months := duration.Hours() / 24 / 30 // Approximate months
					if months > 0 {
						cp.MonthlyRevenue = *contract.FixedPrice / months
					} else {
						cp.MonthlyRevenue = *contract.FixedPrice // Show full amount for short contracts
					}
				} else {
					// No end date, assume 3 month project by default
					cp.MonthlyRevenue = *contract.FixedPrice / 3
				}
			}
		}

		projection.ContractBreakdown = append(projection.ContractBreakdown, cp)
		projection.TotalMonthlyRevenue += cp.MonthlyRevenue
	}

	// Calculate average hourly rate
	if hourlyContractCount > 0 {
		projection.AverageHourlyRate = projection.HourlyContractsRevenue / totalBillableHours
	}

	// Calculate last month's actual invoiced amount
	lastMonthStart := time.Now().AddDate(0, -1, 0)
	lastMonthStart = time.Date(lastMonthStart.Year(), lastMonthStart.Month(), 1, 0, 0, 0, 0, lastMonthStart.Location())
	lastMonthEnd := lastMonthStart.AddDate(0, 1, 0).Add(-time.Second)

	var lastMonthInvoices []models.Invoice
	db.Where("issued_date >= ? AND issued_date <= ?", lastMonthStart, lastMonthEnd).Find(&lastMonthInvoices)

	for _, inv := range lastMonthInvoices {
		projection.LastMonthActual += inv.Amount
	}

	// Calculate current month's actual invoiced amount
	currentMonthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())

	var currentMonthInvoices []models.Invoice
	db.Where("issued_date >= ?", currentMonthStart).Find(&currentMonthInvoices)

	for _, inv := range currentMonthInvoices {
		projection.CurrentMonthActual += inv.Amount
	}

	RespondJSON(w, projection, http.StatusOK)
}

// DashboardSummary represents overall dashboard metrics
type DashboardSummary struct {
	TotalClients        int64   `json:"total_clients"`
	ActiveContracts     int64   `json:"active_contracts"`
	PendingInvoices     int64   `json:"pending_invoices"`
	OverdueInvoices     int64   `json:"overdue_invoices"`
	MonthlyRevenue      float64 `json:"monthly_revenue"`
	UnpaidAmount        float64 `json:"unpaid_amount"`
	ThisMonthHours      float64 `json:"this_month_hours"`
	LastMonthHours      float64 `json:"last_month_hours"`
	ThisMonthExpenses   float64 `json:"this_month_expenses"`
	Currency            string  `json:"currency"`
}

// GetSummary handles GET /api/v1/dashboard/summary
func (c *DashboardController) GetSummary(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	summary := DashboardSummary{
		Currency: "USD",
	}

	// Total clients
	db.Model(&models.Client{}).Count(&summary.TotalClients)

	// Active contracts
	db.Model(&models.Contract{}).Where("active = ?", true).Count(&summary.ActiveContracts)

	// Pending invoices
	db.Model(&models.Invoice{}).Where("status = ?", models.StatusPending).Count(&summary.PendingInvoices)

	// Overdue invoices
	db.Model(&models.Invoice{}).Where("status = ?", models.StatusOverdue).Count(&summary.OverdueInvoices)

	// Unpaid amount (pending + overdue)
	var unpaidInvoices []models.Invoice
	db.Where("status IN ?", []models.InvoiceStatus{models.StatusPending, models.StatusOverdue}).Find(&unpaidInvoices)
	for _, inv := range unpaidInvoices {
		summary.UnpaidAmount += inv.Amount
	}

	// This month's hours
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var thisMonthSessions []models.TrackingSession
	db.Where("start_time >= ? AND billable = ?", monthStart, true).Find(&thisMonthSessions)
	for _, session := range thisMonthSessions {
		if session.Hours != nil {
			summary.ThisMonthHours += *session.Hours
		}
	}

	// Last month's hours
	lastMonthStart := monthStart.AddDate(0, -1, 0)
	lastMonthEnd := monthStart.Add(-time.Second)
	var lastMonthSessions []models.TrackingSession
	db.Where("start_time >= ? AND start_time <= ? AND billable = ?",
		lastMonthStart, lastMonthEnd, true).Find(&lastMonthSessions)
	for _, session := range lastMonthSessions {
		if session.Hours != nil {
			summary.LastMonthHours += *session.Hours
		}
	}

	// This month's expenses
	var thisMonthExpenses []models.Expense
	db.Where("date >= ?", monthStart).Find(&thisMonthExpenses)
	for _, expense := range thisMonthExpenses {
		summary.ThisMonthExpenses += expense.Amount
	}

	RespondJSON(w, summary, http.StatusOK)
}
