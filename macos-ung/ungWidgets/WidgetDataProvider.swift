//
//  WidgetDataProvider.swift
//  ungWidgets
//
//  Shared data provider for widget content
//

import Foundation
import WidgetKit

// MARK: - Shared Data Keys
enum WidgetDataKeys {
    static let appGroupIdentifier = "group.com.ung.app"
    static let isTracking = "widget_isTracking"
    static let trackingProject = "widget_trackingProject"
    static let trackingStartTime = "widget_trackingStartTime"
    static let trackingClient = "widget_trackingClient"
    static let todayHours = "widget_todayHours"
    static let weeklyHours = "widget_weeklyHours"
    static let weeklyTarget = "widget_weeklyTarget"
    static let pendingInvoices = "widget_pendingInvoices"
    static let pendingAmount = "widget_pendingAmount"
    static let pomodoroActive = "widget_pomodoroActive"
    static let pomodoroIsBreak = "widget_pomodoroIsBreak"
    static let pomodoroSecondsRemaining = "widget_pomodoroSecondsRemaining"
    static let pomodoroSessionsCompleted = "widget_pomodoroSessionsCompleted"
    static let lastUpdated = "widget_lastUpdated"
}

// MARK: - Widget Data
struct WidgetData {
    // Tracking
    var isTracking: Bool = false
    var trackingProject: String = ""
    var trackingStartTime: Date?
    var trackingClient: String = ""

    // Stats
    var todayHours: Double = 0
    var weeklyHours: Double = 0
    var weeklyTarget: Double = 40
    var pendingInvoices: Int = 0
    var pendingAmount: Double = 0

    // Pomodoro
    var pomodoroActive: Bool = false
    var pomodoroIsBreak: Bool = false
    var pomodoroSecondsRemaining: Int = 0
    var pomodoroSessionsCompleted: Int = 0

    var lastUpdated: Date = Date()

    // Computed
    var trackingDuration: String {
        guard let startTime = trackingStartTime else { return "00:00:00" }
        let elapsed = Int(Date().timeIntervalSince(startTime))
        let hours = elapsed / 3600
        let minutes = (elapsed % 3600) / 60
        let seconds = elapsed % 60
        return String(format: "%d:%02d:%02d", hours, minutes, seconds)
    }

    var weeklyProgress: Double {
        guard weeklyTarget > 0 else { return 0 }
        return min(weeklyHours / weeklyTarget, 1.0)
    }

    var pomodoroTimeFormatted: String {
        let minutes = pomodoroSecondsRemaining / 60
        let seconds = pomodoroSecondsRemaining % 60
        return String(format: "%02d:%02d", minutes, seconds)
    }

    var formattedPendingAmount: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        return formatter.string(from: NSNumber(value: pendingAmount)) ?? "$0"
    }

    // Load from UserDefaults (App Group)
    static func load() -> WidgetData {
        guard let defaults = UserDefaults(suiteName: WidgetDataKeys.appGroupIdentifier) else {
            return WidgetData()
        }

        var data = WidgetData()
        data.isTracking = defaults.bool(forKey: WidgetDataKeys.isTracking)
        data.trackingProject = defaults.string(forKey: WidgetDataKeys.trackingProject) ?? ""
        data.trackingStartTime = defaults.object(forKey: WidgetDataKeys.trackingStartTime) as? Date
        data.trackingClient = defaults.string(forKey: WidgetDataKeys.trackingClient) ?? ""
        data.todayHours = defaults.double(forKey: WidgetDataKeys.todayHours)
        data.weeklyHours = defaults.double(forKey: WidgetDataKeys.weeklyHours)
        data.weeklyTarget = defaults.double(forKey: WidgetDataKeys.weeklyTarget)
        if data.weeklyTarget == 0 { data.weeklyTarget = 40 }
        data.pendingInvoices = defaults.integer(forKey: WidgetDataKeys.pendingInvoices)
        data.pendingAmount = defaults.double(forKey: WidgetDataKeys.pendingAmount)
        data.pomodoroActive = defaults.bool(forKey: WidgetDataKeys.pomodoroActive)
        data.pomodoroIsBreak = defaults.bool(forKey: WidgetDataKeys.pomodoroIsBreak)
        data.pomodoroSecondsRemaining = defaults.integer(forKey: WidgetDataKeys.pomodoroSecondsRemaining)
        data.pomodoroSessionsCompleted = defaults.integer(forKey: WidgetDataKeys.pomodoroSessionsCompleted)
        data.lastUpdated = defaults.object(forKey: WidgetDataKeys.lastUpdated) as? Date ?? Date()

        return data
    }

    // Save to UserDefaults (App Group) - called from main app
    func save() {
        guard let defaults = UserDefaults(suiteName: WidgetDataKeys.appGroupIdentifier) else {
            return
        }

        defaults.set(isTracking, forKey: WidgetDataKeys.isTracking)
        defaults.set(trackingProject, forKey: WidgetDataKeys.trackingProject)
        defaults.set(trackingStartTime, forKey: WidgetDataKeys.trackingStartTime)
        defaults.set(trackingClient, forKey: WidgetDataKeys.trackingClient)
        defaults.set(todayHours, forKey: WidgetDataKeys.todayHours)
        defaults.set(weeklyHours, forKey: WidgetDataKeys.weeklyHours)
        defaults.set(weeklyTarget, forKey: WidgetDataKeys.weeklyTarget)
        defaults.set(pendingInvoices, forKey: WidgetDataKeys.pendingInvoices)
        defaults.set(pendingAmount, forKey: WidgetDataKeys.pendingAmount)
        defaults.set(pomodoroActive, forKey: WidgetDataKeys.pomodoroActive)
        defaults.set(pomodoroIsBreak, forKey: WidgetDataKeys.pomodoroIsBreak)
        defaults.set(pomodoroSecondsRemaining, forKey: WidgetDataKeys.pomodoroSecondsRemaining)
        defaults.set(pomodoroSessionsCompleted, forKey: WidgetDataKeys.pomodoroSessionsCompleted)
        defaults.set(Date(), forKey: WidgetDataKeys.lastUpdated)
    }

    // Trigger widget refresh
    static func refreshWidgets() {
        WidgetCenter.shared.reloadAllTimelines()
    }
}

// MARK: - Widget Helper to update from main app
class WidgetDataManager {
    static let shared = WidgetDataManager()

    func updateTrackingStatus(isTracking: Bool, project: String, client: String, startTime: Date?) {
        var data = WidgetData.load()
        data.isTracking = isTracking
        data.trackingProject = project
        data.trackingClient = client
        data.trackingStartTime = startTime
        data.save()
        WidgetData.refreshWidgets()
    }

    func updateStats(todayHours: Double, weeklyHours: Double, weeklyTarget: Double, pendingInvoices: Int, pendingAmount: Double) {
        var data = WidgetData.load()
        data.todayHours = todayHours
        data.weeklyHours = weeklyHours
        data.weeklyTarget = weeklyTarget
        data.pendingInvoices = pendingInvoices
        data.pendingAmount = pendingAmount
        data.save()
        WidgetData.refreshWidgets()
    }

    func updatePomodoro(active: Bool, isBreak: Bool, secondsRemaining: Int, sessionsCompleted: Int) {
        var data = WidgetData.load()
        data.pomodoroActive = active
        data.pomodoroIsBreak = isBreak
        data.pomodoroSecondsRemaining = secondsRemaining
        data.pomodoroSessionsCompleted = sessionsCompleted
        data.save()
        WidgetData.refreshWidgets()
    }
}
