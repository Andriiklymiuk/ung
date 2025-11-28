//
//  MainDashboardContent.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct MainDashboardContent: View {
  @EnvironmentObject var appState: AppState

  // Threshold for switching to vertical layout
  private let compactWidthThreshold: CGFloat = 700

  var body: some View {
    // Show full-screen onboarding when setup is incomplete
    if !appState.setupStatus.isComplete {
      OnboardingDashboardView()
    } else {
      regularDashboard
    }
  }

  private var regularDashboard: some View {
    GeometryReader { geometry in
      let isCompact = geometry.size.width < compactWidthThreshold

      ScrollView(showsIndicators: false) {
        VStack(spacing: 24) {
          // Key metrics row - responsive grid
          LazyVGrid(
            columns: [GridItem(.adaptive(minimum: isCompact ? 140 : 180))],
            spacing: 16
          ) {
            MetricCardLarge(
              title: "Total Revenue",
              value: appState.formatCurrency(appState.metrics.totalRevenue),
              icon: "chart.line.uptrend.xyaxis",
              color: .green,
              trend: nil
            )

            MetricCardLarge(
              title: "This Month",
              value: appState.formatCurrency(appState.metrics.monthlyRevenue),
              icon: "calendar",
              color: .blue,
              trend: nil
            )

            MetricCardLarge(
              title: "Pending",
              value: appState.formatCurrency(appState.metrics.pendingAmount),
              icon: "clock.fill",
              color: .orange,
              trend: nil
            )

            MetricCardLarge(
              title: "Hours This Week",
              value: appState.formatHours(appState.metrics.weeklyHours),
              icon: "clock.fill",
              color: .purple,
              trend: nil
            )
          }
          .drawingGroup()  // Performance optimization for cards

          // Two column layout - responsive
          if isCompact {
            // Vertical layout for narrow windows
            VStack(spacing: 20) {
              quickActionsCard
              weeklyProgressCard

              if appState.metrics.pendingAmount > 0 || appState.metrics.overdueAmount > 0 {
                unpaidInvoicesCard
              }

              recentActivityCard
            }
          } else {
            // Horizontal layout for wide windows
            HStack(alignment: .top, spacing: 20) {
              // Left column
              VStack(spacing: 20) {
                quickActionsCard
                recentActivityCard
              }
              .frame(maxWidth: .infinity)

              // Right column
              VStack(spacing: 20) {
                weeklyProgressCard

                if appState.metrics.pendingAmount > 0 || appState.metrics.overdueAmount > 0 {
                  unpaidInvoicesCard
                }
              }
              .frame(maxWidth: .infinity)
            }
          }

          Spacer(minLength: 24)
        }
        .padding(24)
      }
    }
  }

  // MARK: - Metrics Row
  private var metricsRow: some View {
    HStack(spacing: 16) {
      MetricCardLarge(
        title: "Total Revenue",
        value: appState.formatCurrency(appState.metrics.totalRevenue),
        icon: "chart.line.uptrend.xyaxis",
        color: .green,
        trend: nil
      )

      MetricCardLarge(
        title: "This Month",
        value: appState.formatCurrency(appState.metrics.monthlyRevenue),
        icon: "calendar",
        color: .blue,
        trend: nil
      )

      MetricCardLarge(
        title: "Pending",
        value: appState.formatCurrency(appState.metrics.pendingAmount),
        icon: "clock.fill",
        color: .orange,
        trend: nil
      )

      MetricCardLarge(
        title: "Hours This Week",
        value: appState.formatHours(appState.metrics.weeklyHours),
        icon: "clock.fill",
        color: .purple,
        trend: nil
      )
    }
  }

  // MARK: - Quick Actions Card
  private var quickActionsCard: some View {
    DashboardCard(title: "Quick Actions", icon: "bolt.fill") {
      VStack(spacing: 8) {
        QuickActionCardButton(
          icon: "play.fill",
          title: "Start Tracking",
          subtitle: "Track your time",
          color: .green
        ) {
          appState.selectedTab = .tracking
        }

        QuickActionCardButton(
          icon: "doc.text.fill",
          title: "Create Invoice",
          subtitle: "Bill your clients",
          color: .teal
        ) {
          appState.selectedTab = .invoices
        }

        QuickActionCardButton(
          icon: "dollarsign.circle.fill",
          title: "Log Expense",
          subtitle: "Track spending",
          color: .orange
        ) {
          appState.selectedTab = .expenses
        }

        QuickActionCardButton(
          icon: "brain.head.profile",
          title: "Focus Mode",
          subtitle: "Start Pomodoro",
          color: .red
        ) {
          appState.selectedTab = .pomodoro
        }
      }
    }
  }

  // MARK: - Recent Activity Card
  private var recentActivityCard: some View {
    DashboardCard(title: "Recent Activity", icon: "clock.arrow.circlepath") {
      VStack(spacing: 8) {
        if appState.recentSessions.isEmpty && appState.recentInvoices.isEmpty {
          Text("No recent activity")
            .font(.system(size: 13))
            .foregroundColor(.secondary)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 20)
        } else {
          ForEach(appState.recentSessions.prefix(3)) { session in
            ActivityRow(
              icon: "clock.fill",
              iconColor: .blue,
              title: session.project,
              subtitle: session.date,
              trailing: appState.secureMode ? "**:**" : session.duration
            )
          }

          ForEach(appState.recentInvoices.prefix(2)) { invoice in
            ActivityRow(
              icon: "doc.text.fill",
              iconColor: statusColor(invoice.status),
              title: invoice.invoiceNum,
              subtitle: invoice.client,
              trailing: appState.secureMode ? "****" : invoice.amount
            )
          }
        }
      }
    }
  }

  // MARK: - Weekly Progress Card
  private var weeklyProgressCard: some View {
    DashboardCard(title: "Weekly Progress", icon: "chart.bar.fill") {
      VStack(spacing: 16) {
        // Circular progress
        ZStack {
          Circle()
            .stroke(Color.blue.opacity(0.2), lineWidth: 12)

          Circle()
            .trim(
              from: 0, to: min(appState.metrics.weeklyHours / appState.metrics.weeklyTarget, 1.0)
            )
            .stroke(
              LinearGradient(
                colors: [.blue, .purple], startPoint: .topLeading, endPoint: .bottomTrailing),
              style: StrokeStyle(lineWidth: 12, lineCap: .round)
            )
            .rotationEffect(.degrees(-90))

          VStack(spacing: 4) {
            Text(appState.formatHours(appState.metrics.weeklyHours))
              .font(.system(size: 20, weight: .bold, design: .rounded))
              .foregroundColor(.primary)

            Text("of \(Int(appState.metrics.weeklyTarget))h")
              .font(.system(size: 12))
              .foregroundColor(.secondary)
          }
        }
        .frame(width: 120, height: 120)

        // Streak badge
        if appState.metrics.trackingStreak > 0 {
          HStack(spacing: 6) {
            Image(systemName: "flame.fill")
              .foregroundColor(.orange)
            Text("\(appState.metrics.trackingStreak) day streak")
              .font(.system(size: 12, weight: .medium))
              .foregroundColor(.orange)
          }
          .padding(.horizontal, 12)
          .padding(.vertical, 6)
          .background(Color.orange.opacity(0.15))
          .cornerRadius(8)
        }
      }
      .frame(maxWidth: .infinity)
    }
  }

  // MARK: - Setup Checklist Card
  private var setupChecklistCard: some View {
    DashboardCard(title: "Setup Checklist", icon: "checklist") {
      VStack(spacing: 12) {
        SetupStepRow(
          title: "Company Profile",
          isComplete: appState.setupStatus.hasCompany,
          action: { appState.selectedTab = .settings }
        )

        SetupStepRow(
          title: "First Client",
          isComplete: appState.setupStatus.hasClient,
          action: { appState.selectedTab = .clients }
        )

        SetupStepRow(
          title: "First Contract",
          isComplete: appState.setupStatus.hasContract,
          action: { appState.selectedTab = .contracts }
        )
      }
    }
  }

  // MARK: - Unpaid Invoices Card
  private var unpaidInvoicesCard: some View {
    VStack(alignment: .leading, spacing: 12) {
      HStack {
        Image(systemName: "exclamationmark.triangle.fill")
          .foregroundColor(.orange)
        Text("Unpaid Invoices")
          .font(.system(size: 14, weight: .semibold))
        Spacer()
      }

      HStack(spacing: 16) {
        if appState.metrics.pendingAmount > 0 {
          VStack(alignment: .leading, spacing: 2) {
            Text("Pending")
              .font(.system(size: 11))
              .foregroundColor(.secondary)
            Text(appState.formatCurrency(appState.metrics.pendingAmount))
              .font(.system(size: 16, weight: .bold))
              .foregroundColor(.orange)
          }
        }

        if appState.metrics.overdueAmount > 0 {
          VStack(alignment: .leading, spacing: 2) {
            Text("Overdue")
              .font(.system(size: 11))
              .foregroundColor(.secondary)
            Text(appState.formatCurrency(appState.metrics.overdueAmount))
              .font(.system(size: 16, weight: .bold))
              .foregroundColor(.red)
          }
        }

        Spacer()

        Button("View Invoices") {
          appState.selectedTab = .invoices
        }
        .buttonStyle(.plain)
        .font(.system(size: 12, weight: .medium))
        .foregroundColor(.blue)
      }
    }
    .padding(16)
    .background(
      RoundedRectangle(cornerRadius: 12)
        .fill(Color.orange.opacity(0.1))
        .overlay(
          RoundedRectangle(cornerRadius: 12)
            .stroke(Color.orange.opacity(0.3), lineWidth: 1)
        )
    )
  }

  private func statusColor(_ status: String) -> Color {
    switch status.lowercased() {
    case "paid": return .green
    case "sent", "pending": return .orange
    case "overdue": return .red
    default: return .gray
    }
  }
}

