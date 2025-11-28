//
//  SettingsSection.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import LocalAuthentication
import SwiftUI
import UniformTypeIdentifiers

struct SettingsSection: View {
  @EnvironmentObject var appState: AppState
  @Environment(\.colorScheme) var colorScheme
  @State private var showImportPicker = false
  @State private var importError: String?
  @State private var showImportSuccess = false
  @State private var showResetConfirmation = false
  @State private var isCheckingForUpdates = false
  @State private var updateCheckResult: UpdateCheckResult?
  @State private var showUpdateAlert = false
  @State private var biometricsAvailable = false
  @State private var biometricType: LABiometryType = .none

  enum UpdateCheckResult {
    case upToDate
    case updateAvailable(version: String, url: String)
    case error(String)
  }

  var body: some View {
    VStack(spacing: Design.Spacing.md) {
      // Security & Privacy - Most important first
      securityCard

      // App Lock (Touch ID / Password)
      appLockCard

      // Data Management
      dataCard

      // About & Updates
      aboutCard
    }
    .onAppear {
      checkBiometrics()
    }
  }

  private func checkBiometrics() {
    let context = LAContext()
    var error: NSError?
    if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
      biometricsAvailable = true
      biometricType = context.biometryType
    }
  }

  // MARK: - Security Card
  private var securityCard: some View {
    SettingsCard(title: "Privacy", icon: "lock.shield.fill", color: .blue) {
      VStack(spacing: Design.Spacing.sm) {
        SecureModeToggle()
      }
    }
  }

  // MARK: - App Lock Card
  private var appLockCard: some View {
    SettingsCard(
      title: "App Lock",
      icon: biometricType == .touchID
        ? "touchid" : (biometricType == .faceID ? "faceid" : "lock.fill"),
      color: .green
    ) {
      VStack(spacing: Design.Spacing.sm) {
        // Enable App Lock
        SettingsRow(
          icon: "lock.fill",
          title: "Require Authentication",
          subtitle: appState.appLockEnabled ? "App locks when you leave" : "App is not protected"
        ) {
          Toggle(
            "",
            isOn: Binding(
              get: { appState.appLockEnabled },
              set: { appState.setAppLockEnabled($0) }
            )
          )
          .toggleStyle(.switch)
          .controlSize(.small)
          .labelsHidden()
        }

        if appState.appLockEnabled && biometricsAvailable {
          Divider().padding(.vertical, 4)

          SettingsRow(
            icon: biometricType == .touchID ? "touchid" : "faceid",
            title: biometricType == .touchID ? "Use Touch ID" : "Use Face ID",
            subtitle: "Quick unlock with biometrics"
          ) {
            Toggle(
              "",
              isOn: Binding(
                get: { appState.useTouchID },
                set: { appState.setUseTouchID($0) }
              )
            )
            .toggleStyle(.switch)
            .controlSize(.small)
            .labelsHidden()
          }
        }

        if !biometricsAvailable && appState.appLockEnabled {
          HStack(spacing: Design.Spacing.xs) {
            Image(systemName: "info.circle.fill")
              .font(.system(size: 12))
              .foregroundColor(.blue)
            Text("Using system password for authentication")
              .font(Design.Typography.bodySmall)
              .foregroundColor(Design.Colors.textSecondary)
          }
          .padding(.top, 4)
        }
      }
    }
  }

  // MARK: - Data Card
  private var dataCard: some View {
    SettingsCard(title: "Data", icon: "externaldrive.fill", color: .purple) {
      VStack(spacing: 0) {
        // Import
        SettingsActionRow(
          icon: "square.and.arrow.down.fill",
          title: "Import Database",
          subtitle: "Restore from backup",
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

        Divider().padding(.leading, 44)

        // Export
        SettingsActionRow(
          icon: "square.and.arrow.up.fill",
          title: "Export Database",
          subtitle: "Save a copy of your data",
          color: .green
        ) {
          exportDatabase()
        }

        Divider().padding(.leading, 44)

        // Backup
        SettingsActionRow(
          icon: "clock.arrow.circlepath",
          title: "Create Backup",
          subtitle: "Full database backup",
          color: .orange
        ) {
          createBackup()
        }

        Divider().padding(.leading, 44)

        // Reset (Danger Zone)
        SettingsActionRow(
          icon: "trash.fill",
          title: "Reset Database",
          subtitle: "Delete all data permanently",
          color: .red
        ) {
          showResetConfirmation = true
        }
        .confirmationDialog(
          "Reset Database?",
          isPresented: $showResetConfirmation,
          titleVisibility: .visible
        ) {
          Button("Reset All Data", role: .destructive) {
            resetDatabase()
          }
          Button("Cancel", role: .cancel) {}
        } message: {
          Text(
            "This will permanently delete all your data including clients, contracts, invoices, expenses, and tracking sessions. This cannot be undone."
          )
        }
      }
    }
    .alert("Import Successful", isPresented: $showImportSuccess) {
      Button("OK") {
        Task { await appState.refreshDashboard() }
      }
    } message: {
      Text("Database imported successfully.")
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

  // MARK: - About Card
  private var aboutCard: some View {
    SettingsCard(title: "About", icon: "info.circle.fill", color: .gray) {
      VStack(spacing: 0) {
        // App info
        HStack(spacing: Design.Spacing.sm) {
          if let appIcon = NSImage(named: "AppIcon") {
            Image(nsImage: appIcon)
              .resizable()
              .frame(width: 48, height: 48)
              .cornerRadius(12)
          } else {
            RoundedRectangle(cornerRadius: 12)
              .fill(Color.blue)
              .frame(width: 48, height: 48)
              .overlay(
                Image(systemName: "clock.badge.checkmark")
                  .font(.system(size: 20))
                  .foregroundColor(.white)
              )
          }

          VStack(alignment: .leading, spacing: 2) {
            Text("UNG")
              .font(Design.Typography.headingSmall)
            Text("Time Tracking & Invoicing")
              .font(Design.Typography.bodySmall)
              .foregroundColor(Design.Colors.textSecondary)
          }

          Spacer()

          VStack(alignment: .trailing, spacing: 2) {
            Text("v\(AppVersion.current)")
              .font(Design.Typography.labelMedium)
              .foregroundColor(Design.Colors.textSecondary)
            if case .updateAvailable(let version, _) = updateCheckResult {
              Text("v\(version) available")
                .font(Design.Typography.labelSmall)
                .foregroundColor(.green)
            }
          }
        }
        .padding(.bottom, Design.Spacing.sm)

        Divider()

        // Check for updates
        SettingsActionRow(
          icon: "arrow.triangle.2.circlepath",
          title: "Check for Updates",
          subtitle: isCheckingForUpdates ? "Checking..." : "Get the latest version",
          color: .blue,
          showSpinner: isCheckingForUpdates
        ) {
          checkForUpdates()
        }
        .disabled(isCheckingForUpdates)

        Divider().padding(.leading, 44)

        // Links
        HStack(spacing: Design.Spacing.lg) {
          LinkButton(title: "Docs", icon: "book.fill") {
            openURL("https://andriiklymiuk.github.io/ung/")
          }

          LinkButton(title: "GitHub", icon: "chevron.left.forwardslash.chevron.right") {
            openURL("https://github.com/Andriiklymiuk/ung")
          }

          LinkButton(title: "Report Bug", icon: "ladybug.fill") {
            openURL("https://github.com/Andriiklymiuk/ung/issues")
          }
        }
        .padding(.vertical, Design.Spacing.sm)

        Divider()

        // Quit
        SettingsActionRow(
          icon: "power",
          title: "Quit UNG",
          subtitle: "⌘Q",
          color: .red
        ) {
          NSApplication.shared.terminate(nil)
        }
      }
    }
    .alert(updateAlertTitle, isPresented: $showUpdateAlert) {
      if case .updateAvailable(_, let url) = updateCheckResult {
        Button("Download") {
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
    case .upToDate: return "You're Up to Date"
    case .updateAvailable: return "Update Available"
    case .error: return "Update Check Failed"
    case .none: return ""
    }
  }

  private var updateAlertMessage: String {
    switch updateCheckResult {
    case .upToDate:
      return "UNG v\(AppVersion.current) is the latest version."
    case .updateAvailable(let version, _):
      return "Version \(version) is available. You're running v\(AppVersion.current)."
    case .error(let message):
      return message
    case .none:
      return ""
    }
  }

  // MARK: - Actions
  private func openURL(_ urlString: String) {
    if let url = URL(string: urlString) {
      NSWorkspace.shared.open(url)
    }
  }

  private func checkForUpdates() {
    isCheckingForUpdates = true

    Task {
      defer {
        Task { @MainActor in
          isCheckingForUpdates = false
          showUpdateAlert = true
        }
      }

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
          updateCheckResult = .error("Could not reach GitHub.")
          return
        }

        if let json = try JSONSerialization.jsonObject(with: data) as? [String: Any],
          let tagName = json["tag_name"] as? String,
          let htmlUrl = json["html_url"] as? String
        {
          let latestVersion = tagName.hasPrefix("v") ? String(tagName.dropFirst()) : tagName

          if compareVersions(latestVersion, AppVersion.current) > 0 {
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
          importError = "Failed to import database."
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

// MARK: - Reusable Components

struct SettingsCard<Content: View>: View {
  let title: String
  let icon: String
  let color: Color
  @ViewBuilder let content: Content
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
      // Header
      HStack(spacing: Design.Spacing.xs) {
        Image(systemName: icon)
          .font(.system(size: 11, weight: .semibold))
          .foregroundColor(color)

        Text(title)
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textSecondary)
          .textCase(.uppercase)
          .tracking(0.5)
      }

      // Content
      VStack(spacing: 0) {
        content
      }
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.md)
          .fill(Design.Colors.surfaceElevated(colorScheme))
      )
    }
  }
}

struct SecureModeToggle: View {
  @EnvironmentObject var appState: AppState
  @State private var isHovered = false

  var body: some View {
    Button(action: { appState.secureMode.toggle() }) {
      HStack(spacing: Design.Spacing.sm) {
        ZStack {
          RoundedRectangle(cornerRadius: 10)
            .fill(appState.secureMode ? Color.green.opacity(0.15) : Color.secondary.opacity(0.1))
            .frame(width: 40, height: 40)

          Image(systemName: appState.secureMode ? "eye.slash.fill" : "eye.fill")
            .font(.system(size: 16, weight: .medium))
            .foregroundColor(appState.secureMode ? .green : .secondary)
        }

        VStack(alignment: .leading, spacing: 2) {
          Text("Secure Mode")
            .font(Design.Typography.labelLarge)
            .foregroundColor(Design.Colors.textPrimary)

          Text(appState.secureMode ? "All amounts hidden" : "Amounts visible")
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textSecondary)
        }

        Spacer()

        HStack(spacing: 4) {
          Circle()
            .fill(appState.secureMode ? Color.green : Color.secondary.opacity(0.3))
            .frame(width: 8, height: 8)

          Text(appState.secureMode ? "On" : "Off")
            .font(Design.Typography.labelMedium)
            .foregroundColor(appState.secureMode ? .green : Design.Colors.textTertiary)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 6)
        .background(
          Capsule()
            .fill(appState.secureMode ? Color.green.opacity(0.1) : Color.secondary.opacity(0.1))
        )
      }
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(isHovered ? Color.secondary.opacity(0.05) : Color.clear)
      )
      .overlay(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .stroke(appState.secureMode ? Color.green.opacity(0.2) : Color.clear, lineWidth: 1)
      )
    }
    .buttonStyle(.plain)
    .onHover { hovering in
      withAnimation(Design.Animation.quick) {
        isHovered = hovering
      }
    }
  }
}

struct SettingsRow<Trailing: View>: View {
  let icon: String
  let title: String
  let subtitle: String
  @ViewBuilder let trailing: Trailing

  var body: some View {
    HStack(spacing: Design.Spacing.sm) {
      Image(systemName: icon)
        .font(.system(size: 14))
        .foregroundColor(Design.Colors.textSecondary)
        .frame(width: 24)

      VStack(alignment: .leading, spacing: 1) {
        Text(title)
          .font(Design.Typography.labelMedium)
          .foregroundColor(Design.Colors.textPrimary)

        Text(subtitle)
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textTertiary)
      }

      Spacer()

      trailing
    }
    .padding(.vertical, 4)
  }
}

