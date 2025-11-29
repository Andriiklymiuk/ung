//
//  LiveActivityService.swift
//  ung
//
//  Manages Live Activities for real-time tracking and pomodoro display
//

import Foundation

#if os(iOS)
import ActivityKit
import UIKit
#endif

// MARK: - Shared Activity Attributes (must match widget extension)
#if os(iOS)
@available(iOS 16.1, *)
struct TrackingActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var elapsedSeconds: Int
        var isActive: Bool
    }

    var projectName: String
    var clientName: String
    var startTime: Date
}

@available(iOS 16.1, *)
struct PomodoroActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var secondsRemaining: Int
        var isBreak: Bool
        var isPaused: Bool
    }

    var sessionsCompleted: Int
    var workMinutes: Int
    var breakMinutes: Int
}
#endif

// MARK: - Live Activity Service
@MainActor
class LiveActivityService: ObservableObject {
    static let shared = LiveActivityService()

    #if os(iOS)
    @available(iOS 16.1, *)
    private var trackingActivity: Activity<TrackingActivityAttributes>?

    @available(iOS 16.1, *)
    private var pomodoroActivity: Activity<PomodoroActivityAttributes>?
    #endif

    private var trackingUpdateTimer: Timer?
    private var pomodoroUpdateTimer: Timer?

    // MARK: - Tracking Live Activity

    func startTrackingActivity(project: String, client: String, startTime: Date) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else {
            print("[LiveActivity] Activities not enabled")
            return
        }

        // End any existing activity
        endTrackingActivity()

        let attributes = TrackingActivityAttributes(
            projectName: project,
            clientName: client,
            startTime: startTime
        )

        let initialState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: Int(Date().timeIntervalSince(startTime)),
            isActive: true
        )

        do {
            trackingActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: initialState, staleDate: nil),
                pushType: nil
            )
            print("[LiveActivity] Started tracking activity")

            // Start update timer
            startTrackingUpdateTimer(startTime: startTime)
        } catch {
            print("[LiveActivity] Error starting tracking activity: \(error)")
        }
        #endif
    }

    func updateTrackingActivity(elapsedSeconds: Int) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = trackingActivity else { return }

        let updatedState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: elapsedSeconds,
            isActive: true
        )

        Task {
            await activity.update(
                ActivityContent(state: updatedState, staleDate: nil)
            )
        }
        #endif
    }

    func endTrackingActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        // Stop update timer
        trackingUpdateTimer?.invalidate()
        trackingUpdateTimer = nil

        guard let activity = trackingActivity else { return }

        let finalState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: 0,
            isActive: false
        )

        Task {
            await activity.end(
                ActivityContent(state: finalState, staleDate: nil),
                dismissalPolicy: .immediate
            )
            print("[LiveActivity] Ended tracking activity")
        }

        trackingActivity = nil
        #endif
    }

    private func startTrackingUpdateTimer(startTime: Date) {
        trackingUpdateTimer?.invalidate()
        trackingUpdateTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                let elapsed = Int(Date().timeIntervalSince(startTime))
                self?.updateTrackingActivity(elapsedSeconds: elapsed)
            }
        }
    }

    // MARK: - Pomodoro Live Activity

    func startPomodoroActivity(sessionsCompleted: Int, workMinutes: Int, breakMinutes: Int, secondsRemaining: Int, isBreak: Bool) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else {
            print("[LiveActivity] Activities not enabled")
            return
        }

        // End any existing activity
        endPomodoroActivity()

        let attributes = PomodoroActivityAttributes(
            sessionsCompleted: sessionsCompleted,
            workMinutes: workMinutes,
            breakMinutes: breakMinutes
        )

        let initialState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: secondsRemaining,
            isBreak: isBreak,
            isPaused: false
        )

        do {
            pomodoroActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: initialState, staleDate: nil),
                pushType: nil
            )
            print("[LiveActivity] Started pomodoro activity")
        } catch {
            print("[LiveActivity] Error starting pomodoro activity: \(error)")
        }
        #endif
    }

    func updatePomodoroActivity(secondsRemaining: Int, isBreak: Bool, isPaused: Bool) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = pomodoroActivity else { return }

        let updatedState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: secondsRemaining,
            isBreak: isBreak,
            isPaused: isPaused
        )

        Task {
            await activity.update(
                ActivityContent(state: updatedState, staleDate: nil)
            )
        }
        #endif
    }

    func endPomodoroActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        guard let activity = pomodoroActivity else { return }

        let finalState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: 0,
            isBreak: false,
            isPaused: false
        )

        Task {
            await activity.end(
                ActivityContent(state: finalState, staleDate: nil),
                dismissalPolicy: .immediate
            )
            print("[LiveActivity] Ended pomodoro activity")
        }

        pomodoroActivity = nil
        #endif
    }

    // MARK: - Check Authorization

    func areActivitiesEnabled() -> Bool {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return false }
        return ActivityAuthorizationInfo().areActivitiesEnabled
        #else
        return false
        #endif
    }
}
