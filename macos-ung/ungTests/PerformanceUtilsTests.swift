//
//  PerformanceUtilsTests.swift
//  ungTests
//
//  Fast unit tests for performance optimization utilities
//

import XCTest
@testable import ung

final class PerformanceUtilsTests: XCTestCase {

    // MARK: - Debouncer Tests

    func testDebouncerDelaysExecution() {
        let expectation = XCTestExpectation(description: "Debounced action executed")
        var executionCount = 0

        let debouncer = Debouncer(delay: 0.1)

        // Fire multiple times rapidly
        for _ in 0..<5 {
            debouncer.debounce {
                executionCount += 1
                expectation.fulfill()
            }
        }

        wait(for: [expectation], timeout: 1.0)
        XCTAssertEqual(executionCount, 1, "Should only execute once")
    }

    func testDebouncerCancellation() {
        let debouncer = Debouncer(delay: 0.5)
        var executed = false

        debouncer.debounce {
            executed = true
        }

        debouncer.cancel()

        // Wait a bit longer than the delay
        Thread.sleep(forTimeInterval: 0.6)
        XCTAssertFalse(executed, "Should not execute after cancel")
    }

    func testDebouncerAllowsDelayedExecution() {
        let expectation = XCTestExpectation(description: "First debounce")
        let expectation2 = XCTestExpectation(description: "Second debounce")
        var executionCount = 0

        let debouncer = Debouncer(delay: 0.05)

        debouncer.debounce {
            executionCount += 1
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 1.0)

        // After delay, fire again
        debouncer.debounce {
            executionCount += 1
            expectation2.fulfill()
        }

        wait(for: [expectation2], timeout: 1.0)
        XCTAssertEqual(executionCount, 2, "Should execute twice with delay between")
    }

    // MARK: - Throttler Tests

    func testThrottlerAllowsFirstExecution() async {
        let throttler = Throttler(minimumInterval: 0.5)
        let allowed = await throttler.shouldExecute()
        XCTAssertTrue(allowed, "First execution should be allowed")
    }

    func testThrottlerBlocksRapidExecution() async {
        let throttler = Throttler(minimumInterval: 0.5)

        let first = await throttler.shouldExecute()
        XCTAssertTrue(first)

        // Immediately try again
        let second = await throttler.shouldExecute()
        XCTAssertFalse(second, "Rapid second execution should be blocked")
    }

    func testThrottlerAllowsAfterInterval() async {
        let throttler = Throttler(minimumInterval: 0.05)

        let first = await throttler.shouldExecute()
        XCTAssertTrue(first)

        // Wait for interval
        try? await Task.sleep(nanoseconds: 60_000_000) // 60ms

        let second = await throttler.shouldExecute()
        XCTAssertTrue(second, "Should allow after interval")
    }

    func testThrottlerReset() async {
        let throttler = Throttler(minimumInterval: 1.0)

        _ = await throttler.shouldExecute()

        // Immediately blocked
        let blocked = await throttler.shouldExecute()
        XCTAssertFalse(blocked)

        // Reset
        await throttler.reset()

        // Should allow again
        let allowed = await throttler.shouldExecute()
        XCTAssertTrue(allowed, "Should allow after reset")
    }

    // MARK: - DataCache Tests

    func testCacheSetAndGet() async {
        let cache = DataCache.shared
        let key = "test_key_\(UUID().uuidString)"

        await cache.set(key, value: "test_value", ttl: 60)

        let retrieved: String? = await cache.get(key)
        XCTAssertEqual(retrieved, "test_value")

        // Cleanup
        await cache.invalidate(key)
    }

    func testCacheExpiry() async {
        let cache = DataCache.shared
        let key = "test_expiry_\(UUID().uuidString)"

        await cache.set(key, value: "expires_soon", ttl: 0.05)

        // Immediate get should work
        let immediate: String? = await cache.get(key)
        XCTAssertEqual(immediate, "expires_soon")

        // Wait for expiry
        try? await Task.sleep(nanoseconds: 60_000_000) // 60ms

        let expired: String? = await cache.get(key)
        XCTAssertNil(expired, "Should be nil after expiry")
    }

