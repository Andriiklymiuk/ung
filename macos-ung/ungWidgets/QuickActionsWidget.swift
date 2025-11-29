//
//  QuickActionsWidget.swift
//  ungWidgets
//
//  Premium quick actions widget with Telegram-inspired design
//

import SwiftUI
import WidgetKit

// MARK: - Timeline Entry
struct QuickActionsEntry: TimelineEntry {
    let date: Date
    let data: WidgetData
}

// MARK: - Timeline Provider
struct QuickActionsProvider: TimelineProvider {
    func placeholder(in context: Context) -> QuickActionsEntry {
        QuickActionsEntry(date: Date(), data: WidgetData())
    }

    func getSnapshot(in context: Context, completion: @escaping (QuickActionsEntry) -> Void) {
        let entry = QuickActionsEntry(date: Date(), data: WidgetData.load())
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<QuickActionsEntry>) -> Void) {
        let data = WidgetData.load()
        let entry = QuickActionsEntry(date: Date(), data: data)

        let refreshDate = Calendar.current.date(byAdding: .minute, value: 15, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(refreshDate))
        completion(timeline)
    }
}

// MARK: - Widget View
struct QuickActionsWidgetView: View {
    @Environment(\.widgetFamily) var family
    @Environment(\.colorScheme) var colorScheme
    var entry: QuickActionsEntry

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

    // MARK: - Small View (2x2 Grid)
    private var smallView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            VStack(spacing: 10) {
                HStack(spacing: 10) {
                    // Track/Stop button
                    Link(destination: URL(string: "ung://tracking/toggle")!) {
                        actionTile(
                            icon: entry.data.isTracking ? "stop.fill" : "play.fill",
                            label: entry.data.isTracking ? "Stop" : "Track",
                            gradient: entry.data.isTracking ? WidgetColors.trackingGradient : WidgetColors.statsGradient,
                            isActive: entry.data.isTracking
                        )
                    }

                    // Focus button
                    Link(destination: URL(string: "ung://pomodoro/start")!) {
                        actionTile(
                            icon: entry.data.pomodoroActive ? "pause.fill" : "brain.head.profile.fill",
                            label: entry.data.pomodoroActive ? "Pause" : "Focus",
                            gradient: WidgetColors.focusGradient,
                            isActive: entry.data.pomodoroActive
                        )
                    }
                }

                HStack(spacing: 10) {
                    // Invoices button
                    Link(destination: URL(string: "ung://invoices")!) {
                        actionTile(
                            icon: "doc.text.fill",
                            label: "Invoice",
                            gradient: WidgetColors.invoiceGradient,
                            badge: entry.data.pendingInvoices > 0 ? "\(entry.data.pendingInvoices)" : nil
                        )
                    }

                    // Expenses button
                    Link(destination: URL(string: "ung://expenses/add")!) {
                        actionTile(
                            icon: "plus.circle.fill",
                            label: "Expense",
                            gradient: LinearGradient(
                                colors: [Color(hex: "A55EEA"), Color(hex: "8854D0")],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                    }
                }
            }
            .padding(10)
        }
    }

