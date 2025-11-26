package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// RateController handles rate calculation endpoints
type RateController struct{}

// NewRateController creates a new rate controller
func NewRateController() *RateController {
	return &RateController{}
}

// RateCalculation represents the output of rate calculation
type RateCalculation struct {
	Inputs struct {
		TargetAnnual   float64 `json:"target_annual"`
		TargetMonthly  float64 `json:"target_monthly"`
		HoursPerWeek   float64 `json:"hours_per_week"`
		WeeksPerYear   int     `json:"weeks_per_year"`
		Expenses       float64 `json:"expenses"`
		TaxPercent     float64 `json:"tax_percent"`
		ProfitMargin   float64 `json:"profit_margin"`
	} `json:"inputs"`
	Calculation struct {
		NetIncomeTarget float64 `json:"net_income_target"`
		TaxAmount       float64 `json:"tax_amount"`
		GrossNeeded     float64 `json:"gross_needed"`
		ExpensesAdded   float64 `json:"expenses_added"`
		MarginAmount    float64 `json:"margin_amount"`
		TotalRevenue    float64 `json:"total_revenue"`
		AnnualHours     float64 `json:"annual_hours"`
	} `json:"calculation"`
	RecommendedRates struct {
		Hourly  float64 `json:"hourly"`
		Daily   float64 `json:"daily"`
		Weekly  float64 `json:"weekly"`
		Monthly float64 `json:"monthly"`
	} `json:"recommended_rates"`
	RateTiers struct {
		Budget      float64 `json:"budget"`
		Standard    float64 `json:"standard"`
		Premium     float64 `json:"premium"`
		RushWeekend float64 `json:"rush_weekend"`
	} `json:"rate_tiers"`
}

