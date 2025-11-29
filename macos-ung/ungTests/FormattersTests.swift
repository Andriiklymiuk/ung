//
//  FormattersTests.swift
//  ungTests
//
//  Fast unit tests for Formatters utilities
//

import XCTest
@testable import ung

final class FormattersTests: XCTestCase {

    // MARK: - Currency Formatter Tests

    func testCurrencyFormatsPositive() {
        let result = Formatters.currency.string(from: NSNumber(value: 1234.56))
        XCTAssertNotNil(result)
        XCTAssertTrue(result!.contains("1,234") || result!.contains("1234"))
        XCTAssertTrue(result!.contains("$") || result!.contains("USD"))
    }

    func testCurrencyFormatsZero() {
        let result = Formatters.currency.string(from: NSNumber(value: 0))
        XCTAssertNotNil(result)
        XCTAssertTrue(result!.contains("0"))
    }

    func testCurrencyFormatsNegative() {
        let result = Formatters.currency.string(from: NSNumber(value: -500))
        XCTAssertNotNil(result)
        XCTAssertTrue(result!.contains("500"))
    }

    func testFormatCurrencyWithCode() {
        let usd = Formatters.formatCurrency(100, currencyCode: "USD")
        XCTAssertTrue(usd.contains("$") || usd.contains("100"))

        let eur = Formatters.formatCurrency(100, currencyCode: "EUR")
        XCTAssertTrue(eur.contains("100"))
    }

    // MARK: - Date Formatter Tests

    func testShortDateFormat() {
        let date = Date()
        let result = Formatters.shortDate.string(from: date)
        XCTAssertFalse(result.isEmpty)
    }

    func testMediumDateFormat() {
        let date = Date()
        let result = Formatters.mediumDate.string(from: date)
        XCTAssertFalse(result.isEmpty)
        // Medium format should be longer than short
        XCTAssertGreaterThan(result.count, Formatters.shortDate.string(from: date).count)
    }

    func testTimeFormat() {
        let date = Date()
        let result = Formatters.time.string(from: date)
        XCTAssertFalse(result.isEmpty)
        // Time format should contain : for hours:minutes
        XCTAssertTrue(result.contains(":"))
    }

    // MARK: - Duration Formatter Tests

    func testFormatDurationSeconds() {
        let result = Formatters.formatDuration(seconds: 45)
        XCTAssertEqual(result, "00:45")
    }

    func testFormatDurationMinutes() {
        let result = Formatters.formatDuration(seconds: 125)
        XCTAssertEqual(result, "02:05")
    }

    func testFormatDurationHours() {
        let result = Formatters.formatDuration(seconds: 3665) // 1h 1m 5s
        XCTAssertEqual(result, "1:01:05")
    }

    func testFormatDurationZero() {
        let result = Formatters.formatDuration(seconds: 0)
        XCTAssertEqual(result, "00:00")
    }

    func testFormatDurationLarge() {
        let result = Formatters.formatDuration(seconds: 36000) // 10 hours
        XCTAssertEqual(result, "10:00:00")
    }

    // MARK: - Hours Formatter Tests

    func testFormatHours() {
        XCTAssertEqual(Formatters.formatHours(1.0), "1.0h")
        XCTAssertEqual(Formatters.formatHours(1.5), "1.5h")
        XCTAssertEqual(Formatters.formatHours(10.25), "10.2h") // Rounds to 1 decimal
    }

    func testFormatHoursZero() {
        XCTAssertEqual(Formatters.formatHours(0), "0.0h")
    }

    // MARK: - ISO8601 Formatter Tests

    func testISO8601Format() {
        let date = Date()
        let result = Formatters.iso8601.string(from: date)
        XCTAssertFalse(result.isEmpty)
        XCTAssertTrue(result.contains("T")) // ISO8601 contains T separator
    }

    func testISO8601RoundTrip() {
        let original = Date()
        let formatted = Formatters.iso8601.string(from: original)
        let parsed = Formatters.iso8601.date(from: formatted)

        XCTAssertNotNil(parsed)
        if let parsed = parsed {
            // Should be within 1 second due to rounding
            XCTAssertEqual(original.timeIntervalSince1970, parsed.timeIntervalSince1970, accuracy: 1)
        }
    }

    // MARK: - Thread Safety Tests

    func testConcurrentCurrencyFormatting() {
        let expectation = XCTestExpectation(description: "Concurrent formatting")
        expectation.expectedFulfillmentCount = 100

        let queue = DispatchQueue.global(qos: .userInteractive)
        for i in 0..<100 {
            queue.async {
                let result = Formatters.currency.string(from: NSNumber(value: Double(i) * 100))
                XCTAssertNotNil(result)
                expectation.fulfill()
            }
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testConcurrentDateFormatting() {
        let expectation = XCTestExpectation(description: "Concurrent date formatting")
        expectation.expectedFulfillmentCount = 100

        let queue = DispatchQueue.global(qos: .userInteractive)
        let date = Date()

        for _ in 0..<100 {
            queue.async {
                let result = Formatters.shortDate.string(from: date)
                XCTAssertFalse(result.isEmpty)
                expectation.fulfill()
            }
        }

        wait(for: [expectation], timeout: 5.0)
    }
}
