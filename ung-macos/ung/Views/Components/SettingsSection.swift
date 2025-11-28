//
//  SettingsSection.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI
import UniformTypeIdentifiers

struct SettingsSection: View {
    @EnvironmentObject var appState: AppState
    @State private var showImportPicker = false
    @State private var showExportPicker = false
    @State private var importError: String?
    @State private var showImportSuccess = false
    @State private var showResetConfirmation = false

    var body: some View {
        VStack(spacing: 12) {
            // Display settings
            displaySettingsCard

            // Data management
            dataManagementCard

            // Database section
            databaseCard

            // About section
            aboutCard
        }
    }

    // MARK: - Display Settings
    private var displaySettingsCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Display")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            SettingsToggle(
                icon: "eye.slash.fill",
                title: "Secure Mode",
                subtitle: "Hide sensitive amounts",
                isOn: $appState.secureMode
            )

            SettingsToggle(
                icon: "globe",
                title: "Global Database",
                subtitle: "Use ~/.ung/ung.db",
                isOn: $appState.useGlobalDatabase
            )
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }

    // MARK: - Data Management
    private var dataManagementCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Data")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            // Import SQL
            SettingsButton(
                icon: "square.and.arrow.down.fill",
                title: "Import Database",
                subtitle: "Import from SQL file",
                color: .blue
            ) {
                showImportPicker = true
            }
            .fileImporter(
                isPresented: $showImportPicker,
                allowedContentTypes: [.init(filenameExtension: "sql")!, .init(filenameExtension: "db")!, .database],
                allowsMultipleSelection: false
            ) { result in
                handleImport(result)
            }

            // Export database
            SettingsButton(
                icon: "square.and.arrow.up.fill",
                title: "Export Database",
                subtitle: "Export to SQL/CSV",
                color: .green
            ) {
                exportDatabase()
            }

            // Backup
            SettingsButton(
                icon: "externaldrive.fill",
                title: "Backup Data",
                subtitle: "Create full backup",
                color: .purple
            ) {
                createBackup()
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
        .alert("Import Successful", isPresented: $showImportSuccess) {
            Button("OK") {
                Task { await appState.refreshDashboard() }
            }
        } message: {
            Text("Database imported successfully. Your data has been updated.")
        }
        .alert("Import Error", isPresented: .init(
            get: { importError != nil },
            set: { if !$0 { importError = nil } }
        )) {
            Button("OK") { importError = nil }
        } message: {
            Text(importError ?? "Unknown error")
        }
    }

    // MARK: - Database Card
    private var databaseCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Database")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            // Database location
            HStack(spacing: 8) {
                Image(systemName: "folder.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)

                VStack(alignment: .leading, spacing: 2) {
                    Text("Location")
                        .font(.system(size: 11, weight: .medium))
                    Text(appState.useGlobalDatabase ? "~/.ung/ung.db" : ".ung/ung.db")
                        .font(.system(size: 10, design: .monospaced))
                        .foregroundColor(.secondary)
                }

                Spacer()

                Button(action: openDatabaseFolder) {
                    Image(systemName: "arrow.up.forward.square")
                        .font(.system(size: 11))
                        .foregroundColor(.blue)
                }
                .buttonStyle(.plain)
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color(nsColor: .separatorColor).opacity(0.2))
            )

            // Reset database (dangerous)
            SettingsButton(
                icon: "trash.fill",
                title: "Reset Database",
                subtitle: "Delete all data (cannot be undone)",
                color: .red
            ) {
                showResetConfirmation = true
            }
            .confirmationDialog(
                "Reset Database?",
                isPresented: $showResetConfirmation,
                titleVisibility: .visible
            ) {
                Button("Reset", role: .destructive) {
                    resetDatabase()
                }
                Button("Cancel", role: .cancel) {}
            } message: {
                Text("This will permanently delete all your data including clients, contracts, invoices, expenses, and tracking sessions. This action cannot be undone.")
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }

    // MARK: - About Card
    private var aboutCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("About")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text("UNG")
                        .font(.system(size: 13, weight: .semibold))
                    Text("Time tracking & invoicing")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }

                Spacer()

                Text("v1.0")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color(nsColor: .separatorColor).opacity(0.3))
                    .cornerRadius(6)
            }

            Divider()

            HStack(spacing: 16) {
                Button("Documentation") {
                    if let url = URL(string: "https://github.com/Andriiklymiuk/ung") {
                        NSWorkspace.shared.open(url)
                    }
                }
                .buttonStyle(.plain)
                .font(.system(size: 11))
                .foregroundColor(.blue)

                Button("Report Issue") {
                    if let url = URL(string: "https://github.com/Andriiklymiuk/ung/issues") {
                        NSWorkspace.shared.open(url)
                    }
                }
                .buttonStyle(.plain)
                .font(.system(size: 11))
                .foregroundColor(.blue)
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }

    // MARK: - Actions
    private func handleImport(_ result: Result<[URL], Error>) {
        switch result {
        case .success(let urls):
            guard let url = urls.first else { return }

            // Get security-scoped access
            guard url.startAccessingSecurityScopedResource() else {
                importError = "Could not access the selected file"
                return
            }
            defer { url.stopAccessingSecurityScopedResource() }

            Task {
                let success = await appState.cliService.importDatabase(from: url.path)
                if success {
                    showImportSuccess = true
                } else {
                    importError = "Failed to import database. Please check the file format."
                }
            }

        case .failure(let error):
            importError = error.localizedDescription
        }
    }

    private func exportDatabase() {
        Task {
            let panel = NSSavePanel()
            panel.title = "Export Database"
            panel.nameFieldStringValue = "ung_export.sql"
            panel.allowedContentTypes = [.init(filenameExtension: "sql")!]

            let response = await panel.beginSheetModal(for: NSApp.keyWindow ?? NSWindow())

            if response == .OK, let url = panel.url {
                let success = await appState.cliService.exportDatabase(to: url.path)
                if !success {
                    importError = "Failed to export database"
                }
            }
        }
    }

    private func createBackup() {
        Task {
            let dateFormatter = DateFormatter()
            dateFormatter.dateFormat = "yyyy-MM-dd_HHmmss"
            let timestamp = dateFormatter.string(from: Date())

            let panel = NSSavePanel()
            panel.title = "Create Backup"
            panel.nameFieldStringValue = "ung_backup_\(timestamp).db"
            panel.allowedContentTypes = [.database]

            let response = await panel.beginSheetModal(for: NSApp.keyWindow ?? NSWindow())

            if response == .OK, let url = panel.url {
                let success = await appState.cliService.backupDatabase(to: url.path)
                if !success {
                    importError = "Failed to create backup"
                }
            }
        }
    }

    private func openDatabaseFolder() {
        let path = appState.useGlobalDatabase
            ? "\(NSHomeDirectory())/.ung"
            : FileManager.default.currentDirectoryPath + "/.ung"

        if let url = URL(string: "file://\(path)") {
            NSWorkspace.shared.open(url)
        }
    }

    private func resetDatabase() {
        Task {
            let success = await appState.cliService.resetDatabase()
            if success {
                appState.checkStatus()
            } else {
                importError = "Failed to reset database"
            }
        }
    }
}

// MARK: - Settings Toggle
struct SettingsToggle: View {
    let icon: String
    let title: String
    let subtitle: String
    @Binding var isOn: Bool

    var body: some View {
        Toggle(isOn: $isOn) {
            HStack(spacing: 10) {
                Image(systemName: icon)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                    .frame(width: 20)

                VStack(alignment: .leading, spacing: 1) {
                    Text(title)
                        .font(.system(size: 12, weight: .medium))
                    Text(subtitle)
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }
            }
        }
        .toggleStyle(.switch)
        .controlSize(.small)
    }
}

// MARK: - Settings Button
struct SettingsButton: View {
    let icon: String
    let title: String
    let subtitle: String
    let color: Color
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 10) {
                Image(systemName: icon)
                    .font(.system(size: 12))
                    .foregroundColor(color)
                    .frame(width: 20)

                VStack(alignment: .leading, spacing: 1) {
                    Text(title)
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.primary)
                    Text(subtitle)
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }

                Spacer()

                Image(systemName: "chevron.right")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color(nsColor: .separatorColor).opacity(0.2))
            )
        }
        .buttonStyle(.plain)
    }
}

#Preview {
    SettingsSection()
        .padding()
        .frame(width: 320)
        .environmentObject(AppState())
}
