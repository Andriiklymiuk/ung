package controllers

import (
	"net/http"
	"sort"
	"time"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
)

// ReportController handles report endpoints
type ReportController struct{}

// NewReportController creates a new report controller
func NewReportController() *ReportController {
	return &ReportController{}
}

// WeeklyReportData represents weekly report data
type WeeklyReportData struct {
	WeekStart    string                `json:"week_start"`
	WeekEnd      string                `json:"week_end"`
	TotalHours   float64               `json:"total_hours"`
	TotalRevenue float64               `json:"total_revenue"`
	DayBreakdown []DayData             `json:"day_breakdown"`
	ByClient     []ClientReportData    `json:"by_client"`
	ByContract   []ContractReportData  `json:"by_contract"`
	Sessions     int                   `json:"sessions"`
}

// DayData represents daily breakdown
type DayData struct {
	Date   string  `json:"date"`
	Day    string  `json:"day"`
	Hours  float64 `json:"hours"`
	Amount float64 `json:"amount"`
}

// ClientReportData represents client-specific report data
type ClientReportData struct {
	ClientID   uint    `json:"client_id"`
	ClientName string  `json:"client_name"`
	Hours      float64 `json:"hours"`
	Revenue    float64 `json:"revenue"`
	Sessions   int     `json:"sessions"`
}

// ContractReportData represents contract-specific report data
type ContractReportData struct {
	ContractID   uint    `json:"contract_id"`
	ContractName string  `json:"contract_name"`
	ClientName   string  `json:"client_name"`
	Hours        float64 `json:"hours"`
	Revenue      float64 `json:"revenue"`
	HourlyRate   float64 `json:"hourly_rate"`
}

// Weekly handles GET /api/v1/reports/weekly
func (c *ReportController) Weekly(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Calculate week boundaries
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -weekday+1)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, now.Location())
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 0, now.Location())

	report := WeeklyReportData{
		WeekStart:    weekStart.Format("2006-01-02"),
		WeekEnd:      weekEnd.Format("2006-01-02"),
		DayBreakdown: make([]DayData, 7),
		ByClient:     []ClientReportData{},
		ByContract:   []ContractReportData{},
	}

	// Initialize day breakdown
	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		report.DayBreakdown[i] = DayData{
			Date:  day.Format("2006-01-02"),
			Day:   day.Format("Mon"),
			Hours: 0,
		}
	}

	// Fetch tracking sessions for the week
	var sessions []models.TrackingSession
	db.Preload("Client").Preload("Contract").
		Where("start_time >= ? AND start_time <= ? AND billable = ?", weekStart, weekEnd, true).
		Find(&sessions)

	report.Sessions = len(sessions)

	// Client and contract aggregation maps
	clientData := make(map[uint]*ClientReportData)
	contractData := make(map[uint]*ContractReportData)

	for _, session := range sessions {
		hours := float64(0)
		if session.Hours != nil {
			hours = *session.Hours
		}

		report.TotalHours += hours

		// Calculate revenue if contract has hourly rate
		var revenue float64
		var hourlyRate float64
		if session.Contract != nil && session.Contract.HourlyRate != nil {
			hourlyRate = *session.Contract.HourlyRate
			revenue = hours * hourlyRate
			report.TotalRevenue += revenue
		}

		// Day breakdown
		dayIndex := int(session.StartTime.Weekday())
		if dayIndex == 0 {
			dayIndex = 6
		} else {
			dayIndex--
		}
		if dayIndex >= 0 && dayIndex < 7 {
			report.DayBreakdown[dayIndex].Hours += hours
			report.DayBreakdown[dayIndex].Amount += revenue
		}

		// Client aggregation
		if session.ClientID != nil && session.Client != nil {
			if _, ok := clientData[*session.ClientID]; !ok {
				clientData[*session.ClientID] = &ClientReportData{
					ClientID:   *session.ClientID,
					ClientName: session.Client.Name,
				}
			}
			clientData[*session.ClientID].Hours += hours
			clientData[*session.ClientID].Revenue += revenue
			clientData[*session.ClientID].Sessions++
		}

		// Contract aggregation
		if session.ContractID != nil && session.Contract != nil {
			if _, ok := contractData[*session.ContractID]; !ok {
				clientName := ""
				if session.Client != nil {
					clientName = session.Client.Name
				}
				contractData[*session.ContractID] = &ContractReportData{
					ContractID:   *session.ContractID,
					ContractName: session.Contract.Name,
					ClientName:   clientName,
					HourlyRate:   hourlyRate,
				}
			}
			contractData[*session.ContractID].Hours += hours
			contractData[*session.ContractID].Revenue += revenue
		}
	}

	// Convert maps to slices
	for _, cd := range clientData {
		report.ByClient = append(report.ByClient, *cd)
	}
	for _, ctd := range contractData {
		report.ByContract = append(report.ByContract, *ctd)
	}

	// Sort by hours descending
	sort.Slice(report.ByClient, func(i, j int) bool {
		return report.ByClient[i].Hours > report.ByClient[j].Hours
	})
	sort.Slice(report.ByContract, func(i, j int) bool {
		return report.ByContract[i].Hours > report.ByContract[j].Hours
	})

	RespondJSON(w, report, http.StatusOK)
}

