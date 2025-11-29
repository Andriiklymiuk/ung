//
//  RecurringInvoicesSheet.swift
//  ung
//
//  Manage recurring invoices - create, view, pause, generate
//

import SwiftUI

struct RecurringInvoicesSheet: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var recurringInvoices: [RecurringInvoice] = []
    @State private var isLoading = true
    @State private var showAddSheet = false
    @State private var isGenerating = false
    @State private var generatedCount = 0
    @State private var showGenerateSuccess = false

    var body: some View {
        VStack(spacing: 0) {
            // Header
            headerSection

            Divider()

            // Content
            if isLoading {
                Spacer()
                ProgressView()
                    .scaleEffect(0.8)
                Spacer()
            } else if recurringInvoices.isEmpty {
                Spacer()
                emptyState
                Spacer()
            } else {
                ScrollView {
                    LazyVStack(spacing: 8) {
                        ForEach(recurringInvoices) { recurring in
                            RecurringInvoiceRow(
                                recurring: recurring,
                                clientName: clientName(for: recurring.clientId),
                                onPause: { pauseRecurring(recurring) },
                                onResume: { resumeRecurring(recurring) },
                                onDelete: { deleteRecurring(recurring) },
                                onGenerate: { generateSingle(recurring) }
                            )
                        }
                    }
                    .padding(16)
                }
            }

            Divider()

            // Footer
            footerSection
        }
        .frame(width: 500, height: 500)
        .background(Design.Colors.windowBackground)
        .onAppear { loadRecurringInvoices() }
        .sheet(isPresented: $showAddSheet) {
            AddRecurringInvoiceSheet(onSave: {
                loadRecurringInvoices()
            })
            .environmentObject(appState)
        }
        .alert("Invoices Generated", isPresented: $showGenerateSuccess) {
            Button("OK") {}
        } message: {
            Text("\(generatedCount) invoice(s) were generated from recurring templates.")
        }
    }

    // MARK: - Header
    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text("Recurring Invoices")
                    .font(.system(size: 14, weight: .semibold))
                Text("Automate your invoicing")
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

    // MARK: - Empty State
    private var emptyState: some View {
        VStack(spacing: 12) {
            Image(systemName: "arrow.triangle.2.circlepath.doc.on.clipboard")
                .font(.system(size: 40))
                .foregroundColor(.secondary.opacity(0.5))

            Text("No Recurring Invoices")
                .font(.system(size: 14, weight: .medium))

            Text("Set up recurring invoices to automatically generate invoices on a schedule.")
                .font(.system(size: 12))
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 40)

            Button(action: { showAddSheet = true }) {
                HStack(spacing: 4) {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 12))
                    Text("Create Recurring Invoice")
                        .font(.system(size: 12, weight: .medium))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color.blue)
                )
            }
            .buttonStyle(.plain)
            .padding(.top, 8)
        }
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack(spacing: 12) {
            Button(action: { showAddSheet = true }) {
                HStack(spacing: 4) {
                    Image(systemName: "plus")
                        .font(.system(size: 10))
                    Text("Add New")
                        .font(.system(size: 12, weight: .medium))
                }
            }
            .buttonStyle(.bordered)

            Spacer()

            // Generate due invoices button
            Button(action: generateDueInvoices) {
                HStack(spacing: 6) {
                    if isGenerating {
                        ProgressView()
                            .scaleEffect(0.6)
                    } else {
                        Image(systemName: "bolt.fill")
                            .font(.system(size: 10))
                    }
                    Text("Generate Due")
                        .font(.system(size: 12, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(hasDueInvoices && !isGenerating ? Color.green : Color.gray)
                )
            }
            .buttonStyle(.plain)
            .disabled(!hasDueInvoices || isGenerating)
        }
        .padding(16)
    }

    // MARK: - Computed Properties
    private var hasDueInvoices: Bool {
        recurringInvoices.contains { $0.active && $0.nextGenerationDate <= Date() }
    }

    private func clientName(for clientId: Int64) -> String {
        appState.clients.first { $0.id == Int(clientId) }?.name ?? "Unknown"
    }

    // MARK: - Actions
    private func loadRecurringInvoices() {
        Task {
            isLoading = true
            do {
                recurringInvoices = try await appState.database.getRecurringInvoices()
            } catch {
                print("Failed to load recurring invoices: \(error)")
            }
            isLoading = false
        }
    }

    private func pauseRecurring(_ recurring: RecurringInvoice) {
        guard let id = recurring.id else { return }
        Task {
            try? await appState.database.pauseRecurringInvoice(id: id)
            loadRecurringInvoices()
        }
    }

    private func resumeRecurring(_ recurring: RecurringInvoice) {
        guard let id = recurring.id else { return }
        Task {
            try? await appState.database.resumeRecurringInvoice(id: id)
            loadRecurringInvoices()
        }
    }

    private func deleteRecurring(_ recurring: RecurringInvoice) {
        guard let id = recurring.id else { return }
        Task {
            try? await appState.database.deleteRecurringInvoice(id: id)
            loadRecurringInvoices()
        }
    }

    private func generateSingle(_ recurring: RecurringInvoice) {
        Task {
            if let _ = try? await appState.database.generateInvoiceFromRecurring(recurring) {
                generatedCount = 1
                showGenerateSuccess = true
                await appState.refreshDashboard()
                loadRecurringInvoices()
            }
        }
    }

    private func generateDueInvoices() {
        isGenerating = true
        Task {
            do {
                let invoices = try await appState.database.generateDueRecurringInvoices()
                generatedCount = invoices.count
                if generatedCount > 0 {
                    showGenerateSuccess = true
                    await appState.refreshDashboard()
                }
                loadRecurringInvoices()
            } catch {
                print("Failed to generate invoices: \(error)")
            }
            isGenerating = false
        }
    }
}

