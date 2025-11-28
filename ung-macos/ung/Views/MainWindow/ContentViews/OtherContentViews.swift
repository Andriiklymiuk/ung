//
//  OtherContentViews.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

// MARK: - Tracking Content
struct TrackingContent: View {
    @EnvironmentObject var appState: AppState
    @State private var projectName = ""
    @State private var selectedClientId: Int?
    @State private var isStarting = false

    var body: some View {
        VStack(spacing: 24) {
            // Active tracking banner
            if appState.isTracking, let session = appState.activeSession {
                activeTrackingBanner(session)
            }

            HStack(alignment: .top, spacing: 24) {
                // Start tracking form
                startTrackingForm
                    .frame(maxWidth: 400)

                // Recent sessions
                recentSessionsList
            }

            Spacer()
        }
        .padding(24)
    }

    private func activeTrackingBanner(_ session: ActiveSession) -> some View {
        HStack(spacing: 16) {
            Circle()
                .fill(Color.red)
                .frame(width: 12, height: 12)

            VStack(alignment: .leading, spacing: 2) {
                Text(session.project)
                    .font(.system(size: 16, weight: .semibold))
                if !session.client.isEmpty {
                    Text(session.client)
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            Text(session.formattedDuration)
                .font(.system(size: 28, weight: .bold, design: .monospaced))
                .foregroundColor(.primary)

            Button(action: { Task { await appState.stopTracking() } }) {
                HStack(spacing: 6) {
                    Image(systemName: "stop.fill")
                    Text("Stop")
                }
                .font(.system(size: 14, weight: .medium))
                .foregroundColor(.white)
                .padding(.horizontal, 20)
                .padding(.vertical, 10)
                .background(RoundedRectangle(cornerRadius: 10).fill(Color.red))
            }
            .buttonStyle(.plain)
        }
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(Color.red.opacity(0.1))
                .overlay(
                    RoundedRectangle(cornerRadius: 16)
                        .stroke(Color.red.opacity(0.3), lineWidth: 1)
                )
        )
    }

    private var startTrackingForm: some View {
        ContentCard(title: "Start Tracking", icon: "play.fill") {
            VStack(spacing: 16) {
                // Project name
                VStack(alignment: .leading, spacing: 6) {
                    Text("Project / Task")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)

                    TextField("What are you working on?", text: $projectName)
                        .textFieldStyle(.plain)
                        .padding(12)
                        .background(
                            RoundedRectangle(cornerRadius: 8)
                                .fill(Color(nsColor: .textBackgroundColor))
                        )
                }

                // Client selection
                VStack(alignment: .leading, spacing: 6) {
                    Text("Client (Optional)")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)

                    Picker("", selection: $selectedClientId) {
                        Text("No Client").tag(nil as Int?)
                        ForEach(appState.clients) { client in
                            Text(client.name).tag(client.id as Int?)
                        }
                    }
                    .pickerStyle(.menu)
                }

                // Start button
                Button(action: startTracking) {
                    HStack(spacing: 8) {
                        if isStarting {
                            ProgressView().scaleEffect(0.8)
                        } else {
                            Image(systemName: "play.fill")
                        }
                        Text("Start Tracking")
                    }
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 12)
                    .background(
                        RoundedRectangle(cornerRadius: 10)
                            .fill(projectName.isEmpty ? Color.gray : Color.green)
                    )
                }
                .buttonStyle(.plain)
                .disabled(projectName.isEmpty || isStarting || appState.isTracking)
            }
        }
    }

    private var recentSessionsList: some View {
        ContentCard(title: "Recent Sessions", icon: "clock.fill") {
            if appState.recentSessions.isEmpty {
                Text("No sessions yet")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
                    .padding(.vertical, 20)
            } else {
                VStack(spacing: 8) {
                    ForEach(appState.recentSessions) { session in
                        HStack {
                            VStack(alignment: .leading, spacing: 2) {
                                Text(session.project)
                                    .font(.system(size: 13, weight: .medium))
                                Text(session.date)
                                    .font(.system(size: 11))
                                    .foregroundColor(.secondary)
                            }
                            Spacer()
                            Text(appState.secureMode ? "**:**" : session.duration)
                                .font(.system(size: 13, weight: .medium, design: .monospaced))
                                .foregroundColor(.blue)
                        }
                        .padding(12)
                        .background(RoundedRectangle(cornerRadius: 8).fill(Color(nsColor: .controlBackgroundColor)))
                    }
                }
            }
        }
    }

    private func startTracking() {
        isStarting = true
        Task {
            await appState.startTracking(project: projectName, clientId: selectedClientId)
            projectName = ""
            isStarting = false
        }
    }
}

