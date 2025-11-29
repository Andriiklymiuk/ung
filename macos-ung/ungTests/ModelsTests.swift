//
//  ModelsTests.swift
//  ungTests
//
//  Fast unit tests for database models logic
//

import XCTest
@testable import ung

final class ModelsTests: XCTestCase {

    // MARK: - TrackingSession Tests

    func testTrackingSessionIsActive() {
        var session = TrackingSession(
            startTime: Date()
        )

        XCTAssertTrue(session.isActive, "Session without endTime should be active")

        session.endTime = Date()
        XCTAssertFalse(session.isActive, "Session with endTime should not be active")
    }

    func testTrackingSessionIsActiveWithDeletedAt() {
        var session = TrackingSession(
            startTime: Date()
        )

        session.deletedAt = Date()
        XCTAssertFalse(session.isActive, "Deleted session should not be active")
    }

    func testTrackingSessionCalculatedDuration() {
        let now = Date()
        let oneHourAgo = now.addingTimeInterval(-3600)

        let session = TrackingSession(
            startTime: oneHourAgo,
            endTime: now
        )

        let duration = session.calculatedDuration
        XCTAssertEqual(duration, 3600, accuracy: 1)
    }

    func testTrackingSessionFormattedDuration() {
        let now = Date()
        let startTime = now.addingTimeInterval(-3665) // 1h 1m 5s ago

        let session = TrackingSession(
            startTime: startTime,
            endTime: now
        )

        let formatted = session.formattedDuration
        XCTAssertEqual(formatted, "1:01:05")
    }

    func testTrackingSessionFormattedDurationZero() {
        let now = Date()

        let session = TrackingSession(
            startTime: now,
            endTime: now
        )

        let formatted = session.formattedDuration
        XCTAssertEqual(formatted, "0:00:00")
    }

    // MARK: - RecurringFrequency Tests

    func testRecurringFrequencyDisplayNames() {
        XCTAssertEqual(RecurringFrequency.weekly.displayName, "Weekly")
        XCTAssertEqual(RecurringFrequency.biweekly.displayName, "Bi-weekly")
        XCTAssertEqual(RecurringFrequency.monthly.displayName, "Monthly")
        XCTAssertEqual(RecurringFrequency.quarterly.displayName, "Quarterly")
        XCTAssertEqual(RecurringFrequency.yearly.displayName, "Yearly")
    }

    func testRecurringFrequencyCaseIterable() {
        let allCases = RecurringFrequency.allCases
        XCTAssertEqual(allCases.count, 5)
    }

    // MARK: - RecurringInvoice Tests

    func testRecurringInvoiceFrequencyType() {
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "monthly",
            dayOfMonth: 1,
            dayOfWeek: 1,
            nextGenerationDate: Date(),
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        XCTAssertEqual(invoice.frequencyType, .monthly)
    }

    func testRecurringInvoiceInvalidFrequencyDefaultsToMonthly() {
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "invalid",
            dayOfMonth: 1,
            dayOfWeek: 1,
            nextGenerationDate: Date(),
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        XCTAssertEqual(invoice.frequencyType, .monthly)
    }

    func testRecurringInvoiceCalculateNextWeekly() {
        let now = Date()
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "weekly",
            dayOfMonth: 1,
            dayOfWeek: 1,
            nextGenerationDate: now,
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        let next = invoice.calculateNextGenerationDate(from: now)
        let calendar = Calendar.current
        let days = calendar.dateComponents([.day], from: now, to: next).day!

        XCTAssertEqual(days, 7, accuracy: 1)
    }

    func testRecurringInvoiceCalculateNextBiweekly() {
        let now = Date()
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "biweekly",
            dayOfMonth: 1,
            dayOfWeek: 1,
            nextGenerationDate: now,
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        let next = invoice.calculateNextGenerationDate(from: now)
        let calendar = Calendar.current
        let days = calendar.dateComponents([.day], from: now, to: next).day!

        XCTAssertEqual(days, 14, accuracy: 1)
    }

