//
//  TemplateEditorView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI
#if os(iOS)
import UIKit
#endif

// MARK: - Template Models
struct InvoiceTemplate: Codable, Identifiable {
    var id = UUID()
    var name: String = "Default Template"
    var blocks: [TemplateBlock] = TemplateBlock.defaultBlocks
    var styling: TemplateStyling = TemplateStyling()
}

struct TemplateBlock: Codable, Identifiable, Equatable {
    var id = UUID()
    var type: BlockType
    var isVisible: Bool = true
    var customContent: String?

    enum BlockType: String, Codable, CaseIterable {
        case header = "Header"
        case companyInfo = "Company Info"
        case clientInfo = "Client Info"
        case invoiceDetails = "Invoice Details"
        case lineItems = "Line Items"
        case subtotal = "Subtotal"
        case tax = "Tax"
        case total = "Total"
        case notes = "Notes"
        case paymentTerms = "Payment Terms"
        case bankDetails = "Bank Details"
        case footer = "Footer"

        var icon: String {
            switch self {
            case .header: return "text.alignleft"
            case .companyInfo: return "building.2"
            case .clientInfo: return "person.fill"
            case .invoiceDetails: return "doc.text"
            case .lineItems: return "list.bullet.rectangle"
            case .subtotal: return "sum"
            case .tax: return "percent"
            case .total: return "dollarsign.circle"
            case .notes: return "note.text"
            case .paymentTerms: return "calendar.badge.clock"
            case .bankDetails: return "creditcard"
            case .footer: return "text.aligncenter"
            }
        }

        var description: String {
            switch self {
            case .header: return "Your logo and invoice title"
            case .companyInfo: return "Your business details"
            case .clientInfo: return "Client billing information"
            case .invoiceDetails: return "Invoice number, date, due date"
            case .lineItems: return "Services and products"
            case .subtotal: return "Sum before tax"
            case .tax: return "Tax calculations"
            case .total: return "Final amount due"
            case .notes: return "Additional notes"
            case .paymentTerms: return "Payment conditions"
            case .bankDetails: return "Bank account information"
            case .footer: return "Thank you message"
            }
        }
    }

    static var defaultBlocks: [TemplateBlock] {
        [
            TemplateBlock(type: .header),
            TemplateBlock(type: .companyInfo),
            TemplateBlock(type: .clientInfo),
            TemplateBlock(type: .invoiceDetails),
            TemplateBlock(type: .lineItems),
            TemplateBlock(type: .subtotal),
            TemplateBlock(type: .tax),
            TemplateBlock(type: .total),
            TemplateBlock(type: .notes),
            TemplateBlock(type: .paymentTerms),
            TemplateBlock(type: .bankDetails),
            TemplateBlock(type: .footer)
        ]
    }
}

struct TemplateStyling: Codable {
    var primaryColor: String = "#2563EB"
    var secondaryColor: String = "#64748B"
    var fontFamily: String = "Helvetica"
    var fontSize: Int = 12
    var headerSize: Int = 24
    var pageMargin: Int = 40
    var showLogo: Bool = true
    var logoPosition: LogoPosition = .left

    enum LogoPosition: String, Codable, CaseIterable {
        case left = "Left"
        case center = "Center"
        case right = "Right"
    }
}

// MARK: - Template Editor View
struct TemplateEditorView: View {
    @Environment(\.dismiss) var dismiss
    @State private var template = InvoiceTemplate()
    @State private var selectedBlockId: UUID?
    @State private var isDragging = false
    @State private var showSaveSuccess = false

    var body: some View {
        HStack(spacing: 0) {
            // Left panel - Block list and settings
            leftPanel
                .frame(width: 300)

            Divider()

            // Center - Preview
            previewPanel
                .frame(maxWidth: .infinity)

            Divider()

            // Right panel - Block settings
            rightPanel
                .frame(width: 280)
        }
        .frame(minWidth: 1000, minHeight: 700)
        .alert("Template Saved", isPresented: $showSaveSuccess) {
            Button("OK") {}
        } message: {
            Text("Your template has been saved successfully.")
        }
    }

