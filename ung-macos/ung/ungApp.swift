//
//  ungApp.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

@main
struct ungApp: App {
    @StateObject private var appState = AppState()
    @Environment(\.openWindow) var openWindow

    var body: some Scene {
        // Menu Bar
        MenuBarExtra {
            MenuBarView(openMainWindow: { openWindow(id: "main-window") })
                .environmentObject(appState)
        } label: {
            MenuBarLabel()
                .environmentObject(appState)
        }
        .menuBarExtraStyle(.window)

        // Main Window (Slack-style)
        Window("UNG", id: "main-window") {
            MainWindowView()
                .environmentObject(appState)
                .frame(minWidth: 900, minHeight: 600)
        }
        .windowStyle(.hiddenTitleBar)
        .defaultSize(width: 1100, height: 700)
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
                Image(systemName: appState.pomodoroState.isBreak ? "cup.and.saucer.fill" : "brain.head.profile")
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