// MARK: - Clients Content
struct ClientsContent: View {
    @EnvironmentObject var appState: AppState
    @State private var showAddClient = false
    @State private var newClientName = ""
    @State private var newClientEmail = ""

    var body: some View {
        VStack(spacing: 24) {
            if appState.clients.isEmpty {
                emptyState
            } else {
                clientsList
            }
        }
        .padding(24)
        .sheet(isPresented: $showAddClient) {
            addClientSheet
        }
    }

    private var emptyState: some View {
        VStack(spacing: 16) {
            Image(systemName: "person.2.fill")
                .font(.system(size: 48))
                .foregroundColor(.secondary.opacity(0.5))

            Text("No Clients Yet")
                .font(.system(size: 18, weight: .semibold))

            Text("Add your first client to start tracking time and creating invoices")
                .font(.system(size: 14))
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)

            Button("Add Client") {
                showAddClient = true
            }
            .buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var clientsList: some View {
        LazyVGrid(columns: [GridItem(.adaptive(minimum: 280))], spacing: 16) {
            ForEach(appState.clients) { client in
                ClientCard(client: client)
            }

            // Add client card
            Button(action: { showAddClient = true }) {
                VStack(spacing: 12) {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 32))
                        .foregroundColor(.purple.opacity(0.6))
                    Text("Add Client")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundColor(.secondary)
                }
                .frame(maxWidth: .infinity)
                .frame(height: 120)
                .background(
                    RoundedRectangle(cornerRadius: 12)
                        .strokeBorder(style: StrokeStyle(lineWidth: 2, dash: [8]))
                        .foregroundColor(.secondary.opacity(0.3))
                )
            }
            .buttonStyle(.plain)
        }
    }

    private var addClientSheet: some View {
        VStack(spacing: 20) {
            Text("Add New Client")
                .font(.system(size: 18, weight: .bold))

            VStack(alignment: .leading, spacing: 16) {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Name")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                    TextField("Client name", text: $newClientName)
                        .textFieldStyle(.roundedBorder)
                }

                VStack(alignment: .leading, spacing: 6) {
                    Text("Email")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                    TextField("client@example.com", text: $newClientEmail)
                        .textFieldStyle(.roundedBorder)
                }
            }

            HStack {
                Button("Cancel") {
                    showAddClient = false
                    newClientName = ""
                    newClientEmail = ""
                }
                .buttonStyle(.bordered)

                Spacer()

                Button("Add Client") {
                    Task {
                        _ = await appState.cliService.createClient(name: newClientName, email: newClientEmail.isEmpty ? nil : newClientEmail)
                        await appState.refreshDashboard()
                        showAddClient = false
                        newClientName = ""
                        newClientEmail = ""
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(newClientName.isEmpty)
            }
        }
        .padding(24)
        .frame(width: 400)
    }
}

struct ClientCard: View {
    let client: Client
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(Color.purple.opacity(0.15))
                    .frame(width: 48, height: 48)

                Text(String(client.name.prefix(1)).uppercased())
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(.purple)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(client.name)
                    .font(.system(size: 14, weight: .semibold))

                if !client.email.isEmpty {
                    Text(client.email)
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            Image(systemName: "chevron.right")
                .font(.system(size: 12))
                .foregroundColor(.secondary)
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }
}

// MARK: - Contracts Content
struct ContractsContent: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(spacing: 24) {
            if appState.contracts.isEmpty {
                VStack(spacing: 16) {
                    Image(systemName: "doc.text.fill")
                        .font(.system(size: 48))
                        .foregroundColor(.secondary.opacity(0.5))

                    Text("No Contracts Yet")
                        .font(.system(size: 18, weight: .semibold))

                    Text("Create contracts to define rates and terms with your clients")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)

                    Button("Create Contract") {}
                        .buttonStyle(.borderedProminent)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                LazyVGrid(columns: [GridItem(.adaptive(minimum: 300))], spacing: 16) {
                    ForEach(appState.contracts) { contract in
                        ContractCard(contract: contract)
                    }
                }
            }
        }
        .padding(24)
    }
}

struct ContractCard: View {
    let contract: Contract
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "doc.text.fill")
                    .foregroundColor(.indigo)
                Text(contract.name)
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
            }

            Text(contract.clientName)
                .font(.system(size: 12))
                .foregroundColor(.secondary)

            HStack {
                Text(contract.type.capitalized)
                    .font(.system(size: 11, weight: .medium))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.indigo.opacity(0.15))
                    .cornerRadius(6)

                Spacer()

                Text("$\(Int(contract.rate))/hr")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.green)
            }
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }
}

