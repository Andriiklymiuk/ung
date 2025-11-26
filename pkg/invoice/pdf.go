package invoice

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// CurrencySymbols maps currency codes to their symbols
var CurrencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"CNY": "¥",
	"CHF": "CHF ",
	"CAD": "C$",
	"AUD": "A$",
	"NZD": "NZ$",
	"INR": "₹",
	"KRW": "₩",
	"BRL": "R$",
	"MXN": "MX$",
	"RUB": "₽",
	"TRY": "₺",
	"PLN": "zł",
	"SEK": "kr",
	"NOK": "kr",
	"DKK": "kr",
	"CZK": "Kč",
	"HUF": "Ft",
	"UAH": "₴",
	"ILS": "₪",
	"SGD": "S$",
	"HKD": "HK$",
	"THB": "฿",
	"ZAR": "R",
}

// FormatCurrency formats an amount with the appropriate currency symbol
func FormatCurrency(amount float64, currency string) string {
	symbol, ok := CurrencySymbols[strings.ToUpper(currency)]
	if !ok {
		symbol = currency + " "
	}
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// GeneratePDF creates a professional PDF invoice with enhanced features
func GeneratePDF(invoice models.Invoice, company models.Company, client models.Client, lineItems []models.InvoiceLineItem) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Load configuration
	cfg, _ := config.Load()
	pdfCfg := cfg.PDF

	// Set up page footer with page numbers
	if pdfCfg.ShowPageNumber {
		pdf.SetFooterFunc(func() {
			pdf.SetY(-15)
			pdf.SetFont("Arial", "I", 8)
			pdf.SetTextColor(128, 128, 128)
			pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
		})
		pdf.AliasNbPages("")
	}

	pdf.AddPage()

	// Draw watermark if enabled
	if pdfCfg.ShowWatermark {
		drawWatermark(pdf, invoice.Status, pdfCfg)
	}

	// Set margins
	leftMargin := 15.0
	rightMargin := 15.0
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - rightMargin

	// Header section with logo and company name
	headerY := 15.0
	logoEndX := leftMargin

	// Draw company logo if available and enabled
	if pdfCfg.ShowLogo && company.LogoPath != "" {
		logoEndX = drawLogo(pdf, company.LogoPath, leftMargin, headerY)
	}

	// Company name (after logo or at start)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(logoEndX+5, headerY)
	pdf.Cell(contentWidth/2-logoEndX, 10, company.Name)

	// INVOICE title on right with custom color
	pdf.SetFont("Arial", "B", 28)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.SetXY(pageWidth-rightMargin-60, headerY)
	pdf.Cell(60, 10, cfg.Invoice.InvoiceLabel)

	// Reset text color
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)

	// Company details below name
	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(leftMargin, 27)

	// Registration number and Tax ID
	if company.TaxID != "" {
		pdf.Cell(contentWidth/2, 4, fmt.Sprintf("Tax ID: %s", company.TaxID))
		pdf.Ln(4)
	}

	// Bank Details
	if company.BankAccount != "" {
		pdf.SetX(leftMargin)
		bankInfo := "Bank: " + company.BankAccount
		if company.BankName != "" {
			bankInfo = company.BankName + " | " + company.BankAccount
		}
		if company.BankSWIFT != "" {
			bankInfo += " | SWIFT: " + company.BankSWIFT
		}
		pdf.Cell(contentWidth/2, 4, bankInfo)
		pdf.Ln(4)
	}

	// Address
	if company.Address != "" {
		pdf.SetX(leftMargin)
		pdf.MultiCell(contentWidth/2-10, 4, company.Address, "", "L", false)
	}

	// Bill To section
	billToY := pdf.GetY() + 8
	pdf.SetXY(leftMargin, billToY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(40, 5, cfg.Invoice.BillToLabel)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(leftMargin, billToY+7)
	pdf.Cell(contentWidth/2, 5, client.Name)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	currentY := billToY + 13

	// Client tax ID
	if client.TaxID != "" {
		pdf.SetXY(leftMargin, currentY)
		pdf.Cell(contentWidth/2, 4, fmt.Sprintf("Tax ID: %s", client.TaxID))
		currentY += 4
	}

	// Client address
	if client.Address != "" {
		pdf.SetXY(leftMargin, currentY)
		pdf.MultiCell(contentWidth/2, 4, client.Address, "", "L", false)
	}

	// Invoice metadata on the right side
	metaStartY := billToY
	metaLabelX := pageWidth - rightMargin - 80
	metaValueX := pageWidth - rightMargin - 40

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)

	// Invoice#
	pdf.SetXY(metaLabelX, metaStartY)
	pdf.Cell(40, 5, "Invoice#")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(metaValueX, metaStartY)
	pdf.Cell(40, 5, invoice.InvoiceNum)

	// Invoice Date
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)
	pdf.SetXY(metaLabelX, metaStartY+6)
	pdf.Cell(40, 5, "Invoice Date")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(metaValueX, metaStartY+6)
	pdf.Cell(40, 5, invoice.IssuedDate.Format("02 Jan 2006"))

	// Due Date
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)
	pdf.SetXY(metaLabelX, metaStartY+12)
	pdf.Cell(40, 5, "Due Date")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(metaValueX, metaStartY+12)
	pdf.Cell(40, 5, invoice.DueDate.Format("02 Jan 2006"))

	// Status badge
	pdf.SetXY(metaLabelX, metaStartY+20)
	drawStatusBadge(pdf, invoice.Status, pdfCfg)

	// Line Items Table
	tableY := pdf.GetY() + 15
	if tableY < currentY+15 {
		tableY = currentY + 15
	}
	pdf.SetXY(leftMargin, tableY)

	// Ensure line items have proper names
	for i := range lineItems {
		if lineItems[i].ItemName == "" {
			lineItems[i].ItemName = fmt.Sprintf("Software services in %s", invoice.IssuedDate.Format("January"))
		}
	}

	totals := drawLineItemsTable(pdf, lineItems, invoice.Currency, cfg.Invoice, pdfCfg, leftMargin, contentWidth)

	// Draw tax breakdown if enabled
	if pdfCfg.ShowTaxBreakdown && pdfCfg.TaxRate > 0 {
		drawTaxBreakdown(pdf, totals, invoice.Currency, pdfCfg, leftMargin, contentWidth)
	}

	// Balance Due section with highlighting
	drawBalanceDue(pdf, totals.GrandTotal, invoice.Currency, invoice.Status, pdfCfg, leftMargin, contentWidth)

	// Notes section
	notesY := pdf.GetY() + 15
	pdf.SetXY(leftMargin, notesY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(40, 5, cfg.Invoice.NotesLabel)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(leftMargin, notesY+6)
	notesText := "Thank you for your business!"
	if invoice.Description != "" {
		notesText = invoice.Description
	}
	pdf.MultiCell(contentWidth, 4, notesText, "", "L", false)

	// Terms & Conditions
	termsY := pdf.GetY() + 10
	pdf.SetXY(leftMargin, termsY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(40, 5, cfg.Invoice.TermsLabel)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(leftMargin, termsY+6)
	pdf.MultiCell(contentWidth, 4, cfg.Invoice.Terms, "", "L", false)

	// Save PDF
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

// InvoiceTotals holds calculated invoice totals
type InvoiceTotals struct {
	Subtotal    float64
	Discount    float64
	TaxableAmount float64
	TaxAmount   float64
	GrandTotal  float64
}

// drawLogo draws the company logo and returns the X position after the logo
func drawLogo(pdf *gofpdf.Fpdf, logoPath string, x, y float64) float64 {
	// Expand ~ in path
	if strings.HasPrefix(logoPath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			logoPath = filepath.Join(home, logoPath[1:])
		}
	}

	// Check if file exists
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		return x
	}

	// Determine image type
	ext := strings.ToLower(filepath.Ext(logoPath))
	var imgType string
	switch ext {
	case ".png":
		imgType = "PNG"
	case ".jpg", ".jpeg":
		imgType = "JPEG"
	case ".gif":
		imgType = "GIF"
	default:
		return x
	}

	// Register and draw image
	pdf.RegisterImage(logoPath, imgType)
	imgInfo := pdf.GetImageInfo(logoPath)
	if imgInfo == nil {
		return x
	}

	// Scale logo to max height of 15mm while maintaining aspect ratio
	maxHeight := 15.0
	ratio := maxHeight / imgInfo.Height()
	width := imgInfo.Width() * ratio

	pdf.Image(logoPath, x, y, width, maxHeight, false, imgType, 0, "")

	return x + width
}

