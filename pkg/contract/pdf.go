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

// GeneratePDF creates a professional PDF for a contract as a clean one-page document
func GeneratePDF(contract models.Contract, company models.Company, client models.Client) (string, error) {
	cfg, _ := config.Load()
	pdfCfg := cfg.PDF

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	leftMargin := 20.0
	rightMargin := 20.0
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - rightMargin

	// Document Title - centered at top
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(leftMargin, 20)
	pdf.CellFormat(contentWidth, 10, "SERVICE AGREEMENT", "", 1, "C", false, 0, "")

	// Contract reference line
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetXY(leftMargin, 32)
	pdf.CellFormat(contentWidth, 6, fmt.Sprintf("Contract Reference: %s", contract.ContractNum), "", 1, "C", false, 0, "")

	// Horizontal line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(leftMargin, 42, pageWidth-rightMargin, 42)

	// Introduction paragraph
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetXY(leftMargin, 50)

	introText := fmt.Sprintf("This Service Agreement is entered into as of %s between the following parties:",
		contract.StartDate.Format("January 2, 2006"))
	pdf.MultiCell(contentWidth, 6, introText, "", "L", false)

	pdf.Ln(8)

	// Parties section
	currentY := pdf.GetY()

	// Provider (Company)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(leftMargin, currentY)
	pdf.Cell(contentWidth/2-5, 6, "Service Provider:")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	providerY := currentY + 7
	pdf.SetXY(leftMargin, providerY)
	pdf.Cell(contentWidth/2-5, 5, company.Name)
	providerY += 5
	if company.Address != "" {
		pdf.SetXY(leftMargin, providerY)
		pdf.MultiCell(contentWidth/2-10, 5, company.Address, "", "L", false)
		providerY = pdf.GetY()
	}
	if company.TaxID != "" {
		pdf.SetXY(leftMargin, providerY)
		pdf.Cell(contentWidth/2-5, 5, fmt.Sprintf("Tax ID: %s", company.TaxID))
		providerY += 5
	}

	// Client
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(pageWidth/2+5, currentY)
	pdf.Cell(contentWidth/2-5, 6, "Client:")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	clientY := currentY + 7
	pdf.SetXY(pageWidth/2+5, clientY)
	pdf.Cell(contentWidth/2-5, 5, client.Name)
	clientY += 5
	if client.Address != "" {
		pdf.SetXY(pageWidth/2+5, clientY)
		pdf.MultiCell(contentWidth/2-10, 5, client.Address, "", "L", false)
		clientY = pdf.GetY()
	}
	if client.TaxID != "" {
		pdf.SetXY(pageWidth/2+5, clientY)
		pdf.Cell(contentWidth/2-5, 5, fmt.Sprintf("Tax ID: %s", client.TaxID))
		clientY += 5
	}

	// Move past both columns
	maxY := providerY
	if clientY > maxY {
		maxY = clientY
	}
	pdf.SetY(maxY + 10)

	// Terms of Agreement section
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetX(leftMargin)
	pdf.Cell(contentWidth, 8, "Terms of Agreement")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)

	// Build terms text
	termsItems := []string{}

	// Contract type and rate
	if contract.HourlyRate != nil {
		termsItems = append(termsItems, fmt.Sprintf("The Service Provider agrees to provide services at an hourly rate of %s.",
			invoice.FormatCurrency(*contract.HourlyRate, contract.Currency)))
	}
	if contract.FixedPrice != nil {
		termsItems = append(termsItems, fmt.Sprintf("The total fixed price for services is %s.",
			invoice.FormatCurrency(*contract.FixedPrice, contract.Currency)))
	}

	// Duration
	if contract.EndDate != nil {
		termsItems = append(termsItems, fmt.Sprintf("This agreement is effective from %s until %s.",
			contract.StartDate.Format("January 2, 2006"), contract.EndDate.Format("January 2, 2006")))
	} else {
		termsItems = append(termsItems, fmt.Sprintf("This agreement is effective from %s and continues until terminated by either party.",
			contract.StartDate.Format("January 2, 2006")))
	}

	// Render terms as numbered list
	for i, term := range termsItems {
		pdf.SetX(leftMargin)
		pdf.MultiCell(contentWidth, 6, fmt.Sprintf("%d. %s", i+1, term), "", "L", false)
		pdf.Ln(2)
	}

	// Notes section (if any)
	if contract.Notes != "" {
		pdf.Ln(5)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(40, 40, 40)
		pdf.SetX(leftMargin)
		pdf.Cell(contentWidth, 8, "Additional Terms")
		pdf.Ln(8)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		pdf.SetX(leftMargin)
		pdf.MultiCell(contentWidth, 5, contract.Notes, "", "L", false)
	}

	// Payment Information
	if company.BankAccount != "" {
		pdf.Ln(8)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(40, 40, 40)
		pdf.SetX(leftMargin)
		pdf.Cell(contentWidth, 8, "Payment Information")
		pdf.Ln(8)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		pdf.SetX(leftMargin)

		bankInfo := ""
		if company.BankName != "" {
			bankInfo += fmt.Sprintf("Bank: %s\n", company.BankName)
		}
		bankInfo += fmt.Sprintf("Account: %s", company.BankAccount)
		if company.BankSWIFT != "" {
			bankInfo += fmt.Sprintf("\nSWIFT: %s", company.BankSWIFT)
		}
		pdf.MultiCell(contentWidth, 5, bankInfo, "", "L", false)
	}

	// General Terms
	if cfg.Invoice.Terms != "" {
		pdf.Ln(8)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(40, 40, 40)
		pdf.SetX(leftMargin)
		pdf.Cell(contentWidth, 8, "General Conditions")
		pdf.Ln(8)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
		pdf.SetX(leftMargin)
		pdf.MultiCell(contentWidth, 5, cfg.Invoice.Terms, "", "L", false)
	}

	// Signature section
	pdf.Ln(15)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(pdfCfg.TextColor.R, pdfCfg.TextColor.G, pdfCfg.TextColor.B)
	pdf.SetX(leftMargin)
	pdf.MultiCell(contentWidth, 5, "By signing below, both parties agree to the terms and conditions set forth in this agreement.", "", "L", false)

	pdf.Ln(10)
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