// MARK: - Invoices Content
struct InvoicesContent: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(spacing: 24) {
            // Invoice stats
            HStack(spacing: 16) {
                InvoiceStatCard(title: "Total", value: "\(appState.invoiceCount)", color: .blue)
                InvoiceStatCard(title: "Pending", value: appState.formatCurrency(appState.metrics.pendingAmount), color: .orange)
                InvoiceStatCard(title: "Overdue", value: appState.formatCurrency(appState.metrics.overdueAmount), color: .red)
            }

            // Invoice list
            if appState.recentInvoices.isEmpty {
                VStack(spacing: 16) {
                    Image(systemName: "doc.plaintext.fill")
                        .font(.system(size: 48))
                        .foregroundColor(.secondary.opacity(0.5))

                    Text("No Invoices Yet")
                        .font(.system(size: 18, weight: .semibold))

                    Text("Create your first invoice from tracked time")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ContentCard(title: "Recent Invoices", icon: "doc.plaintext.fill") {
                    VStack(spacing: 8) {
                        ForEach(appState.recentInvoices) { invoice in
                            InvoiceRow(invoice: invoice)
                        }
                    }
                }
            }
        }
        .padding(24)
    }
}

struct InvoiceStatCard: View {
    let title: String
    let value: String
    let color: Color
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.system(size: 12))
                .foregroundColor(.secondary)
            Text(value)
                .font(.system(size: 20, weight: .bold))
                .foregroundColor(color)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }
}

struct InvoiceRow: View {
    @EnvironmentObject var appState: AppState
    let invoice: RecentInvoice

    var statusColor: Color {
        switch invoice.status.lowercased() {
        case "paid": return .green
        case "sent", "pending": return .orange
        case "overdue": return .red
        default: return .gray
        }
    }

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(invoice.invoiceNum)
                    .font(.system(size: 13, weight: .medium))
                Text(invoice.client)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Text(appState.secureMode ? "****" : invoice.amount)
                .font(.system(size: 13, weight: .semibold))

            Text(invoice.status.capitalized)
                .font(.system(size: 10, weight: .medium))
                .foregroundColor(statusColor)
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(statusColor.opacity(0.15))
                .cornerRadius(6)
        }
        .padding(12)
        .background(RoundedRectangle(cornerRadius: 8).fill(Color(nsColor: .controlBackgroundColor)))
    }
}

// MARK: - Expenses Content
struct ExpensesContent: View {
    @EnvironmentObject var appState: AppState
    @State private var showLogExpense = false

    var body: some View {
        VStack(spacing: 24) {
            if appState.recentExpenses.isEmpty {
                VStack(spacing: 16) {
                    Image(systemName: "dollarsign.circle.fill")
                        .font(.system(size: 48))
                        .foregroundColor(.secondary.opacity(0.5))

                    Text("No Expenses Yet")
                        .font(.system(size: 18, weight: .semibold))

                    Text("Track your business expenses for tax deductions")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)

                    Button("Log Expense") {
                        showLogExpense = true
                    }
                    .buttonStyle(.borderedProminent)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ContentCard(title: "Recent Expenses", icon: "dollarsign.circle.fill") {
                    VStack(spacing: 8) {
                        ForEach(appState.recentExpenses) { expense in
                            ExpenseContentRow(expense: expense)
                        }
                    }
                }
            }
        }
        .padding(24)
    }
}

struct ExpenseContentRow: View {
    @EnvironmentObject var appState: AppState
    let expense: RecentExpense

    var categoryIcon: String {
        switch expense.category.lowercased() {
        case "software": return "app.badge.fill"
        case "hardware": return "desktopcomputer"
        case "travel": return "airplane"
        case "meals": return "fork.knife"
        default: return "tag.fill"
        }
    }

    var body: some View {
        HStack {
            Image(systemName: categoryIcon)
                .font(.system(size: 14))
                .foregroundColor(.orange)
                .frame(width: 32)

            VStack(alignment: .leading, spacing: 2) {
                Text(expense.description)
                    .font(.system(size: 13, weight: .medium))
                Text(expense.category)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Text(appState.secureMode ? "****" : expense.amount)
                .font(.system(size: 13, weight: .semibold))
                .foregroundColor(.orange)
        }
        .padding(12)
        .background(RoundedRectangle(cornerRadius: 8).fill(Color(nsColor: .controlBackgroundColor)))
    }
}

// MARK: - Reports Content
struct ReportsContent: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(spacing: 24) {
            // Period selector
            HStack {
                Text("Analytics & Reports")
                    .font(.system(size: 16, weight: .semibold))
                Spacer()
            }

