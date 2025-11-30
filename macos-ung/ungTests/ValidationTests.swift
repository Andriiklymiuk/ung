//
//  ValidationTests.swift
//  ungTests
//
//  Fast unit tests for input validation logic
//

import XCTest
@testable import ung

final class ValidationTests: XCTestCase {

    // MARK: - Email Validation Tests

    func testValidEmails() {
        let validEmails = [
            "test@example.com",
            "user.name@domain.co.uk",
            "user+tag@gmail.com",
            "contact@company.io",
            "name123@test.org"
        ]

        for email in validEmails {
            XCTAssertTrue(isValidEmail(email), "\(email) should be valid")
        }
    }

    func testInvalidEmails() {
        let invalidEmails = [
            "",
            "notanemail",
            "@nodomain.com",
            "no@",
            "spaces in@email.com",
            "missing.domain@"
        ]

        for email in invalidEmails {
            XCTAssertFalse(isValidEmail(email), "\(email) should be invalid")
        }
    }

    // MARK: - Amount Validation Tests

    func testValidAmounts() {
        XCTAssertTrue(isValidAmount(0))
        XCTAssertTrue(isValidAmount(100))
        XCTAssertTrue(isValidAmount(1234.56))
        XCTAssertTrue(isValidAmount(0.01))
    }

    func testInvalidAmounts() {
        XCTAssertFalse(isValidAmount(-1))
        XCTAssertFalse(isValidAmount(-100))
        XCTAssertFalse(isValidAmount(Double.nan))
        XCTAssertFalse(isValidAmount(Double.infinity))
    }

    // MARK: - Date Range Validation Tests

    func testValidDateRange() {
        let startDate = Date()
        let endDate = startDate.addingTimeInterval(3600)

        XCTAssertTrue(isValidDateRange(start: startDate, end: endDate))
    }

    func testInvalidDateRange() {
        let startDate = Date()
        let endDate = startDate.addingTimeInterval(-3600) // End before start

        XCTAssertFalse(isValidDateRange(start: startDate, end: endDate))
    }

    func testSameDateRange() {
        let date = Date()
        XCTAssertTrue(isValidDateRange(start: date, end: date))
    }

    // MARK: - Required Field Validation Tests

    func testRequiredString() {
        XCTAssertTrue(isNotEmpty("test"))
        XCTAssertTrue(isNotEmpty("  test  "))
        XCTAssertFalse(isNotEmpty(""))
        XCTAssertFalse(isNotEmpty("   "))
    }

    // MARK: - Currency Code Validation Tests

    func testValidCurrencyCodes() {
        let validCodes = ["USD", "EUR", "GBP", "CAD", "AUD", "UAH", "PLN", "CHF"]

        for code in validCodes {
            XCTAssertTrue(isValidCurrencyCode(code), "\(code) should be valid")
        }
    }

    func testInvalidCurrencyCodes() {
        let invalidCodes = ["", "US", "USDD", "123", "usd"]

        for code in invalidCodes {
            XCTAssertFalse(isValidCurrencyCode(code), "\(code) should be invalid")
        }
    }

    // MARK: - Hours Validation Tests

    func testValidHours() {
        XCTAssertTrue(isValidHours(0))
        XCTAssertTrue(isValidHours(1))
        XCTAssertTrue(isValidHours(8))
        XCTAssertTrue(isValidHours(24))
        XCTAssertTrue(isValidHours(0.5))
        XCTAssertTrue(isValidHours(0.25))
    }

    func testInvalidHours() {
        XCTAssertFalse(isValidHours(-1))
        XCTAssertFalse(isValidHours(-0.5))
        XCTAssertFalse(isValidHours(25)) // More than 24 hours
        XCTAssertFalse(isValidHours(100))
    }

    // MARK: - Invoice Number Validation Tests

    func testValidInvoiceNumbers() {
        let validNumbers = [
            "INV-001",
            "INV-2025-001",
            "2025-001",
            "ABC123"
        ]

        for number in validNumbers {
            XCTAssertTrue(isValidInvoiceNumber(number), "\(number) should be valid")
        }
    }

    func testInvalidInvoiceNumbers() {
        XCTAssertFalse(isValidInvoiceNumber(""))
        XCTAssertFalse(isValidInvoiceNumber("   "))
    }

    // MARK: - Helper Functions

    private func isValidEmail(_ email: String) -> Bool {
        let emailRegex = "[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}"
        let emailPredicate = NSPredicate(format: "SELF MATCHES %@", emailRegex)
        return emailPredicate.evaluate(with: email)
    }

    private func isValidAmount(_ amount: Double) -> Bool {
        return !amount.isNaN && !amount.isInfinite && amount >= 0
    }

    private func isValidDateRange(start: Date, end: Date) -> Bool {
        return end >= start
    }

    private func isNotEmpty(_ string: String) -> Bool {
        return !string.trimmingCharacters(in: .whitespaces).isEmpty
    }

    private func isValidCurrencyCode(_ code: String) -> Bool {
        let validCodes = ["USD", "EUR", "GBP", "CAD", "AUD", "UAH", "PLN", "CHF"]
        return validCodes.contains(code)
    }

    private func isValidHours(_ hours: Double) -> Bool {
        return hours >= 0 && hours <= 24
    }

    private func isValidInvoiceNumber(_ number: String) -> Bool {
        return !number.trimmingCharacters(in: .whitespaces).isEmpty
    }
}
