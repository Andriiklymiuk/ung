package cmd

import (
	"testing"

	"github.com/Andriiklymiuk/ung/internal/db"
)

func TestClientCRUD(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Test Create
	result, err := db.DB.Exec(`
		INSERT INTO clients (name, email, address, tax_id)
		VALUES ('Test Client', 'test@example.com', '123 Test St', 'TAX123')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test client: %v", err)
	}

	id, _ := result.LastInsertId()
	if id < 1 {
		t.Errorf("Expected client ID >= 1, got %d", id)
	}

	// Test Read
	var name, email, address string
	var taxID *string
	err = db.DB.QueryRow("SELECT name, email, address, tax_id FROM clients WHERE id = ?", id).
		Scan(&name, &email, &address, &taxID)
	if err != nil {
		t.Fatalf("Failed to query client: %v", err)
	}

	if name != "Test Client" {
		t.Errorf("Expected name 'Test Client', got '%s'", name)
	}
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", email)
	}

	// Test Update
	_, err = db.DB.Exec("UPDATE clients SET name = 'Updated Name' WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to update client: %v", err)
	}

	err = db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query updated client: %v", err)
	}
	if name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got '%s'", name)
	}

	// Test Delete
	_, err = db.DB.Exec("DELETE FROM clients WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to delete client: %v", err)
	}

	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM clients WHERE id = ?", id).Scan(&count)
	if count != 0 {
		t.Errorf("Expected client to be deleted, found %d", count)
	}
}

func TestClientListMultiple(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Insert multiple clients
	clients := []struct {
		name  string
		email string
	}{
		{"Client A", "a@example.com"},
		{"Client B", "b@example.com"},
		{"Client C", "c@example.com"},
	}

	for _, c := range clients {
		_, err := db.DB.Exec("INSERT INTO clients (name, email) VALUES (?, ?)", c.name, c.email)
		if err != nil {
			t.Fatalf("Failed to insert client %s: %v", c.name, err)
		}
	}

	// Query all clients from this test
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM clients WHERE name LIKE 'Client %'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count clients: %v", err)
	}

	if count < 3 {
		t.Errorf("Expected at least 3 clients, got %d", count)
	}
}

func TestClientRequiredFields(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Name and email are required
	_, err := db.DB.Exec("INSERT INTO clients (name, email) VALUES ('Required Test', 'required@test.com')")
	if err != nil {
		t.Fatalf("Failed to insert client with required fields: %v", err)
	}

	// Verify insertion
	var name, email string
	err = db.DB.QueryRow("SELECT name, email FROM clients WHERE name = 'Required Test'").Scan(&name, &email)
	if err != nil {
		t.Fatalf("Failed to query client: %v", err)
	}

	if name != "Required Test" {
		t.Errorf("Expected name 'Required Test', got '%s'", name)
	}
}
