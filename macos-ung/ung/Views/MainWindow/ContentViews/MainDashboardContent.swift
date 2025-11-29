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
            .font(Design.Typography.bodyMedium)
            .foregroundColor(Design.Colors.textSecondary)
            .frame(maxWidth: .infinity)
            .padding(.vertical, Design.Spacing.lg)
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

          VStack(spacing: Design.Spacing.xxs) {
            Text(appState.formatHours(appState.metrics.weeklyHours))
              .font(Design.Typography.headingMedium)
              .foregroundColor(Design.Colors.textPrimary)

            Text("of \(Int(appState.metrics.weeklyTarget))h")
              .font(Design.Typography.bodySmall)
              .foregroundColor(Design.Colors.textSecondary)
          }
        }
        .frame(width: 120, height: 120)

        // Streak badge
        if appState.metrics.trackingStreak > 0 {
          HStack(spacing: Design.Spacing.xs) {
            Image(systemName: "flame.fill")
              .foregroundColor(Design.Colors.warning)
            Text("\(appState.metrics.trackingStreak) day streak")
              .font(Design.Typography.labelMedium)
              .foregroundColor(Design.Colors.warning)
          }
          .padding(.horizontal, Design.Spacing.sm)
          .padding(.vertical, Design.Spacing.xs)
          .background(Design.Colors.warningLight)
          .cornerRadius(Design.Radius.sm)
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
    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
      HStack {
        Image(systemName: "exclamationmark.triangle.fill")
          .foregroundColor(Design.Colors.warning)
        Text("Unpaid Invoices")
          .font(Design.Typography.headingSmall)
        Spacer()
      }

      HStack(spacing: Design.Spacing.md) {
        if appState.metrics.pendingAmount > 0 {
          VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
            Text("Pending")
              .font(Design.Typography.labelSmall)
              .foregroundColor(Design.Colors.textSecondary)
            Text(appState.formatCurrency(appState.metrics.pendingAmount))
              .font(Design.Typography.headingSmall)
              .foregroundColor(Design.Colors.warning)
          }
        }

        if appState.metrics.overdueAmount > 0 {
          VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
            Text("Overdue")
              .font(Design.Typography.labelSmall)
              .foregroundColor(Design.Colors.textSecondary)
            Text(appState.formatCurrency(appState.metrics.overdueAmount))
              .font(Design.Typography.headingSmall)
              .foregroundColor(Design.Colors.error)
          }
        }

        Spacer()

        Button("View Invoices") {
          appState.selectedTab = .invoices
        }
        .buttonStyle(.plain)
        .font(Design.Typography.labelMedium)
        .foregroundColor(Design.Colors.primary)
      }
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.warningLight)
        .overlay(
          RoundedRectangle(cornerRadius: Design.Radius.md)
            .stroke(Design.Colors.warning.opacity(0.3), lineWidth: 1)
        )
    )
  }

  private func statusColor(_ status: String) -> Color {
    switch status.lowercased() {
    case "paid": return Design.Colors.success
    case "sent", "pending": return Design.Colors.warning
    case "overdue": return Design.Colors.error
    default: return Design.Colors.textTertiary
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
    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
      HStack {
        Image(systemName: icon)
          .font(.system(size: Design.IconSize.sm))
          .foregroundColor(color)

        if let trend = trend {
          Spacer()
          Text(trend)
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.success)
            .padding(.horizontal, Design.Spacing.xs)
            .padding(.vertical, Design.Spacing.xxxs)
            .background(Design.Colors.successLight)
            .cornerRadius(Design.Radius.xs)
        }
      }

      Text(value)
        .font(Design.Typography.headingLarge)
        .foregroundColor(Design.Colors.textPrimary)

      Text(title)
        .font(Design.Typography.bodySmall)
        .foregroundColor(Design.Colors.textSecondary)
    }
    .frame(maxWidth: .infinity, alignment: .leading)
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: Design.Shadow.sm.color,
          radius: Design.Shadow.sm.radius,
          y: Design.Shadow.sm.y
        )
    )
  }
}

struct DashboardCard<Content: View>: View {
  let title: String
  let icon: String
  @ViewBuilder let content: () -> Content
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.md) {
      HStack {
        Image(systemName: icon)
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(Design.Colors.textSecondary)
        Text(title)
          .font(Design.Typography.headingSmall)
        Spacer()
      }
      .accessibleHeader(label: title)

      content()
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: Design.Shadow.sm.color,
          radius: Design.Shadow.sm.radius,
          y: Design.Shadow.sm.y
        )
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
      HStack(spacing: Design.Spacing.sm) {
        ZStack {
          RoundedRectangle(cornerRadius: Design.Radius.sm)
            .fill(color.opacity(0.15))
            .frame(width: 40, height: 40)

          Image(systemName: icon)
            .font(.system(size: Design.IconSize.md))
            .foregroundColor(color)
        }

        VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
          Text(title)
            .font(Design.Typography.bodyMedium)
            .fontWeight(.medium)
            .foregroundColor(Design.Colors.textPrimary)

          Text(subtitle)
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textSecondary)
        }

        Spacer()

        Image(systemName: "chevron.right")
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(Design.Colors.textTertiary)
      }
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(Design.Colors.backgroundSecondary(colorScheme))
      )
    }
    .buttonStyle(.plain)
    .accessibleButton(label: title, hint: subtitle)
  }
}

struct ActivityRow: View {
  let icon: String
  let iconColor: Color
  let title: String
  let subtitle: String
  let trailing: String

  var body: some View {
    HStack(spacing: Design.Spacing.sm) {
      Image(systemName: icon)
        .font(.system(size: Design.IconSize.xs))
        .foregroundColor(iconColor)
        .frame(width: Design.Spacing.lg)

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(title)
          .font(Design.Typography.labelMedium)
          .foregroundColor(Design.Colors.textPrimary)
          .lineLimit(1)

        Text(subtitle)
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textSecondary)
      }

      Spacer()

      Text(trailing)
        .font(Design.Typography.monoSmall)
        .foregroundColor(Design.Colors.textSecondary)
    }
    .padding(.vertical, Design.Spacing.xs)
  }
}

struct SetupStepRow: View {
  let title: String
  let isComplete: Bool
  let action: () -> Void

  var body: some View {
    Button(action: action) {
      HStack(spacing: Design.Spacing.sm) {
        Image(systemName: isComplete ? "checkmark.circle.fill" : "circle")
          .font(.system(size: Design.IconSize.md))
          .foregroundColor(isComplete ? Design.Colors.success : Design.Colors.textSecondary)

        Text(title)
          .font(Design.Typography.bodyMedium)
          .foregroundColor(isComplete ? Design.Colors.textSecondary : Design.Colors.textPrimary)
          .strikethrough(isComplete)

        Spacer()

        if !isComplete {
          Image(systemName: "chevron.right")
            .font(.system(size: Design.IconSize.xs))
            .foregroundColor(Design.Colors.textSecondary)
        }
      }
    }
    .buttonStyle(.plain)
    .disabled(isComplete)
    .accessibleButton(label: title, hint: isComplete ? "Completed" : "Tap to complete")
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
