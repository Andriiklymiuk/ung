//
//  WeeklyStatsWidget.swift
//  ungWidgets
//
//  Premium weekly stats widget with Telegram-inspired design
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

        let refreshDate = Calendar.current.date(byAdding: .minute, value: 30, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(refreshDate))
        completion(timeline)
    }
}

// MARK: - Widget View
struct WeeklyStatsWidgetView: View {
    @Environment(\.widgetFamily) var family
    @Environment(\.colorScheme) var colorScheme
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

    // MARK: - Small View (Elegant Progress Ring)
    private var smallView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            VStack(spacing: 8) {
                // Premium progress ring
                ZStack {
                    // Background ring with gradient
                    Circle()
                        .stroke(
                            WidgetColors.statsGradient.opacity(0.15),
                            lineWidth: 8
                        )
                        .frame(width: 72, height: 72)

                    // Progress ring
                    Circle()
                        .trim(from: 0, to: entry.data.weeklyProgress)
                        .stroke(
                            WidgetColors.statsGradient,
                            style: StrokeStyle(lineWidth: 8, lineCap: .round)
                        )
                        .frame(width: 72, height: 72)
                        .rotationEffect(.degrees(-90))

                    // Subtle glow
                    Circle()
                        .trim(from: 0, to: entry.data.weeklyProgress)
                        .stroke(
                            WidgetColors.statsGradient,
                            style: StrokeStyle(lineWidth: 16, lineCap: .round)
                        )
                        .frame(width: 72, height: 72)
                        .rotationEffect(.degrees(-90))
                        .blur(radius: 8)
                        .opacity(0.4)

                    // Center content
                    VStack(spacing: 0) {
                        Text("\(Int(entry.data.weeklyProgress * 100))")
                            .font(.system(size: 22, weight: .bold))
                            .foregroundColor(WidgetColors.textPrimary)
                        Text("%")
                            .font(.system(size: 10, weight: .medium))
                            .foregroundColor(WidgetColors.textTertiary)
                    }
                }

                // Hours label
                VStack(spacing: 2) {
                    Text(formatHoursCompact(entry.data.weeklyHours))
                        .font(.system(size: 14, weight: .bold))
                        .foregroundColor(WidgetColors.textPrimary)

                    Text("of \(formatHoursCompact(entry.data.weeklyTarget))")
                        .font(.system(size: 10))
                        .foregroundColor(WidgetColors.textTertiary)
                }
            }
            .padding(14)
        }
        .widgetURL(URL(string: "ung://reports"))
    }

    // MARK: - Medium View (Stats Dashboard)
    private var mediumView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            HStack(spacing: 16) {
                // Left - Progress ring
                ZStack {
                    // Background ring
                    Circle()
                        .stroke(
                            WidgetColors.statsGradient.opacity(0.15),
                            lineWidth: 10
                        )
                        .frame(width: 90, height: 90)

                    // Progress ring
                    Circle()
                        .trim(from: 0, to: entry.data.weeklyProgress)
                        .stroke(
                            WidgetColors.statsGradient,
                            style: StrokeStyle(lineWidth: 10, lineCap: .round)
                        )
                        .frame(width: 90, height: 90)
                        .rotationEffect(.degrees(-90))

                    // Glow
                    Circle()
                        .trim(from: 0, to: entry.data.weeklyProgress)
                        .stroke(
                            WidgetColors.statsGradient,
                            style: StrokeStyle(lineWidth: 20, lineCap: .round)
                        )
                        .frame(width: 90, height: 90)
                        .rotationEffect(.degrees(-90))
                        .blur(radius: 10)
                        .opacity(0.4)

                    // Center
                    VStack(spacing: 0) {
                        Text("\(Int(entry.data.weeklyProgress * 100))")
                            .font(.system(size: 26, weight: .bold))
                            .foregroundColor(WidgetColors.textPrimary)
                        Text("percent")
                            .font(.system(size: 9, weight: .medium))
                            .foregroundColor(WidgetColors.textTertiary)
                    }
                }

                // Right - Stats grid
                VStack(alignment: .leading, spacing: 10) {
                    // Week hours
                    statRow(
                        icon: "calendar.badge.clock",
                        title: "This Week",
                        value: formatHours(entry.data.weeklyHours),
                        gradient: WidgetColors.statsGradient
                    )

                    // Today hours
                    statRow(
                        icon: "sun.max.fill",
                        title: "Today",
                        value: formatHours(entry.data.todayHours),
                        gradient: WidgetColors.focusGradient
                    )

                    // Pending amount
                    statRow(
                        icon: "banknote.fill",
                        title: "Pending",
                        value: entry.data.formattedPendingAmount,
                        gradient: WidgetColors.invoiceGradient
                    )
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .padding(16)
        }
        .widgetURL(URL(string: "ung://reports"))
    }

    // MARK: - Large View (Full Dashboard)
    private var largeView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            VStack(spacing: 16) {
                // Header
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Weekly Progress")
                            .font(.system(size: 17, weight: .semibold))
                            .foregroundColor(WidgetColors.textPrimary)
                        Text(weekDateRange())
                            .font(.system(size: 12))
                            .foregroundColor(WidgetColors.textTertiary)
                    }

                    Spacer()

                    // Achievement badge
                    if entry.data.weeklyProgress >= 1.0 {
                        HStack(spacing: 4) {
                            Image(systemName: "checkmark.seal.fill")
                                .font(.system(size: 12))
                            Text("Goal Met!")
                                .font(.system(size: 11, weight: .semibold))
                        }
                        .foregroundColor(.white)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 5)
                        .background(WidgetColors.breakGradient)
                        .clipShape(Capsule())
                    }
                }

                // Progress section
                VStack(spacing: 12) {
                    // Hours row
                    HStack(alignment: .firstTextBaseline) {
                        Text(formatHours(entry.data.weeklyHours))
                            .font(.system(size: 36, weight: .bold))
                            .foregroundColor(WidgetColors.textPrimary)

                        Text("of \(formatHours(entry.data.weeklyTarget))")
                            .font(.system(size: 14))
                            .foregroundColor(WidgetColors.textTertiary)

                        Spacer()

                        Text("\(Int(entry.data.weeklyProgress * 100))%")
                            .font(.system(size: 18, weight: .bold))
                            .foregroundStyle(WidgetColors.statsGradient)
                    }

                    // Progress bar
                    GeometryReader { geometry in
                        ZStack(alignment: .leading) {
                            // Background
                            RoundedRectangle(cornerRadius: 6)
                                .fill(WidgetColors.statsGradient.opacity(0.15))
                                .frame(height: 12)

                            // Progress
                            RoundedRectangle(cornerRadius: 6)
                                .fill(WidgetColors.statsGradient)
                                .frame(width: geometry.size.width * min(entry.data.weeklyProgress, 1.0), height: 12)

                            // Glow
                            RoundedRectangle(cornerRadius: 6)
                                .fill(WidgetColors.statsGradient)
                                .frame(width: geometry.size.width * min(entry.data.weeklyProgress, 1.0), height: 12)
                                .blur(radius: 6)
                                .opacity(0.5)
                        }
                    }
                    .frame(height: 12)
                }

                Divider()
                    .background(Color.gray.opacity(0.2))

                // Stats grid
                LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                    largeStatCard(
                        icon: "sun.max.fill",
                        title: "Today",
                        value: formatHours(entry.data.todayHours),
                        gradient: WidgetColors.focusGradient
                    )

                    largeStatCard(
                        icon: "doc.text.fill",
                        title: "Pending Invoices",
                        value: "\(entry.data.pendingInvoices)",
                        gradient: WidgetColors.trackingGradient
                    )

                    largeStatCard(
                        icon: "banknote.fill",
                        title: "Pending Amount",
                        value: entry.data.formattedPendingAmount,
                        gradient: WidgetColors.invoiceGradient
                    )

                    largeStatCard(
                        icon: "target",
                        title: "Weekly Target",
                        value: formatHours(entry.data.weeklyTarget),
                        gradient: WidgetColors.statsGradient
                    )
                }

                Spacer()
            }
            .padding(16)
        }
        .widgetURL(URL(string: "ung://reports"))
    }

    // MARK: - Helper Views
    private func statRow(icon: String, title: String, value: String, gradient: LinearGradient) -> some View {
        HStack(spacing: 10) {
            // Icon badge
            ZStack {
                Circle()
                    .fill(gradient.opacity(0.15))
                    .frame(width: 32, height: 32)

                Image(systemName: icon)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(gradient)
            }

            VStack(alignment: .leading, spacing: 1) {
                Text(title)
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(WidgetColors.textTertiary)

                Text(value)
                    .font(.system(size: 15, weight: .bold))
                    .foregroundColor(WidgetColors.textPrimary)
            }
        }
    }

    private func largeStatCard(icon: String, title: String, value: String, gradient: LinearGradient) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            ZStack {
                Circle()
                    .fill(gradient.opacity(0.15))
                    .frame(width: 36, height: 36)

                Image(systemName: icon)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(gradient)
            }

            Text(title)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(WidgetColors.textTertiary)
                .lineLimit(1)

            Text(value)
                .font(.system(size: 18, weight: .bold))
                .foregroundColor(WidgetColors.textPrimary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark
                    ? Color.white.opacity(0.05)
                    : Color.black.opacity(0.03)
                )
        )
    }

    private func formatHours(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        return String(format: "%dh %dm", h, m)
    }

    private func formatHoursCompact(_ hours: Double) -> String {
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        if h > 0 {
            return String(format: "%dh", h)
        }
        return String(format: "%dm", m)
    }

    private func weekDateRange() -> String {
        let calendar = Calendar.current
        let now = Date()
        let weekday = calendar.component(.weekday, from: now)
        let daysToMonday = (weekday + 5) % 7

        guard let monday = calendar.date(byAdding: .day, value: -daysToMonday, to: now),
              let sunday = calendar.date(byAdding: .day, value: 6, to: monday) else {
            return ""
        }

        let formatter = DateFormatter()
        formatter.dateFormat = "MMM d"
        return "\(formatter.string(from: monday)) - \(formatter.string(from: sunday))"
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
        .description("Track your weekly progress with beautiful insights.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
        .contentMarginsDisabled()
    }
}

#Preview(as: .systemLarge) {
    WeeklyStatsWidget()
} timeline: {
    var data = WidgetData()
    data.weeklyHours = 32.5
    data.weeklyTarget = 40
    data.todayHours = 6.25
    data.pendingInvoices = 3
    data.pendingAmount = 4850.00
    return WeeklyStatsEntry(date: .now, data: data)
}
