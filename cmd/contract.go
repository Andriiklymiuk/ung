package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/internal/repository"
	"github.com/Andriiklymiuk/ung/pkg/contract"
	"github.com/Andriiklymiuk/ung/pkg/idgen"
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

var contractPDFCmd = &cobra.Command{
	Use:   "pdf [contract-id]",
	Short: "Generate PDF for a contract",
	Args:  cobra.ExactArgs(1),
	RunE:  runContractPDF,
}

var contractEmailCmd = &cobra.Command{
	Use:   "email [contract-id]",
	Short: "Export contract to email client",
	Args:  cobra.ExactArgs(1),
	RunE:  runContractEmail,
}

var contractDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a contract",
	Long: `Delete a contract permanently.

⚠️  This action cannot be undone!

Examples:
  ung contract delete 5      Delete contract #5`,
	Args: cobra.ExactArgs(1),
	RunE: runContractDelete,
}

var (
	contractDeleteYes bool
	contractClientID  int
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
	contractCmd.AddCommand(contractPDFCmd)
	contractCmd.AddCommand(contractEmailCmd)
	contractCmd.AddCommand(contractDeleteCmd)

	// Add flags (optional - if not provided, will use interactive mode)
	contractAddCmd.Flags().IntVar(&contractClientID, "client", 0, "Client ID")
	contractAddCmd.Flags().StringVar(&contractName, "name", "", "Contract name")
	contractAddCmd.Flags().StringVar(&contractType, "type", "", "Contract type (hourly, fixed_price, retainer)")
	contractAddCmd.Flags().Float64Var(&contractRate, "rate", 0, "Hourly rate (for hourly contracts)")
	contractAddCmd.Flags().Float64Var(&contractPrice, "price", 0, "Fixed price (for fixed_price contracts)")
	contractAddCmd.Flags().StringVar(&contractCurrency, "currency", "USD", "Currency")

	// Edit flags
	contractEditCmd.Flags().StringVar(&contractName, "name", "", "Contract name")
	contractEditCmd.Flags().Float64Var(&contractRate, "rate", 0, "Hourly rate")
	contractEditCmd.Flags().Float64Var(&contractPrice, "price", 0, "Fixed price")
	contractEditCmd.Flags().StringVar(&contractCurrency, "currency", "", "Currency")
	contractEditCmd.Flags().BoolVar(&contractActive, "active", true, "Contract active status")
	contractEditCmd.Flags().StringVar(&contractNotes, "notes", "", "Contract notes")

	// Delete flags
	contractDeleteCmd.Flags().BoolVarP(&contractDeleteYes, "yes", "y", false, "Skip confirmation prompt")
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

	// Get client name for contract number generation
	var clientName string
	if err := db.DB.QueryRow("SELECT name FROM clients WHERE id = ?", contractClientID).Scan(&clientName); err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Use current time as start date
	startDate := time.Now()

	// Generate human-readable contract number
	contractNum, err := idgen.GenerateContractNumber(db.GormDB, clientName, startDate)
	if err != nil {
		return fmt.Errorf("failed to generate contract number: %w", err)
	}

	// Insert contract
	query := `
		INSERT INTO contracts (contract_num, client_id, name, contract_type, hourly_rate, fixed_price, currency, start_date, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
	`

	var ratePtr *float64
	if contractRate > 0 {
		ratePtr = &contractRate
	}

	var pricePtr *float64
	if contractPrice > 0 {
		pricePtr = &contractPrice
	}

	result, err := db.DB.Exec(query, contractNum, contractClientID, contractName, ct, ratePtr, pricePtr, contractCurrency, startDate)
	if err != nil {
		return fmt.Errorf("failed to add contract: %w", err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("✓ Contract added successfully (ID: %d)\n", id)
	fmt.Printf("  Contract Number: %s\n", contractNum)
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
		SELECT c.id, c.contract_num, c.name, c.contract_type, c.hourly_rate, c.fixed_price, c.currency, c.active, cl.name
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
	fmt.Fprintln(w, "ID\tCONTRACT#\tNAME\tCLIENT\tTYPE\tRATE/PRICE\tACTIVE")

	for rows.Next() {
		var id int
		var contractNum, name, contractType, currency, clientName string
		var hourlyRate, fixedPrice *float64
		var active bool

		if err := rows.Scan(&id, &contractNum, &name, &contractType, &hourlyRate, &fixedPrice, &currency, &active, &clientName); err != nil {
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

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			id, contractNum, name, clientName, contractType, ratePrice, activeStr)
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
			ORDER BY c.active DESC, c.id DESC
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
			return fmt.Errorf("no contracts found")
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

	// Check if any flags were provided (non-interactive mode)
	hasFlags := cmd.Flags().Changed("name") || cmd.Flags().Changed("rate") ||
		cmd.Flags().Changed("price") || cmd.Flags().Changed("currency") ||
		cmd.Flags().Changed("active") || cmd.Flags().Changed("notes")

	if hasFlags {
		// Non-interactive mode - use flags to update
		updates := make(map[string]interface{})
		if cmd.Flags().Changed("name") {
			updates["name"] = contractName
		}
		if cmd.Flags().Changed("rate") {
			updates["hourly_rate"] = contractRate
		}
		if cmd.Flags().Changed("price") {
			updates["fixed_price"] = contractPrice
		}
		if cmd.Flags().Changed("currency") {
			updates["currency"] = contractCurrency
		}
		if cmd.Flags().Changed("active") {
			updates["active"] = contractActive
		}
		if cmd.Flags().Changed("notes") {
			updates["notes"] = contractNotes
		}

		if len(updates) == 0 {
			return fmt.Errorf("no fields to update")
		}

		query := "UPDATE contracts SET "
		argsList := []interface{}{}
		i := 0
		for key, val := range updates {
			if i > 0 {
				query += ", "
			}
			query += key + " = ?"
			argsList = append(argsList, val)
			i++
		}
		query += ", updated_at = ? WHERE id = ?"
		argsList = append(argsList, time.Now(), contractID)

		result, err := db.DB.Exec(query, argsList...)
		if err != nil {
			return fmt.Errorf("failed to update contract: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("contract with ID %d not found", contractID)
		}

		fmt.Printf("✓ Contract %d updated successfully\n", contractID)
		return nil
	}

	// Interactive mode - fetch current values and show form
	var currentName, currentCurrency string
	var currentRate, currentPrice *float64
	var currentActive bool

	err = db.DB.QueryRow(`
		SELECT name, hourly_rate, fixed_price, currency, active
		FROM contracts WHERE id = ?
	`, contractID).Scan(&currentName, &currentRate, &currentPrice, &currentCurrency, &currentActive)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	// Prepare form values
	editName := currentName
	editCurrency := currentCurrency
	editActive := currentActive
	editRateStr := ""
	editPriceStr := ""
	if currentRate != nil {
		editRateStr = fmt.Sprintf("%.2f", *currentRate)
	}
	if currentPrice != nil {
		editPriceStr = fmt.Sprintf("%.2f", *currentPrice)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Contract Name").
				Value(&editName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Hourly Rate").
				Description("Leave empty if not applicable").
				Value(&editRateStr),

			huh.NewInput().
				Title("Fixed Price").
				Description("Leave empty if not applicable").
				Value(&editPriceStr),

			huh.NewInput().
				Title("Currency").
				Value(&editCurrency),

			huh.NewConfirm().
				Title("Is this contract active?").
				Value(&editActive),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}

	// Build update query
	updates := make(map[string]interface{})

	if editName != currentName {
		updates["name"] = editName
	}
	if editCurrency != currentCurrency {
		updates["currency"] = editCurrency
	}
	if editActive != currentActive {
		updates["active"] = editActive
	}

	// Handle rate changes
	if editRateStr != "" {
		newRate, err := strconv.ParseFloat(editRateStr, 64)
		if err == nil && (currentRate == nil || newRate != *currentRate) {
			updates["hourly_rate"] = newRate
		}
	} else if currentRate != nil {
		updates["hourly_rate"] = nil
	}

	// Handle price changes
	if editPriceStr != "" {
		newPrice, err := strconv.ParseFloat(editPriceStr, 64)
		if err == nil && (currentPrice == nil || newPrice != *currentPrice) {
			updates["fixed_price"] = newPrice
		}
	} else if currentPrice != nil {
		updates["fixed_price"] = nil
	}

	if len(updates) == 0 {
		fmt.Println("No changes made")
		return nil
	}

	query := "UPDATE contracts SET "
	argsList := []interface{}{}
	i := 0
	for key, val := range updates {
		if i > 0 {
			query += ", "
		}
		query += key + " = ?"
		argsList = append(argsList, val)
		i++
	}
	query += ", updated_at = ? WHERE id = ?"
	argsList = append(argsList, time.Now(), contractID)

	_, err = db.DB.Exec(query, argsList...)
	if err != nil {
		return fmt.Errorf("failed to update contract: %w", err)
	}

	fmt.Printf("✓ Contract %d updated successfully\n", contractID)
	return nil
}

func runContractPDF(cmd *cobra.Command, args []string) error {
	contractID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid contract ID: %w", err)
	}

	contractRepo := repository.NewContractRepository()
	companyRepo := repository.NewCompanyRepository()

	// Get contract with client preloaded
	contractModel, err := contractRepo.GetByID(uint(contractID))
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	// Get company (assuming first company for now)
	companies, err := companyRepo.List()
	if err != nil || len(companies) == 0 {
		return fmt.Errorf("no company found: %w", err)
	}
	company := companies[0]

	// Generate PDF
	pdfPath, err := contract.GeneratePDF(*contractModel, company, contractModel.Client)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Update contract with PDF path
	contractModel.PDFPath = pdfPath
	if err := contractRepo.Update(contractModel); err != nil {
		return fmt.Errorf("failed to update contract: %w", err)
	}

	fmt.Printf("✓ PDF generated successfully: %s\n", pdfPath)
	return nil
}

func runContractEmail(cmd *cobra.Command, args []string) error {
	contractID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid contract ID: %w", err)
	}

	contractRepo := repository.NewContractRepository()
	companyRepo := repository.NewCompanyRepository()

	// Get contract
	contractModel, err := contractRepo.GetByID(uint(contractID))
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	// Get company
	companies, err := companyRepo.List()
	if err != nil || len(companies) == 0 {
		return fmt.Errorf("no company found: %w", err)
	}
	company := companies[0]

	// Ensure PDF is generated
	var pdfPath string
	if contractModel.PDFPath == "" {
		fmt.Println("📄 Generating PDF first...")
		pdfPath, err = contract.GeneratePDF(*contractModel, company, contractModel.Client)
		if err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
		contractModel.PDFPath = pdfPath
		contractRepo.Update(contractModel)
	} else {
		pdfPath = contractModel.PDFPath
	}

	// Prepare email details
	subject := fmt.Sprintf("Contract: %s - %s", company.Name, contractModel.Name)
	body := fmt.Sprintf("Hi,\n\nPlease find attached the contract for %s.\n\nBest regards,\n%s",
		contractModel.Name, company.Name)

	// Prompt for email client selection
	var emailClient string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select email client").
				Options(
					huh.NewOption("Apple Mail", "apple"),
					huh.NewOption("Outlook", "outlook"),
					huh.NewOption("Gmail (Browser)", "gmail"),
				).
				Value(&emailClient),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("email export cancelled: %w", err)
	}

	// Export to selected email client
	switch emailClient {
	case "apple":
		return exportContractToAppleMail(subject, body, pdfPath)
	case "outlook":
		return exportContractToOutlook(subject, body, pdfPath)
	case "gmail":
		return exportContractToGmail(subject, body, pdfPath)
	default:
		return fmt.Errorf("unknown email client: %s", emailClient)
	}
}

func exportContractToAppleMail(subject, body, attachmentPath string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("Apple Mail is only available on macOS")
	}

	script := fmt.Sprintf(`
tell application "Mail"
	activate
	set newMessage to make new outgoing message with properties {subject:"%s", content:"%s", visible:true}
	tell newMessage
		make new attachment with properties {file name:POSIX file "%s"} at after the last paragraph
	end tell
end tell
`, escapeAppleScript(subject), escapeAppleScript(body), attachmentPath)

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open Apple Mail: %w", err)
	}

	fmt.Println("✓ Email draft created in Apple Mail with attachment")
	return nil
}