// MARK: - Supporting Views
struct MetricCardLarge: View {
  let title: String
  let value: String
  let icon: String
  let color: Color
  let trend: String?
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: 12) {
      HStack {
        Image(systemName: icon)
          .font(.system(size: 14))
          .foregroundColor(color)

        if let trend = trend {
          Spacer()
          Text(trend)
            .font(.system(size: 11, weight: .medium))
            .foregroundColor(.green)
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(Color.green.opacity(0.15))
            .cornerRadius(4)
        }
      }

      Text(value)
        .font(.system(size: 24, weight: .bold, design: .rounded))
        .foregroundColor(.primary)

      Text(title)
        .font(.system(size: 12))
        .foregroundColor(.secondary)
    }
    .frame(maxWidth: .infinity, alignment: .leading)
    .padding(16)
    .background(
      RoundedRectangle(cornerRadius: 12)
        .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
        .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
    )
  }
}

struct DashboardCard<Content: View>: View {
  let title: String
  let icon: String
  @ViewBuilder let content: () -> Content
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: 16) {
      HStack {
        Image(systemName: icon)
          .font(.system(size: 12))
          .foregroundColor(.secondary)
        Text(title)
          .font(.system(size: 14, weight: .semibold))
        Spacer()
      }

      content()
    }
    .padding(16)
    .background(
      RoundedRectangle(cornerRadius: 12)
        .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
        .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
    )
  }
}

