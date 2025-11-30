//
//  LiveActivityService.swift
//  ung
//
//  Premium Live Activities service for real-time tracking and pomodoro display
//  Enhanced with milestone celebrations, invoice tracking, break reminders
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

// MARK: - Milestone Activity Attributes
@available(iOS 16.1, *)
struct MilestoneActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var currentAmount: Double
        var isAchieved: Bool
        var celebrationPhase: Int // 0: approaching, 1: achieved, 2: exceeded
    }

    var milestoneType: String // "daily", "weekly", "monthly"
    var targetAmount: Double
    var currency: String
    var milestoneName: String
}

// MARK: - Invoice Status Activity Attributes
@available(iOS 16.1, *)
struct InvoiceStatusActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var status: String // "sent", "viewed", "paid", "overdue"
        var viewedAt: Date?
        var paidAt: Date?
        var daysUntilDue: Int
    }

    var invoiceNumber: String
    var clientName: String
    var amount: Double
    var currency: String
    var dueDate: Date
    var sentAt: Date
}

// MARK: - Daily Summary Activity Attributes
@available(iOS 16.1, *)
struct DailySummaryActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var hoursWorked: Double
        var earnings: Double
        var sessionsCompleted: Int
        var streakDays: Int
        var isNewRecord: Bool
    }

    var summaryType: String // "daily", "weekly"
    var targetHours: Double
    var currency: String
    var userName: String?
}

