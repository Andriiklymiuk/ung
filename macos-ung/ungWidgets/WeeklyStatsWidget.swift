//
//  WeeklyStatsWidget.swift
//  ungWidgets
//
//  Shows weekly progress and stats
//

import SwiftUI
import WidgetKit

// MARK: - Timeline Entry
struct WeeklyStatsEntry: TimelineEntry {
    let date: Date
    let data: WidgetData
}

// MARK: - Timeline Provider
struct WeeklyStatsProvider: TimelineProvider {
    func placeholder(in context: Context) -> WeeklyStatsEntry {
        WeeklyStatsEntry(date: Date(), data: WidgetData())
    }

    func getSnapshot(in context: Context, completion: @escaping (WeeklyStatsEntry) -> Void) {
        let entry = WeeklyStatsEntry(date: Date(), data: WidgetData.load())
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<WeeklyStatsEntry>) -> Void) {
        let data = WidgetData.load()
        let entry = WeeklyStatsEntry(date: Date(), data: data)

        // Refresh every 30 minutes
        let refreshDate = Calendar.current.date(byAdding: .minute, value: 30, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(refreshDate))
        completion(timeline)
    }
}

// MARK: - Widget View
struct WeeklyStatsWidgetView: View {
    @Environment(\.widgetFamily) var family
    var entry: WeeklyStatsEntry

    var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemMedium:
            mediumView
        case .systemLarge:
            largeView
        default:
            smallView
        }
    }

    // MARK: - Small View (Progress Ring)
    private var smallView: some View {
        VStack(spacing: 8) {
            // Progress ring
            ZStack {
                Circle()
                    .stroke(Color.blue.opacity(0.2), lineWidth: 8)
                    .frame(width: 70, height: 70)

                Circle()
                    .trim(from: 0, to: entry.data.weeklyProgress)
                    .stroke(
                        LinearGradient(
                            colors: [.blue, .cyan],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        style: StrokeStyle(lineWidth: 8, lineCap: .round)
                    )
                    .frame(width: 70, height: 70)
                    .rotationEffect(.degrees(-90))

                VStack(spacing: 0) {
                    Text("\(Int(entry.data.weeklyProgress * 100))%")
                        .font(.system(size: 16, weight: .bold))
                    Text("goal")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                }
            }

            Text(formatHours(entry.data.weeklyHours))
                .font(.system(size: 14, weight: .semibold))

            Text("of \(formatHours(entry.data.weeklyTarget))")
                .font(.system(size: 10))
                .foregroundColor(.secondary)
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://reports"))
    }

    // MARK: - Medium View (Stats Row)
    private var mediumView: some View {
        HStack(spacing: 16) {
            // Progress ring
            ZStack {
                Circle()
                    .stroke(Color.blue.opacity(0.2), lineWidth: 10)
                    .frame(width: 80, height: 80)

                Circle()
                    .trim(from: 0, to: entry.data.weeklyProgress)
                    .stroke(
                        LinearGradient(
                            colors: [.blue, .cyan],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        style: StrokeStyle(lineWidth: 10, lineCap: .round)
                    )
                    .frame(width: 80, height: 80)
                    .rotationEffect(.degrees(-90))

                VStack(spacing: 2) {
                    Text("\(Int(entry.data.weeklyProgress * 100))%")
                        .font(.system(size: 18, weight: .bold))
                    Text("Weekly")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                }
            }

            // Stats
            VStack(alignment: .leading, spacing: 12) {
                statRow(
                    icon: "clock.fill",
                    title: "This Week",
                    value: formatHours(entry.data.weeklyHours),
                    color: .blue
                )

                statRow(
                    icon: "sun.max.fill",
                    title: "Today",
                    value: formatHours(entry.data.todayHours),
                    color: .orange
                )

                statRow(
                    icon: "doc.plaintext.fill",
                    title: "Pending",
                    value: entry.data.formattedPendingAmount,
                    color: .green
                )
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://reports"))
    }

    // MARK: - Large View (Full Stats)
    private var largeView: some View {
        VStack(spacing: 16) {
            // Header
            HStack {
                Text("Weekly Progress")
                    .font(.system(size: 16, weight: .semibold))
                Spacer()
                Image(systemName: "chart.bar.fill")
                    .foregroundColor(.blue)
            }

            // Progress bar
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text(formatHours(entry.data.weeklyHours))
                        .font(.system(size: 28, weight: .bold))
                    Text("of \(formatHours(entry.data.weeklyTarget))")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)
                    Spacer()
                    Text("\(Int(entry.data.weeklyProgress * 100))%")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundColor(.blue)
                }

                GeometryReader { geometry in
                    ZStack(alignment: .leading) {
                        RoundedRectangle(cornerRadius: 6)
                            .fill(Color.blue.opacity(0.2))
                            .frame(height: 12)

                        RoundedRectangle(cornerRadius: 6)
                            .fill(
                                LinearGradient(
                                    colors: [.blue, .cyan],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                            )
                            .frame(width: geometry.size.width * entry.data.weeklyProgress, height: 12)
                    }
                }
                .frame(height: 12)
            }

            Divider()

            // Stats grid
            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 16) {
                largeStatCard(
                    icon: "sun.max.fill",
                    title: "Today",
                    value: formatHours(entry.data.todayHours),
                    color: .orange
                )

                largeStatCard(
                    icon: "doc.plaintext.fill",
                    title: "Pending Invoices",
                    value: "\(entry.data.pendingInvoices)",
                    color: .purple
                )

                largeStatCard(
                    icon: "dollarsign.circle.fill",
                    title: "Pending Amount",
                    value: entry.data.formattedPendingAmount,
                    color: .green
                )

                largeStatCard(
                    icon: "target",
                    title: "Weekly Target",
                    value: formatHours(entry.data.weeklyTarget),
                    color: .blue
                )
            }

            Spacer()
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://reports"))
    }

    // MARK: - Helper Views
    private func statRow(icon: String, title: String, value: String, color: Color) -> some View {
        HStack(spacing: 8) {
            Image(systemName: icon)
                .font(.system(size: 12))
                .foregroundColor(color)
                .frame(width: 20)

            VStack(alignment: .leading, spacing: 1) {
                Text(title)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
                Text(value)
                    .font(.system(size: 14, weight: .semibold))
            }
        }
    }

    private func largeStatCard(icon: String, title: String, value: String, color: Color) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Image(systemName: icon)
                    .font(.system(size: 14))
                    .foregroundColor(color)
                Spacer()
            }

            Text(title)
                .font(.system(size: 11))
                .foregroundColor(.secondary)

            Text(value)
                .font(.system(size: 18, weight: .bold))
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(color.opacity(0.1))
        )
    }

    private func formatHours(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        return String(format: "%dh %02dm", h, m)
    }
}

// MARK: - Widget Configuration
struct WeeklyStatsWidget: Widget {
    let kind: String = "WeeklyStatsWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: WeeklyStatsProvider()) { entry in
            WeeklyStatsWidgetView(entry: entry)
        }
        .configurationDisplayName("Weekly Stats")
        .description("Track your weekly progress and goals at a glance.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}

#Preview(as: .systemMedium) {
    WeeklyStatsWidget()
} timeline: {
    var data = WidgetData()
    data.weeklyHours = 28.5
    data.weeklyTarget = 40
    data.todayHours = 4.25
    data.pendingInvoices = 3
    data.pendingAmount = 2450.00
    return WeeklyStatsEntry(date: .now, data: data)
}
