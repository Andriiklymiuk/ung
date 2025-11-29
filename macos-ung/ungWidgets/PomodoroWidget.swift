//
//  PomodoroWidget.swift
//  ungWidgets
//
//  Shows pomodoro timer status
//

import SwiftUI
import WidgetKit

// MARK: - Timeline Entry
struct PomodoroEntry: TimelineEntry {
    let date: Date
    let data: WidgetData
}

// MARK: - Timeline Provider
struct PomodoroProvider: TimelineProvider {
    func placeholder(in context: Context) -> PomodoroEntry {
        PomodoroEntry(date: Date(), data: WidgetData())
    }

    func getSnapshot(in context: Context, completion: @escaping (PomodoroEntry) -> Void) {
        let entry = PomodoroEntry(date: Date(), data: WidgetData.load())
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<PomodoroEntry>) -> Void) {
        let data = WidgetData.load()
        let currentDate = Date()

        var entries: [PomodoroEntry] = []

        if data.pomodoroActive {
            // Update every second for active timer (up to 60 entries)
            for secondOffset in stride(from: 0, to: min(data.pomodoroSecondsRemaining, 60), by: 1) {
                let entryDate = Calendar.current.date(byAdding: .second, value: secondOffset, to: currentDate)!
                var updatedData = data
                updatedData.pomodoroSecondsRemaining = max(0, data.pomodoroSecondsRemaining - secondOffset)
                entries.append(PomodoroEntry(date: entryDate, data: updatedData))
            }
        } else {
            entries.append(PomodoroEntry(date: currentDate, data: data))
        }

        let refreshDate = Calendar.current.date(byAdding: .minute, value: data.pomodoroActive ? 1 : 15, to: currentDate)!
        let timeline = Timeline(entries: entries, policy: .after(refreshDate))
        completion(timeline)
    }
}

// MARK: - Widget View
struct PomodoroWidgetView: View {
    @Environment(\.widgetFamily) var family
    var entry: PomodoroEntry

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

    private var timerColor: Color {
        if !entry.data.pomodoroActive {
            return .gray
        }
        return entry.data.pomodoroIsBreak ? .green : .orange
    }

    private var statusIcon: String {
        if !entry.data.pomodoroActive {
            return "brain.head.profile"
        }
        return entry.data.pomodoroIsBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill"
    }

    private var statusText: String {
        if !entry.data.pomodoroActive {
            return "Ready"
        }
        return entry.data.pomodoroIsBreak ? "Break Time" : "Focus Mode"
    }

    // MARK: - Small View
    private var smallView: some View {
        VStack(spacing: 8) {
            // Status icon
            ZStack {
                Circle()
                    .fill(timerColor.opacity(0.15))
                    .frame(width: 50, height: 50)

                if entry.data.pomodoroActive {
                    // Progress ring
                    Circle()
                        .trim(from: 0, to: calculateProgress())
                        .stroke(timerColor, style: StrokeStyle(lineWidth: 4, lineCap: .round))
                        .frame(width: 50, height: 50)
                        .rotationEffect(.degrees(-90))
                }

                Image(systemName: statusIcon)
                    .font(.system(size: 20))
                    .foregroundColor(timerColor)
            }

            if entry.data.pomodoroActive {
                Text(entry.data.pomodoroTimeFormatted)
                    .font(.system(size: 24, weight: .bold, design: .monospaced))
                    .foregroundColor(timerColor)
                    .monospacedDigit()

                Text(statusText)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
            } else {
                Text("Pomodoro")
                    .font(.system(size: 14, weight: .semibold))

                Text("Tap to start")
                    .font(.system(size: 11))
                    .foregroundColor(.blue)
            }

            // Sessions completed
            if entry.data.pomodoroSessionsCompleted > 0 {
                HStack(spacing: 2) {
                    ForEach(0..<min(entry.data.pomodoroSessionsCompleted, 4), id: \.self) { _ in
                        Circle()
                            .fill(Color.orange)
                            .frame(width: 6, height: 6)
                    }
                    if entry.data.pomodoroSessionsCompleted > 4 {
                        Text("+\(entry.data.pomodoroSessionsCompleted - 4)")
                            .font(.system(size: 8))
                            .foregroundColor(.secondary)
                    }
                }
            }
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://pomodoro"))
    }

    // MARK: - Medium View
    private var mediumView: some View {
        HStack(spacing: 20) {
            // Timer circle
            ZStack {
                Circle()
                    .stroke(timerColor.opacity(0.2), lineWidth: 8)
                    .frame(width: 90, height: 90)

                if entry.data.pomodoroActive {
                    Circle()
                        .trim(from: 0, to: calculateProgress())
                        .stroke(timerColor, style: StrokeStyle(lineWidth: 8, lineCap: .round))
                        .frame(width: 90, height: 90)
                        .rotationEffect(.degrees(-90))
                }

                VStack(spacing: 2) {
                    Image(systemName: statusIcon)
                        .font(.system(size: 20))
                        .foregroundColor(timerColor)

                    if entry.data.pomodoroActive {
                        Text(entry.data.pomodoroTimeFormatted)
                            .font(.system(size: 16, weight: .bold, design: .monospaced))
                            .foregroundColor(timerColor)
                            .monospacedDigit()
                    }
                }
            }

            // Status and info
            VStack(alignment: .leading, spacing: 8) {
                Text(statusText)
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(entry.data.pomodoroActive ? timerColor : .primary)

                if entry.data.pomodoroActive {
                    Text(entry.data.pomodoroIsBreak ? "Take a break!" : "Stay focused")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                } else {
                    Text("Start a focus session")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                }

                Spacer()

                // Sessions today
                HStack(spacing: 4) {
                    Image(systemName: "flame.fill")
                        .font(.system(size: 12))
                        .foregroundColor(.orange)
                    Text("\(entry.data.pomodoroSessionsCompleted) sessions")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                }

                // Session indicators
                HStack(spacing: 3) {
                    ForEach(0..<4, id: \.self) { index in
                        RoundedRectangle(cornerRadius: 2)
                            .fill(index < entry.data.pomodoroSessionsCompleted % 4 ? Color.orange : Color.gray.opacity(0.3))
                            .frame(width: 20, height: 4)
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding()
        .containerBackground(for: .widget) {
            Color(.systemBackground)
        }
        .widgetURL(URL(string: "ung://pomodoro"))
    }

    private func calculateProgress() -> Double {
        let totalSeconds: Int
        if entry.data.pomodoroIsBreak {
            totalSeconds = entry.data.pomodoroSessionsCompleted % 4 == 0 ? 15 * 60 : 5 * 60
        } else {
            totalSeconds = 25 * 60
        }
        return 1.0 - (Double(entry.data.pomodoroSecondsRemaining) / Double(totalSeconds))
    }
}

// MARK: - Widget Configuration
struct PomodoroWidget: Widget {
    let kind: String = "PomodoroWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: PomodoroProvider()) { entry in
            PomodoroWidgetView(entry: entry)
        }
        .configurationDisplayName("Pomodoro Timer")
        .description("Quick access to your focus timer.")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

#Preview(as: .systemSmall) {
    PomodoroWidget()
} timeline: {
    PomodoroEntry(date: .now, data: WidgetData())
    var active = WidgetData()
    active.pomodoroActive = true
    active.pomodoroIsBreak = false
    active.pomodoroSecondsRemaining = 15 * 60 + 30
    active.pomodoroSessionsCompleted = 3
    return PomodoroEntry(date: .now, data: active)
}
