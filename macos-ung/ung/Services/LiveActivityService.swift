//
//  LiveActivityService.swift
//  ung
//
//  Premium Live Activities service for real-time tracking and pomodoro display
//  Supports Dynamic Island and Lock Screen on iOS 16.1+
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
        var billableAmount: Double?
    }

    var projectName: String
    var clientName: String
    var startTime: Date
    var hourlyRate: Double?
    var currency: String
}

@available(iOS 16.1, *)
struct PomodoroActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var secondsRemaining: Int
        var isBreak: Bool
        var isPaused: Bool
        var currentSessionNumber: Int
    }

    var sessionsCompleted: Int
    var workMinutes: Int
    var breakMinutes: Int
    var longBreakMinutes: Int
    var projectName: String?
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

    // Store start time for tracking updates
    private var trackingStartTime: Date?
    private var trackingHourlyRate: Double?
    private var trackingCurrency: String = "USD"

    // MARK: - Tracking Live Activity

    /// Start a new tracking Live Activity with optional billing info
    func startTrackingActivity(
        project: String,
        client: String,
        startTime: Date,
        hourlyRate: Double? = nil,
        currency: String = "USD"
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else {
            print("[LiveActivity] Activities not enabled by user")
            return
        }

        // End any existing activity first
        endTrackingActivity()

        // Store for timer updates
        trackingStartTime = startTime
        trackingHourlyRate = hourlyRate
        trackingCurrency = currency

        let attributes = TrackingActivityAttributes(
            projectName: project,
            clientName: client,
            startTime: startTime,
            hourlyRate: hourlyRate,
            currency: currency
        )

        let elapsed = Int(Date().timeIntervalSince(startTime))
        let billable = hourlyRate.map { (Double(elapsed) / 3600.0) * $0 }

        let initialState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: elapsed,
            isActive: true,
            billableAmount: billable
        )

        // Calculate stale date (show as stale if no update for 5 seconds)
        let staleDate = Date().addingTimeInterval(5)

        do {
            trackingActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: initialState, staleDate: staleDate),
                pushType: nil
            )
            print("[LiveActivity] Started tracking activity for '\(project)'")

            // Start automatic updates
            startTrackingUpdateTimer(startTime: startTime)
        } catch {
            print("[LiveActivity] Error starting tracking activity: \(error.localizedDescription)")
        }
        #endif
    }

    /// Update the tracking Live Activity with new elapsed time
    func updateTrackingActivity(elapsedSeconds: Int) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = trackingActivity else { return }

        let billable = trackingHourlyRate.map { (Double(elapsedSeconds) / 3600.0) * $0 }

        let updatedState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: elapsedSeconds,
            isActive: true,
            billableAmount: billable
        )

        // Set next stale date
        let staleDate = Date().addingTimeInterval(5)

        Task {
            await activity.update(
                ActivityContent(state: updatedState, staleDate: staleDate)
            )
        }
        #endif
    }

    /// End the tracking Live Activity
    func endTrackingActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        // Stop update timer
        trackingUpdateTimer?.invalidate()
        trackingUpdateTimer = nil
        trackingStartTime = nil

        guard let activity = trackingActivity else { return }

        // Calculate final elapsed time
        let finalElapsed = trackingStartTime.map { Int(Date().timeIntervalSince($0)) } ?? 0
        let billable = trackingHourlyRate.map { (Double(finalElapsed) / 3600.0) * $0 }

        let finalState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: finalElapsed,
            isActive: false,
            billableAmount: billable
        )

        Task {
            // End with default dismissal (lingers briefly then disappears)
            await activity.end(
                ActivityContent(state: finalState, staleDate: nil),
                dismissalPolicy: .default
            )
            print("[LiveActivity] Ended tracking activity")
        }

        trackingActivity = nil
        #endif
    }

    /// End tracking with immediate dismissal
    func endTrackingActivityImmediately() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        trackingUpdateTimer?.invalidate()
        trackingUpdateTimer = nil

        guard let activity = trackingActivity else { return }

        let finalState = TrackingActivityAttributes.ContentState(
            elapsedSeconds: 0,
            isActive: false,
            billableAmount: nil
        )

        Task {
            await activity.end(
                ActivityContent(state: finalState, staleDate: nil),
                dismissalPolicy: .immediate
            )
        }

        trackingActivity = nil
        #endif
    }

    private func startTrackingUpdateTimer(startTime: Date) {
        trackingUpdateTimer?.invalidate()

        // Update every second
        trackingUpdateTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                let elapsed = Int(Date().timeIntervalSince(startTime))
                self?.updateTrackingActivity(elapsedSeconds: elapsed)
            }
        }

        // Make sure timer runs even when scrolling
        RunLoop.current.add(trackingUpdateTimer!, forMode: .common)
    }

    // MARK: - Pomodoro Live Activity

    /// Start a new Pomodoro Live Activity
    func startPomodoroActivity(
        sessionsCompleted: Int,
        workMinutes: Int,
        breakMinutes: Int,
        longBreakMinutes: Int = 15,
        secondsRemaining: Int,
        isBreak: Bool,
        currentSession: Int = 1,
        projectName: String? = nil
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else {
            print("[LiveActivity] Activities not enabled by user")
            return
        }

        // End any existing pomodoro activity
        endPomodoroActivity()

        let attributes = PomodoroActivityAttributes(
            sessionsCompleted: sessionsCompleted,
            workMinutes: workMinutes,
            breakMinutes: breakMinutes,
            longBreakMinutes: longBreakMinutes,
            projectName: projectName
        )

        let initialState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: secondsRemaining,
            isBreak: isBreak,
            isPaused: false,
            currentSessionNumber: currentSession
        )

        // Stale if no update for 3 seconds
        let staleDate = Date().addingTimeInterval(3)

        do {
            pomodoroActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: initialState, staleDate: staleDate),
                pushType: nil
            )
            print("[LiveActivity] Started pomodoro activity - \(isBreak ? "Break" : "Focus") mode")
        } catch {
            print("[LiveActivity] Error starting pomodoro activity: \(error.localizedDescription)")
        }
        #endif
    }

    /// Update Pomodoro Live Activity state
    func updatePomodoroActivity(
        secondsRemaining: Int,
        isBreak: Bool,
        isPaused: Bool,
        currentSession: Int = 1
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = pomodoroActivity else { return }

        let updatedState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: secondsRemaining,
            isBreak: isBreak,
            isPaused: isPaused,
            currentSessionNumber: currentSession
        )

        // Shorter stale time when paused
        let staleDate = isPaused
            ? Date().addingTimeInterval(60) // 1 minute when paused
            : Date().addingTimeInterval(3)  // 3 seconds when active

        Task {
            await activity.update(
                ActivityContent(state: updatedState, staleDate: staleDate)
            )
        }
        #endif
    }

    /// Update Pomodoro with new session count (e.g., when completing a session)
    @available(iOS 16.2, *)
    func updatePomodoroWithAlert(
        secondsRemaining: Int,
        isBreak: Bool,
        isPaused: Bool,
        sessionsCompleted: Int,
        currentSession: Int,
        alertTitle: String,
        alertBody: String
    ) {
        #if os(iOS)
        guard #available(iOS 16.2, *) else { return }
        guard let activity = pomodoroActivity else { return }

        let updatedState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: secondsRemaining,
            isBreak: isBreak,
            isPaused: isPaused,
            currentSessionNumber: currentSession
        )

        let alertConfig = AlertConfiguration(
            title: LocalizedStringResource(stringLiteral: alertTitle),
            body: LocalizedStringResource(stringLiteral: alertBody),
            sound: .default
        )

        Task {
            await activity.update(
                ActivityContent(state: updatedState, staleDate: Date().addingTimeInterval(3)),
                alertConfiguration: alertConfig
            )
        }
        #endif
    }

    /// End the Pomodoro Live Activity
    func endPomodoroActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        guard let activity = pomodoroActivity else { return }

        let finalState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: 0,
            isBreak: false,
            isPaused: false,
            currentSessionNumber: 0
        )

        Task {
            await activity.end(
                ActivityContent(state: finalState, staleDate: nil),
                dismissalPolicy: .default
            )
            print("[LiveActivity] Ended pomodoro activity")
        }

        pomodoroActivity = nil
        #endif
    }

    /// End Pomodoro with completion alert
    @available(iOS 16.2, *)
    func endPomodoroWithAlert(sessionsCompleted: Int, totalFocusMinutes: Int) {
        #if os(iOS)
        guard #available(iOS 16.2, *) else { return }
        guard let activity = pomodoroActivity else { return }

        let finalState = PomodoroActivityAttributes.ContentState(
            secondsRemaining: 0,
            isBreak: false,
            isPaused: false,
            currentSessionNumber: 0
        )

        Task {
            await activity.end(
                ActivityContent(state: finalState, staleDate: nil),
                dismissalPolicy: .after(Date().addingTimeInterval(10)) // Linger for 10 seconds
            )
        }

        pomodoroActivity = nil
        #endif
    }

    // MARK: - Check Authorization

    /// Check if Live Activities are enabled
    func areActivitiesEnabled() -> Bool {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return false }
        return ActivityAuthorizationInfo().areActivitiesEnabled
        #else
        return false
        #endif
    }

    /// Check if there's an active tracking session
    var hasActiveTrackingActivity: Bool {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return false }
        return trackingActivity != nil
        #else
        return false
        #endif
    }

    /// Check if there's an active pomodoro session
    var hasActivePomodoroActivity: Bool {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return false }
        return pomodoroActivity != nil
        #else
        return false
        #endif
    }

    // MARK: - Cleanup

    /// End all active Live Activities
    func endAllActivities() {
        endTrackingActivity()
        endPomodoroActivity()
    }

    /// Clean up any orphaned activities from previous sessions
    func cleanupOrphanedActivities() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        // Clean up tracking activities
        for activity in Activity<TrackingActivityAttributes>.activities {
            Task {
                await activity.end(
                    ActivityContent(
                        state: TrackingActivityAttributes.ContentState(
                            elapsedSeconds: 0,
                            isActive: false,
                            billableAmount: nil
                        ),
                        staleDate: nil
                    ),
                    dismissalPolicy: .immediate
                )
            }
        }

        // Clean up pomodoro activities
        for activity in Activity<PomodoroActivityAttributes>.activities {
            Task {
                await activity.end(
                    ActivityContent(
                        state: PomodoroActivityAttributes.ContentState(
                            secondsRemaining: 0,
                            isBreak: false,
                            isPaused: false,
                            currentSessionNumber: 0
                        ),
                        staleDate: nil
                    ),
                    dismissalPolicy: .immediate
                )
            }
        }

        print("[LiveActivity] Cleaned up orphaned activities")
        #endif
    }
}

// MARK: - Live Activity URL Handling
extension LiveActivityService {
    /// Handle deep links from Live Activity taps
    func handleActivityURL(_ url: URL) -> Bool {
        guard let host = url.host else { return false }

        switch host {
        case "tracking":
            // Open tracking view
            NotificationCenter.default.post(name: .openTracking, object: nil)
            return true
        case "pomodoro":
            // Open pomodoro view
            NotificationCenter.default.post(name: .openPomodoro, object: nil)
            return true
        default:
            return false
        }
    }
}

// MARK: - Notification Names
extension Notification.Name {
    static let openTracking = Notification.Name("openTracking")
    static let openPomodoro = Notification.Name("openPomodoro")
}
