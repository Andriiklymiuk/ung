//
//  TrackingStatusWidget.swift
//  ungWidgets
//
//  Premium tracking status widget with Telegram-inspired design
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
    @Environment(\.colorScheme) var colorScheme
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

    // MARK: - Small Widget (Premium Design)
    private var smallView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            VStack(alignment: .leading, spacing: 0) {
                // Status header
                HStack(spacing: 6) {
                    if entry.data.isTracking {
                        PulsingDot(color: WidgetColors.trackingRed, size: 6)
                    } else {
                        Circle()
                            .fill(Color.gray.opacity(0.3))
                            .frame(width: 6, height: 6)
                    }

                    Text(entry.data.isTracking ? "RECORDING" : "READY")
                        .font(.system(size: 9, weight: .bold))
                        .tracking(1.2)
                        .foregroundColor(entry.data.isTracking ? WidgetColors.trackingRed : WidgetColors.textTertiary)
                }
                .padding(.bottom, 8)

                Spacer()

                if entry.data.isTracking {
                    // Active tracking view
                    VStack(alignment: .leading, spacing: 4) {
                        Text(entry.data.trackingProject)
                            .font(.system(size: 14, weight: .semibold))
                            .lineLimit(2)
                            .foregroundColor(WidgetColors.textPrimary)

                        if !entry.data.trackingClient.isEmpty {
                            Text(entry.data.trackingClient)
                                .font(.system(size: 11))
                                .foregroundColor(WidgetColors.textTertiary)
                                .lineLimit(1)
                        }
                    }

                    Spacer()

                    // Timer
                    MonospacedTimer(
                        time: entry.data.trackingDuration,
                        size: 26,
                        color: WidgetColors.trackingRed
                    )
                } else {
                    // Idle state
                    VStack(alignment: .leading, spacing: 6) {
                        Image(systemName: "play.circle.fill")
                            .font(.system(size: 32))
                            .foregroundStyle(WidgetColors.statsGradient)

                        Text("Start tracking")
                            .font(.system(size: 13, weight: .medium))
                            .foregroundColor(WidgetColors.textSecondary)
                    }

                    Spacer()

                    // Today's hours hint
                    HStack(spacing: 4) {
                        Text("Today")
                            .font(.system(size: 10))
                            .foregroundColor(WidgetColors.textTertiary)
                        Text(formatHours(entry.data.todayHours))
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(WidgetColors.textSecondary)
                    }
                }
            }
            .padding(14)
        }
        .widgetURL(URL(string: "ung://tracking"))
    }

    // MARK: - Medium Widget (Premium Design)
    private var mediumView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            HStack(spacing: 16) {
                // Left section - Main content
                VStack(alignment: .leading, spacing: 0) {
                    // Status header
                    HStack(spacing: 6) {
                        if entry.data.isTracking {
                            PulsingDot(color: WidgetColors.trackingRed, size: 6)
                        } else {
                            Circle()
                                .fill(Color.gray.opacity(0.3))
                                .frame(width: 6, height: 6)
                        }

                        Text(entry.data.isTracking ? "RECORDING" : "READY")
                            .font(.system(size: 9, weight: .bold))
                            .tracking(1.2)
                            .foregroundColor(entry.data.isTracking ? WidgetColors.trackingRed : WidgetColors.textTertiary)
                    }

                    Spacer()

                    if entry.data.isTracking {
                        // Project info
                        Text(entry.data.trackingProject)
                            .font(.system(size: 16, weight: .semibold))
                            .lineLimit(1)
                            .foregroundColor(WidgetColors.textPrimary)

                        if !entry.data.trackingClient.isEmpty {
                            Text(entry.data.trackingClient)
                                .font(.system(size: 12))
                                .foregroundColor(WidgetColors.textTertiary)
                                .lineLimit(1)
                        }

                        Spacer()

                        // Timer
                        MonospacedTimer(
                            time: entry.data.trackingDuration,
                            size: 32,
                            color: WidgetColors.trackingRed
                        )
                    } else {
                        // Idle state
                        HStack(spacing: 12) {
                            Image(systemName: "play.circle.fill")
                                .font(.system(size: 40))
                                .foregroundStyle(WidgetColors.statsGradient)

                            VStack(alignment: .leading, spacing: 2) {
                                Text("Start tracking")
                                    .font(.system(size: 15, weight: .semibold))
                                    .foregroundColor(WidgetColors.textPrimary)
                                Text("Tap to begin a session")
                                    .font(.system(size: 12))
                                    .foregroundColor(WidgetColors.textTertiary)
                            }
                        }

                        Spacer()
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)

                // Right section - Stats cards
                VStack(spacing: 8) {
                    // Today card
                    statsCard(
                        icon: "sun.max.fill",
                        title: "Today",
                        value: formatHours(entry.data.todayHours),
                        gradient: WidgetColors.focusGradient
                    )

                    // Week card
                    statsCard(
                        icon: "calendar",
                        title: "Week",
                        value: formatHours(entry.data.weeklyHours),
                        gradient: WidgetColors.statsGradient
                    )
                }
                .frame(width: 100)
            }
            .padding(14)
        }
        .widgetURL(URL(string: "ung://tracking"))
    }

    // MARK: - Helper Views
    private func statsCard(icon: String, title: String, value: String, gradient: LinearGradient) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 4) {
                Image(systemName: icon)
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundStyle(gradient)
                Text(title)
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(WidgetColors.textTertiary)
            }

            Text(value)
                .font(.system(size: 15, weight: .bold))
                .foregroundColor(WidgetColors.textPrimary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(colorScheme == .dark
                    ? Color.white.opacity(0.05)
                    : Color.black.opacity(0.03)
                )
        )
    }

    private func formatHours(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        if h > 0 {
            return String(format: "%dh %dm", h, m)
        }
        return String(format: "%dm", m)
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
        .description("Track your work sessions in real-time.")
        .supportedFamilies([.systemSmall, .systemMedium])
        .contentMarginsDisabled()
    }
}

#Preview(as: .systemSmall) {
    TrackingStatusWidget()
} timeline: {
    TrackingStatusEntry(date: .now, data: WidgetData())
    var tracking = WidgetData()
    tracking.isTracking = true
    tracking.trackingProject = "Mobile App Design"
    tracking.trackingClient = "Acme Corp"
    tracking.trackingStartTime = Date().addingTimeInterval(-5423)
    return TrackingStatusEntry(date: .now, data: tracking)
}
