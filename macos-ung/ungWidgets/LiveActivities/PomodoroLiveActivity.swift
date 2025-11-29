//
//  PomodoroLiveActivity.swift
//  ungWidgets
//
//  Premium Live Activity for Pomodoro timer on Lock Screen & Dynamic Island
//

#if os(iOS)
import ActivityKit
import SwiftUI
import WidgetKit

// MARK: - Activity Attributes
@available(iOS 16.1, *)
struct PomodoroActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var secondsRemaining: Int
        var isBreak: Bool
        var isPaused: Bool
        var currentSessionNumber: Int // Which session in the cycle (1-4)
    }

    var sessionsCompleted: Int
    var workMinutes: Int
    var breakMinutes: Int
    var longBreakMinutes: Int
    var projectName: String?
}

// MARK: - Live Activity Widget
@available(iOS 16.1, *)
struct PomodoroLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: PomodoroActivityAttributes.self) { context in
            // Lock Screen / Banner view - Premium design
            LockScreenPomodoroView(context: context)
                .activityBackgroundTint(Color.black.opacity(0.85))
                .activitySystemActionForegroundColor(Color.white)
        } dynamicIsland: { context in
            DynamicIsland {
                // Expanded Dynamic Island View
                DynamicIslandExpandedRegion(.leading) {
                    PomodoroExpandedLeading(context: context)
                }

                DynamicIslandExpandedRegion(.trailing) {
                    PomodoroExpandedTrailing(context: context)
                }

                DynamicIslandExpandedRegion(.center) {
                    PomodoroExpandedCenter(context: context)
                }

                DynamicIslandExpandedRegion(.bottom) {
                    PomodoroExpandedBottom(context: context)
                }
            } compactLeading: {
                PomodoroCompactLeading(context: context)
            } compactTrailing: {
                PomodoroCompactTrailing(context: context)
            } minimal: {
                PomodoroMinimalView(context: context)
            }
            .widgetURL(URL(string: "ung://pomodoro"))
            .keylineTint(context.state.isBreak ? Color.green : Color.orange)
        }
    }
}

// MARK: - Theme Colors
@available(iOS 16.1, *)
private struct PomodoroTheme {
    let context: ActivityViewContext<PomodoroActivityAttributes>

    var primaryColor: Color {
        context.state.isBreak ? .green : .orange
    }

    var gradientColors: [Color] {
        context.state.isBreak
            ? [Color.green, Color.teal]
            : [Color.orange, Color.red]
    }

    var icon: String {
        if context.state.isPaused {
            return "pause.circle.fill"
        }
        return context.state.isBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill"
    }

    var statusText: String {
        if context.state.isPaused {
            return "PAUSED"
        }
        return context.state.isBreak ? "Break Time" : "Focus Mode"
    }
}

// MARK: - Lock Screen View
@available(iOS 16.1, *)
private struct LockScreenPomodoroView: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>

    private var theme: PomodoroTheme {
        PomodoroTheme(context: context)
    }

    var body: some View {
        HStack(spacing: 16) {
            // Premium progress ring
            PomodoroProgressRing(context: context)

            // Status info
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 8) {
                    Text(theme.statusText)
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.white)

                    if context.state.isPaused {
                        PausedBadge()
                    }
                }

                // Session progress dots
                SessionProgressDots(
                    completed: context.attributes.sessionsCompleted,
                    current: context.state.currentSessionNumber,
                    isBreak: context.state.isBreak
                )

                if let project = context.attributes.projectName, !project.isEmpty {
                    HStack(spacing: 4) {
                        Image(systemName: "folder.fill")
                            .font(.system(size: 10))
                            .foregroundColor(.gray)
                        Text(project)
                            .font(.system(size: 12))
                            .foregroundColor(.gray)
                            .lineLimit(1)
                    }
                }
            }

            Spacer()

            // Large timer display
            VStack(alignment: .trailing, spacing: 4) {
                Text(formatTime(context.state.secondsRemaining))
                    .font(.system(size: 36, weight: .bold, design: .monospaced))
                    .foregroundStyle(
                        LinearGradient(
                            colors: theme.gradientColors,
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    )
                    .monospacedDigit()
                    .contentTransition(.numericText())

                Text(context.state.isBreak ? "until focus" : "remaining")
                    .font(.system(size: 11))
                    .foregroundColor(.gray)
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
    }

    private func formatTime(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d", minutes, secs)
    }
}

