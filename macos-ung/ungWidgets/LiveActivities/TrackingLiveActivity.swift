//
//  TrackingLiveActivity.swift
//  ungWidgets
//
//  Premium Live Activity for real-time tracking display on Lock Screen & Dynamic Island
//

#if os(iOS)
import ActivityKit
import SwiftUI
import WidgetKit

// MARK: - Activity Attributes
@available(iOS 16.1, *)
struct TrackingActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var elapsedSeconds: Int
        var isActive: Bool
        var billableAmount: Double? // Optional real-time earnings
    }

    var projectName: String
    var clientName: String
    var startTime: Date
    var hourlyRate: Double?
    var currency: String
}

// MARK: - Live Activity Widget
@available(iOS 16.1, *)
struct TrackingLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: TrackingActivityAttributes.self) { context in
            // Lock Screen / Banner view - Premium design
            LockScreenTrackingView(context: context)
                .activityBackgroundTint(Color.black.opacity(0.85))
                .activitySystemActionForegroundColor(Color.white)
        } dynamicIsland: { context in
            DynamicIsland {
                // Expanded Dynamic Island View
                DynamicIslandExpandedRegion(.leading) {
                    ExpandedLeadingView(context: context)
                }

                DynamicIslandExpandedRegion(.trailing) {
                    ExpandedTrailingView(context: context)
                }

                DynamicIslandExpandedRegion(.center) {
                    ExpandedCenterView(context: context)
                }

                DynamicIslandExpandedRegion(.bottom) {
                    ExpandedBottomView(context: context)
                }
            } compactLeading: {
                CompactLeadingView(context: context)
            } compactTrailing: {
                CompactTrailingView(context: context)
            } minimal: {
                MinimalView(context: context)
            }
            .widgetURL(URL(string: "ung://tracking/active"))
            .keylineTint(Color.red)
        }
    }
}

// MARK: - Lock Screen View
@available(iOS 16.1, *)
private struct LockScreenTrackingView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        HStack(spacing: 16) {
            // Animated recording indicator
            RecordingIndicator(isActive: context.state.isActive)

            // Project & client info
            VStack(alignment: .leading, spacing: 4) {
                Text(context.attributes.projectName)
                    .font(.system(size: 17, weight: .semibold))
                    .lineLimit(1)
                    .foregroundColor(.white)

                if !context.attributes.clientName.isEmpty {
                    HStack(spacing: 4) {
                        Image(systemName: "building.2.fill")
                            .font(.system(size: 10))
                            .foregroundColor(.gray)
                        Text(context.attributes.clientName)
                            .font(.system(size: 13))
                            .foregroundColor(.gray)
                            .lineLimit(1)
                    }
                }
            }

            Spacer()

            // Timer display with optional earnings
            VStack(alignment: .trailing, spacing: 4) {
                Text(formatDuration(context.state.elapsedSeconds))
                    .font(.system(size: 28, weight: .bold, design: .monospaced))
                    .foregroundStyle(
                        LinearGradient(
                            colors: [.red, .orange],
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    )
                    .monospacedDigit()
                    .contentTransition(.numericText())

                if let rate = context.attributes.hourlyRate, rate > 0 {
                    let earnings = (Double(context.state.elapsedSeconds) / 3600.0) * rate
                    Text("\(context.attributes.currency) \(earnings, specifier: "%.2f")")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.green)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
    }

    private func formatDuration(_ seconds: Int) -> String {
        let hours = seconds / 3600
        let minutes = (seconds % 3600) / 60
        let secs = seconds % 60
        return String(format: "%d:%02d:%02d", hours, minutes, secs)
    }
}

// MARK: - Recording Indicator
@available(iOS 16.1, *)
private struct RecordingIndicator: View {
    let isActive: Bool

    var body: some View {
        ZStack {
            // Outer glow ring
            Circle()
                .fill(
                    RadialGradient(
                        colors: [Color.red.opacity(0.4), Color.red.opacity(0)],
                        center: .center,
                        startRadius: 15,
                        endRadius: 30
                    )
                )
                .frame(width: 60, height: 60)

            // Pulsing ring
            Circle()
                .stroke(Color.red.opacity(0.3), lineWidth: 2)
                .frame(width: 44, height: 44)

            // Inner fill
            Circle()
                .fill(
                    RadialGradient(
                        colors: [Color.red, Color.red.opacity(0.8)],
                        center: .center,
                        startRadius: 0,
                        endRadius: 10
                    )
                )
                .frame(width: 16, height: 16)
                .shadow(color: .red.opacity(0.8), radius: 8, x: 0, y: 0)
        }
    }
}

// MARK: - Dynamic Island Expanded Views
@available(iOS 16.1, *)
private struct ExpandedLeadingView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        HStack(spacing: 8) {
            // Pulsing red dot
            ZStack {
                Circle()
                    .fill(Color.red.opacity(0.3))
                    .frame(width: 16, height: 16)
                Circle()
                    .fill(Color.red)
                    .frame(width: 8, height: 8)
                    .shadow(color: .red, radius: 4)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(context.attributes.projectName)
                    .font(.system(size: 14, weight: .semibold))
                    .lineLimit(1)

                if !context.attributes.clientName.isEmpty {
                    Text(context.attributes.clientName)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                        .lineLimit(1)
                }
            }
        }
    }
}