struct QuickActionCardButton: View {
  let icon: String
  let title: String
  let subtitle: String
  let color: Color
  let action: () -> Void
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    Button(action: action) {
      HStack(spacing: 12) {
        ZStack {
          RoundedRectangle(cornerRadius: 10)
            .fill(color.opacity(0.15))
            .frame(width: 40, height: 40)

          Image(systemName: icon)
            .font(.system(size: 16))
            .foregroundColor(color)
        }

        VStack(alignment: .leading, spacing: 2) {
          Text(title)
            .font(.system(size: 13, weight: .medium))
            .foregroundColor(.primary)

          Text(subtitle)
            .font(.system(size: 11))
            .foregroundColor(.secondary)
        }

        Spacer()

        Image(systemName: "chevron.right")
          .font(.system(size: 11))
          .foregroundColor(.secondary)
      }
      .padding(12)
      .background(
        RoundedRectangle(cornerRadius: 10)
          .fill(colorScheme == .dark ? Color(white: 0.1) : Color(white: 0.97))
      )
    }
    .buttonStyle(.plain)
  }
}

struct ActivityRow: View {
  let icon: String
  let iconColor: Color
  let title: String
  let subtitle: String
  let trailing: String

  var body: some View {
    HStack(spacing: 10) {
      Image(systemName: icon)
        .font(.system(size: 12))
        .foregroundColor(iconColor)
        .frame(width: 24)

      VStack(alignment: .leading, spacing: 1) {
        Text(title)
          .font(.system(size: 12, weight: .medium))
          .foregroundColor(.primary)
          .lineLimit(1)

        Text(subtitle)
          .font(.system(size: 10))
          .foregroundColor(.secondary)
      }

      Spacer()

      Text(trailing)
        .font(.system(size: 11, weight: .medium, design: .monospaced))
        .foregroundColor(.secondary)
    }
    .padding(.vertical, 6)
  }
}

struct SetupStepRow: View {
  let title: String
  let isComplete: Bool
  let action: () -> Void

