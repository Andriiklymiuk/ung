//
//  PomodoroWatchView.swift
//  ung Watch App
//
//  Pomodoro timer interface for Apple Watch
//

import SwiftUI

struct PomodoroWatchView: View {
    @EnvironmentObject var watchState: WatchAppState

    var timerColor: Color {
        watchState.pomodoroData.isBreak ? .green : .orange
    }

    var body: some View {
        ScrollView {
            VStack(spacing: 12) {
                if watchState.pomodoroData.isActive {
                    activeTimerView
                } else {
                    startView
                }
            }
            .padding(.horizontal, 8)
        }
        .navigationTitle("Focus")
    }

    // MARK: - Active Timer View
    private var activeTimerView: some View {
        VStack(spacing: 16) {
            // Status
            Text(statusText)
                .font(.system(size: 12, weight: .medium))
                .foregroundColor(timerColor)
                .padding(.horizontal, 12)
                .padding(.vertical, 4)
                .background(timerColor.opacity(0.2))
                .cornerRadius(8)

            // Circular Timer
            ZStack {
                // Background ring
                Circle()
                    .stroke(timerColor.opacity(0.2), lineWidth: 10)

                // Progress ring
                Circle()
                    .trim(from: 0, to: watchState.pomodoroProgress)
                    .stroke(
                        LinearGradient(
                            colors: watchState.pomodoroData.isBreak
                                ? [.green, .mint]
                                : [.orange, .red],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        style: StrokeStyle(lineWidth: 10, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))
                    .animation(.linear(duration: 1), value: watchState.pomodoroData.secondsRemaining)

                // Time display
                VStack(spacing: 2) {
                    Text(watchState.formattedPomodoroTime)
                        .font(.system(size: 32, weight: .bold, design: .rounded))
                        .monospacedDigit()
                        .foregroundColor(timerColor)

                    Text(watchState.pomodoroData.isBreak ? "Break" : "Focus")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }
            }
            .frame(width: 130, height: 130)

            // Sessions counter
            HStack(spacing: 4) {
                ForEach(0..<4, id: \.self) { index in
                    Circle()
                        .fill(index < watchState.pomodoroData.sessionsCompleted % 4 ? timerColor : Color.gray.opacity(0.3))
                        .frame(width: 8, height: 8)
                }
                Text("\(watchState.pomodoroData.sessionsCompleted)")
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(.secondary)
            }

            // Control buttons
            HStack(spacing: 12) {
                // Pause/Resume
                Button(action: {
                    if watchState.pomodoroData.isPaused {
                        watchState.resumePomodoro()
                    } else {
                        watchState.pausePomodoro()
                    }
                }) {
                    Image(systemName: watchState.pomodoroData.isPaused ? "play.fill" : "pause.fill")
                        .font(.system(size: 16))
                        .foregroundColor(.white)
                        .frame(width: 44, height: 44)
                        .background(timerColor)
                        .cornerRadius(22)
                }
                .buttonStyle(.plain)

                // Skip
                Button(action: { watchState.skipPomodoro() }) {
                    Image(systemName: "forward.fill")
                        .font(.system(size: 14))
                        .foregroundColor(.white)
                        .frame(width: 40, height: 40)
                        .background(Color.gray.opacity(0.5))
                        .cornerRadius(20)
                }
                .buttonStyle(.plain)

                // Stop
                Button(action: { watchState.stopPomodoro() }) {
                    Image(systemName: "stop.fill")
                        .font(.system(size: 14))
                        .foregroundColor(.white)
                        .frame(width: 40, height: 40)
                        .background(Color.red.opacity(0.8))
                        .cornerRadius(20)
                }
                .buttonStyle(.plain)
            }
        }
    }

    // MARK: - Start View
    private var startView: some View {
        VStack(spacing: 20) {
            // Icon
            ZStack {
                Circle()
                    .fill(Color.orange.opacity(0.2))
                    .frame(width: 80, height: 80)

                Image(systemName: "brain.head.profile")
                    .font(.system(size: 32))
                    .foregroundColor(.orange)
            }

            // Title
            VStack(spacing: 4) {
                Text("Pomodoro")
                    .font(.system(size: 18, weight: .bold))

                Text("\(watchState.pomodoroData.workMinutes) min focus")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            // Start Button
            Button(action: { watchState.startPomodoro() }) {
                HStack(spacing: 8) {
                    Image(systemName: "play.fill")
                        .font(.system(size: 14))
                    Text("Start Focus")
                        .font(.system(size: 14, weight: .semibold))
                }
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 14)
                .background(
                    LinearGradient(
                        colors: [.orange, .red],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                )
                .cornerRadius(14)
            }
            .buttonStyle(.plain)

            // Quick tip
            Text("Stay focused for \(watchState.pomodoroData.workMinutes) minutes")
                .font(.system(size: 10))
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
        }
        .padding(.vertical, 8)
    }

    private var statusText: String {
        if watchState.pomodoroData.isPaused {
            return "Paused"
        }
        if watchState.pomodoroData.isBreak {
            return watchState.pomodoroData.sessionsCompleted % 4 == 0 ? "Long Break" : "Short Break"
        }
        return "Focus Time"
    }
}

#Preview {
    PomodoroWatchView()
        .environmentObject(WatchAppState())
}
