//
//  DailySummaryLiveActivity.swift
//  ungWidgets
//
//  Premium Live Activity for daily/weekly summary and celebrations
//  "Celebrate your wins" - Marketing hook
//

#if os(iOS)
import ActivityKit
import SwiftUI
import WidgetKit

// MARK: - Daily Summary Activity Attributes
@available(iOS 16.1, *)
struct DailySummaryActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var hoursWorked: Double
        var earnings: Double
        var sessionsCompleted: Int
        var streakDays: Int
        var isNewRecord: Bool // Personal best!
    }

    var summaryType: String // "daily", "weekly"
    var targetHours: Double
    var currency: String
    var userName: String?
}

// MARK: - Break Reminder Activity Attributes
@available(iOS 16.1, *)
struct BreakReminderActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var minutesSinceLastBreak: Int
        var totalWorkMinutesToday: Int
        var breaksTakenToday: Int
        var urgencyLevel: Int // 0: normal, 1: suggested, 2: recommended, 3: urgent
    }

    var breakIntervalMinutes: Int
    var recommendedBreakMinutes: Int
}

// MARK: - Daily Summary Live Activity Widget
@available(iOS 16.1, *)
struct DailySummaryLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: DailySummaryActivityAttributes.self) { context in
            DailySummaryLockScreenView(context: context)
                .activityBackgroundTint(Color.purple.opacity(0.85))
                .activitySystemActionForegroundColor(Color.white)
        } dynamicIsland: { context in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    SummaryExpandedLeading(context: context)
                }

                DynamicIslandExpandedRegion(.trailing) {
                    SummaryExpandedTrailing(context: context)
                }

                DynamicIslandExpandedRegion(.center) {
                    SummaryExpandedCenter(context: context)
                }

                DynamicIslandExpandedRegion(.bottom) {
                    SummaryExpandedBottom(context: context)
                }
            } compactLeading: {
                SummaryCompactLeading(context: context)
            } compactTrailing: {
                SummaryCompactTrailing(context: context)
            } minimal: {
                SummaryMinimalView(context: context)
            }
            .widgetURL(URL(string: "ung://dashboard"))
            .keylineTint(Color.purple)
        }
    }
}

// MARK: - Break Reminder Live Activity Widget
@available(iOS 16.1, *)
struct BreakReminderLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: BreakReminderActivityAttributes.self) { context in
            BreakReminderLockScreenView(context: context)
                .activityBackgroundTint(urgencyBackground(context.state.urgencyLevel))
                .activitySystemActionForegroundColor(Color.white)
        } dynamicIsland: { context in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    BreakExpandedLeading(context: context)
                }

                DynamicIslandExpandedRegion(.trailing) {
                    BreakExpandedTrailing(context: context)
                }

                DynamicIslandExpandedRegion(.center) {
                    BreakExpandedCenter(context: context)
                }

                DynamicIslandExpandedRegion(.bottom) {
                    BreakExpandedBottom(context: context)
                }
            } compactLeading: {
                BreakCompactLeading(context: context)
            } compactTrailing: {
                BreakCompactTrailing(context: context)
            } minimal: {
                BreakMinimalView(context: context)
            }
            .widgetURL(URL(string: "ung://pomodoro"))
            .keylineTint(urgencyColor(context.state.urgencyLevel))
        }
    }

    private func urgencyBackground(_ level: Int) -> Color {
        switch level {
        case 3: return Color.red.opacity(0.85)
        case 2: return Color.orange.opacity(0.85)
        case 1: return Color.yellow.opacity(0.85)
        default: return Color.teal.opacity(0.85)
        }
    }

    private func urgencyColor(_ level: Int) -> Color {
        switch level {
        case 3: return .red
        case 2: return .orange
        case 1: return .yellow
        default: return .teal
        }
    }
}