  var body: some View {
    Button(action: action) {
      HStack(spacing: 10) {
        Image(systemName: isComplete ? "checkmark.circle.fill" : "circle")
          .font(.system(size: 16))
          .foregroundColor(isComplete ? .green : .secondary)

        Text(title)
          .font(.system(size: 13))
          .foregroundColor(isComplete ? .secondary : .primary)
          .strikethrough(isComplete)

        Spacer()

        if !isComplete {
          Image(systemName: "chevron.right")
            .font(.system(size: 10))
            .foregroundColor(.secondary)
        }
      }
    }
    .buttonStyle(.plain)
    .disabled(isComplete)
  }
}

// MARK: - Onboarding Dashboard View
struct OnboardingDashboardView: View {
  @EnvironmentObject var appState: AppState
  @Environment(\.colorScheme) var colorScheme

  private var completedSteps: Int {
    var count = 0
    if appState.setupStatus.hasCompany { count += 1 }
    if appState.setupStatus.hasClient { count += 1 }
    if appState.setupStatus.hasContract { count += 1 }
    return count
  }

  var body: some View {
    VStack(spacing: 24) {
      // Header - more compact
      VStack(spacing: 8) {
        Image(systemName: "sparkles")
          .font(.system(size: 40))
          .foregroundStyle(
            LinearGradient(
              colors: [.blue, .purple],
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            )
          )

        Text("Welcome to UNG")
          .font(.system(size: 24, weight: .bold, design: .rounded))
          .foregroundColor(.primary)

        Text("Let's get you set up in 3 easy steps")
          .font(.system(size: 14))
          .foregroundColor(.secondary)

        // Progress indicator
        HStack(spacing: 8) {
          Text("\(completedSteps) of 3 complete")
            .font(.system(size: 12, weight: .medium))
            .foregroundColor(.secondary)

          // Progress bar
          ZStack(alignment: .leading) {
            RoundedRectangle(cornerRadius: 4)
              .fill(Color.secondary.opacity(0.2))
              .frame(width: 100, height: 6)

            RoundedRectangle(cornerRadius: 4)
              .fill(
                LinearGradient(
                  colors: [.blue, .purple],
                  startPoint: .leading,
                  endPoint: .trailing
                )
              )
              .frame(width: 100 * CGFloat(completedSteps) / 3.0, height: 6)
          }
        }
        .padding(.top, 4)
      }
      .padding(.top, 16)

      // Setup Steps - more compact
      VStack(spacing: 12) {
        OnboardingStepCard(
          stepNumber: 1,
          title: "Set Up Your Company",
          description: "Add your business details",
          icon: "building.2.fill",
          isComplete: appState.setupStatus.hasCompany,
          isEnabled: true,
          action: { appState.selectedTab = .settings }
        )

        OnboardingStepCard(
          stepNumber: 2,
          title: "Add Your First Client",
          description: "Who will you be working with?",
          icon: "person.2.fill",
          isComplete: appState.setupStatus.hasClient,
          isEnabled: appState.setupStatus.hasCompany,
          action: { appState.selectedTab = .clients }
        )

        OnboardingStepCard(
          stepNumber: 3,
          title: "Create a Contract",
          description: "Define your working terms",
          icon: "doc.text.fill",
          isComplete: appState.setupStatus.hasContract,
          isEnabled: appState.setupStatus.hasClient,
          action: { appState.selectedTab = .contracts }
        )
      }
      .frame(maxWidth: 480)

      // Quick Actions - more compact
      VStack(alignment: .leading, spacing: 12) {
        Text("Quick Actions")
          .font(.system(size: 13, weight: .semibold))
          .foregroundColor(.secondary)

        HStack(spacing: 10) {
          OnboardingQuickAction(
            icon: "play.fill",
            title: "Start Tracking",
            color: .green,
            action: { appState.selectedTab = .tracking }
          )

          OnboardingQuickAction(
            icon: "brain.head.profile",
            title: "Focus Mode",
            color: .orange,
            action: { appState.selectedTab = .pomodoro }
          )

          OnboardingQuickAction(
            icon: "chart.bar.fill",
            title: "View Reports",
            color: .blue,
            action: { appState.selectedTab = .reports }
          )
        }
      }
      .frame(maxWidth: 480)

      Spacer()
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity)
    .padding(20)
  }
}

// MARK: - Onboarding Step Card
struct OnboardingStepCard: View {
  let stepNumber: Int
  let title: String
  let description: String
  let icon: String
  let isComplete: Bool
  let isEnabled: Bool
  let action: () -> Void
  @Environment(\.colorScheme) var colorScheme
  @State private var isHovered = false

