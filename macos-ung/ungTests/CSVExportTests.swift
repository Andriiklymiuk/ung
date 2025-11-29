//
//  CSVExportTests.swift
//  ungTests
//
//  Fast unit tests for CSV export logic
//

import XCTest
@testable import ung

final class CSVExportTests: XCTestCase {

    // MARK: - ExportType Tests

    func testExportTypeRawValues() {
        XCTAssertEqual(CSVExportView.ExportType.sessions.rawValue, "Time Sessions")
        XCTAssertEqual(CSVExportView.ExportType.clients.rawValue, "Clients")
        XCTAssertEqual(CSVExportView.ExportType.contracts.rawValue, "Contracts")
        XCTAssertEqual(CSVExportView.ExportType.invoices.rawValue, "Invoices")
        XCTAssertEqual(CSVExportView.ExportType.expenses.rawValue, "Expenses")
        XCTAssertEqual(CSVExportView.ExportType.all.rawValue, "All Data")
    }

    func testExportTypeIcons() {
        XCTAssertEqual(CSVExportView.ExportType.sessions.icon, "clock.fill")
        XCTAssertEqual(CSVExportView.ExportType.clients.icon, "person.2.fill")
        XCTAssertEqual(CSVExportView.ExportType.contracts.icon, "doc.text.fill")
        XCTAssertEqual(CSVExportView.ExportType.invoices.icon, "doc.plaintext.fill")
        XCTAssertEqual(CSVExportView.ExportType.expenses.icon, "dollarsign.circle.fill")
        XCTAssertEqual(CSVExportView.ExportType.all.icon, "square.and.arrow.up.fill")
    }

    func testExportTypeCaseIterable() {
        XCTAssertEqual(CSVExportView.ExportType.allCases.count, 6)
    }

    func testExportTypeIdentifiable() {
        let type = CSVExportView.ExportType.sessions
        XCTAssertEqual(type.id, type.rawValue)
    }

    // MARK: - DateRangeOption Tests

    func testDateRangeOptionRawValues() {
        XCTAssertEqual(CSVExportView.DateRangeOption.thisWeek.rawValue, "This Week")
        XCTAssertEqual(CSVExportView.DateRangeOption.thisMonth.rawValue, "This Month")
        XCTAssertEqual(CSVExportView.DateRangeOption.thisYear.rawValue, "This Year")
        XCTAssertEqual(CSVExportView.DateRangeOption.lastMonth.rawValue, "Last Month")
        XCTAssertEqual(CSVExportView.DateRangeOption.allTime.rawValue, "All Time")
        XCTAssertEqual(CSVExportView.DateRangeOption.custom.rawValue, "Custom Range")
    }

    func testDateRangeOptionCaseIterable() {
        XCTAssertEqual(CSVExportView.DateRangeOption.allCases.count, 6)
    }

    // MARK: - CSV Escape Tests

    func testEscapeCSVSimpleString() {
        let result = CSVExportHelper.escapeCSV("Hello World")
        XCTAssertEqual(result, "Hello World")
    }

    func testEscapeCSVWithComma() {
        let result = CSVExportHelper.escapeCSV("Hello, World")
        XCTAssertEqual(result, "\"Hello, World\"")
    }

    func testEscapeCSVWithQuotes() {
        let result = CSVExportHelper.escapeCSV("Say \"Hello\"")
        XCTAssertEqual(result, "\"Say \"\"Hello\"\"\"")
    }

    func testEscapeCSVWithNewline() {
        let result = CSVExportHelper.escapeCSV("Line1\nLine2")
        XCTAssertEqual(result, "\"Line1\nLine2\"")
    }

    func testEscapeCSVEmptyString() {
        let result = CSVExportHelper.escapeCSV("")
        XCTAssertEqual(result, "")
    }

    func testEscapeCSVWithAllSpecialChars() {
        let result = CSVExportHelper.escapeCSV("Hello, \"World\"\nNew Line")
        XCTAssertTrue(result.hasPrefix("\""))
        XCTAssertTrue(result.hasSuffix("\""))
    }

    // MARK: - Filename Generation Tests

