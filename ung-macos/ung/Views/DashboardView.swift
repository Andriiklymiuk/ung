//
//  DashboardView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct DashboardView: View {
    @EnvironmentObject var appState: AppState
    @State private var currentTab: DashboardTab = .overview
    @State private var showStartTracking = false
    @State private var showLogExpense = false
    @State private var showCreateInvoice = false

    enum DashboardTab: String, CaseIterable {
        case overview = "Overview"
        case activity = "Activity"
        case settings = "Settings"
    }

    var body: some View {
        VStack(spacing: 0) {
            // Active tracking banner (if tracking)
            if appState.isTracking, let session = appState.activeSession {
                ActiveTrackingBanner(session: session)
            }

            // Tab selector
            tabSelector

            // Content based on tab
            ScrollView {
                VStack(spacing: 12) {
                    switch currentTab {
                    case .overview:
                        overviewContent
                    case .activity:
                        activityContent
                    case .settings:
                        settingsContent
                    }
                }
                .padding(12)
            }
            .frame(maxHeight: appState.isTracking ? 380 : 420)

            // Footer
            footerBar
        }
        .sheet(isPresented: $showStartTracking) {
            StartTrackingSheet()
                .environmentObject(appState)
        }
        .sheet(isPresented: $showLogExpense) {
            LogExpenseSheet()
                .environmentObject(appState)
        }
        .sheet(isPresented: $showCreateInvoice) {
            CreateInvoiceSheet()
                .environmentObject(appState)
        }
    }

    // MARK: - Tab Selector
    private var tabSelector: some View {
        HStack(spacing: 0) {
            ForEach(DashboardTab.allCases, id: \.self) { tab in
                Button(action: { withAnimation(.spring(response: 0.25)) { currentTab = tab } }) {
                    Text(tab.rawValue)
                        .font(.system(size: 12, weight: currentTab == tab ? .semibold : .regular))
                        .foregroundColor(currentTab == tab ? .primary : .secondary)
                        .padding(.vertical, 8)
                        .frame(maxWidth: .infinity)
                        .background(
                            Group {
                                if currentTab == tab {
                                    RoundedRectangle(cornerRadius: 6)
                                        .fill(Color(nsColor: .controlBackgroundColor))
                                        .shadow(color: .black.opacity(0.05), radius: 1, y: 1)
                                }
                            }
                        )
                }
                .buttonStyle(.plain)
            }
        }
        .padding(4)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(nsColor: .separatorColor).opacity(0.3))
        )
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
    }

    // MARK: - Overview Content
    private var overviewContent: some View {
        VStack(spacing: 12) {
            // Setup checklist (if incomplete)
            if !appState.setupStatus.isComplete {
                SetupChecklistCard()
            }

            // Quick actions
            QuickActionsSection(
                onStartTracking: { showStartTracking = true },
                onLogExpense: { showLogExpense = true },
                onCreateInvoice: { showCreateInvoice = true }
            )

            // Metrics cards
            MetricsSection()

            // Weekly progress
            WeeklyProgressCard()
        }
    }

    // MARK: - Activity Content
    private var activityContent: some View {
        VStack(spacing: 12) {
            // Recent sessions
            RecentSessionsCard()

            // Recent invoices
            RecentInvoicesCard()

            // Recent expenses
            RecentExpensesCard()
        }
    }

    // MARK: - Settings Content
    private var settingsContent: some View {
        VStack(spacing: 12) {
            SettingsSection()
        }
    }

    // MARK: - Footer Bar
    private var footerBar: some View {
        HStack {
            Button(action: {
                Task { await appState.refreshDashboard() }
            }) {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
            .disabled(appState.isRefreshing)

            Spacer()

            if appState.isRefreshing {
                ProgressView()
                    .scaleEffect(0.6)
            }

            Spacer()

            Button(action: {
                NSApplication.shared.terminate(nil)
            }) {
                Text("Quit")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(Color(nsColor: .separatorColor).opacity(0.2))
    }
}

// MARK: - Active Tracking Banner
struct ActiveTrackingBanner: View {
    @EnvironmentObject var appState: AppState
    let session: ActiveSession
    @State private var isPulsing = false

    var body: some View {
        HStack(spacing: 10) {
            // Pulsing indicator
            Circle()
                .fill(Color.red)
                .frame(width: 8, height: 8)
                .scaleEffect(isPulsing ? 1.2 : 1.0)
                .animation(.easeInOut(duration: 0.8).repeatForever(autoreverses: true), value: isPulsing)

            VStack(alignment: .leading, spacing: 2) {
                Text(session.project)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.white)
                    .lineLimit(1)

                if !session.client.isEmpty {
                    Text(session.client)
                        .font(.system(size: 10))
                        .foregroundColor(.white.opacity(0.8))
                        .lineLimit(1)
                }
            }

            Spacer()

            // Timer
            Text(session.formattedDuration)
                .font(.system(size: 14, weight: .bold, design: .monospaced))
                .foregroundColor(.white)

            // Stop button
            Button(action: {
                Task { await appState.stopTracking() }
            }) {
                Image(systemName: "stop.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.white)
                    .padding(6)
                    .background(Circle().fill(Color.white.opacity(0.2)))
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(
            LinearGradient(
                colors: [.red, .orange],
                startPoint: .leading,
                endPoint: .trailing
            )
        )
        .onAppear { isPulsing = true }
    }
}

// MARK: - Setup Checklist Card
struct SetupChecklistCard: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Image(systemName: "checklist")
                    .font(.system(size: 12))
                    .foregroundColor(.orange)
                Text("Complete Setup")
                    .font(.system(size: 12, weight: .semibold))
                Spacer()
            }

            VStack(spacing: 6) {
                ChecklistItem(
                    title: "Company Profile",
                    isComplete: appState.setupStatus.hasCompany,
                    action: {}
                )
                ChecklistItem(
                    title: "First Client",
                    isComplete: appState.setupStatus.hasClient,
                    action: {}
                )
                ChecklistItem(
                    title: "First Contract",
                    isComplete: appState.setupStatus.hasContract,
                    action: {}
                )
            }

            if !appState.setupStatus.isComplete {
                Text("Next: \(appState.setupStatus.nextStep)")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.orange.opacity(0.1))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(Color.orange.opacity(0.3), lineWidth: 1)
                )
        )
    }
}

