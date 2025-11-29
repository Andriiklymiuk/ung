//
//  TrackingStatusWidget.swift
//  ungWidgets
//
//  Shows current tracking status on home screen
//

import SwiftUI
import WidgetKit

// MARK: - Timeline Entry
struct TrackingStatusEntry: TimelineEntry {
    let date: Date
    let data: WidgetData
}

// MARK: - Timeline Provider
struct TrackingStatusProvider: TimelineProvider {
    func placeholder(in context: Context) -> TrackingStatusEntry {
        TrackingStatusEntry(date: Date(), data: WidgetData())
    }

    func getSnapshot(in context: Context, completion: @escaping (TrackingStatusEntry) -> Void) {
        let entry = TrackingStatusEntry(date: Date(), data: WidgetData.load())
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<TrackingStatusEntry>) -> Void) {
        let data = WidgetData.load()
        let currentDate = Date()

        // If tracking is active, update every minute
        var entries: [TrackingStatusEntry] = []
        if data.isTracking {
            for minuteOffset in 0..<60 {
                let entryDate = Calendar.current.date(byAdding: .minute, value: minuteOffset, to: currentDate)!
                entries.append(TrackingStatusEntry(date: entryDate, data: data))
            }
        } else {
            entries.append(TrackingStatusEntry(date: currentDate, data: data))
        }

        let refreshDate = Calendar.current.date(byAdding: .minute, value: data.isTracking ? 1 : 15, to: currentDate)!
        let timeline = Timeline(entries: entries, policy: .after(refreshDate))
        completion(timeline)
    }
}

// MARK: - Widget View
struct TrackingStatusWidgetView: View {
    @Environment(\.widgetFamily) var family
    var entry: TrackingStatusEntry

    var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemMedium:
            mediumView
        default:
            smallView
        }
    }

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Header
            HStack {
                Image(systemName: entry.data.isTracking ? "record.circle.fill" : "clock.fill")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(entry.data.isTracking ? .red : .blue)
                Text(entry.data.isTracking ? "Tracking" : "Ready")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.secondary)
            }

            Spacer()

            if entry.data.isTracking {
                // Active tracking
                Text(entry.data.trackingProject)
                    .font(.system(size: 14, weight: .medium))
                    .lineLimit(2)
                    .foregroundColor(.primary)

                Text(entry.data.trackingDuration)
                    .font(.system(size: 28, weight: .bold, design: .monospaced))
                    .foregroundColor(.red)
                    .monospacedDigit()
            } else {
                // Not tracking
                Text("No active session")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)

                Text("Tap to start")
                    .font(.system(size: 11))
                    .foregroundColor(.blue)
            }
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://tracking"))
    }

    private var mediumView: some View {
        HStack(spacing: 16) {
            // Left side - tracking status
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Circle()
                        .fill(entry.data.isTracking ? Color.red : Color.gray.opacity(0.3))
                        .frame(width: 10, height: 10)
                    Text(entry.data.isTracking ? "Active Session" : "Not Tracking")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.secondary)
                }

                if entry.data.isTracking {
                    Text(entry.data.trackingProject)
                        .font(.system(size: 16, weight: .semibold))
                        .lineLimit(1)

                    if !entry.data.trackingClient.isEmpty {
                        Text(entry.data.trackingClient)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                            .lineLimit(1)
                    }

                    Spacer()

                    Text(entry.data.trackingDuration)
                        .font(.system(size: 32, weight: .bold, design: .monospaced))
                        .foregroundColor(.red)
                        .monospacedDigit()
                } else {
                    Text("Start tracking your work")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)

                    Spacer()

                    Label("Start Tracking", systemImage: "play.fill")
                        .font(.system(size: 13, weight: .medium))
                        .foregroundColor(.blue)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            // Right side - today's stats
            VStack(alignment: .trailing, spacing: 4) {
                Text("Today")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)

                Text(formatHours(entry.data.todayHours))
                    .font(.system(size: 20, weight: .bold))
                    .foregroundColor(.primary)

                Spacer()

                Text("This Week")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)

                Text(formatHours(entry.data.weeklyHours))
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.blue)
            }
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://tracking"))
    }

    private func formatHours(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        return String(format: "%dh %02dm", h, m)
    }
}

// MARK: - Widget Configuration
struct TrackingStatusWidget: Widget {
    let kind: String = "TrackingStatusWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: TrackingStatusProvider()) { entry in
            TrackingStatusWidgetView(entry: entry)
        }
        .configurationDisplayName("Time Tracking")
        .description("See your current tracking status and start sessions quickly.")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

#Preview(as: .systemSmall) {
    TrackingStatusWidget()
} timeline: {
    TrackingStatusEntry(date: .now, data: WidgetData())
    var tracking = WidgetData()
    tracking.isTracking = true
    tracking.trackingProject = "Client Work"
    tracking.trackingStartTime = Date().addingTimeInterval(-3600)
    return TrackingStatusEntry(date: .now, data: tracking)
}