// MARK: - Recurring Invoice Row
struct RecurringInvoiceRow: View {
    let recurring: RecurringInvoice
    let clientName: String
    let onPause: () -> Void
    let onResume: () -> Void
    let onDelete: () -> Void
    let onGenerate: () -> Void

    @State private var isHovered = false
    @Environment(\.colorScheme) var colorScheme

    private var isDue: Bool {
        recurring.nextGenerationDate <= Date()
    }

    private var dateFormatter: DateFormatter {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        return formatter
    }

    var body: some View {
        HStack(spacing: 12) {
            // Status indicator
            Circle()
                .fill(recurring.active ? (isDue ? Color.green : Color.blue) : Color.gray)
                .frame(width: 8, height: 8)

            // Info
            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Text(clientName)
                        .font(.system(size: 13, weight: .medium))

                    Text("•")
                        .foregroundColor(.secondary)

                    Text(recurring.frequencyType.displayName)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                HStack(spacing: 8) {
                    Text(formatCurrency(recurring.amount, recurring.currency))
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.primary)

                    if let description = recurring.description, !description.isEmpty {
                        Text(description)
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                            .lineLimit(1)
                    }
                }

                Text("Next: \(dateFormatter.string(from: recurring.nextGenerationDate))")
                    .font(.system(size: 10))
                    .foregroundColor(isDue ? .green : .secondary)
            }

            Spacer()

