package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var contractCmd = &cobra.Command{
	Use:   "contract",
	Short: "Manage contracts",
	Long:  "Add, list, and edit client contracts",
}

var contractAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new contract",
	RunE:  runContractAdd,
}

var contractListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all contracts",
	RunE:    runContractList,
}

var contractEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing contract",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runContractEdit,
}

var (
	contractClientID int
	contractName     string
	contractType     string
	contractRate     float64
	contractPrice    float64
	contractCurrency string
	contractActive   bool
	contractNotes    string
)

func init() {
	contractCmd.AddCommand(contractAddCmd)
	contractCmd.AddCommand(contractListCmd)
	contractCmd.AddCommand(contractEditCmd)

	// Add flags (optional - if not provided, will use interactive mode)
	contractAddCmd.Flags().IntVar(&contractClientID, "client", 0, "Client ID")
	contractAddCmd.Flags().StringVar(&contractName, "name", "", "Contract name")
	contractAddCmd.Flags().StringVar(&contractType, "type", "", "Contract type (hourly, fixed_price, retainer)")
	contractAddCmd.Flags().Float64Var(&contractRate, "rate", 0, "Hourly rate (for hourly contracts)")
	contractAddCmd.Flags().Float64Var(&contractPrice, "price", 0, "Fixed price (for fixed_price contracts)")
	contractAddCmd.Flags().StringVar(&contractCurrency, "currency", "USD", "Currency")
}

func getClients() ([]models.Client, error) {
	rows, err := db.DB.Query("SELECT id, name, email FROM clients ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.Name, &c.Email); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}

func runContractAdd(cmd *cobra.Command, args []string) error {
	// Ensure prerequisites exist
	fmt.Println("🚀 Let's create a contract!")

	if err := ensureCompanyExists(); err != nil {
		return err
	}

	if err := ensureClientExists(); err != nil {
		return err
	}

	// Get all clients for selection
	clients, err := getClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %w", err)
	}

	if len(clients) == 0 {
		return fmt.Errorf("no clients found. This shouldn't happen after ensureClientExists")
	}

	// Use interactive mode if required fields not provided
	if contractClientID == 0 || contractName == "" || contractType == "" {
		// Build client options
		clientOptions := make([]huh.Option[int], len(clients))
		for i, c := range clients {
			clientOptions[i] = huh.NewOption(fmt.Sprintf("%s (%s)", c.Name, c.Email), int(c.ID))
		}

		var selectedClientID int
		var selectedContractType string
		var selectedName string
		var selectedRateStr string
		var selectedPriceStr string
		var selectedCurrency string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Client").
					Options(clientOptions...).
					Value(&selectedClientID),

				huh.NewInput().
					Title("Contract Name").
					Placeholder("e.g., Website Development Q1 2025").
					Value(&selectedName).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("contract name is required")
						}
						return nil
					}),

				huh.NewSelect[string]().
					Title("Contract Type").
					Options(
						huh.NewOption("Hourly Rate", "hourly"),
						huh.NewOption("Fixed Price", "fixed_price"),
						huh.NewOption("Retainer", "retainer"),
					).
					Value(&selectedContractType),
			),

			huh.NewGroup(
				huh.NewInput().
					Title("Hourly Rate (if applicable)").
					Placeholder("e.g., 75.00").
					Value(&selectedRateStr).
					Validate(func(s string) error {
						if selectedContractType == "hourly" && s == "" {
							return fmt.Errorf("hourly rate is required for hourly contracts")
						}
						return nil
					}),

				huh.NewInput().
					Title("Fixed Price (if applicable)").
					Placeholder("e.g., 5000.00").
					Value(&selectedPriceStr),

				huh.NewInput().
					Title("Currency").
					Placeholder("USD").
					Value(&selectedCurrency).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("currency is required")
						}
						return nil
					}),
			).WithHideFunc(func() bool {
				return selectedContractType == ""
			}),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form cancelled: %w", err)
		}

		// Convert string inputs to float64
		if selectedRateStr != "" {
			rate, err := strconv.ParseFloat(selectedRateStr, 64)
			if err != nil {
				return fmt.Errorf("invalid hourly rate: %w", err)
			}
			contractRate = rate
		}

		if selectedPriceStr != "" {
			price, err := strconv.ParseFloat(selectedPriceStr, 64)
			if err != nil {
				return fmt.Errorf("invalid fixed price: %w", err)
			}
			contractPrice = price
		}

		contractClientID = selectedClientID
		contractName = selectedName
		contractType = selectedContractType
		if selectedCurrency != "" {
			contractCurrency = selectedCurrency
		}
	}

	// Validate contract type
	var ct models.ContractType
	switch contractType {
	case "hourly":
		ct = models.ContractTypeHourly
		if contractRate == 0 {
			return fmt.Errorf("hourly rate is required for hourly contracts")
		}
	case "fixed_price":
		ct = models.ContractTypeFixedPrice
		if contractPrice == 0 {
			return fmt.Errorf("fixed price is required for fixed-price contracts")
		}
	case "retainer":
		ct = models.ContractTypeRetainer
	default:
		return fmt.Errorf("invalid contract type: %s (use hourly, fixed_price, or retainer)", contractType)
	}

	// Insert contract
	query := `
		INSERT INTO contracts (client_id, name, contract_type, hourly_rate, fixed_price, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1)
	`

	var ratePtr *float64
	if contractRate > 0 {
		ratePtr = &contractRate
	}

	var pricePtr *float64
	if contractPrice > 0 {
		pricePtr = &contractPrice
	}

	result, err := db.DB.Exec(query, contractClientID, contractName, ct, ratePtr, pricePtr, contractCurrency, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add contract: %w", err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("✓ Contract added successfully (ID: %d)\n", id)
	fmt.Printf("  Client ID: %d\n", contractClientID)
	fmt.Printf("  Name: %s\n", contractName)
	fmt.Printf("  Type: %s\n", ct)
	if ratePtr != nil {
		fmt.Printf("  Hourly Rate: %.2f %s\n", *ratePtr, contractCurrency)
	}
	if pricePtr != nil {
		fmt.Printf("  Fixed Price: %.2f %s\n", *pricePtr, contractCurrency)
	}

	return nil
}

