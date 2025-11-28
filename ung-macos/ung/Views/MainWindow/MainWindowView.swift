//
//  MainWindowView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct MainWindowView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        HStack(spacing: 0) {
            // Slack-style sidebar
            SidebarView()

            // Main content
            ContentAreaView()
        }
        .background(backgroundColor)
        .onAppear {
            Task {
                await appState.refreshDashboard()
            }
        }
    }

    private var backgroundColor: Color {
        colorScheme == .dark
            ? Color(nsColor: .windowBackgroundColor)
            : Color(nsColor: .windowBackgroundColor)
    }
}

// MARK: - Sidebar View
struct SidebarView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(spacing: 0) {
            // App header with logo
            appHeader

            // Active status banner
            if appState.isTracking || appState.pomodoroState.isActive {
                activeStatusBanner
            }

            // Navigation items
            ScrollView {
                VStack(spacing: 4) {
                    ForEach(SidebarTab.allCases) { tab in
                        if tab == .settings {
                            Spacer()
                                .frame(height: 16)
                        }
                        SidebarItem(tab: tab)
                    }
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 12)
            }

            Spacer()

            // Secure mode toggle at bottom
            secureModeToggle
        }
        .frame(width: 220)
        .background(sidebarBackground)
    }

    private var sidebarBackground: some View {
        colorScheme == .dark
            ? Color(white: 0.12)
            : Color(white: 0.95)
    }

    private var appHeader: some View {
        HStack(spacing: 10) {
            // Logo
            ZStack {
                RoundedRectangle(cornerRadius: 10)
                    .fill(
                        LinearGradient(
                            colors: [.blue, .purple],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .frame(width: 36, height: 36)

                Image(systemName: "clock.badge.checkmark")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.white)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text("UNG")
                    .font(.system(size: 15, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                Text("Freelance Toolkit")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }

            Spacer()
        }
        .padding(16)
        .background(
            colorScheme == .dark
                ? Color(white: 0.08)
                : Color(white: 0.92)
        )
    }

    private var activeStatusBanner: some View {
        HStack(spacing: 8) {
            // Pulsing dot
            Circle()
                .fill(appState.isTracking ? Color.red : (appState.pomodoroState.isBreak ? Color.green : Color.orange))
                .frame(width: 8, height: 8)

            if appState.isTracking, let session = appState.activeSession {
                VStack(alignment: .leading, spacing: 1) {
                    Text(session.project)
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.primary)
                        .lineLimit(1)

                    Text(session.formattedDuration)
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                        .foregroundColor(.secondary)
                }
            } else if appState.pomodoroState.isActive {
                VStack(alignment: .leading, spacing: 1) {
                    Text(appState.pomodoroState.statusText)
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.primary)

                    Text(appState.pomodoroState.formattedTime)
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            // Quick stop button
            Button(action: {
                if appState.isTracking {
                    Task { await appState.stopTracking() }
                } else {
                    appState.stopPomodoro()
                }
            }) {
                Image(systemName: "stop.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.white)
                    .padding(5)
                    .background(Circle().fill(Color.red.opacity(0.8)))
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(appState.isTracking
                      ? Color.red.opacity(0.15)
                      : (appState.pomodoroState.isBreak ? Color.green.opacity(0.15) : Color.orange.opacity(0.15)))
        )
        .padding(.horizontal, 12)
        .padding(.bottom, 8)
    }

    private var secureModeToggle: some View {
        Button(action: { appState.secureMode.toggle() }) {
            HStack(spacing: 8) {
                Image(systemName: appState.secureMode ? "eye.slash.fill" : "eye.fill")
                    .font(.system(size: 12))
                    .foregroundColor(appState.secureMode ? .orange : .secondary)

                Text(appState.secureMode ? "Secure Mode On" : "Secure Mode")
                    .font(.system(size: 11))
                    .foregroundColor(appState.secureMode ? .orange : .secondary)

                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
        }
        .buttonStyle(.plain)
        .background(
            appState.secureMode
                ? Color.orange.opacity(0.1)
                : Color.clear
        )
    }
}

// MARK: - Sidebar Item
struct SidebarItem: View {
    @EnvironmentObject var appState: AppState
    let tab: SidebarTab
    @Environment(\.colorScheme) var colorScheme

    var isSelected: Bool {
        appState.selectedTab == tab
    }

    var body: some View {
        Button(action: { appState.selectedTab = tab }) {
            HStack(spacing: 10) {
                // Icon
                Image(systemName: tab.icon)
                    .font(.system(size: 14, weight: isSelected ? .semibold : .regular))
                    .foregroundColor(isSelected ? tab.color : .secondary)
                    .frame(width: 24)

                // Label
                Text(tab.rawValue)
                    .font(.system(size: 13, weight: isSelected ? .medium : .regular))
                    .foregroundColor(isSelected ? .primary : .secondary)

                Spacer()

                // Badge for specific tabs
                if tab == .invoices && appState.metrics.overdueAmount > 0 {
                    Text("\(appState.invoiceCount)")
                        .font(.system(size: 10, weight: .semibold))
                        .foregroundColor(.white)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Capsule().fill(Color.red))
                }
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 8)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(isSelected
                          ? (colorScheme == .dark ? Color.white.opacity(0.1) : tab.color.opacity(0.12))
                          : Color.clear)
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Content Area
struct ContentAreaView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(spacing: 0) {
            // Header
            contentHeader

            Divider()

            // Content
            ScrollView {
                switch appState.selectedTab {
                case .dashboard:
                    MainDashboardContent()
                case .tracking:
                    TrackingContent()
                case .clients:
                    ClientsContent()
                case .contracts:
                    ContractsContent()
                case .invoices:
                    InvoicesContent()
                case .expenses:
                    ExpensesContent()
                case .pomodoro:
                    PomodoroContent()
                case .reports:
                    ReportsContent()
                case .settings:
                    SettingsContent()
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(colorScheme == .dark ? Color(white: 0.16) : Color.white)
    }

    private var contentHeader: some View {
        HStack(spacing: 16) {
            // Title
            VStack(alignment: .leading, spacing: 2) {
                Text(appState.selectedTab.rawValue)
                    .font(.system(size: 20, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                Text(headerSubtitle)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            Spacer()

            // Refresh button
            Button(action: { Task { await appState.refreshDashboard() } }) {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: 14))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
            .disabled(appState.isRefreshing)

            // Action button (context-specific)
            if let action = primaryAction {
                Button(action: action.action) {
                    HStack(spacing: 6) {
                        Image(systemName: action.icon)
                            .font(.system(size: 12))
                        Text(action.title)
                            .font(.system(size: 13, weight: .medium))
                    }
                    .foregroundColor(.white)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 8)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(action.color)
                    )
                }
                .buttonStyle(.plain)
            }
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 16)
    }

    private var headerSubtitle: String {
        switch appState.selectedTab {
        case .dashboard: return "Overview of your business"
        case .tracking: return "Manage your time sessions"
        case .clients: return "\(appState.clientCount) clients"
        case .contracts: return "\(appState.contractCount) contracts"
        case .invoices: return "\(appState.invoiceCount) invoices"
        case .expenses: return "Track your business expenses"
        case .pomodoro: return "Focus sessions completed: \(appState.pomodoroState.sessionsCompleted)"
        case .reports: return "Analytics and insights"
        case .settings: return "Configure your preferences"
        }
    }

    private var primaryAction: (title: String, icon: String, color: Color, action: () -> Void)? {
        switch appState.selectedTab {
        case .tracking:
            return ("Start Tracking", "play.fill", .green, {})
        case .clients:
            return ("Add Client", "plus", .purple, {})
        case .contracts:
            return ("New Contract", "plus", .indigo, {})
        case .invoices:
            return ("Create Invoice", "plus", .teal, {})
        case .expenses:
            return ("Log Expense", "plus", .orange, {})
        case .pomodoro:
            return appState.pomodoroState.isActive ? nil : ("Start Focus", "play.fill", .red, { appState.startPomodoro() })
        default:
            return nil
        }
    }
}

#Preview {
    MainWindowView()
        .environmentObject(AppState())
        .frame(width: 1100, height: 700)
}
