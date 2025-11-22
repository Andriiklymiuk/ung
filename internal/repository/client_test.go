package repository

import (
	"testing"

	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestClientRepository_Create(t *testing.T) {
	setupTestDB(t)
	repo := NewClientRepository()

	client := &models.Client{
		Name:    "Test Client",
		Email:   "client@example.com",
		Address: "456 Client St",
		TaxID:   "67890",
	}

	err := repo.Create(client)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.ID == 0 {
		t.Error("expected client ID to be set")
	}
}

func TestClientRepository_GetByID(t *testing.T) {
	setupTestDB(t)
	repo := NewClientRepository()

	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
	}
	repo.Create(client)

	retrieved, err := repo.GetByID(client.ID)
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}

	if retrieved.Name != client.Name {
		t.Errorf("expected name %s, got %s", client.Name, retrieved.Name)
	}
}

func TestClientRepository_List(t *testing.T) {
	setupTestDB(t)
	repo := NewClientRepository()

	clients := []*models.Client{
		{Name: "Client 1", Email: "c1@example.com"},
		{Name: "Client 2", Email: "c2@example.com"},
	}

	for _, c := range clients {
		repo.Create(c)
	}

	result, err := repo.List()
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 clients, got %d", len(result))
	}
}

func TestClientRepository_Update(t *testing.T) {
	setupTestDB(t)
	repo := NewClientRepository()

	client := &models.Client{
		Name:  "Original Name",
		Email: "client@example.com",
	}
	repo.Create(client)

	client.Name = "Updated Name"
	err := repo.Update(client)
	if err != nil {
		t.Fatalf("failed to update client: %v", err)
	}

	updated, _ := repo.GetByID(client.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", updated.Name)
	}
}

func TestClientRepository_Delete(t *testing.T) {
	setupTestDB(t)
	repo := NewClientRepository()

	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
	}
	repo.Create(client)

	err := repo.Delete(client.ID)
	if err != nil {
		t.Fatalf("failed to delete client: %v", err)
	}

	_, err = repo.GetByID(client.ID)
	if err == nil {
		t.Error("expected error when getting deleted client")
	}
}

func TestClientRepository_Count(t *testing.T) {
	setupTestDB(t)
	repo := NewClientRepository()

	count, err := repo.Count()
	if err != nil {
		t.Fatalf("failed to count clients: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	repo.Create(&models.Client{Name: "Client 1", Email: "c1@example.com"})

	count, err = repo.Count()
	if err != nil {
		t.Fatalf("failed to count clients: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}
