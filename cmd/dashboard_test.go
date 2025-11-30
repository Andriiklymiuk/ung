package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
)

func TestDashboardRevenueCalculation(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline
	var baselinePaid float64
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM invoices WHERE status = 'paid'").Scan(&baselinePaid)

	// Setup company
	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Dashboard Rev Test Co', 'dashrevtest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	// Create paid invoices
	invoices := []struct {
		amount float64
		status string
	}{
		{1000.00, "paid"},
		{500.00, "paid"},
		{750.00, "paid"},
		{300.00, "pending"}, // Not paid, shouldn't count
	}

	for i, inv := range invoices {
		_, err := db.DB.Exec(`
			INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
			VALUES (?, ?, ?, 'USD', ?, ?, ?)
		`, companyID, "INV-DASHREV-"+string(rune('0'+i)), inv.amount, inv.status, time.Now(), time.Now().AddDate(0, 0, 30))
		if err != nil {
			t.Fatalf("Failed to create invoice: %v", err)
		}
	}

	// Calculate paid revenue
	var paidTotal float64
	err := db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM invoices WHERE status = 'paid'").
		Scan(&paidTotal)
	if err != nil {
		t.Fatalf("Failed to sum paid invoices: %v", err)
	}

	expectedIncrease := 1000.00 + 500.00 + 750.00
	expected := baselinePaid + expectedIncrease
	if paidTotal != expected {
		t.Errorf("Expected paid total %f, got %f", expected, paidTotal)
	}
}

func TestDashboardPendingInvoices(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline
	var baselinePending int
	db.DB.QueryRow("SELECT COUNT(*) FROM invoices WHERE status IN ('pending', 'sent', 'overdue')").Scan(&baselinePending)

	companyResult, _ := db.DB.Exec("INSERT INTO companies (name, email) VALUES ('Dashboard Pending Test Co', 'dashpendingtest@test.com')")
	companyID, _ := companyResult.LastInsertId()

	// Create mix of invoices
	statuses := []string{"pending", "pending", "sent", "paid", "overdue"}

	for i, status := range statuses {
		_, err := db.DB.Exec(`
			INSERT INTO invoices (company_id, invoice_num, amount, currency, status, issued_date, due_date)
			VALUES (?, ?, 100.00, 'USD', ?, ?, ?)
		`, companyID, "INV-DASHPEND-"+string(rune('0'+i)), status, time.Now(), time.Now().AddDate(0, 0, 30))
		if err != nil {
			t.Fatalf("Failed to create invoice: %v", err)
		}
	}

	// Count pending invoices
	var pendingCount int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM invoices WHERE status IN ('pending', 'sent', 'overdue')
	`).Scan(&pendingCount)
	if err != nil {
		t.Fatalf("Failed to count pending: %v", err)
	}

	expectedIncrease := 4 // 2 pending + 1 sent + 1 overdue
	expected := baselinePending + expectedIncrease
	if pendingCount != expected {
		t.Errorf("Expected %d pending invoices, got %d", expected, pendingCount)
	}
}

func TestDashboardActiveTracking(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline
	var baselineActive int
	db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE end_time IS NULL AND deleted_at IS NULL").Scan(&baselineActive)

	// Start a session
	_, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, billable)
		VALUES ('Dashboard Active Test', ?, 1)
	`, time.Now())
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}

	// Check active session exists
	var activeCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE end_time IS NULL AND deleted_at IS NULL").
		Scan(&activeCount)
	if err != nil {
		t.Fatalf("Failed to count active sessions: %v", err)
	}

	expected := baselineActive + 1
	if activeCount != expected {
		t.Errorf("Expected %d active session(s), got %d", expected, activeCount)
	}
}