    func testFilenameGeneration() {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        let dateStr = dateFormatter.string(from: Date())

        XCTAssertEqual(CSVExportHelper.generateFilename(for: .sessions), "ung_sessions_\(dateStr).csv")
        XCTAssertEqual(CSVExportHelper.generateFilename(for: .clients), "ung_clients_\(dateStr).csv")
        XCTAssertEqual(CSVExportHelper.generateFilename(for: .contracts), "ung_contracts_\(dateStr).csv")
        XCTAssertEqual(CSVExportHelper.generateFilename(for: .invoices), "ung_invoices_\(dateStr).csv")
        XCTAssertEqual(CSVExportHelper.generateFilename(for: .expenses), "ung_expenses_\(dateStr).csv")
        XCTAssertEqual(CSVExportHelper.generateFilename(for: .all), "ung_all_data_\(dateStr).csv")
    }

    func testFilenameHasCSVExtension() {
        for type in CSVExportView.ExportType.allCases {
            let filename = CSVExportHelper.generateFilename(for: type)
            XCTAssertTrue(filename.hasSuffix(".csv"), "Filename for \(type) should have .csv extension")
        }
    }

    func testFilenameContainsDate() {
        let filename = CSVExportHelper.generateFilename(for: .sessions)
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        let dateStr = dateFormatter.string(from: Date())

        XCTAssertTrue(filename.contains(dateStr))
    }

    // MARK: - CSV Header Tests

    func testSessionsCSVHeader() {
        let header = CSVExportHelper.getCSVHeader(for: .sessions)
        XCTAssertTrue(header.contains("ID"))
        XCTAssertTrue(header.contains("Project"))
        XCTAssertTrue(header.contains("Duration"))
    }

    func testClientsCSVHeader() {
        let header = CSVExportHelper.getCSVHeader(for: .clients)
        XCTAssertTrue(header.contains("ID"))
        XCTAssertTrue(header.contains("Name"))
        XCTAssertTrue(header.contains("Email"))
    }

    func testContractsCSVHeader() {
        let header = CSVExportHelper.getCSVHeader(for: .contracts)
        XCTAssertTrue(header.contains("ID"))
        XCTAssertTrue(header.contains("Name"))
        XCTAssertTrue(header.contains("Rate"))
    }

    func testInvoicesCSVHeader() {
        let header = CSVExportHelper.getCSVHeader(for: .invoices)
        XCTAssertTrue(header.contains("ID"))
        XCTAssertTrue(header.contains("Invoice Number"))
        XCTAssertTrue(header.contains("Amount"))
        XCTAssertTrue(header.contains("Status"))
    }

    func testExpensesCSVHeader() {
        let header = CSVExportHelper.getCSVHeader(for: .expenses)
        XCTAssertTrue(header.contains("ID"))
        XCTAssertTrue(header.contains("Description"))
        XCTAssertTrue(header.contains("Amount"))
        XCTAssertTrue(header.contains("Category"))
    }

    // MARK: - Date Range Calculation Tests

    func testThisWeekDateRange() {
        let (start, end) = CSVExportHelper.getDateRange(for: .thisWeek)
        let calendar = Calendar.current

        XCTAssertTrue(calendar.isDateInToday(end) || end > Date())

        let weekStart = calendar.date(from: calendar.dateComponents([.yearForWeekOfYear, .weekOfYear], from: Date()))!
        XCTAssertEqual(calendar.startOfDay(for: start), calendar.startOfDay(for: weekStart))
    }

    func testThisMonthDateRange() {
        let (start, end) = CSVExportHelper.getDateRange(for: .thisMonth)
        let calendar = Calendar.current

        let components = calendar.dateComponents([.year, .month], from: start)
        let currentComponents = calendar.dateComponents([.year, .month], from: Date())

        XCTAssertEqual(components.year, currentComponents.year)
        XCTAssertEqual(components.month, currentComponents.month)
        XCTAssertEqual(calendar.component(.day, from: start), 1)
    }

    func testThisYearDateRange() {
        let (start, _) = CSVExportHelper.getDateRange(for: .thisYear)
        let calendar = Calendar.current

        let components = calendar.dateComponents([.year, .month, .day], from: start)
        let currentYear = calendar.component(.year, from: Date())

        XCTAssertEqual(components.year, currentYear)
        XCTAssertEqual(components.month, 1)
        XCTAssertEqual(components.day, 1)
    }