// Calculate handles POST /api/v1/rate/calculate
func (c *RateController) Calculate(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		Annual       *float64 `json:"annual"`
		Monthly      *float64 `json:"monthly"`
		HoursPerWeek *float64 `json:"hours_per_week"`
		WeeksPerYear *int     `json:"weeks_per_year"`
		Expenses     *float64 `json:"expenses"`
		TaxPercent   *float64 `json:"tax_percent"`
		ProfitMargin *float64 `json:"profit_margin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user settings for defaults
	var settings models.UserSettings
	if err := db.First(&settings).Error; err != nil {
		settings = models.UserSettings{
			HoursPerWeek:      40,
			WeeksPerYear:      48,
			DefaultTaxPercent: 25,
			DefaultMargin:     20,
			AnnualExpenses:    0,
		}
	}

	// Use provided values or defaults
	hoursPerWeek := settings.HoursPerWeek
	if req.HoursPerWeek != nil {
		hoursPerWeek = *req.HoursPerWeek
	}

	weeksPerYear := settings.WeeksPerYear
	if req.WeeksPerYear != nil {
		weeksPerYear = *req.WeeksPerYear
	}

	expenses := settings.AnnualExpenses
	if req.Expenses != nil {
		expenses = *req.Expenses
	}

	taxPercent := settings.DefaultTaxPercent
	if req.TaxPercent != nil {
		taxPercent = *req.TaxPercent
	}

	profitMargin := settings.DefaultMargin
	if req.ProfitMargin != nil {
		profitMargin = *req.ProfitMargin
	}

	// Calculate target annual
	var targetAnnual float64
	if req.Annual != nil && *req.Annual > 0 {
		targetAnnual = *req.Annual
	} else if req.Monthly != nil && *req.Monthly > 0 {
		targetAnnual = *req.Monthly * 12
	} else {
		RespondError(w, "Either annual or monthly target is required", http.StatusBadRequest)
		return
	}

	// Calculate gross needed (to cover taxes)
	taxMultiplier := 1 / (1 - taxPercent/100)
	grossNeeded := targetAnnual * taxMultiplier

	// Add expenses
	totalNeeded := grossNeeded + expenses

	// Add profit margin
	withMargin := totalNeeded * (1 + profitMargin/100)

	// Calculate billable hours
	annualHours := hoursPerWeek * float64(weeksPerYear)

	// Calculate hourly rate
	hourlyRate := withMargin / annualHours

	result := RateCalculation{}
	result.Inputs.TargetAnnual = targetAnnual
	result.Inputs.TargetMonthly = targetAnnual / 12
	result.Inputs.HoursPerWeek = hoursPerWeek
	result.Inputs.WeeksPerYear = weeksPerYear
	result.Inputs.Expenses = expenses
	result.Inputs.TaxPercent = taxPercent
	result.Inputs.ProfitMargin = profitMargin

	result.Calculation.NetIncomeTarget = targetAnnual
	result.Calculation.TaxAmount = grossNeeded - targetAnnual
	result.Calculation.GrossNeeded = grossNeeded
	result.Calculation.ExpensesAdded = expenses
	result.Calculation.MarginAmount = withMargin - totalNeeded
	result.Calculation.TotalRevenue = withMargin
	result.Calculation.AnnualHours = annualHours

	result.RecommendedRates.Hourly = hourlyRate
	result.RecommendedRates.Daily = hourlyRate * 8
	result.RecommendedRates.Weekly = hourlyRate * hoursPerWeek
	result.RecommendedRates.Monthly = withMargin / 12

	result.RateTiers.Budget = hourlyRate * 0.7
	result.RateTiers.Standard = hourlyRate
	result.RateTiers.Premium = hourlyRate * 1.3
	result.RateTiers.RushWeekend = hourlyRate * 1.5

	RespondJSON(w, result, http.StatusOK)
}

// RateAnalysis represents the output of rate analysis
type RateAnalysis struct {
	TotalBillableHours     float64            `json:"total_billable_hours"`
	ContractBasedRevenue   float64            `json:"contract_based_revenue"`
	InvoiceRevenue         float64            `json:"invoice_revenue"`
	EffectiveRateContracts float64            `json:"effective_rate_contracts"`
	EffectiveRateInvoices  float64            `json:"effective_rate_invoices"`
	ByClient               []ClientRateStats  `json:"by_client"`
	Recommendation         string             `json:"recommendation"`
}

// ClientRateStats represents rate statistics for a client
type ClientRateStats struct {
	ClientName    string  `json:"client_name"`
	Hours         float64 `json:"hours"`
	Revenue       float64 `json:"revenue"`
	EffectiveRate float64 `json:"effective_rate"`
}

// Analyze handles GET /api/v1/rate/analyze
func (c *RateController) Analyze(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var sessions []models.TrackingSession
	db.Where("billable = ?", true).
		Preload("Client").
		Preload("Contract").
		Find(&sessions)

	clientStats := make(map[string]*ClientRateStats)
	var totalHours, totalRevenue float64

	for _, s := range sessions {
		clientName := "Unknown"
		if s.Client != nil {
			clientName = s.Client.Name
		}

		if clientStats[clientName] == nil {
			clientStats[clientName] = &ClientRateStats{ClientName: clientName}
		}

		hours := 0.0
		if s.Hours != nil {
			hours = *s.Hours
		}

		rate := 0.0
		if s.Contract != nil && s.Contract.HourlyRate != nil {
			rate = *s.Contract.HourlyRate
		}

		clientStats[clientName].Hours += hours
		clientStats[clientName].Revenue += hours * rate
		totalHours += hours
		totalRevenue += hours * rate
	}

	// Get paid invoices for the last year
	yearAgo := time.Now().AddDate(-1, 0, 0)
	var invoices []models.Invoice
	db.Where("status = ? AND updated_at >= ?", models.StatusPaid, yearAgo).Find(&invoices)

	var invoiceRevenue float64
	for _, inv := range invoices {
		invoiceRevenue += inv.Amount
	}

	analysis := RateAnalysis{
		TotalBillableHours:   totalHours,
		ContractBasedRevenue: totalRevenue,
		InvoiceRevenue:       invoiceRevenue,
	}

	if totalHours > 0 {
		analysis.EffectiveRateContracts = totalRevenue / totalHours
		analysis.EffectiveRateInvoices = invoiceRevenue / totalHours
	}

	for _, stats := range clientStats {
		if stats.Hours > 0 {
			stats.EffectiveRate = stats.Revenue / stats.Hours
		}
		analysis.ByClient = append(analysis.ByClient, *stats)
	}

	// Add recommendation
	if analysis.EffectiveRateContracts > 0 {
		analysis.Recommendation = "Consider raising rates if clients accept at " +
			"$" + formatFloat(analysis.EffectiveRateContracts*1.1) + "-$" + formatFloat(analysis.EffectiveRateContracts*1.25) + "/hr"
	}

	RespondJSON(w, analysis, http.StatusOK)
}

// RateComparison represents the output of rate comparison
type RateComparison struct {
	HourlyContracts []ContractRateInfo `json:"hourly_contracts"`
	Summary         struct {
		ContractCount int     `json:"contract_count"`
		MinRate       float64 `json:"min_rate"`
		MaxRate       float64 `json:"max_rate"`
		AverageRate   float64 `json:"average_rate"`
		RateSpread    float64 `json:"rate_spread"`
	} `json:"summary"`
	Last30Days struct {
		BillableHours   float64 `json:"billable_hours"`
		RevenueEarned   float64 `json:"revenue_earned"`
		ActualRate      float64 `json:"actual_rate"`
		VsAverageAmount float64 `json:"vs_average_amount"`
		VsAveragePercent float64 `json:"vs_average_percent"`
	} `json:"last_30_days"`
}

// ContractRateInfo represents rate information for a contract
type ContractRateInfo struct {
	ContractID   uint    `json:"contract_id"`
	ContractName string  `json:"contract_name"`
	ClientName   string  `json:"client_name"`
	ContractType string  `json:"contract_type"`
	HourlyRate   float64 `json:"hourly_rate,omitempty"`
	FixedPrice   float64 `json:"fixed_price,omitempty"`
}

// Compare handles GET /api/v1/rate/compare
func (c *RateController) Compare(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var contracts []models.Contract
	db.Where("active = ?", true).Preload("Client").Find(&contracts)

	comparison := RateComparison{}
	var rateSum float64

	for _, c := range contracts {
		info := ContractRateInfo{
			ContractID:   c.ID,
			ContractName: c.Name,
			ClientName:   c.Client.Name,
			ContractType: string(c.ContractType),
		}

		if c.ContractType == models.ContractTypeHourly && c.HourlyRate != nil {
			info.HourlyRate = *c.HourlyRate
			comparison.HourlyContracts = append(comparison.HourlyContracts, info)
			rateSum += *c.HourlyRate

			if comparison.Summary.MinRate == 0 || *c.HourlyRate < comparison.Summary.MinRate {
				comparison.Summary.MinRate = *c.HourlyRate
			}
			if *c.HourlyRate > comparison.Summary.MaxRate {
				comparison.Summary.MaxRate = *c.HourlyRate
			}
		} else if c.FixedPrice != nil {
			info.FixedPrice = *c.FixedPrice
		}
	}

	comparison.Summary.ContractCount = len(comparison.HourlyContracts)
	if comparison.Summary.ContractCount > 0 {
		comparison.Summary.AverageRate = rateSum / float64(comparison.Summary.ContractCount)
		comparison.Summary.RateSpread = comparison.Summary.MaxRate - comparison.Summary.MinRate
	}

	// Last 30 days analysis
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var sessions []models.TrackingSession
	db.Where("billable = ? AND start_time >= ?", true, thirtyDaysAgo).Find(&sessions)

	for _, s := range sessions {
		if s.Hours != nil {
			comparison.Last30Days.BillableHours += *s.Hours
		}
	}

	var paidInvoices []models.Invoice
	db.Where("status = ? AND updated_at >= ?", models.StatusPaid, thirtyDaysAgo).Find(&paidInvoices)

	for _, inv := range paidInvoices {
		comparison.Last30Days.RevenueEarned += inv.Amount
	}

	if comparison.Last30Days.BillableHours > 0 {
		comparison.Last30Days.ActualRate = comparison.Last30Days.RevenueEarned / comparison.Last30Days.BillableHours

		if comparison.Summary.AverageRate > 0 {
			comparison.Last30Days.VsAverageAmount = comparison.Last30Days.ActualRate - comparison.Summary.AverageRate
			comparison.Last30Days.VsAveragePercent = (comparison.Last30Days.VsAverageAmount / comparison.Summary.AverageRate) * 100
		}
	}

	RespondJSON(w, comparison, http.StatusOK)
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.0f", f)
}