    // MARK: - Left Panel
    private var leftPanel: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                VStack(alignment: .leading, spacing: 2) {
                    Text("Template Editor")
                        .font(.system(size: 16, weight: .bold))
                    Text("Drag to reorder blocks")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }
                Spacer()
            }
            .padding(16)
            .background(Design.Colors.windowBackground)

            Divider()

            // Blocks list
            ScrollView {
                VStack(spacing: 8) {
                    ForEach(Array(template.blocks.enumerated()), id: \.element.id) { index, block in
                        BlockRow(
                            block: block,
                            isSelected: selectedBlockId == block.id,
                            onSelect: { selectedBlockId = block.id },
                            onToggle: { template.blocks[index].isVisible.toggle() },
                            onMoveUp: { moveBlock(at: index, direction: -1) },
                            onMoveDown: { moveBlock(at: index, direction: 1) },
                            canMoveUp: index > 0,
                            canMoveDown: index < template.blocks.count - 1
                        )
                    }
                }
                .padding(12)
            }

            Divider()

            // Actions
            HStack(spacing: 12) {
                Button("Reset") {
                    template = InvoiceTemplate()
                }
                .buttonStyle(.bordered)

                Spacer()

                Button("Save Template") {
                    saveTemplate()
                }
                .buttonStyle(.borderedProminent)
            }
            .padding(16)
        }
        .background(Design.Colors.controlBackground.opacity(0.5))
    }

    // MARK: - Preview Panel
    private var previewPanel: some View {
        VStack(spacing: 0) {
            // Preview header
            HStack {
                Text("Preview")
                    .font(.system(size: 14, weight: .semibold))

                Spacer()

                // Zoom controls would go here
            }
            .padding(16)
            .background(Design.Colors.windowBackground)

            Divider()

            // Invoice preview
            ScrollView {
                invoicePreview
                    .padding(40)
            }
            .background(Color.gray.opacity(0.1))
        }
    }

    private var invoicePreview: some View {
        VStack(spacing: 0) {
            // Paper
            VStack(alignment: .leading, spacing: 20) {
                ForEach(template.blocks.filter { $0.isVisible }) { block in
                    previewBlock(block)
                }
            }
            .padding(CGFloat(template.styling.pageMargin))
            .frame(width: 595) // A4 width at 72 DPI
            .background(Color.white)
            .shadow(color: .black.opacity(0.15), radius: 10, y: 5)
        }
    }

    @ViewBuilder
    private func previewBlock(_ block: TemplateBlock) -> some View {
        switch block.type {
        case .header:
            HStack {
                if template.styling.logoPosition == .left && template.styling.showLogo {
                    logoPlaceholder
                }
                Spacer()
                if template.styling.logoPosition == .center && template.styling.showLogo {
                    logoPlaceholder
                }
                Spacer()
                VStack(alignment: .trailing) {
                    Text("INVOICE")
                        .font(.system(size: CGFloat(template.styling.headerSize), weight: .bold))
                        .foregroundColor(Color(hex: template.styling.primaryColor))
                }
                if template.styling.logoPosition == .right && template.styling.showLogo {
                    logoPlaceholder
                }
            }

        case .companyInfo:
            VStack(alignment: .leading, spacing: 4) {
                Text("Your Company Name")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(Color(hex: template.styling.primaryColor))
                Text("123 Business Street")
                Text("City, State 12345")
                Text("contact@company.com")
            }
            .font(.system(size: CGFloat(template.styling.fontSize)))
            .foregroundColor(Color(hex: template.styling.secondaryColor))

        case .clientInfo:
            VStack(alignment: .leading, spacing: 4) {
                Text("BILL TO")
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(Color(hex: template.styling.secondaryColor))
                Text("Client Name")
                    .font(.system(size: 14, weight: .medium))
                Text("Client Address")
                Text("client@email.com")
            }
            .font(.system(size: CGFloat(template.styling.fontSize)))

        case .invoiceDetails:
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    DetailRow(label: "Invoice #", value: "INV-001")
                    DetailRow(label: "Date", value: "Nov 28, 2025")
                    DetailRow(label: "Due Date", value: "Dec 28, 2025")
                }
                Spacer()
            }

        case .lineItems:
            VStack(spacing: 0) {
                // Header
                HStack {
                    Text("Description")
                        .frame(maxWidth: .infinity, alignment: .leading)
                    Text("Qty")
                        .frame(width: 50)
                    Text("Rate")
                        .frame(width: 80)
                    Text("Amount")
                        .frame(width: 80, alignment: .trailing)
                }
                .font(.system(size: 10, weight: .semibold))
                .foregroundColor(Color(hex: template.styling.secondaryColor))
                .padding(.vertical, 8)
                .background(Color.gray.opacity(0.1))

                // Sample items
                ForEach(0..<3) { i in
                    HStack {
                        Text("Service Item \(i + 1)")
                            .frame(maxWidth: .infinity, alignment: .leading)
                        Text("\(i + 1)")
                            .frame(width: 50)
                        Text("$100.00")
                            .frame(width: 80)
                        Text("$\((i + 1) * 100).00")
                            .frame(width: 80, alignment: .trailing)
                    }
                    .font(.system(size: CGFloat(template.styling.fontSize)))
                    .padding(.vertical, 8)
                    Divider()
                }
            }

        case .subtotal:
            HStack {
                Spacer()
                Text("Subtotal")
                    .foregroundColor(Color(hex: template.styling.secondaryColor))
                Text("$600.00")
                    .frame(width: 80, alignment: .trailing)
                    .fontWeight(.medium)
            }
            .font(.system(size: CGFloat(template.styling.fontSize)))

        case .tax:
            HStack {
                Spacer()
                Text("Tax (10%)")
                    .foregroundColor(Color(hex: template.styling.secondaryColor))
                Text("$60.00")
                    .frame(width: 80, alignment: .trailing)
            }
            .font(.system(size: CGFloat(template.styling.fontSize)))

        case .total:
            HStack {
                Spacer()
                Text("TOTAL")
                    .font(.system(size: 14, weight: .bold))
                    .foregroundColor(Color(hex: template.styling.primaryColor))
                Text("$660.00")
                    .font(.system(size: 18, weight: .bold))
                    .foregroundColor(Color(hex: template.styling.primaryColor))
                    .frame(width: 100, alignment: .trailing)
            }
            .padding(.top, 8)

        case .notes:
            VStack(alignment: .leading, spacing: 4) {
                Text("Notes")
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(Color(hex: template.styling.secondaryColor))
                Text(block.customContent ?? "Thank you for your business. Payment is due within 30 days.")
                    .font(.system(size: CGFloat(template.styling.fontSize)))
            }

        case .paymentTerms:
            VStack(alignment: .leading, spacing: 4) {
                Text("Payment Terms")
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(Color(hex: template.styling.secondaryColor))
                Text("Net 30 days. Late payments subject to 1.5% monthly interest.")
                    .font(.system(size: CGFloat(template.styling.fontSize)))
            }

        case .bankDetails:
            VStack(alignment: .leading, spacing: 4) {
                Text("Bank Details")
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(Color(hex: template.styling.secondaryColor))
                Text("Bank: Example Bank")
                Text("Account: 1234567890")
                Text("Routing: 987654321")
            }
            .font(.system(size: CGFloat(template.styling.fontSize)))

        case .footer:
            Text(block.customContent ?? "Thank you for your business!")
                .font(.system(size: CGFloat(template.styling.fontSize)))
                .foregroundColor(Color(hex: template.styling.secondaryColor))
                .frame(maxWidth: .infinity, alignment: .center)
                .padding(.top, 20)
        }
    }

    private var logoPlaceholder: some View {
        RoundedRectangle(cornerRadius: 8)
            .fill(Color.gray.opacity(0.2))
            .frame(width: 80, height: 40)
            .overlay(
                Text("LOGO")
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(.gray)
            )
    }

    // MARK: - Right Panel
    private var rightPanel: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("Settings")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
            }
            .padding(16)
            .background(Design.Colors.windowBackground)

            Divider()

            ScrollView {
                VStack(spacing: 20) {
                    // Template name
                    SettingSection(title: "Template Name") {
                        TextField("Name", text: $template.name)
                            .textFieldStyle(.roundedBorder)
                    }

                    // Colors
                    SettingSection(title: "Colors") {
                        HStack {
                            Text("Primary")
                                .font(.system(size: 12))
                            Spacer()
                            ColorPicker("", selection: Binding(
                                get: { Color(hex: template.styling.primaryColor) },
                                set: { template.styling.primaryColor = $0.hexString }
                            ))
                        }

                        HStack {
                            Text("Secondary")
                                .font(.system(size: 12))
                            Spacer()
                            ColorPicker("", selection: Binding(
                                get: { Color(hex: template.styling.secondaryColor) },
                                set: { template.styling.secondaryColor = $0.hexString }
                            ))
                        }
                    }

                    // Typography
                    SettingSection(title: "Typography") {
                        HStack {
                            Text("Font Size")
                                .font(.system(size: 12))
                            Spacer()
                            Stepper("\(template.styling.fontSize)pt", value: $template.styling.fontSize, in: 10...16)
                        }

                        HStack {
                            Text("Header Size")
                                .font(.system(size: 12))
                            Spacer()
                            Stepper("\(template.styling.headerSize)pt", value: $template.styling.headerSize, in: 18...36)
                        }
                    }

                    // Layout
                    SettingSection(title: "Layout") {
                        HStack {
                            Text("Page Margin")
                                .font(.system(size: 12))
                            Spacer()
                            Stepper("\(template.styling.pageMargin)px", value: $template.styling.pageMargin, in: 20...60, step: 5)
                        }

                        Toggle("Show Logo", isOn: $template.styling.showLogo)
                            .font(.system(size: 12))

                        if template.styling.showLogo {
                            Picker("Logo Position", selection: $template.styling.logoPosition) {
                                ForEach(TemplateStyling.LogoPosition.allCases, id: \.self) { position in
                                    Text(position.rawValue).tag(position)
                                }
                            }
                            .pickerStyle(.segmented)
                        }
                    }

                    // Block settings
                    if let selectedId = selectedBlockId,
                       let blockIndex = template.blocks.firstIndex(where: { $0.id == selectedId }) {
                        SettingSection(title: "Block Settings") {
                            let block = template.blocks[blockIndex]

                            if block.type == .notes || block.type == .footer {
                                Text("Custom Text")
                                    .font(.system(size: 12))
                                TextEditor(text: Binding(
                                    get: { template.blocks[blockIndex].customContent ?? "" },
                                    set: { template.blocks[blockIndex].customContent = $0 }
                                ))
                                .frame(height: 80)
                                .border(Color.gray.opacity(0.3))
                            }
                        }
                    }
                }
                .padding(16)
            }
        }
        .background(Design.Colors.controlBackground.opacity(0.5))
    }

    // MARK: - Actions
    private func moveBlock(at index: Int, direction: Int) {
        let newIndex = index + direction
        guard newIndex >= 0 && newIndex < template.blocks.count else { return }
        template.blocks.swapAt(index, newIndex)
    }

    private func saveTemplate() {
        // In a real app, save to disk or CLI
        showSaveSuccess = true
    }
}