func exportContractToOutlook(subject, body, attachmentPath string) error {
	mailtoURL := fmt.Sprintf("mailto:?subject=%s&body=%s",
		url.QueryEscape(subject),
		url.QueryEscape(body))

	if err := openURL(mailtoURL); err != nil {
		return fmt.Errorf("failed to open Outlook: %w", err)
	}

	fmt.Println("✓ Email draft created in Outlook")
	fmt.Printf("📎 Please manually attach the PDF: %s\n", attachmentPath)
	return nil
}

func exportContractToGmail(subject, body, attachmentPath string) error {
	gmailURL := fmt.Sprintf("https://mail.google.com/mail/?view=cm&fs=1&su=%s&body=%s",
		url.QueryEscape(subject),
		url.QueryEscape(body))

	if err := openURL(gmailURL); err != nil {
		return fmt.Errorf("failed to open Gmail: %w", err)
	}

	fmt.Println("✓ Gmail compose opened in browser")
	fmt.Printf("📎 Please manually attach the PDF: %s\n", attachmentPath)
	return nil
}

// runContractDelete deletes a contract
func runContractDelete(cmd *cobra.Command, args []string) error {
	contractID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid contract ID: %s", args[0])
	}

	// Get contract details first
	var contractNum, name, clientName string
	err = db.DB.QueryRow(`
		SELECT c.contract_num, c.name, cl.name
		FROM contracts c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.id = ?
	`, contractID).Scan(&contractNum, &name, &clientName)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	// Show contract info
	fmt.Printf("\n📄 Contract: %s\n", contractNum)
	fmt.Printf("   Name: %s\n", name)
	fmt.Printf("   Client: %s\n\n", clientName)

	// Confirm deletion unless --yes flag is set
	if !contractDeleteYes {
		// Confirmation
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Are you sure you want to delete this contract?").
					Description("This action cannot be undone!").
					Affirmative("Yes, delete it").
					Negative("Cancel").
					Value(&confirm),
			),
		)

		if err := form.Run(); err != nil || !confirm {
			fmt.Println("❌ Deletion cancelled")
			return nil
		}
	}

	// Check for related time entries
	var timeEntryCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE contract_id = ?", contractID).Scan(&timeEntryCount)
	if timeEntryCount > 0 {
		fmt.Printf("⚠️  Warning: This contract has %d associated time entries. They will be orphaned.\n", timeEntryCount)
		if !contractDeleteYes {
			var confirmOrphan bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Continue with deletion?").
						Affirmative("Yes, continue").
						Negative("Cancel").
						Value(&confirmOrphan),
				),
			)
			if err := form.Run(); err != nil || !confirmOrphan {
				fmt.Println("❌ Deletion cancelled")
				return nil
			}
		}
	}

	// Delete contract from database
	_, err = db.DB.Exec("DELETE FROM contracts WHERE id = ?", contractID)
	if err != nil {
		return fmt.Errorf("failed to delete contract: %w", err)
	}

	fmt.Printf("✓ Contract %s deleted successfully!\n", contractNum)
	return nil
}
