package invoice

import (
	"fmt"
	"path/filepath"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// GeneratePDF creates a PDF invoice
func GeneratePDF(invoice models.Invoice, company models.Company, client models.Client) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(190, 10, "INVOICE")
	pdf.Ln(15)

	// Invoice details
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 6, fmt.Sprintf("Invoice Number: %s", invoice.InvoiceNum))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Issue Date: %s", invoice.IssuedDate.Format("2006-01-02")))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Due Date: %s", invoice.DueDate.Format("2006-01-02")))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Status: %s", invoice.Status))
	pdf.Ln(12)

	// From (Company)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(95, 8, "From:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 6, company.Name)
	pdf.Ln(6)
	pdf.Cell(95, 6, company.Email)
	pdf.Ln(6)
	if company.Address != "" {
		pdf.Cell(95, 6, company.Address)
		pdf.Ln(6)
	}
	if company.TaxID != "" {
		pdf.Cell(95, 6, fmt.Sprintf("Tax ID: %s", company.TaxID))
		pdf.Ln(6)
	}
	pdf.Ln(8)

	// To (Client)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(95, 8, "Bill To:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 6, client.Name)
	pdf.Ln(6)
	pdf.Cell(95, 6, client.Email)
	pdf.Ln(6)
	if client.Address != "" {
		pdf.Cell(95, 6, client.Address)
		pdf.Ln(6)
	}
	if client.TaxID != "" {
		pdf.Cell(95, 6, fmt.Sprintf("Tax ID: %s", client.TaxID))
		pdf.Ln(6)
	}
	pdf.Ln(12)

	// Description
	if invoice.Description != "" {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "Description:")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(190, 6, invoice.Description, "", "", false)
		pdf.Ln(6)
	}

	// Amount
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(140, 10, "Total Amount:", "1", 0, "R", true, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%.2f %s", invoice.Amount, invoice.Currency), "1", 0, "R", true, 0, "")
	pdf.Ln(10)

	// Save PDF
	invoicesDir := db.GetInvoicesDir()
	filename := fmt.Sprintf("%s.pdf", invoice.InvoiceNum)
	pdfPath := filepath.Join(invoicesDir, filename)

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}
