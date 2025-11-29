//
//  PDFGenerator.swift
//  ung
//
//  PDF generation for invoices - cross-platform (macOS/iOS)
//

import SwiftUI
#if os(iOS)
import UIKit
#else
import AppKit
#endif

// MARK: - PDF Generator

class PDFGenerator {

    // MARK: - Configuration

    struct PDFConfig {
        var primaryColor: Color = Color(red: 232/255, green: 119/255, blue: 34/255) // Orange
        var secondaryColor: Color = Color(red: 100/255, green: 100/255, blue: 100/255)
        var textColor: Color = Color(red: 40/255, green: 40/255, blue: 40/255)
        var showWatermark: Bool = true
        var showLogo: Bool = true
        var showPageNumber: Bool = true
        var taxRate: Double = 0.0
        var taxLabel: String = "VAT"
        var taxInclusive: Bool = false
    }

    private let config: PDFConfig

    // Page dimensions (A4 at 72 DPI)
    private let pageWidth: CGFloat = 595
    private let pageHeight: CGFloat = 842
    private let margin: CGFloat = 50

    init(config: PDFConfig = PDFConfig()) {
        self.config = config
    }

    // MARK: - Public API

    /// Generate a PDF for an invoice
    func generateInvoicePDF(
        invoice: Invoice,
        company: Company,
        client: ClientModel,
        lineItems: [InvoiceLineItem]
    ) -> Data? {
        let pageRect = CGRect(x: 0, y: 0, width: pageWidth, height: pageHeight)

        #if os(iOS)
        let renderer = UIGraphicsPDFRenderer(bounds: pageRect)
        return renderer.pdfData { context in
            context.beginPage()
            let cgContext = context.cgContext
            drawInvoicePage(in: cgContext, pageRect: pageRect, invoice: invoice, company: company, client: client, lineItems: lineItems)
        }
        #else
        let pdfData = NSMutableData()
        var mediaBox = pageRect

        guard let consumer = CGDataConsumer(data: pdfData as CFMutableData),
              let context = CGContext(consumer: consumer, mediaBox: &mediaBox, nil) else {
            return nil
        }

        context.beginPDFPage(nil)
        drawInvoicePage(in: context, pageRect: pageRect, invoice: invoice, company: company, client: client, lineItems: lineItems)
        context.endPDFPage()
        context.closePDF()

        return pdfData as Data
        #endif
    }

    /// Save PDF to file and return the path
    func saveInvoicePDF(
        invoice: Invoice,
        company: Company,
        client: ClientModel,
        lineItems: [InvoiceLineItem]
    ) -> URL? {
        guard let pdfData = generateInvoicePDF(invoice: invoice, company: company, client: client, lineItems: lineItems) else {
            return nil
        }

        // Create invoices directory if needed
        let fileManager = FileManager.default
        let documentsPath = fileManager.urls(for: .documentDirectory, in: .userDomainMask)[0]
        let invoicesDir = documentsPath.appendingPathComponent("Invoices")

        try? fileManager.createDirectory(at: invoicesDir, withIntermediateDirectories: true)

        // Save with invoice number as filename
        let filename = "\(invoice.invoiceNum).pdf"
        let fileURL = invoicesDir.appendingPathComponent(filename)

        do {
            try pdfData.write(to: fileURL)
            return fileURL
        } catch {
            print("Failed to save PDF: \(error)")
            return nil
        }
    }

    // MARK: - Drawing Functions

    private func drawInvoicePage(
        in context: CGContext,
        pageRect: CGRect,
        invoice: Invoice,
        company: Company,
        client: ClientModel,
        lineItems: [InvoiceLineItem]
    ) {
        // Flip coordinate system for text drawing
        context.translateBy(x: 0, y: pageRect.height)
        context.scaleBy(x: 1.0, y: -1.0)

        var yOffset: CGFloat = margin

        // Draw header with company info and invoice title
        yOffset = drawHeader(in: context, pageRect: pageRect, invoice: invoice, company: company, yOffset: yOffset)

        // Draw client info
        yOffset = drawClientInfo(in: context, pageRect: pageRect, client: client, yOffset: yOffset)

        // Draw invoice details (number, date, due date)
        yOffset = drawInvoiceDetails(in: context, pageRect: pageRect, invoice: invoice, yOffset: yOffset)

        // Draw line items table
        yOffset = drawLineItems(in: context, pageRect: pageRect, lineItems: lineItems, invoice: invoice, yOffset: yOffset)

        // Draw totals
        yOffset = drawTotals(in: context, pageRect: pageRect, lineItems: lineItems, invoice: invoice, yOffset: yOffset)

        // Draw notes if present
        if let notes = invoice.notes, !notes.isEmpty {
            yOffset = drawNotes(in: context, pageRect: pageRect, notes: notes, yOffset: yOffset)
        }

        // Draw bank details
        yOffset = drawBankDetails(in: context, pageRect: pageRect, company: company, yOffset: yOffset)

        // Draw watermark if paid
        if config.showWatermark && invoice.status.lowercased() == "paid" {
            drawWatermark(in: context, pageRect: pageRect, text: "PAID")
        }

        // Draw page number
        if config.showPageNumber {
            drawPageNumber(in: context, pageRect: pageRect, pageNumber: 1)
        }
    }