// MARK: - Progress Ring
@available(iOS 16.1, *)
private struct PomodoroProgressRing: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>

    private var progress: Double {
        let totalSeconds: Int
        if context.state.isBreak {
            // Check if it's a long break (every 4 sessions)
            let isLongBreak = context.attributes.sessionsCompleted > 0 &&
                              context.attributes.sessionsCompleted % 4 == 0
            totalSeconds = isLongBreak
                ? context.attributes.longBreakMinutes * 60
                : context.attributes.breakMinutes * 60
        } else {
            totalSeconds = context.attributes.workMinutes * 60
        }
        return 1.0 - (Double(context.state.secondsRemaining) / Double(max(totalSeconds, 1)))
    }

    private var color: Color {
        context.state.isBreak ? .green : .orange
    }

    var body: some View {
        ZStack {
            // Background ring
            Circle()
                .stroke(color.opacity(0.2), lineWidth: 6)
                .frame(width: 64, height: 64)

            // Progress ring with gradient
            Circle()
                .trim(from: 0, to: progress)
                .stroke(
                    AngularGradient(
                        colors: context.state.isBreak
                            ? [.green, .teal, .green]
                            : [.orange, .red, .orange],
                        center: .center,
                        startAngle: .degrees(-90),
                        endAngle: .degrees(270)
                    ),
                    style: StrokeStyle(lineWidth: 6, lineCap: .round)
                )
                .frame(width: 64, height: 64)
                .rotationEffect(.degrees(-90))
                .shadow(color: color.opacity(0.5), radius: 4)

            // Center icon
            Image(systemName: context.state.isBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill")
                .font(.system(size: 22))
                .foregroundStyle(
                    LinearGradient(
                        colors: context.state.isBreak ? [.green, .teal] : [.orange, .red],
                        startPoint: .top,
                        endPoint: .bottom
                    )
                )
        }
    }
}

// MARK: - Session Progress Dots
@available(iOS 16.1, *)
private struct SessionProgressDots: View {
    let completed: Int
    let current: Int
    let isBreak: Bool

    var body: some View {
        HStack(spacing: 6) {
            ForEach(1...4, id: \.self) { index in
                let sessionInCycle = (completed % 4) + (isBreak ? 0 : 1)
                let isFilled = index <= sessionInCycle

                Circle()
                    .fill(isFilled ? (isBreak ? Color.green : Color.orange) : Color.gray.opacity(0.3))
                    .frame(width: 10, height: 10)
                    .overlay {
                        if index == current && !isBreak {
                            Circle()
                                .stroke(Color.orange, lineWidth: 2)
                                .frame(width: 14, height: 14)
                        }
                    }
            }

            Text("·")
                .foregroundColor(.gray)

            Text("\(completed) total")
                .font(.system(size: 11))
                .foregroundColor(.gray)
        }
    }
}

// MARK: - Paused Badge
@available(iOS 16.1, *)
private struct PausedBadge: View {
    var body: some View {
        HStack(spacing: 3) {
            Image(systemName: "pause.fill")
                .font(.system(size: 8))
            Text("PAUSED")
                .font(.system(size: 9, weight: .bold))
        }
        .foregroundColor(.yellow)
        .padding(.horizontal, 6)
        .padding(.vertical, 3)
        .background(Color.yellow.opacity(0.2))
        .cornerRadius(4)
    }
}

// MARK: - Dynamic Island Expanded Views
@available(iOS 16.1, *)
private struct PomodoroExpandedLeading: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>
    private var theme: PomodoroTheme { PomodoroTheme(context: context) }

    var body: some View {
        HStack(spacing: 8) {
            // Icon with glow
            ZStack {
                Circle()
                    .fill(theme.primaryColor.opacity(0.2))
                    .frame(width: 28, height: 28)

                Image(systemName: theme.icon)
                    .font(.system(size: 14))
                    .foregroundColor(theme.primaryColor)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(context.state.isBreak ? "Break" : "Focus")
                    .font(.system(size: 14, weight: .semibold))

                if context.state.isPaused {
                    Text("Paused")
                        .font(.system(size: 10))
                        .foregroundColor(.yellow)
                } else if let project = context.attributes.projectName {
                    Text(project)
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                        .lineLimit(1)
                }
            }
        }
    }
}

@available(iOS 16.1, *)
private struct PomodoroExpandedTrailing: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>
    private var theme: PomodoroTheme { PomodoroTheme(context: context) }

    var body: some View {
        Text(formatTime(context.state.secondsRemaining))
            .font(.system(size: 26, weight: .bold, design: .monospaced))
            .foregroundStyle(
                LinearGradient(
                    colors: theme.gradientColors,
                    startPoint: .leading,
                    endPoint: .trailing
                )
            )
            .monospacedDigit()
            .contentTransition(.numericText())
    }

    private func formatTime(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d", minutes, secs)
    }
}

@available(iOS 16.1, *)
private struct PomodoroExpandedCenter: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>

    var body: some View {
        if context.state.isPaused {
            HStack(spacing: 4) {
                Image(systemName: "pause.fill")
                    .font(.system(size: 10))
                Text("Tap to resume")
                    .font(.system(size: 10))
            }
            .foregroundColor(.yellow)
        }
    }
}

