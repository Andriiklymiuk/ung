package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/pkg/invoice"
	"github.com/jung-kurt/gofpdf"
)

// GeneratePDF creates a professional PDF for a contract with enhanced features
func GeneratePDF(contract models.Contract, company models.Company, client models.Client) (string, error) {
	cfg, _ := config.Load()
	pdfCfg := cfg.PDF

	pdf := gofpdf.New("P", "mm", "A4", "")

	// Note: Page numbers disabled for contracts since they are typically single-page documents
	// For multi-page contracts, the page number would show "Page 1/1" which looks odd

	pdf.AddPage()

	// Note: Watermark disabled for contracts - not needed for professional look
	// if pdfCfg.ShowWatermark {
	// 	drawContractWatermark(pdf, contract.Active, pdfCfg)
	// }

	// Set margins
	leftMargin := 15.0
	rightMargin := 15.0
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - rightMargin

	// Header section with logo
	headerY := 15.0
	logoEndX := leftMargin

	// Draw company logo if available and enabled
	if pdfCfg.ShowLogo && company.LogoPath != "" {
		logoEndX = drawLogo(pdf, company.LogoPath, leftMargin, headerY)
	}

	// Company name (after logo or at start)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(logoEndX+5, headerY)
	pdf.Cell(contentWidth/2-logoEndX, 10, company.Name)

	// CONTRACT title on right with custom color
	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.SetXY(pageWidth-rightMargin-60, headerY)
	pdf.Cell(60, 10, "CONTRACT")
	pdf.Ln(15)

	// Contract metadata section (right side)
	metaStartY := 32.0
	metaLabelX := pageWidth - rightMargin - 80
	metaValueX := pageWidth - rightMargin - 40

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)

	// Contract Number
	pdf.SetXY(metaLabelX, metaStartY)
	pdf.Cell(40, 5, "Contract#")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(metaValueX, metaStartY)
	pdf.Cell(45, 5, contract.ContractNum)

	// Type (moved up since Name is removed)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)
	pdf.SetXY(metaLabelX, metaStartY+6)
	pdf.Cell(40, 5, "Type")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(metaValueX, metaStartY+6)
	pdf.Cell(45, 5, formatContractType(contract.ContractType))

	// Start Date
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)
	pdf.SetXY(metaLabelX, metaStartY+12)
	pdf.Cell(40, 5, "Start Date")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(metaValueX, metaStartY+12)
	pdf.Cell(45, 5, contract.StartDate.Format("02 Jan 2006"))

	// End Date
	if contract.EndDate != nil {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)
		pdf.SetXY(metaLabelX, metaStartY+18)
		pdf.Cell(40, 5, "End Date")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		pdf.SetXY(metaValueX, metaStartY+18)
		pdf.Cell(45, 5, contract.EndDate.Format("02 Jan 2006"))
	}

	// Note: Status badge removed - not needed for professional contracts

	// Two-column layout for Bill To/From
	// Position below the metadata section to avoid overlap
	currentY := 75.0

	// Left column - Bill To (Client)
	pdf.SetXY(leftMargin, currentY)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(90, 6, cfg.Invoice.BillToLabel)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(leftMargin, currentY+7)
	pdf.Cell(90, 5, client.Name)

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	clientY := currentY + 13
	if client.Email != "" {
		pdf.SetXY(leftMargin, clientY)
		pdf.Cell(90, 5, client.Email)
		clientY += 5
	}
	if client.Address != "" {
		pdf.SetXY(leftMargin, clientY)
		pdf.MultiCell(85, 5, client.Address, "", "L", false)
		clientY = pdf.GetY()
	}
	if client.TaxID != "" {
		pdf.SetXY(leftMargin, clientY)
		pdf.Cell(90, 5, fmt.Sprintf("Tax ID: %s", client.TaxID))
		clientY += 5
	}

	leftColumnEndY := clientY

	// Right column - From (Company)
	pdf.SetXY(110, currentY)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(85, 6, cfg.Invoice.FromLabel)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(110, currentY+7)
	pdf.Cell(85, 5, company.Name)

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	companyY := currentY + 13
	if company.Email != "" {
		pdf.SetXY(110, companyY)
		pdf.Cell(85, 5, company.Email)
		companyY += 5
	}
	if company.Phone != "" {
		pdf.SetXY(110, companyY)
		pdf.Cell(85, 5, company.Phone)
		companyY += 5
	}
	if company.Address != "" {
		pdf.SetXY(110, companyY)
		pdf.MultiCell(80, 5, company.Address, "", "L", false)
		companyY = pdf.GetY()
	}
	if company.TaxID != "" {
		pdf.SetXY(110, companyY)
		pdf.Cell(85, 5, fmt.Sprintf("Tax ID: %s", company.TaxID))
		companyY += 5
	}

	// Bank details
	if company.BankName != "" || company.BankAccount != "" {
		pdf.SetXY(110, companyY+2)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(pdfCfg.SecondaryColor.R, pdfCfg.SecondaryColor.G, pdfCfg.SecondaryColor.B)
		pdf.Cell(85, 5, "Bank Details:")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		companyY += 7
		if company.BankName != "" {
			pdf.SetXY(110, companyY)
			pdf.Cell(85, 5, company.BankName)
			companyY += 5
		}
		if company.BankAccount != "" {
			pdf.SetXY(110, companyY)
			pdf.Cell(85, 5, fmt.Sprintf("Account: %s", company.BankAccount))
			companyY += 5
		}
		if company.BankSWIFT != "" {
			pdf.SetXY(110, companyY)
			pdf.Cell(85, 5, fmt.Sprintf("SWIFT: %s", company.BankSWIFT))
			companyY += 5
		}
	}

	rightColumnEndY := companyY

	// Move past both columns
	maxY := leftColumnEndY
	if rightColumnEndY > maxY {
		maxY = rightColumnEndY
	}
	pdf.SetY(maxY + 10)

	// Contract Details Section
	pdf.SetX(leftMargin)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(190, 8, "Contract Details")
	pdf.Ln(10)

	// Contract terms table with colored header
	pdf.SetX(leftMargin)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(60, 8, "Term", "1", 0, "L", true, 0, "")
	pdf.CellFormat(contentWidth-60, 8, "Value", "1", 1, "L", true, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetFillColor(255, 255, 255)

	// Contract Type
	pdf.SetX(leftMargin)
	pdf.CellFormat(60, 7, "Contract Type", "1", 0, "L", false, 0, "")
	pdf.CellFormat(contentWidth-60, 7, formatContractType(contract.ContractType), "1", 1, "L", false, 0, "")

	// Rate/Price with currency formatting
	if contract.HourlyRate != nil {
		pdf.SetX(leftMargin)
		pdf.CellFormat(60, 7, "Hourly Rate", "1", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(contentWidth-60, 7, fmt.Sprintf("%s/hour", invoice.FormatCurrency(*contract.HourlyRate, contract.Currency)), "1", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
	}
	if contract.FixedPrice != nil {
		pdf.SetX(leftMargin)
		pdf.CellFormat(60, 7, "Fixed Price", "1", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(contentWidth-60, 7, invoice.FormatCurrency(*contract.FixedPrice, contract.Currency), "1", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
	}

	// Dates
	pdf.SetX(leftMargin)
	pdf.CellFormat(60, 7, "Start Date", "1", 0, "L", false, 0, "")
	pdf.CellFormat(contentWidth-60, 7, contract.StartDate.Format("January 2, 2006"), "1", 1, "L", false, 0, "")

	if contract.EndDate != nil {
		pdf.SetX(leftMargin)
		pdf.CellFormat(60, 7, "End Date", "1", 0, "L", false, 0, "")
		pdf.CellFormat(contentWidth-60, 7, contract.EndDate.Format("January 2, 2006"), "1", 1, "L", false, 0, "")
	}

	// Note: Status row removed - not needed for professional contracts

	pdf.Ln(5)

	// Notes section
	if contract.Notes != "" {
		pdf.SetX(leftMargin)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
		pdf.Cell(190, 8, cfg.Invoice.NotesLabel)
		pdf.Ln(8)
		pdf.SetX(leftMargin)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		pdf.MultiCell(contentWidth, 5, contract.Notes, "", "L", false)
		pdf.Ln(5)
	}

	// Terms & Conditions
	pdf.SetX(leftMargin)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(pdfCfg.PrimaryColor.R, pdfCfg.PrimaryColor.G, pdfCfg.PrimaryColor.B)
	pdf.Cell(190, 8, cfg.Invoice.TermsLabel)
	pdf.Ln(8)
	pdf.SetX(leftMargin)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.MultiCell(contentWidth, 4, cfg.Invoice.Terms, "", "L", false)

	// Signature lines at bottom
	pdf.Ln(15)
	drawSignatureLines(pdf, company.Name, client.Name, leftMargin, contentWidth, pdfCfg)

	// Save PDF
	contractsDir := getContractsDir()
	filename := fmt.Sprintf("%s_%s.pdf", client.Name, contract.Name)
	filename = sanitizeFilename(filename)
	pdfPath := filepath.Join(contractsDir, filename)

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
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

	// Scale logo to max height of 12mm while maintaining aspect ratio
	maxHeight := 12.0
	ratio := maxHeight / imgInfo.Height()
	width := imgInfo.Width() * ratio

	pdf.Image(logoPath, x, y, width, maxHeight, false, imgType, 0, "")

	return x + width
}

// drawContractWatermark draws a diagonal watermark based on contract status
func drawContractWatermark(pdf *gofpdf.Fpdf, active bool, cfg config.PDFConfig) {
	var text string
	var r, g, b int

	if active {
		text = "ACTIVE"
		r, g, b = 0, 150, 0 // Green
	} else {
		text = "INACTIVE"
		r, g, b = 150, 150, 150 // Gray
	}

	if cfg.WatermarkText != "" {
		text = cfg.WatermarkText
	}

	// Save current state
	pdf.SetAlpha(0.08, "Normal")
	pdf.SetFont("Helvetica", "B", 70)
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

// drawStatusBadge draws a colored status badge for contracts
func drawStatusBadge(pdf *gofpdf.Fpdf, active bool, cfg config.PDFConfig) {
	var text string
	var r, g, b int

	if active {
		text = "ACTIVE"
		r, g, b = 34, 139, 34 // Forest Green
	} else {
		text = "INACTIVE"
		r, g, b = 128, 128, 128 // Gray
	}

	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 8)

	width := pdf.GetStringWidth(text) + 8
	pdf.CellFormat(width, 6, text, "", 0, "C", true, 0, "")
}

// drawSignatureLines draws signature lines for both parties
func drawSignatureLines(pdf *gofpdf.Fpdf, companyName, clientName string, leftMargin, contentWidth float64, cfg config.PDFConfig) {
	lineWidth := contentWidth/2 - 10

	// Company signature (left)
	pdf.SetX(leftMargin)
	pdf.SetDrawColor(100, 100, 100)
	pdf.Line(leftMargin, pdf.GetY(), leftMargin+lineWidth, pdf.GetY())
	pdf.Ln(2)
	pdf.SetX(leftMargin)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(cfg.TextColor.R, cfg.TextColor.G, cfg.TextColor.B)
	pdf.Cell(lineWidth, 5, companyName)
	pdf.Ln(4)
	pdf.SetX(leftMargin)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(lineWidth, 5, "Signature & Date")

	// Client signature (right)
	clientStartX := leftMargin + contentWidth/2 + 10
	pdf.SetXY(clientStartX, pdf.GetY()-11)
	pdf.Line(clientStartX, pdf.GetY(), clientStartX+lineWidth, pdf.GetY())
	pdf.Ln(2)
	pdf.SetX(clientStartX)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(cfg.TextColor.R, cfg.TextColor.G, cfg.TextColor.B)
	pdf.Cell(lineWidth, 5, clientName)
	pdf.Ln(4)
	pdf.SetX(clientStartX)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(lineWidth, 5, "Signature & Date")
}

// formatContractType returns a human-readable contract type
func formatContractType(ct models.ContractType) string {
	switch ct {
	case models.ContractTypeHourly:
		return "Hourly Rate"
	case models.ContractTypeFixedPrice:
		return "Fixed Price"
	case models.ContractTypeRetainer:
		return "Retainer"
	default:
		return string(ct)
	}
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getContractsDir returns the contracts directory path
func getContractsDir() string {
	contractsDir := config.GetContractsDir()

	// Ensure directory exists
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		// Fallback to invoices dir if we can't create contracts dir
		return db.GetInvoicesDir()
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