struct ChecklistItem: View {
    let title: String
    let isComplete: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                Image(systemName: isComplete ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 12))
                    .foregroundColor(isComplete ? .green : .secondary)

                Text(title)
                    .font(.system(size: 11))
                    .foregroundColor(isComplete ? .secondary : .primary)
                    .strikethrough(isComplete)

                Spacer()
            }
        }
        .buttonStyle(.plain)
        .disabled(isComplete)
    }
}

// MARK: - Quick Actions Section
struct QuickActionsSection: View {
    let onStartTracking: () -> Void
    let onLogExpense: () -> Void
    let onCreateInvoice: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Quick Actions")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack(spacing: 8) {
                QuickActionButton(
                    icon: "play.fill",
                    title: "Track",
                    color: .green,
                    action: onStartTracking
                )

                QuickActionButton(
                    icon: "dollarsign.circle.fill",
                    title: "Expense",
                    color: .orange,
                    action: onLogExpense
                )

                QuickActionButton(
                    icon: "doc.text.fill",
                    title: "Invoice",
                    color: .blue,
                    action: onCreateInvoice
                )
            }
        }
    }
}

struct QuickActionButton: View {
    let icon: String
    let title: String
    let color: Color
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 16))
                    .foregroundColor(color)

                Text(title)
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(.primary)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 12)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(color.opacity(0.1))
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Metrics Section
struct MetricsSection: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("This Month")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack(spacing: 8) {
                MetricCard(
                    title: "Revenue",
                    value: appState.formatCurrency(appState.metrics.monthlyRevenue),
                    icon: "chart.line.uptrend.xyaxis",
                    color: .green
                )

                MetricCard(
                    title: "Pending",
                    value: appState.formatCurrency(appState.metrics.pendingAmount),
                    icon: "clock.fill",
                    color: .orange
                )
            }

            // Counts row
            HStack(spacing: 8) {
                CountBadge(count: appState.clientCount, label: "Clients", icon: "person.2.fill")
                CountBadge(count: appState.contractCount, label: "Contracts", icon: "doc.text.fill")
                CountBadge(count: appState.invoiceCount, label: "Invoices", icon: "doc.fill")
            }
        }
    }
}

struct MetricCard: View {
    let title: String
    let value: String
    let icon: String
    let color: Color

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 4) {
                Image(systemName: icon)
                    .font(.system(size: 10))
                    .foregroundColor(color)
                Text(title)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }

            Text(value)
                .font(.system(size: 16, weight: .bold, design: .rounded))
                .foregroundColor(.primary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

struct CountBadge: View {
    let count: Int
    let label: String
    let icon: String

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.system(size: 10))
                .foregroundColor(.secondary)

            Text("\(count)")
                .font(.system(size: 12, weight: .semibold))
                .foregroundColor(.primary)

            Text(label)
                .font(.system(size: 10))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 8)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

// MARK: - Weekly Progress Card
struct WeeklyProgressCard: View {
    @EnvironmentObject var appState: AppState

    var progress: Double {
        guard appState.metrics.weeklyTarget > 0 else { return 0 }
        return min(appState.metrics.weeklyHours / appState.metrics.weeklyTarget, 1.0)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("Weekly Progress")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)

                Spacer()

                // Streak badge
                if appState.metrics.trackingStreak > 0 {
                    HStack(spacing: 3) {
                        Image(systemName: "flame.fill")
                            .font(.system(size: 9))
                            .foregroundColor(.orange)
                        Text("\(appState.metrics.trackingStreak) days")
                            .font(.system(size: 9, weight: .medium))
                            .foregroundColor(.orange)
                    }
                    .padding(.horizontal, 6)
                    .padding(.vertical, 3)
                    .background(Color.orange.opacity(0.15))
                    .cornerRadius(6)
                }
            }

            // Progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.blue.opacity(0.2))
                        .frame(height: 8)

                    RoundedRectangle(cornerRadius: 4)
                        .fill(
                            LinearGradient(
                                colors: [.blue, .purple],
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )
                        .frame(width: geometry.size.width * progress, height: 8)
                }
            }
            .frame(height: 8)

            // Hours text
            HStack {
                Text(appState.formatHours(appState.metrics.weeklyHours))
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.primary)

                Text("/ \(Int(appState.metrics.weeklyTarget))h target")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)

                Spacer()

                Text("\(Int(progress * 100))%")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.blue)
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

#Preview {
    DashboardView()
        .environmentObject(AppState())
        .frame(width: 320)
}
