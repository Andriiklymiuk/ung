//
//  CreateInvoiceSheet.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct CreateInvoiceSheet: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var selectedClientId: Int?
    @State private var isCreating = false
    @State private var showNoClientsAlert = false

    var body: some View {
        VStack(spacing: 0) {
            // Header
            headerSection

            Divider()

            // Content
            ScrollView {
                VStack(spacing: 16) {
                    // Info banner
                    infoBanner

                    // Client selection
                    clientSelectionSection

                    // Options hint
                    optionsHint
                }
                .padding(16)
            }

            Divider()

            // Footer actions
            footerSection
        }
        .frame(width: 320, height: 380)
        .background(Design.Colors.windowBackground)
        .onAppear {
            Task {
                await appState.refreshDashboard()
                if appState.clients.isEmpty {
                    showNoClientsAlert = true
                }
            }
        }
        .alert("No Clients", isPresented: $showNoClientsAlert) {
            Button("OK") { dismiss() }
        } message: {
            Text("You need to create a client first before generating an invoice.")
        }
    }

    // MARK: - Header
    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text("Create Invoice")
                    .font(Design.Typography.headingSmall)
                Text("Generate from tracked time")
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textSecondary)
            }

            Spacer()

            Button(action: { dismiss() }) {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 18))
                    .foregroundColor(Design.Colors.textTertiary)
            }
            .buttonStyle(DSInteractiveStyle())
        }
        .padding(Design.Spacing.md)
    }

    // MARK: - Info Banner
    private var infoBanner: some View {
        HStack(spacing: Design.Spacing.sm) {
            Image(systemName: "info.circle.fill")
                .font(.system(size: Design.IconSize.sm))
                .foregroundColor(Design.Colors.primary)

            Text("Invoice will be generated from unbilled time tracked for the selected client.")
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textSecondary)
        }
        .padding(Design.Spacing.sm)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: Design.Radius.sm)
                .fill(Design.Colors.primary.opacity(0.1))
        )
    }

    // MARK: - Client Selection
    private var clientSelectionSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            Text("Select Client")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textSecondary)

            if appState.clients.isEmpty {
                HStack(spacing: Design.Spacing.sm) {
                    Image(systemName: "person.fill.questionmark")
                        .font(.system(size: Design.IconSize.md))
                        .foregroundColor(Design.Colors.textTertiary)

                    VStack(alignment: .leading, spacing: 2) {
                        Text("No clients found")
                            .font(Design.Typography.labelMedium)
                        Text("Create a client to start invoicing")
                            .font(Design.Typography.labelSmall)
                            .foregroundColor(Design.Colors.textSecondary)
                    }
                }
                .padding(Design.Spacing.sm)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(
                    RoundedRectangle(cornerRadius: Design.Radius.sm)
                        .fill(Design.Colors.controlBackground)
                )
            } else {
                VStack(spacing: Design.Spacing.xxs) {
                    ForEach(appState.clients) { client in
                        InvoiceClientRow(
                            client: client,
                            isSelected: selectedClientId == client.id,
                            action: { withAnimation(Design.Animation.snappy) { selectedClientId = client.id } }
                        )
                    }
                }
            }
        }
    }

    // MARK: - Options Hint
    private var optionsHint: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            Text("Invoice includes:")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textSecondary)

            VStack(alignment: .leading, spacing: Design.Spacing.xs) {
                IncludeItem(icon: "clock.fill", text: "Unbilled time sessions")
                IncludeItem(icon: "dollarsign.circle.fill", text: "Contract rates applied")
                IncludeItem(icon: "calendar", text: "Current billing period")
            }
        }
        .padding(Design.Spacing.sm)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: Design.Radius.sm)
                .fill(Design.Colors.controlBackground)
        )
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack(spacing: Design.Spacing.sm) {
            Button("Cancel") {
                dismiss()
            }
            .buttonStyle(DSGhostButtonStyle())

            Spacer()

            Button(action: createInvoice) {
                HStack(spacing: 6) {
                    if isCreating {
                        ProgressView()
                            .scaleEffect(0.7)
                            .tint(.white)
                    } else {
                        Image(systemName: "doc.text.fill")
                            .font(.system(size: 10))
                    }
                    Text("Generate")
                }
            }
            .buttonStyle(DSCompactButtonStyle(isDisabled: selectedClientId == nil))
            .disabled(selectedClientId == nil || isCreating)
        }
        .padding(Design.Spacing.md)
    }

    // MARK: - Actions
    private func createInvoice() {
        guard let clientId = selectedClientId else { return }
        isCreating = true
        Task {
            // Get company for invoice
            guard let company = try? await appState.database.getCompany(),
                  let companyId = company?.id else {
                isCreating = false
                return
            }

            // Generate invoice number
            let year = Calendar.current.component(.year, from: Date())
            let count = (try? await appState.database.getInvoiceCount()) ?? 0
            let invoiceNum = "INV-\(year)-\(String(format: "%04d", count + 1))"

            // Create invoice
            var invoice = Invoice(
                invoiceNum: invoiceNum,
                companyId: companyId,
                amount: 0,
                currency: "USD",
                status: "pending"
            )
            invoice.issuedDate = Date()
            invoice.dueDate = Calendar.current.date(byAdding: .day, value: 30, to: Date())

            _ = try? await appState.database.createInvoice(invoice)
            await appState.refreshDashboard()
            isCreating = false
            dismiss()
        }
    }
}

// MARK: - Invoice Client Row
struct InvoiceClientRow: View {
    let client: Client
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: Design.Spacing.sm) {
                // Selection indicator
                Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: Design.IconSize.sm))
                    .foregroundColor(isSelected ? Design.Colors.primary : Design.Colors.textTertiary)

                // Avatar
                ZStack {
                    Circle()
                        .fill(Design.Colors.primary.opacity(0.15))
                        .frame(width: 32, height: 32)

                    Text(String(client.name.prefix(1)).uppercased())
                        .font(Design.Typography.labelMedium)
                        .foregroundColor(Design.Colors.primary)
                }

                // Name
                Text(client.name)
                    .font(Design.Typography.labelLarge)
                    .foregroundColor(Design.Colors.textPrimary)

                Spacer()
            }
            .padding(Design.Spacing.sm)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.sm)
                    .fill(isSelected ? Design.Colors.primary.opacity(0.1) : Design.Colors.controlBackground)
                    .overlay(
                        RoundedRectangle(cornerRadius: Design.Radius.sm)
                            .stroke(isSelected ? Design.Colors.primary.opacity(0.5) : Color.clear, lineWidth: 1)
                    )
            )
        }
        .buttonStyle(DSInteractiveStyle())
    }
}

// MARK: - Include Item
struct IncludeItem: View {
    let icon: String
    let text: String

    var body: some View {
        HStack(spacing: Design.Spacing.xs) {
            Image(systemName: icon)
                .font(.system(size: 10))
                .foregroundColor(Design.Colors.success)
                .frame(width: Design.Spacing.md)

            Text(text)
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textPrimary)
        }
    }
}

#Preview {
    CreateInvoiceSheet()
        .environmentObject(AppState())
}
