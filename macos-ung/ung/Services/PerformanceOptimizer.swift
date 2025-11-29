//
//  PerformanceOptimizer.swift
//  ung
//
//  Performance optimization utilities: caching, debouncing, and lazy loading
//

import Foundation
import Combine

// MARK: - Data Cache

/// Thread-safe cache for expensive operations with TTL support
actor DataCache {
    static let shared = DataCache()

    private var cache: [String: CacheEntry] = [:]
    private var lastCleanup: Date = Date()

    struct CacheEntry {
        let value: Any
        let expiry: Date

        var isExpired: Bool {
            Date() > expiry
        }
    }

    func get<T>(_ key: String) -> T? {
        guard let entry = cache[key], !entry.isExpired else {
            cache.removeValue(forKey: key)
            return nil
        }
        return entry.value as? T
    }

    func set<T>(_ key: String, value: T, ttl: TimeInterval = 30) {
        cache[key] = CacheEntry(value: value, expiry: Date().addingTimeInterval(ttl))
        cleanupIfNeeded()
    }

    func invalidate(_ key: String) {
        cache.removeValue(forKey: key)
    }

    func invalidatePrefix(_ prefix: String) {
        cache = cache.filter { !$0.key.hasPrefix(prefix) }
    }

    func invalidateAll() {
        cache.removeAll()
    }

    private func cleanupIfNeeded() {
        guard Date().timeIntervalSince(lastCleanup) > 60 else { return }
        cache = cache.filter { !$0.value.isExpired }
        lastCleanup = Date()
    }
}

// MARK: - Debouncer

/// Debounces rapid calls to prevent excessive operations
class Debouncer {
    private var workItem: DispatchWorkItem?
    private let queue: DispatchQueue
    private let delay: TimeInterval

    init(delay: TimeInterval = 0.3, queue: DispatchQueue = .main) {
        self.delay = delay
        self.queue = queue
    }

    func debounce(action: @escaping () -> Void) {
        workItem?.cancel()
        workItem = DispatchWorkItem { action() }
        queue.asyncAfter(deadline: .now() + delay, execute: workItem!)
    }

    func cancel() {
        workItem?.cancel()
        workItem = nil
    }
}

// MARK: - Throttler

/// Throttles calls to ensure minimum time between executions
actor Throttler {
    private var lastExecutionTime: Date?
    private let minimumInterval: TimeInterval

    init(minimumInterval: TimeInterval = 0.5) {
        self.minimumInterval = minimumInterval
    }

    func shouldExecute() -> Bool {
        let now = Date()
        if let last = lastExecutionTime,
           now.timeIntervalSince(last) < minimumInterval {
            return false
        }
        lastExecutionTime = now
        return true
    }

    func reset() {
        lastExecutionTime = nil
    }
}

// MARK: - Formatters (Static/Cached)

/// Pre-configured formatters to avoid repeated initialization
enum Formatters {
    static let currency: NumberFormatter = {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        return formatter
    }()

    static let shortDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        return formatter
    }()

    static let mediumDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        return formatter
    }()

    static let time: DateFormatter = {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter
    }()

    static let iso8601: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    static func formatCurrency(_ amount: Double, currencyCode: String = "USD") -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = currencyCode
        return formatter.string(from: NSNumber(value: amount)) ?? "$0.00"
    }

    static func formatDuration(seconds: Int) -> String {
        let hours = seconds / 3600
        let minutes = (seconds % 3600) / 60
        let secs = seconds % 60
        if hours > 0 {
            return String(format: "%d:%02d:%02d", hours, minutes, secs)
        }
        return String(format: "%02d:%02d", minutes, secs)
    }

    static func formatHours(_ hours: Double) -> String {
        return String(format: "%.1fh", hours)
    }
}

// MARK: - Lazy Loader

/// Manages lazy loading of data with pagination
class LazyLoader<T> {
    private var items: [T] = []
    private var isLoading = false
    private var hasMore = true
    private let pageSize: Int
    private let loadPage: (Int) async throws -> [T]

    init(pageSize: Int = 20, loadPage: @escaping (Int) async throws -> [T]) {
        self.pageSize = pageSize
        self.loadPage = loadPage
    }

    var allItems: [T] { items }
    var count: Int { items.count }
    var canLoadMore: Bool { hasMore && !isLoading }

    func reset() {
        items = []
        hasMore = true
        isLoading = false
    }

    func loadInitial() async throws -> [T] {
        reset()
        return try await loadMore()
    }

    func loadMore() async throws -> [T] {
        guard !isLoading && hasMore else { return items }

        isLoading = true
        defer { isLoading = false }

        let page = items.count / pageSize
        let newItems = try await loadPage(page)

        if newItems.count < pageSize {
            hasMore = false
        }

        items.append(contentsOf: newItems)
        return items
    }
}

// MARK: - Refresh Coordinator

/// Coordinates data refresh to prevent redundant operations
@MainActor
class RefreshCoordinator: ObservableObject {
    static let shared = RefreshCoordinator()