// MARK: - Daily Summary Lock Screen View
@available(iOS 16.1, *)
private struct DailySummaryLockScreenView: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    private var progress: Double {
        min(context.state.hoursWorked / max(context.attributes.targetHours, 1), 1.5)
    }

    var body: some View {
        HStack(spacing: 16) {
            // Achievement badge
            SummaryAchievementBadge(
                progress: progress,
                isNewRecord: context.state.isNewRecord,
                streakDays: context.state.streakDays
            )

            // Summary info
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 8) {
                    Text(context.state.isNewRecord ? "New Record!" : "\(context.attributes.summaryType.capitalized) Summary")
                        .font(.system(size: 17, weight: .semibold))
                        .foregroundColor(.white)

                    if context.state.streakDays > 0 {
                        HStack(spacing: 2) {
                            Image(systemName: "flame.fill")
                                .font(.system(size: 10))
                            Text("\(context.state.streakDays)")
                                .font(.system(size: 10, weight: .bold))
                        }
                        .foregroundColor(.orange)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color.orange.opacity(0.2))
                        .cornerRadius(4)
                    }
                }

                HStack(spacing: 12) {
                    // Sessions
                    HStack(spacing: 4) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 10))
                            .foregroundColor(.green)
                        Text("\(context.state.sessionsCompleted) sessions")
                            .font(.system(size: 12))
                            .foregroundColor(.gray)
                    }

                    // Goal progress
                    let percentage = Int(progress * 100)
                    Text("\(percentage)% of goal")
                        .font(.system(size: 12))
                        .foregroundColor(progress >= 1.0 ? .green : .gray)
                }
            }

            Spacer()

            // Stats column
            VStack(alignment: .trailing, spacing: 8) {
                // Hours
                VStack(alignment: .trailing, spacing: 2) {
                    Text(formatHours(context.state.hoursWorked))
                        .font(.system(size: 24, weight: .bold, design: .rounded))
                        .foregroundStyle(
                            LinearGradient(colors: [.purple, .pink], startPoint: .leading, endPoint: .trailing)
                        )

                    Text("hours worked")
                        .font(.system(size: 10))
                        .foregroundColor(.gray)
                }

                // Earnings
                if context.state.earnings > 0 {
                    Text("+\(context.attributes.currency)\(context.state.earnings, specifier: "%.0f")")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.green)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
    }

    private func formatHours(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        return m > 0 ? "\(h)h \(m)m" : "\(h)h"
    }
}

// MARK: - Summary Achievement Badge
@available(iOS 16.1, *)
private struct SummaryAchievementBadge: View {
    let progress: Double
    let isNewRecord: Bool
    let streakDays: Int

    var body: some View {
        ZStack {
            // Celebration glow for achievements
            if isNewRecord || progress >= 1.0 {
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [Color.purple.opacity(0.4), Color.purple.opacity(0)],
                            center: .center,
                            startRadius: 20,
                            endRadius: 45
                        )
                    )
                    .frame(width: 90, height: 90)
            }

            // Progress ring
            Circle()
                .stroke(Color.purple.opacity(0.2), lineWidth: 8)
                .frame(width: 64, height: 64)

            Circle()
                .trim(from: 0, to: min(progress, 1.0))
                .stroke(
                    AngularGradient(
                        colors: [.purple, .pink, .purple],
                        center: .center,
                        startAngle: .degrees(-90),
                        endAngle: .degrees(270)
                    ),
                    style: StrokeStyle(lineWidth: 8, lineCap: .round)
                )
                .frame(width: 64, height: 64)
                .rotationEffect(.degrees(-90))

            // Center content
            if isNewRecord {
                Image(systemName: "trophy.fill")
                    .font(.system(size: 24))
                    .foregroundStyle(
                        LinearGradient(colors: [.yellow, .orange], startPoint: .top, endPoint: .bottom)
                    )
            } else if progress >= 1.0 {
                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 24))
                    .foregroundColor(.green)
            } else {
                Image(systemName: "chart.bar.fill")
                    .font(.system(size: 22))
                    .foregroundStyle(
                        LinearGradient(colors: [.purple, .pink], startPoint: .top, endPoint: .bottom)
                    )
            }
        }
    }
}

// MARK: - Dynamic Island Views for Summary
@available(iOS 16.1, *)
private struct SummaryExpandedLeading: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    var body: some View {
        HStack(spacing: 8) {
            ZStack {
                Circle()
                    .fill(Color.purple.opacity(0.2))
                    .frame(width: 28, height: 28)

                Image(systemName: context.state.isNewRecord ? "trophy.fill" : "chart.bar.fill")
                    .font(.system(size: 14))
                    .foregroundColor(context.state.isNewRecord ? .yellow : .purple)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(context.state.isNewRecord ? "New Record!" : "Summary")
                    .font(.system(size: 14, weight: .semibold))

                Text(context.attributes.summaryType.capitalized)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }
        }
    }
}

