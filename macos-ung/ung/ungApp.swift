//
//  ungApp.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI
import UniformTypeIdentifiers
import UserNotifications

@main
struct ungApp: App {
  @StateObject private var appState = AppState()
  @StateObject private var themeManager = ThemeManager.shared
  @State private var importedFileURL: URL?
  @State private var showImportPasswordPrompt = false
  @State private var showImportConfirmation = false

  init() {
    // Set up notification delegate
    UNUserNotificationCenter.current().delegate = NotificationDelegate.shared

    // Configure notification action handlers
    NotificationDelegate.shared.onPomodoroAction = { action in
      Task { @MainActor in
        let appState = AppState()
        switch action {
        case "START_BREAK":
          // Already handled by pomodoro completion
          break
        case "SKIP_BREAK":
          appState.skipPomodoro()
        case "START_WORK":
          if !appState.pomodoroState.isActive {
            appState.startPomodoro()
          }
        default:
          break
        }
      }
    }

    NotificationDelegate.shared.onTrackingAction = { action in
      Task { @MainActor in
        let appState = AppState()
        switch action {
        case "STOP_TRACKING":
          await appState.stopTracking()
        case "CONTINUE_TRACKING":
          // Just dismiss - reschedule reminder
          NotificationService.shared.scheduleTrackingReminder()
        default:
          break
        }
      }
    }

    // Request notification authorization on first launch
    Task {
      _ = await NotificationService.shared.requestAuthorization()
    }
  }