    @Published private(set) var isRefreshing = false
    private var pendingRefreshes: Set<RefreshType> = []
    private let debouncer = Debouncer(delay: 0.3)
    private let throttler = Throttler(minimumInterval: 0.5)

    enum RefreshType: Hashable {
        case metrics
        case invoices
        case sessions
        case expenses
        case clients
        case contracts
        case all
    }

    private var refreshTask: Task<Void, Never>?

    func requestRefresh(_ types: Set<RefreshType>) {
        pendingRefreshes.formUnion(types)
        debouncer.debounce { [weak self] in
            self?.executePendingRefreshes()
        }
    }

    func requestRefresh(_ type: RefreshType) {
        requestRefresh([type])
    }

    private func executePendingRefreshes() {
        guard !pendingRefreshes.isEmpty else { return }

        let typesToRefresh = pendingRefreshes
        pendingRefreshes.removeAll()

        refreshTask?.cancel()
        refreshTask = Task { [weak self] in
            guard let self = self else { return }
            guard await self.throttler.shouldExecute() else {
                // Re-queue if throttled
                self.pendingRefreshes.formUnion(typesToRefresh)
                return
            }

            self.isRefreshing = true
            // Actual refresh would be handled by AppState
            // This is just coordination
            self.isRefreshing = false
        }
    }

    func cancelAll() {
        refreshTask?.cancel()
        pendingRefreshes.removeAll()
        debouncer.cancel()
    }
}

// MARK: - Performance Monitor

/// Monitors and logs performance metrics
class PerformanceMonitor {
    static let shared = PerformanceMonitor()

    private var operationTimes: [String: [TimeInterval]] = [:]
    private let queue = DispatchQueue(label: "performance.monitor")

    func measure<T>(_ operation: String, block: () throws -> T) rethrows -> T {
        let start = CFAbsoluteTimeGetCurrent()
        defer {
            let elapsed = CFAbsoluteTimeGetCurrent() - start
            recordTime(operation: operation, time: elapsed)
        }
        return try block()
    }

    func measure<T>(_ operation: String, block: () async throws -> T) async rethrows -> T {
        let start = CFAbsoluteTimeGetCurrent()
        defer {
            let elapsed = CFAbsoluteTimeGetCurrent() - start
            recordTime(operation: operation, time: elapsed)
        }
        return try await block()
    }

    private func recordTime(operation: String, time: TimeInterval) {
        queue.async { [weak self] in
            guard let self = self else { return }
            var times = self.operationTimes[operation] ?? []
            times.append(time)
            // Keep last 100 measurements
            if times.count > 100 {
                times.removeFirst(times.count - 100)
            }
            self.operationTimes[operation] = times

            // Log slow operations
            if time > 0.1 {
                print("⚠️ Slow operation: \(operation) took \(String(format: "%.3f", time))s")
            }
        }
    }

    func averageTime(for operation: String) -> TimeInterval? {
        queue.sync {
            guard let times = operationTimes[operation], !times.isEmpty else { return nil }
            return times.reduce(0, +) / Double(times.count)
        }
    }

    func report() -> [String: TimeInterval] {
        queue.sync {
            var report: [String: TimeInterval] = [:]
            for (op, times) in operationTimes {
                if !times.isEmpty {
                    report[op] = times.reduce(0, +) / Double(times.count)
                }
            }
            return report
        }
    }
}

// MARK: - Batch Processor

/// Batches multiple operations for efficient execution
actor BatchProcessor<T, R> {
    private var pending: [(T, CheckedContinuation<R, Error>)] = []
    private var isProcessing = false
    private let batchSize: Int
    private let batchDelay: TimeInterval
    private let processor: ([T]) async throws -> [R]

    init(batchSize: Int = 10, batchDelay: TimeInterval = 0.05, processor: @escaping ([T]) async throws -> [R]) {
        self.batchSize = batchSize
        self.batchDelay = batchDelay
        self.processor = processor
    }

    func process(_ item: T) async throws -> R {
        return try await withCheckedThrowingContinuation { continuation in
            Task {
                await self.enqueue(item, continuation: continuation)
            }
        }
    }

    private func enqueue(_ item: T, continuation: CheckedContinuation<R, Error>) {
        pending.append((item, continuation))

        if pending.count >= batchSize {
            executeBatch()
        } else if !isProcessing {
            isProcessing = true
            Task {
                try? await Task.sleep(nanoseconds: UInt64(batchDelay * 1_000_000_000))
                await self.executeBatch()
            }
        }
    }

    private func executeBatch() {
        guard !pending.isEmpty else {
            isProcessing = false
            return
        }

        let batch = pending
        pending = []

        Task {
            do {
                let items = batch.map { $0.0 }
                let results = try await processor(items)

                for (index, (_, continuation)) in batch.enumerated() {
                    if index < results.count {
                        continuation.resume(returning: results[index])
                    } else {
                        continuation.resume(throwing: BatchError.resultMismatch)
                    }
                }
            } catch {
                for (_, continuation) in batch {
                    continuation.resume(throwing: error)
                }
            }

            await self.executeBatch()
        }
    }

    enum BatchError: Error {
        case resultMismatch
    }
}
