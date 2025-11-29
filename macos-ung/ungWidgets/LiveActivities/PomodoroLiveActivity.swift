//
//  PomodoroLiveActivity.swift
//  ungWidgets
//
//  Live Activity for real-time pomodoro timer on Lock Screen & Dynamic Island
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
    }

    var sessionsCompleted: Int
    var workMinutes: Int
    var breakMinutes: Int
}

// MARK: - Live Activity Widget
@available(iOS 16.1, *)
struct PomodoroLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: PomodoroActivityAttributes.self) { context in
            // Lock Screen / Banner view
            lockScreenView(context: context)
        } dynamicIsland: { context in
            DynamicIsland {
                // Expanded view
                DynamicIslandExpandedRegion(.leading) {
                    HStack(spacing: 8) {
                        Image(systemName: context.state.isBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill")
                            .font(.system(size: 16))
                            .foregroundColor(context.state.isBreak ? .green : .orange)

                        Text(context.state.isBreak ? "Break" : "Focus")
                            .font(.system(size: 14, weight: .semibold))
                    }
                }

                DynamicIslandExpandedRegion(.trailing) {
                    Text(formatTime(context.state.secondsRemaining))
                        .font(.system(size: 24, weight: .bold, design: .monospaced))
                        .foregroundColor(context.state.isBreak ? .green : .orange)
                        .monospacedDigit()
                }

                DynamicIslandExpandedRegion(.center) {
                    if context.state.isPaused {
                        Label("Paused", systemImage: "pause.fill")
                            .font(.system(size: 12))
                            .foregroundColor(.yellow)
                    }
                }

                DynamicIslandExpandedRegion(.bottom) {
                    // Progress bar
                    GeometryReader { geometry in
                        ZStack(alignment: .leading) {
                            RoundedRectangle(cornerRadius: 4)
                                .fill(Color.gray.opacity(0.3))
                                .frame(height: 6)

                            RoundedRectangle(cornerRadius: 4)
                                .fill(context.state.isBreak ? Color.green : Color.orange)
                                .frame(
                                    width: geometry.size.width * calculateProgress(context: context),
                                    height: 6
                                )
                        }
                    }
                    .frame(height: 6)
                    .padding(.horizontal, 4)

                    // Session indicators
                    HStack(spacing: 4) {
                        ForEach(0..<4, id: \.self) { index in
                            Circle()
                                .fill(index < context.attributes.sessionsCompleted % 4 ? Color.orange : Color.gray.opacity(0.3))
                                .frame(width: 8, height: 8)
                        }

                        Spacer()

                        Text("\(context.attributes.sessionsCompleted) sessions")
                            .font(.system(size: 10))
                            .foregroundColor(.secondary)
                    }
                    .padding(.horizontal, 4)
                    .padding(.top, 8)
                }
            } compactLeading: {
                // Compact leading
                Image(systemName: context.state.isBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill")
                    .font(.system(size: 14))
                    .foregroundColor(context.state.isBreak ? .green : .orange)
            } compactTrailing: {
                // Compact trailing
                Text(formatTimeCompact(context.state.secondsRemaining))
                    .font(.system(size: 12, weight: .bold, design: .monospaced))
                    .foregroundColor(context.state.isBreak ? .green : .orange)
                    .monospacedDigit()
            } minimal: {
                // Minimal view
                ZStack {
                    Circle()
                        .stroke(context.state.isBreak ? Color.green : Color.orange, lineWidth: 2)

                    // Mini progress
                    Circle()
                        .trim(from: 0, to: calculateProgress(context: context))
                        .stroke(context.state.isBreak ? Color.green : Color.orange, style: StrokeStyle(lineWidth: 2, lineCap: .round))
                        .rotationEffect(.degrees(-90))
                }
            }
        }
    }

    @ViewBuilder
    private func lockScreenView(context: ActivityViewContext<PomodoroActivityAttributes>) -> some View {
        HStack(spacing: 16) {
            // Progress ring
            ZStack {
                Circle()
                    .stroke(Color.gray.opacity(0.3), lineWidth: 6)
                    .frame(width: 60, height: 60)

                Circle()
                    .trim(from: 0, to: calculateProgress(context: context))
                    .stroke(
                        context.state.isBreak ? Color.green : Color.orange,
                        style: StrokeStyle(lineWidth: 6, lineCap: .round)
                    )
                    .frame(width: 60, height: 60)
                    .rotationEffect(.degrees(-90))

                Image(systemName: context.state.isBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill")
                    .font(.system(size: 20))
                    .foregroundColor(context.state.isBreak ? .green : .orange)
            }

            // Status
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text(context.state.isBreak ? "Break Time" : "Focus Mode")
                        .font(.system(size: 16, weight: .semibold))

                    if context.state.isPaused {
                        Text("PAUSED")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundColor(.yellow)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.yellow.opacity(0.2))
                            .cornerRadius(4)
                    }
                }

                HStack(spacing: 3) {
                    ForEach(0..<4, id: \.self) { index in
                        RoundedRectangle(cornerRadius: 2)
                            .fill(index < context.attributes.sessionsCompleted % 4 ? Color.orange : Color.gray.opacity(0.3))
                            .frame(width: 16, height: 4)
                    }
                    Text("\(context.attributes.sessionsCompleted)")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            // Timer
            Text(formatTime(context.state.secondsRemaining))
                .font(.system(size: 32, weight: .bold, design: .monospaced))
                .foregroundColor(context.state.isBreak ? .green : .orange)
                .monospacedDigit()
        }
        .padding()
        .activityBackgroundTint(Color.black.opacity(0.8))
        .activitySystemActionForegroundColor(Color.white)
    }

    private func calculateProgress(context: ActivityViewContext<PomodoroActivityAttributes>) -> Double {
        let totalSeconds: Int
        if context.state.isBreak {
            totalSeconds = context.attributes.breakMinutes * 60
        } else {
            totalSeconds = context.attributes.workMinutes * 60
        }
        return 1.0 - (Double(context.state.secondsRemaining) / Double(totalSeconds))
    }

    private func formatTime(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d", minutes, secs)
    }

    private func formatTimeCompact(_ seconds: Int) -> String {
        let minutes = seconds / 60
        let secs = seconds % 60
        return String(format: "%d:%02d", minutes, secs)
    }
}
#endif