// MARK: - Supporting Views
struct BlockRow: View {
    let block: TemplateBlock
    let isSelected: Bool
    let onSelect: () -> Void
    let onToggle: () -> Void
    let onMoveUp: () -> Void
    let onMoveDown: () -> Void
    let canMoveUp: Bool
    let canMoveDown: Bool
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        HStack(spacing: 10) {
            // Drag handle
            VStack(spacing: 2) {
                Button(action: onMoveUp) {
                    Image(systemName: "chevron.up")
                        .font(.system(size: 10))
                }
                .buttonStyle(.plain)
                .disabled(!canMoveUp)
                .opacity(canMoveUp ? 1 : 0.3)

                Button(action: onMoveDown) {
                    Image(systemName: "chevron.down")
                        .font(.system(size: 10))
                }
                .buttonStyle(.plain)
                .disabled(!canMoveDown)
                .opacity(canMoveDown ? 1 : 0.3)
            }
            .foregroundColor(.secondary)

            // Icon
            Image(systemName: block.type.icon)
                .font(.system(size: 12))
                .foregroundColor(isSelected ? .blue : .secondary)
                .frame(width: 24)

            // Content
            VStack(alignment: .leading, spacing: 2) {
                Text(block.type.rawValue)
                    .font(.system(size: 12, weight: isSelected ? .semibold : .regular))
                    .foregroundColor(block.isVisible ? .primary : .secondary)

                Text(block.type.description)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
                    .lineLimit(1)
            }

            Spacer()

            // Visibility toggle
            Button(action: onToggle) {
                Image(systemName: block.isVisible ? "eye.fill" : "eye.slash.fill")
                    .font(.system(size: 12))
                    .foregroundColor(block.isVisible ? .blue : .secondary)
            }
            .buttonStyle(.plain)
        }
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(isSelected
                      ? (colorScheme == .dark ? Color.blue.opacity(0.2) : Color.blue.opacity(0.1))
                      : (colorScheme == .dark ? Color(white: 0.15) : Color.white))
                .shadow(color: .black.opacity(0.05), radius: 2, y: 1)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 8)
                .stroke(isSelected ? Color.blue.opacity(0.5) : Color.clear, lineWidth: 1)
        )
        .opacity(block.isVisible ? 1 : 0.6)
        .onTapGesture { onSelect() }
    }
}

struct DetailRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .foregroundColor(.secondary)
            Text(value)
                .fontWeight(.medium)
        }
        .font(.system(size: 11))
    }
}

struct SettingSection<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title)
                .font(.system(size: 11, weight: .semibold))
                .foregroundColor(.secondary)
                .textCase(.uppercase)

            content()
        }
    }
}

// MARK: - Color Extensions
extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3:
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6:
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8:
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (255, 0, 0, 0)
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }

    var hexString: String {
        #if os(macOS)
        guard let components = NSColor(self).cgColor.components else { return "#000000" }
        #else
        guard let components = UIColor(self).cgColor.components else { return "#000000" }
        #endif
        let r = Int(components[0] * 255)
        let g = Int(components[1] * 255)
        let b = Int(components[2] * 255)
        return String(format: "#%02X%02X%02X", r, g, b)
    }
}

#Preview {
    TemplateEditorView()
        .frame(width: 1200, height: 800)
}