// MonthlyReportData represents monthly report data
type MonthlyReportData struct {
	Month         string               `json:"month"`
	Year          int                  `json:"year"`
	TotalHours    float64              `json:"total_hours"`
	TotalRevenue  float64              `json:"total_revenue"`
	TotalExpenses float64              `json:"total_expenses"`
	Profit        float64              `json:"profit"`
	InvoicesSent  int                  `json:"invoices_sent"`
	InvoicesPaid  int                  `json:"invoices_paid"`
	ByClient      []ClientReportData   `json:"by_client"`
	ByContract    []ContractReportData `json:"by_contract"`
	WeekBreakdown []WeekData           `json:"week_breakdown"`
}

// WeekData represents weekly breakdown within monthly report
type WeekData struct {
	WeekNum int     `json:"week_num"`
	Hours   float64 `json:"hours"`
	Revenue float64 `json:"revenue"`
}

// Monthly handles GET /api/v1/reports/monthly
func (c *ReportController) Monthly(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	report := MonthlyReportData{
		Month:         monthStart.Format("January"),
		Year:          now.Year(),
		ByClient:      []ClientReportData{},
		ByContract:    []ContractReportData{},
		WeekBreakdown: []WeekData{},
	}

	// Fetch tracking sessions
	var sessions []models.TrackingSession
	db.Preload("Client").Preload("Contract").
		Where("start_time >= ? AND start_time <= ? AND billable = ?", monthStart, monthEnd, true).
		Find(&sessions)

	clientData := make(map[uint]*ClientReportData)
	contractData := make(map[uint]*ContractReportData)
	weekData := make(map[int]*WeekData)

	for _, session := range sessions {
		hours := float64(0)
		if session.Hours != nil {
			hours = *session.Hours
		}
		report.TotalHours += hours

		var revenue float64
		var hourlyRate float64
		if session.Contract != nil && session.Contract.HourlyRate != nil {
			hourlyRate = *session.Contract.HourlyRate
			revenue = hours * hourlyRate
			report.TotalRevenue += revenue
		}

		// Week breakdown
		_, weekNum := session.StartTime.ISOWeek()
		if _, ok := weekData[weekNum]; !ok {
			weekData[weekNum] = &WeekData{WeekNum: weekNum}
		}
		weekData[weekNum].Hours += hours
		weekData[weekNum].Revenue += revenue

		// Client aggregation
		if session.ClientID != nil && session.Client != nil {
			if _, ok := clientData[*session.ClientID]; !ok {
				clientData[*session.ClientID] = &ClientReportData{
					ClientID:   *session.ClientID,
					ClientName: session.Client.Name,
				}
			}
			clientData[*session.ClientID].Hours += hours
			clientData[*session.ClientID].Revenue += revenue
			clientData[*session.ClientID].Sessions++
		}

		// Contract aggregation
		if session.ContractID != nil && session.Contract != nil {
			if _, ok := contractData[*session.ContractID]; !ok {
				clientName := ""
				if session.Client != nil {
					clientName = session.Client.Name
				}
				contractData[*session.ContractID] = &ContractReportData{
					ContractID:   *session.ContractID,
					ContractName: session.Contract.Name,
					ClientName:   clientName,
					HourlyRate:   hourlyRate,
				}
			}
			contractData[*session.ContractID].Hours += hours
			contractData[*session.ContractID].Revenue += revenue
		}
	}

	// Fetch expenses
	var expenses []models.Expense
	db.Where("date >= ? AND date <= ?", monthStart, monthEnd).Find(&expenses)
	for _, exp := range expenses {
		report.TotalExpenses += exp.Amount
	}
	report.Profit = report.TotalRevenue - report.TotalExpenses

	// Fetch invoices
	var sentInvoices []models.Invoice
	db.Where("status IN ? AND issued_date >= ? AND issued_date <= ?",
		[]models.InvoiceStatus{models.StatusSent, models.StatusPaid}, monthStart, monthEnd).Find(&sentInvoices)
	report.InvoicesSent = len(sentInvoices)

	var paidInvoices []models.Invoice
	db.Where("status = ? AND updated_at >= ? AND updated_at <= ?",
		models.StatusPaid, monthStart, monthEnd).Find(&paidInvoices)
	report.InvoicesPaid = len(paidInvoices)

	// Convert maps to slices
	for _, cd := range clientData {
		report.ByClient = append(report.ByClient, *cd)
	}
	for _, ctd := range contractData {
		report.ByContract = append(report.ByContract, *ctd)
	}
	for _, wd := range weekData {
		report.WeekBreakdown = append(report.WeekBreakdown, *wd)
	}

	sort.Slice(report.ByClient, func(i, j int) bool {
		return report.ByClient[i].Hours > report.ByClient[j].Hours
	})
	sort.Slice(report.ByContract, func(i, j int) bool {
		return report.ByContract[i].Hours > report.ByContract[j].Hours
	})
	sort.Slice(report.WeekBreakdown, func(i, j int) bool {
		return report.WeekBreakdown[i].WeekNum < report.WeekBreakdown[j].WeekNum
	})

	RespondJSON(w, report, http.StatusOK)
}

