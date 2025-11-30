//
//  MilestoneLiveActivity.swift
//  ungWidgets
//
//  Premium Live Activity for celebrating earnings milestones
//  "Watch your money grow in real-time" - Marketing hook
//

#if os(iOS)
import ActivityKit
import SwiftUI
import WidgetKit

// MARK: - Milestone Activity Attributes
@available(iOS 16.1, *)
struct MilestoneActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var currentAmount: Double
        var isAchieved: Bool
        var celebrationPhase: Int // 0: approaching, 1: achieved, 2: exceeded
    }

    var milestoneType: String // "daily", "weekly", "monthly"
    var targetAmount: Double
    var currency: String
    var milestoneName: String // e.g., "$100 day", "Weekly goal"
}

// MARK: - Milestone Live Activity Widget
@available(iOS 16.1, *)
struct MilestoneLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: MilestoneActivityAttributes.self) { context in
            // Lock Screen / Banner view
            MilestoneLockScreenView(context: context)
                .activityBackgroundTint(context.state.isAchieved ? Color.green.opacity(0.85) : Color.black.opacity(0.85))
                .activitySystemActionForegroundColor(Color.white)
        } dynamicIsland: { context in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    MilestoneExpandedLeading(context: context)
                }

                DynamicIslandExpandedRegion(.trailing) {
                    MilestoneExpandedTrailing(context: context)
                }

                DynamicIslandExpandedRegion(.center) {
                    MilestoneExpandedCenter(context: context)
                }

                DynamicIslandExpandedRegion(.bottom) {
                    MilestoneExpandedBottom(context: context)
                }
            } compactLeading: {
                MilestoneCompactLeading(context: context)
            } compactTrailing: {
                MilestoneCompactTrailing(context: context)
            } minimal: {
                MilestoneMinimalView(context: context)
            }
            .widgetURL(URL(string: "ung://dashboard"))
            .keylineTint(context.state.isAchieved ? Color.green : Color.yellow)
        }
    }
}

// MARK: - Lock Screen View
@available(iOS 16.1, *)
private struct MilestoneLockScreenView: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>

    private var progress: Double {
        min(context.state.currentAmount / max(context.attributes.targetAmount, 1), 1.5)
    }

    private var theme: MilestoneTheme {
        MilestoneTheme(isAchieved: context.state.isAchieved, celebrationPhase: context.state.celebrationPhase)
    }

    var body: some View {
        HStack(spacing: 16) {
            // Progress ring with celebration effects
            MilestoneProgressRing(
                progress: progress,
                isAchieved: context.state.isAchieved,
                celebrationPhase: context.state.celebrationPhase
            )

            // Milestone info
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 8) {
                    Text(context.state.isAchieved ? "🎉 Goal Achieved!" : context.attributes.milestoneName)
                        .font(.system(size: 17, weight: .semibold))
                        .foregroundColor(.white)

                    if context.state.celebrationPhase == 2 {
                        Text("EXCEEDED")
                            .font(.system(size: 9, weight: .bold))
                            .foregroundColor(.yellow)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.yellow.opacity(0.2))
                            .cornerRadius(4)
                    }
                }

                // Progress indicator
                Text("\(milestoneTypeLabel) goal: \(context.attributes.currency)\(Int(context.attributes.targetAmount))")
                    .font(.system(size: 13))
                    .foregroundColor(.gray)
            }

            Spacer()

            // Current amount display
            VStack(alignment: .trailing, spacing: 4) {
                Text("\(context.attributes.currency)\(context.state.currentAmount, specifier: "%.0f")")
                    .font(.system(size: 28, weight: .bold, design: .rounded))
                    .foregroundStyle(theme.gradient)
                    .contentTransition(.numericText())

                let remaining = context.attributes.targetAmount - context.state.currentAmount
                if remaining > 0 {
                    Text("\(context.attributes.currency)\(remaining, specifier: "%.0f") to go")
                        .font(.system(size: 11))
                        .foregroundColor(.gray)
                } else {
                    Text("+\(context.attributes.currency)\(abs(remaining), specifier: "%.0f") over")
                        .font(.system(size: 11))
                        .foregroundColor(.green)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
    }

    private var milestoneTypeLabel: String {
        switch context.attributes.milestoneType {
        case "daily": return "Daily"
        case "weekly": return "Weekly"
        case "monthly": return "Monthly"
        default: return "Goal"
        }
    }
}