@available(iOS 16.1, *)
private struct SummaryExpandedTrailing: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            Text(formatHoursCompact(context.state.hoursWorked))
                .font(.system(size: 20, weight: .bold, design: .rounded))
                .foregroundStyle(
                    LinearGradient(colors: [.purple, .pink], startPoint: .leading, endPoint: .trailing)
                )

            if context.state.earnings > 0 {
                Text("+\(context.attributes.currency)\(context.state.earnings, specifier: "%.0f")")
                    .font(.system(size: 10))
                    .foregroundColor(.green)
            }
        }
    }

    private func formatHoursCompact(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        return m > 0 ? "\(h):\(String(format: "%02d", m))" : "\(h)h"
    }
}

@available(iOS 16.1, *)
private struct SummaryExpandedCenter: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    var body: some View {
        if context.state.isNewRecord {
            HStack(spacing: 4) {
                Image(systemName: "star.fill")
                    .font(.system(size: 10))
                Text("PERSONAL BEST")
                    .font(.system(size: 10, weight: .bold))
            }
            .foregroundColor(.yellow)
        } else if context.state.streakDays > 0 {
            HStack(spacing: 4) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 10))
                Text("\(context.state.streakDays) day streak")
                    .font(.system(size: 10))
            }
            .foregroundColor(.orange)
        }
    }
}

@available(iOS 16.1, *)
private struct SummaryExpandedBottom: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    private var progress: Double {
        min(context.state.hoursWorked / max(context.attributes.targetHours, 1), 1.0)
    }

    var body: some View {
        VStack(spacing: 8) {
            // Progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.gray.opacity(0.3))
                        .frame(height: 6)

                    RoundedRectangle(cornerRadius: 4)
                        .fill(
                            LinearGradient(colors: [.purple, .pink], startPoint: .leading, endPoint: .trailing)
                        )
                        .frame(width: geometry.size.width * progress, height: 6)
                }
            }
            .frame(height: 6)
            .padding(.horizontal, 4)

            // Stats row
            HStack {
                HStack(spacing: 4) {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.green)
                    Text("\(context.state.sessionsCompleted) sessions")
                        .font(.system(size: 10))
                }
                .foregroundColor(.secondary)

                Spacer()

                Text("\(Int(progress * 100))% of \(formatHoursShort(context.attributes.targetHours)) goal")
                    .font(.system(size: 10))
                    .foregroundColor(progress >= 1.0 ? .green : .secondary)
            }
            .padding(.horizontal, 4)
        }
    }

    private func formatHoursShort(_ hours: Double) -> String {
        return "\(Int(hours))h"
    }
}

// MARK: - Compact Views for Summary
@available(iOS 16.1, *)
private struct SummaryCompactLeading: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    var body: some View {
        Image(systemName: context.state.isNewRecord ? "trophy.fill" : "chart.bar.fill")
            .font(.system(size: 14))
            .foregroundColor(context.state.isNewRecord ? .yellow : .purple)
    }
}

@available(iOS 16.1, *)
private struct SummaryCompactTrailing: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    var body: some View {
        Text("\(Int(context.state.hoursWorked))h")
            .font(.system(size: 13, weight: .bold, design: .rounded))
            .foregroundStyle(
                LinearGradient(colors: [.purple, .pink], startPoint: .leading, endPoint: .trailing)
            )
    }
}

@available(iOS 16.1, *)
private struct SummaryMinimalView: View {
    let context: ActivityViewContext<DailySummaryActivityAttributes>

