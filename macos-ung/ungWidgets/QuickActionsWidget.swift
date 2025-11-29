//
//  QuickActionsWidget.swift
//  ungWidgets
//
//  Quick action buttons for common tasks
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

        // Refresh every 15 minutes
        let refreshDate = Calendar.current.date(byAdding: .minute, value: 15, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(refreshDate))
        completion(timeline)
    }
}

// MARK: - Widget View
struct QuickActionsWidgetView: View {
    @Environment(\.widgetFamily) var family
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
        VStack(spacing: 8) {
            HStack(spacing: 8) {
                actionButton(
                    icon: entry.data.isTracking ? "stop.fill" : "play.fill",
                    label: entry.data.isTracking ? "Stop" : "Track",
                    color: entry.data.isTracking ? .red : .blue,
                    url: "ung://tracking/toggle"
                )

                actionButton(
                    icon: "brain.head.profile.fill",
                    label: "Focus",
                    color: .orange,
                    url: "ung://pomodoro/start"
                )
            }

            HStack(spacing: 8) {
                actionButton(
                    icon: "doc.plaintext.fill",
                    label: "Invoice",
                    color: .green,
                    url: "ung://invoices"
                )

                actionButton(
                    icon: "chart.bar.fill",
                    label: "Reports",
                    color: .purple,
                    url: "ung://reports"
                )
            }
        }
        .padding(8)
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
    }

    // MARK: - Medium View (4 buttons in row)
    private var mediumView: some View {
        HStack(spacing: 12) {
            Link(destination: URL(string: "ung://tracking/toggle")!) {
                VStack(spacing: 8) {
                    ZStack {
                        Circle()
                            .fill(entry.data.isTracking ? Color.red.opacity(0.15) : Color.blue.opacity(0.15))
                            .frame(width: 50, height: 50)

                        Image(systemName: entry.data.isTracking ? "stop.fill" : "play.fill")
                            .font(.system(size: 20))
                            .foregroundColor(entry.data.isTracking ? .red : .blue)
                    }

                    Text(entry.data.isTracking ? "Stop" : "Start")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.primary)

                    if entry.data.isTracking {
                        Text(entry.data.trackingDuration)
                            .font(.system(size: 9, design: .monospaced))
                            .foregroundColor(.red)
                    }
                }
            }

            Divider()

            Link(destination: URL(string: "ung://pomodoro/start")!) {
                VStack(spacing: 8) {
                    ZStack {
                        Circle()
                            .fill(Color.orange.opacity(0.15))
                            .frame(width: 50, height: 50)

                        Image(systemName: entry.data.pomodoroActive ? "pause.fill" : "brain.head.profile.fill")
                            .font(.system(size: 20))
                            .foregroundColor(.orange)
                    }

                    Text(entry.data.pomodoroActive ? "Pause" : "Focus")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.primary)

                    if entry.data.pomodoroActive {
                        Text(entry.data.pomodoroTimeFormatted)
                            .font(.system(size: 9, design: .monospaced))
                            .foregroundColor(.orange)
                    }
                }
            }

            Divider()

            Link(destination: URL(string: "ung://invoices")!) {
                VStack(spacing: 8) {
                    ZStack {
                        Circle()
                            .fill(Color.green.opacity(0.15))
                            .frame(width: 50, height: 50)

                        Image(systemName: "doc.plaintext.fill")
                            .font(.system(size: 20))
                            .foregroundColor(.green)
                    }

                    Text("Invoices")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.primary)

                    if entry.data.pendingInvoices > 0 {
                        Text("\(entry.data.pendingInvoices) pending")
                            .font(.system(size: 9))
                            .foregroundColor(.green)
                    }
                }
            }

            Divider()

            Link(destination: URL(string: "ung://expenses/add")!) {
                VStack(spacing: 8) {
                    ZStack {
                        Circle()
                            .fill(Color.purple.opacity(0.15))
                            .frame(width: 50, height: 50)

                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 20))
                            .foregroundColor(.purple)
                    }

                    Text("Expense")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.primary)

                    Text("Log new")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                }
            }
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
    }

    // MARK: - Helper
    private func actionButton(icon: String, label: String, color: Color, url: String) -> some View {
        Link(destination: URL(string: url)!) {
            VStack(spacing: 4) {
                ZStack {
                    RoundedRectangle(cornerRadius: 12)
                        .fill(color.opacity(0.15))
                        .frame(width: 44, height: 44)

                    Image(systemName: icon)
                        .font(.system(size: 18))
                        .foregroundColor(color)
                }

                Text(label)
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(.primary)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
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
        .description("Quick access to common tasks like tracking and invoicing.")
        .supportedFamilies([.systemSmall, .systemMedium])
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