// MARK: - Milestone Theme
@available(iOS 16.1, *)
private struct MilestoneTheme {
    let isAchieved: Bool
    let celebrationPhase: Int

    var gradient: LinearGradient {
        if isAchieved {
            return LinearGradient(
                colors: [Color.green, Color.mint],
                startPoint: .leading,
                endPoint: .trailing
            )
        } else if celebrationPhase == 0 {
            return LinearGradient(
                colors: [Color.yellow, Color.orange],
                startPoint: .leading,
                endPoint: .trailing
            )
        } else {
            return LinearGradient(
                colors: [Color.green, Color.cyan],
                startPoint: .leading,
                endPoint: .trailing
            )
        }
    }

    var primaryColor: Color {
        isAchieved ? .green : .yellow
    }

    var icon: String {
        if isAchieved {
            return "checkmark.seal.fill"
        } else if celebrationPhase == 0 {
            return "target"
        } else {
            return "star.fill"
        }
    }
}

// MARK: - Progress Ring
@available(iOS 16.1, *)
private struct MilestoneProgressRing: View {
    let progress: Double
    let isAchieved: Bool
    let celebrationPhase: Int

    private var theme: MilestoneTheme {
        MilestoneTheme(isAchieved: isAchieved, celebrationPhase: celebrationPhase)
    }

    var body: some View {
        ZStack {
            // Celebration glow when achieved
            if isAchieved {
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [Color.green.opacity(0.4), Color.green.opacity(0)],
                            center: .center,
                            startRadius: 20,
                            endRadius: 45
                        )
                    )
                    .frame(width: 90, height: 90)
            }

            // Background ring
            Circle()
                .stroke(theme.primaryColor.opacity(0.2), lineWidth: 8)
                .frame(width: 64, height: 64)

            // Progress ring
            Circle()
                .trim(from: 0, to: min(progress, 1.0))
                .stroke(
                    AngularGradient(
                        colors: isAchieved
                            ? [.green, .mint, .green]
                            : [.yellow, .orange, .yellow],
                        center: .center,
                        startAngle: .degrees(-90),
                        endAngle: .degrees(270)
                    ),
                    style: StrokeStyle(lineWidth: 8, lineCap: .round)
                )
                .frame(width: 64, height: 64)
                .rotationEffect(.degrees(-90))
                .shadow(color: theme.primaryColor.opacity(0.5), radius: 4)

            // Overflow indicator (when exceeded)
            if progress > 1.0 {
                Circle()
                    .trim(from: 0, to: min(progress - 1.0, 0.5))
                    .stroke(
                        Color.cyan,
                        style: StrokeStyle(lineWidth: 4, lineCap: .round)
                    )
                    .frame(width: 52, height: 52)
                    .rotationEffect(.degrees(-90))
            }

            // Center icon
            Image(systemName: theme.icon)
                .font(.system(size: 24))
                .foregroundStyle(theme.gradient)
        }
    }
}

// MARK: - Dynamic Island Expanded Views
@available(iOS 16.1, *)
private struct MilestoneExpandedLeading: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>
    private var theme: MilestoneTheme {
        MilestoneTheme(isAchieved: context.state.isAchieved, celebrationPhase: context.state.celebrationPhase)
    }

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
                Text(context.state.isAchieved ? "Achieved!" : "Goal")
                    .font(.system(size: 14, weight: .semibold))

                Text(context.attributes.milestoneName)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
                    .lineLimit(1)
            }
        }
    }
}

@available(iOS 16.1, *)
private struct MilestoneExpandedTrailing: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>
    private var theme: MilestoneTheme {
        MilestoneTheme(isAchieved: context.state.isAchieved, celebrationPhase: context.state.celebrationPhase)
    }

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            Text("\(context.attributes.currency)\(context.state.currentAmount, specifier: "%.0f")")
                .font(.system(size: 22, weight: .bold, design: .rounded))
                .foregroundStyle(theme.gradient)
                .contentTransition(.numericText())

            let progress = (context.state.currentAmount / context.attributes.targetAmount) * 100
            Text("\(Int(min(progress, 999)))%")
                .font(.system(size: 10, weight: .medium))
                .foregroundColor(context.state.isAchieved ? .green : .orange)
        }
    }
}

