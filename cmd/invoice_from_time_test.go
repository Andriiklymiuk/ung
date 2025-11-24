package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestTimeSessionGroup_HourlyContract(t *testing.T) {
	hourlyRate := 150.0
	group := timeSessionGroup{
		ClientID:     1,
		ClientName:   "Test Client",
		ContractType: "hourly",
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		TotalHours:   10.5,
	}

	if group.ContractType != "hourly" {
		t.Errorf("Expected contract type 'hourly', got '%s'", group.ContractType)
	}

	if *group.HourlyRate != 150.0 {
		t.Errorf("Expected hourly rate 150.0, got %f", *group.HourlyRate)
	}

	expectedAmount := group.TotalHours * (*group.HourlyRate)
	if expectedAmount != 1575.0 {
		t.Errorf("Expected total 1575.0 (10.5 * 150), got %f", expectedAmount)
	}
}

func TestTimeSessionGroup_FixedPriceContract(t *testing.T) {
	fixedPrice := 5000.0
	group := timeSessionGroup{
		ClientID:     1,
		ClientName:   "Test Client",
		ContractType: "fixed_price",
		FixedPrice:   &fixedPrice,
		Currency:     "USD",
		TotalHours:   45.0,
	}

	if group.ContractType != "fixed_price" {
		t.Errorf("Expected contract type 'fixed_price', got '%s'", group.ContractType)
	}

	if *group.FixedPrice != 5000.0 {
		t.Errorf("Expected fixed price 5000.0, got %f", *group.FixedPrice)
	}

	// For fixed price, amount should be fixed price, not hours * rate
	if *group.FixedPrice != 5000.0 {
		t.Errorf("Expected amount 5000.0 regardless of hours, got %f", *group.FixedPrice)
	}
}

func TestGetUnbilledTimeSessions_ExcludesInvoiced(t *testing.T) {
	setupTestDB(t)

	// Create test data
	clientResult, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Test Client", "test@example.com")
	clientID, _ := clientResult.LastInsertId()

	contractResult, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.test.2025", clientID, "Test Contract", models.ContractTypeHourly, 150.0, "USD", true)
	contractID, _ := contractResult.LastInsertId()

	// Insert billable session without invoice marker
	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, clientID, contractID, "Project A", time.Now().Add(-2*time.Hour), time.Now(), 2.0, true, "Unbilled work")

	// Insert billable session WITH invoice marker (should be excluded)
	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, clientID, contractID, "Project B", time.Now().Add(-4*time.Hour), time.Now().Add(-2*time.Hour), 2.0, true, "Already billed [Invoiced: inv.test.123]")

	// Insert non-billable session (should be excluded)
	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, clientID, contractID, "Project C", time.Now().Add(-6*time.Hour), time.Now().Add(-4*time.Hour), 2.0, false, "Non-billable")

	groups, err := getUnbilledTimeSessions()
	if err != nil {
		t.Fatalf("getUnbilledTimeSessions failed: %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	if len(groups) > 0 {
		if groups[0].TotalHours != 2.0 {
			t.Errorf("Expected 2.0 total hours (only unbilled session), got %f", groups[0].TotalHours)
		}
	}
}

func TestGetUnbilledTimeSessions_GroupsByContract(t *testing.T) {
	setupTestDB(t)

	// Create test client
	clientResult, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Test Client", "test@example.com")
	clientID, _ := clientResult.LastInsertId()

	// Create two different contracts
	contract1Result, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.test.1.2025", clientID, "Contract 1", models.ContractTypeHourly, 100.0, "USD", true)
	contract1ID, _ := contract1Result.LastInsertId()

	contract2Result, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.test.2.2025", clientID, "Contract 2", models.ContractTypeHourly, 150.0, "USD", true)
	contract2ID, _ := contract2Result.LastInsertId()

	// Add sessions for both contracts
	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, clientID, contract1ID, "Project A", time.Now().Add(-2*time.Hour), time.Now(), 3.0, true)

	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, clientID, contract2ID, "Project B", time.Now().Add(-4*time.Hour), time.Now().Add(-2*time.Hour), 5.0, true)

	groups, err := getUnbilledTimeSessions()
	if err != nil {
		t.Fatalf("getUnbilledTimeSessions failed: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups (one per contract), got %d", len(groups))
	}

	// Verify each group has correct hours
	totalHours := 0.0
	for _, group := range groups {
		totalHours += group.TotalHours
	}

	if totalHours != 8.0 {
		t.Errorf("Expected total 8.0 hours across both groups, got %f", totalHours)
	}
}

func TestGetUnbilledTimeSessions_FixedPriceContract(t *testing.T) {
	setupTestDB(t)

	// Create test client
	clientResult, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Fixed Price Client", "test@example.com")
	clientID, _ := clientResult.LastInsertId()

	// Create fixed price contract
	fixedPrice := 5000.0
	contractResult, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, fixed_price, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.fixed.2025", clientID, "Fixed Contract", models.ContractTypeFixedPrice, fixedPrice, "USD", true)
	contractID, _ := contractResult.LastInsertId()

	// Add time sessions (hours are tracked but amount is fixed)
	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, clientID, contractID, "Project X", time.Now().Add(-3*time.Hour), time.Now(), 45.0, true)

	groups, err := getUnbilledTimeSessions()
	if err != nil {
		t.Fatalf("getUnbilledTimeSessions failed: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	group := groups[0]

	if group.ContractType != string(models.ContractTypeFixedPrice) {
		t.Errorf("Expected contract type fixed_price, got %s", group.ContractType)
	}

	if group.FixedPrice == nil {
		t.Fatal("Expected fixed price to be set")
	}

	if *group.FixedPrice != 5000.0 {
		t.Errorf("Expected fixed price 5000.0, got %f", *group.FixedPrice)
	}

	if group.TotalHours != 45.0 {
		t.Errorf("Expected 45.0 hours tracked, got %f", group.TotalHours)
	}
}

func TestGetUnbilledTimeSessionsForClient_FiltersCorrectly(t *testing.T) {
	setupTestDB(t)

	// Create two clients
	client1Result, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Client 1", "client1@example.com")
	client1ID, _ := client1Result.LastInsertId()

	client2Result, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Client 2", "client2@example.com")
	client2ID, _ := client2Result.LastInsertId()

	// Create contracts for both
	contract1Result, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.1.2025", client1ID, "Contract 1", models.ContractTypeHourly, 100.0, "USD", true)
	contract1ID, _ := contract1Result.LastInsertId()

	contract2Result, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.2.2025", client2ID, "Contract 2", models.ContractTypeHourly, 150.0, "USD", true)
	contract2ID, _ := contract2Result.LastInsertId()

	// Add sessions for both clients
	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, client1ID, contract1ID, "Project A", time.Now(), time.Now(), 5.0, true)

	db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, contract_id, project_name, start_time, end_time, hours, billable)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, client2ID, contract2ID, "Project B", time.Now(), time.Now(), 10.0, true)

	// Get sessions for client 1 only
	groups, err := getUnbilledTimeSessionsForClient(uint(client1ID))
	if err != nil {
		t.Fatalf("getUnbilledTimeSessionsForClient failed: %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group for client 1, got %d", len(groups))
	}

	if len(groups) > 0 && groups[0].TotalHours != 5.0 {
		t.Errorf("Expected 5.0 hours for client 1, got %f", groups[0].TotalHours)
	}
}