func runContractList(cmd *cobra.Command, args []string) error {
	query := `
		SELECT c.id, c.name, c.contract_type, c.hourly_rate, c.fixed_price, c.currency, c.active, cl.name
		FROM contracts c
		JOIN clients cl ON c.client_id = cl.id
		ORDER BY c.active DESC, c.id DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query contracts: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCLIENT\tTYPE\tRATE/PRICE\tACTIVE")

	for rows.Next() {
		var id int
		var name, contractType, currency, clientName string
		var hourlyRate, fixedPrice *float64
		var active bool

		if err := rows.Scan(&id, &name, &contractType, &hourlyRate, &fixedPrice, &currency, &active, &clientName); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		ratePrice := "-"
		if hourlyRate != nil {
			ratePrice = fmt.Sprintf("%.2f %s/hr", *hourlyRate, currency)
		} else if fixedPrice != nil {
			ratePrice = fmt.Sprintf("%.2f %s", *fixedPrice, currency)
		}

		activeStr := "✓"
		if !active {
			activeStr = "✗"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			id, name, clientName, contractType, ratePrice, activeStr)
	}

	w.Flush()
	return nil
}

func runContractEdit(cmd *cobra.Command, args []string) error {
	// Get contract ID
	var contractID int
	var err error

	if len(args) == 0 {
		// Interactive mode - show list and let user select
		rows, err := db.DB.Query(`
			SELECT c.id, c.name, cl.name
			FROM contracts c
			JOIN clients cl ON c.client_id = cl.id
			WHERE c.active = 1
			ORDER BY c.id DESC
		`)
		if err != nil {
			return fmt.Errorf("failed to query contracts: %w", err)
		}
		defer rows.Close()

		type contractOption struct {
			id         int
			name       string
			clientName string
		}

		var contracts []contractOption
		for rows.Next() {
			var c contractOption
			if err := rows.Scan(&c.id, &c.name, &c.clientName); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			contracts = append(contracts, c)
		}

		if len(contracts) == 0 {
			return fmt.Errorf("no active contracts found")
		}

		// Build options
		options := make([]huh.Option[int], len(contracts))
		for i, c := range contracts {
			options[i] = huh.NewOption(fmt.Sprintf("%s (%s)", c.name, c.clientName), c.id)
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Contract to Edit").
					Options(options...).
					Value(&contractID),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form cancelled: %w", err)
		}
	} else {
		contractID, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid contract ID: %w", err)
		}
	}

	// For now, just allow toggling active status
	var active bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Is this contract still active?").
				Value(&active),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}

	_, err = db.DB.Exec("UPDATE contracts SET active = ?, updated_at = ? WHERE id = ?", active, time.Now(), contractID)
	if err != nil {
		return fmt.Errorf("failed to update contract: %w", err)
	}

	fmt.Printf("✓ Contract %d updated successfully\n", contractID)
	return nil
}
