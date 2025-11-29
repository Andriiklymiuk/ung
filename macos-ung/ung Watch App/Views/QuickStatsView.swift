//
//  QuickStatsView.swift
//  ung Watch App
//
//  Quick stats and glanceable information for Apple Watch
//

import SwiftUI

struct QuickStatsView: View {
    @EnvironmentObject var watchState: WatchAppState

    var body: some View {
        ScrollView {
            VStack(spacing: 12) {
                // Weekly Progress Ring
                weeklyProgressCard

                // Today's Hours
                todayCard

                // Pending Work
                pendingCard

                // Sync Status
                syncStatusCard
            }
            .padding(.horizontal, 8)
        }
        .navigationTitle("Stats")
    }

    // MARK: - Weekly Progress Card
    private var weeklyProgressCard: some View {
        VStack(spacing: 8) {
            ZStack {
                // Background ring
                Circle()
                    .stroke(Color.blue.opacity(0.2), lineWidth: 8)

                // Progress ring
                Circle()
                    .trim(from: 0, to: watchState.weeklyProgress)
                    .stroke(
                        LinearGradient(
                            colors: [.blue, .cyan],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        style: StrokeStyle(lineWidth: 8, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))

                VStack(spacing: 2) {
                    Text(String(format: "%.1fh", watchState.statsData.weeklyHours))
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundColor(.blue)

                    Text("of \(Int(watchState.statsData.weeklyTarget))h")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                }
            }
            .frame(width: 90, height: 90)

            Text("Weekly Progress")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)
        }
        .padding(.vertical, 8)
    }

    // MARK: - Today Card
    private var todayCard: some View {
        HStack(spacing: 12) {
            Image(systemName: "sun.max.fill")
                .font(.system(size: 20))
                .foregroundColor(.yellow)

            VStack(alignment: .leading, spacing: 2) {
                Text("Today")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)

                Text(String(format: "%.1fh", watchState.statsData.todayHours))
                    .font(.system(size: 16, weight: .bold))
            }

            Spacer()
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(Color.yellow.opacity(0.15))
        .cornerRadius(12)
    }

    // MARK: - Pending Card
    private var pendingCard: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(Color.orange.opacity(0.2))
                    .frame(width: 36, height: 36)

                Text("\(watchState.statsData.pendingInvoices)")
                    .font(.system(size: 14, weight: .bold))
                    .foregroundColor(.orange)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text("Pending")
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)

                Text("Invoices")
                    .font(.system(size: 12, weight: .medium))
            }

            Spacer()

            Image(systemName: "chevron.right")
                .font(.system(size: 10))
                .foregroundColor(.secondary)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(Color.orange.opacity(0.1))
        .cornerRadius(12)
    }

    // MARK: - Sync Status Card
    private var syncStatusCard: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(watchState.isConnected ? Color.green : Color.gray)
                .frame(width: 8, height: 8)

            Text(watchState.isConnected ? "Connected" : "Offline")
                .font(.system(size: 10))
                .foregroundColor(.secondary)

            Spacer()

            if let lastSync = watchState.lastSyncTime {
                Text(lastSync, style: .relative)
                    .font(.system(size: 9))
                    .foregroundColor(.secondary)
            }

            Button(action: { watchState.requestSync() }) {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: 12))
                    .foregroundColor(.blue)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(Color(.darkGray).opacity(0.3))
        .cornerRadius(10)
    }
}

// MARK: - Complication Views
struct CircularComplicationView: View {
    let hours: Double
    let target: Double

    var progress: Double {
        guard target > 0 else { return 0 }
        return min(hours / target, 1.0)
    }

    var body: some View {
        ZStack {
            Circle()
                .stroke(Color.blue.opacity(0.3), lineWidth: 4)

            Circle()
                .trim(from: 0, to: progress)
                .stroke(Color.blue, style: StrokeStyle(lineWidth: 4, lineCap: .round))
                .rotationEffect(.degrees(-90))

            Text(String(format: "%.0f", hours))
                .font(.system(size: 14, weight: .bold, design: .rounded))
        }
    }
}

struct CornerComplicationView: View {
    let isTracking: Bool
    let hours: Double

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            Image(systemName: isTracking ? "timer" : "clock.fill")
                .font(.system(size: 12))
                .foregroundColor(isTracking ? .green : .blue)

            Text(String(format: "%.1fh", hours))
                .font(.system(size: 10, weight: .medium, design: .rounded))
        }
    }
}

struct InlineComplicationView: View {
    let isTracking: Bool
    let projectName: String
    let hours: Double

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: isTracking ? "timer" : "clock")
                .font(.system(size: 10))

            if isTracking {
                Text(projectName)
                    .font(.system(size: 10))
                    .lineLimit(1)
            } else {
                Text(String(format: "%.1fh today", hours))
                    .font(.system(size: 10))
            }
        }
    }
}

#Preview {
    QuickStatsView()
        .environmentObject(WatchAppState())
}