  var body: some Scene {
    // Main Window (shows on launch like a regular macOS app)
    WindowGroup("UNG", id: "main-window") {
      MainWindowView()
        .environmentObject(appState)
        .environmentObject(themeManager)
        .onOpenURL { url in
          handleOpenURL(url)
        }
        .sheet(isPresented: $showImportPasswordPrompt) {
          DatabaseImportPasswordView(
            fileURL: importedFileURL,
            onImport: { password in
              Task {
                await importEncryptedDatabase(url: importedFileURL!, password: password)
              }
            },
            onCancel: {
              showImportPasswordPrompt = false
              importedFileURL = nil
            }
          )
          .environmentObject(appState)
        }
        .alert("Import Database?", isPresented: $showImportConfirmation) {
          Button("Import & Merge") {
            Task {
              await importPlainDatabase(url: importedFileURL!)
            }
          }
          Button("Cancel", role: .cancel) {
            importedFileURL = nil
          }
        } message: {
          Text("Data from the imported database will be merged with your existing data. Existing records will be kept, new records will be added.")
        }
    }
    .windowStyle(.hiddenTitleBar)
    .windowResizability(.contentMinSize)
    .defaultSize(width: 1100, height: 750)
    .commands {
      // Replace default New command with our tracking
      CommandGroup(replacing: .newItem) {
        Button("Start Tracking") {
          appState.selectedTab = .tracking
        }
        .keyboardShortcut("n", modifiers: [.command])
        .disabled(appState.isTracking)

        Button("Stop Tracking") {
          Task { await appState.stopTracking() }
        }
        .keyboardShortcut(".", modifiers: [.command])
        .disabled(!appState.isTracking)

        Divider()

        Button("New Client") {
          appState.selectedTab = .clients
        }
        .keyboardShortcut("n", modifiers: [.command, .shift])

        Button("New Invoice") {
          appState.selectedTab = .invoices
        }
        .keyboardShortcut("i", modifiers: [.command, .shift])
      }

      // Search command
      CommandGroup(after: .textEditing) {
        Button("Search Everything") {
          appState.showGlobalSearch = true
        }
        .keyboardShortcut("k", modifiers: [.command])
      }

      // Navigation commands
      CommandMenu("Navigate") {
        Button("Dashboard") {
          appState.selectedTab = .dashboard
        }
        .keyboardShortcut("1", modifiers: [.command])

        Button("Time Tracking") {
          appState.selectedTab = .tracking
        }
        .keyboardShortcut("2", modifiers: [.command])

        Button("Clients") {
          appState.selectedTab = .clients
        }
        .keyboardShortcut("3", modifiers: [.command])

        Button("Contracts") {
          appState.selectedTab = .contracts
        }
        .keyboardShortcut("4", modifiers: [.command])

        Button("Invoices") {
          appState.selectedTab = .invoices
        }
        .keyboardShortcut("5", modifiers: [.command])

        Button("Expenses") {
          appState.selectedTab = .expenses
        }
        .keyboardShortcut("6", modifiers: [.command])

        Button("Focus Timer") {
          appState.selectedTab = .pomodoro
        }
        .keyboardShortcut("7", modifiers: [.command])

        Button("Reports") {
          appState.selectedTab = .reports
        }
        .keyboardShortcut("8", modifiers: [.command])

        Button("Settings") {
          appState.selectedTab = .settings
        }
        .keyboardShortcut(",", modifiers: [.command])
      }

      // Focus Timer commands
      CommandMenu("Focus") {
        Button("Start Focus Session") {
          appState.startPomodoro()
        }
        .keyboardShortcut("f", modifiers: [.command, .shift])
        .disabled(appState.pomodoroState.isActive)

        Button("Stop Focus Session") {
          appState.stopPomodoro()
        }
        .keyboardShortcut("f", modifiers: [.command, .shift, .option])
        .disabled(!appState.pomodoroState.isActive)

        Button("Skip to Next") {
          appState.skipPomodoro()
        }
        .keyboardShortcut("s", modifiers: [.command, .shift])
        .disabled(!appState.pomodoroState.isActive)
      }

      // Data commands
      CommandGroup(after: .importExport) {
        Divider()

        Button("Refresh Data") {
          Task { await appState.refreshDashboard() }
        }
        .keyboardShortcut("r", modifiers: [.command])
        .disabled(appState.isRefreshing)

        Button("Sync with iCloud") {
          appState.checkICloudSync()
        }
        .keyboardShortcut("s", modifiers: [.command, .option])
        .disabled(!appState.iCloudEnabled)
      }

      // Security commands
      CommandGroup(after: .windowArrangement) {
        Divider()

        Button(appState.secureMode ? "Disable Secure Mode" : "Enable Secure Mode") {
          appState.secureMode.toggle()
          let message = appState.secureMode
            ? "Secure Mode Enabled - Sensitive data hidden"
            : "Secure Mode Disabled - All data visible"
          appState.showToastMessage(message, type: appState.secureMode ? .warning : .info)
        }
        .keyboardShortcut("h", modifiers: [.command, .shift])

        if appState.appLockEnabled {
          Button("Lock App") {
            appState.lockApp()
          }
          .keyboardShortcut("l", modifiers: [.command, .control])
        }
      }
    }

    // Menu Bar Extra (still accessible from menu bar)
    MenuBarExtra {
      MenuBarView(openMainWindow: {
        NSApp.activate(ignoringOtherApps: true)
        if let window = NSApp.windows.first(where: { $0.title == "UNG" }) {
          window.makeKeyAndOrderFront(nil)
        }
      })
      .environmentObject(appState)
      .environmentObject(themeManager)
    } label: {
      MenuBarLabel()
        .environmentObject(appState)
    }
    .menuBarExtraStyle(.window)
  }

  // MARK: - File Import Handling

