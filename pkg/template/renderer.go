package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// TemplateDefinition represents a customizable PDF template layout
type TemplateDefinition struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Type        string            `json:"type" yaml:"type"` // "invoice" or "contract"
	PageSize    string            `json:"page_size" yaml:"page_size"` // "A4", "Letter"
	Margins     Margins           `json:"margins" yaml:"margins"`
	Colors      ColorScheme       `json:"colors" yaml:"colors"`
	Fonts       FontConfig        `json:"fonts" yaml:"fonts"`
	Blocks      []TemplateBlock   `json:"blocks" yaml:"blocks"`
}

// Margins defines page margins in mm
type Margins struct {
	Top    float64 `json:"top" yaml:"top"`
	Bottom float64 `json:"bottom" yaml:"bottom"`
	Left   float64 `json:"left" yaml:"left"`
	Right  float64 `json:"right" yaml:"right"`
}

// ColorScheme defines template colors
type ColorScheme struct {
	Primary   HexColor `json:"primary" yaml:"primary"`
	Secondary HexColor `json:"secondary" yaml:"secondary"`
	Text      HexColor `json:"text" yaml:"text"`
	Muted     HexColor `json:"muted" yaml:"muted"`
	Accent    HexColor `json:"accent" yaml:"accent"`
}

// HexColor is a color in hex format (#RRGGBB)
type HexColor string

