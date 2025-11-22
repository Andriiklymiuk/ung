package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// GeneratePDF creates a professional PDF for a contract
func GeneratePDF(contract models.Contract, company models.Company, client models.Client) (string, error) {
	cfg, _ := config.Load()

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set fonts
	pdf.SetFont("Helvetica", "B", 24)

	// Title
	pdf.CellFormat(190, 10, cfg.Invoice.InvoiceLabel+" / CONTRACT", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Contract metadata (right-aligned)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(120)
	pdf.CellFormat(70, 5, fmt.Sprintf("Contract: %s", contract.Name), "", 1, "L", false, 0, "")
	pdf.SetX(120)
	pdf.CellFormat(70, 5, fmt.Sprintf("Type: %s", contract.ContractType), "", 1, "L", false, 0, "")
	pdf.SetX(120)
	pdf.CellFormat(70, 5, fmt.Sprintf("Start Date: %s", contract.StartDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
	if contract.EndDate != nil {
		pdf.SetX(120)
		pdf.CellFormat(70, 5, fmt.Sprintf("End Date: %s", contract.EndDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
	}
	pdf.SetX(120)
	status := "Active"
	if !contract.Active {
		status = "Inactive"
	}
	pdf.CellFormat(70, 5, fmt.Sprintf("Status: %s", status), "", 1, "L", false, 0, "")

	pdf.Ln(5)

	// Two-column layout for From/Bill To
	currentY := pdf.GetY()

	// Left column - From (Company)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetY(currentY)
	pdf.CellFormat(90, 6, cfg.Invoice.FromLabel, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(90, 5, company.Name, "", 1, "L", false, 0, "")
	if company.Email != "" {
		pdf.CellFormat(90, 5, company.Email, "", 1, "L", false, 0, "")
	}
	if company.Phone != "" {
		pdf.CellFormat(90, 5, company.Phone, "", 1, "L", false, 0, "")
	}
	if company.Address != "" {
		pdf.CellFormat(90, 5, company.Address, "", 1, "L", false, 0, "")
	}
	if company.RegistrationAddress != "" && company.RegistrationAddress != company.Address {
		pdf.CellFormat(90, 5, fmt.Sprintf("Reg: %s", company.RegistrationAddress), "", 1, "L", false, 0, "")
	}
	if company.TaxID != "" {
		pdf.CellFormat(90, 5, fmt.Sprintf("Tax ID: %s", company.TaxID), "", 1, "L", false, 0, "")
	}

	// Bank details
	if company.BankName != "" || company.BankAccount != "" {
		pdf.Ln(2)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(90, 5, "Bank Details:", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		if company.BankName != "" {
			pdf.CellFormat(90, 5, company.BankName, "", 1, "L", false, 0, "")
		}
		if company.BankAccount != "" {
			pdf.CellFormat(90, 5, fmt.Sprintf("Account: %s", company.BankAccount), "", 1, "L", false, 0, "")
		}
		if company.BankSWIFT != "" {
			pdf.CellFormat(90, 5, fmt.Sprintf("SWIFT: %s", company.BankSWIFT), "", 1, "L", false, 0, "")
		}
	}

	leftColumnEndY := pdf.GetY()

	// Right column - Bill To (Client)
	pdf.SetXY(105, currentY)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(85, 6, cfg.Invoice.BillToLabel, "", 1, "L", false, 0, "")

	pdf.SetX(105)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(85, 5, client.Name, "", 1, "L", false, 0, "")
	pdf.SetX(105)
	if client.Email != "" {
		pdf.CellFormat(85, 5, client.Email, "", 1, "L", false, 0, "")
		pdf.SetX(105)
	}
	if client.Address != "" {
		pdf.CellFormat(85, 5, client.Address, "", 1, "L", false, 0, "")
		pdf.SetX(105)
	}
	if client.TaxID != "" {
		pdf.CellFormat(85, 5, fmt.Sprintf("Tax ID: %s", client.TaxID), "", 1, "L", false, 0, "")
	}

	rightColumnEndY := pdf.GetY()

	// Move past both columns
	maxY := leftColumnEndY
	if rightColumnEndY > maxY {
		maxY = rightColumnEndY
	}
	pdf.SetY(maxY + 10)

	// Contract Details Section
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(190, 8, "Contract Details", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Contract terms table
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(60, 7, "Term", "1", 0, "L", true, 0, "")
	pdf.CellFormat(130, 7, "Value", "1", 1, "L", true, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetFillColor(255, 255, 255)

	// Contract Type
	pdf.CellFormat(60, 6, "Contract Type", "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, string(contract.ContractType), "1", 1, "L", false, 0, "")

	// Rate/Price
	if contract.HourlyRate != nil {
		pdf.CellFormat(60, 6, "Hourly Rate", "1", 0, "L", false, 0, "")
		pdf.CellFormat(130, 6, fmt.Sprintf("%.2f %s/hour", *contract.HourlyRate, contract.Currency), "1", 1, "L", false, 0, "")
	}
	if contract.FixedPrice != nil {
		pdf.CellFormat(60, 6, "Fixed Price", "1", 0, "L", false, 0, "")
		pdf.CellFormat(130, 6, fmt.Sprintf("%.2f %s", *contract.FixedPrice, contract.Currency), "1", 1, "L", false, 0, "")
	}

	// Dates
	pdf.CellFormat(60, 6, "Start Date", "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, contract.StartDate.Format("January 2, 2006"), "1", 1, "L", false, 0, "")

	if contract.EndDate != nil {
		pdf.CellFormat(60, 6, "End Date", "1", 0, "L", false, 0, "")
		pdf.CellFormat(130, 6, contract.EndDate.Format("January 2, 2006"), "1", 1, "L", false, 0, "")
	}

	// Status
	pdf.CellFormat(60, 6, "Status", "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, status, "1", 1, "L", false, 0, "")

	pdf.Ln(5)

	// Notes section
	if contract.Notes != "" {
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(190, 8, cfg.Invoice.NotesLabel, "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(190, 5, contract.Notes, "", "L", false)
		pdf.Ln(5)
	}

	// Terms & Conditions
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(190, 8, cfg.Invoice.TermsLabel, "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(190, 4, cfg.Invoice.Terms, "", "L", false)

	// Save PDF
	contractsDir := getContractsDir()
	filename := fmt.Sprintf("%s_%s.pdf", client.Name, contract.Name)
	// Sanitize filename
	filename = sanitizeFilename(filename)
	pdfPath := filepath.Join(contractsDir, filename)

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}

// getContractsDir returns the contracts directory path
func getContractsDir() string {
	invoicesDir := db.GetInvoicesDir()
	contractsDir := filepath.Join(filepath.Dir(invoicesDir), "contracts")

	// Ensure directory exists
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		// Fallback to invoices dir if we can't create contracts dir
		return invoicesDir
	}

	return contractsDir
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	result := filename
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
