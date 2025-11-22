package repository

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestContractRepository_Create(t *testing.T) {
	setupTestDB(t)
	clientRepo := NewClientRepository()
	contractRepo := NewContractRepository()

	// Create client first
	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
	}
	clientRepo.Create(client)

	hourlyRate := 100.0
	contract := &models.Contract{
		ClientID:     client.ID,
		Name:         "Test Contract",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       true,
	}

	err := contractRepo.Create(contract)
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}

	if contract.ID == 0 {
		t.Error("expected contract ID to be set")
	}
}

func TestContractRepository_GetByID(t *testing.T) {
	setupTestDB(t)
	clientRepo := NewClientRepository()
	contractRepo := NewContractRepository()

	client := &models.Client{Name: "Test Client", Email: "client@example.com"}
	clientRepo.Create(client)

	hourlyRate := 100.0
	contract := &models.Contract{
		ClientID:     client.ID,
		Name:         "Test Contract",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       true,
	}
	contractRepo.Create(contract)

	retrieved, err := contractRepo.GetByID(contract.ID)
	if err != nil {
		t.Fatalf("failed to get contract: %v", err)
	}

	if retrieved.Name != contract.Name {
		t.Errorf("expected name %s, got %s", contract.Name, retrieved.Name)
	}

	if retrieved.Client.ID != client.ID {
		t.Error("expected client to be preloaded")
	}
}

func TestContractRepository_ListActive(t *testing.T) {
	setupTestDB(t)
	clientRepo := NewClientRepository()
	contractRepo := NewContractRepository()

	client := &models.Client{Name: "Test Client", Email: "client@example.com"}
	clientRepo.Create(client)

	hourlyRate := 100.0
	// Create active contract
	activeContract := &models.Contract{
		ClientID:     client.ID,
		Name:         "Active Contract",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       true,
	}
	contractRepo.Create(activeContract)

	// Create inactive contract
	inactiveContract := &models.Contract{
		ClientID:     client.ID,
		Name:         "Inactive Contract",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       false,
	}
	contractRepo.Create(inactiveContract)

	result, err := contractRepo.ListActive()
	if err != nil {
		t.Fatalf("failed to list active contracts: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 active contract, got %d", len(result))
	}

	if !result[0].Active {
		t.Error("expected contract to be active")
	}
}

func TestContractRepository_CountActive(t *testing.T) {
	setupTestDB(t)
	clientRepo := NewClientRepository()
	contractRepo := NewContractRepository()

	client := &models.Client{Name: "Test Client", Email: "client@example.com"}
	clientRepo.Create(client)

	hourlyRate := 100.0
	contractRepo.Create(&models.Contract{
		ClientID:     client.ID,
		Name:         "Active 1",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       true,
	})

	contractRepo.Create(&models.Contract{
		ClientID:     client.ID,
		Name:         "Inactive",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    time.Now(),
		Active:       false,
	})

	count, err := contractRepo.CountActive()
	if err != nil {
		t.Fatalf("failed to count active contracts: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 active contract, got %d", count)
	}
}