@available(iOS 16.1, *)
private struct MilestoneExpandedCenter: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>

    var body: some View {
        if context.state.isAchieved {
            HStack(spacing: 4) {
                Image(systemName: "sparkles")
                    .font(.system(size: 10))
                Text("MILESTONE REACHED")
                    .font(.system(size: 10, weight: .bold))
            }
            .foregroundColor(.green)
        }
    }
}

@available(iOS 16.1, *)
private struct MilestoneExpandedBottom: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>

    private var progress: Double {
        min(context.state.currentAmount / max(context.attributes.targetAmount, 1), 1.0)
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
                            LinearGradient(
                                colors: context.state.isAchieved ? [.green, .mint] : [.yellow, .orange],
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )
                        .frame(width: geometry.size.width * progress, height: 6)
                }
            }
            .frame(height: 6)
            .padding(.horizontal, 4)

            // Info row
            HStack {
                Text(context.attributes.milestoneType.capitalized)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)

                Spacer()

                let remaining = max(context.attributes.targetAmount - context.state.currentAmount, 0)
                if remaining > 0 {
                    Text("\(context.attributes.currency)\(remaining, specifier: "%.0f") remaining")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                } else {
                    Text("Goal exceeded!")
                        .font(.system(size: 10))
                        .foregroundColor(.green)
                }
            }
            .padding(.horizontal, 4)
        }
    }
}

// MARK: - Compact Views
@available(iOS 16.1, *)
private struct MilestoneCompactLeading: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>
    private var theme: MilestoneTheme {
        MilestoneTheme(isAchieved: context.state.isAchieved, celebrationPhase: context.state.celebrationPhase)
    }

    var body: some View {
        ZStack {
            // Mini progress ring
            Circle()
                .stroke(theme.primaryColor.opacity(0.3), lineWidth: 2)
                .frame(width: 20, height: 20)

            let progress = min(context.state.currentAmount / context.attributes.targetAmount, 1.0)
            Circle()
                .trim(from: 0, to: progress)
                .stroke(theme.primaryColor, style: StrokeStyle(lineWidth: 2, lineCap: .round))
                .frame(width: 20, height: 20)
                .rotationEffect(.degrees(-90))

            if context.state.isAchieved {
                Image(systemName: "checkmark")
                    .font(.system(size: 8, weight: .bold))
                    .foregroundColor(.green)
            } else {
                Image(systemName: "dollarsign")
                    .font(.system(size: 8))
                    .foregroundColor(theme.primaryColor)
            }
        }
    }
}

@available(iOS 16.1, *)
private struct MilestoneCompactTrailing: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>
    private var theme: MilestoneTheme {
        MilestoneTheme(isAchieved: context.state.isAchieved, celebrationPhase: context.state.celebrationPhase)
    }

    var body: some View {
        let progress = (context.state.currentAmount / context.attributes.targetAmount) * 100
        Text("\(Int(min(progress, 999)))%")
            .font(.system(size: 13, weight: .bold, design: .rounded))
            .foregroundStyle(theme.gradient)
            .contentTransition(.numericText())
    }
}

// MARK: - Minimal View
@available(iOS 16.1, *)
private struct MilestoneMinimalView: View {
    let context: ActivityViewContext<MilestoneActivityAttributes>

    var body: some View {
        ZStack {
            if context.state.isAchieved {
                // Celebration glow
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [Color.green.opacity(0.4), Color.clear],
                            center: .center,
                            startRadius: 2,
                            endRadius: 14
                        )
                    )

                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 14))
                    .foregroundColor(.green)
            } else {
                // Progress indicator
                let progress = min(context.state.currentAmount / context.attributes.targetAmount, 1.0)
                Circle()
                    .trim(from: 0, to: progress)
                    .stroke(Color.yellow, style: StrokeStyle(lineWidth: 2.5, lineCap: .round))
                    .frame(width: 18, height: 18)
                    .rotationEffect(.degrees(-90))

                Circle()
                    .fill(Color.yellow)
                    .frame(width: 6, height: 6)
            }
        }
    }
}
#endif