// RevenueReportData represents revenue report data
type RevenueReportData struct {
	Period        string               `json:"period"`
	TotalRevenue  float64              `json:"total_revenue"`
	PaidRevenue   float64              `json:"paid_revenue"`
	PendingRevenue float64             `json:"pending_revenue"`
	OverdueRevenue float64             `json:"overdue_revenue"`
	ByClient      []ClientRevenueData  `json:"by_client"`
	MonthlyTrend  []MonthlyRevenueData `json:"monthly_trend"`
}

// ClientRevenueData represents client revenue breakdown
type ClientRevenueData struct {
	ClientName string  `json:"client_name"`
	Total      float64 `json:"total"`
	Paid       float64 `json:"paid"`
	Pending    float64 `json:"pending"`
	Overdue    float64 `json:"overdue"`
}

// MonthlyRevenueData represents monthly revenue trend
type MonthlyRevenueData struct {
	Month    string  `json:"month"`
	Revenue  float64 `json:"revenue"`
	Paid     float64 `json:"paid"`
	Invoices int     `json:"invoices"`
}

// Revenue handles GET /api/v1/reports/revenue
func (c *ReportController) Revenue(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	report := RevenueReportData{
		Period:       now.Format("2006"),
		ByClient:     []ClientRevenueData{},
		MonthlyTrend: []MonthlyRevenueData{},
	}

	// Fetch all invoices for the year
	var invoices []models.Invoice
	db.Where("issued_date >= ?", startOfYear).Find(&invoices)

	// Build client revenue map using invoice recipients
	clientRevenue := make(map[string]*ClientRevenueData)

	for _, inv := range invoices {
		report.TotalRevenue += inv.Amount

		switch inv.Status {
		case models.StatusPaid:
			report.PaidRevenue += inv.Amount
		case models.StatusPending:
			report.PendingRevenue += inv.Amount
		case models.StatusOverdue:
			report.OverdueRevenue += inv.Amount
		}

		// Get invoice recipients for client breakdown
		var recipients []models.InvoiceRecipient
		db.Preload("Client").Where("invoice_id = ?", inv.ID).Find(&recipients)

		for _, rec := range recipients {
			var client models.Client
			if err := db.First(&client, rec.ClientID).Error; err != nil {
				continue
			}

			if _, ok := clientRevenue[client.Name]; !ok {
				clientRevenue[client.Name] = &ClientRevenueData{ClientName: client.Name}
			}
			clientRevenue[client.Name].Total += inv.Amount

			switch inv.Status {
			case models.StatusPaid:
				clientRevenue[client.Name].Paid += inv.Amount
			case models.StatusPending:
				clientRevenue[client.Name].Pending += inv.Amount
			case models.StatusOverdue:
				clientRevenue[client.Name].Overdue += inv.Amount
			}
		}
	}

	// Monthly trend
	for m := 1; m <= int(now.Month()); m++ {
		monthStart := time.Date(now.Year(), time.Month(m), 1, 0, 0, 0, 0, now.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var monthInvoices []models.Invoice
		db.Where("issued_date >= ? AND issued_date <= ?", monthStart, monthEnd).Find(&monthInvoices)

		monthData := MonthlyRevenueData{
			Month:    monthStart.Format("Jan"),
			Invoices: len(monthInvoices),
		}

		for _, inv := range monthInvoices {
			monthData.Revenue += inv.Amount
			if inv.Status == models.StatusPaid {
				monthData.Paid += inv.Amount
			}
		}

		report.MonthlyTrend = append(report.MonthlyTrend, monthData)
	}

	// Convert map to slice
	for _, cr := range clientRevenue {
		report.ByClient = append(report.ByClient, *cr)
	}

	sort.Slice(report.ByClient, func(i, j int) bool {
		return report.ByClient[i].Total > report.ByClient[j].Total
	})

	RespondJSON(w, report, http.StatusOK)
}

// ClientsReportData represents clients report
type ClientsReportData struct {
	TotalClients    int                   `json:"total_clients"`
	ActiveClients   int                   `json:"active_clients"`
	ClientDetails   []ClientDetailData    `json:"client_details"`
}

// ClientDetailData represents detailed client info for report
type ClientDetailData struct {
	ClientID        uint    `json:"client_id"`
	ClientName      string  `json:"client_name"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalHours      float64 `json:"total_hours"`
	ActiveContracts int     `json:"active_contracts"`
	LastInvoice     string  `json:"last_invoice"`
	LastActivity    string  `json:"last_activity"`
}

// Clients handles GET /api/v1/reports/clients
func (c *ReportController) Clients(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	report := ClientsReportData{
		ClientDetails: []ClientDetailData{},
	}

	var clients []models.Client
	db.Find(&clients)
	report.TotalClients = len(clients)

	for _, client := range clients {
		detail := ClientDetailData{
			ClientID:   client.ID,
			ClientName: client.Name,
		}

		// Count active contracts
		var contractCount int64
		db.Model(&models.Contract{}).Where("client_id = ? AND active = ?", client.ID, true).Count(&contractCount)
		detail.ActiveContracts = int(contractCount)

		// Total hours from tracking
		var sessions []models.TrackingSession
		db.Where("client_id = ?", client.ID).Find(&sessions)
		for _, s := range sessions {
			if s.Hours != nil {
				detail.TotalHours += *s.Hours
			}
		}

		// Total revenue from invoices
		var invoiceRecipients []models.InvoiceRecipient
		db.Preload("Invoice").Where("client_id = ?", client.ID).Find(&invoiceRecipients)
		for _, ir := range invoiceRecipients {
			var inv models.Invoice
			if err := db.First(&inv, ir.InvoiceID).Error; err == nil {
				detail.TotalRevenue += inv.Amount
				if detail.LastInvoice == "" || inv.IssuedDate.Format("2006-01-02") > detail.LastInvoice {
					detail.LastInvoice = inv.IssuedDate.Format("2006-01-02")
				}
			}
		}

		// Last activity
		var lastSession models.TrackingSession
		if err := db.Where("client_id = ?", client.ID).Order("start_time DESC").First(&lastSession).Error; err == nil {
			detail.LastActivity = lastSession.StartTime.Format("2006-01-02")
		}

		if detail.ActiveContracts > 0 || detail.TotalHours > 0 {
			report.ActiveClients++
		}

		report.ClientDetails = append(report.ClientDetails, detail)
	}

	sort.Slice(report.ClientDetails, func(i, j int) bool {
		return report.ClientDetails[i].TotalRevenue > report.ClientDetails[j].TotalRevenue
	})

	RespondJSON(w, report, http.StatusOK)
}

// OverdueInvoiceData represents overdue invoice data
type OverdueInvoiceData struct {
	ID           uint    `json:"id"`
	InvoiceNum   string  `json:"invoice_num"`
	ClientName   string  `json:"client_name"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	DueDate      string  `json:"due_date"`
	DaysOverdue  int     `json:"days_overdue"`
}

// OverdueReportData represents overdue invoices report
type OverdueReportData struct {
	TotalOverdue float64              `json:"total_overdue"`
	Count        int                  `json:"count"`
	Invoices     []OverdueInvoiceData `json:"invoices"`
}

// Overdue handles GET /api/v1/reports/overdue
func (c *ReportController) Overdue(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	now := time.Now()
	report := OverdueReportData{
		Invoices: []OverdueInvoiceData{},
	}

	// Update overdue status first
	db.Model(&models.Invoice{}).
		Where("status = ? AND due_date < ?", models.StatusPending, now).
		Update("status", models.StatusOverdue)

	// Fetch overdue invoices
	var invoices []models.Invoice
	db.Where("status = ?", models.StatusOverdue).Order("due_date ASC").Find(&invoices)

	report.Count = len(invoices)

	for _, inv := range invoices {
		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)

		// Get client name from recipients
		clientName := ""
		var recipients []models.InvoiceRecipient
		db.Where("invoice_id = ?", inv.ID).Find(&recipients)
		if len(recipients) > 0 {
			var client models.Client
			if err := db.First(&client, recipients[0].ClientID).Error; err == nil {
				clientName = client.Name
			}
		}

		report.TotalOverdue += inv.Amount
		report.Invoices = append(report.Invoices, OverdueInvoiceData{
			ID:          inv.ID,
			InvoiceNum:  inv.InvoiceNum,
			ClientName:  clientName,
			Amount:      inv.Amount,
			Currency:    inv.Currency,
			DueDate:     inv.DueDate.Format("2006-01-02"),
			DaysOverdue: daysOverdue,
		})
	}

	RespondJSON(w, report, http.StatusOK)
}

