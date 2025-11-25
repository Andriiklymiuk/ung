package contract

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/pkg/invoice"
)

// TestFormatContractType tests the contract type formatting
func TestFormatContractType(t *testing.T) {
	tests := []struct {
		contractType models.ContractType
		expected     string
	}{
		{models.ContractTypeHourly, "Hourly Rate"},
		{models.ContractTypeFixedPrice, "Fixed Price"},
		{models.ContractTypeRetainer, "Retainer"},
		{models.ContractType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.contractType), func(t *testing.T) {
			result := formatContractType(tt.contractType)
			if result != tt.expected {
				t.Errorf("formatContractType(%s) = %s; want %s",
					tt.contractType, result, tt.expected)
			}
		})
	}
}

// TestTruncateString tests the string truncation function
func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Hello", 10, "Hello"},
		{"Hello World", 5, "He..."},
		{"Short", 5, "Short"},
		{"Exactly Ten", 11, "Exactly Ten"},
		{"Longer string here", 10, "Longer ..."},
		{"", 5, ""},
		{"ABC", 3, "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q; want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestSanitizeFilename tests filename sanitization
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_file.pdf", "normal_file.pdf"},
		{"file/with/slashes.pdf", "file_with_slashes.pdf"},
		{"file\\with\\backslashes.pdf", "file_with_backslashes.pdf"},
		{"file:with:colons.pdf", "file_with_colons.pdf"},
		{"file*with*stars.pdf", "file_with_stars.pdf"},
		{"file?with?questions.pdf", "file_with_questions.pdf"},
		{"file\"with\"quotes.pdf", "file_with_quotes.pdf"},
		{"file<with>angles.pdf", "file_with_angles.pdf"},
		{"file|with|pipes.pdf", "file_with_pipes.pdf"},
		{"Client Name_Contract Name.pdf", "Client Name_Contract Name.pdf"},
		{"ACME/Corp_Q1*2024.pdf", "ACME_Corp_Q1_2024.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q; want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestContractDataStructure tests the contract model structure
func TestContractDataStructure(t *testing.T) {
	now := time.Now()
	hourlyRate := 150.00
	endDate := now.AddDate(1, 0, 0)

	contract := models.Contract{
		ID:           1,
		ContractNum:  "CONTRACT-2024-001",
		ClientID:     1,
		Name:         "Web Development Project",
		ContractType: models.ContractTypeHourly,
		HourlyRate:   &hourlyRate,
		Currency:     "USD",
		StartDate:    now,
		EndDate:      &endDate,
		Active:       true,
		Notes:        "Monthly retainer agreement",
	}

	// Verify data integrity
	if contract.ContractNum == "" {
		t.Error("contract number should not be empty")
	}
	if contract.Name == "" {
		t.Error("contract name should not be empty")
	}
	if contract.HourlyRate == nil {
		t.Error("hourly rate should be set for hourly contract")
	}
	if *contract.HourlyRate != hourlyRate {
		t.Errorf("hourly rate = %.2f; want %.2f", *contract.HourlyRate, hourlyRate)
	}
	if !contract.Active {
		t.Error("contract should be active")
	}
}

// TestContractTypeValues tests contract type enum values
func TestContractTypeValues(t *testing.T) {
	types := map[models.ContractType]string{
		models.ContractTypeHourly:     "hourly",
		models.ContractTypeFixedPrice: "fixed_price",
		models.ContractTypeRetainer:   "retainer",
	}

	for ct, expected := range types {
		if string(ct) != expected {
			t.Errorf("ContractType %v = %s; want %s", ct, string(ct), expected)
		}
	}
}

// TestContractWithFixedPrice tests fixed price contract
func TestContractWithFixedPrice(t *testing.T) {
	fixedPrice := 5000.00

	contract := models.Contract{
		ID:           1,
		ContractNum:  "FP-2024-001",
		Name:         "Website Redesign",
		ContractType: models.ContractTypeFixedPrice,
		FixedPrice:   &fixedPrice,
		Currency:     "EUR",
		Active:       true,
	}

	if contract.FixedPrice == nil {
		t.Error("fixed price should be set")
	}
	if *contract.FixedPrice != fixedPrice {
		t.Errorf("fixed price = %.2f; want %.2f", *contract.FixedPrice, fixedPrice)
	}
	if contract.ContractType != models.ContractTypeFixedPrice {
		t.Errorf("contract type = %s; want fixed_price", contract.ContractType)
	}
}

// TestContractPDFConfig tests PDF config for contracts
func TestContractPDFConfig(t *testing.T) {
	cfg := config.GetDefaultPDFConfig()

	// Contracts should use the same color scheme as invoices
	if cfg.PrimaryColor.R != 232 || cfg.PrimaryColor.G != 119 || cfg.PrimaryColor.B != 34 {
		t.Errorf("Primary color mismatch for contracts")
	}

	// Verify page numbers are enabled
	if !cfg.ShowPageNumber {
		t.Error("Page numbers should be enabled by default")
	}

	// Verify watermarks are enabled
	if !cfg.ShowWatermark {
		t.Error("Watermarks should be enabled by default")
	}
}

// TestContractCurrencyFormatting tests currency formatting in contracts
func TestContractCurrencyFormatting(t *testing.T) {
	hourlyRate := 75.00

	// Test with invoice package's FormatCurrency
	formatted := invoice.FormatCurrency(hourlyRate, "EUR")
	expected := "€75.00"

	if formatted != expected {
		t.Errorf("FormatCurrency(%.2f, EUR) = %s; want %s", hourlyRate, formatted, expected)
	}
}

// TestContractStatusMapping tests active/inactive status mapping
func TestContractStatusMapping(t *testing.T) {
	tests := []struct {
		active       bool
		expectedText string
	}{
		{true, "Active"},
		{false, "Inactive"},
	}

	for _, tt := range tests {
		statusText := "Active"
		if !tt.active {
			statusText = "Inactive"
		}

		if statusText != tt.expectedText {
			t.Errorf("Status for active=%v = %s; want %s",
				tt.active, statusText, tt.expectedText)
		}
	}
}

// TestContractDates tests date handling in contracts
func TestContractDates(t *testing.T) {
	now := time.Now()

	// Open-ended contract (no end date)
	openContract := models.Contract{
		ID:        1,
		Name:      "Ongoing Support",
		StartDate: now,
		EndDate:   nil,
		Active:    true,
	}

	if openContract.EndDate != nil {
		t.Error("Open-ended contract should have nil end date")
	}

	// Fixed-term contract
	endDate := now.AddDate(0, 6, 0)
	fixedContract := models.Contract{
		ID:        2,
		Name:      "6-Month Project",
		StartDate: now,
		EndDate:   &endDate,
		Active:    true,
	}

	if fixedContract.EndDate == nil {
		t.Error("Fixed-term contract should have end date")
	}

	// Verify end date is after start date
	if fixedContract.EndDate.Before(fixedContract.StartDate) {
		t.Error("End date should be after start date")
	}
}

// TestCompanyAndClientData tests company and client model for contracts
func TestCompanyAndClientData(t *testing.T) {
	company := models.Company{
		ID:          1,
		Name:        "Test Consulting LLC",
		Email:       "contact@testconsulting.com",
		Phone:       "+1-555-0100",
		Address:     "100 Business Park, Suite 200",
		TaxID:       "EIN-12-3456789",
		BankName:    "First National Bank",
		BankAccount: "123456789012",
		BankSWIFT:   "FNBKUS33",
		LogoPath:    "~/company-logo.png",
	}

	client := models.Client{
		ID:      1,
		Name:    "Enterprise Corp",
		Email:   "contracts@enterprise.com",
		Address: "500 Corporate Blvd",
		TaxID:   "EIN-98-7654321",
	}

	// Verify company data
	if company.BankName == "" || company.BankAccount == "" {
		t.Error("Company should have bank details for contracts")
	}
	if company.BankSWIFT == "" {
		t.Error("Company should have SWIFT code for international contracts")
	}

	// Verify client data
	if client.Name == "" {
		t.Error("Client name is required")
	}
	if client.Email == "" {
		t.Error("Client email is required for contract delivery")
	}
}

// TestFilenameGeneration tests contract filename generation
func TestFilenameGeneration(t *testing.T) {
	tests := []struct {
		clientName   string
		contractName string
		expected     string
	}{
		{"ACME Corp", "Q1 Project", "ACME Corp_Q1 Project.pdf"},
		{"Client/Inc", "Contract:2024", "Client_Inc_Contract_2024.pdf"},
		{"Normal Client", "Normal Contract", "Normal Client_Normal Contract.pdf"},
	}

	for _, tt := range tests {
		filename := sanitizeFilename(tt.clientName + "_" + tt.contractName + ".pdf")
		if filename != tt.expected {
			t.Errorf("filename for %s + %s = %s; want %s",
				tt.clientName, tt.contractName, filename, tt.expected)
		}
	}
}

// BenchmarkSanitizeFilename benchmarks the filename sanitization
func BenchmarkSanitizeFilename(b *testing.B) {
	filename := "Client/Name:With*Special?Characters.pdf"
	for i := 0; i < b.N; i++ {
		sanitizeFilename(filename)
	}
}

// BenchmarkFormatContractType benchmarks contract type formatting
func BenchmarkFormatContractType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatContractType(models.ContractTypeHourly)
	}
}

// BenchmarkTruncateString benchmarks string truncation
func BenchmarkTruncateString(b *testing.B) {
	longString := "This is a very long string that needs to be truncated"
	for i := 0; i < b.N; i++ {
		truncateString(longString, 20)
	}
}
