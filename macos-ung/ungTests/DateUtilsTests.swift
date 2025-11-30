//
//  DateUtilsTests.swift
//  ungTests
//
//  Fast unit tests for date utilities
//

import XCTest
@testable import ung

final class DateUtilsTests: XCTestCase {

    // MARK: - Date Calculations Tests

    func testDaysBetweenDates() {
        let calendar = Calendar.current
        let today = Date()
        let tomorrow = calendar.date(byAdding: .day, value: 1, to: today)!
        let nextWeek = calendar.date(byAdding: .day, value: 7, to: today)!

        XCTAssertEqual(daysBetween(today, and: tomorrow), 1)
        XCTAssertEqual(daysBetween(today, and: nextWeek), 7)
        XCTAssertEqual(daysBetween(today, and: today), 0)
    }

    func testDaysBetweenNegative() {
        let calendar = Calendar.current
        let today = Date()
        let yesterday = calendar.date(byAdding: .day, value: -1, to: today)!

        XCTAssertEqual(daysBetween(today, and: yesterday), -1)
    }

    // MARK: - Start/End of Day Tests

    func testStartOfDay() {
        let date = Date()
        let start = startOfDay(date)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.hour, .minute, .second], from: start)

        XCTAssertEqual(components.hour, 0)
        XCTAssertEqual(components.minute, 0)
        XCTAssertEqual(components.second, 0)
    }

    func testEndOfDay() {
        let date = Date()
        let end = endOfDay(date)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.hour, .minute, .second], from: end)

        XCTAssertEqual(components.hour, 23)
        XCTAssertEqual(components.minute, 59)
        XCTAssertEqual(components.second, 59)
    }

    // MARK: - Week Calculations Tests

    func testStartOfWeek() {
        let date = Date()
        let start = startOfWeek(date)
        let calendar = Calendar.current
        let weekday = calendar.component(.weekday, from: start)

        XCTAssertEqual(weekday, calendar.firstWeekday)
    }

    func testEndOfWeek() {
        let date = Date()
        let start = startOfWeek(date)
        let end = endOfWeek(date)
        let days = daysBetween(start, and: end)

        XCTAssertEqual(days, 6) // 7 days - 1 (start day)
    }

    // MARK: - Month Calculations Tests

    func testStartOfMonth() {
        let date = Date()
        let start = startOfMonth(date)
        let calendar = Calendar.current
        let day = calendar.component(.day, from: start)

        XCTAssertEqual(day, 1)
    }

    func testEndOfMonth() {
        let calendar = Calendar.current
        let components = DateComponents(year: 2025, month: 1, day: 15)
        let date = calendar.date(from: components)!
        let end = endOfMonth(date)
        let day = calendar.component(.day, from: end)

        XCTAssertEqual(day, 31) // January has 31 days
    }

    func testEndOfMonthFebruary() {
        let calendar = Calendar.current
        let components = DateComponents(year: 2025, month: 2, day: 10)
        let date = calendar.date(from: components)!
        let end = endOfMonth(date)
        let day = calendar.component(.day, from: end)

        XCTAssertEqual(day, 28) // 2025 is not a leap year
    }

    func testEndOfMonthFebruaryLeapYear() {
        let calendar = Calendar.current
        let components = DateComponents(year: 2024, month: 2, day: 10)
        let date = calendar.date(from: components)!
        let end = endOfMonth(date)
        let day = calendar.component(.day, from: end)

        XCTAssertEqual(day, 29) // 2024 is a leap year
    }

    // MARK: - Year Calculations Tests

    func testStartOfYear() {
        let date = Date()
        let start = startOfYear(date)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.month, .day], from: start)

        XCTAssertEqual(components.month, 1)
        XCTAssertEqual(components.day, 1)
    }

    func testEndOfYear() {
        let date = Date()
        let end = endOfYear(date)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.month, .day], from: end)

        XCTAssertEqual(components.month, 12)
        XCTAssertEqual(components.day, 31)
    }

    // MARK: - Due Date Tests

    func testAddDays() {
        let today = Date()
        let in30Days = addDays(30, to: today)
        let days = daysBetween(today, and: in30Days)

        XCTAssertEqual(days, 30)
    }

    func testIsOverdue() {
        let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date())!
        let tomorrow = Calendar.current.date(byAdding: .day, value: 1, to: Date())!

        XCTAssertTrue(isOverdue(yesterday))
        XCTAssertFalse(isOverdue(tomorrow))
    }

    func testIsDueToday() {
        let today = Date()
        XCTAssertTrue(isDueToday(today))

        let tomorrow = Calendar.current.date(byAdding: .day, value: 1, to: today)!
        XCTAssertFalse(isDueToday(tomorrow))
    }

    // MARK: - Helper Functions

    private func daysBetween(_ start: Date, and end: Date) -> Int {
        let calendar = Calendar.current
        let components = calendar.dateComponents([.day], from: start, to: end)
        return components.day ?? 0
    }

    private func startOfDay(_ date: Date) -> Date {
        return Calendar.current.startOfDay(for: date)
    }

    private func endOfDay(_ date: Date) -> Date {
        var components = DateComponents()
        components.day = 1
        components.second = -1
        return Calendar.current.date(byAdding: components, to: startOfDay(date))!
    }

    private func startOfWeek(_ date: Date) -> Date {
        let calendar = Calendar.current
        let components = calendar.dateComponents([.yearForWeekOfYear, .weekOfYear], from: date)
        return calendar.date(from: components)!
    }

    private func endOfWeek(_ date: Date) -> Date {
        var components = DateComponents()
        components.weekOfYear = 1
        components.second = -1
        return Calendar.current.date(byAdding: components, to: startOfWeek(date))!
    }

    private func startOfMonth(_ date: Date) -> Date {
        let calendar = Calendar.current
        let components = calendar.dateComponents([.year, .month], from: date)
        return calendar.date(from: components)!
    }

    private func endOfMonth(_ date: Date) -> Date {
        var components = DateComponents()
        components.month = 1
        components.second = -1
        return Calendar.current.date(byAdding: components, to: startOfMonth(date))!
    }

    private func startOfYear(_ date: Date) -> Date {
        let calendar = Calendar.current
        let components = calendar.dateComponents([.year], from: date)
        return calendar.date(from: components)!
    }

    private func endOfYear(_ date: Date) -> Date {
        var components = DateComponents()
        components.year = 1
        components.second = -1
        return Calendar.current.date(byAdding: components, to: startOfYear(date))!
    }

    private func addDays(_ days: Int, to date: Date) -> Date {
        return Calendar.current.date(byAdding: .day, value: days, to: date)!
    }

    private func isOverdue(_ dueDate: Date) -> Bool {
        return dueDate < startOfDay(Date())
    }

    private func isDueToday(_ date: Date) -> Bool {
        return Calendar.current.isDateInToday(date)
    }
}