// MARK: - Break Reminder Activity Attributes
@available(iOS 16.1, *)
struct BreakReminderActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var minutesSinceLastBreak: Int
        var totalWorkMinutesToday: Int
        var breaksTakenToday: Int
        var urgencyLevel: Int // 0: normal, 1: suggested, 2: recommended, 3: urgent
    }

    var breakIntervalMinutes: Int
    var recommendedBreakMinutes: Int
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

    @available(iOS 16.1, *)
    private var milestoneActivity: Activity<MilestoneActivityAttributes>?

    @available(iOS 16.1, *)
    private var invoiceActivity: Activity<InvoiceStatusActivityAttributes>?

    @available(iOS 16.1, *)
    private var summaryActivity: Activity<DailySummaryActivityAttributes>?

    @available(iOS 16.1, *)
    private var breakReminderActivity: Activity<BreakReminderActivityAttributes>?
    #endif

    private var trackingUpdateTimer: Timer?
    private var pomodoroUpdateTimer: Timer?
    private var breakReminderTimer: Timer?

    // Store start time for tracking updates
    private var trackingStartTime: Date?
    private var trackingHourlyRate: Double?
    private var trackingCurrency: String = "USD"

    // Break tracking
    private var lastBreakTime: Date?
    private var totalWorkMinutesToday: Int = 0
    private var breaksTakenToday: Int = 0

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

    // MARK: - Milestone Live Activity

    /// Start a milestone tracking Live Activity
    func startMilestoneActivity(
        milestoneType: String,
        targetAmount: Double,
        currentAmount: Double,
        currency: String = "USD",
        milestoneName: String
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }

        endMilestoneActivity()

        let attributes = MilestoneActivityAttributes(
            milestoneType: milestoneType,
            targetAmount: targetAmount,
            currency: currency,
            milestoneName: milestoneName
        )

        let isAchieved = currentAmount >= targetAmount
        let celebrationPhase = currentAmount > targetAmount * 1.1 ? 2 : (isAchieved ? 1 : 0)

        let state = MilestoneActivityAttributes.ContentState(
            currentAmount: currentAmount,
            isAchieved: isAchieved,
            celebrationPhase: celebrationPhase
        )

        do {
            milestoneActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: state, staleDate: Date().addingTimeInterval(60)),
                pushType: nil
            )
            print("[LiveActivity] Started milestone activity: \(milestoneName)")
        } catch {
            print("[LiveActivity] Error starting milestone: \(error)")
        }
        #endif
    }

    /// Update milestone progress
    func updateMilestoneActivity(currentAmount: Double) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = milestoneActivity else { return }

        let targetAmount = activity.attributes.targetAmount
        let isAchieved = currentAmount >= targetAmount
        let celebrationPhase = currentAmount > targetAmount * 1.1 ? 2 : (isAchieved ? 1 : 0)

        let state = MilestoneActivityAttributes.ContentState(
            currentAmount: currentAmount,
            isAchieved: isAchieved,
            celebrationPhase: celebrationPhase
        )

        Task {
            await activity.update(ActivityContent(state: state, staleDate: Date().addingTimeInterval(60)))
        }
        #endif
    }

    /// End milestone activity
    func endMilestoneActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = milestoneActivity else { return }

        Task {
            await activity.end(
                ActivityContent(
                    state: MilestoneActivityAttributes.ContentState(
                        currentAmount: 0,
                        isAchieved: false,
                        celebrationPhase: 0
                    ),
                    staleDate: nil
                ),
                dismissalPolicy: .default
            )
        }
        milestoneActivity = nil
        #endif
    }

    // MARK: - Invoice Status Live Activity

    /// Start an invoice tracking Live Activity
    func startInvoiceActivity(
        invoiceNumber: String,
        clientName: String,
        amount: Double,
        currency: String = "USD",
        dueDate: Date,
        sentAt: Date = Date()
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }

        endInvoiceActivity()

        let attributes = InvoiceStatusActivityAttributes(
            invoiceNumber: invoiceNumber,
            clientName: clientName,
            amount: amount,
            currency: currency,
            dueDate: dueDate,
            sentAt: sentAt
        )

        let daysUntilDue = Calendar.current.dateComponents([.day], from: Date(), to: dueDate).day ?? 0

        let state = InvoiceStatusActivityAttributes.ContentState(
            status: "sent",
            viewedAt: nil,
            paidAt: nil,
            daysUntilDue: daysUntilDue
        )

        do {
            invoiceActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: state, staleDate: Date().addingTimeInterval(3600)),
                pushType: nil
            )
            print("[LiveActivity] Started invoice activity: #\(invoiceNumber)")
        } catch {
            print("[LiveActivity] Error starting invoice activity: \(error)")
        }
        #endif
    }

    /// Update invoice status
    func updateInvoiceActivity(status: String, viewedAt: Date? = nil, paidAt: Date? = nil) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = invoiceActivity else { return }

        let dueDate = activity.attributes.dueDate
        let daysUntilDue = Calendar.current.dateComponents([.day], from: Date(), to: dueDate).day ?? 0

        let state = InvoiceStatusActivityAttributes.ContentState(
            status: status,
            viewedAt: viewedAt,
            paidAt: paidAt,
            daysUntilDue: daysUntilDue
        )

        Task {
            await activity.update(ActivityContent(state: state, staleDate: Date().addingTimeInterval(3600)))

            // Auto-end if paid
            if status == "paid" {
                try? await Task.sleep(nanoseconds: 10_000_000_000) // 10 seconds
                await activity.end(
                    ActivityContent(state: state, staleDate: nil),
                    dismissalPolicy: .default
                )
            }
        }
        #endif
    }

    /// End invoice activity
    func endInvoiceActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = invoiceActivity else { return }

        Task {
            await activity.end(
                ActivityContent(
                    state: InvoiceStatusActivityAttributes.ContentState(
                        status: "sent",
                        viewedAt: nil,
                        paidAt: nil,
                        daysUntilDue: 0
                    ),
                    staleDate: nil
                ),
                dismissalPolicy: .immediate
            )
        }
        invoiceActivity = nil
        #endif
    }

    // MARK: - Daily Summary Live Activity

    /// Start a daily/weekly summary Live Activity
    func startSummaryActivity(
        summaryType: String = "daily",
        targetHours: Double,
        hoursWorked: Double,
        earnings: Double,
        sessionsCompleted: Int,
        streakDays: Int,
        isNewRecord: Bool = false,
        currency: String = "USD",
        userName: String? = nil
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }

        endSummaryActivity()

        let attributes = DailySummaryActivityAttributes(
            summaryType: summaryType,
            targetHours: targetHours,
            currency: currency,
            userName: userName
        )

        let state = DailySummaryActivityAttributes.ContentState(
            hoursWorked: hoursWorked,
            earnings: earnings,
            sessionsCompleted: sessionsCompleted,
            streakDays: streakDays,
            isNewRecord: isNewRecord
        )

        do {
            summaryActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: state, staleDate: Date().addingTimeInterval(300)),
                pushType: nil
            )
            print("[LiveActivity] Started \(summaryType) summary activity")
        } catch {
            print("[LiveActivity] Error starting summary: \(error)")
        }
        #endif
    }

    /// Update summary with new stats
    func updateSummaryActivity(
        hoursWorked: Double,
        earnings: Double,
        sessionsCompleted: Int,
        streakDays: Int,
        isNewRecord: Bool
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = summaryActivity else { return }

        let state = DailySummaryActivityAttributes.ContentState(
            hoursWorked: hoursWorked,
            earnings: earnings,
            sessionsCompleted: sessionsCompleted,
            streakDays: streakDays,
            isNewRecord: isNewRecord
        )

        Task {
            await activity.update(ActivityContent(state: state, staleDate: Date().addingTimeInterval(300)))
        }
        #endif
    }

    /// End summary activity
    func endSummaryActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = summaryActivity else { return }

        Task {
            await activity.end(
                ActivityContent(
                    state: DailySummaryActivityAttributes.ContentState(
                        hoursWorked: 0,
                        earnings: 0,
                        sessionsCompleted: 0,
                        streakDays: 0,
                        isNewRecord: false
                    ),
                    staleDate: nil
                ),
                dismissalPolicy: .default
            )
        }
        summaryActivity = nil
        #endif
    }

    // MARK: - Break Reminder Live Activity

    /// Start a break reminder Live Activity (health-focused)
    func startBreakReminderActivity(
        breakIntervalMinutes: Int = 45,
        recommendedBreakMinutes: Int = 5
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }

        endBreakReminderActivity()

        lastBreakTime = Date()

        let attributes = BreakReminderActivityAttributes(
            breakIntervalMinutes: breakIntervalMinutes,
            recommendedBreakMinutes: recommendedBreakMinutes
        )

        let state = BreakReminderActivityAttributes.ContentState(
            minutesSinceLastBreak: 0,
            totalWorkMinutesToday: totalWorkMinutesToday,
            breaksTakenToday: breaksTakenToday,
            urgencyLevel: 0
        )

        do {
            breakReminderActivity = try Activity.request(
                attributes: attributes,
                content: .init(state: state, staleDate: Date().addingTimeInterval(60)),
                pushType: nil
            )

            // Start timer to update break reminder
            startBreakReminderTimer(interval: breakIntervalMinutes)

            print("[LiveActivity] Started break reminder activity")
        } catch {
            print("[LiveActivity] Error starting break reminder: \(error)")
        }
        #endif
    }

    private func startBreakReminderTimer(interval: Int) {
        breakReminderTimer?.invalidate()

        breakReminderTimer = Timer.scheduledTimer(withTimeInterval: 60, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                self?.updateBreakReminderFromTimer(interval: interval)
            }
        }
        RunLoop.current.add(breakReminderTimer!, forMode: .common)
    }

    private func updateBreakReminderFromTimer(interval: Int) {
        guard let lastBreak = lastBreakTime else { return }

        let minutesSinceBreak = Int(Date().timeIntervalSince(lastBreak) / 60)
        totalWorkMinutesToday += 1

        // Calculate urgency level
        let urgencyLevel: Int
        if minutesSinceBreak >= interval * 2 {
            urgencyLevel = 3 // Urgent
        } else if minutesSinceBreak >= Int(Double(interval) * 1.5) {
            urgencyLevel = 2 // Recommended
        } else if minutesSinceBreak >= interval {
            urgencyLevel = 1 // Suggested
        } else {
            urgencyLevel = 0 // Normal
        }

        updateBreakReminderActivity(
            minutesSinceLastBreak: minutesSinceBreak,
            totalWorkMinutesToday: totalWorkMinutesToday,
            breaksTakenToday: breaksTakenToday,
            urgencyLevel: urgencyLevel
        )
    }

    /// Update break reminder
    func updateBreakReminderActivity(
        minutesSinceLastBreak: Int,
        totalWorkMinutesToday: Int,
        breaksTakenToday: Int,
        urgencyLevel: Int
    ) {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }
        guard let activity = breakReminderActivity else { return }

        let state = BreakReminderActivityAttributes.ContentState(
            minutesSinceLastBreak: minutesSinceLastBreak,
            totalWorkMinutesToday: totalWorkMinutesToday,
            breaksTakenToday: breaksTakenToday,
            urgencyLevel: urgencyLevel
        )

        Task {
            await activity.update(ActivityContent(state: state, staleDate: Date().addingTimeInterval(60)))
        }
        #endif
    }

    /// Record a break taken
    func recordBreakTaken() {
        lastBreakTime = Date()
        breaksTakenToday += 1

        updateBreakReminderActivity(
            minutesSinceLastBreak: 0,
            totalWorkMinutesToday: totalWorkMinutesToday,
            breaksTakenToday: breaksTakenToday,
            urgencyLevel: 0
        )
    }

    /// End break reminder activity
    func endBreakReminderActivity() {
        #if os(iOS)
        guard #available(iOS 16.1, *) else { return }

        breakReminderTimer?.invalidate()
        breakReminderTimer = nil

        guard let activity = breakReminderActivity else { return }

        Task {
            await activity.end(
                ActivityContent(
                    state: BreakReminderActivityAttributes.ContentState(
                        minutesSinceLastBreak: 0,
                        totalWorkMinutesToday: 0,
                        breaksTakenToday: 0,
                        urgencyLevel: 0
                    ),
                    staleDate: nil
                ),
                dismissalPolicy: .immediate
            )
        }
        breakReminderActivity = nil
        #endif
    }

    /// Reset daily break stats (call at start of day)
    func resetDailyBreakStats() {
        totalWorkMinutesToday = 0
        breaksTakenToday = 0
        lastBreakTime = Date()
    }

    // MARK: - Cleanup

    /// End all active Live Activities
    func endAllActivities() {
        endTrackingActivity()
        endPomodoroActivity()
        endMilestoneActivity()
        endInvoiceActivity()
        endSummaryActivity()
        endBreakReminderActivity()
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

        // Clean up milestone activities
        for activity in Activity<MilestoneActivityAttributes>.activities {
            Task {
                await activity.end(
                    ActivityContent(
                        state: MilestoneActivityAttributes.ContentState(
                            currentAmount: 0,
                            isAchieved: false,
                            celebrationPhase: 0
                        ),
                        staleDate: nil
                    ),
                    dismissalPolicy: .immediate
                )
            }
        }

        // Clean up invoice activities
        for activity in Activity<InvoiceStatusActivityAttributes>.activities {
            Task {
                await activity.end(
                    ActivityContent(
                        state: InvoiceStatusActivityAttributes.ContentState(
                            status: "sent",
                            viewedAt: nil,
                            paidAt: nil,
                            daysUntilDue: 0
                        ),
                        staleDate: nil
                    ),
                    dismissalPolicy: .immediate
                )
            }
        }

        // Clean up summary activities
        for activity in Activity<DailySummaryActivityAttributes>.activities {
            Task {
                await activity.end(
                    ActivityContent(
                        state: DailySummaryActivityAttributes.ContentState(
                            hoursWorked: 0,
                            earnings: 0,
                            sessionsCompleted: 0,
                            streakDays: 0,
                            isNewRecord: false
                        ),
                        staleDate: nil
                    ),
                    dismissalPolicy: .immediate
                )
            }
        }

        // Clean up break reminder activities
        for activity in Activity<BreakReminderActivityAttributes>.activities {
            Task {
                await activity.end(
                    ActivityContent(
                        state: BreakReminderActivityAttributes.ContentState(
                            minutesSinceLastBreak: 0,
                            totalWorkMinutesToday: 0,
                            breaksTakenToday: 0,
                            urgencyLevel: 0
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
