//
//  SmartSearchServiceTests.swift
//  ungTests
//
//  Fast unit tests for SmartSearchService NLP query parsing
//

import XCTest
@testable import ung

final class SmartSearchServiceTests: XCTestCase {
    var service: SmartSearchService!

    override func setUp() {
        super.setUp()
        service = SmartSearchService.shared
    }

    // MARK: - Category Detection Tests

    func testDetectsInvoiceCategory() {
        let queries = [
            "show invoices",
            "unpaid invoices",
            "all bills",
            "overdue payments"
        ]

        for query in queries {
            let result = service.parse(query)
            XCTAssertEqual(result.category, .invoices, "Failed for query: \(query)")
        }
    }

    func testDetectsSessionCategory() {
        let queries = [
            "hours today",
            "time this week",
            "sessions last month",
            "work tracking"
        ]

        for query in queries {
            let result = service.parse(query)
            XCTAssertEqual(result.category, .sessions, "Failed for query: \(query)")
        }
    }

    func testDetectsClientCategory() {
        let queries = [
            "my clients",
            "client list",
            "all customers"
        ]

        for query in queries {
            let result = service.parse(query)
            XCTAssertEqual(result.category, .clients, "Failed for query: \(query)")
        }
    }

    func testDetectsExpenseCategory() {
        let queries = [
            "expenses this month",
            "costs last week",
            "spending"
        ]

        for query in queries {
            let result = service.parse(query)
            XCTAssertEqual(result.category, .expenses, "Failed for query: \(query)")
        }
    }

    // MARK: - Date Range Tests

    func testParsesToday() {
        let result = service.parse("invoices today")
        XCTAssertNotNil(result.dateRange)

        if let range = result.dateRange {
            let calendar = Calendar.current
            XCTAssertTrue(calendar.isDateInToday(range.start))
        }
    }

    func testParsesThisWeek() {
        let result = service.parse("hours this week")
        XCTAssertNotNil(result.dateRange)

        if let range = result.dateRange {
            let calendar = Calendar.current
            let now = Date()
            let weekStart = calendar.date(from: calendar.dateComponents([.yearForWeekOfYear, .weekOfYear], from: now))!
            XCTAssertEqual(calendar.startOfDay(for: range.start), calendar.startOfDay(for: weekStart))
        }
    }

    func testParsesLastMonth() {
        let result = service.parse("invoices last month")
        XCTAssertNotNil(result.dateRange)

        if let range = result.dateRange {
            let calendar = Calendar.current
            let components = calendar.dateComponents([.month, .year], from: range.start)
            let currentMonth = calendar.component(.month, from: Date())
            let expectedMonth = currentMonth == 1 ? 12 : currentMonth - 1
            XCTAssertEqual(components.month, expectedMonth)
        }
    }

    func testParsesLast30Days() {
        let result = service.parse("expenses last 30 days")
        XCTAssertNotNil(result.dateRange)

        if let range = result.dateRange {
            let calendar = Calendar.current
            let days = calendar.dateComponents([.day], from: range.start, to: range.end).day!
            XCTAssertEqual(days, 30, accuracy: 1)
        }
    }

    // MARK: - Amount Filter Tests

    func testParsesOverAmount() {
        let result = service.parse("invoices over $500")
        XCTAssertNotNil(result.amountFilter)

        if let filter = result.amountFilter {
            if case .greaterThan = filter.op {
                XCTAssertEqual(filter.value, 500)
            } else {
                XCTFail("Expected greaterThan operator")
            }
        }
    }

    func testParsesUnderAmount() {
        let result = service.parse("expenses under $100")
        XCTAssertNotNil(result.amountFilter)

        if let filter = result.amountFilter {
            if case .lessThan = filter.op {
                XCTAssertEqual(filter.value, 100)
            } else {
                XCTFail("Expected lessThan operator")
            }
        }
    }

    func testParsesAmountWithoutDollarSign() {
        let result = service.parse("invoices more than 1000")
        XCTAssertNotNil(result.amountFilter)
        XCTAssertEqual(result.amountFilter?.value, 1000)
    }

    func testParsesBetweenAmount() {
        let result = service.parse("invoices between $100 and $500")
        XCTAssertNotNil(result.amountFilter)

        if let filter = result.amountFilter {
            if case .between(let low, let high) = filter.op {
                XCTAssertEqual(low, 100)
                XCTAssertEqual(high, 500)
            } else {
                XCTFail("Expected between operator")
            }
        }
    }

    // MARK: - Status Filter Tests

    func testParsesUnpaidStatus() {
        let result = service.parse("unpaid invoices")
        XCTAssertEqual(result.statusFilter, .unpaid)
    }

    func testParsesPaidStatus() {
        let result = service.parse("paid invoices this month")
        XCTAssertEqual(result.statusFilter, .paid)
    }

    func testParsesOverdueStatus() {
        let result = service.parse("overdue invoices")
        XCTAssertEqual(result.statusFilter, .overdue)
    }

    func testParsesPendingStatus() {
        let result = service.parse("pending payments")
        XCTAssertEqual(result.statusFilter, .pending)
    }

    // MARK: - Limit Tests

    func testParsesTopN() {
        let result = service.parse("top 5 invoices")
        XCTAssertEqual(result.limit, 5)
    }

    func testParsesFirstN() {
        let result = service.parse("first 10 clients")
        XCTAssertEqual(result.limit, 10)
    }

    func testLimitCappedAt100() {
        let result = service.parse("top 500 invoices")
        XCTAssertEqual(result.limit, 100)
    }

    // MARK: - Combined Query Tests

    func testComplexQuery() {
        let result = service.parse("unpaid invoices last month over $500")

        XCTAssertEqual(result.category, .invoices)
        XCTAssertEqual(result.statusFilter, .unpaid)
        XCTAssertNotNil(result.dateRange)
        XCTAssertNotNil(result.amountFilter)
        XCTAssertEqual(result.amountFilter?.value, 500)
    }

    func testEmptyQuery() {
        let result = service.parse("")
        XCTAssertTrue(result.isEmpty)
        XCTAssertEqual(result.category, .all)
    }

    func testWhitespaceQuery() {
        let result = service.parse("   ")
        XCTAssertTrue(result.isEmpty)
    }

    // MARK: - Suggestions Tests

    func testSuggestionsForEmpty() {
        let suggestions = service.getSuggestions(for: "")
        XCTAssertFalse(suggestions.isEmpty)
    }

    func testSuggestionsForInvoice() {
        let suggestions = service.getSuggestions(for: "inv")
        XCTAssertTrue(suggestions.contains { $0.contains("invoice") })
    }

    func testSuggestionsForLast() {
        let suggestions = service.getSuggestions(for: "last")
        XCTAssertTrue(suggestions.contains { $0.contains("week") || $0.contains("month") })
    }
}
