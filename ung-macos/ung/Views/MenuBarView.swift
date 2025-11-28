//
//  MenuBarView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct MenuBarView: View {
    @EnvironmentObject var appState: AppState
    let openMainWindow: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            switch appState.status {
            case .loading:
                LoadingView()
            case .cliNotInstalled:
                OnboardingView(state: .notInstalled)
            case .notInitialized:
                OnboardingView(state: .notInitialized)
            case .ready:
                CompactDashboardView(openMainWindow: openMainWindow)
            }
        }
        .frame(width: 320)
        .background(Color(nsColor: .windowBackgroundColor))
    }
}

// MARK: - Compact Dashboard View (Menu Bar)
struct CompactDashboardView: View {
    @EnvironmentObject var appState: AppState
    let openMainWindow: () -> Void

    @State private var currentTab: CompactTab = .overview
    @State private var showStartTracking = false

    enum CompactTab: String, CaseIterable {
        case overview = "Overview"
        case pomodoro = "Focus"
    }

    var body: some View {
        VStack(spacing: 0) {
            // Active status banner
            if appState.isTracking, let session = appState.activeSession {
                activeTrackingBanner(session)
            } else if appState.pomodoroState.isActive {
                pomodoroBanner
            }

            // Tab selector
            tabSelector

            // Content
            ScrollView {
                VStack(spacing: 12) {
                    switch currentTab {
                    case .overview:
                        overviewContent
                    case .pomodoro:
                        pomodoroContent
                    }
                }
                .padding(12)
            }
            .frame(maxHeight: 360)

            // Footer
            footerBar
        }
        .sheet(isPresented: $showStartTracking) {
            StartTrackingSheet()
                .environmentObject(appState)
        }
    }

