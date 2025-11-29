//
//  WidgetDataTests.swift
//  ungTests
//
//  Fast unit tests for WidgetData computed properties
//

import XCTest
@testable import ung

final class WidgetDataTests: XCTestCase {

    // MARK: - WidgetData Keys Tests

    func testWidgetDataKeysExist() {
        XCTAssertEqual(WidgetDataKeys.appGroupIdentifier, "group.com.ung.app")
        XCTAssertEqual(WidgetDataKeys.isTracking, "widget_isTracking")
        XCTAssertEqual(WidgetDataKeys.trackingProject, "widget_trackingProject")
        XCTAssertEqual(WidgetDataKeys.todayHours, "widget_todayHours")
        XCTAssertEqual(WidgetDataKeys.weeklyHours, "widget_weeklyHours")
        XCTAssertEqual(WidgetDataKeys.weeklyTarget, "widget_weeklyTarget")
        XCTAssertEqual(WidgetDataKeys.pendingInvoices, "widget_pendingInvoices")
        XCTAssertEqual(WidgetDataKeys.pomodoroActive, "widget_pomodoroActive")
        XCTAssertEqual(WidgetDataKeys.lastUpdated, "widget_lastUpdated")
    }

    // MARK: - WidgetData Default Values

    func testWidgetDataDefaultValues() {
        let data = WidgetData()

        XCTAssertFalse(data.isTracking)
        XCTAssertEqual(data.trackingProject, "")
        XCTAssertNil(data.trackingStartTime)
        XCTAssertEqual(data.trackingClient, "")
        XCTAssertEqual(data.todayHours, 0)
        XCTAssertEqual(data.weeklyHours, 0)
        XCTAssertEqual(data.weeklyTarget, 40)
        XCTAssertEqual(data.pendingInvoices, 0)
        XCTAssertEqual(data.pendingAmount, 0)
        XCTAssertFalse(data.pomodoroActive)
        XCTAssertFalse(data.pomodoroIsBreak)
        XCTAssertEqual(data.pomodoroSecondsRemaining, 0)
        XCTAssertEqual(data.pomodoroSessionsCompleted, 0)
    }

    // MARK: - Tracking Duration Tests

    func testTrackingDurationWithNoStartTime() {
        let data = WidgetData()
        XCTAssertEqual(data.trackingDuration, "00:00:00")
    }

    func testTrackingDurationCalculation() {
        var data = WidgetData()
        data.trackingStartTime = Date().addingTimeInterval(-3665) // 1h 1m 5s ago

        let duration = data.trackingDuration
        // Should be approximately "1:01:05" but may vary slightly
        XCTAssertTrue(duration.hasPrefix("1:01:"))
    }

    // MARK: - Weekly Progress Tests

    func testWeeklyProgressZeroTarget() {
        var data = WidgetData()
        data.weeklyTarget = 0
        data.weeklyHours = 10

        XCTAssertEqual(data.weeklyProgress, 0)
    }

    func testWeeklyProgressNormal() {
        var data = WidgetData()
        data.weeklyTarget = 40
        data.weeklyHours = 20

        XCTAssertEqual(data.weeklyProgress, 0.5, accuracy: 0.01)
    }

    func testWeeklyProgressExceeds100() {
        var data = WidgetData()
        data.weeklyTarget = 40
        data.weeklyHours = 50

        // Should cap at 1.0
        XCTAssertEqual(data.weeklyProgress, 1.0, accuracy: 0.01)
    }

    func testWeeklyProgressZeroHours() {
        var data = WidgetData()
        data.weeklyTarget = 40
        data.weeklyHours = 0

        XCTAssertEqual(data.weeklyProgress, 0)
    }

    // MARK: - Pomodoro Time Formatted Tests

    func testPomodoroTimeFormattedZero() {
        let data = WidgetData()
        XCTAssertEqual(data.pomodoroTimeFormatted, "00:00")
    }

    func testPomodoroTimeFormatted25Minutes() {
        var data = WidgetData()
        data.pomodoroSecondsRemaining = 25 * 60

        XCTAssertEqual(data.pomodoroTimeFormatted, "25:00")
    }

    func testPomodoroTimeFormattedWithSeconds() {
        var data = WidgetData()
        data.pomodoroSecondsRemaining = 5 * 60 + 30 // 5:30

        XCTAssertEqual(data.pomodoroTimeFormatted, "05:30")
    }

    func testPomodoroTimeFormattedLessThanMinute() {
        var data = WidgetData()
        data.pomodoroSecondsRemaining = 45

        XCTAssertEqual(data.pomodoroTimeFormatted, "00:45")
    }

    // MARK: - Formatted Pending Amount Tests

    func testFormattedPendingAmountZero() {
        let data = WidgetData()
        let formatted = data.formattedPendingAmount

        XCTAssertTrue(formatted.contains("0"))
    }

    func testFormattedPendingAmount() {
        var data = WidgetData()
        data.pendingAmount = 1234.56

        let formatted = data.formattedPendingAmount
        // Should contain currency symbol and amount
        XCTAssertTrue(formatted.contains("1,234") || formatted.contains("1234"))
    }

    func testFormattedPendingAmountNegative() {
        var data = WidgetData()
        data.pendingAmount = -500

        let formatted = data.formattedPendingAmount
        XCTAssertTrue(formatted.contains("500"))
    }

    // MARK: - Combined State Tests

    func testTrackingState() {
        var data = WidgetData()
        data.isTracking = true
        data.trackingProject = "iOS Development"
        data.trackingClient = "Acme Corp"
        data.trackingStartTime = Date().addingTimeInterval(-1800) // 30 minutes ago

        XCTAssertTrue(data.isTracking)
        XCTAssertEqual(data.trackingProject, "iOS Development")
        XCTAssertEqual(data.trackingClient, "Acme Corp")
        XCTAssertNotNil(data.trackingStartTime)
    }

    func testPomodoroState() {
        var data = WidgetData()
        data.pomodoroActive = true
        data.pomodoroIsBreak = false
        data.pomodoroSecondsRemaining = 15 * 60
        data.pomodoroSessionsCompleted = 2

        XCTAssertTrue(data.pomodoroActive)
        XCTAssertFalse(data.pomodoroIsBreak)
        XCTAssertEqual(data.pomodoroSecondsRemaining, 900)
        XCTAssertEqual(data.pomodoroSessionsCompleted, 2)
        XCTAssertEqual(data.pomodoroTimeFormatted, "15:00")
    }

    func testBreakState() {
        var data = WidgetData()
        data.pomodoroActive = true
        data.pomodoroIsBreak = true
        data.pomodoroSecondsRemaining = 5 * 60

        XCTAssertTrue(data.pomodoroActive)
        XCTAssertTrue(data.pomodoroIsBreak)
        XCTAssertEqual(data.pomodoroTimeFormatted, "05:00")
    }

    // MARK: - Stats State Tests

    func testStatsState() {
        var data = WidgetData()
        data.todayHours = 6.5
        data.weeklyHours = 32
        data.weeklyTarget = 40
        data.pendingInvoices = 3
        data.pendingAmount = 5000

        XCTAssertEqual(data.todayHours, 6.5)
        XCTAssertEqual(data.weeklyHours, 32)
        XCTAssertEqual(data.weeklyTarget, 40)
        XCTAssertEqual(data.weeklyProgress, 0.8, accuracy: 0.01)
        XCTAssertEqual(data.pendingInvoices, 3)
        XCTAssertEqual(data.pendingAmount, 5000)
    }
}