// UnpaidInvoiceData represents unpaid invoice data
type UnpaidInvoiceData struct {
	ID          uint    `json:"id"`
	InvoiceNum  string  `json:"invoice_num"`
	ClientName  string  `json:"client_name"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	IssuedDate  string  `json:"issued_date"`
	DueDate     string  `json:"due_date"`
	Status      string  `json:"status"`
	DaysUntilDue int    `json:"days_until_due"`
}

// UnpaidReportData represents unpaid invoices report
type UnpaidReportData struct {
	TotalUnpaid float64             `json:"total_unpaid"`
	Pending     float64             `json:"pending"`
	Overdue     float64             `json:"overdue"`
	Count       int                 `json:"count"`
	Invoices    []UnpaidInvoiceData `json:"invoices"`
}

// Unpaid handles GET /api/v1/reports/unpaid
func (c *ReportController) Unpaid(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	now := time.Now()
	report := UnpaidReportData{
		Invoices: []UnpaidInvoiceData{},
	}

	// Fetch pending and overdue invoices
	var invoices []models.Invoice
	db.Where("status IN ?", []models.InvoiceStatus{models.StatusPending, models.StatusOverdue}).
		Order("due_date ASC").Find(&invoices)

	report.Count = len(invoices)

	for _, inv := range invoices {
		daysUntilDue := int(inv.DueDate.Sub(now).Hours() / 24)

		// Get client name
		clientName := ""
		var recipients []models.InvoiceRecipient
		db.Where("invoice_id = ?", inv.ID).Find(&recipients)
		if len(recipients) > 0 {
			var client models.Client
			if err := db.First(&client, recipients[0].ClientID).Error; err == nil {
				clientName = client.Name
			}
		}

		report.TotalUnpaid += inv.Amount
		if inv.Status == models.StatusPending {
			report.Pending += inv.Amount
		} else {
			report.Overdue += inv.Amount
		}

		report.Invoices = append(report.Invoices, UnpaidInvoiceData{
			ID:           inv.ID,
			InvoiceNum:   inv.InvoiceNum,
			ClientName:   clientName,
			Amount:       inv.Amount,
			Currency:     inv.Currency,
			IssuedDate:   inv.IssuedDate.Format("2006-01-02"),
			DueDate:      inv.DueDate.Format("2006-01-02"),
			Status:       string(inv.Status),
			DaysUntilDue: daysUntilDue,
		})
	}

	RespondJSON(w, report, http.StatusOK)
}
