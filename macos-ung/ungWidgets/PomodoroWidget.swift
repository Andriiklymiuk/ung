//
//  PomodoroWidget.swift
//  ungWidgets
//
//  Premium pomodoro timer widget with Telegram-inspired design
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
    @Environment(\.colorScheme) var colorScheme
    var entry: PomodoroEntry

    private var timerGradient: LinearGradient {
        if !entry.data.pomodoroActive {
            return LinearGradient(colors: [Color.gray.opacity(0.5), Color.gray.opacity(0.3)], startPoint: .top, endPoint: .bottom)
        }
        return entry.data.pomodoroIsBreak ? WidgetColors.breakGradient : WidgetColors.focusGradient
    }

    private var timerColor: Color {
        if !entry.data.pomodoroActive {
            return .gray
        }
        return entry.data.pomodoroIsBreak ? WidgetColors.breakGreen : WidgetColors.focusOrange
    }

    private var statusIcon: String {
        if !entry.data.pomodoroActive {
            return "brain.head.profile"
        }
        return entry.data.pomodoroIsBreak ? "cup.and.saucer.fill" : "brain.head.profile.fill"
    }

    private var statusText: String {
        if !entry.data.pomodoroActive {
            return "Ready to Focus"
        }
        return entry.data.pomodoroIsBreak ? "Break Time" : "Stay Focused"
    }

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

    // MARK: - Small View
    private var smallView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            VStack(spacing: 8) {
                // Timer ring
                ZStack {
                    // Background ring
                    Circle()
                        .stroke(timerGradient.opacity(0.2), lineWidth: 6)
                        .frame(width: 70, height: 70)

                    if entry.data.pomodoroActive {
                        // Progress ring
                        Circle()
                            .trim(from: 0, to: calculateProgress())
                            .stroke(timerGradient, style: StrokeStyle(lineWidth: 6, lineCap: .round))
                            .frame(width: 70, height: 70)
                            .rotationEffect(.degrees(-90))

                        // Glow
                        Circle()
                            .trim(from: 0, to: calculateProgress())
                            .stroke(timerGradient, style: StrokeStyle(lineWidth: 12, lineCap: .round))
                            .frame(width: 70, height: 70)
                            .rotationEffect(.degrees(-90))
                            .blur(radius: 6)
                            .opacity(0.5)
                    }

                    // Center content
                    VStack(spacing: 2) {
                        Image(systemName: statusIcon)
                            .font(.system(size: 18))
                            .foregroundStyle(timerGradient)

                        if entry.data.pomodoroActive {
                            Text(entry.data.pomodoroTimeFormatted)
                                .font(.system(size: 14, weight: .bold, design: .monospaced))
                                .foregroundColor(timerColor)
                                .monospacedDigit()
                        }
                    }
                }

                // Status
                Text(statusText)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(entry.data.pomodoroActive ? timerColor : WidgetColors.textTertiary)

                // Session dots
                if entry.data.pomodoroSessionsCompleted > 0 || entry.data.pomodoroActive {
                    HStack(spacing: 4) {
                        ForEach(0..<4, id: \.self) { index in
                            Circle()
                                .fill(index < entry.data.pomodoroSessionsCompleted % 4 ? timerColor : Color.gray.opacity(0.2))
                                .frame(width: 6, height: 6)
                        }
                    }
                }
            }
            .padding(12)
        }
        .widgetURL(URL(string: "ung://pomodoro"))
    }

    // MARK: - Medium View
    private var mediumView: some View {
        ZStack {
            // Background
            ContainerRelativeShape()
                .fill(colorScheme == .dark
                    ? Color(hex: "1C1C1E")
                    : Color(hex: "F8F9FA")
                )

            HStack(spacing: 20) {
                // Left - Timer ring
                ZStack {
                    // Background ring
                    Circle()
                        .stroke(timerGradient.opacity(0.2), lineWidth: 8)
                        .frame(width: 90, height: 90)

                    if entry.data.pomodoroActive {
                        // Progress ring
                        Circle()
                            .trim(from: 0, to: calculateProgress())
                            .stroke(timerGradient, style: StrokeStyle(lineWidth: 8, lineCap: .round))
                            .frame(width: 90, height: 90)
                            .rotationEffect(.degrees(-90))

                        // Glow
                        Circle()
                            .trim(from: 0, to: calculateProgress())
                            .stroke(timerGradient, style: StrokeStyle(lineWidth: 16, lineCap: .round))
                            .frame(width: 90, height: 90)
                            .rotationEffect(.degrees(-90))
                            .blur(radius: 8)
                            .opacity(0.5)
                    }

                    // Center content
                    VStack(spacing: 2) {
                        Image(systemName: statusIcon)
                            .font(.system(size: 22))
                            .foregroundStyle(timerGradient)

                        if entry.data.pomodoroActive {
                            Text(entry.data.pomodoroTimeFormatted)
                                .font(.system(size: 18, weight: .bold, design: .monospaced))
                                .foregroundColor(timerColor)
                                .monospacedDigit()
                        }
                    }
                }

                // Right - Status and info
                VStack(alignment: .leading, spacing: 8) {
                    // Status header
                    HStack(spacing: 6) {
                        if entry.data.pomodoroActive {
                            PulsingDot(color: timerColor, size: 6)
                        } else {
                            Circle()
                                .fill(Color.gray.opacity(0.3))
                                .frame(width: 6, height: 6)
                        }

                        Text(entry.data.pomodoroActive ? (entry.data.pomodoroIsBreak ? "BREAK" : "FOCUS") : "READY")
                            .font(.system(size: 10, weight: .bold))
                            .tracking(1)
                            .foregroundColor(entry.data.pomodoroActive ? timerColor : WidgetColors.textTertiary)
                    }

                    // Status text
                    Text(statusText)
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundColor(WidgetColors.textPrimary)

                    if !entry.data.pomodoroActive {
                        Text("Tap to start a focus session")
                            .font(.system(size: 12))
                            .foregroundColor(WidgetColors.textTertiary)
                    }

                    Spacer()

                    // Session progress
                    VStack(alignment: .leading, spacing: 4) {
                        HStack(spacing: 4) {
                            Image(systemName: "flame.fill")
                                .font(.system(size: 11))
                                .foregroundColor(.orange)
                            Text("\(entry.data.pomodoroSessionsCompleted) sessions today")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(WidgetColors.textTertiary)
                        }

                        // Session indicators
                        HStack(spacing: 4) {
                            ForEach(0..<4, id: \.self) { index in
                                RoundedRectangle(cornerRadius: 2)
                                    .fill(index < entry.data.pomodoroSessionsCompleted % 4 ? timerColor : Color.gray.opacity(0.2))
                                    .frame(width: 24, height: 4)
                            }
                        }
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .padding(16)
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
        .configurationDisplayName("Focus Timer")
        .description("Track your Pomodoro sessions with style.")
        .supportedFamilies([.systemSmall, .systemMedium])
        .contentMarginsDisabled()
    }
}

#Preview(as: .systemMedium) {
    PomodoroWidget()
} timeline: {
    PomodoroEntry(date: .now, data: WidgetData())
    var active = WidgetData()
    active.pomodoroActive = true
    active.pomodoroIsBreak = false
    active.pomodoroSecondsRemaining = 18 * 60 + 45
    active.pomodoroSessionsCompleted = 3
    return PomodoroEntry(date: .now, data: active)
}