    var body: some View {
        ZStack {
            Circle()
                .fill(
                    RadialGradient(
                        colors: [Color.purple.opacity(0.3), Color.clear],
                        center: .center,
                        startRadius: 2,
                        endRadius: 14
                    )
                )

            if context.state.isNewRecord {
                Image(systemName: "trophy.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.yellow)
            } else {
                Image(systemName: "chart.bar.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.purple)
            }
        }
    }
}

// MARK: - Break Reminder Views
@available(iOS 16.1, *)
private struct BreakReminderLockScreenView: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>

    private var theme: BreakTheme {
        BreakTheme(urgencyLevel: context.state.urgencyLevel)
    }

    var body: some View {
        HStack(spacing: 16) {
            // Health indicator
            BreakHealthIndicator(
                minutesSinceBreak: context.state.minutesSinceLastBreak,
                interval: context.attributes.breakIntervalMinutes,
                urgencyLevel: context.state.urgencyLevel
            )

            // Break info
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 8) {
                    Text(theme.title)
                        .font(.system(size: 17, weight: .semibold))
                        .foregroundColor(.white)

                    if context.state.urgencyLevel >= 2 {
                        Text(theme.urgencyLabel)
                            .font(.system(size: 9, weight: .bold))
                            .foregroundColor(theme.primaryColor)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(theme.primaryColor.opacity(0.2))
                            .cornerRadius(4)
                    }
                }

                HStack(spacing: 12) {
                    HStack(spacing: 4) {
                        Image(systemName: "clock.fill")
                            .font(.system(size: 10))
                        Text("\(context.state.minutesSinceLastBreak)m since break")
                            .font(.system(size: 12))
                    }
                    .foregroundColor(.gray)

                    HStack(spacing: 4) {
                        Image(systemName: "cup.and.saucer.fill")
                            .font(.system(size: 10))
                        Text("\(context.state.breaksTakenToday) today")
                            .font(.system(size: 12))
                    }
                    .foregroundColor(.gray)
                }
            }

            Spacer()

            // Action suggestion
            VStack(alignment: .trailing, spacing: 4) {
                Text(theme.actionText)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(theme.gradient)

                Text("\(context.attributes.recommendedBreakMinutes)min break")
                    .font(.system(size: 11))
                    .foregroundColor(.gray)

                // Total work time
                let hours = context.state.totalWorkMinutesToday / 60
                let mins = context.state.totalWorkMinutesToday % 60
                Text("\(hours)h \(mins)m worked")
                    .font(.system(size: 10))
                    .foregroundColor(.gray)
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
    }
}

// MARK: - Break Theme
@available(iOS 16.1, *)
private struct BreakTheme {
    let urgencyLevel: Int

    var primaryColor: Color {
        switch urgencyLevel {
        case 3: return .red
        case 2: return .orange
        case 1: return .yellow
        default: return .teal
        }
    }

    var gradient: LinearGradient {
        switch urgencyLevel {
        case 3: return LinearGradient(colors: [.red, .orange], startPoint: .leading, endPoint: .trailing)
        case 2: return LinearGradient(colors: [.orange, .yellow], startPoint: .leading, endPoint: .trailing)
        case 1: return LinearGradient(colors: [.yellow, .green], startPoint: .leading, endPoint: .trailing)
        default: return LinearGradient(colors: [.teal, .cyan], startPoint: .leading, endPoint: .trailing)
        }
    }

    var title: String {
        switch urgencyLevel {
        case 3: return "Break Needed!"
        case 2: return "Time for a Break"
        case 1: return "Break Suggested"
        default: return "Break Reminder"
        }
    }

    var urgencyLabel: String {
        switch urgencyLevel {
        case 3: return "URGENT"
        case 2: return "RECOMMENDED"
        default: return ""
        }
    }

    var actionText: String {
        switch urgencyLevel {
        case 3: return "Take a break"
        case 2: return "Break time"
        case 1: return "Consider break"
        default: return "On track"
        }
    }

    var icon: String {
        switch urgencyLevel {
        case 3: return "exclamationmark.triangle.fill"
        case 2: return "clock.badge.exclamationmark.fill"
        case 1: return "cup.and.saucer.fill"
        default: return "heart.fill"
        }
    }
}

// MARK: - Break Health Indicator
@available(iOS 16.1, *)
private struct BreakHealthIndicator: View {
    let minutesSinceBreak: Int
    let interval: Int
    let urgencyLevel: Int

    private var theme: BreakTheme {
        BreakTheme(urgencyLevel: urgencyLevel)
    }

    private var progress: Double {
        min(Double(minutesSinceBreak) / Double(max(interval, 1)), 1.0)
    }

    var body: some View {
        ZStack {
            // Warning glow for high urgency
            if urgencyLevel >= 2 {
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [theme.primaryColor.opacity(0.4), theme.primaryColor.opacity(0)],
                            center: .center,
                            startRadius: 20,
                            endRadius: 45
                        )
                    )
                    .frame(width: 90, height: 90)
            }

            // Progress ring (fills up as break becomes more needed)
            Circle()
                .stroke(theme.primaryColor.opacity(0.2), lineWidth: 8)
                .frame(width: 64, height: 64)