struct SettingsActionRow: View {
  let icon: String
  let title: String
  let subtitle: String
  let color: Color
  var showSpinner: Bool = false
  let action: () -> Void

  @State private var isHovered = false

  var body: some View {
    Button(action: action) {
      HStack(spacing: Design.Spacing.sm) {
        ZStack {
          if showSpinner {
            ProgressView()
              .scaleEffect(0.6)
              .frame(width: 24, height: 24)
          } else {
            Image(systemName: icon)
              .font(.system(size: 14))
              .foregroundColor(color)
              .frame(width: 24)
          }
        }

        VStack(alignment: .leading, spacing: 1) {
          Text(title)
            .font(Design.Typography.labelMedium)
            .foregroundColor(Design.Colors.textPrimary)

          Text(subtitle)
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textTertiary)
        }

        Spacer()

        Image(systemName: "chevron.right")
          .font(.system(size: 10, weight: .semibold))
          .foregroundColor(Design.Colors.textTertiary)
      }
      .padding(.vertical, Design.Spacing.xs)
      .padding(.horizontal, Design.Spacing.xxs)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.xs)
          .fill(isHovered ? Color.secondary.opacity(0.08) : Color.clear)
      )
    }
    .buttonStyle(.plain)
    .onHover { hovering in
      withAnimation(Design.Animation.quick) {
        isHovered = hovering
      }
    }
  }
}

struct LinkButton: View {
  let title: String
  let icon: String
  let action: () -> Void

  @State private var isHovered = false

  var body: some View {
    Button(action: action) {
      HStack(spacing: 4) {
        Image(systemName: icon)
          .font(.system(size: 10))
        Text(title)
          .font(Design.Typography.labelSmall)
      }
      .foregroundColor(isHovered ? .blue : Design.Colors.textSecondary)
    }
    .buttonStyle(.plain)
    .onHover { hovering in
      withAnimation(Design.Animation.quick) {
        isHovered = hovering
      }
    }
  }
}

#Preview {
  SettingsSection()
    .padding()
    .frame(width: 400)
    .environmentObject(AppState())
}
