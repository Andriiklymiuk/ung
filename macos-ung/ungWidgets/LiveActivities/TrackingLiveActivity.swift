//
//  TrackingLiveActivity.swift
//  ungWidgets
//
//  Live Activity for real-time tracking display on Lock Screen & Dynamic Island
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
    }

    var projectName: String
    var clientName: String
    var startTime: Date
}

// MARK: - Live Activity Widget
@available(iOS 16.1, *)
struct TrackingLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: TrackingActivityAttributes.self) { context in
            // Lock Screen / Banner view
            lockScreenView(context: context)
        } dynamicIsland: { context in
            DynamicIsland {
                // Expanded view
                DynamicIslandExpandedRegion(.leading) {
                    HStack(spacing: 8) {
                        Circle()
                            .fill(Color.red)
                            .frame(width: 10, height: 10)
                        Text(context.attributes.projectName)
                            .font(.system(size: 14, weight: .semibold))
                            .lineLimit(1)
                    }
                }

                DynamicIslandExpandedRegion(.trailing) {
                    Text(formatDuration(context.state.elapsedSeconds))
                        .font(.system(size: 20, weight: .bold, design: .monospaced))
                        .foregroundColor(.red)
                        .monospacedDigit()
                }

                DynamicIslandExpandedRegion(.center) {
                    if !context.attributes.clientName.isEmpty {
                        Text(context.attributes.clientName)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                    }
                }

                DynamicIslandExpandedRegion(.bottom) {
                    HStack {
                        Label("Started", systemImage: "clock")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)

                        Spacer()

                        Text(context.attributes.startTime, style: .time)
                            .font(.system(size: 11, weight: .medium))
                    }
                    .padding(.horizontal, 4)
                }
            } compactLeading: {
                // Compact leading (pill left side)
                HStack(spacing: 4) {
                    Circle()
                        .fill(Color.red)
                        .frame(width: 8, height: 8)
                    Text(String(context.attributes.projectName.prefix(3)))
                        .font(.system(size: 12, weight: .semibold))
                }
            } compactTrailing: {
                // Compact trailing (pill right side)
                Text(formatDurationCompact(context.state.elapsedSeconds))
                    .font(.system(size: 12, weight: .bold, design: .monospaced))
                    .foregroundColor(.red)
                    .monospacedDigit()
            } minimal: {
                // Minimal view (when multiple activities)
                ZStack {
                    Circle()
                        .fill(Color.red.opacity(0.3))
                    Image(systemName: "record.circle.fill")
                        .font(.system(size: 12))
                        .foregroundColor(.red)
                }
            }
        }
    }

    @ViewBuilder
    private func lockScreenView(context: ActivityViewContext<TrackingActivityAttributes>) -> some View {
        HStack(spacing: 16) {
            // Status indicator
            ZStack {
                Circle()
                    .fill(Color.red.opacity(0.2))
                    .frame(width: 50, height: 50)

                Circle()
                    .fill(Color.red)
                    .frame(width: 12, height: 12)
            }

            // Project info
            VStack(alignment: .leading, spacing: 4) {
                Text(context.attributes.projectName)
                    .font(.system(size: 16, weight: .semibold))
                    .lineLimit(1)

                if !context.attributes.clientName.isEmpty {
                    Text(context.attributes.clientName)
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            // Timer
            VStack(alignment: .trailing, spacing: 4) {
                Text(formatDuration(context.state.elapsedSeconds))
                    .font(.system(size: 24, weight: .bold, design: .monospaced))
                    .foregroundColor(.red)
                    .monospacedDigit()

                Text("tracking")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
        }
        .padding()
        .activityBackgroundTint(Color.black.opacity(0.8))
        .activitySystemActionForegroundColor(Color.white)
    }

    private func formatDuration(_ seconds: Int) -> String {
        let hours = seconds / 3600
        let minutes = (seconds % 3600) / 60
        let secs = seconds % 60
        return String(format: "%d:%02d:%02d", hours, minutes, secs)
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
#endif