// drawWatermark draws a diagonal watermark based on invoice status
func drawWatermark(pdf *gofpdf.Fpdf, status models.InvoiceStatus, cfg config.PDFConfig) {
	var text string
	var r, g, b int

	switch status {
	case models.StatusPaid:
		text = cfg.PaidLabel
		r, g, b = 0, 150, 0 // Green
	case models.StatusOverdue:
		text = cfg.OverdueLabel
		r, g, b = 200, 0, 0 // Red
	case models.StatusPending:
		// No watermark for pending
		return
	case models.StatusSent:
		// No watermark for sent
		return
	default:
		text = cfg.DraftLabel
		r, g, b = 150, 150, 150 // Gray
	}

	if cfg.WatermarkText != "" {
		text = cfg.WatermarkText
	}

	if text == "" {
		return
	}

	// Save current state
	pdf.SetAlpha(0.1, "Normal")
	pdf.SetFont("Arial", "B", 80)
	pdf.SetTextColor(r, g, b)

	// Calculate center position and draw rotated text
	pageWidth, pageHeight := pdf.GetPageSize()
	textWidth := pdf.GetStringWidth(text)

	// Draw watermark diagonally across the page
	pdf.TransformBegin()
	pdf.TransformRotate(-45, pageWidth/2, pageHeight/2)
	pdf.Text(pageWidth/2-textWidth/2, pageHeight/2, text)
	pdf.TransformEnd()

	// Reset alpha
	pdf.SetAlpha(1.0, "Normal")
}

