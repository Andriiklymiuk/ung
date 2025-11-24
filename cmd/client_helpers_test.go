package cmd

import (
	"testing"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestSearchClients_SingleMatch(t *testing.T) {
	setupTestDB(t)

	// Insert test client
	db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Humbrella Corporation", "contact@humbrella.com")

	matches, err := searchClients("humbrella")
	if err != nil {
		t.Fatalf("searchClients failed: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}

	if matches[0].Name != "Humbrella Corporation" {
		t.Errorf("Expected 'Humbrella Corporation', got '%s'", matches[0].Name)
	}
}

func TestSearchClients_MultipleMatches(t *testing.T) {
	setupTestDB(t)

	// Insert multiple test clients with similar names
	clients := []string{"Humbrella Corporation", "Humberto Industries", "Humdinger LLC"}
	for _, name := range clients {
		db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", name, name+"@example.com")
	}

	matches, err := searchClients("hum")
	if err != nil {
		t.Fatalf("searchClients failed: %v", err)
	}

	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// Verify alphabetical ordering
	expectedOrder := []string{"Humberto Industries", "Humbrella Corporation", "Humdinger LLC"}
	for i, expected := range expectedOrder {
		if matches[i].Name != expected {
			t.Errorf("Expected match[%d] to be '%s', got '%s'", i, expected, matches[i].Name)
		}
	}
}

func TestSearchClients_NoMatch(t *testing.T) {
	setupTestDB(t)

	db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Acme Corp", "contact@acme.com")

	matches, err := searchClients("nonexistent")
	if err != nil {
		t.Fatalf("searchClients failed: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestSearchClients_CaseInsensitive(t *testing.T) {
	setupTestDB(t)

	db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Humbrella Corporation", "contact@humbrella.com")

	testCases := []string{"HUMBRELLA", "humbrella", "HuMbReLLa", "hum"}
	for _, searchTerm := range testCases {
		matches, err := searchClients(searchTerm)
		if err != nil {
			t.Fatalf("searchClients failed for '%s': %v", searchTerm, err)
		}

		if len(matches) != 1 {
			t.Errorf("Search term '%s' expected 1 match, got %d", searchTerm, len(matches))
		}
	}
}

func TestSearchClients_PartialMatch(t *testing.T) {
	setupTestDB(t)

	db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Humbrella Corporation", "contact@humbrella.com")

	// Should match anywhere in the name
	partialTerms := []string{"hum", "brella", "corp", "ration"}
	for _, term := range partialTerms {
		matches, err := searchClients(term)
		if err != nil {
			t.Fatalf("searchClients failed for '%s': %v", term, err)
		}

		if len(matches) != 1 {
			t.Errorf("Partial term '%s' expected 1 match, got %d", term, len(matches))
		}
	}
}

func TestFindActiveContractForClient_Success(t *testing.T) {
	setupTestDB(t)

	// Insert client
	result, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Test Client", "test@example.com")
	clientID, _ := result.LastInsertId()

	// Insert active contract
	contractResult, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.test.2025", clientID, "Test Contract", models.ContractTypeHourly, 150.0, "USD", true)
	expectedContractID, _ := contractResult.LastInsertId()

	contractID, err := FindActiveContractForClient(uint(clientID))
	if err != nil {
		t.Fatalf("FindActiveContractForClient failed: %v", err)
	}

	if int64(contractID) != expectedContractID {
		t.Errorf("Expected contract ID %d, got %d", expectedContractID, contractID)
	}
}

func TestFindActiveContractForClient_NoActiveContract(t *testing.T) {
	setupTestDB(t)

	// Insert client
	result, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Test Client", "test@example.com")
	clientID, _ := result.LastInsertId()

	// Insert inactive contract
	db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.test.2025", clientID, "Test Contract", models.ContractTypeHourly, 150.0, "USD", false)

	_, err := FindActiveContractForClient(uint(clientID))
	if err == nil {
		t.Error("Expected error for client with no active contract")
	}
}

func TestFindActiveContractForClient_MultipleContracts(t *testing.T) {
	setupTestDB(t)

	// Insert client
	result, _ := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", "Test Client", "test@example.com")
	clientID, _ := result.LastInsertId()

	// Insert older contract
	db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now', '-1 month'), ?)
	`, "contract.test.old.2025", clientID, "Old Contract", models.ContractTypeHourly, 100.0, "USD", true)

	// Insert newer contract (should be selected)
	newerResult, _ := db.DB.Exec(`
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, date('now'), ?)
	`, "contract.test.new.2025", clientID, "New Contract", models.ContractTypeHourly, 150.0, "USD", true)
	expectedContractID, _ := newerResult.LastInsertId()

	contractID, err := FindActiveContractForClient(uint(clientID))
	if err != nil {
		t.Fatalf("FindActiveContractForClient failed: %v", err)
	}

	if int64(contractID) != expectedContractID {
		t.Errorf("Expected most recent contract ID %d, got %d", expectedContractID, contractID)
	}
}

func TestClientMatch_Struct(t *testing.T) {
	match := clientMatch{
		ID:   123,
		Name: "Test Client",
	}

	if match.ID != 123 {
		t.Errorf("Expected ID 123, got %d", match.ID)
	}

	if match.Name != "Test Client" {
		t.Errorf("Expected Name 'Test Client', got '%s'", match.Name)
	}
}
