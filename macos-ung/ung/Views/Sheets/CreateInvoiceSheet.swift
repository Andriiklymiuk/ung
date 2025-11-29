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
                    .font(.system(size: 14, weight: .semibold))
                Text("Generate from tracked time")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Button(action: { dismiss() }) {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 18))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
        }
        .padding(16)
    }

    // MARK: - Info Banner
    private var infoBanner: some View {
        HStack(spacing: 10) {
            Image(systemName: "info.circle.fill")
                .font(.system(size: 14))
                .foregroundColor(.blue)

            Text("Invoice will be generated from unbilled time tracked for the selected client.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
        .padding(12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color.blue.opacity(0.1))
        )
    }

    // MARK: - Client Selection
    private var clientSelectionSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Select Client")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            if appState.clients.isEmpty {
                HStack(spacing: 10) {
                    Image(systemName: "person.fill.questionmark")
                        .font(.system(size: 16))
                        .foregroundColor(.secondary)

                    VStack(alignment: .leading, spacing: 2) {
                        Text("No clients found")
                            .font(.system(size: 12, weight: .medium))
                        Text("Create a client to start invoicing")
                            .font(.system(size: 10))
                            .foregroundColor(.secondary)
                    }
                }
                .padding(12)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Design.Colors.controlBackground)
                )
            } else {
                VStack(spacing: 4) {
                    ForEach(appState.clients) { client in
                        InvoiceClientRow(
                            client: client,
                            isSelected: selectedClientId == client.id,
                            action: { selectedClientId = client.id }
                        )
                    }
                }
            }
        }
    }

    // MARK: - Options Hint
    private var optionsHint: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Invoice includes:")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            VStack(alignment: .leading, spacing: 6) {
                IncludeItem(icon: "clock.fill", text: "Unbilled time sessions")
                IncludeItem(icon: "dollarsign.circle.fill", text: "Contract rates applied")
                IncludeItem(icon: "calendar", text: "Current billing period")
            }
        }
        .padding(12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Design.Colors.controlBackground)
        )
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack(spacing: 12) {
            Button("Cancel") {
                dismiss()
            }
            .buttonStyle(.plain)
            .foregroundColor(.secondary)

            Spacer()

            Button(action: createInvoice) {
                HStack(spacing: 6) {
                    if isCreating {
                        ProgressView()
                            .scaleEffect(0.7)
                    } else {
                        Image(systemName: "doc.text.fill")
                            .font(.system(size: 10))
                    }
                    Text("Generate")
                        .font(.system(size: 12, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(selectedClientId != nil ? Color.blue : Color.gray)
                )
            }
            .buttonStyle(.plain)
            .disabled(selectedClientId == nil || isCreating)
        }
        .padding(16)
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
            HStack(spacing: 10) {
                // Selection indicator
                Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 14))
                    .foregroundColor(isSelected ? .blue : .secondary)

                // Avatar
                ZStack {
                    Circle()
                        .fill(Color.blue.opacity(0.15))
                        .frame(width: 32, height: 32)

                    Text(String(client.name.prefix(1)).uppercased())
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.blue)
                }

                // Name
                Text(client.name)
                    .font(.system(size: 13, weight: .medium))
                    .foregroundColor(.primary)

                Spacer()
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(isSelected ? Color.blue.opacity(0.1) : Design.Colors.controlBackground)
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(isSelected ? Color.blue.opacity(0.5) : Color.clear, lineWidth: 1)
                    )
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Include Item
struct IncludeItem: View {
    let icon: String
    let text: String

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: icon)
                .font(.system(size: 10))
                .foregroundColor(.green)
                .frame(width: 16)

            Text(text)
                .font(.system(size: 11))
                .foregroundColor(.primary)
        }
    }
}

#Preview {
    CreateInvoiceSheet()
        .environmentObject(AppState())
}