// drawStatusBadge draws a colored status badge (only for paid and overdue)
func drawStatusBadge(pdf *gofpdf.Fpdf, status models.InvoiceStatus, cfg config.PDFConfig) {
	var text string
	var r, g, b int

	switch status {
	case models.StatusPaid:
		text = cfg.PaidLabel
		r, g, b = 34, 139, 34 // Forest Green
	case models.StatusOverdue:
		text = cfg.OverdueLabel
		r, g, b = 220, 20, 60 // Crimson
	default:
		// Don't show status badge for pending or sent invoices
		return
	}

	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 8)

	width := pdf.GetStringWidth(text) + 8
	pdf.CellFormat(width, 6, text, "", 0, "C", true, 0, "")
}

// drawLineItemsTable creates the line items table and returns totals
func drawLineItemsTable(pdf *gofpdf.Fpdf, items []models.InvoiceLineItem, currency string, labels config.InvoiceConfig, pdfCfg config.PDFConfig, leftMargin, contentWidth float64) InvoiceTotals {
	// Table header with primary color
	pdf.SetFillColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	// Check if any items have discounts
	hasDiscounts := false
	for _, item := range items {
		if item.Discount > 0 || item.DiscountPct > 0 {
			hasDiscounts = true
			break
		}
	}

	// Column widths - adjust based on whether discounts are shown
	var itemWidth, qtyWidth, rateWidth, discountWidth, amountWidth float64
	if hasDiscounts {
		itemWidth = contentWidth * 0.35
		qtyWidth = contentWidth * 0.12
		rateWidth = contentWidth * 0.18
		discountWidth = contentWidth * 0.15
		amountWidth = contentWidth * 0.20
	} else {
		itemWidth = contentWidth * 0.45
		qtyWidth = contentWidth * 0.15
		rateWidth = contentWidth * 0.20
		amountWidth = contentWidth * 0.20
		discountWidth = 0
	}

	// Header row
	pdf.CellFormat(itemWidth, 8, labels.ItemLabel, "", 0, "L", true, 0, "")
	pdf.CellFormat(qtyWidth, 8, labels.QuantityLabel, "", 0, "C", true, 0, "")
	pdf.CellFormat(rateWidth, 8, labels.RateLabel, "", 0, "R", true, 0, "")
	if hasDiscounts {
		pdf.CellFormat(discountWidth, 8, pdfCfg.DiscountLabel, "", 0, "R", true, 0, "")
	}
	pdf.CellFormat(amountWidth, 8, labels.AmountLabel, "", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetFillColor(255, 255, 255)

	var totals InvoiceTotals
	for _, item := range items {
		pdf.SetX(leftMargin)

		// Calculate effective discount
		discount := item.Discount
		if item.DiscountPct > 0 {
			discount = item.Amount * (item.DiscountPct / 100)
		}

		lineAmount := item.Amount - discount

		// Item name (with description if available)
		itemText := item.ItemName
		pdf.CellFormat(itemWidth, 8, itemText, "", 0, "L", false, 0, "")

		// Quantity - show decimal only if needed
		qtyStr := fmt.Sprintf("%.0f", item.Quantity)
		if item.Quantity != float64(int(item.Quantity)) {
			qtyStr = fmt.Sprintf("%.2f", item.Quantity)
		}
		pdf.CellFormat(qtyWidth, 8, qtyStr, "", 0, "C", false, 0, "")

		// Rate
		pdf.CellFormat(rateWidth, 8, FormatCurrency(item.Rate, currency), "", 0, "R", false, 0, "")

		// Discount
		if hasDiscounts {
			if discount > 0 {
				pdf.SetTextColor(220, 20, 60) // Red for discount
				pdf.CellFormat(discountWidth, 8, "-"+FormatCurrency(discount, currency), "", 0, "R", false, 0, "")
				pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
			} else {
				pdf.CellFormat(discountWidth, 8, "-", "", 0, "R", false, 0, "")
			}
		}

		// Amount
		pdf.CellFormat(amountWidth, 8, FormatCurrency(lineAmount, currency), "", 0, "R", false, 0, "")
		pdf.Ln(-1)

		// Draw separator line
		pdf.SetDrawColor(220, 220, 220)
		pdf.Line(leftMargin, pdf.GetY(), leftMargin+contentWidth, pdf.GetY())

		totals.Subtotal += item.Amount
		totals.Discount += discount
		totals.TaxAmount += item.TaxAmount
	}

	totals.TaxableAmount = totals.Subtotal - totals.Discount
	totals.GrandTotal = totals.TaxableAmount + totals.TaxAmount

	// Subtotal row
	pdf.Ln(3)
	pdf.SetX(leftMargin)
	pdf.SetFont("Arial", "", 9)
	summaryLabelWidth := itemWidth + qtyWidth
	if hasDiscounts {
		summaryLabelWidth += discountWidth
	}
	pdf.CellFormat(summaryLabelWidth, 7, "", "", 0, "R", false, 0, "")
	pdf.CellFormat(rateWidth, 7, pdfCfg.SubtotalLabel, "", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(amountWidth, 7, FormatCurrency(totals.Subtotal, currency), "", 0, "R", false, 0, "")
	pdf.Ln(-1)

	// Discount row (if any)
	if totals.Discount > 0 {
		pdf.SetX(leftMargin)
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(summaryLabelWidth, 7, "", "", 0, "R", false, 0, "")
		pdf.SetTextColor(220, 20, 60)
		pdf.CellFormat(rateWidth, 7, pdfCfg.DiscountLabel, "", 0, "R", false, 0, "")
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(amountWidth, 7, "-"+FormatCurrency(totals.Discount, currency), "", 0, "R", false, 0, "")
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		pdf.Ln(-1)
	}

	return totals
}

// drawTaxBreakdown draws the tax/VAT breakdown section
func drawTaxBreakdown(pdf *gofpdf.Fpdf, totals InvoiceTotals, currency string, cfg config.PDFConfig, leftMargin, contentWidth float64) {
	taxAmount := totals.TaxableAmount * cfg.TaxRate

	pdf.SetX(leftMargin)
	pdf.SetFont("Arial", "", 9)

	// Tax row
	rateWidth := contentWidth * 0.20
	amountWidth := contentWidth * 0.20
	labelWidth := contentWidth - rateWidth - amountWidth

	pdf.CellFormat(labelWidth, 7, "", "", 0, "R", false, 0, "")
	pdf.CellFormat(rateWidth, 7, fmt.Sprintf("%s (%.0f%%)", cfg.TaxLabel, cfg.TaxRate*100), "", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(amountWidth, 7, FormatCurrency(taxAmount, currency), "", 0, "R", false, 0, "")
	pdf.Ln(-1)
}

// drawBalanceDue draws the total section without colored background
func drawBalanceDue(pdf *gofpdf.Fpdf, total float64, currency string, status models.InvoiceStatus, cfg config.PDFConfig, leftMargin, contentWidth float64) {
	pdf.Ln(5)
	pdf.SetX(leftMargin)

	rateWidth := contentWidth * 0.20
	amountWidth := contentWidth * 0.20
	labelWidth := contentWidth - rateWidth - amountWidth

	// Draw a simple line separator
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(leftMargin+labelWidth, pdf.GetY(), leftMargin+contentWidth, pdf.GetY())
	pdf.Ln(3)
	pdf.SetX(leftMargin)

	// Use dark text on white background for cleaner look
	pdf.SetTextColor(cfg.TextColor.R, cfg.TextColor.G, cfg.TextColor.B)
	pdf.SetFont("Arial", "B", 12)

	// Always use "Total" label
	label := "Total"

	pdf.CellFormat(labelWidth, 10, "", "", 0, "R", false, 0, "")
	pdf.CellFormat(rateWidth, 10, label, "", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(amountWidth, 10, FormatCurrency(total, currency), "", 0, "R", false, 0, "")
	pdf.Ln(-1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
