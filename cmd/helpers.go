package cmd

import (
	"fmt"

	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/internal/repository"
	"github.com/charmbracelet/huh"
)

// ensureCompanyExists checks if a company exists, and prompts to create one if not
func ensureCompanyExists() error {
	repo := repository.NewCompanyRepository()

	count, err := repo.Count()
	if err != nil {
		return fmt.Errorf("failed to check companies: %w", err)
	}

	if count > 0 {
		return nil // Company exists, we're good
	}

	// No company - prompt to create one
	fmt.Println("⚠️  No company information found. Let's add your business details first!")
	fmt.Println()

	var shouldAdd bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to add your company information now?").
				Value(&shouldAdd),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled")
	}

	if !shouldAdd {
		return fmt.Errorf("company information is required. Run 'ung company add' to add your business details")
	}

	// Interactive company creation
	var name, email, address, taxID string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Company Name").
				Placeholder("e.g., John Doe Studio").
				Value(&name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("company name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Email").
				Placeholder("e.g., hello@example.com").
				Value(&email).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("email is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Address (optional)").
				Placeholder("e.g., 123 Main St, City, Country").
				Value(&address),

			huh.NewInput().
				Title("Tax ID (optional)").
				Placeholder("e.g., US123456789").
				Value(&taxID),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("company creation cancelled: %w", err)
	}

	// Insert company
	company := &models.Company{
		Name:    name,
		Email:   email,
		Address: address,
		TaxID:   taxID,
	}

	if err := repo.Create(company); err != nil {
		return fmt.Errorf("failed to add company: %w", err)
	}

	fmt.Printf("\n✓ Company added successfully (ID: %d)\n\n", company.ID)

	return nil
}

// ensureClientExists checks if a client exists, and prompts to create one if not
func ensureClientExists() error {
	repo := repository.NewClientRepository()

	count, err := repo.Count()
	if err != nil {
		return fmt.Errorf("failed to check clients: %w", err)
	}

	if count > 0 {
		return nil // Clients exist, we're good
	}

	// No clients - prompt to create one
	fmt.Println("⚠️  No clients found. Let's add your first client!")
	fmt.Println()

	var shouldAdd bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to add a client now?").
				Value(&shouldAdd),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled")
	}

	if !shouldAdd {
		return fmt.Errorf("at least one client is required. Run 'ung client add' to add a client")
	}

	// Interactive client creation
	var name, email, address, taxID string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Client Name").
				Placeholder("e.g., Acme Corporation").
				Value(&name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("client name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Email").
				Placeholder("e.g., contact@acme.com").
				Value(&email).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("email is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Address (optional)").
				Placeholder("e.g., 456 Business Ave, City").
				Value(&address),

			huh.NewInput().
				Title("Tax ID (optional)").
				Placeholder("e.g., EU987654321").
				Value(&taxID),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("client creation cancelled: %w", err)
	}

	// Insert client
	client := &models.Client{
		Name:    name,
		Email:   email,
		Address: address,
		TaxID:   taxID,
	}

	if err := repo.Create(client); err != nil {
		return fmt.Errorf("failed to add client: %w", err)
	}

	fmt.Printf("\n✓ Client added successfully (ID: %d)\n\n", client.ID)

	return nil
}

// ensureContractExists checks if a contract exists, and prompts to create one if not
func ensureContractExists() error {
	repo := repository.NewContractRepository()

	count, err := repo.CountActive()
	if err != nil {
		return fmt.Errorf("failed to check contracts: %w", err)
	}

	if count > 0 {
		return nil // Active contracts exist
	}

	// No contracts - prompt to create one
	fmt.Println("⚠️  No active contracts found. You need a contract to track time!")
	fmt.Println()

	var shouldAdd bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to create a contract now?").
				Affirmative("Yes, let's do it!").
				Negative("No, I'll do it later").
				Value(&shouldAdd),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled")
	}

	if !shouldAdd {
		return fmt.Errorf("at least one contract is required. Run 'ung contract add' to create a contract")
	}

	fmt.Println("\n📋 Let's create your first contract!")

	// Redirect to contract add - but we need to ensure prerequisites first
	if err := ensureCompanyExists(); err != nil {
		return err
	}

	if err := ensureClientExists(); err != nil {
		return err
	}

	// Now we can show the contract creation - we'll just return and tell them to use the command
	// since we're already in a complex flow
	fmt.Println("✓ Prerequisites ready! Now let's create your contract...")
	fmt.Println()

	return fmt.Errorf("please run 'ung contract add' to create your first contract, then try 'ung track log' again")
}
