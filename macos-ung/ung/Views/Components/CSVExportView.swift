//
//  CSVExportView.swift
//  ung
//
//  CSV export functionality for all data types
//

import SwiftUI
#if os(macOS)
import AppKit
#else
import UIKit
#endif

// MARK: - CSV Export View
struct CSVExportView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme

    @State private var selectedExportType: ExportType = .sessions
    @State private var dateRange: DateRangeOption = .allTime
    @State private var customStartDate = Calendar.current.date(byAdding: .month, value: -1, to: Date()) ?? Date()
    @State private var customEndDate = Date()
    @State private var isExporting = false
    @State private var showExportSuccess = false
    @State private var showExportError = false
    @State private var errorMessage = ""
    @State private var exportedFileURL: URL?

    enum ExportType: String, CaseIterable, Identifiable {
        case sessions = "Time Sessions"
        case clients = "Clients"
        case contracts = "Contracts"
        case invoices = "Invoices"
        case expenses = "Expenses"
        case all = "All Data"

        var id: String { rawValue }

        var icon: String {
            switch self {
            case .sessions: return "clock.fill"
            case .clients: return "person.2.fill"
            case .contracts: return "doc.text.fill"
            case .invoices: return "doc.plaintext.fill"
            case .expenses: return "dollarsign.circle.fill"
            case .all: return "square.and.arrow.up.fill"
            }
        }

        var color: Color {
            switch self {
            case .sessions: return Design.Colors.primary
            case .clients: return Design.Colors.purple
            case .contracts: return Design.Colors.indigo
            case .invoices: return Design.Colors.teal
            case .expenses: return Design.Colors.warning
            case .all: return Design.Colors.success
            }
        }
    }

    enum DateRangeOption: String, CaseIterable, Identifiable {
        case thisWeek = "This Week"
        case thisMonth = "This Month"
        case thisYear = "This Year"
        case lastMonth = "Last Month"
        case allTime = "All Time"
        case custom = "Custom Range"

        var id: String { rawValue }
    }

    var body: some View {
        ScrollView {
            VStack(spacing: Design.Spacing.lg) {
                // Header
                headerSection

                // Export Type Selection
                exportTypeSection

                // Date Range (for applicable types)
                if selectedExportType == .sessions || selectedExportType == .invoices || selectedExportType == .expenses || selectedExportType == .all {
                    dateRangeSection
                }

                // Export Preview
                previewSection

                // Export Button
                exportButton
            }
            .padding(Design.Spacing.md)
        }
        .alert("Export Successful", isPresented: $showExportSuccess) {
            #if os(macOS)
            Button("Show in Finder") {
                if let url = exportedFileURL {
                    NSWorkspace.shared.selectFile(url.path, inFileViewerRootedAtPath: url.deletingLastPathComponent().path)
                }
            }
            #endif
            Button("OK", role: .cancel) {}
        } message: {
            Text("Your data has been exported successfully.")
        }
        .alert("Export Failed", isPresented: $showExportError) {
            Button("OK", role: .cancel) {}
        } message: {
            Text(errorMessage)
        }
    }

    // MARK: - Header Section
    private var headerSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            HStack {
                Image(systemName: "square.and.arrow.up.fill")
                    .font(.system(size: 24))
                    .foregroundColor(Design.Colors.primary)
                Text("Export Data")
                    .font(Design.Typography.headingMedium)
            }
            Text("Export your data to CSV format for use in spreadsheets, accounting software, or backup purposes.")
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textSecondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    // MARK: - Export Type Section
    private var exportTypeSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
            Text("What to Export")
                .font(Design.Typography.labelMedium)
                .foregroundColor(Design.Colors.textSecondary)

            LazyVGrid(columns: [GridItem(.adaptive(minimum: 140))], spacing: Design.Spacing.sm) {
                ForEach(ExportType.allCases) { type in
                    ExportTypeCard(
                        type: type,
                        isSelected: selectedExportType == type,
                        action: { selectedExportType = type }
                    )
                }
            }
        }
    }

    // MARK: - Date Range Section
    private var dateRangeSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
            Text("Date Range")
                .font(Design.Typography.labelMedium)
                .foregroundColor(Design.Colors.textSecondary)

            VStack(spacing: Design.Spacing.xs) {
                Picker("Date Range", selection: $dateRange) {
                    ForEach(DateRangeOption.allCases) { option in
                        Text(option.rawValue).tag(option)
                    }
                }
                .pickerStyle(.segmented)

                if dateRange == .custom {
                    HStack(spacing: Design.Spacing.md) {
                        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
                            Text("Start Date")
                                .font(Design.Typography.labelSmall)
                                .foregroundColor(Design.Colors.textTertiary)
                            DatePicker("", selection: $customStartDate, displayedComponents: .date)
                                .labelsHidden()
                        }

                        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
                            Text("End Date")
                                .font(Design.Typography.labelSmall)
                                .foregroundColor(Design.Colors.textTertiary)
                            DatePicker("", selection: $customEndDate, displayedComponents: .date)
                                .labelsHidden()
                        }
                    }
                    .padding(.top, Design.Spacing.xs)
                }
            }
            .padding(Design.Spacing.md)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.md)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
            )
        }
    }

    // MARK: - Preview Section
    private var previewSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
            Text("Export Preview")
                .font(Design.Typography.labelMedium)
                .foregroundColor(Design.Colors.textSecondary)

            VStack(alignment: .leading, spacing: Design.Spacing.md) {
                HStack {
                    Image(systemName: selectedExportType.icon)
                        .foregroundColor(selectedExportType.color)
                    Text(selectedExportType.rawValue)
                        .font(Design.Typography.labelMedium)
                    Spacer()
                    Text(previewRecordCount)
                        .font(Design.Typography.bodySmall)
                        .foregroundColor(Design.Colors.textSecondary)
                }

                Divider()

                // Preview columns
                Text("Columns included:")
                    .font(Design.Typography.labelSmall)
                    .foregroundColor(Design.Colors.textTertiary)

                Text(previewColumns)
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textSecondary)
            }
            .padding(Design.Spacing.md)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.md)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
            )
        }
    }

    private var previewRecordCount: String {
        switch selectedExportType {
        case .sessions: return "\(appState.recentSessions.count) records"
        case .clients: return "\(appState.clients.count) records"
        case .contracts: return "\(appState.contracts.count) records"
        case .invoices: return "\(appState.recentInvoices.count) records"
        case .expenses: return "\(appState.recentExpenses.count) records"
        case .all: return "All data"
        }
    }

    private var previewColumns: String {
        switch selectedExportType {
        case .sessions:
            return "ID, Project, Client, Start Time, End Time, Duration, Notes"
        case .clients:
            return "ID, Name, Email, Address, Tax ID, Created At"
        case .contracts:
            return "ID, Name, Client, Type, Rate, Currency, Start Date, Active"
        case .invoices:
            return "ID, Invoice Number, Client, Amount, Currency, Status, Issued Date, Due Date"
        case .expenses:
            return "ID, Description, Amount, Currency, Category, Date"
        case .all:
            return "All columns from all data types"
        }
    }

    // MARK: - Export Button
    private var exportButton: some View {
        Button(action: performExport) {
            HStack(spacing: Design.Spacing.xs) {
                if isExporting {
                    ProgressView()
                        .scaleEffect(0.8)
                        #if os(macOS)
                        .controlSize(.small)
                        #endif
                } else {
                    Image(systemName: "square.and.arrow.up")
                }
                Text(isExporting ? "Exporting..." : "Export to CSV")
            }
        }
        .buttonStyle(DSPrimaryButtonStyle(color: selectedExportType.color))
        .disabled(isExporting)
    }

    // MARK: - Export Logic
    private func performExport() {
        isExporting = true

        Task {
            do {
                let csvContent = try await generateCSV()
                let filename = generateFilename()

                #if os(macOS)
                try saveCSVMacOS(content: csvContent, filename: filename)
                #else
                try await saveCSViOS(content: csvContent, filename: filename)
                #endif

                await MainActor.run {
                    isExporting = false
                    showExportSuccess = true
                }
            } catch {
                await MainActor.run {
                    isExporting = false
                    errorMessage = error.localizedDescription
                    showExportError = true
                }
            }
        }
    }

    private func generateFilename() -> String {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        let dateStr = dateFormatter.string(from: Date())

        switch selectedExportType {
        case .sessions: return "ung_sessions_\(dateStr).csv"
        case .clients: return "ung_clients_\(dateStr).csv"
        case .contracts: return "ung_contracts_\(dateStr).csv"
        case .invoices: return "ung_invoices_\(dateStr).csv"
        case .expenses: return "ung_expenses_\(dateStr).csv"
        case .all: return "ung_all_data_\(dateStr).csv"
        }
    }

    private func generateCSV() async throws -> String {
        switch selectedExportType {
        case .sessions:
            return generateSessionsCSV()
        case .clients:
            return generateClientsCSV()
        case .contracts:
            return generateContractsCSV()
        case .invoices:
            return generateInvoicesCSV()
        case .expenses:
            return generateExpensesCSV()
        case .all:
            return try await generateAllDataCSV()
        }
    }

    private func generateSessionsCSV() -> String {
        var csv = "ID,Project,Client,Start Time,End Time,Duration,Notes\n"

        for session in appState.recentSessions {
            let row = [
                String(session.id),
                escapeCSV(session.project),
                escapeCSV(session.date),
                "",  // Start time would need database query
                "",  // End time would need database query
                session.duration,
                ""   // Notes would need database query
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return csv
    }

    private func generateClientsCSV() -> String {
        var csv = "ID,Name,Email,Address,Tax ID\n"

        for client in appState.clients {
            let row = [
                String(client.id),
                escapeCSV(client.name),
                escapeCSV(client.email),
                escapeCSV(client.address),
                escapeCSV(client.taxId)
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return csv
    }

    private func generateContractsCSV() -> String {
        var csv = "ID,Name,Client,Type,Hourly Rate,Fixed Price,Currency,Active\n"

        for contract in appState.contracts {
            let row = [
                String(contract.id),
                escapeCSV(contract.name),
                escapeCSV(contract.clientName),
                escapeCSV(contract.type),
                String(format: "%.2f", contract.rate),
                String(format: "%.2f", contract.price),
                escapeCSV(contract.currency),
                "true"  // Active status would need database query
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return csv
    }

    private func generateInvoicesCSV() -> String {
        var csv = "ID,Invoice Number,Client,Amount,Currency,Status,Issued Date\n"

        for invoice in appState.recentInvoices {
            let row = [
                String(invoice.id),
                escapeCSV(invoice.invoiceNum),
                escapeCSV(invoice.client),
                invoice.amount.replacingOccurrences(of: "$", with: "").replacingOccurrences(of: ",", with: ""),
                "USD",
                escapeCSV(invoice.status),
                ""  // Issued date would need database query
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return csv
    }

    private func generateExpensesCSV() -> String {
        var csv = "ID,Description,Amount,Currency,Category,Date\n"

        for expense in appState.recentExpenses {
            let row = [
                String(expense.id),
                escapeCSV(expense.description),
                expense.amount.replacingOccurrences(of: "$", with: "").replacingOccurrences(of: ",", with: ""),
                "USD",
                escapeCSV(expense.category),
                expense.date
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return csv
    }

    private func generateAllDataCSV() async throws -> String {
        var csv = ""

        csv += "=== CLIENTS ===\n"
        csv += generateClientsCSV()
        csv += "\n"

        csv += "=== CONTRACTS ===\n"
        csv += generateContractsCSV()
        csv += "\n"

        csv += "=== TIME SESSIONS ===\n"
        csv += generateSessionsCSV()
        csv += "\n"

        csv += "=== INVOICES ===\n"
        csv += generateInvoicesCSV()
        csv += "\n"

        csv += "=== EXPENSES ===\n"
        csv += generateExpensesCSV()

        return csv
    }

    private func escapeCSV(_ value: String) -> String {
        if value.contains(",") || value.contains("\"") || value.contains("\n") {
            return "\"\(value.replacingOccurrences(of: "\"", with: "\"\""))\""
        }
        return value
    }

    #if os(macOS)
    private func saveCSVMacOS(content: String, filename: String) throws {
        let panel = NSSavePanel()
        panel.allowedContentTypes = [.commaSeparatedText]
        panel.nameFieldStringValue = filename
        panel.canCreateDirectories = true

        if panel.runModal() == .OK, let url = panel.url {
            try content.write(to: url, atomically: true, encoding: .utf8)
            exportedFileURL = url
        }
    }
    #else
    private func saveCSViOS(content: String, filename: String) async throws {
        let tempDir = FileManager.default.temporaryDirectory
        let fileURL = tempDir.appendingPathComponent(filename)

        try content.write(to: fileURL, atomically: true, encoding: .utf8)
        exportedFileURL = fileURL

        // Present share sheet
        await MainActor.run {
            if let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
               let rootVC = windowScene.windows.first?.rootViewController {
                let activityVC = UIActivityViewController(activityItems: [fileURL], applicationActivities: nil)

                if let popover = activityVC.popoverPresentationController {
                    popover.sourceView = rootVC.view
                    popover.sourceRect = CGRect(x: rootVC.view.bounds.midX, y: rootVC.view.bounds.midY, width: 0, height: 0)
                    popover.permittedArrowDirections = []
                }

                rootVC.present(activityVC, animated: true)
            }
        }
    }
    #endif
}

// MARK: - Export Type Card
struct ExportTypeCard: View {
    let type: CSVExportView.ExportType
    let isSelected: Bool
    let action: () -> Void
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        Button(action: action) {
            VStack(spacing: Design.Spacing.xs) {
                ZStack {
                    Circle()
                        .fill(type.color.opacity(isSelected ? 0.2 : 0.1))
                        .frame(width: 44, height: 44)

                    Image(systemName: type.icon)
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(type.color)
                }

                Text(type.rawValue)
                    .font(Design.Typography.labelSmall)
                    .foregroundColor(isSelected ? type.color : Design.Colors.textSecondary)
                    .multilineTextAlignment(.center)
            }
            .frame(maxWidth: .infinity)
            .padding(Design.Spacing.md)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.md)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
                    .overlay(
                        RoundedRectangle(cornerRadius: Design.Radius.md)
                            .stroke(isSelected ? type.color : Color.clear, lineWidth: 2)
                    )
            )
        }
        .buttonStyle(.plain)
    }
}

#Preview {
    CSVExportView()
        .environmentObject(AppState())
        .frame(width: 500, height: 700)
}
