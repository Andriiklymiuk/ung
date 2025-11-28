//
//  PomodoroContent.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct PomodoroContent: View {
    @EnvironmentObject var appState: AppState
    @State private var showSettings = false

    var body: some View {
        HStack(spacing: 40) {
            // Main timer display
            timerSection
                .frame(maxWidth: 500)

            // Stats and settings
            VStack(spacing: 24) {
                statsCard
                settingsCard
            }
            .frame(maxWidth: 350)
        }
        .padding(40)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // MARK: - Timer Section
    private var timerSection: some View {
        VStack(spacing: 32) {
            // Status text
            Text(appState.pomodoroState.statusText)
                .font(.system(size: 18, weight: .medium))
                .foregroundColor(.secondary)

            // Circular timer
            ZStack {
                // Background circle
                Circle()
                    .stroke(Color.gray.opacity(0.2), lineWidth: 20)
                    .frame(width: 280, height: 280)

                // Progress circle
                Circle()
                    .trim(from: 0, to: appState.pomodoroState.progress)
                    .stroke(
                        LinearGradient(
                            colors: appState.pomodoroState.isBreak
                                ? [.green, .mint]
                                : [.red, .orange],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        style: StrokeStyle(lineWidth: 20, lineCap: .round)
                    )
                    .frame(width: 280, height: 280)
                    .rotationEffect(.degrees(-90))
                    .animation(.linear(duration: 1), value: appState.pomodoroState.secondsRemaining)

                // Timer display
                VStack(spacing: 8) {
                    Text(appState.pomodoroState.formattedTime)
                        .font(.system(size: 64, weight: .bold, design: .rounded))
                        .foregroundColor(.primary)
                        .monospacedDigit()

                    if appState.pomodoroState.isActive {
                        Text(appState.pomodoroState.isBreak ? "Take a break" : "Stay focused")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)
                    }
                }
            }

            // Control buttons
            HStack(spacing: 20) {
                if appState.pomodoroState.isActive {
                    // Pause/Resume button
                    Button(action: {
                        if appState.pomodoroState.isPaused {
                            appState.resumePomodoro()
                        } else {
                            appState.pausePomodoro()
                        }
                    }) {
                        HStack(spacing: 8) {
                            Image(systemName: appState.pomodoroState.isPaused ? "play.fill" : "pause.fill")
                            Text(appState.pomodoroState.isPaused ? "Resume" : "Pause")
                        }
                        .font(.system(size: 15, weight: .medium))
                        .foregroundColor(.primary)
                        .padding(.horizontal, 24)
                        .padding(.vertical, 12)
                        .background(
                            RoundedRectangle(cornerRadius: 12)
                                .fill(Color(nsColor: .controlBackgroundColor))
                        )
                    }
                    .buttonStyle(.plain)

                    // Skip button
                    Button(action: { appState.skipPomodoro() }) {
                        HStack(spacing: 8) {
                            Image(systemName: "forward.fill")
                            Text("Skip")
                        }
                        .font(.system(size: 15, weight: .medium))
                        .foregroundColor(.secondary)
                        .padding(.horizontal, 24)
                        .padding(.vertical, 12)
                        .background(
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(Color.secondary.opacity(0.3), lineWidth: 1)
                        )
                    }
                    .buttonStyle(.plain)

                    // Stop button
                    Button(action: { appState.stopPomodoro() }) {
                        HStack(spacing: 8) {
                            Image(systemName: "stop.fill")
                            Text("Stop")
                        }
                        .font(.system(size: 15, weight: .medium))
                        .foregroundColor(.red)
                        .padding(.horizontal, 24)
                        .padding(.vertical, 12)
                        .background(
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(Color.red.opacity(0.3), lineWidth: 1)
                        )
                    }
                    .buttonStyle(.plain)
                } else {
                    // Start button
                    Button(action: { appState.startPomodoro() }) {
                        HStack(spacing: 10) {
                            Image(systemName: "play.fill")
                                .font(.system(size: 18))
                            Text("Start Focus Session")
                                .font(.system(size: 16, weight: .semibold))
                        }
                        .foregroundColor(.white)
                        .padding(.horizontal, 32)
                        .padding(.vertical, 16)
                        .background(
                            RoundedRectangle(cornerRadius: 16)
                                .fill(
                                    LinearGradient(
                                        colors: [.red, .orange],
                                        startPoint: .leading,
                                        endPoint: .trailing
                                    )
                                )
                                .shadow(color: .red.opacity(0.3), radius: 10, y: 5)
                        )
                    }
                    .buttonStyle(.plain)
                }
            }
        }
    }

    // MARK: - Stats Card
    private var statsCard: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Today's Focus")
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.primary)

            HStack(spacing: 24) {
                StatItem(
                    value: "\(appState.pomodoroState.sessionsCompleted)",
                    label: "Sessions",
                    icon: "checkmark.circle.fill",
                    color: .green
                )

                StatItem(
                    value: "\(appState.pomodoroState.sessionsCompleted * appState.pomodoroState.workMinutes)m",
                    label: "Focus Time",
                    icon: "brain.head.profile",
                    color: .orange
                )

                StatItem(
                    value: "\(appState.pomodoroState.sessionsCompleted / appState.pomodoroState.sessionsUntilLongBreak)",
                    label: "Cycles",
                    icon: "arrow.triangle.2.circlepath",
                    color: .blue
                )
            }
        }
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }

    // MARK: - Settings Card
    private var settingsCard: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Timer Settings")
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.primary)

            VStack(spacing: 12) {
                TimerSettingRow(
                    label: "Work Duration",
                    value: $appState.pomodoroState.workMinutes,
                    range: 15...60,
                    unit: "min"
                )

                TimerSettingRow(
                    label: "Short Break",
                    value: $appState.pomodoroState.breakMinutes,
                    range: 3...15,
                    unit: "min"
                )

                TimerSettingRow(
                    label: "Long Break",
                    value: $appState.pomodoroState.longBreakMinutes,
                    range: 10...30,
                    unit: "min"
                )

                TimerSettingRow(
                    label: "Sessions Until Long Break",
                    value: $appState.pomodoroState.sessionsUntilLongBreak,
                    range: 2...6,
                    unit: ""
                )
            }
        }
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

// MARK: - Supporting Views
struct StatItem: View {
    let value: String
    let label: String
    let icon: String
    let color: Color

    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.system(size: 20))
                .foregroundColor(color)

            Text(value)
                .font(.system(size: 20, weight: .bold, design: .rounded))
                .foregroundColor(.primary)

            Text(label)
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}

struct TimerSettingRow: View {
    let label: String
    @Binding var value: Int
    let range: ClosedRange<Int>
    let unit: String

    var body: some View {
        HStack {
            Text(label)
                .font(.system(size: 13))
                .foregroundColor(.secondary)

            Spacer()

            HStack(spacing: 8) {
                Button(action: { if value > range.lowerBound { value -= 1 } }) {
                    Image(systemName: "minus.circle.fill")
                        .font(.system(size: 18))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)

                Text("\(value)\(unit)")
                    .font(.system(size: 14, weight: .medium, design: .monospaced))
                    .foregroundColor(.primary)
                    .frame(width: 50)

                Button(action: { if value < range.upperBound { value += 1 } }) {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 18))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
            }
        }
    }
}

#Preview {
    PomodoroContent()
        .environmentObject(AppState())
        .frame(width: 900, height: 600)
}