// ToRGB converts hex color to RGB values
func (c HexColor) ToRGB() (int, int, int) {
	hex := strings.TrimPrefix(string(c), "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// FontConfig defines fonts used in template
type FontConfig struct {
	Family    string  `json:"family" yaml:"family"` // "Arial", "Helvetica", "Times"
	SizeTitle float64 `json:"size_title" yaml:"size_title"`
	SizeBody  float64 `json:"size_body" yaml:"size_body"`
	SizeSmall float64 `json:"size_small" yaml:"size_small"`
}

// TemplateBlock represents a visual block in the template
type TemplateBlock struct {
	Type     string            `json:"type" yaml:"type"` // "header", "company_info", "client_info", "line_items", "totals", "notes", "terms", "spacer"
	Position BlockPosition     `json:"position" yaml:"position"`
	Style    BlockStyle        `json:"style" yaml:"style"`
	Options  map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
}

// BlockPosition defines where a block appears
type BlockPosition struct {
	X      float64 `json:"x" yaml:"x"`           // X position (0 = left margin)
	Y      float64 `json:"y" yaml:"y"`           // Y position (0 = auto)
	Width  float64 `json:"width" yaml:"width"`   // Width (0 = auto/full)
	Height float64 `json:"height" yaml:"height"` // Height (0 = auto)
	Align  string  `json:"align" yaml:"align"`   // "left", "center", "right"
}

// BlockStyle defines visual style for a block
type BlockStyle struct {
	BackgroundColor HexColor `json:"background_color,omitempty" yaml:"background_color,omitempty"`
	TextColor       HexColor `json:"text_color,omitempty" yaml:"text_color,omitempty"`
	BorderColor     HexColor `json:"border_color,omitempty" yaml:"border_color,omitempty"`
	BorderWidth     float64  `json:"border_width,omitempty" yaml:"border_width,omitempty"`
	Padding         float64  `json:"padding,omitempty" yaml:"padding,omitempty"`
	FontSize        float64  `json:"font_size,omitempty" yaml:"font_size,omitempty"`
	FontStyle       string   `json:"font_style,omitempty" yaml:"font_style,omitempty"` // "", "B", "I", "BI"
}

// InvoiceData contains all data for rendering an invoice
type InvoiceData struct {
	Invoice   models.Invoice
	Company   models.Company
	Client    models.Client
	LineItems []models.InvoiceLineItem
	Config    *config.Config
}

// ContractData contains all data for rendering a contract
type ContractData struct {
	Contract models.Contract
	Company  models.Company
	Client   models.Client
	Config   *config.Config
}

// Renderer handles template-based PDF rendering
type Renderer struct {
	template *TemplateDefinition
	pdf      *gofpdf.Fpdf
	data     interface{}
	cfg      *config.Config
	currentY float64
}

// NewRenderer creates a renderer with a template
func NewRenderer(template *TemplateDefinition) *Renderer {
	return &Renderer{
		template: template,
	}
}

// LoadTemplate loads a template from a JSON file
func LoadTemplate(path string) (*TemplateDefinition, error) {
	path = expandPath(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	var tmpl TemplateDefinition
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &tmpl, nil
}

// SaveTemplate saves a template to a JSON file
func SaveTemplate(tmpl *TemplateDefinition, path string) error {
	path = expandPath(path)

	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// GetDefaultInvoiceTemplate returns the default invoice template definition
func GetDefaultInvoiceTemplate() *TemplateDefinition {
	return &TemplateDefinition{
		Name:        "default",
		Description: "Default professional invoice template",
		Type:        "invoice",
		PageSize:    "A4",
		Margins: Margins{
			Top:    15,
			Bottom: 15,
			Left:   15,
			Right:  15,
		},
		Colors: ColorScheme{
			Primary:   "#E87722",
			Secondary: "#505050",
			Text:      "#3C3C3C",
			Muted:     "#808080",
			Accent:    "#2C3E50",
		},
		Fonts: FontConfig{
			Family:    "Arial",
			SizeTitle: 28,
			SizeBody:  10,
			SizeSmall: 9,
		},
		Blocks: []TemplateBlock{
			{
				Type: "header",
				Position: BlockPosition{Align: "left"},
				Style: BlockStyle{FontSize: 28, FontStyle: "B"},
				Options: map[string]interface{}{
					"show_logo": true,
					"show_title": true,
				},
			},
			{
				Type: "spacer",
				Position: BlockPosition{Height: 10},
			},
			{
				Type: "company_info",
				Position: BlockPosition{Width: 50, Align: "left"},
				Style: BlockStyle{FontSize: 9},
			},
			{
				Type: "invoice_meta",
				Position: BlockPosition{Width: 50, Align: "right"},
				Style: BlockStyle{FontSize: 10},
			},
			{
				Type: "spacer",
				Position: BlockPosition{Height: 10},
			},
			{
				Type: "client_info",
				Position: BlockPosition{Width: 50, Align: "left"},
				Style: BlockStyle{FontSize: 9},
			},
			{
				Type: "spacer",
				Position: BlockPosition{Height: 15},
			},
			{
				Type: "line_items",
				Position: BlockPosition{Width: 100},
				Style: BlockStyle{FontSize: 9},
				Options: map[string]interface{}{
					"show_header": true,
					"show_borders": true,
					"zebra_striping": false,
				},
			},
			{
				Type: "totals",
				Position: BlockPosition{Align: "right"},
				Style: BlockStyle{FontSize: 10, FontStyle: "B"},
			},
			{
				Type: "spacer",
				Position: BlockPosition{Height: 15},
			},
			{
				Type: "notes",
				Position: BlockPosition{Width: 100},
				Style: BlockStyle{FontSize: 9},
			},
			{
				Type: "terms",
				Position: BlockPosition{Width: 100},
				Style: BlockStyle{FontSize: 9},
			},
		},
	}
}

// RenderInvoice renders an invoice using a template
func (r *Renderer) RenderInvoice(data InvoiceData, outputPath string) error {
	r.data = data
	r.cfg = data.Config

	// Initialize PDF
	pageSize := r.template.PageSize
	if pageSize == "" {
		pageSize = "A4"
	}
	r.pdf = gofpdf.New("P", "mm", pageSize, "")
	r.pdf.AddPage()

	// Set margins
	r.pdf.SetMargins(r.template.Margins.Left, r.template.Margins.Top, r.template.Margins.Right)
	r.currentY = r.template.Margins.Top

	// Render each block
	for _, block := range r.template.Blocks {
		if err := r.renderBlock(block); err != nil {
			return fmt.Errorf("failed to render block %s: %w", block.Type, err)
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return r.pdf.OutputFileAndClose(outputPath)
}

// renderBlock renders a single template block
func (r *Renderer) renderBlock(block TemplateBlock) error {
	invoiceData, ok := r.data.(InvoiceData)
	if !ok {
		return fmt.Errorf("invalid data type for invoice rendering")
	}

	switch block.Type {
	case "header":
		return r.renderHeader(block, invoiceData)
	case "company_info":
		return r.renderCompanyInfo(block, invoiceData)
	case "client_info":
		return r.renderClientInfo(block, invoiceData)
	case "invoice_meta":
		return r.renderInvoiceMeta(block, invoiceData)
	case "line_items":
		return r.renderLineItems(block, invoiceData)
	case "totals":
		return r.renderTotals(block, invoiceData)
	case "notes":
		return r.renderNotes(block, invoiceData)
	case "terms":
		return r.renderTerms(block, invoiceData)
	case "spacer":
		r.currentY += block.Position.Height
		return nil
	default:
		return nil // Unknown block types are ignored
	}
}

// renderHeader renders the header block
func (r *Renderer) renderHeader(block TemplateBlock, data InvoiceData) error {
	leftMargin := r.template.Margins.Left
	pageWidth := 210.0 // A4 width
	contentWidth := pageWidth - leftMargin - r.template.Margins.Right

	// Company name
	r.pdf.SetFont(r.template.Fonts.Family, "B", 18)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(leftMargin, r.currentY)
	r.pdf.Cell(contentWidth/2, 10, data.Company.Name)

	// Invoice title
	r.pdf.SetFont(r.template.Fonts.Family, "B", r.template.Fonts.SizeTitle)
	r.setTextColor(r.template.Colors.Primary)
	r.pdf.SetXY(pageWidth-r.template.Margins.Right-60, r.currentY)
	r.pdf.Cell(60, 10, r.cfg.Invoice.InvoiceLabel)

	r.currentY += 15
	return nil
}

// renderCompanyInfo renders company information
func (r *Renderer) renderCompanyInfo(block TemplateBlock, data InvoiceData) error {
	leftMargin := r.template.Margins.Left
	fontSize := block.Style.FontSize
	if fontSize == 0 {
		fontSize = r.template.Fonts.SizeSmall
	}

	r.pdf.SetFont(r.template.Fonts.Family, "", fontSize)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(leftMargin, r.currentY)

	company := data.Company
	lineHeight := fontSize * 0.4

	if company.TaxID != "" {
		r.pdf.Cell(80, lineHeight, fmt.Sprintf("Tax ID: %s", company.TaxID))
		r.currentY += lineHeight
		r.pdf.SetXY(leftMargin, r.currentY)
	}

	if company.BankAccount != "" {
		bankInfo := "Bank: " + company.BankAccount
		if company.BankName != "" {
			bankInfo = company.BankName + " | " + company.BankAccount
		}
		if company.BankSWIFT != "" {
			bankInfo += " | SWIFT: " + company.BankSWIFT
		}
		r.pdf.Cell(80, lineHeight, bankInfo)
		r.currentY += lineHeight
		r.pdf.SetXY(leftMargin, r.currentY)
	}

	if company.Address != "" {
		r.pdf.MultiCell(80, lineHeight, company.Address, "", "L", false)
		r.currentY = r.pdf.GetY()
	}

	return nil
}

// renderClientInfo renders client information
func (r *Renderer) renderClientInfo(block TemplateBlock, data InvoiceData) error {
	leftMargin := r.template.Margins.Left
	fontSize := block.Style.FontSize
	if fontSize == 0 {
		fontSize = r.template.Fonts.SizeSmall
	}

	// Bill To label
	r.pdf.SetFont(r.template.Fonts.Family, "B", 10)
	r.setTextColor(r.template.Colors.Primary)
	r.pdf.SetXY(leftMargin, r.currentY)
	r.pdf.Cell(40, 5, r.cfg.Invoice.BillToLabel)
	r.currentY += 7

	// Client name
	r.pdf.SetFont(r.template.Fonts.Family, "B", 11)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(leftMargin, r.currentY)
	r.pdf.Cell(80, 5, data.Client.Name)
	r.currentY += 6

	r.pdf.SetFont(r.template.Fonts.Family, "", fontSize)

	client := data.Client
	lineHeight := fontSize * 0.4

	if client.TaxID != "" {
		r.pdf.SetXY(leftMargin, r.currentY)
		r.pdf.Cell(80, lineHeight, fmt.Sprintf("Tax ID: %s", client.TaxID))
		r.currentY += lineHeight
	}

	if client.Address != "" {
		r.pdf.SetXY(leftMargin, r.currentY)
		r.pdf.MultiCell(80, lineHeight, client.Address, "", "L", false)
		r.currentY = r.pdf.GetY()
	}

	return nil
}

// renderInvoiceMeta renders invoice metadata (number, dates)
func (r *Renderer) renderInvoiceMeta(block TemplateBlock, data InvoiceData) error {
	pageWidth := 210.0
	rightMargin := r.template.Margins.Right
	metaLabelX := pageWidth - rightMargin - 80
	metaValueX := pageWidth - rightMargin - 40

	r.pdf.SetFont(r.template.Fonts.Family, "B", 10)
	r.setTextColor(r.template.Colors.Secondary)

	// Invoice#
	r.pdf.SetXY(metaLabelX, r.currentY)
	r.pdf.Cell(40, 5, "Invoice#")
	r.pdf.SetFont(r.template.Fonts.Family, "", 10)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(metaValueX, r.currentY)
	r.pdf.Cell(40, 5, data.Invoice.InvoiceNum)

	// Invoice Date
	r.pdf.SetFont(r.template.Fonts.Family, "B", 10)
	r.setTextColor(r.template.Colors.Secondary)
	r.pdf.SetXY(metaLabelX, r.currentY+6)
	r.pdf.Cell(40, 5, "Invoice Date")
	r.pdf.SetFont(r.template.Fonts.Family, "", 10)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(metaValueX, r.currentY+6)
	r.pdf.Cell(40, 5, data.Invoice.IssuedDate.Format("02 Jan 2006"))

	// Due Date
	r.pdf.SetFont(r.template.Fonts.Family, "B", 10)
	r.setTextColor(r.template.Colors.Secondary)
	r.pdf.SetXY(metaLabelX, r.currentY+12)
	r.pdf.Cell(40, 5, "Due Date")
	r.pdf.SetFont(r.template.Fonts.Family, "", 10)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(metaValueX, r.currentY+12)
	r.pdf.Cell(40, 5, data.Invoice.DueDate.Format("02 Jan 2006"))

	return nil
}

// renderLineItems renders the line items table
func (r *Renderer) renderLineItems(block TemplateBlock, data InvoiceData) error {
	leftMargin := r.template.Margins.Left
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - r.template.Margins.Right

	// Column widths
	itemWidth := contentWidth * 0.45
	qtyWidth := contentWidth * 0.15
	rateWidth := contentWidth * 0.20
	amountWidth := contentWidth * 0.20

	// Table header
	pr, pg, pb := r.template.Colors.Primary.ToRGB()
	r.pdf.SetFillColor(pr, pg, pb)
	r.pdf.SetTextColor(255, 255, 255)
	r.pdf.SetFont(r.template.Fonts.Family, "B", 9)
	r.pdf.SetXY(leftMargin, r.currentY)

	r.pdf.CellFormat(itemWidth, 8, r.cfg.Invoice.ItemLabel, "", 0, "L", true, 0, "")
	r.pdf.CellFormat(qtyWidth, 8, r.cfg.Invoice.QuantityLabel, "", 0, "C", true, 0, "")
	r.pdf.CellFormat(rateWidth, 8, r.cfg.Invoice.RateLabel, "", 0, "R", true, 0, "")
	r.pdf.CellFormat(amountWidth, 8, r.cfg.Invoice.AmountLabel, "", 0, "R", true, 0, "")
	r.pdf.Ln(-1)
	r.currentY = r.pdf.GetY()

	// Table rows
	r.pdf.SetFont(r.template.Fonts.Family, "", 9)
	r.setTextColor(r.template.Colors.Text)

	for _, item := range data.LineItems {
		r.pdf.SetX(leftMargin)

		// Item name
		r.pdf.CellFormat(itemWidth, 8, item.ItemName, "", 0, "L", false, 0, "")

		// Quantity
		qtyStr := fmt.Sprintf("%.0f", item.Quantity)
		if item.Quantity != float64(int(item.Quantity)) {
			qtyStr = fmt.Sprintf("%.2f", item.Quantity)
		}
		r.pdf.CellFormat(qtyWidth, 8, qtyStr, "", 0, "C", false, 0, "")

		// Rate
		r.pdf.CellFormat(rateWidth, 8, formatCurrency(item.Rate, data.Invoice.Currency), "", 0, "R", false, 0, "")

		// Amount
		r.pdf.CellFormat(amountWidth, 8, formatCurrency(item.Amount, data.Invoice.Currency), "", 0, "R", false, 0, "")
		r.pdf.Ln(-1)

		// Draw separator
		r.pdf.SetDrawColor(220, 220, 220)
		r.pdf.Line(leftMargin, r.pdf.GetY(), leftMargin+contentWidth, r.pdf.GetY())
	}

	r.currentY = r.pdf.GetY()
	return nil
}

// renderTotals renders the totals section
func (r *Renderer) renderTotals(block TemplateBlock, data InvoiceData) error {
	leftMargin := r.template.Margins.Left
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - r.template.Margins.Right

	rateWidth := contentWidth * 0.20
	amountWidth := contentWidth * 0.20
	labelWidth := contentWidth - rateWidth - amountWidth

	// Calculate total
	var total float64
	for _, item := range data.LineItems {
		total += item.Amount
	}

	r.pdf.Ln(5)
	r.pdf.SetX(leftMargin)

	// Total line
	r.pdf.SetDrawColor(200, 200, 200)
	r.pdf.Line(leftMargin+labelWidth, r.pdf.GetY(), leftMargin+contentWidth, r.pdf.GetY())
	r.pdf.Ln(3)
	r.pdf.SetX(leftMargin)

	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetFont(r.template.Fonts.Family, "B", 12)

	r.pdf.CellFormat(labelWidth, 10, "", "", 0, "R", false, 0, "")
	r.pdf.CellFormat(rateWidth, 10, r.cfg.Invoice.TotalLabel, "", 0, "R", false, 0, "")
	r.pdf.SetFont(r.template.Fonts.Family, "B", 14)
	r.pdf.CellFormat(amountWidth, 10, formatCurrency(total, data.Invoice.Currency), "", 0, "R", false, 0, "")

	r.currentY = r.pdf.GetY() + 10
	return nil
}

// renderNotes renders the notes section
func (r *Renderer) renderNotes(block TemplateBlock, data InvoiceData) error {
	if data.Invoice.Description == "" {
		return nil
	}

	leftMargin := r.template.Margins.Left
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - r.template.Margins.Right

	r.pdf.SetXY(leftMargin, r.currentY)
	r.pdf.SetFont(r.template.Fonts.Family, "B", 10)
	r.setTextColor(r.template.Colors.Primary)
	r.pdf.Cell(40, 5, r.cfg.Invoice.NotesLabel)
	r.pdf.SetFont(r.template.Fonts.Family, "", 9)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(leftMargin, r.currentY+6)
	r.pdf.MultiCell(contentWidth, 4, data.Invoice.Description, "", "L", false)

	r.currentY = r.pdf.GetY() + 5
	return nil
}

// renderTerms renders the terms section
func (r *Renderer) renderTerms(block TemplateBlock, data InvoiceData) error {
	leftMargin := r.template.Margins.Left
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - r.template.Margins.Right

	r.pdf.SetXY(leftMargin, r.currentY)
	r.pdf.SetFont(r.template.Fonts.Family, "B", 10)
	r.setTextColor(r.template.Colors.Primary)
	r.pdf.Cell(40, 5, r.cfg.Invoice.TermsLabel)
	r.pdf.SetFont(r.template.Fonts.Family, "", 9)
	r.setTextColor(r.template.Colors.Text)
	r.pdf.SetXY(leftMargin, r.currentY+6)
	r.pdf.MultiCell(contentWidth, 4, r.cfg.Invoice.Terms, "", "L", false)

	r.currentY = r.pdf.GetY()
	return nil
}

// setTextColor sets the text color from HexColor
func (r *Renderer) setTextColor(color HexColor) {
	red, green, blue := color.ToRGB()
	r.pdf.SetTextColor(red, green, blue)
}

// formatCurrency formats amount with currency symbol
func formatCurrency(amount float64, currency string) string {
	symbols := map[string]string{
		"USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥",
		"CAD": "C$", "AUD": "A$", "CHF": "CHF ", "UAH": "₴",
		"PLN": "zł", "CZK": "Kč", "SEK": "kr", "NOK": "kr",
	}
	symbol, ok := symbols[strings.ToUpper(currency)]
	if !ok {
		symbol = currency + " "
	}
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// TemplateInfo contains information about a template
type TemplateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	IsBuiltin   bool   `json:"is_builtin"`
	IsCustom    bool   `json:"is_custom"`
}

// ListTemplates returns available templates
func ListTemplates(templatesDir string) ([]TemplateInfo, error) {
	templates := []TemplateInfo{
		{
			Name:        "default",
			Description: "Default professional template (built-in)",
			Type:        "invoice",
			IsBuiltin:   true,
		},
	}

	// Look for custom templates
	if templatesDir != "" {
		files, err := os.ReadDir(expandPath(templatesDir))
		if err == nil {
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".json") {
					path := filepath.Join(templatesDir, f.Name())
					if tmpl, err := LoadTemplate(path); err == nil {
						templates = append(templates, TemplateInfo{
							Name:        tmpl.Name,
							Description: tmpl.Description,
							Type:        tmpl.Type,
							Path:        path,
							IsCustom:    true,
						})
					}
				}
			}
		}
	}

	return templates, nil
}

// GetSampleInvoiceData returns sample data for template preview
func GetSampleInvoiceData(cfg *config.Config) InvoiceData {
	return InvoiceData{
		Invoice: models.Invoice{
			InvoiceNum:  "INV-2024-001",
			IssuedDate:  time.Now(),
			DueDate:     time.Now().AddDate(0, 0, 30),
			Amount:      5000.00,
			Currency:    "USD",
			Status:      models.StatusPending,
			Description: "Sample invoice for template preview",
		},
		Company: models.Company{
			Name:        "Acme Corporation",
			Email:       "billing@acme.com",
			Phone:       "+1 (555) 123-4567",
			Address:     "123 Business St, Suite 100\nNew York, NY 10001",
			TaxID:       "12-3456789",
			BankName:    "First National Bank",
			BankAccount: "1234567890",
			BankSWIFT:   "FNBKUS33",
		},
		Client: models.Client{
			Name:    "Client Company Inc.",
			Email:   "accounts@client.com",
			Address: "456 Client Ave\nLos Angeles, CA 90001",
			TaxID:   "98-7654321",
		},
		LineItems: []models.InvoiceLineItem{
			{
				ItemName: "Web Development",
				Quantity: 40,
				Rate:     100,
				Amount:   4000,
			},
			{
				ItemName: "Consulting",
				Quantity: 10,
				Rate:     100,
				Amount:   1000,
			},
		},
		Config: cfg,
	}
}

// HasCustomTemplate checks if a custom template is configured
func HasCustomTemplate(cfg *config.Config, docType string) bool {
	switch docType {
	case "invoice":
		return cfg.Templates.InvoiceHTML != ""
	case "contract":
		return cfg.Templates.ContractHTML != ""
	default:
		return false
	}
}

// RenderWithTemplate renders using a custom template if configured
func RenderWithTemplate(templatePath string, data InvoiceData, outputPath string) error {
	tmpl, err := LoadTemplate(templatePath)
	if err != nil {
		return err
	}

	renderer := NewRenderer(tmpl)
	return renderer.RenderInvoice(data, outputPath)
}
