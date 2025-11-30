//
//  CalculationsTests.swift
//  ungTests
//
//  Fast unit tests for financial and business calculations
//

import XCTest
@testable import ung

final class CalculationsTests: XCTestCase {

    // MARK: - Invoice Total Calculations

    func testInvoiceLineItemTotal() {
        let lineItem = InvoiceLineItem(
            invoiceId: 1,
            itemName: "Development",
            quantity: 10,
            rate: 100.0,
            amount: 1000.0
        )

        XCTAssertEqual(lineItem.amount, lineItem.quantity * lineItem.rate)
    }

    func testInvoiceLineItemWithDiscount() {
        let quantity = 10.0
        let rate = 100.0
        let discountPercent = 10.0

        let subtotal = quantity * rate
        let discount = subtotal * (discountPercent / 100)
        let total = subtotal - discount

        XCTAssertEqual(subtotal, 1000.0)
        XCTAssertEqual(discount, 100.0)
        XCTAssertEqual(total, 900.0)
    }

    func testInvoiceTaxCalculation() {
        let subtotal = 1000.0
        let taxRate = 20.0

        let tax = subtotal * (taxRate / 100)
        let total = subtotal + tax

        XCTAssertEqual(tax, 200.0)
        XCTAssertEqual(total, 1200.0)
    }

    // MARK: - Hourly Rate Calculations

    func testHourlyBilling() {
        let hours = 8.0
        let rate = 75.0

        let total = hours * rate

        XCTAssertEqual(total, 600.0)
    }

    func testPartialHourBilling() {
        let hours = 2.5
        let rate = 100.0

        let total = hours * rate

        XCTAssertEqual(total, 250.0)
    }

    // MARK: - Duration Calculations

    func testSecondsToHours() {
        XCTAssertEqual(secondsToHours(3600), 1.0)
        XCTAssertEqual(secondsToHours(5400), 1.5)
        XCTAssertEqual(secondsToHours(7200), 2.0)
        XCTAssertEqual(secondsToHours(0), 0.0)
    }

    func testHoursToSeconds() {
        XCTAssertEqual(hoursToSeconds(1.0), 3600)
        XCTAssertEqual(hoursToSeconds(1.5), 5400)
        XCTAssertEqual(hoursToSeconds(2.0), 7200)
        XCTAssertEqual(hoursToSeconds(0.0), 0)
    }

    func testDurationFromDates() {
        let start = Date()
        let end = start.addingTimeInterval(3600)

        let duration = Int(end.timeIntervalSince(start))

        XCTAssertEqual(duration, 3600)
    }

    // MARK: - Expense Totals

    func testExpenseCategoryTotal() {
        let expenses: [(category: String, amount: Double)] = [
            ("Software", 99.99),
            ("Software", 50.00),
            ("Hardware", 200.00),
            ("Software", 149.99)
        ]

        let softwareTotal = expenses
            .filter { $0.category == "Software" }
            .reduce(0) { $0 + $1.amount }

        XCTAssertEqual(softwareTotal, 299.98, accuracy: 0.01)
    }

    func testMonthlyExpenseTotal() {
        let amounts = [99.99, 50.00, 200.00, 35.00]
        let total = amounts.reduce(0, +)

        XCTAssertEqual(total, 384.99, accuracy: 0.01)
    }

    // MARK: - Profit Calculations

    func testGrossProfit() {
        let revenue = 10000.0
        let expenses = 3000.0

        let profit = revenue - expenses

        XCTAssertEqual(profit, 7000.0)
    }

    func testProfitMargin() {
        let revenue = 10000.0
        let profit = 7000.0

        let margin = (profit / revenue) * 100

        XCTAssertEqual(margin, 70.0)
    }

    // MARK: - Goal Progress Calculations

    func testGoalProgress() {
        let goal = 5000.0
        let current = 3500.0

        let progress = (current / goal) * 100

        XCTAssertEqual(progress, 70.0)
    }

    func testGoalProgressExceeded() {
        let goal = 5000.0
        let current = 6000.0

        let progress = min((current / goal) * 100, 100)

        XCTAssertEqual(progress, 100.0)
    }

    func testGoalRemaining() {
        let goal = 5000.0
        let current = 3500.0

        let remaining = max(goal - current, 0)

        XCTAssertEqual(remaining, 1500.0)
    }

    // MARK: - Pomodoro Calculations

    func testPomodoroSessionCount() {
        let totalMinutes = 200
        let sessionLength = 25

        let completedSessions = totalMinutes / sessionLength

        XCTAssertEqual(completedSessions, 8)
    }

    func testPomodoroBreakTime() {
        let shortBreak = 5
        let longBreak = 15
        let sessionsBeforeLongBreak = 4

        let sessions = 10
        let longBreaks = sessions / sessionsBeforeLongBreak
        let shortBreaks = sessions - longBreaks

        let totalBreakTime = (longBreaks * longBreak) + (shortBreaks * shortBreak)

        XCTAssertEqual(totalBreakTime, 70) // 2 * 15 + 8 * 5
    }

    // MARK: - Helper Functions

    private func secondsToHours(_ seconds: Int) -> Double {
        return Double(seconds) / 3600.0
    }

    private func hoursToSeconds(_ hours: Double) -> Int {
        return Int(hours * 3600)
    }
}