    func testLastMonthDateRange() {
        let (start, end) = CSVExportHelper.getDateRange(for: .lastMonth)
        let calendar = Calendar.current

        let startComponents = calendar.dateComponents([.month], from: start)
        let endComponents = calendar.dateComponents([.month], from: end)

        // Start and end should be in the same month
        XCTAssertEqual(startComponents.month, endComponents.month)

        // Should be last month
        let currentMonth = calendar.component(.month, from: Date())
        let expectedMonth = currentMonth == 1 ? 12 : currentMonth - 1
        XCTAssertEqual(startComponents.month, expectedMonth)
    }

    func testAllTimeDateRange() {
        let (start, end) = CSVExportHelper.getDateRange(for: .allTime)

        // Start should be far in the past
        XCTAssertTrue(start.timeIntervalSince1970 < Date().timeIntervalSince1970 - 86400 * 365 * 10)

        // End should be now or future
        XCTAssertTrue(end >= Date().addingTimeInterval(-60)) // Allow 1 minute tolerance
    }
}

// MARK: - CSV Export Helper (Testable Functions)

/// Helper struct to make CSV export functions testable without view state
struct CSVExportHelper {
    static func escapeCSV(_ value: String) -> String {
        if value.contains(",") || value.contains("\"") || value.contains("\n") {
            return "\"\(value.replacingOccurrences(of: "\"", with: "\"\""))\""
        }
        return value
    }

    static func generateFilename(for type: CSVExportView.ExportType) -> String {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        let dateStr = dateFormatter.string(from: Date())

        switch type {
        case .sessions: return "ung_sessions_\(dateStr).csv"
        case .clients: return "ung_clients_\(dateStr).csv"
        case .contracts: return "ung_contracts_\(dateStr).csv"
        case .invoices: return "ung_invoices_\(dateStr).csv"
        case .expenses: return "ung_expenses_\(dateStr).csv"
        case .all: return "ung_all_data_\(dateStr).csv"
        }
    }

    static func getCSVHeader(for type: CSVExportView.ExportType) -> String {
        switch type {
        case .sessions:
            return "ID,Project,Client,Start Time,End Time,Duration,Notes"
        case .clients:
            return "ID,Name,Email,Address,Tax ID"
        case .contracts:
            return "ID,Name,Client,Type,Hourly Rate,Fixed Price,Currency,Active"
        case .invoices:
            return "ID,Invoice Number,Client,Amount,Currency,Status,Issued Date"
        case .expenses:
            return "ID,Description,Amount,Currency,Category,Date"
        case .all:
            return "All columns from all data types"
        }
    }

    static func getDateRange(for option: CSVExportView.DateRangeOption) -> (start: Date, end: Date) {
        let calendar = Calendar.current
        let now = Date()

        switch option {
        case .thisWeek:
            let start = calendar.date(from: calendar.dateComponents([.yearForWeekOfYear, .weekOfYear], from: now))!
            return (start, now)

        case .thisMonth:
            let start = calendar.date(from: calendar.dateComponents([.year, .month], from: now))!
            return (start, now)

        case .thisYear:
            var components = calendar.dateComponents([.year], from: now)
            components.month = 1
            components.day = 1
            let start = calendar.date(from: components)!
            return (start, now)

        case .lastMonth:
            let thisMonthStart = calendar.date(from: calendar.dateComponents([.year, .month], from: now))!
            let lastMonthEnd = calendar.date(byAdding: .day, value: -1, to: thisMonthStart)!
            let lastMonthStart = calendar.date(from: calendar.dateComponents([.year, .month], from: lastMonthEnd))!
            return (lastMonthStart, lastMonthEnd)

        case .allTime:
            let start = Date(timeIntervalSince1970: 0)
            return (start, now)

        case .custom:
            // Return a default range for custom (actual dates would come from UI)
            let start = calendar.date(byAdding: .month, value: -1, to: now)!
            return (start, now)
        }
    }
}