            // Status badge
            if !recurring.active {
                Text("Paused")
                    .font(.system(size: 9, weight: .medium))
                    .foregroundColor(.orange)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.orange.opacity(0.15))
                    .cornerRadius(4)
            } else if isDue {
                Text("Due")
                    .font(.system(size: 9, weight: .medium))
                    .foregroundColor(.green)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.green.opacity(0.15))
                    .cornerRadius(4)
            }

            // Generated count
            HStack(spacing: 2) {
                Image(systemName: "doc.on.doc")
                    .font(.system(size: 10))
                Text("\(recurring.generatedCount)")
                    .font(.system(size: 10, weight: .medium))
            }
            .foregroundColor(.secondary)

            // Action buttons (on hover)
            if isHovered {
                HStack(spacing: 4) {
                    if isDue && recurring.active {
                        Button(action: onGenerate) {
                            Image(systemName: "bolt.fill")
                                .font(.system(size: 12))
                                .foregroundColor(.green)
                        }
                        .buttonStyle(.plain)
                        .help("Generate Now")
                    }

                    if recurring.active {
                        Button(action: onPause) {
                            Image(systemName: "pause.circle")
                                .font(.system(size: 12))
                                .foregroundColor(.orange)
                        }
                        .buttonStyle(.plain)
                        .help("Pause")
                    } else {
                        Button(action: onResume) {
                            Image(systemName: "play.circle")
                                .font(.system(size: 12))
                                .foregroundColor(.green)
                        }
                        .buttonStyle(.plain)
                        .help("Resume")
                    }

                    Button(action: onDelete) {
                        Image(systemName: "trash")
                            .font(.system(size: 12))
                            .foregroundColor(.red)
                    }
                    .buttonStyle(.plain)
                    .help("Delete")
                }
                .transition(.scale.combined(with: .opacity))
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Design.Colors.controlBackground)
        )
        .animation(.easeInOut(duration: 0.2), value: isHovered)
        .onHover { hovering in isHovered = hovering }
    }

    private func formatCurrency(_ amount: Double, _ currency: String) -> String {
        let symbols: [String: String] = [
            "USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥",
            "CAD": "CA$", "AUD": "A$", "CHF": "CHF", "UAH": "₴"
        ]
        let symbol = symbols[currency.uppercased()] ?? currency
        return "\(symbol)\(String(format: "%.2f", amount))"
    }
}