@available(iOS 16.1, *)
private struct ExpandedTrailingView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            Text(formatDuration(context.state.elapsedSeconds))
                .font(.system(size: 22, weight: .bold, design: .monospaced))
                .foregroundStyle(
                    LinearGradient(
                        colors: [.red, .orange],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                )
                .monospacedDigit()
                .contentTransition(.numericText())

            if let rate = context.attributes.hourlyRate, rate > 0 {
                let earnings = (Double(context.state.elapsedSeconds) / 3600.0) * rate
                Text("+\(context.attributes.currency)\(earnings, specifier: "%.2f")")
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(.green)
            }
        }
    }

    private func formatDuration(_ seconds: Int) -> String {
        let hours = seconds / 3600
        let minutes = (seconds % 3600) / 60
        let secs = seconds % 60
        return String(format: "%d:%02d:%02d", hours, minutes, secs)
    }
}

@available(iOS 16.1, *)
private struct ExpandedCenterView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        if context.state.isActive {
            HStack(spacing: 4) {
                Image(systemName: "record.circle")
                    .font(.system(size: 10))
                    .foregroundColor(.red)
                Text("RECORDING")
                    .font(.system(size: 10, weight: .bold))
                    .foregroundColor(.red)
            }
        }
    }
}

@available(iOS 16.1, *)
private struct ExpandedBottomView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        HStack {
            // Start time
            HStack(spacing: 4) {
                Image(systemName: "clock.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
                Text("Started")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
                Text(context.attributes.startTime, style: .time)
                    .font(.system(size: 11, weight: .medium))
            }

            Spacer()

            // Tap hint
            HStack(spacing: 4) {
                Image(systemName: "hand.tap.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
                Text("Tap to stop")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            }
        }
        .padding(.horizontal, 4)
    }
}

// MARK: - Compact Views
@available(iOS 16.1, *)
private struct CompactLeadingView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        HStack(spacing: 4) {
            // Animated recording dot
            ZStack {
                Circle()
                    .fill(Color.red.opacity(0.3))
                    .frame(width: 12, height: 12)
                Circle()
                    .fill(Color.red)
                    .frame(width: 6, height: 6)
            }

            // Project abbreviation
            Text(abbreviate(context.attributes.projectName))
                .font(.system(size: 12, weight: .semibold))
                .lineLimit(1)
        }
    }

    private func abbreviate(_ name: String) -> String {
        let words = name.split(separator: " ")
        if words.count >= 2 {
            return String(words[0].prefix(1) + words[1].prefix(1)).uppercased()
        }
        return String(name.prefix(3))
    }
}

@available(iOS 16.1, *)
private struct CompactTrailingView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        Text(formatDurationCompact(context.state.elapsedSeconds))
            .font(.system(size: 13, weight: .bold, design: .monospaced))
            .foregroundStyle(
                LinearGradient(
                    colors: [.red, .orange],
                    startPoint: .leading,
                    endPoint: .trailing
                )
            )
            .monospacedDigit()
            .contentTransition(.numericText())
    }

    private func formatDurationCompact(_ seconds: Int) -> String {
        let hours = seconds / 3600
        let minutes = (seconds % 3600) / 60
        if hours > 0 {
            return String(format: "%d:%02d", hours, minutes)
        }
        let secs = seconds % 60
        return String(format: "%d:%02d", minutes, secs)
    }
}

// MARK: - Minimal View
@available(iOS 16.1, *)
private struct MinimalView: View {
    let context: ActivityViewContext<TrackingActivityAttributes>

    var body: some View {
        ZStack {
            // Background glow
            Circle()
                .fill(
                    RadialGradient(
                        colors: [Color.red.opacity(0.4), Color.clear],
                        center: .center,
                        startRadius: 2,
                        endRadius: 14
                    )
                )

            // Recording icon
            Image(systemName: "record.circle.fill")
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(
                    LinearGradient(
                        colors: [.red, .orange],
                        startPoint: .top,
                        endPoint: .bottom
                    )
                )
        }
    }
}
#endif
