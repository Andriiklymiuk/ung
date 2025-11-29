//
//  TimeTrackingView.swift
//  ung Watch App
//
//  Time tracking interface for Apple Watch
//

import SwiftUI

struct TimeTrackingView: View {
    @EnvironmentObject var watchState: WatchAppState
    @State private var showProjectPicker = false

    var body: some View {
        ScrollView {
            VStack(spacing: 12) {
                if watchState.sessionData.isTracking {
                    activeTrackingView
                } else {
                    startTrackingView
                }
            }
            .padding(.horizontal, 8)
        }
        .navigationTitle("Track")
        .sheet(isPresented: $showProjectPicker) {
            ProjectPickerView { project in
                watchState.startTracking(project: project)
                showProjectPicker = false
            }
        }
    }

    // MARK: - Active Tracking View
    private var activeTrackingView: some View {
        VStack(spacing: 16) {
            // Timer Display
            VStack(spacing: 4) {
                Text(watchState.formattedTrackingTime)
                    .font(.system(size: 36, weight: .bold, design: .rounded))
                    .monospacedDigit()
                    .foregroundColor(.green)

                Text(watchState.sessionData.projectName)
                    .font(.system(size: 14, weight: .medium))
                    .foregroundColor(.secondary)
                    .lineLimit(1)
            }
            .padding(.vertical, 8)

            // Progress Ring
            ZStack {
                Circle()
                    .stroke(Color.green.opacity(0.2), lineWidth: 6)

                Circle()
                    .trim(from: 0, to: min(Double(watchState.sessionData.elapsedSeconds) / 3600.0, 1.0))
                    .stroke(Color.green, style: StrokeStyle(lineWidth: 6, lineCap: .round))
                    .rotationEffect(.degrees(-90))
                    .animation(.linear(duration: 1), value: watchState.sessionData.elapsedSeconds)

                VStack(spacing: 2) {
                    Image(systemName: "timer")
                        .font(.system(size: 20))
                        .foregroundColor(.green)

                    Text("\(watchState.sessionData.elapsedSeconds / 60)m")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                }
            }
            .frame(width: 80, height: 80)

            // Stop Button
            Button(action: { watchState.stopTracking() }) {
                HStack(spacing: 6) {
                    Image(systemName: "stop.fill")
                        .font(.system(size: 14))
                    Text("Stop")
                        .font(.system(size: 14, weight: .semibold))
                }
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 12)
                .background(Color.red)
                .cornerRadius(12)
            }
            .buttonStyle(.plain)
        }
    }

    // MARK: - Start Tracking View
    private var startTrackingView: some View {
        VStack(spacing: 16) {
            // Header
            VStack(spacing: 4) {
                Image(systemName: "clock.fill")
                    .font(.system(size: 32))
                    .foregroundColor(.blue)

                Text("Start Tracking")
                    .font(.system(size: 16, weight: .semibold))
            }
            .padding(.top, 8)

            // Quick Start Buttons
            VStack(spacing: 8) {
                ForEach(watchState.quickProjects.prefix(4), id: \.self) { project in
                    QuickStartButton(project: project) {
                        watchState.startTracking(project: project)
                    }
                }
            }

            // More Projects Button
            Button(action: { showProjectPicker = true }) {
                HStack(spacing: 4) {
                    Image(systemName: "ellipsis.circle")
                        .font(.system(size: 12))
                    Text("More")
                        .font(.system(size: 12, weight: .medium))
                }
                .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
        }
    }
}

// MARK: - Quick Start Button
struct QuickStartButton: View {
    let project: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                Image(systemName: "play.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.green)

                Text(project)
                    .font(.system(size: 13, weight: .medium))
                    .lineLimit(1)

                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .background(Color.green.opacity(0.15))
            .cornerRadius(10)
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Project Picker View
struct ProjectPickerView: View {
    @EnvironmentObject var watchState: WatchAppState
    let onSelect: (String) -> Void

    var body: some View {
        ScrollView {
            VStack(spacing: 8) {
                Text("Select Project")
                    .font(.system(size: 14, weight: .semibold))
                    .padding(.bottom, 8)

                ForEach(watchState.quickProjects, id: \.self) { project in
                    Button(action: { onSelect(project) }) {
                        HStack {
                            Text(project)
                                .font(.system(size: 14))
                            Spacer()
                            Image(systemName: "chevron.right")
                                .font(.system(size: 10))
                                .foregroundColor(.secondary)
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 10)
                        .background(Color(.darkGray).opacity(0.3))
                        .cornerRadius(8)
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, 8)
        }
    }
}

#Preview {
    TimeTrackingView()
        .environmentObject(WatchAppState())
}