  var body: some View {
    Button(action: {
      if isEnabled && !isComplete {
        action()
      }
    }) {
      HStack(spacing: 16) {
        // Step indicator
        ZStack {
          Circle()
            .fill(
              isComplete
                ? Color.green
                : (isEnabled ? Color.accentColor : Color.secondary.opacity(0.3))
            )
            .frame(width: 44, height: 44)

          if isComplete {
            Image(systemName: "checkmark")
              .font(.system(size: 18, weight: .bold))
              .foregroundColor(.white)
          } else {
            Text("\(stepNumber)")
              .font(.system(size: 18, weight: .bold, design: .rounded))
              .foregroundColor(isEnabled ? .white : .secondary)
          }
        }

        // Content
        VStack(alignment: .leading, spacing: 4) {
          Text(title)
            .font(.system(size: 15, weight: .semibold))
            .foregroundColor(isComplete ? .secondary : .primary)
            .strikethrough(isComplete, color: .secondary)

          Text(description)
            .font(.system(size: 13))
            .foregroundColor(.secondary)
            .lineLimit(2)
        }

        Spacer()

        // Status/Action indicator
        if isComplete {
          Image(systemName: "checkmark.circle.fill")
            .font(.system(size: 20))
            .foregroundColor(.green)
        } else if isEnabled {
          HStack(spacing: 6) {
            Text("Set Up")
              .font(.system(size: 13, weight: .medium))
            Image(systemName: "arrow.right")
              .font(.system(size: 11))
          }
          .foregroundColor(.accentColor)
        } else {
          HStack(spacing: 4) {
            Image(systemName: "lock.fill")
              .font(.system(size: 10))
            Text("Complete step \(stepNumber - 1)")
              .font(.system(size: 11))
          }
          .foregroundColor(.secondary)
        }
      }
      .padding(16)
      .background(
        RoundedRectangle(cornerRadius: 12)
          .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
          .shadow(
            color: .black.opacity(isHovered && isEnabled && !isComplete ? 0.08 : 0.04),
            radius: isHovered ? 8 : 4, y: isHovered ? 4 : 2)
      )
      .overlay(
        RoundedRectangle(cornerRadius: 12)
          .stroke(
            isEnabled && !isComplete ? Color.accentColor.opacity(isHovered ? 0.5 : 0) : Color.clear,
            lineWidth: 2
          )
      )
      .scaleEffect(isHovered && isEnabled && !isComplete ? 1.01 : 1.0)
      .animation(.easeInOut(duration: 0.15), value: isHovered)
    }
    .buttonStyle(.plain)
    .disabled(isComplete || !isEnabled)
    .onHover { hovering in
      isHovered = hovering
    }
  }
}

// MARK: - Onboarding Quick Action
struct OnboardingQuickAction: View {
  let icon: String
  let title: String
  let color: Color
  let action: () -> Void
  @Environment(\.colorScheme) var colorScheme
  @State private var isHovered = false

  var body: some View {
    Button(action: action) {
      VStack(spacing: 10) {
        ZStack {
          RoundedRectangle(cornerRadius: 12)
            .fill(color.opacity(0.15))
            .frame(width: 48, height: 48)

          Image(systemName: icon)
            .font(.system(size: 20))
            .foregroundColor(color)
        }

        Text(title)
          .font(.system(size: 12, weight: .medium))
          .foregroundColor(.primary)
          .lineLimit(1)
      }
      .frame(maxWidth: .infinity)
      .padding(.vertical, 16)
      .background(
        RoundedRectangle(cornerRadius: 12)
          .fill(colorScheme == .dark ? Color(white: 0.12) : Color.white)
          .shadow(
            color: .black.opacity(isHovered ? 0.08 : 0.04), radius: isHovered ? 8 : 4,
            y: isHovered ? 4 : 2)
      )
      .scaleEffect(isHovered ? 1.02 : 1.0)
      .animation(.easeInOut(duration: 0.15), value: isHovered)
    }
    .buttonStyle(.plain)
    .onHover { hovering in
      isHovered = hovering
    }
  }
}

#Preview {
  MainDashboardContent()
    .environmentObject(AppState())
    .frame(width: 800, height: 600)
}

#Preview("Onboarding") {
  OnboardingDashboardView()
    .environmentObject(AppState())
    .frame(width: 800, height: 700)
}
