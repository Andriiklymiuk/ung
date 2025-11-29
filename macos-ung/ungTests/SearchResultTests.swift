//
//  SearchResultTests.swift
//  ungTests
//
//  Fast unit tests for SearchResult and related models
//

import XCTest
@testable import ung

final class SearchResultTests: XCTestCase {

    // MARK: - SearchResult Tests

    func testSearchResultIdentifiable() {
        let result1 = SearchResult(
            type: .invoices,
            title: "INV-001",
            subtitle: "Client A",
            deepLink: "ung://invoices/1"
        )

        let result2 = SearchResult(
            type: .invoices,
            title: "INV-001",
            subtitle: "Client A",
            deepLink: "ung://invoices/1"
        )

        // Each should have unique ID
        XCTAssertNotEqual(result1.id, result2.id)
    }

    func testSearchResultWithAmount() {
        let result = SearchResult(
            type: .invoices,
            title: "INV-001",
            subtitle: "Client A",
            amount: 1500.00,
            deepLink: "ung://invoices/1"
        )

        XCTAssertEqual(result.amount, 1500.00)
    }

    func testSearchResultWithDate() {
        let date = Date()
        let result = SearchResult(
            type: .sessions,
            title: "Development",
            subtitle: "2h 30m",
            date: date,
            deepLink: "ung://sessions/1"
        )

        XCTAssertEqual(result.date, date)
    }

    func testSearchResultWithDetail() {
        let result = SearchResult(
            type: .contracts,
            title: "Contract A",
            subtitle: "Client B",
            detail: "hourly",
            deepLink: "ung://contracts/1"
        )

        XCTAssertEqual(result.detail, "hourly")
    }

    // MARK: - SearchCategory Tests

    func testSearchCategoryRawValues() {
        XCTAssertEqual(SearchCategory.all.rawValue, "all")
        XCTAssertEqual(SearchCategory.invoices.rawValue, "invoices")
        XCTAssertEqual(SearchCategory.sessions.rawValue, "sessions")
        XCTAssertEqual(SearchCategory.clients.rawValue, "clients")
        XCTAssertEqual(SearchCategory.expenses.rawValue, "expenses")
        XCTAssertEqual(SearchCategory.contracts.rawValue, "contracts")
    }

    // MARK: - ParsedQuery Tests

    func testParsedQueryDefaults() {
        let query = ParsedQuery()

        XCTAssertEqual(query.category, .all)
        XCTAssertEqual(query.searchText, "")
        XCTAssertNil(query.dateRange)
        XCTAssertNil(query.amountFilter)
        XCTAssertNil(query.statusFilter)
        XCTAssertNil(query.clientName)
        XCTAssertNil(query.limit)
    }

    func testParsedQueryIsEmpty() {
        var query = ParsedQuery()
        XCTAssertTrue(query.isEmpty)

        query.searchText = "test"
        XCTAssertFalse(query.isEmpty)
    }

    func testParsedQueryWithCategory() {
        var query = ParsedQuery()
        query.category = .invoices

        XCTAssertEqual(query.category, .invoices)
        XCTAssertFalse(query.isEmpty) // Category alone doesn't make it non-empty
    }

    func testParsedQueryWithDateRange() {
        var query = ParsedQuery()
        let start = Date()
        let end = Date().addingTimeInterval(86400)
        query.dateRange = (start, end)

        XCTAssertNotNil(query.dateRange)
        XCTAssertEqual(query.dateRange?.start, start)
        XCTAssertEqual(query.dateRange?.end, end)
    }

    // MARK: - AmountFilter Tests

    func testAmountFilterGreaterThan() {
        let filter = AmountFilter(op: .greaterThan, value: 500)

        if case .greaterThan = filter.op {
            XCTAssertEqual(filter.value, 500)
        } else {
            XCTFail("Expected greaterThan operator")
        }
    }

    func testAmountFilterLessThan() {
        let filter = AmountFilter(op: .lessThan, value: 100)

        if case .lessThan = filter.op {
            XCTAssertEqual(filter.value, 100)
        } else {
            XCTFail("Expected lessThan operator")
        }
    }

    func testAmountFilterEquals() {
        let filter = AmountFilter(op: .equals, value: 250)

        if case .equals = filter.op {
            XCTAssertEqual(filter.value, 250)
        } else {
            XCTFail("Expected equals operator")
        }
    }

    func testAmountFilterBetween() {
        let filter = AmountFilter(op: .between(low: 100, high: 500), value: 0)

        if case .between(let low, let high) = filter.op {
            XCTAssertEqual(low, 100)
            XCTAssertEqual(high, 500)
        } else {
            XCTFail("Expected between operator")
        }
    }

    // MARK: - StatusFilter Tests

    func testStatusFilterRawValues() {
        XCTAssertEqual(StatusFilter.paid.rawValue, "paid")
        XCTAssertEqual(StatusFilter.unpaid.rawValue, "unpaid")
        XCTAssertEqual(StatusFilter.pending.rawValue, "pending")
        XCTAssertEqual(StatusFilter.overdue.rawValue, "overdue")
    }

    // MARK: - Search Result Icon Tests

    func testSearchResultIconForType() {
        let invoiceResult = SearchResult(type: .invoices, title: "", subtitle: "", deepLink: "")
        XCTAssertEqual(invoiceResult.icon, "doc.plaintext")

        let sessionResult = SearchResult(type: .sessions, title: "", subtitle: "", deepLink: "")
        XCTAssertEqual(sessionResult.icon, "clock")

        let clientResult = SearchResult(type: .clients, title: "", subtitle: "", deepLink: "")
        XCTAssertEqual(clientResult.icon, "person")

        let expenseResult = SearchResult(type: .expenses, title: "", subtitle: "", deepLink: "")
        XCTAssertEqual(expenseResult.icon, "dollarsign.circle")

        let contractResult = SearchResult(type: .contracts, title: "", subtitle: "", deepLink: "")
        XCTAssertEqual(contractResult.icon, "doc.text")
    }
}
