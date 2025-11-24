package invoice

import (
	"fmt"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// GeneratePDF creates a professional PDF invoice with itemized billing
func GeneratePDF(invoice models.Invoice, company models.Company, client models.Client, lineItems []models.InvoiceLineItem) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Load configuration for labels
	cfg, _ := config.Load()

	// Header - Invoice Title
	pdf.SetFont("Arial", "B", 26)
	pdf.SetTextColor(60, 60, 60)
	pdf.Cell(190, 12, cfg.Invoice.InvoiceLabel)
	pdf.Ln(15)

	// Invoice metadata (number, dates) - Right aligned
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)

	// Invoice Number
	pdf.CellFormat(140, 5, "", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 5, "Invoice #:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 5, invoice.InvoiceNum)
	pdf.Ln(5)

	// Invoice Date
	pdf.CellFormat(140, 5, "", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 5, "Issue Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 5, invoice.IssuedDate.Format("02 Jan 2006"))
	pdf.Ln(5)

	// Due Date
	pdf.CellFormat(140, 5, "", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 5, "Due Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 5, invoice.DueDate.Format("02 Jan 2006"))
	pdf.Ln(12)

	// Reset color for main content
	pdf.SetTextColor(0, 0, 0)

	// Two-column layout for company and client info
	colWidth := 95.0

	// From (Company) - Left column
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(colWidth, 6, cfg.Invoice.FromLabel)
	pdf.Ln(7)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(colWidth, 5, company.Name)
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 10)
	if company.Email != "" {
		pdf.Cell(colWidth, 4, company.Email)
		pdf.Ln(4)
	}
	if company.Phone != "" {
		pdf.Cell(colWidth, 4, company.Phone)
		pdf.Ln(4)
	}
	if company.Address != "" {
		pdf.MultiCell(colWidth, 4, company.Address, "", "L", false)
	}
	if company.RegistrationAddress != "" && company.RegistrationAddress != company.Address {
		pdf.SetFont("Arial", "I", 9)
		pdf.Cell(colWidth, 4, "Reg: "+company.RegistrationAddress)
		pdf.SetFont("Arial", "", 10)
		pdf.Ln(4)
	}
	if company.TaxID != "" {
		pdf.Cell(colWidth, 4, "Tax ID: "+company.TaxID)
		pdf.Ln(4)
	}

	// Bank details
	if company.BankName != "" || company.BankAccount != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 9)
		pdf.Cell(colWidth, 4, "Bank Details:")
		pdf.Ln(4)
		pdf.SetFont("Arial", "", 9)
		if company.BankName != "" {
			pdf.Cell(colWidth, 3, company.BankName)
			pdf.Ln(3)
		}
		if company.BankAccount != "" {
			pdf.Cell(colWidth, 3, "Account: "+company.BankAccount)
			pdf.Ln(3)
		}
		if company.BankSWIFT != "" {
			pdf.Cell(colWidth, 3, "SWIFT: "+company.BankSWIFT)
			pdf.Ln(3)
		}
	}

	// Bill To (Client) - Right column (set position back to top-right)
	pdf.SetXY(105, 40) // Start at top of right column

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(colWidth, 6, cfg.Invoice.BillToLabel)
	pdf.Ln(7)

	pdf.SetX(105)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(colWidth, 5, client.Name)
	pdf.Ln(5)

	pdf.SetX(105)
	pdf.SetFont("Arial", "", 10)
	if client.Email != "" {
		pdf.Cell(colWidth, 4, client.Email)
		pdf.Ln(4)
		pdf.SetX(105)
	}
	if client.Address != "" {
		currentY := pdf.GetY()
		pdf.SetXY(105, currentY)
		pdf.MultiCell(colWidth, 4, client.Address, "", "L", false)
	}
	if client.TaxID != "" {
		pdf.SetX(105)
		pdf.Cell(colWidth, 4, "Tax ID: "+client.TaxID)
		pdf.Ln(4)
	}

	// Move to after both columns
	pdf.Ln(100) // Ensure we're below both columns

	// Description section
	if invoice.Description != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(190, 6, cfg.Invoice.DescriptionLabel+":")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(190, 5, invoice.Description, "", "L", false)
		pdf.Ln(4)
	}

	// Line Items Table
	pdf.Ln(6)
	drawLineItemsTable(pdf, lineItems, invoice.Currency, cfg.Invoice)

	// Notes section
	pdf.Ln(8)
	if invoice.Description != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(190, 5, cfg.Invoice.NotesLabel+":")
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.MultiCell(190, 4, cfg.Invoice.PaymentNote, "", "L", false)
		pdf.SetTextColor(0, 0, 0)
	}

	// Terms & Conditions
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 5, cfg.Invoice.TermsLabel+":")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.MultiCell(190, 4, cfg.Invoice.Terms, "", "L", false)

	// Save PDF to current directory
	filename := fmt.Sprintf("%s.pdf", invoice.InvoiceNum)
	pdfPath := filename // Save in current directory

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}

// drawLineItemsTable creates a professional table for invoice line items
func drawLineItemsTable(pdf *gofpdf.Fpdf, items []models.InvoiceLineItem, currency string, labels config.InvoiceConfig) {
	// Table header
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)

	// Column widths
	itemWidth := 70.0
	qtyWidth := 25.0
	rateWidth := 40.0
	amountWidth := 45.0

	// Header row
	pdf.CellFormat(itemWidth, 7, labels.ItemLabel, "1", 0, "L", true, 0, "")
	pdf.CellFormat(qtyWidth, 7, labels.QuantityLabel, "1", 0, "C", true, 0, "")
	pdf.CellFormat(rateWidth, 7, labels.RateLabel, "1", 0, "R", true, 0, "")
	pdf.CellFormat(amountWidth, 7, labels.AmountLabel, "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 255, 255)

	var total float64
	for _, item := range items {
		// Item name (with description if available)
		itemText := item.ItemName
		if item.Description != "" {
			itemText = item.ItemName + "\n" + item.Description
		}

		// Calculate row height based on content
		lines := pdf.SplitLines([]byte(itemText), itemWidth-2)
		rowHeight := float64(len(lines)) * 5.0
		if rowHeight < 7 {
			rowHeight = 7
		}

		currentY := pdf.GetY()

		// Item column
		pdf.MultiCell(itemWidth, 5, itemText, "1", "L", false)

		// Move back to draw other columns at the same Y position
		pdf.SetXY(pdf.GetX()+itemWidth, currentY)

		// Quantity
		pdf.CellFormat(qtyWidth, rowHeight, fmt.Sprintf("%.2f", item.Quantity), "1", 0, "C", false, 0, "")

		// Rate
		pdf.CellFormat(rateWidth, rowHeight, fmt.Sprintf("%.2f %s", item.Rate, currency), "1", 0, "R", false, 0, "")

		// Amount
		pdf.CellFormat(amountWidth, rowHeight, fmt.Sprintf("%.2f %s", item.Amount, currency), "1", 0, "R", false, 0, "")

		pdf.Ln(-1)
		total += item.Amount
	}

	// Total row
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(itemWidth+qtyWidth+rateWidth, 9, labels.TotalLabel+":", "1", 0, "R", true, 0, "")
	pdf.CellFormat(amountWidth, 9, fmt.Sprintf("%.2f %s", total, currency), "1", 0, "R", true, 0, "")
	pdf.Ln(-1)
}
