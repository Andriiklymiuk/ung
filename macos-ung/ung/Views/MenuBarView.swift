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
  @Environment(\.dismiss) private var dismiss
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
      ScrollView(showsIndicators: false) {
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
    HStack(spacing: Design.Spacing.xs) {
      Circle()
        .fill(Color.white)
        .frame(width: 8, height: 8)

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(session.project)
          .font(Design.Typography.labelMedium)
          .foregroundColor(.white)
          .lineLimit(1)

        if !session.client.isEmpty {
          Text(session.client)
            .font(Design.Typography.labelSmall)
            .foregroundColor(.white.opacity(0.8))
            .lineLimit(1)
        }
      }

      Spacer()

      Text(session.formattedDuration)
        .font(Design.Typography.monoSmall)
        .foregroundColor(.white)

      Button(action: { Task { await appState.stopTracking() } }) {
        Image(systemName: "stop.fill")
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(.white)
          .padding(Design.Spacing.xxs)
          .background(Circle().fill(Color.white.opacity(0.2)))
      }
      .buttonStyle(.plain)
      .accessibleButton(label: "Stop tracking", hint: "Stops the current time tracking session")
    }
    .padding(.horizontal, Design.Spacing.sm)
    .padding(.vertical, Design.Spacing.xs)
    .background(
      LinearGradient(colors: [Design.Colors.error, Design.Colors.warning], startPoint: .leading, endPoint: .trailing)
    )
  }

  // MARK: - Pomodoro Banner
  private var pomodoroBanner: some View {
    HStack(spacing: Design.Spacing.xs) {
      Circle()
        .fill(Color.white)
        .frame(width: 8, height: 8)

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(appState.pomodoroState.statusText)
          .font(Design.Typography.labelMedium)
          .foregroundColor(.white)

        Text("Session \(appState.pomodoroState.sessionsCompleted + 1)")
          .font(Design.Typography.labelSmall)
          .foregroundColor(.white.opacity(0.8))
      }

      Spacer()

      Text(appState.pomodoroState.formattedTime)
        .font(Design.Typography.monoSmall)
        .foregroundColor(.white)

      HStack(spacing: Design.Spacing.xxs) {
        Button(action: {
          if appState.pomodoroState.isPaused {
            appState.resumePomodoro()
          } else {
            appState.pausePomodoro()
          }
        }) {
          Image(systemName: appState.pomodoroState.isPaused ? "play.fill" : "pause.fill")
            .font(.system(size: Design.IconSize.xs))
            .foregroundColor(.white)
            .padding(Design.Spacing.xxs)
            .background(Circle().fill(Color.white.opacity(0.2)))
        }
        .buttonStyle(.plain)
        .accessibleButton(label: appState.pomodoroState.isPaused ? "Resume focus" : "Pause focus")

        Button(action: { appState.stopPomodoro() }) {
          Image(systemName: "stop.fill")
            .font(.system(size: Design.IconSize.xs))
            .foregroundColor(.white)
            .padding(Design.Spacing.xxs)
            .background(Circle().fill(Color.white.opacity(0.2)))
        }
        .buttonStyle(.plain)
        .accessibleButton(label: "Stop focus session")
      }
    }
    .padding(.horizontal, Design.Spacing.sm)
    .padding(.vertical, Design.Spacing.xs)
    .background(
      LinearGradient(
        colors: appState.pomodoroState.isBreak ? [Design.Colors.success, Design.Colors.success.opacity(0.7)] : [Design.Colors.warning, Design.Colors.error],
        startPoint: .leading,
        endPoint: .trailing
      )
    )
  }

  // MARK: - Tab Selector
  private var tabSelector: some View {
    HStack(spacing: 0) {
      ForEach(CompactTab.allCases, id: \.self) { tab in
        Text(tab.rawValue)
          .font(currentTab == tab ? Design.Typography.labelMedium : Design.Typography.bodySmall)
          .foregroundColor(currentTab == tab ? Design.Colors.textPrimary : Design.Colors.textSecondary)
          .padding(.vertical, Design.Spacing.xs)
          .frame(maxWidth: .infinity)
          .background(
            Group {
              if currentTab == tab {
                RoundedRectangle(cornerRadius: Design.Radius.xs)
                  .fill(Design.Colors.controlBackground)
                  .shadow(color: Design.Shadow.sm.color, radius: Design.Shadow.sm.radius, y: Design.Shadow.sm.y)
              }
            }
          )
          .contentShape(Rectangle())
          .onTapGesture {
            withAnimation(Design.Animation.smooth) {
              currentTab = tab
            }
          }
          .accessibilityAddTraits(currentTab == tab ? .isSelected : [])
      }
    }
    .padding(Design.Spacing.xxs)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.sm)
        .fill(Design.Colors.divider)
    )
    .padding(.horizontal, Design.Spacing.sm)
    .padding(.vertical, Design.Spacing.xs)
    .accessibilityElement(children: .contain)
    .accessibilityLabel("Tab selector")
  }

  // MARK: - Overview Content
  private var overviewContent: some View {
    VStack(spacing: Design.Spacing.sm) {
      // Quick actions
      HStack(spacing: Design.Spacing.xs) {
        CompactActionButton(
          icon: "play.fill",
          title: "Track",
          color: Design.Colors.success,
          disabled: appState.isTracking
        ) {
          showStartTracking = true
        }

        CompactActionButton(
          icon: "brain.head.profile",
          title: "Focus",
          color: Design.Colors.warning,
          disabled: appState.pomodoroState.isActive
        ) {
          currentTab = .pomodoro
        }

        CompactActionButton(
          icon: "rectangle.expand.vertical",
          title: "Expand",
          color: Design.Colors.brand,
          disabled: false
        ) {
          openMainWindow()
          dismiss()
        }
      }

      // Metrics
      HStack(spacing: Design.Spacing.xs) {
        CompactMetricCard(
          title: "This Week",
          value: appState.formatHours(appState.metrics.weeklyHours),
          icon: "clock.fill",
          color: Design.Colors.brand
        )

        CompactMetricCard(
          title: "Pending",
          value: appState.formatCurrency(appState.metrics.pendingAmount),
          icon: "hourglass",
          color: Design.Colors.warning
        )
      }

      // Recent sessions
      if !appState.recentSessions.isEmpty {
        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
          Text("Recent")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)

          ForEach(appState.recentSessions.prefix(3)) { session in
            HStack {
              Circle()
                .fill(Design.Colors.brand.opacity(0.3))
                .frame(width: 6, height: 6)

              Text(session.project)
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textPrimary)
                .lineLimit(1)

              Spacer()

              Text(appState.secureMode ? "**:**" : session.duration)
                .font(.system(size: 10, design: .monospaced))
                .foregroundColor(Design.Colors.textSecondary)
            }
            .padding(.vertical, Design.Spacing.xxs)
          }
        }
        .padding(Design.Spacing.xs)
        .background(
          RoundedRectangle(cornerRadius: Design.Radius.sm)
            .fill(Design.Colors.controlBackground)
        )
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Recent sessions")
      }
    }
  }

  // MARK: - Pomodoro Content
  private var pomodoroContent: some View {
    VStack(spacing: Design.Spacing.md) {
      // Timer display
      ZStack {
        Circle()
          .stroke(Design.Colors.border, lineWidth: 8)
          .frame(width: 140, height: 140)

        Circle()
          .trim(from: 0, to: appState.pomodoroState.progress)
          .stroke(
            LinearGradient(
              colors: appState.pomodoroState.isBreak ? [Design.Colors.success, Design.Colors.success.opacity(0.7)] : [Design.Colors.error, Design.Colors.warning],
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            ),
            style: StrokeStyle(lineWidth: 8, lineCap: .round)
          )
          .frame(width: 140, height: 140)
          .rotationEffect(.degrees(-90))
          .animation(Design.Animation.smooth, value: appState.pomodoroState.progress)

        VStack(spacing: Design.Spacing.xxs) {
          Text(appState.pomodoroState.formattedTime)
            .font(Design.Typography.displaySmall)
            .monospacedDigit()

          Text(appState.pomodoroState.statusText)
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textSecondary)
        }
      }
      .accessibilityElement(children: .combine)
      .accessibilityLabel("Timer: \(appState.pomodoroState.formattedTime), \(appState.pomodoroState.statusText)")

      // Controls - fixed height to prevent jumping
      Group {
        if appState.pomodoroState.isActive {
          HStack(spacing: Design.Spacing.sm) {
            Button(action: {
              if appState.pomodoroState.isPaused {
                appState.resumePomodoro()
              } else {
                appState.pausePomodoro()
              }
            }) {
              Image(systemName: appState.pomodoroState.isPaused ? "play.fill" : "pause.fill")
                .font(.system(size: Design.IconSize.sm))
                .frame(width: 40, height: 40)
                .background(Circle().fill(Design.Colors.controlBackground))
            }
            .buttonStyle(.plain)
            .accessibleButton(label: appState.pomodoroState.isPaused ? "Resume" : "Pause")

            Button(action: { appState.skipPomodoro() }) {
              Image(systemName: "forward.fill")
                .font(.system(size: Design.IconSize.sm))
                .frame(width: 40, height: 40)
                .background(Circle().fill(Design.Colors.controlBackground))
            }
            .buttonStyle(.plain)
            .accessibleButton(label: "Skip to next session")

            Button(action: { appState.stopPomodoro() }) {
              Image(systemName: "stop.fill")
                .font(.system(size: Design.IconSize.sm))
                .foregroundColor(Design.Colors.error)
                .frame(width: 40, height: 40)
                .background(Circle().stroke(Design.Colors.error.opacity(0.3), lineWidth: 1))
            }
            .buttonStyle(.plain)
            .accessibleButton(label: "Stop focus session")
          }
        } else {
          Button(action: { appState.startPomodoro() }) {
            HStack(spacing: Design.Spacing.xs) {
              Image(systemName: "play.fill")
              Text("Start Focus")
            }
            .font(Design.Typography.labelMedium)
            .foregroundColor(.white)
            .frame(width: 140, height: 40)
            .background(
              RoundedRectangle(cornerRadius: Design.Radius.sm)
                .fill(Design.Colors.error)
            )
          }
          .buttonStyle(.plain)
          .accessibleButton(label: "Start focus session")
        }
      }
      .frame(height: 44)

      // Stats
      HStack(spacing: Design.Spacing.lg) {
        VStack(spacing: Design.Spacing.xxxs) {
          Text("\(appState.pomodoroState.sessionsCompleted)")
            .font(Design.Typography.headingSmall)
            .frame(minWidth: 40)
          Text("Sessions")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)
        }
        .frame(minWidth: 70)

        Divider().frame(height: 30)

        VStack(spacing: Design.Spacing.xxxs) {
          Text("\(appState.pomodoroState.sessionsCompleted * appState.pomodoroState.workMinutes)m")
            .font(Design.Typography.headingSmall)
            .frame(minWidth: 50)
          Text("Focus Time")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)
        }
        .frame(minWidth: 80)
      }
      .accessibilityElement(children: .combine)
      .accessibilityLabel("\(appState.pomodoroState.sessionsCompleted) sessions, \(appState.pomodoroState.sessionsCompleted * appState.pomodoroState.workMinutes) minutes focus time")
    }
  }

  // MARK: - Footer Bar
  private var footerBar: some View {
    HStack {
      Button(action: { Task { await appState.refreshDashboard() } }) {
        Image(systemName: "arrow.clockwise")
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(Design.Colors.textSecondary)
      }
      .buttonStyle(.plain)
      .disabled(appState.isRefreshing)
      .accessibleButton(label: "Refresh", hint: "Refreshes dashboard data")

      if appState.isRefreshing {
        ProgressView().scaleEffect(0.6)
      }

      Spacer()

      // Open main window button
      Button(action: {
        openMainWindow()
        dismiss()
      }) {
        HStack(spacing: Design.Spacing.xxs) {
          Image(systemName: "macwindow")
            .font(.system(size: Design.IconSize.xs))
          Text("Open App")
            .font(Design.Typography.labelSmall)
        }
        .foregroundColor(Design.Colors.brand)
      }
      .buttonStyle(.plain)
      .accessibleButton(label: "Open main window")

      Spacer()

      // Secure mode indicator
      if appState.secureMode {
        Image(systemName: "eye.slash.fill")
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(Design.Colors.warning)
          .accessibilityLabel("Secure mode enabled")
      }

      Button(action: { NSApplication.shared.terminate(nil) }) {
        Text("Quit")
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)
      }
      .buttonStyle(.plain)
      .accessibleButton(label: "Quit application")
    }
    .padding(.horizontal, Design.Spacing.sm)
    .padding(.vertical, Design.Spacing.xs)
    .background(Design.Colors.divider)
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
      VStack(spacing: Design.Spacing.xxs) {
        Image(systemName: icon)
          .font(.system(size: Design.IconSize.sm))
          .foregroundColor(disabled ? Design.Colors.textSecondary : color)

        Text(title)
          .font(Design.Typography.labelSmall)
          .foregroundColor(disabled ? Design.Colors.textSecondary : Design.Colors.textPrimary)
      }
      .frame(maxWidth: .infinity)
      .padding(.vertical, Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(disabled ? Design.Colors.border : color.opacity(0.1))
      )
    }
    .buttonStyle(.plain)
    .disabled(disabled)
    .accessibleButton(label: title, hint: disabled ? "Currently unavailable" : nil)
  }
}

// MARK: - Compact Metric Card
struct CompactMetricCard: View {
  let title: String
  let value: String
  let icon: String
  let color: Color

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
      HStack(spacing: Design.Spacing.xxs) {
        Image(systemName: icon)
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(color)
        Text(title)
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textSecondary)
      }

      Text(value)
        .font(Design.Typography.labelLarge)
        .foregroundColor(Design.Colors.textPrimary)
    }
    .frame(maxWidth: .infinity, alignment: .leading)
    .padding(Design.Spacing.xs)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.sm)
        .fill(Design.Colors.controlBackground)
    )
    .accessibilityElement(children: .combine)
    .accessibilityLabel("\(title): \(value)")
  }
}

#Preview {
  MenuBarView(openMainWindow: {})
    .environmentObject(AppState())
}