    func testCacheInvalidatePrefix() async {
        let cache = DataCache.shared
        let prefix = "test_prefix_\(UUID().uuidString)"

        await cache.set("\(prefix)_1", value: "value1", ttl: 60)
        await cache.set("\(prefix)_2", value: "value2", ttl: 60)
        await cache.set("other_key", value: "other", ttl: 60)

        await cache.invalidatePrefix(prefix)

        let v1: String? = await cache.get("\(prefix)_1")
        let v2: String? = await cache.get("\(prefix)_2")
        let other: String? = await cache.get("other_key")

        XCTAssertNil(v1, "Prefixed key should be invalidated")
        XCTAssertNil(v2, "Prefixed key should be invalidated")
        XCTAssertEqual(other, "other", "Other key should remain")

        // Cleanup
        await cache.invalidate("other_key")
    }

    func testCacheTypeSafety() async {
        let cache = DataCache.shared
        let key = "type_test_\(UUID().uuidString)"

        await cache.set(key, value: 42, ttl: 60)

        // Correct type
        let intValue: Int? = await cache.get(key)
        XCTAssertEqual(intValue, 42)

        // Wrong type
        let stringValue: String? = await cache.get(key)
        XCTAssertNil(stringValue, "Wrong type should return nil")

        // Cleanup
        await cache.invalidate(key)
    }

    // MARK: - LazyLoader Tests

    func testLazyLoaderInitialLoad() async throws {
        var pageLoadCount = 0

        let loader = LazyLoader<Int>(pageSize: 10) { page in
            pageLoadCount += 1
            return (0..<10).map { page * 10 + $0 }
        }

        let items = try await loader.loadInitial()
        XCTAssertEqual(items.count, 10)
        XCTAssertEqual(pageLoadCount, 1)
    }

    func testLazyLoaderPagination() async throws {
        let loader = LazyLoader<Int>(pageSize: 5) { page in
            return (0..<5).map { page * 5 + $0 }
        }

        _ = try await loader.loadInitial()
        XCTAssertEqual(loader.count, 5)

        _ = try await loader.loadMore()
        XCTAssertEqual(loader.count, 10)

        _ = try await loader.loadMore()
        XCTAssertEqual(loader.count, 15)
    }

    func testLazyLoaderDetectsEndOfData() async throws {
        let loader = LazyLoader<Int>(pageSize: 10) { page in
            // Return less than page size on second page
            if page == 0 {
                return Array(0..<10)
            } else {
                return Array(0..<3) // Less than page size
            }
        }

        _ = try await loader.loadInitial()
        XCTAssertTrue(loader.canLoadMore)

        _ = try await loader.loadMore()
        XCTAssertFalse(loader.canLoadMore, "Should detect end of data")
    }

    func testLazyLoaderReset() async throws {
        let loader = LazyLoader<Int>(pageSize: 5) { _ in
            return [1, 2, 3, 4, 5]
        }

        _ = try await loader.loadInitial()
        _ = try await loader.loadMore()
        XCTAssertEqual(loader.count, 10)

        loader.reset()
        XCTAssertEqual(loader.count, 0)
        XCTAssertTrue(loader.canLoadMore)
    }

    // MARK: - PerformanceMonitor Tests

    func testPerformanceMonitorMeasures() {
        let monitor = PerformanceMonitor.shared
        let operationName = "test_operation_\(UUID().uuidString)"

        // Measure a simple operation
        let result: Int = monitor.measure(operationName) {
            Thread.sleep(forTimeInterval: 0.01)
            return 42
        }

        XCTAssertEqual(result, 42)

        // Check that time was recorded
        let average = monitor.averageTime(for: operationName)
        XCTAssertNotNil(average)
        XCTAssertGreaterThan(average!, 0.005)
    }

    func testPerformanceMonitorReport() {
        let monitor = PerformanceMonitor.shared
        let operationName = "report_test_\(UUID().uuidString)"

        // Run a few measurements
        for _ in 0..<3 {
            _ = monitor.measure(operationName) {
                Thread.sleep(forTimeInterval: 0.001)
                return true
            }
        }

        let report = monitor.report()
        XCTAssertNotNil(report[operationName])
    }
}
