package invoice

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// GeneratePDF creates a professional PDF invoice matching Zoho invoice style
func GeneratePDF(invoice models.Invoice, company models.Company, client models.Client, lineItems []models.InvoiceLineItem) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Load configuration for labels
	cfg, _ := config.Load()

	// Set margins
	leftMargin := 15.0
	rightMargin := 15.0
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - rightMargin

	// Header - Company name (large, left) and INVOICE title (right)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(leftMargin, 15)
	pdf.Cell(contentWidth/2, 10, company.Name)

	// INVOICE title on right
	pdf.SetFont("Arial", "B", 28)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetXY(pageWidth-rightMargin-60, 15)
	pdf.Cell(60, 10, cfg.Invoice.InvoiceLabel)

	// Company details below name
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetXY(leftMargin, 27)

	// Registration number and Tax ID
	if company.TaxID != "" {
		pdf.Cell(contentWidth/2, 4, fmt.Sprintf("Registration number: %s, Tax ID:", company.TaxID))
		pdf.Ln(4)
		pdf.SetX(leftMargin)
		pdf.Cell(contentWidth/2, 4, company.TaxID)
		pdf.Ln(4)
	}

	// Bank Details
	if company.BankAccount != "" {
		pdf.SetX(leftMargin)
		pdf.Cell(contentWidth/2, 4, "Bank Details: "+company.BankAccount)
		pdf.Ln(4)
	}

	// Address
	if company.Address != "" {
		pdf.SetX(leftMargin)
		pdf.MultiCell(contentWidth/2-10, 4, company.Address, "", "L", false)
	}

	// Bill To section (after company info, on left)
	billToY := pdf.GetY() + 8
	pdf.SetXY(leftMargin, billToY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(40, 5, cfg.Invoice.BillToLabel)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(leftMargin, billToY+7)
	pdf.Cell(contentWidth/2, 5, client.Name)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(60, 60, 60)
	currentY := billToY + 13

	// Client registration/VAT info
	if client.TaxID != "" {
		pdf.SetXY(leftMargin, currentY)
		pdf.Cell(contentWidth/2, 4, fmt.Sprintf("Registration number: %s, VAT ID:", client.TaxID[:min(len(client.TaxID), 11)]))
		currentY += 4
		pdf.SetXY(leftMargin, currentY)
		pdf.Cell(contentWidth/2, 4, client.TaxID)
		currentY += 4
	}

	// Client address
	if client.Address != "" {
		pdf.SetXY(leftMargin, currentY)
		pdf.MultiCell(contentWidth/2, 4, client.Address, "", "L", false)
	}

	// Invoice metadata on the right side (positioned next to Bill To)
	metaStartY := billToY
	metaLabelX := pageWidth - rightMargin - 80
	metaValueX := pageWidth - rightMargin - 40

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(80, 80, 80)

	// Invoice#
	pdf.SetXY(metaLabelX, metaStartY)
	pdf.Cell(40, 5, "Invoice#")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(metaValueX, metaStartY)
	pdf.Cell(40, 5, invoice.InvoiceNum)

	// Invoice Date
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(metaLabelX, metaStartY+6)
	pdf.Cell(40, 5, "Invoice Date")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(metaValueX, metaStartY+6)
	pdf.Cell(40, 5, invoice.IssuedDate.Format("02 Jan 2006"))

	// Due Date
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(metaLabelX, metaStartY+12)
	pdf.Cell(40, 5, "Due Date")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(metaValueX, metaStartY+12)
	pdf.Cell(40, 5, invoice.DueDate.Format("02 Jan 2006"))

	// Line Items Table
	tableY := pdf.GetY() + 15
	pdf.SetXY(leftMargin, tableY)

	// Ensure line items have proper names
	for i := range lineItems {
		if lineItems[i].ItemName == "" {
			lineItems[i].ItemName = fmt.Sprintf("Software services in %s", invoice.IssuedDate.Format("January"))
		}
	}

	drawLineItemsTable(pdf, lineItems, invoice.Currency, cfg.Invoice, leftMargin, contentWidth)

	// Notes section
	notesY := pdf.GetY() + 15
	pdf.SetXY(leftMargin, notesY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(200, 100, 50) // Orange-ish color like Zoho
	pdf.Cell(40, 5, cfg.Invoice.NotesLabel)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetXY(leftMargin, notesY+6)
	pdf.Cell(contentWidth, 4, "It was great doing business with you.")

	// Terms & Conditions
	termsY := notesY + 25
	pdf.SetXY(leftMargin, termsY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(200, 100, 50) // Orange-ish color
	pdf.Cell(40, 5, cfg.Invoice.TermsLabel)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetXY(leftMargin, termsY+6)
	pdf.Cell(contentWidth, 4, cfg.Invoice.Terms)

	// Save PDF to configured invoices directory
	invoicesDir := config.GetInvoicesDir()
	if err := os.MkdirAll(invoicesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create invoices directory: %w", err)
	}

	filename := fmt.Sprintf("%s.pdf", invoice.InvoiceNum)
	pdfPath := filepath.Join(invoicesDir, filename)

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}

// drawLineItemsTable creates a table matching Zoho invoice style
func drawLineItemsTable(pdf *gofpdf.Fpdf, items []models.InvoiceLineItem, currency string, labels config.InvoiceConfig, leftMargin, contentWidth float64) {
	// Table header with orange background
	pdf.SetFillColor(232, 119, 34) // Orange like Zoho
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	// Column widths
	itemWidth := contentWidth * 0.45
	qtyWidth := contentWidth * 0.15
	rateWidth := contentWidth * 0.20
	amountWidth := contentWidth * 0.20

	// Header row
	pdf.CellFormat(itemWidth, 8, labels.ItemLabel, "", 0, "L", true, 0, "")
	pdf.CellFormat(qtyWidth, 8, labels.QuantityLabel, "", 0, "C", true, 0, "")
	pdf.CellFormat(rateWidth, 8, labels.RateLabel, "", 0, "R", true, 0, "")
	pdf.CellFormat(amountWidth, 8, labels.AmountLabel, "", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetFillColor(255, 255, 255)

	var total float64
	for _, item := range items {
		pdf.SetX(leftMargin)

		// Draw bottom border line
		y := pdf.GetY()

		// Item name
		pdf.CellFormat(itemWidth, 8, item.ItemName, "", 0, "L", false, 0, "")
		// Quantity
		pdf.CellFormat(qtyWidth, 8, fmt.Sprintf("%.0f", item.Quantity), "", 0, "C", false, 0, "")
		// Rate
		pdf.CellFormat(rateWidth, 8, fmt.Sprintf("%.0f", item.Rate), "", 0, "R", false, 0, "")
		// Amount
		pdf.CellFormat(amountWidth, 8, fmt.Sprintf("%.2f", item.Amount), "", 0, "R", false, 0, "")
		pdf.Ln(-1)

		// Draw separator line
		pdf.SetDrawColor(220, 220, 220)
		pdf.Line(leftMargin, pdf.GetY(), leftMargin+contentWidth, pdf.GetY())

		total += item.Amount
		_ = y // avoid unused warning
	}

	// Subtotal row
	pdf.Ln(3)
	pdf.SetX(leftMargin)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(itemWidth+qtyWidth, 7, "", "", 0, "R", false, 0, "")
	pdf.CellFormat(rateWidth, 7, "Subtotal", "", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(amountWidth, 7, fmt.Sprintf("%.2f", total), "", 0, "R", false, 0, "")
	pdf.Ln(-1)

	// Total row with background
	pdf.Ln(5)
	pdf.SetX(leftMargin)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(itemWidth+qtyWidth, 10, "", "", 0, "R", true, 0, "")
	pdf.CellFormat(rateWidth, 10, labels.TotalLabel, "", 0, "R", true, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(amountWidth, 10, fmt.Sprintf("$%.2f", total), "", 0, "R", true, 0, "")
	pdf.Ln(-1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