            // Metrics overview
            HStack(spacing: 16) {
                ReportMetricCard(
                    title: "Total Revenue",
                    value: appState.formatCurrency(appState.metrics.totalRevenue),
                    icon: "chart.line.uptrend.xyaxis",
                    color: .green
                )

                ReportMetricCard(
                    title: "Hours Tracked",
                    value: appState.formatHours(appState.metrics.weeklyHours * 4),
                    icon: "clock.fill",
                    color: .blue
                )

                ReportMetricCard(
                    title: "Active Clients",
                    value: "\(appState.clientCount)",
                    icon: "person.2.fill",
                    color: .purple
                )
            }

            // Coming soon placeholder
            VStack(spacing: 12) {
                Image(systemName: "chart.bar.xaxis")
                    .font(.system(size: 48))
                    .foregroundColor(.secondary.opacity(0.5))

                Text("Detailed Reports Coming Soon")
                    .font(.system(size: 16, weight: .medium))

                Text("Advanced analytics, charts, and exportable reports")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .padding(24)
    }
}

struct ReportMetricCard: View {
    let title: String
    let value: String
    let icon: String
    let color: Color
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 20))
                .foregroundColor(color)

            Text(value)
                .font(.system(size: 24, weight: .bold, design: .rounded))

            Text(title)
                .font(.system(size: 12))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }
}

// MARK: - Settings Content
struct SettingsContent: View {
    @EnvironmentObject var appState: AppState
    @State private var passwordInput = ""
    @State private var showPasswordField = false

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Display settings
                SettingsCard(title: "Display", icon: "eye") {
                    Toggle("Secure Mode", isOn: $appState.secureMode)
                    Text("Hide sensitive amounts and financial data")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                // Database settings
                SettingsCard(title: "Database", icon: "cylinder") {
                    Toggle("Use Global Database", isOn: $appState.useGlobalDatabase)
                    Text("Store data in ~/.ung/ accessible from anywhere")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                    Divider()

                    HStack {
                        Text("Password Protection")
                            .font(.system(size: 13))
                        Spacer()
                        if appState.hasStoredPassword {
                            HStack(spacing: 8) {
                                Image(systemName: "checkmark.shield.fill")
                                    .foregroundColor(.green)
                                Text("Protected")
                                    .font(.system(size: 12))
                                    .foregroundColor(.green)

                                Button("Clear") {
                                    _ = appState.clearPassword()
                                }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                            }
                        } else {
                            Button("Set Password") {
                                showPasswordField.toggle()
                            }
                            .buttonStyle(.bordered)
                            .controlSize(.small)
                        }
                    }

                    if showPasswordField {
                        HStack {
                            SecureField("Enter password", text: $passwordInput)
                                .textFieldStyle(.roundedBorder)
                            Button("Save") {
                                if !passwordInput.isEmpty {
                                    _ = appState.savePassword(passwordInput)
                                    passwordInput = ""
                                    showPasswordField = false
                                }
                            }
                            .buttonStyle(.borderedProminent)
                            .disabled(passwordInput.isEmpty)
                        }
                    }
                }

                // About
                SettingsCard(title: "About", icon: "info.circle") {
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("UNG")
                                .font(.system(size: 14, weight: .semibold))
                            Text("Freelance Time Tracking & Invoicing")
                                .font(.system(size: 12))
                                .foregroundColor(.secondary)
                        }
                        Spacer()
                        Text("v1.0")
                            .font(.system(size: 12, design: .monospaced))
                            .foregroundColor(.secondary)
                    }

                    Divider()

                    HStack(spacing: 16) {
                        Link("Documentation", destination: URL(string: "https://github.com/Andriiklymiuk/ung")!)
                        Link("Report Issue", destination: URL(string: "https://github.com/Andriiklymiuk/ung/issues")!)
                    }
                    .font(.system(size: 13))
                }
            }
            .padding(24)
        }
    }
}

struct SettingsCard<Content: View>: View {
    let title: String
    let icon: String
    @ViewBuilder let content: () -> Content
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: icon)
                    .foregroundColor(.secondary)
                Text(title)
                    .font(.system(size: 14, weight: .semibold))
            }

            content()
        }
        .padding(20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }
}

// MARK: - Content Card Helper
struct ContentCard<Content: View>: View {
    let title: String
    let icon: String
    @ViewBuilder let content: () -> Content
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: icon)
                    .foregroundColor(.secondary)
                Text(title)
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
            }

            content()
        }
        .padding(20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }
}