            Circle()
                .trim(from: 0, to: progress)
                .stroke(
                    AngularGradient(
                        colors: urgencyLevel >= 2 ? [.red, .orange, .red] : [.teal, .cyan, .teal],
                        center: .center,
                        startAngle: .degrees(-90),
                        endAngle: .degrees(270)
                    ),
                    style: StrokeStyle(lineWidth: 8, lineCap: .round)
                )
                .frame(width: 64, height: 64)
                .rotationEffect(.degrees(-90))

            // Center icon
            Image(systemName: theme.icon)
                .font(.system(size: 22))
                .foregroundStyle(theme.gradient)
        }
    }
}

// MARK: - Dynamic Island Views for Break Reminder
@available(iOS 16.1, *)
private struct BreakExpandedLeading: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    var body: some View {
        HStack(spacing: 8) {
            ZStack {
                Circle()
                    .fill(theme.primaryColor.opacity(0.2))
                    .frame(width: 28, height: 28)

                Image(systemName: theme.icon)
                    .font(.system(size: 14))
                    .foregroundColor(theme.primaryColor)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text("Break")
                    .font(.system(size: 14, weight: .semibold))

                Text("\(context.state.minutesSinceLastBreak)m ago")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }
        }
    }
}

@available(iOS 16.1, *)
private struct BreakExpandedTrailing: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            Text(theme.actionText)
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(theme.gradient)

            Text("\(context.attributes.recommendedBreakMinutes)min")
                .font(.system(size: 10))
                .foregroundColor(.secondary)
        }
    }
}

@available(iOS 16.1, *)
private struct BreakExpandedCenter: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    var body: some View {
        if context.state.urgencyLevel >= 2 {
            HStack(spacing: 4) {
                Image(systemName: "heart.fill")
                    .font(.system(size: 10))
                Text("Your health matters")
                    .font(.system(size: 10))
            }
            .foregroundColor(theme.primaryColor)
        }
    }
}

@available(iOS 16.1, *)
private struct BreakExpandedBottom: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    private var progress: Double {
        min(Double(context.state.minutesSinceLastBreak) / Double(max(context.attributes.breakIntervalMinutes, 1)), 1.0)
    }

    var body: some View {
        VStack(spacing: 8) {
            // Urgency progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.gray.opacity(0.3))
                        .frame(height: 6)

                    RoundedRectangle(cornerRadius: 4)
                        .fill(theme.gradient)
                        .frame(width: geometry.size.width * progress, height: 6)
                }
            }
            .frame(height: 6)
            .padding(.horizontal, 4)

            // Info row
            HStack {
                HStack(spacing: 4) {
                    Image(systemName: "cup.and.saucer.fill")
                        .font(.system(size: 10))
                    Text("\(context.state.breaksTakenToday) breaks today")
                        .font(.system(size: 10))
                }
                .foregroundColor(.secondary)

                Spacer()

                let hours = context.state.totalWorkMinutesToday / 60
                let mins = context.state.totalWorkMinutesToday % 60
                Text("\(hours)h \(mins)m worked")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }
            .padding(.horizontal, 4)
        }
    }
}

// MARK: - Compact Views for Break
@available(iOS 16.1, *)
private struct BreakCompactLeading: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    var body: some View {
        Image(systemName: theme.icon)
            .font(.system(size: 14))
            .foregroundColor(theme.primaryColor)
    }
}

@available(iOS 16.1, *)
private struct BreakCompactTrailing: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    var body: some View {
        Text("\(context.state.minutesSinceLastBreak)m")
            .font(.system(size: 13, weight: .bold))
            .foregroundStyle(theme.gradient)
    }
}

@available(iOS 16.1, *)
private struct BreakMinimalView: View {
    let context: ActivityViewContext<BreakReminderActivityAttributes>
    private var theme: BreakTheme { BreakTheme(urgencyLevel: context.state.urgencyLevel) }

    var body: some View {
        ZStack {
            Circle()
                .fill(
                    RadialGradient(
                        colors: [theme.primaryColor.opacity(0.3), Color.clear],
                        center: .center,
                        startRadius: 2,
                        endRadius: 14
                    )
                )

            Image(systemName: context.state.urgencyLevel >= 2 ? "exclamationmark" : "cup.and.saucer.fill")
                .font(.system(size: context.state.urgencyLevel >= 2 ? 12 : 10))
                .foregroundColor(theme.primaryColor)
        }
    }
}
#endif