    func testRecurringInvoiceCalculateNextMonthly() {
        let now = Date()
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "monthly",
            dayOfMonth: 15,
            dayOfWeek: 1,
            nextGenerationDate: now,
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        let next = invoice.calculateNextGenerationDate(from: now)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.month], from: now, to: next)

        XCTAssertEqual(components.month, 1)
    }

    func testRecurringInvoiceCalculateNextQuarterly() {
        let now = Date()
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "quarterly",
            dayOfMonth: 1,
            dayOfWeek: 1,
            nextGenerationDate: now,
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        let next = invoice.calculateNextGenerationDate(from: now)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.month], from: now, to: next)

        XCTAssertEqual(components.month, 3)
    }

    func testRecurringInvoiceCalculateNextYearly() {
        let now = Date()
        let invoice = RecurringInvoice(
            clientId: 1,
            amount: 1000,
            currency: "USD",
            frequency: "yearly",
            dayOfMonth: 1,
            dayOfWeek: 1,
            nextGenerationDate: now,
            active: true,
            autoPdf: false,
            autoSend: false,
            generatedCount: 0
        )

        let next = invoice.calculateNextGenerationDate(from: now)
        let calendar = Calendar.current
        let components = calendar.dateComponents([.year], from: now, to: next)

        XCTAssertEqual(components.year, 1)
    }

    // MARK: - ExpenseCategory Tests

    func testExpenseCategoryRawValues() {
        XCTAssertEqual(ExpenseCategory.software.rawValue, "Software")
        XCTAssertEqual(ExpenseCategory.hardware.rawValue, "Hardware")
        XCTAssertEqual(ExpenseCategory.travel.rawValue, "Travel")
        XCTAssertEqual(ExpenseCategory.meals.rawValue, "Meals")
        XCTAssertEqual(ExpenseCategory.office.rawValue, "Office")
        XCTAssertEqual(ExpenseCategory.marketing.rawValue, "Marketing")
        XCTAssertEqual(ExpenseCategory.education.rawValue, "Education")
        XCTAssertEqual(ExpenseCategory.other.rawValue, "Other")
    }

    func testExpenseCategoryCaseIterable() {
        XCTAssertEqual(ExpenseCategory.allCases.count, 8)
    }

    // MARK: - InvoiceStatus Tests

    func testInvoiceStatusRawValues() {
        XCTAssertEqual(InvoiceStatus.pending.rawValue, "pending")
        XCTAssertEqual(InvoiceStatus.sent.rawValue, "sent")
        XCTAssertEqual(InvoiceStatus.paid.rawValue, "paid")
        XCTAssertEqual(InvoiceStatus.overdue.rawValue, "overdue")
    }

    // MARK: - ContractType Tests

    func testContractTypeRawValues() {
        XCTAssertEqual(ContractType.hourly.rawValue, "hourly")
        XCTAssertEqual(ContractType.fixedPrice.rawValue, "fixed_price")
        XCTAssertEqual(ContractType.retainer.rawValue, "retainer")
    }

    // MARK: - Database Table Names

    func testDatabaseTableNames() {
        XCTAssertEqual(Company.databaseTableName, "companies")
        XCTAssertEqual(ClientModel.databaseTableName, "clients")
        XCTAssertEqual(ContractModel.databaseTableName, "contracts")
        XCTAssertEqual(Invoice.databaseTableName, "invoices")
        XCTAssertEqual(InvoiceLineItem.databaseTableName, "invoice_line_items")
        XCTAssertEqual(InvoiceRecipient.databaseTableName, "invoice_recipients")
        XCTAssertEqual(TrackingSession.databaseTableName, "tracking_sessions")
        XCTAssertEqual(Expense.databaseTableName, "expenses")
        XCTAssertEqual(SchemaMigration.databaseTableName, "schema_migrations")
    }
}
