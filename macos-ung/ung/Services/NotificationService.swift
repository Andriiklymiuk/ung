//
//  NotificationService.swift
//  ung
//
//  Comprehensive notification management for the app
//

import Foundation
import UserNotifications
#if os(iOS)
import UIKit
#elseif os(macOS)
import AppKit
#endif

// MARK: - Notification Types

enum NotificationType: String {
    case pomodoroWorkComplete = "pomodoro.work.complete"
    case pomodoroBreakComplete = "pomodoro.break.complete"
    case pomodoroLongBreakComplete = "pomodoro.longbreak.complete"
    case trackingReminder = "tracking.reminder"
    case trackingLongSession = "tracking.longsession"
    case invoiceDueSoon = "invoice.due.soon"
    case invoiceOverdue = "invoice.overdue"
    case recurringInvoiceGenerated = "recurring.generated"
    case weeklyGoalProgress = "weekly.goal.progress"
    case weeklyGoalAchieved = "weekly.goal.achieved"
    case dailySummary = "daily.summary"
}

// MARK: - Notification Settings

struct NotificationSettings: Codable {
    var pomodoroEnabled: Bool = true
    var pomodoroSound: Bool = true
    var trackingRemindersEnabled: Bool = true
    var trackingReminderInterval: Int = 60 // minutes
    var longSessionAlertMinutes: Int = 120 // 2 hours
    var invoiceRemindersEnabled: Bool = true
    var invoiceDueDaysBefore: Int = 3
    var weeklyGoalRemindersEnabled: Bool = true
    var dailySummaryEnabled: Bool = false
    var dailySummaryTime: Date = Calendar.current.date(from: DateComponents(hour: 18, minute: 0)) ?? Date()
    var quietHoursEnabled: Bool = false
    var quietHoursStart: Date = Calendar.current.date(from: DateComponents(hour: 22, minute: 0)) ?? Date()
    var quietHoursEnd: Date = Calendar.current.date(from: DateComponents(hour: 8, minute: 0)) ?? Date()
}

// MARK: - Notification Service

@MainActor
class NotificationService: ObservableObject {
    static let shared = NotificationService()

    @Published var isAuthorized = false
    @Published var settings = NotificationSettings()

    private let center = UNUserNotificationCenter.current()
    private let settingsKey = "notification_settings"

    init() {
        loadSettings()
        checkAuthorizationStatus()
    }

    // MARK: - Authorization

    func requestAuthorization() async -> Bool {
        do {
            let granted = try await center.requestAuthorization(options: [.alert, .sound, .badge])
            await MainActor.run {
                self.isAuthorized = granted
            }
            if granted {
                await registerCategories()
            }
            return granted
        } catch {
            print("Notification authorization error: \(error)")
            return false
        }
    }

    func checkAuthorizationStatus() {
        center.getNotificationSettings { [weak self] settings in
            DispatchQueue.main.async {
                self?.isAuthorized = settings.authorizationStatus == .authorized
            }
        }
    }

    // MARK: - Register Categories & Actions

    private func registerCategories() async {
        // Pomodoro actions
        let startBreakAction = UNNotificationAction(
            identifier: "START_BREAK",
            title: "Start Break",
            options: .foreground
        )
        let skipBreakAction = UNNotificationAction(
            identifier: "SKIP_BREAK",
            title: "Skip Break",
            options: []
        )
        let startWorkAction = UNNotificationAction(
            identifier: "START_WORK",
            title: "Start Focus",
            options: .foreground
        )

        let pomodoroCategory = UNNotificationCategory(
            identifier: "POMODORO",
            actions: [startBreakAction, skipBreakAction, startWorkAction],
            intentIdentifiers: [],
            options: .customDismissAction
        )

        // Tracking actions
        let stopTrackingAction = UNNotificationAction(
            identifier: "STOP_TRACKING",
            title: "Stop Tracking",
            options: .foreground
        )
        let continueTrackingAction = UNNotificationAction(
            identifier: "CONTINUE_TRACKING",
            title: "Continue",
            options: []
        )

        let trackingCategory = UNNotificationCategory(
            identifier: "TRACKING",
            actions: [stopTrackingAction, continueTrackingAction],
            intentIdentifiers: [],
            options: .customDismissAction
        )

        // Invoice actions
        let viewInvoiceAction = UNNotificationAction(
            identifier: "VIEW_INVOICE",
            title: "View Invoice",
            options: .foreground
        )
        let markPaidAction = UNNotificationAction(
            identifier: "MARK_PAID",
            title: "Mark as Paid",
            options: []
        )

        let invoiceCategory = UNNotificationCategory(
            identifier: "INVOICE",
            actions: [viewInvoiceAction, markPaidAction],
            intentIdentifiers: [],
            options: .customDismissAction
        )

        center.setNotificationCategories([pomodoroCategory, trackingCategory, invoiceCategory])
    }