@available(iOS 16.1, *)
private struct PomodoroExpandedBottom: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>
    private var theme: PomodoroTheme { PomodoroTheme(context: context) }

    private var progress: Double {
        let totalSeconds: Int
        if context.state.isBreak {
            totalSeconds = context.attributes.breakMinutes * 60
        } else {
            totalSeconds = context.attributes.workMinutes * 60
        }
        return 1.0 - (Double(context.state.secondsRemaining) / Double(max(totalSeconds, 1)))
    }

    var body: some View {
        VStack(spacing: 8) {
            // Premium progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    // Background
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.gray.opacity(0.3))
                        .frame(height: 6)

                    // Progress with gradient
                    RoundedRectangle(cornerRadius: 4)
                        .fill(
                            LinearGradient(
                                colors: theme.gradientColors,
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )
                        .frame(width: geometry.size.width * progress, height: 6)
                        .shadow(color: theme.primaryColor.opacity(0.5), radius: 2)
                }
            }
            .frame(height: 6)
            .padding(.horizontal, 4)

            // Session info row
            HStack {
                // Session dots
                HStack(spacing: 4) {
                    ForEach(0..<4, id: \.self) { index in
                        Circle()
                            .fill(
                                index < context.attributes.sessionsCompleted % 4
                                    ? theme.primaryColor
                                    : Color.gray.opacity(0.3)
                            )
                            .frame(width: 8, height: 8)
                    }
                }

                Spacer()

                // Session counter
                HStack(spacing: 4) {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.green)
                    Text("\(context.attributes.sessionsCompleted) done")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }
            }
            .padding(.horizontal, 4)
        }
    }
}

// MARK: - Compact Views
@available(iOS 16.1, *)
private struct PomodoroCompactLeading: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>
    private var theme: PomodoroTheme { PomodoroTheme(context: context) }

    var body: some View {
        ZStack {
            // Mini progress ring
            Circle()
                .stroke(theme.primaryColor.opacity(0.3), lineWidth: 2)
                .frame(width: 20, height: 20)

            Circle()
                .trim(from: 0, to: calculateProgress())
                .stroke(theme.primaryColor, style: StrokeStyle(lineWidth: 2, lineCap: .round))
                .frame(width: 20, height: 20)
                .rotationEffect(.degrees(-90))

            Image(systemName: context.state.isBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill")
                .font(.system(size: 8))
                .foregroundColor(theme.primaryColor)
        }
    }

    private func calculateProgress() -> Double {
        let totalSeconds: Int
        if context.state.isBreak {
            totalSeconds = context.attributes.breakMinutes * 60
        } else {
            totalSeconds = context.attributes.workMinutes * 60
        }
        return 1.0 - (Double(context.state.secondsRemaining) / Double(max(totalSeconds, 1)))
    }
}

@available(iOS 16.1, *)
private struct PomodoroCompactTrailing: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>
    private var theme: PomodoroTheme { PomodoroTheme(context: context) }

    var body: some View {
        HStack(spacing: 2) {
            if context.state.isPaused {
                Image(systemName: "pause.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.yellow)
            }

            Text(formatTimeCompact(context.state.secondsRemaining))
                .font(.system(size: 13, weight: .bold, design: .monospaced))
                .foregroundStyle(
                    context.state.isPaused
                        ? AnyShapeStyle(Color.yellow)
                        : AnyShapeStyle(LinearGradient(colors: theme.gradientColors, startPoint: .leading, endPoint: .trailing))
                )
                .monospacedDigit()
                .contentTransition(.numericText())
        }
    }

    private func formatTimeCompact(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let secs = seconds % 60
        return String(format: "%d:%02d", minutes, secs)
    }
}

// MARK: - Minimal View
@available(iOS 16.1, *)
private struct PomodoroMinimalView: View {
    let context: ActivityViewContext<PomodoroActivityAttributes>
    private var theme: PomodoroTheme { PomodoroTheme(context: context) }

    private var progress: Double {
        let totalSeconds: Int
        if context.state.isBreak {
            totalSeconds = context.attributes.breakMinutes * 60
        } else {
            totalSeconds = context.attributes.workMinutes * 60
        }
        return 1.0 - (Double(context.state.secondsRemaining) / Double(max(totalSeconds, 1)))
    }

    var body: some View {
        ZStack {
            // Background glow
            Circle()
                .fill(
                    RadialGradient(
                        colors: [theme.primaryColor.opacity(0.3), Color.clear],
                        center: .center,
                        startRadius: 2,
                        endRadius: 14
                    )
                )

            // Progress ring
            Circle()
                .trim(from: 0, to: progress)
                .stroke(
                    AngularGradient(
                        colors: theme.gradientColors + [theme.gradientColors[0]],
                        center: .center
                    ),
                    style: StrokeStyle(lineWidth: 2.5, lineCap: .round)
                )
                .frame(width: 18, height: 18)
                .rotationEffect(.degrees(-90))

            // Center indicator
            if context.state.isPaused {
                Image(systemName: "pause.fill")
                    .font(.system(size: 8))
                    .foregroundColor(.yellow)
            } else {
                Circle()
                    .fill(theme.primaryColor)
                    .frame(width: 6, height: 6)
            }
        }
    }
}
#endif