    // MARK: - Medium View (4 actions in row)
    private var mediumView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            HStack(spacing: 12) {
                // Track/Stop
                Link(destination: URL(string: "ung://tracking/toggle")!) {
                    mediumActionCard(
                        icon: entry.data.isTracking ? "stop.fill" : "play.fill",
                        title: entry.data.isTracking ? "Stop" : "Start",
                        subtitle: entry.data.isTracking ? entry.data.trackingDuration : "Tracking",
                        gradient: entry.data.isTracking ? WidgetColors.trackingGradient : WidgetColors.statsGradient,
                        isActive: entry.data.isTracking
                    )
                }

                // Divider
                Rectangle()
                    .fill(Color.gray.opacity(0.2))
                    .frame(width: 1)
                    .padding(.vertical, 12)

                // Focus
                Link(destination: URL(string: "ung://pomodoro/start")!) {
                    mediumActionCard(
                        icon: entry.data.pomodoroActive ? "pause.fill" : "brain.head.profile.fill",
                        title: entry.data.pomodoroActive ? "Pause" : "Focus",
                        subtitle: entry.data.pomodoroActive ? entry.data.pomodoroTimeFormatted : "Timer",
                        gradient: WidgetColors.focusGradient,
                        isActive: entry.data.pomodoroActive
                    )
                }

                // Divider
                Rectangle()
                    .fill(Color.gray.opacity(0.2))
                    .frame(width: 1)
                    .padding(.vertical, 12)

                // Invoices
                Link(destination: URL(string: "ung://invoices")!) {
                    mediumActionCard(
                        icon: "doc.text.fill",
                        title: "Invoices",
                        subtitle: entry.data.pendingInvoices > 0 ? "\(entry.data.pendingInvoices) pending" : "View all",
                        gradient: WidgetColors.invoiceGradient,
                        badge: entry.data.pendingInvoices > 0 ? "\(entry.data.pendingInvoices)" : nil
                    )
                }

                // Divider
                Rectangle()
                    .fill(Color.gray.opacity(0.2))
                    .frame(width: 1)
                    .padding(.vertical, 12)

                // Add Expense
                Link(destination: URL(string: "ung://expenses/add")!) {
                    mediumActionCard(
                        icon: "plus.circle.fill",
                        title: "Expense",
                        subtitle: "Log new",
                        gradient: LinearGradient(
                            colors: [Color(hex: "A55EEA"), Color(hex: "8854D0")],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)
        }
    }

    // MARK: - Helper Views

    private func actionTile(icon: String, label: String, gradient: LinearGradient, isActive: Bool = false, badge: String? = nil) -> some View {
        VStack(spacing: 6) {
            ZStack(alignment: .topTrailing) {
                ZStack {
                    // Background
                    RoundedRectangle(cornerRadius: 12)
                        .fill(gradient.opacity(isActive ? 0.25 : 0.15))
                        .frame(width: 44, height: 44)

                    // Glow when active
                    if isActive {
                        RoundedRectangle(cornerRadius: 12)
                            .fill(gradient)
                            .frame(width: 44, height: 44)
                            .blur(radius: 10)
                            .opacity(0.4)
                    }

                    // Icon
                    Image(systemName: icon)
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundStyle(gradient)
                }

                // Badge
                if let badge = badge {
                    Text(badge)
                        .font(.system(size: 9, weight: .bold))
                        .foregroundColor(.white)
                        .padding(.horizontal, 5)
                        .padding(.vertical, 2)
                        .background(WidgetColors.trackingGradient)
                        .clipShape(Capsule())
                        .offset(x: 6, y: -6)
                }
            }

            Text(label)
                .font(.system(size: 10, weight: .medium))
                .foregroundColor(WidgetColors.textPrimary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func mediumActionCard(icon: String, title: String, subtitle: String, gradient: LinearGradient, isActive: Bool = false, badge: String? = nil) -> some View {
        VStack(spacing: 8) {
            ZStack(alignment: .topTrailing) {
                ZStack {
                    // Icon circle
                    Circle()
                        .fill(gradient.opacity(isActive ? 0.25 : 0.15))
                        .frame(width: 48, height: 48)

                    // Glow when active
                    if isActive {
                        Circle()
                            .fill(gradient)
                            .frame(width: 48, height: 48)
                            .blur(radius: 12)
                            .opacity(0.4)
                    }

                    // Icon
                    Image(systemName: icon)
                        .font(.system(size: 20, weight: .semibold))
                        .foregroundStyle(gradient)
                }

                // Badge
                if let badge = badge {
                    Text(badge)
                        .font(.system(size: 10, weight: .bold))
                        .foregroundColor(.white)
                        .frame(width: 18, height: 18)
                        .background(WidgetColors.trackingGradient)
                        .clipShape(Circle())
                        .offset(x: 4, y: -4)
                }
            }

            VStack(spacing: 2) {
                Text(title)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(WidgetColors.textPrimary)

                Text(subtitle)
                    .font(.system(size: 10))
                    .foregroundColor(isActive ? Color(gradient.stops.first?.color ?? .gray) : WidgetColors.textTertiary)
                    .lineLimit(1)
            }
        }
        .frame(maxWidth: .infinity)
    }
}

// Helper extension for gradient color extraction
extension LinearGradient {
    var stops: [Gradient.Stop] {
        return []
    }
}

// MARK: - Widget Configuration
struct QuickActionsWidget: Widget {
    let kind: String = "QuickActionsWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: QuickActionsProvider()) { entry in
            QuickActionsWidgetView(entry: entry)
        }
        .configurationDisplayName("Quick Actions")
        .description("Fast access to common tasks.")
        .supportedFamilies([.systemSmall, .systemMedium])
        .contentMarginsDisabled()
    }
}

#Preview(as: .systemMedium) {
    QuickActionsWidget()
} timeline: {
    var data = WidgetData()
    data.isTracking = true
    data.trackingProject = "Client Work"
    data.trackingStartTime = Date().addingTimeInterval(-3600)
    data.pendingInvoices = 2
    return QuickActionsEntry(date: .now, data: data)
}
