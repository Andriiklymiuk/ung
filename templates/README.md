# UNG Custom Templates

This directory contains example HTML templates for generating PDFs. You can customize these templates to match your branding and layout preferences.

## Using Custom Templates

To use custom HTML templates:

1. Copy the example templates to your own directory:
   ```bash
   cp -r templates ~/my-ung-templates
   ```

2. Customize the HTML/CSS in the templates

3. Configure UNG to use your templates in `.ung.yaml`:
   ```yaml
   templates:
     invoice_html: "~/my-ung-templates/invoice.html"
     contract_html: "~/my-ung-templates/contract.html"
   ```

## Template Variables

### Invoice Template (`invoice.html`)

Available variables:
- `{{.InvoiceNum}}` - Invoice number
- `{{.IssuedDate}}` - Issue date (formatted)
- `{{.DueDate}}` - Due date (formatted)
- `{{.Amount}}` - Total amount
- `{{.Currency}}` - Currency code
- `{{.Description}}` - Invoice description/notes
- `{{.Company}}` - Company object with fields:
  - `.Name`, `.Email`, `.Phone`, `.Address`
  - `.RegistrationAddress`, `.TaxID`
  - `.BankName`, `.BankAccount`, `.BankSWIFT`
- `{{.Client}}` - Client object with fields:
  - `.Name`, `.Email`, `.Address`, `.TaxID`
- `{{.LineItems}}` - Array of line items, each with:
  - `.ItemName`, `.Description`
  - `.Quantity`, `.Rate`, `.Amount`
- `{{.Labels}}` - Configured labels (multi-language support):
  - `.InvoiceLabel`, `.FromLabel`, `.BillToLabel`
  - `.ItemLabel`, `.QuantityLabel`, `.RateLabel`, `.AmountLabel`
  - `.TotalLabel`, `.NotesLabel`, `.TermsLabel`, `.Terms`

### Contract Template (`contract.html`)

Available variables:
- `{{.Name}}` - Contract name
- `{{.ContractType}}` - Type (hourly/fixed_price/retainer)
- `{{.HourlyRate}}` - Hourly rate (if applicable)
- `{{.FixedPrice}}` - Fixed price (if applicable)
- `{{.Currency}}` - Currency code
- `{{.StartDate}}` - Start date (formatted)
- `{{.EndDate}}` - End date (formatted, may be nil)
- `{{.Active}}` - Boolean active status
- `{{.Notes}}` - Contract notes
- `{{.Company}}` - Company object (same as invoice)
- `{{.Client}}` - Client object (same as invoice)
- `{{.Labels}}` - Configured labels (same as invoice)

## Rendering

**Note:** Currently, UNG uses gofpdf for PDF generation, which provides a programmatic way to create PDFs. HTML template support is planned for future versions using tools like wkhtmltopdf or chromedp.

These HTML templates serve as:
1. **Documentation** of available data fields
2. **Examples** for creating custom layouts
3. **Future compatibility** when HTML-to-PDF rendering is implemented

To customize PDFs now, you can modify the Go code in:
- `pkg/invoice/pdf.go` - Invoice PDF generation
- `pkg/contract/pdf.go` - Contract PDF generation

## Tips for Customization

1. **Colors**: Change colors in the `<style>` section
2. **Fonts**: Modify font-family in body styles
3. **Layout**: Adjust the `.parties` flex layout or table styles
4. **Logo**: Add `{{if .Company.LogoPath}}` to include company logo
5. **Language**: Use `{{.Labels.*}}` for multi-language support

## Future Enhancement

When HTML rendering is implemented, these templates will be directly used to generate PDFs. The current gofpdf implementation ensures consistent output, while HTML templates provide an easier customization path for the future.