// MARK: - Add Recurring Invoice Sheet
struct AddRecurringInvoiceSheet: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    var onSave: () -> Void

    @State private var selectedClientId: Int?
    @State private var amount: String = ""
    @State private var currency: String = "USD"
    @State private var description: String = ""
    @State private var frequency: RecurringFrequency = .monthly
    @State private var dayOfMonth: Int = 1
    @State private var autoPdf: Bool = true
    @State private var isSaving = false

    let currencies = ["USD", "EUR", "GBP", "CHF", "CAD", "AUD", "JPY", "UAH"]
    let daysOfMonth = [1, 5, 10, 15, 20, 25, 28]

    var isValid: Bool {
        selectedClientId != nil && (Double(amount) ?? 0) > 0
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("New Recurring Invoice")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
                Button(action: { dismiss() }) {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 18))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
            }
            .padding(16)

            Divider()

            // Form
            ScrollView {
                VStack(spacing: 16) {
                    // Client selection
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Client")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(.secondary)

                        if appState.clients.isEmpty {
                            Text("No clients available")
                                .font(.system(size: 12))
                                .foregroundColor(.secondary)
                        } else {
                            Picker("Client", selection: $selectedClientId) {
                                Text("Select a client").tag(nil as Int?)
                                ForEach(appState.clients) { client in
                                    Text(client.name).tag(client.id as Int?)
                                }
                            }
                            .pickerStyle(.menu)
                        }
                    }

                    // Amount and currency
                    HStack(spacing: 12) {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Amount")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(.secondary)

                            TextField("0.00", text: $amount)
                                .textFieldStyle(.roundedBorder)
                        }

                        VStack(alignment: .leading, spacing: 8) {
                            Text("Currency")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(.secondary)

                            Picker("Currency", selection: $currency) {
                                ForEach(currencies, id: \.self) { curr in
                                    Text(curr).tag(curr)
                                }
                            }
                            .pickerStyle(.menu)
                            .frame(width: 100)
                        }
                    }

                    // Description
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Description")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(.secondary)

                        TextField("Service description", text: $description)
                            .textFieldStyle(.roundedBorder)
                    }

                    // Frequency
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Frequency")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(.secondary)

                        Picker("Frequency", selection: $frequency) {
                            ForEach(RecurringFrequency.allCases, id: \.self) { freq in
                                Text(freq.displayName).tag(freq)
                            }
                        }
                        .pickerStyle(.segmented)
                    }

                    // Day of month (for monthly/quarterly/yearly)
                    if frequency == .monthly || frequency == .quarterly || frequency == .yearly {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Day of Month")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(.secondary)

                            Picker("Day", selection: $dayOfMonth) {
                                ForEach(daysOfMonth, id: \.self) { day in
                                    Text("\(day)").tag(day)
                                }
                            }
                            .pickerStyle(.segmented)
                        }
                    }

                    // Auto PDF toggle
                    Toggle("Automatically generate PDF", isOn: $autoPdf)
                        .font(.system(size: 12))
                }
                .padding(16)
            }

            Divider()

            // Footer
            HStack {
                Button("Cancel") { dismiss() }
                    .buttonStyle(.plain)
                    .foregroundColor(.secondary)

                Spacer()

                Button(action: saveRecurring) {
                    HStack(spacing: 6) {
                        if isSaving {
                            ProgressView()
                                .scaleEffect(0.6)
                        } else {
                            Image(systemName: "plus.circle.fill")
                                .font(.system(size: 10))
                        }
                        Text("Create")
                            .font(.system(size: 12, weight: .semibold))
                    }
                    .foregroundColor(.white)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(isValid ? Color.blue : Color.gray)
                    )
                }
                .buttonStyle(.plain)
                .disabled(!isValid || isSaving)
            }
            .padding(16)
        }
        .frame(width: 400, height: 500)
        .background(Design.Colors.windowBackground)
    }

    private func saveRecurring() {
        guard let clientId = selectedClientId,
              let amountValue = Double(amount.replacingOccurrences(of: ",", with: ".")) else { return }

        isSaving = true
        Task {
            let recurring = RecurringInvoice(
                clientId: Int64(clientId),
                contractId: nil,
                amount: amountValue,
                currency: currency,
                description: description.isEmpty ? nil : description,
                frequency: frequency.rawValue,
                dayOfMonth: dayOfMonth,
                dayOfWeek: 1,
                nextGenerationDate: calculateFirstGenerationDate(),
                lastGeneratedDate: nil,
                lastInvoiceId: nil,
                active: true,
                autoPdf: autoPdf,
                autoSend: false,
                emailApp: nil,
                generatedCount: 0,
                notes: nil
            )

            _ = try? await appState.database.createRecurringInvoice(recurring)
            onSave()
            isSaving = false
            dismiss()
        }
    }

    private func calculateFirstGenerationDate() -> Date {
        let calendar = Calendar.current
        var components = calendar.dateComponents([.year, .month], from: Date())
        components.day = min(dayOfMonth, 28)

        guard let thisMonthDate = calendar.date(from: components) else { return Date() }

        // If the day has passed this month, schedule for next occurrence
        if thisMonthDate <= Date() {
            switch frequency {
            case .weekly:
                return calendar.date(byAdding: .weekOfYear, value: 1, to: Date()) ?? Date()
            case .biweekly:
                return calendar.date(byAdding: .weekOfYear, value: 2, to: Date()) ?? Date()
            case .monthly:
                return calendar.date(byAdding: .month, value: 1, to: thisMonthDate) ?? thisMonthDate
            case .quarterly:
                return calendar.date(byAdding: .month, value: 3, to: thisMonthDate) ?? thisMonthDate
            case .yearly:
                return calendar.date(byAdding: .year, value: 1, to: thisMonthDate) ?? thisMonthDate
            }
        }

        return thisMonthDate
    }
}

#Preview {
    RecurringInvoicesSheet()
        .environmentObject(AppState())
}