func TestDashboardHoursTracked(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline
	var baselineHours float64
	db.DB.QueryRow("SELECT COALESCE(SUM(hours), 0) FROM tracking_sessions WHERE deleted_at IS NULL").Scan(&baselineHours)

	// Add completed tracking sessions
	sessions := []float64{2.0, 3.5, 1.5, 4.0}

	for i, hours := range sessions {
		duration := int(hours * 3600)
		startTime := time.Now().Add(-time.Duration(duration) * time.Second)
		_, err := db.DB.Exec(`
			INSERT INTO tracking_sessions (project_name, start_time, end_time, duration, hours, billable)
			VALUES (?, ?, ?, ?, ?, 1)
		`, "Dashboard Hours Task "+string(rune('0'+i)), startTime, time.Now(), duration, hours)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	// Calculate total hours
	var totalHours float64
	err := db.DB.QueryRow(`
		SELECT COALESCE(SUM(hours), 0) FROM tracking_sessions WHERE deleted_at IS NULL
	`).Scan(&totalHours)
	if err != nil {
		t.Fatalf("Failed to sum hours: %v", err)
	}

	expectedIncrease := 2.0 + 3.5 + 1.5 + 4.0
	expected := baselineHours + expectedIncrease
	if totalHours != expected {
		t.Errorf("Expected total hours %f, got %f", expected, totalHours)
	}
}

func TestDashboardBillableVsNonBillable(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baselines
	var baseBillable, baseNonBillable float64
	db.DB.QueryRow("SELECT COALESCE(SUM(hours), 0) FROM tracking_sessions WHERE billable = 1 AND deleted_at IS NULL").Scan(&baseBillable)
	db.DB.QueryRow("SELECT COALESCE(SUM(hours), 0) FROM tracking_sessions WHERE billable = 0 AND deleted_at IS NULL").Scan(&baseNonBillable)

	// Add billable sessions
	for i := 0; i < 3; i++ {
		_, err := db.DB.Exec(`
			INSERT INTO tracking_sessions (project_name, start_time, end_time, duration, hours, billable)
			VALUES (?, ?, ?, 3600, 1.0, 1)
		`, "Billable "+string(rune('0'+i)), time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			t.Fatalf("Failed to create billable session: %v", err)
		}
	}

	// Add non-billable sessions
	for i := 0; i < 2; i++ {
		_, err := db.DB.Exec(`
			INSERT INTO tracking_sessions (project_name, start_time, end_time, duration, hours, billable)
			VALUES (?, ?, ?, 1800, 0.5, 0)
		`, "Non-billable "+string(rune('0'+i)), time.Now().Add(-30*time.Minute), time.Now())
		if err != nil {
			t.Fatalf("Failed to create non-billable session: %v", err)
		}
	}

	// Count billable hours
	var billableHours, nonBillableHours float64
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(hours), 0) FROM tracking_sessions WHERE billable = 1 AND deleted_at IS NULL
	`).Scan(&billableHours)
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(hours), 0) FROM tracking_sessions WHERE billable = 0 AND deleted_at IS NULL
	`).Scan(&nonBillableHours)

	expectedBillable := baseBillable + 3.0
	expectedNonBillable := baseNonBillable + 1.0

	if billableHours != expectedBillable {
		t.Errorf("Expected %f billable hours, got %f", expectedBillable, billableHours)
	}
	if nonBillableHours != expectedNonBillable {
		t.Errorf("Expected %f non-billable hours, got %f", expectedNonBillable, nonBillableHours)
	}
}

func TestDashboardExpenseSummary(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline
	var baselineExpenses float64
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses").Scan(&baselineExpenses)

	// Add expenses
	expenses := []struct {
		amount   float64
		category string
	}{
		{99.99, "software"},
		{50.00, "software"},
		{200.00, "hardware"},
		{35.00, "meals"},
	}

	for _, exp := range expenses {
		_, err := db.DB.Exec(`
			INSERT INTO expenses (description, amount, currency, category, date)
			VALUES ('Dashboard Expense Test', ?, 'USD', ?, ?)
		`, exp.amount, exp.category, time.Now())
		if err != nil {
			t.Fatalf("Failed to add expense: %v", err)
		}
	}

	// Get total expenses
	var totalExpenses float64
	err := db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses").Scan(&totalExpenses)
	if err != nil {
		t.Fatalf("Failed to sum expenses: %v", err)
	}

	expectedIncrease := 99.99 + 50.00 + 200.00 + 35.00
	expected := baselineExpenses + expectedIncrease
	if totalExpenses != expected {
		t.Errorf("Expected total expenses %f, got %f", expected, totalExpenses)
	}
}

func TestDashboardContractCount(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Get baseline
	var baselineActive int
	db.DB.QueryRow("SELECT COUNT(*) FROM contracts WHERE active = 1").Scan(&baselineActive)

	// Setup
	clientResult, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES ('Dashboard Contract Client', 'dashcontract@test.com')")
	clientID, _ := clientResult.LastInsertId()

	// Add contracts
	contracts := []struct {
		name   string
		active bool
	}{
		{"Active Contract 1", true},
		{"Active Contract 2", true},
		{"Inactive Contract", false},
	}

	for i, c := range contracts {
		_, err := db.DB.Exec(`
			INSERT INTO contracts (client_id, name, contract_type, contract_num, active, start_date)
			VALUES (?, ?, 'hourly', ?, ?, ?)
		`, clientID, c.name, "DASH-C-00"+string(rune('1'+i)), c.active, time.Now())
		if err != nil {
			t.Fatalf("Failed to add contract: %v", err)
		}
	}

	// Count active contracts
	var activeCount int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM contracts WHERE active = 1").Scan(&activeCount)
	if err != nil {
		t.Fatalf("Failed to count active contracts: %v", err)
	}

	expectedActive := baselineActive + 2
	if activeCount != expectedActive {
		t.Errorf("Expected %d active contracts, got %d", expectedActive, activeCount)
	}
}