    // MARK: - Pomodoro Notifications

    func schedulePomodoroComplete(isBreak: Bool, isLongBreak: Bool = false, delay: TimeInterval) {
        guard settings.pomodoroEnabled else { return }
        guard !isInQuietHours() else { return }

        let content = UNMutableNotificationContent()

        if isBreak {
            content.title = isLongBreak ? "Long Break Complete! 🎉" : "Break Over!"
            content.body = "Time to get back to focused work."
            content.sound = settings.pomodoroSound ? .default : nil
        } else {
            content.title = "Focus Session Complete! 🎯"
            content.body = "Great work! Take a well-deserved break."
            content.sound = settings.pomodoroSound ? .default : nil
        }

        content.categoryIdentifier = "POMODORO"
        content.userInfo = ["type": isBreak ? NotificationType.pomodoroBreakComplete.rawValue : NotificationType.pomodoroWorkComplete.rawValue]

        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: delay, repeats: false)
        let request = UNNotificationRequest(
            identifier: "pomodoro_\(UUID().uuidString)",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    func cancelPomodoroNotifications() {
        center.removePendingNotificationRequests(withIdentifiers: ["pomodoro"])
        // Also remove all pending pomodoro notifications
        center.getPendingNotificationRequests { requests in
            let pomodoroIds = requests.filter { $0.identifier.hasPrefix("pomodoro_") }.map { $0.identifier }
            self.center.removePendingNotificationRequests(withIdentifiers: pomodoroIds)
        }
    }

    // MARK: - Tracking Notifications

    func scheduleTrackingReminder() {
        guard settings.trackingRemindersEnabled else { return }
        guard !isInQuietHours() else { return }

        // Cancel existing reminder
        center.removePendingNotificationRequests(withIdentifiers: ["tracking_reminder"])

        let content = UNMutableNotificationContent()
        content.title = "Start Tracking? ⏱️"
        content.body = "Don't forget to track your work time today."
        content.sound = .default
        content.categoryIdentifier = "TRACKING"

        // Schedule for the configured interval
        let trigger = UNTimeIntervalNotificationTrigger(
            timeInterval: TimeInterval(settings.trackingReminderInterval * 60),
            repeats: false
        )

        let request = UNNotificationRequest(
            identifier: "tracking_reminder",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    func scheduleLongSessionAlert(projectName: String, elapsedMinutes: Int) {
        guard settings.trackingRemindersEnabled else { return }
        guard elapsedMinutes >= settings.longSessionAlertMinutes else { return }
        guard !isInQuietHours() else { return }

        let content = UNMutableNotificationContent()
        content.title = "Long Session Alert 💪"
        content.body = "You've been tracking '\(projectName)' for \(elapsedMinutes / 60) hours. Take a break?"
        content.sound = .default
        content.categoryIdentifier = "TRACKING"
        content.userInfo = ["type": NotificationType.trackingLongSession.rawValue]

        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 1, repeats: false)
        let request = UNNotificationRequest(
            identifier: "long_session_\(UUID().uuidString)",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    func cancelTrackingReminders() {
        center.removePendingNotificationRequests(withIdentifiers: ["tracking_reminder"])
    }

    // MARK: - Invoice Notifications

    func scheduleInvoiceDueReminder(invoiceNum: String, clientName: String, amount: String, dueDate: Date) {
        guard settings.invoiceRemindersEnabled else { return }

        let daysBefore = settings.invoiceDueDaysBefore
        guard let reminderDate = Calendar.current.date(byAdding: .day, value: -daysBefore, to: dueDate) else { return }

        // Don't schedule if reminder date is in the past
        guard reminderDate > Date() else { return }

        let content = UNMutableNotificationContent()
        content.title = "Invoice Due Soon 📄"
        content.body = "\(invoiceNum) for \(clientName) (\(amount)) is due in \(daysBefore) days."
        content.sound = .default
        content.categoryIdentifier = "INVOICE"
        content.userInfo = [
            "type": NotificationType.invoiceDueSoon.rawValue,
            "invoiceNum": invoiceNum
        ]

        let dateComponents = Calendar.current.dateComponents([.year, .month, .day, .hour, .minute], from: reminderDate)
        let trigger = UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: false)

        let request = UNNotificationRequest(
            identifier: "invoice_due_\(invoiceNum)",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    func notifyInvoiceOverdue(invoiceNum: String, clientName: String, amount: String) {
        guard settings.invoiceRemindersEnabled else { return }

        let content = UNMutableNotificationContent()
        content.title = "Invoice Overdue ⚠️"
        content.body = "\(invoiceNum) for \(clientName) (\(amount)) is now overdue."
        content.sound = .default
        content.categoryIdentifier = "INVOICE"
        content.userInfo = [
            "type": NotificationType.invoiceOverdue.rawValue,
            "invoiceNum": invoiceNum
        ]

        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 1, repeats: false)
        let request = UNNotificationRequest(
            identifier: "invoice_overdue_\(invoiceNum)",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    func notifyRecurringInvoiceGenerated(count: Int) {
        guard settings.invoiceRemindersEnabled else { return }

        let content = UNMutableNotificationContent()
        content.title = "Recurring Invoices Generated 🔄"
        content.body = "\(count) invoice\(count == 1 ? "" : "s") generated from recurring templates."
        content.sound = .default
        content.userInfo = ["type": NotificationType.recurringInvoiceGenerated.rawValue]

        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 1, repeats: false)
        let request = UNNotificationRequest(
            identifier: "recurring_\(UUID().uuidString)",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    // MARK: - Weekly Goal Notifications

    func notifyWeeklyGoalProgress(currentHours: Double, targetHours: Double) {
        guard settings.weeklyGoalRemindersEnabled else { return }

        let progress = currentHours / targetHours
        let content = UNMutableNotificationContent()

        if progress >= 1.0 {
            content.title = "Weekly Goal Achieved! 🏆"
            content.body = "You've reached \(String(format: "%.1f", currentHours))h of your \(String(format: "%.0f", targetHours))h goal!"
            content.userInfo = ["type": NotificationType.weeklyGoalAchieved.rawValue]
        } else if progress >= 0.75 {
            content.title = "Almost There! 💪"
            content.body = "\(String(format: "%.1f", currentHours))h tracked - just \(String(format: "%.1f", targetHours - currentHours))h to reach your weekly goal."
            content.userInfo = ["type": NotificationType.weeklyGoalProgress.rawValue]
        } else {
            return // Don't notify for lower progress
        }

        content.sound = .default

        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 1, repeats: false)
        let request = UNNotificationRequest(
            identifier: "weekly_goal_\(UUID().uuidString)",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    // MARK: - Daily Summary

    func scheduleDailySummary(hoursTracked: Double, sessionsCompleted: Int) {
        guard settings.dailySummaryEnabled else { return }

        let content = UNMutableNotificationContent()
        content.title = "Daily Summary 📊"
        content.body = "Today: \(String(format: "%.1f", hoursTracked))h tracked across \(sessionsCompleted) session\(sessionsCompleted == 1 ? "" : "s")."
        content.sound = .default
        content.userInfo = ["type": NotificationType.dailySummary.rawValue]

        // Schedule for the configured time
        var dateComponents = Calendar.current.dateComponents([.hour, .minute], from: settings.dailySummaryTime)
        dateComponents.second = 0

        let trigger = UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)
        let request = UNNotificationRequest(
            identifier: "daily_summary",
            content: content,
            trigger: trigger
        )

        center.add(request)
    }

    func cancelDailySummary() {
        center.removePendingNotificationRequests(withIdentifiers: ["daily_summary"])
    }

    // MARK: - Quiet Hours

    private func isInQuietHours() -> Bool {
        guard settings.quietHoursEnabled else { return false }

        let now = Date()
        let calendar = Calendar.current
        let currentHour = calendar.component(.hour, from: now)
        let currentMinute = calendar.component(.minute, from: now)
        let currentTime = currentHour * 60 + currentMinute

        let startHour = calendar.component(.hour, from: settings.quietHoursStart)
        let startMinute = calendar.component(.minute, from: settings.quietHoursStart)
        let startTime = startHour * 60 + startMinute

        let endHour = calendar.component(.hour, from: settings.quietHoursEnd)
        let endMinute = calendar.component(.minute, from: settings.quietHoursEnd)
        let endTime = endHour * 60 + endMinute

        if startTime < endTime {
            // Same day: e.g., 9:00 to 17:00
            return currentTime >= startTime && currentTime < endTime
        } else {
            // Overnight: e.g., 22:00 to 8:00
            return currentTime >= startTime || currentTime < endTime
        }
    }

    // MARK: - Settings Persistence

    func saveSettings() {
        if let data = try? JSONEncoder().encode(settings) {
            UserDefaults.standard.set(data, forKey: settingsKey)
        }
    }

    private func loadSettings() {
        if let data = UserDefaults.standard.data(forKey: settingsKey),
           let decoded = try? JSONDecoder().decode(NotificationSettings.self, from: data) {
            settings = decoded
        }
    }

    // MARK: - Clear All

    func clearAllNotifications() {
        center.removeAllPendingNotificationRequests()
        center.removeAllDeliveredNotifications()
    }

    func clearDeliveredNotifications() {
        center.removeAllDeliveredNotifications()
    }

    // MARK: - Badge Management

    func updateBadge(count: Int) {
        #if os(iOS)
        UNUserNotificationCenter.current().setBadgeCount(count)
        #elseif os(macOS)
        NSApplication.shared.dockTile.badgeLabel = count > 0 ? "\(count)" : nil
        #endif
    }

    func clearBadge() {
        updateBadge(count: 0)
    }
}

// MARK: - Notification Delegate

class NotificationDelegate: NSObject, UNUserNotificationCenterDelegate {
    static let shared = NotificationDelegate()

    var onPomodoroAction: ((String) -> Void)?
    var onTrackingAction: ((String) -> Void)?
    var onInvoiceAction: ((String, String) -> Void)?

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        // Show notification even when app is in foreground
        completionHandler([.banner, .sound, .badge])
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let actionIdentifier = response.actionIdentifier
        let userInfo = response.notification.request.content.userInfo

        switch actionIdentifier {
        case "START_BREAK", "SKIP_BREAK", "START_WORK":
            onPomodoroAction?(actionIdentifier)

        case "STOP_TRACKING", "CONTINUE_TRACKING":
            onTrackingAction?(actionIdentifier)

        case "VIEW_INVOICE", "MARK_PAID":
            if let invoiceNum = userInfo["invoiceNum"] as? String {
                onInvoiceAction?(actionIdentifier, invoiceNum)
            }

        case UNNotificationDefaultActionIdentifier:
            // User tapped the notification itself
            if let type = userInfo["type"] as? String {
                handleNotificationTap(type: type, userInfo: userInfo)
            }

        default:
            break
        }

        completionHandler()
    }

    private func handleNotificationTap(type: String, userInfo: [AnyHashable: Any]) {
        guard let notificationType = NotificationType(rawValue: type) else { return }

        // Post notification for app to handle navigation
        NotificationCenter.default.post(
            name: .didTapNotification,
            object: nil,
            userInfo: ["type": notificationType, "info": userInfo]
        )
    }
}

// MARK: - Notification Names

extension Notification.Name {
    static let didTapNotification = Notification.Name("didTapNotification")
}
