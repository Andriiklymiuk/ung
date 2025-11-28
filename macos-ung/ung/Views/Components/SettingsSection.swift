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
  @State private var isCheckingForUpdates = false
  @State private var updateCheckResult: UpdateCheckResult?
  @State private var showUpdateAlert = false

  enum UpdateCheckResult {
    case upToDate
    case updateAvailable(version: String, url: String)
    case error(String)
  }

  var body: some View {
    VStack(spacing: 12) {
      // Display settings
      displaySettingsCard

      // Security settings
      securitySettingsCard

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

  // MARK: - Security Settings
  private var securitySettingsCard: some View {
    VStack(alignment: .leading, spacing: 10) {
      Text("Privacy & Security")
        .font(.system(size: 11, weight: .medium))
        .foregroundColor(.secondary)

      // Secure mode toggle with better UI
      HStack(spacing: 12) {
        ZStack {
          RoundedRectangle(cornerRadius: 10)
            .fill(appState.secureMode ? Color.green.opacity(0.15) : Color.secondary.opacity(0.1))
            .frame(width: 40, height: 40)

          Image(systemName: appState.secureMode ? "eye.slash.fill" : "eye.fill")
            .font(.system(size: 16))
            .foregroundColor(appState.secureMode ? .green : .secondary)
        }

        VStack(alignment: .leading, spacing: 2) {
          Text("Secure Mode")
            .font(.system(size: 12, weight: .medium))

          Text(appState.secureMode ? "Amounts are hidden" : "Amounts are visible")
            .font(.system(size: 10))
            .foregroundColor(.secondary)
        }

        Spacer()

        Toggle("", isOn: $appState.secureMode)
          .toggleStyle(.switch)
          .controlSize(.small)
          .labelsHidden()
      }
      .padding(10)
      .background(
        RoundedRectangle(cornerRadius: 10)
          .fill(
            appState.secureMode
              ? Color.green.opacity(0.05) : Color(nsColor: .separatorColor).opacity(0.2))
      )
      .overlay(
        RoundedRectangle(cornerRadius: 10)
          .stroke(appState.secureMode ? Color.green.opacity(0.3) : Color.clear, lineWidth: 1)
      )
      .animation(.easeInOut(duration: 0.2), value: appState.secureMode)

      Text("When enabled, all monetary values will be replaced with asterisks for privacy.")
        .font(.system(size: 10))
        .foregroundColor(.secondary)
        .padding(.horizontal, 4)
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
        allowedContentTypes: [
          .init(filenameExtension: "sql")!, .init(filenameExtension: "db")!, .database,
        ],
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
    .alert(
      "Import Error",
      isPresented: .init(
        get: { importError != nil },
        set: { if !$0 { importError = nil } }
      )
    ) {
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
        Text(
          "This will permanently delete all your data including clients, contracts, invoices, expenses, and tracking sessions. This action cannot be undone."
        )
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

        Text("v\(AppVersion.current)")
          .font(.system(size: 11, weight: .medium))
          .foregroundColor(.secondary)
          .padding(.horizontal, 8)
          .padding(.vertical, 4)
          .background(Color(nsColor: .separatorColor).opacity(0.3))
          .cornerRadius(6)
      }

      // Check for Updates button
      Button(action: checkForUpdates) {
        HStack(spacing: 10) {
          if isCheckingForUpdates {
            ProgressView()
              .scaleEffect(0.6)
              .frame(width: 20)
          } else {
            Image(systemName: "arrow.triangle.2.circlepath")
              .font(.system(size: 12))
              .foregroundColor(.blue)
              .frame(width: 20)
          }

          VStack(alignment: .leading, spacing: 1) {
            Text("Check for Updates")
              .font(.system(size: 12, weight: .medium))
              .foregroundColor(.primary)
            Text(isCheckingForUpdates ? "Checking..." : "Current: v\(AppVersion.current)")
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
      .disabled(isCheckingForUpdates)

      Divider()

      HStack(spacing: 16) {
        Button("Documentation") {
          if let url = URL(string: "https://andriiklymiuk.github.io/ung/") {
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

        Spacer()

        Button("GitHub") {
          if let url = URL(string: "https://github.com/Andriiklymiuk/ung") {
            NSWorkspace.shared.open(url)
          }
        }
        .buttonStyle(.plain)
        .font(.system(size: 11))
        .foregroundColor(.blue)
      }

      Divider()

      // Quit button
      Button(action: { NSApplication.shared.terminate(nil) }) {
        HStack(spacing: 10) {
          Image(systemName: "power")
            .font(.system(size: 12))
            .foregroundColor(.red)
            .frame(width: 20)

          Text("Quit UNG")
            .font(.system(size: 12, weight: .medium))
            .foregroundColor(.primary)

          Spacer()

          Text("⌘Q")
            .font(.system(size: 10, design: .monospaced))
            .foregroundColor(.secondary)
        }
        .padding(10)
        .background(
          RoundedRectangle(cornerRadius: 8)
            .fill(Color(nsColor: .separatorColor).opacity(0.2))
        )
      }
      .buttonStyle(.plain)
      .keyboardShortcut("q", modifiers: .command)
    }
    .padding(12)
    .background(
      RoundedRectangle(cornerRadius: 10)
        .fill(Color(nsColor: .controlBackgroundColor))
    )
    .alert(updateAlertTitle, isPresented: $showUpdateAlert) {
      if case .updateAvailable(_, let url) = updateCheckResult {
        Button("Download Update") {
          if let downloadUrl = URL(string: url) {
            NSWorkspace.shared.open(downloadUrl)
          }
        }
        Button("Later", role: .cancel) {}
      } else {
        Button("OK") {}
      }
    } message: {
      Text(updateAlertMessage)
    }
  }

  private var updateAlertTitle: String {
    switch updateCheckResult {
    case .upToDate:
      return "You're up to date!"
    case .updateAvailable:
      return "Update Available"
    case .error:
      return "Update Check Failed"
    case .none:
      return ""
    }
  }

  private var updateAlertMessage: String {
    switch updateCheckResult {
    case .upToDate:
      return "UNG v\(AppVersion.current) is the latest version."
    case .updateAvailable(let version, _):
      return
        "A new version (v\(version)) is available. You are currently running v\(AppVersion.current)."
    case .error(let message):
      return message
    case .none:
      return ""
    }
  }

  // MARK: - Actions
  private func checkForUpdates() {
    isCheckingForUpdates = true

    Task {
      defer {
        Task { @MainActor in
          isCheckingForUpdates = false
          showUpdateAlert = true
        }
      }

      // Check GitHub releases API for latest version
      guard let url = URL(string: "https://api.github.com/repos/Andriiklymiuk/ung/releases/latest")
      else {
        updateCheckResult = .error("Invalid URL")
        return
      }

      var request = URLRequest(url: url)
      request.setValue("application/vnd.github.v3+json", forHTTPHeaderField: "Accept")
      request.timeoutInterval = 10

      do {
        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
          updateCheckResult = .error(
            "Could not reach GitHub. Please check your internet connection.")
          return
        }

        if let json = try JSONSerialization.jsonObject(with: data) as? [String: Any],
          let tagName = json["tag_name"] as? String,
          let htmlUrl = json["html_url"] as? String
        {
          // Remove "v" prefix if present for comparison
          let latestVersion = tagName.hasPrefix("v") ? String(tagName.dropFirst()) : tagName
          let currentVersion = AppVersion.current

          // Compare versions
          if compareVersions(latestVersion, currentVersion) > 0 {
            updateCheckResult = .updateAvailable(version: latestVersion, url: htmlUrl)
          } else {
            updateCheckResult = .upToDate
          }
        } else {
          updateCheckResult = .upToDate
        }
      } catch {
        updateCheckResult = .error("Network error: \(error.localizedDescription)")
      }
    }
  }

  private func compareVersions(_ v1: String, _ v2: String) -> Int {
    let parts1 = v1.split(separator: ".").compactMap { Int($0) }
    let parts2 = v2.split(separator: ".").compactMap { Int($0) }

    let maxCount = max(parts1.count, parts2.count)

    for i in 0..<maxCount {
      let p1 = i < parts1.count ? parts1[i] : 0
      let p2 = i < parts2.count ? parts2[i] : 0

      if p1 > p2 { return 1 }
      if p1 < p2 { return -1 }
    }

    return 0
  }

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
    let path =
      appState.useGlobalDatabase
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
