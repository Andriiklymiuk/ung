//
//  ungApp.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI
import UserNotifications

@main
struct ungApp: App {
  @StateObject private var appState = AppState()

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
    } label: {
      MenuBarLabel()
        .environmentObject(appState)
    }
    .menuBarExtraStyle(.window)
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