    // MARK: - Active Tracking Banner
    private func activeTrackingBanner(_ session: ActiveSession) -> some View {
        HStack(spacing: 10) {
            Circle()
                .fill(Color.red)
                .frame(width: 8, height: 8)

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

            Text(session.formattedDuration)
                .font(.system(size: 14, weight: .bold, design: .monospaced))
                .foregroundColor(.white)

            Button(action: { Task { await appState.stopTracking() } }) {
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
            LinearGradient(colors: [.red, .orange], startPoint: .leading, endPoint: .trailing)
        )
    }

    // MARK: - Pomodoro Banner
    private var pomodoroBanner: some View {
        HStack(spacing: 10) {
            Circle()
                .fill(appState.pomodoroState.isBreak ? Color.green : Color.orange)
                .frame(width: 8, height: 8)

            VStack(alignment: .leading, spacing: 2) {
                Text(appState.pomodoroState.statusText)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.white)

                Text("Session \(appState.pomodoroState.sessionsCompleted + 1)")
                    .font(.system(size: 10))
                    .foregroundColor(.white.opacity(0.8))
            }

            Spacer()

            Text(appState.pomodoroState.formattedTime)
                .font(.system(size: 14, weight: .bold, design: .monospaced))
                .foregroundColor(.white)

            HStack(spacing: 4) {
                Button(action: {
                    if appState.pomodoroState.isPaused {
                        appState.resumePomodoro()
                    } else {
                        appState.pausePomodoro()
                    }
                }) {
                    Image(systemName: appState.pomodoroState.isPaused ? "play.fill" : "pause.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.white)
                        .padding(5)
                        .background(Circle().fill(Color.white.opacity(0.2)))
                }
                .buttonStyle(.plain)

                Button(action: { appState.stopPomodoro() }) {
                    Image(systemName: "stop.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.white)
                        .padding(5)
                        .background(Circle().fill(Color.white.opacity(0.2)))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(
            LinearGradient(
                colors: appState.pomodoroState.isBreak ? [.green, .mint] : [.orange, .red],
                startPoint: .leading,
                endPoint: .trailing
            )
        )
    }

    // MARK: - Tab Selector
    private var tabSelector: some View {
        HStack(spacing: 0) {
            ForEach(CompactTab.allCases, id: \.self) { tab in
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
            // Quick actions
            HStack(spacing: 8) {
                CompactActionButton(
                    icon: "play.fill",
                    title: "Track",
                    color: .green,
                    disabled: appState.isTracking
                ) {
                    showStartTracking = true
                }

                CompactActionButton(
                    icon: "brain.head.profile",
                    title: "Focus",
                    color: .orange,
                    disabled: appState.pomodoroState.isActive
                ) {
                    currentTab = .pomodoro
                }

                CompactActionButton(
                    icon: "rectangle.expand.vertical",
                    title: "Expand",
                    color: .blue,
                    disabled: false
                ) {
                    openMainWindow()
                }
            }

            // Metrics
            HStack(spacing: 8) {
                CompactMetricCard(
                    title: "This Week",
                    value: appState.formatHours(appState.metrics.weeklyHours),
                    icon: "clock.fill",
                    color: .blue
                )

                CompactMetricCard(
                    title: "Pending",
                    value: appState.formatCurrency(appState.metrics.pendingAmount),
                    icon: "hourglass",
                    color: .orange
                )
            }

            // Recent sessions
            if !appState.recentSessions.isEmpty {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Recent")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.secondary)

                    ForEach(appState.recentSessions.prefix(3)) { session in
                        HStack {
                            Circle()
                                .fill(Color.blue.opacity(0.3))
                                .frame(width: 6, height: 6)

                            Text(session.project)
                                .font(.system(size: 11))
                                .foregroundColor(.primary)
                                .lineLimit(1)

                            Spacer()

                            Text(appState.secureMode ? "**:**" : session.duration)
                                .font(.system(size: 10, design: .monospaced))
                                .foregroundColor(.secondary)
                        }
                        .padding(.vertical, 4)
                    }
                }
                .padding(10)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color(nsColor: .controlBackgroundColor))
                )
            }
        }
    }

    // MARK: - Pomodoro Content
    private var pomodoroContent: some View {
        VStack(spacing: 16) {
            // Timer display
            ZStack {
                Circle()
                    .stroke(Color.gray.opacity(0.2), lineWidth: 8)
                    .frame(width: 140, height: 140)

                Circle()
                    .trim(from: 0, to: appState.pomodoroState.progress)
                    .stroke(
                        LinearGradient(
                            colors: appState.pomodoroState.isBreak ? [.green, .mint] : [.red, .orange],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        style: StrokeStyle(lineWidth: 8, lineCap: .round)
                    )
                    .frame(width: 140, height: 140)
                    .rotationEffect(.degrees(-90))

                VStack(spacing: 4) {
                    Text(appState.pomodoroState.formattedTime)
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .monospacedDigit()

                    Text(appState.pomodoroState.statusText)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }
            }

            // Controls
            if appState.pomodoroState.isActive {
                HStack(spacing: 12) {
                    Button(action: {
                        if appState.pomodoroState.isPaused {
                            appState.resumePomodoro()
                        } else {
                            appState.pausePomodoro()
                        }
                    }) {
                        Image(systemName: appState.pomodoroState.isPaused ? "play.fill" : "pause.fill")
                            .font(.system(size: 14))
                            .frame(width: 40, height: 40)
                            .background(Circle().fill(Color(nsColor: .controlBackgroundColor)))
                    }
                    .buttonStyle(.plain)

                    Button(action: { appState.skipPomodoro() }) {
                        Image(systemName: "forward.fill")
                            .font(.system(size: 14))
                            .frame(width: 40, height: 40)
                            .background(Circle().fill(Color(nsColor: .controlBackgroundColor)))
                    }
                    .buttonStyle(.plain)

                    Button(action: { appState.stopPomodoro() }) {
                        Image(systemName: "stop.fill")
                            .font(.system(size: 14))
                            .foregroundColor(.red)
                            .frame(width: 40, height: 40)
                            .background(Circle().stroke(Color.red.opacity(0.3), lineWidth: 1))
                    }
                    .buttonStyle(.plain)
                }
            } else {
                Button(action: { appState.startPomodoro() }) {
                    HStack(spacing: 8) {
                        Image(systemName: "play.fill")
                        Text("Start Focus")
                    }
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.white)
                    .padding(.horizontal, 24)
                    .padding(.vertical, 12)
                    .background(
                        RoundedRectangle(cornerRadius: 10)
                            .fill(LinearGradient(colors: [.red, .orange], startPoint: .leading, endPoint: .trailing))
                    )
                }
                .buttonStyle(.plain)
            }

            // Stats
            HStack(spacing: 20) {
                VStack(spacing: 2) {
                    Text("\(appState.pomodoroState.sessionsCompleted)")
                        .font(.system(size: 18, weight: .bold))
                    Text("Sessions")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }

                Divider().frame(height: 30)

                VStack(spacing: 2) {
                    Text("\(appState.pomodoroState.sessionsCompleted * appState.pomodoroState.workMinutes)m")
                        .font(.system(size: 18, weight: .bold))
                    Text("Focus Time")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }
            }
        }
    }

    // MARK: - Footer Bar
    private var footerBar: some View {
        HStack {
            Button(action: { Task { await appState.refreshDashboard() } }) {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
            .disabled(appState.isRefreshing)

            if appState.isRefreshing {
                ProgressView().scaleEffect(0.6)
            }

            Spacer()

            // Open main window button
            Button(action: openMainWindow) {
                HStack(spacing: 4) {
                    Image(systemName: "macwindow")
                        .font(.system(size: 10))
                    Text("Open App")
                        .font(.system(size: 10))
                }
                .foregroundColor(.blue)
            }
            .buttonStyle(.plain)

            Spacer()

            // Secure mode indicator
            if appState.secureMode {
                Image(systemName: "eye.slash.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.orange)
            }

            Button(action: { NSApplication.shared.terminate(nil) }) {
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

// MARK: - Compact Action Button
struct CompactActionButton: View {
    let icon: String
    let title: String
    let color: Color
    let disabled: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 16))
                    .foregroundColor(disabled ? .secondary : color)

                Text(title)
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(disabled ? .secondary : .primary)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 12)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(disabled ? Color.gray.opacity(0.1) : color.opacity(0.1))
            )
        }
        .buttonStyle(.plain)
        .disabled(disabled)
    }
}

// MARK: - Compact Metric Card
struct CompactMetricCard: View {
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
                .font(.system(size: 14, weight: .bold, design: .rounded))
                .foregroundColor(.primary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

#Preview {
    MenuBarView(openMainWindow: {})
        .environmentObject(AppState())
}
