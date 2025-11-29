//
//  AppStateTests.swift
//  ungTests
//
//  Fast unit tests for AppState models and logic
//

import XCTest
@testable import ung

final class AppStateTests: XCTestCase {

    // MARK: - PomodoroState Tests

    func testPomodoroStateDefaultValues() {
        let state = PomodoroState()

        XCTAssertFalse(state.isActive)
        XCTAssertFalse(state.isBreak)
        XCTAssertFalse(state.isPaused)
        XCTAssertEqual(state.workMinutes, 25)
        XCTAssertEqual(state.breakMinutes, 5)
        XCTAssertEqual(state.longBreakMinutes, 15)
        XCTAssertEqual(state.sessionsUntilLongBreak, 4)
        XCTAssertEqual(state.sessionsCompleted, 0)
    }

    func testPomodoroStateFormattedTime() {
        var state = PomodoroState()

        state.secondsRemaining = 1505 // 25:05
        XCTAssertEqual(state.formattedTime, "25:05")

        state.secondsRemaining = 65 // 01:05
        XCTAssertEqual(state.formattedTime, "01:05")

        state.secondsRemaining = 0
        XCTAssertEqual(state.formattedTime, "00:00")
    }

    func testPomodoroStateProgress() {
        var state = PomodoroState()
        state.workMinutes = 25
        state.secondsRemaining = 25 * 60 // Full time remaining

        // At start, progress should be 0
        XCTAssertEqual(state.progress, 0.0, accuracy: 0.01)

        // Halfway through
        state.secondsRemaining = (25 * 60) / 2
        XCTAssertEqual(state.progress, 0.5, accuracy: 0.01)

        // At end
        state.secondsRemaining = 0
        XCTAssertEqual(state.progress, 1.0, accuracy: 0.01)
    }

    func testPomodoroStateProgressDuringBreak() {
        var state = PomodoroState()
        state.isBreak = true
        state.breakMinutes = 5
        state.secondsRemaining = 5 * 60

        XCTAssertEqual(state.progress, 0.0, accuracy: 0.01)

        state.secondsRemaining = 0
        XCTAssertEqual(state.progress, 1.0, accuracy: 0.01)
    }

    func testPomodoroStateStatusText() {
        var state = PomodoroState()

        state.isActive = false
        XCTAssertEqual(state.statusText, "Ready to focus")

        state.isActive = true
        state.isPaused = true
        XCTAssertEqual(state.statusText, "Paused")

        state.isPaused = false
        state.isBreak = false
        XCTAssertEqual(state.statusText, "Focus Time")

        state.isBreak = true
        state.sessionsCompleted = 1
        XCTAssertEqual(state.statusText, "Short Break")

        state.sessionsCompleted = 4 // Multiple of sessionsUntilLongBreak
        XCTAssertEqual(state.statusText, "Long Break")
    }

    // MARK: - ActiveSession Tests

    func testActiveSessionFormattedDuration() {
        let session = ActiveSession(
            id: 1,
            project: "Test",
            client: "Client",
            startTime: Date(),
            elapsedSeconds: 3665 // 1h 1m 5s
        )

        XCTAssertEqual(session.formattedDuration, "1:01:05")
    }

    func testActiveSessionFormattedDurationZero() {
        let session = ActiveSession(
            id: 1,
            project: "Test",
            client: "Client",
            startTime: Date(),
            elapsedSeconds: 0
        )

        XCTAssertEqual(session.formattedDuration, "0:00:00")
    }

    func testActiveSessionFormattedDurationLarge() {
        let session = ActiveSession(
            id: 1,
            project: "Test",
            client: "Client",
            startTime: Date(),
            elapsedSeconds: 36000 // 10 hours
        )

        XCTAssertEqual(session.formattedDuration, "10:00:00")
    }

    // MARK: - DashboardMetrics Tests

    func testDashboardMetricsDefaultValues() {
        let metrics = DashboardMetrics()

        XCTAssertEqual(metrics.totalRevenue, 0)
        XCTAssertEqual(metrics.monthlyRevenue, 0)
        XCTAssertEqual(metrics.pendingAmount, 0)
        XCTAssertEqual(metrics.overdueAmount, 0)
        XCTAssertEqual(metrics.weeklyHours, 0)
        XCTAssertEqual(metrics.weeklyTarget, 40)
        XCTAssertEqual(metrics.trackingStreak, 0)
    }

    // MARK: - SetupStatus Tests