    private func drawHeader(
        in context: CGContext,
        pageRect: CGRect,
        invoice: Invoice,
        company: Company,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset
        let contentWidth = pageRect.width - 2 * margin

        // Company name (left)
        let companyNameAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 20, weight: .bold),
            .foregroundColor: platformColor(config.primaryColor)
        ]
        let companyNameStr = NSAttributedString(string: company.name, attributes: companyNameAttrs)
        drawAttributedString(companyNameStr, in: context, at: CGPoint(x: margin, y: y))

        // INVOICE title (right)
        let invoiceTitleAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 28, weight: .bold),
            .foregroundColor: platformColor(config.primaryColor)
        ]
        let invoiceTitleStr = NSAttributedString(string: "INVOICE", attributes: invoiceTitleAttrs)
        let titleWidth = invoiceTitleStr.size().width
        drawAttributedString(invoiceTitleStr, in: context, at: CGPoint(x: pageRect.width - margin - titleWidth, y: y))

        y += 30

        // Company address
        if let address = company.address {
            let addressAttrs: [NSAttributedString.Key: Any] = [
                .font: platformFont(size: 10, weight: .regular),
                .foregroundColor: platformColor(config.secondaryColor)
            ]
            let addressStr = NSAttributedString(string: address, attributes: addressAttrs)
            drawAttributedString(addressStr, in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        // Company email
        let emailAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.secondaryColor)
        ]
        let emailStr = NSAttributedString(string: company.email, attributes: emailAttrs)
        drawAttributedString(emailStr, in: context, at: CGPoint(x: margin, y: y))
        y += 14

        // Tax ID if present
        if let taxId = company.taxId, !taxId.isEmpty {
            let taxIdAttrs: [NSAttributedString.Key: Any] = [
                .font: platformFont(size: 10, weight: .regular),
                .foregroundColor: platformColor(config.secondaryColor)
            ]
            let taxIdStr = NSAttributedString(string: "Tax ID: \(taxId)", attributes: taxIdAttrs)
            drawAttributedString(taxIdStr, in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        y += 20

        // Draw separator line
        context.setStrokeColor(platformColor(config.secondaryColor).cgColor ?? CGColor(gray: 0.5, alpha: 0.3))
        context.setLineWidth(0.5)
        context.move(to: CGPoint(x: margin, y: pageRect.height - y))
        context.addLine(to: CGPoint(x: pageRect.width - margin, y: pageRect.height - y))
        context.strokePath()

        return y + 20
    }

    private func drawClientInfo(
        in context: CGContext,
        pageRect: CGRect,
        client: ClientModel,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset

        // "Bill To" label
        let labelAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .semibold),
            .foregroundColor: platformColor(config.secondaryColor)
        ]
        let labelStr = NSAttributedString(string: "BILL TO", attributes: labelAttrs)
        drawAttributedString(labelStr, in: context, at: CGPoint(x: margin, y: y))
        y += 16

        // Client name
        let nameAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 14, weight: .semibold),
            .foregroundColor: platformColor(config.textColor)
        ]
        let nameStr = NSAttributedString(string: client.name, attributes: nameAttrs)
        drawAttributedString(nameStr, in: context, at: CGPoint(x: margin, y: y))
        y += 18

        // Client address
        if let address = client.address, !address.isEmpty {
            let addressAttrs: [NSAttributedString.Key: Any] = [
                .font: platformFont(size: 10, weight: .regular),
                .foregroundColor: platformColor(config.textColor)
            ]
            let addressStr = NSAttributedString(string: address, attributes: addressAttrs)
            drawAttributedString(addressStr, in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        // Client email
        let emailAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.textColor)
        ]
        let emailStr = NSAttributedString(string: client.email, attributes: emailAttrs)
        drawAttributedString(emailStr, in: context, at: CGPoint(x: margin, y: y))
        y += 14

        // Client tax ID
        if let taxId = client.taxId, !taxId.isEmpty {
            let taxIdAttrs: [NSAttributedString.Key: Any] = [
                .font: platformFont(size: 10, weight: .regular),
                .foregroundColor: platformColor(config.textColor)
            ]
            let taxIdStr = NSAttributedString(string: "Tax ID: \(taxId)", attributes: taxIdAttrs)
            drawAttributedString(taxIdStr, in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        return y + 20
    }

    private func drawInvoiceDetails(
        in context: CGContext,
        pageRect: CGRect,
        invoice: Invoice,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset
        let rightColumn = pageRect.width - margin - 150

        let labelAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.secondaryColor)
        ]

        let valueAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .medium),
            .foregroundColor: platformColor(config.textColor)
        ]

        // Invoice number
        drawAttributedString(NSAttributedString(string: "Invoice Number:", attributes: labelAttrs),
                           in: context, at: CGPoint(x: rightColumn, y: y))
        drawAttributedString(NSAttributedString(string: invoice.invoiceNum, attributes: valueAttrs),
                           in: context, at: CGPoint(x: rightColumn + 100, y: y))
        y += 16

        // Issue date
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "dd MMM yyyy"

        drawAttributedString(NSAttributedString(string: "Invoice Date:", attributes: labelAttrs),
                           in: context, at: CGPoint(x: rightColumn, y: y))
        let issuedDateStr = invoice.issuedDate.map { dateFormatter.string(from: $0) } ?? "-"
        drawAttributedString(NSAttributedString(string: issuedDateStr, attributes: valueAttrs),
                           in: context, at: CGPoint(x: rightColumn + 100, y: y))
        y += 16

        // Due date
        drawAttributedString(NSAttributedString(string: "Due Date:", attributes: labelAttrs),
                           in: context, at: CGPoint(x: rightColumn, y: y))
        let dueDateStr = invoice.dueDate.map { dateFormatter.string(from: $0) } ?? "-"
        drawAttributedString(NSAttributedString(string: dueDateStr, attributes: valueAttrs),
                           in: context, at: CGPoint(x: rightColumn + 100, y: y))

        return y + 30
    }

    private func drawLineItems(
        in context: CGContext,
        pageRect: CGRect,
        lineItems: [InvoiceLineItem],
        invoice: Invoice,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset
        let contentWidth = pageRect.width - 2 * margin

        // Column widths
        let descWidth: CGFloat = contentWidth * 0.45
        let qtyWidth: CGFloat = contentWidth * 0.15
        let rateWidth: CGFloat = contentWidth * 0.20
        let amountWidth: CGFloat = contentWidth * 0.20

        // Table header background
        let headerRect = CGRect(x: margin, y: pageRect.height - y - 20, width: contentWidth, height: 20)
        context.setFillColor(CGColor(gray: 0.95, alpha: 1.0))
        context.fill(headerRect)

        // Table header
        let headerAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 9, weight: .semibold),
            .foregroundColor: platformColor(config.secondaryColor)
        ]

        var xOffset: CGFloat = margin + 5
        drawAttributedString(NSAttributedString(string: "DESCRIPTION", attributes: headerAttrs),
                           in: context, at: CGPoint(x: xOffset, y: y + 5))
        xOffset += descWidth

        drawAttributedString(NSAttributedString(string: "QTY", attributes: headerAttrs),
                           in: context, at: CGPoint(x: xOffset, y: y + 5))
        xOffset += qtyWidth

        drawAttributedString(NSAttributedString(string: "RATE", attributes: headerAttrs),
                           in: context, at: CGPoint(x: xOffset, y: y + 5))
        xOffset += rateWidth

        drawAttributedString(NSAttributedString(string: "AMOUNT", attributes: headerAttrs),
                           in: context, at: CGPoint(x: xOffset, y: y + 5))

        y += 25

        // Line items
        let itemAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.textColor)
        ]

        let currencySymbol = getCurrencySymbol(invoice.currency)

        for item in lineItems {
            xOffset = margin + 5

            drawAttributedString(NSAttributedString(string: item.itemName, attributes: itemAttrs),
                               in: context, at: CGPoint(x: xOffset, y: y))
            xOffset += descWidth

            let qtyStr = item.quantity.truncatingRemainder(dividingBy: 1) == 0
                ? String(format: "%.0f", item.quantity)
                : String(format: "%.2f", item.quantity)
            drawAttributedString(NSAttributedString(string: qtyStr, attributes: itemAttrs),
                               in: context, at: CGPoint(x: xOffset, y: y))
            xOffset += qtyWidth

            drawAttributedString(NSAttributedString(string: "\(currencySymbol)\(String(format: "%.2f", item.rate))", attributes: itemAttrs),
                               in: context, at: CGPoint(x: xOffset, y: y))
            xOffset += rateWidth

            drawAttributedString(NSAttributedString(string: "\(currencySymbol)\(String(format: "%.2f", item.amount))", attributes: itemAttrs),
                               in: context, at: CGPoint(x: xOffset, y: y))

            y += 18

            // Draw separator line
            context.setStrokeColor(CGColor(gray: 0.9, alpha: 1.0))
            context.setLineWidth(0.5)
            context.move(to: CGPoint(x: margin, y: pageRect.height - y + 5))
            context.addLine(to: CGPoint(x: pageRect.width - margin, y: pageRect.height - y + 5))
            context.strokePath()
        }

        return y + 10
    }

    private func drawTotals(
        in context: CGContext,
        pageRect: CGRect,
        lineItems: [InvoiceLineItem],
        invoice: Invoice,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset
        let rightAlign: CGFloat = pageRect.width - margin - 100
        let labelX: CGFloat = rightAlign - 80

        let labelAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.secondaryColor)
        ]

        let valueAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .medium),
            .foregroundColor: platformColor(config.textColor)
        ]

        let currencySymbol = getCurrencySymbol(invoice.currency)
        let subtotal = lineItems.reduce(0) { $0 + $1.amount }

        // Subtotal
        drawAttributedString(NSAttributedString(string: "Subtotal", attributes: labelAttrs),
                           in: context, at: CGPoint(x: labelX, y: y))
        drawAttributedString(NSAttributedString(string: "\(currencySymbol)\(String(format: "%.2f", subtotal))", attributes: valueAttrs),
                           in: context, at: CGPoint(x: rightAlign, y: y))
        y += 18

        // Tax if applicable
        if config.taxRate > 0 {
            let taxAmount = subtotal * config.taxRate
            drawAttributedString(NSAttributedString(string: "\(config.taxLabel) (\(Int(config.taxRate * 100))%)", attributes: labelAttrs),
                               in: context, at: CGPoint(x: labelX, y: y))
            drawAttributedString(NSAttributedString(string: "\(currencySymbol)\(String(format: "%.2f", taxAmount))", attributes: valueAttrs),
                               in: context, at: CGPoint(x: rightAlign, y: y))
            y += 18
        }

        y += 5

        // Total
        let totalAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 14, weight: .bold),
            .foregroundColor: platformColor(config.primaryColor)
        ]

        let totalLabelAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 12, weight: .bold),
            .foregroundColor: platformColor(config.primaryColor)
        ]

        drawAttributedString(NSAttributedString(string: "TOTAL", attributes: totalLabelAttrs),
                           in: context, at: CGPoint(x: labelX, y: y))
        drawAttributedString(NSAttributedString(string: "\(currencySymbol)\(String(format: "%.2f", invoice.amount))", attributes: totalAttrs),
                           in: context, at: CGPoint(x: rightAlign, y: y))

        return y + 30
    }

    private func drawNotes(
        in context: CGContext,
        pageRect: CGRect,
        notes: String,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset

        let labelAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .semibold),
            .foregroundColor: platformColor(config.secondaryColor)
        ]

        let notesAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.textColor)
        ]

        drawAttributedString(NSAttributedString(string: "Notes", attributes: labelAttrs),
                           in: context, at: CGPoint(x: margin, y: y))
        y += 14

        drawAttributedString(NSAttributedString(string: notes, attributes: notesAttrs),
                           in: context, at: CGPoint(x: margin, y: y))

        return y + 30
    }

    private func drawBankDetails(
        in context: CGContext,
        pageRect: CGRect,
        company: Company,
        yOffset: CGFloat
    ) -> CGFloat {
        var y = yOffset

        guard company.bankName != nil || company.bankAccount != nil else {
            return y
        }

        let labelAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .semibold),
            .foregroundColor: platformColor(config.secondaryColor)
        ]

        let valueAttrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 10, weight: .regular),
            .foregroundColor: platformColor(config.textColor)
        ]

        drawAttributedString(NSAttributedString(string: "Bank Details", attributes: labelAttrs),
                           in: context, at: CGPoint(x: margin, y: y))
        y += 14

        if let bankName = company.bankName {
            drawAttributedString(NSAttributedString(string: "Bank: \(bankName)", attributes: valueAttrs),
                               in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        if let bankAccount = company.bankAccount {
            drawAttributedString(NSAttributedString(string: "Account: \(bankAccount)", attributes: valueAttrs),
                               in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        if let bankSwift = company.bankSwift {
            drawAttributedString(NSAttributedString(string: "SWIFT: \(bankSwift)", attributes: valueAttrs),
                               in: context, at: CGPoint(x: margin, y: y))
            y += 14
        }

        return y + 20
    }

    private func drawWatermark(in context: CGContext, pageRect: CGRect, text: String) {
        let attrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 72, weight: .bold),
            .foregroundColor: CGColor(red: 0, green: 0.8, blue: 0, alpha: 0.15)
        ]

        let watermarkStr = NSAttributedString(string: text, attributes: attrs)
        let size = watermarkStr.size()

        // Center and rotate
        context.saveGState()
        context.translateBy(x: pageRect.width / 2, y: pageRect.height / 2)
        context.rotate(by: -.pi / 6) // -30 degrees

        let point = CGPoint(x: -size.width / 2, y: -size.height / 2)
        watermarkStr.draw(at: point)

        context.restoreGState()
    }

    private func drawPageNumber(in context: CGContext, pageRect: CGRect, pageNumber: Int) {
        let attrs: [NSAttributedString.Key: Any] = [
            .font: platformFont(size: 9, weight: .regular),
            .foregroundColor: platformColor(config.secondaryColor)
        ]

        let pageStr = NSAttributedString(string: "Page \(pageNumber)", attributes: attrs)
        let size = pageStr.size()

        let x = (pageRect.width - size.width) / 2
        let y = pageRect.height - margin / 2

        drawAttributedString(pageStr, in: context, at: CGPoint(x: x, y: y))
    }

    // MARK: - Helper Functions

    private func drawAttributedString(_ string: NSAttributedString, in context: CGContext, at point: CGPoint) {
        let line = CTLineCreateWithAttributedString(string)
        context.textPosition = point
        CTLineDraw(line, context)
    }

    private func platformFont(size: CGFloat, weight: Font.Weight) -> Any {
        #if os(iOS)
        switch weight {
        case .bold: return UIFont.boldSystemFont(ofSize: size)
        case .semibold: return UIFont.systemFont(ofSize: size, weight: .semibold)
        case .medium: return UIFont.systemFont(ofSize: size, weight: .medium)
        default: return UIFont.systemFont(ofSize: size)
        }
        #else
        switch weight {
        case .bold: return NSFont.boldSystemFont(ofSize: size)
        case .semibold: return NSFont.systemFont(ofSize: size, weight: .semibold)
        case .medium: return NSFont.systemFont(ofSize: size, weight: .medium)
        default: return NSFont.systemFont(ofSize: size)
        }
        #endif
    }

    private func platformColor(_ color: Color) -> PlatformColor {
        #if os(iOS)
        return UIColor(color)
        #else
        return NSColor(color)
        #endif
    }

    private func getCurrencySymbol(_ currency: String) -> String {
        let symbols: [String: String] = [
            "USD": "$",
            "EUR": "€",
            "GBP": "£",
            "JPY": "¥",
            "CAD": "CA$",
            "AUD": "A$",
            "CHF": "CHF",
            "CNY": "¥",
            "INR": "₹",
            "UAH": "₴",
            "PLN": "zł",
            "KRW": "₩",
            "BRL": "R$",
            "MXN": "MX$",
            "SGD": "S$",
            "HKD": "HK$",
            "NOK": "kr",
            "SEK": "kr",
            "DKK": "kr",
            "NZD": "NZ$",
            "ZAR": "R",
            "RUB": "₽",
            "TRY": "₺",
            "ILS": "₪",
            "THB": "฿",
            "PHP": "₱",
            "CZK": "Kč"
        ]
        return symbols[currency.uppercased()] ?? currency
    }
}

// MARK: - Platform Type Alias

#if os(iOS)
typealias PlatformColor = UIColor
#else
typealias PlatformColor = NSColor
#endif

// MARK: - CGColor Extension

extension CGColor {
    static func from(_ color: Color) -> CGColor? {
        #if os(iOS)
        return UIColor(color).cgColor
        #else
        return NSColor(color).cgColor
        #endif
    }
}