  /// Handle opening a URL (file or deep link)
  private func handleOpenURL(_ url: URL) {
    // Check if it's a file URL
    guard url.isFileURL else {
      // Handle deep links (ung:// scheme)
      handleDeepLink(url)
      return
    }

    // Start accessing security-scoped resource
    guard url.startAccessingSecurityScopedResource() else {
      appState.showError("Cannot access the file. Permission denied.")
      return
    }

    importedFileURL = url

    // Check if it's an encrypted database
    let pathExtension = url.pathExtension.lowercased()
    let fileName = url.lastPathComponent.lowercased()

    if fileName.hasSuffix(".db.encrypted") || pathExtension == "encrypted" {
      // Encrypted database - need password
      showImportPasswordPrompt = true
    } else if pathExtension == "db" || pathExtension == "sqlite" || pathExtension == "sqlite3" {
      // Plain database - confirm import
      showImportConfirmation = true
    } else {
      url.stopAccessingSecurityScopedResource()
      appState.showError("Unsupported file type: \(pathExtension)")
      importedFileURL = nil
    }
  }

  /// Handle deep links (ung:// scheme)
  private func handleDeepLink(_ url: URL) {
    guard let host = url.host else { return }

    switch host {
    case "tracking":
      appState.selectedTab = .tracking
    case "invoices":
      appState.selectedTab = .invoices
    case "clients":
      appState.selectedTab = .clients
    case "settings":
      appState.selectedTab = .settings
    default:
      break
    }
  }

  /// Import an encrypted database file with merge strategy
  @MainActor
  private func importEncryptedDatabase(url: URL, password: String) async {
    defer {
      url.stopAccessingSecurityScopedResource()
      importedFileURL = nil
      showImportPasswordPrompt = false
    }

    do {
      // Use merge import to combine data instead of replacing
      let result = try await appState.database.mergeImportEncryptedDatabase(from: url, password: password)

      // Save password to keychain for future use
      _ = appState.saveEncryptionPasswordToKeychain(password)

      // Refresh dashboard to show new data
      await appState.refreshDashboard()

      // Show import result
      var message = result.summary
      if !result.skippedSummary.isEmpty {
        message += "\n" + result.skippedSummary
      }
      appState.showToastMessage(message, type: .success)
    } catch {
      appState.showError("Failed to import database: \(error.localizedDescription)")
    }
  }

  /// Import a plain (unencrypted) database file with merge strategy
  @MainActor
  private func importPlainDatabase(url: URL) async {
    defer {
      url.stopAccessingSecurityScopedResource()
      importedFileURL = nil
      showImportConfirmation = false
    }

    do {
      // Use merge import to combine data instead of replacing
      let result = try await appState.database.mergeImportDatabase(from: url)

      // Refresh dashboard to show new data
      await appState.refreshDashboard()

      // Show import result
      var message = result.summary
      if !result.skippedSummary.isEmpty {
        message += "\n" + result.skippedSummary
      }
      appState.showToastMessage(message, type: .success)
    } catch {
      appState.showError("Failed to import database: \(error.localizedDescription)")
    }
  }
}

// MARK: - Menu Bar Label (shows timer when active)
struct MenuBarLabel: View {
  @EnvironmentObject var appState: AppState

  var body: some View {
    HStack(spacing: 4) {
      // Icon changes based on state
      if appState.isTracking {
        Image(systemName: "record.circle.fill")
          .symbolRenderingMode(.palette)
          .foregroundStyle(.red, .primary)
      } else if appState.pomodoroState.isActive {
        Image(
          systemName: appState.pomodoroState.isBreak ? "cup.and.saucer.fill" : "brain.head.profile"
        )
        .symbolRenderingMode(.hierarchical)
        .foregroundStyle(appState.pomodoroState.isBreak ? .green : .orange)
      } else {
        Image(systemName: "clock.badge.checkmark")
          .symbolRenderingMode(.hierarchical)
      }

      // Live timer display in menu bar
      if appState.isTracking, let session = appState.activeSession {
        Text(session.formattedDuration)
          .font(.system(size: 11, weight: .medium, design: .monospaced))
          .monospacedDigit()
      } else if appState.pomodoroState.isActive {
        Text(appState.pomodoroState.formattedTime)
          .font(.system(size: 11, weight: .medium, design: .monospaced))
          .monospacedDigit()
      }
    }
  }
}