    func testSetupStatusIsComplete() {
        var status = SetupStatus()

        XCTAssertFalse(status.isComplete)

        status.hasCompany = true
        XCTAssertFalse(status.isComplete)

        status.hasClient = true
        XCTAssertFalse(status.isComplete)

        status.hasContract = true
        XCTAssertTrue(status.isComplete)
    }

    func testSetupStatusNextStep() {
        var status = SetupStatus()

        XCTAssertEqual(status.nextStep, "Create company profile")

        status.hasCompany = true
        XCTAssertEqual(status.nextStep, "Add your first client")

        status.hasClient = true
        XCTAssertEqual(status.nextStep, "Create a contract")

        status.hasContract = true
        XCTAssertEqual(status.nextStep, "Setup complete!")
    }

    // MARK: - SidebarTab Tests

    func testSidebarTabIcons() {
        XCTAssertEqual(SidebarTab.dashboard.icon, "square.grid.2x2")
        XCTAssertEqual(SidebarTab.tracking.icon, "clock")
        XCTAssertEqual(SidebarTab.clients.icon, "person.2")
        XCTAssertEqual(SidebarTab.contracts.icon, "doc.text")
        XCTAssertEqual(SidebarTab.invoices.icon, "doc.plaintext")
        XCTAssertEqual(SidebarTab.expenses.icon, "dollarsign.circle")
        XCTAssertEqual(SidebarTab.pomodoro.icon, "brain.head.profile")
        XCTAssertEqual(SidebarTab.reports.icon, "chart.bar")
        XCTAssertEqual(SidebarTab.settings.icon, "gearshape")
    }

    func testSidebarTabFilledIcons() {
        XCTAssertEqual(SidebarTab.dashboard.iconFilled, "square.grid.2x2.fill")
        XCTAssertEqual(SidebarTab.tracking.iconFilled, "clock.fill")
        XCTAssertEqual(SidebarTab.clients.iconFilled, "person.2.fill")
    }

    func testSidebarTabShortLabels() {
        XCTAssertEqual(SidebarTab.dashboard.shortLabel, "Home")
        XCTAssertEqual(SidebarTab.tracking.shortLabel, "Track")
        XCTAssertEqual(SidebarTab.pomodoro.shortLabel, "Focus")
    }

    func testSidebarTabRawValues() {
        XCTAssertEqual(SidebarTab.dashboard.rawValue, "Dashboard")
        XCTAssertEqual(SidebarTab.tracking.rawValue, "Time Tracking")
        XCTAssertEqual(SidebarTab.pomodoro.rawValue, "Focus Timer")
    }

    func testSidebarTabCaseIterable() {
        XCTAssertEqual(SidebarTab.allCases.count, 9)
    }

    // MARK: - Client Tests

    func testClientEquatable() {
        let client1 = Client(id: 1, name: "Test", email: "test@test.com")
        let client2 = Client(id: 1, name: "Test", email: "test@test.com")
        let client3 = Client(id: 2, name: "Test", email: "test@test.com")

        XCTAssertEqual(client1, client2)
        XCTAssertNotEqual(client1, client3)
    }

    func testClientHashable() {
        let client1 = Client(id: 1, name: "Test", email: "test@test.com")
        let client2 = Client(id: 1, name: "Test", email: "test@test.com")

        var set = Set<Client>()
        set.insert(client1)
        set.insert(client2)

        XCTAssertEqual(set.count, 1)
    }

    // MARK: - Contract Tests

    func testContractDefaultValues() {
        let contract = Contract(id: 1, name: "Test", clientName: "Client")

        XCTAssertEqual(contract.rate, 0)
        XCTAssertEqual(contract.price, 0)
        XCTAssertEqual(contract.type, "hourly")
        XCTAssertEqual(contract.currency, "USD")
        XCTAssertEqual(contract.notes, "")
    }

    // MARK: - RecentExpense Tests

    func testRecentExpenseDefaultDate() {
        let expense = RecentExpense(id: 1, description: "Test", amount: "$100", category: "Software")
        XCTAssertEqual(expense.date, "")
    }

    // MARK: - ToastType Tests

    func testToastTypes() {
        let types: [AppState.ToastType] = [.success, .info, .warning, .error]
        XCTAssertEqual(types.count, 4)
    }

    // MARK: - AppStatus Tests

    func testAppStatusEquatable() {
        XCTAssertEqual(AppStatus.loading, AppStatus.loading)
        XCTAssertEqual(AppStatus.ready, AppStatus.ready)
        XCTAssertEqual(AppStatus.notInitialized, AppStatus.notInitialized)
        XCTAssertNotEqual(AppStatus.loading, AppStatus.ready)
    }
}
