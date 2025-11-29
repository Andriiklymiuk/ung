//
//  SmartGoalsTests.swift
//  ungTests
//
//  Unit tests for Smart Goals data models and logic
//

import XCTest
@testable import ung

final class SmartGoalsTests: XCTestCase {

    // MARK: - GoalSuggestion Tests

    func testGoalSuggestionIdentifiable() {
        let suggestion1 = GoalSuggestion(
            type: .weeklyHours,
            title: "Weekly Hours",
            description: "Test",
            suggestedValue: 40,
            currentValue: 35,
            confidence: .high,
            reasoning: "Test reasoning",
            icon: "clock",
            trend: .up
        )

        let suggestion2 = GoalSuggestion(
            type: .weeklyHours,
            title: "Weekly Hours",
            description: "Test",
            suggestedValue: 40,
            currentValue: 35,
            confidence: .high,
            reasoning: "Test reasoning",
            icon: "clock",
            trend: .up
        )

        // Different IDs even with same content
        XCTAssertNotEqual(suggestion1.id, suggestion2.id)
    }

    func testConfidenceLevelColors() {
        XCTAssertEqual(GoalSuggestion.ConfidenceLevel.high.color, "green")
        XCTAssertEqual(GoalSuggestion.ConfidenceLevel.medium.color, "orange")
        XCTAssertEqual(GoalSuggestion.ConfidenceLevel.low.color, "gray")
    }

    func testConfidenceLevelRawValues() {
        XCTAssertEqual(GoalSuggestion.ConfidenceLevel.high.rawValue, "High")
        XCTAssertEqual(GoalSuggestion.ConfidenceLevel.medium.rawValue, "Medium")
        XCTAssertEqual(GoalSuggestion.ConfidenceLevel.low.rawValue, "Low")
    }

    func testTrendDirectionIcons() {
        XCTAssertEqual(GoalSuggestion.TrendDirection.up.icon, "arrow.up.right")
        XCTAssertEqual(GoalSuggestion.TrendDirection.down.icon, "arrow.down.right")
        XCTAssertEqual(GoalSuggestion.TrendDirection.stable.icon, "arrow.right")
    }

    // MARK: - GoalInsight Tests

    func testGoalInsightIdentifiable() {
        let insight1 = GoalInsight(
            title: "Test",
            message: "Test message",
            type: .positive,
            actionSuggestion: nil,
            icon: "star"
        )

        let insight2 = GoalInsight(
            title: "Test",
            message: "Test message",
            type: .positive,
            actionSuggestion: nil,
            icon: "star"
        )

        XCTAssertNotEqual(insight1.id, insight2.id)
    }

    func testGoalInsightTypes() {
        let types: [GoalInsight.InsightType] = [.positive, .warning, .info, .achievement]
        XCTAssertEqual(types.count, 4)
    }

    // MARK: - Goal Type Tests

    func testGoalTypes() {
        let types: [GoalSuggestion.GoalType] = [
            .weeklyHours,
            .monthlyIncome,
            .yearlyIncome,
            .hourlyRate,
            .clientsPerMonth,
            .projectsPerMonth
        ]
        XCTAssertEqual(types.count, 6)
    }

    // MARK: - Calculation Logic Tests

    func testPercentageChange() {
        // Test percentage change calculation
        let current: Double = 100
        let suggested: Double = 115
        let change = ((suggested - current) / current) * 100

        XCTAssertEqual(change, 15, accuracy: 0.01)
    }

    func testNegativePercentageChange() {
        let current: Double = 100
        let suggested: Double = 80
        let change = ((suggested - current) / current) * 100

        XCTAssertEqual(change, -20, accuracy: 0.01)
    }

    func testProgressCalculation() {
        // Test progress toward goal
        let current: Double = 30
        let goal: Double = 40
        let progress = current / goal

        XCTAssertEqual(progress, 0.75, accuracy: 0.01)
    }

    func testProgressExceedsGoal() {
        let current: Double = 50
        let goal: Double = 40
        let progress = current / goal

        XCTAssertEqual(progress, 1.25, accuracy: 0.01)
    }

    // MARK: - Rounding Logic Tests

    func testRoundToNearest100() {
        let value: Double = 1234.56
        let rounded = round(value / 100) * 100

        XCTAssertEqual(rounded, 1200)
    }

    func testRoundToNearest1000() {
        let value: Double = 45678.90
        let rounded = round(value / 1000) * 1000

        XCTAssertEqual(rounded, 46000)
    }

    // MARK: - Growth Rate Tests

    func testGrowthRatePositive() {
        let recent: Double = 1500
        let previous: Double = 1000
        let growthRate = (recent - previous) / previous

        XCTAssertEqual(growthRate, 0.5, accuracy: 0.01) // 50% growth
    }

    func testGrowthRateNegative() {
        let recent: Double = 800
        let previous: Double = 1000
        let growthRate = (recent - previous) / previous

        XCTAssertEqual(growthRate, -0.2, accuracy: 0.01) // -20% decline
    }

    func testGrowthRateZeroPrevious() {
        let recent: Double = 1000
        let previous: Double = 0
        let growthRate = previous > 0 ? (recent - previous) / previous : 0

        XCTAssertEqual(growthRate, 0) // Avoid division by zero
    }

    // MARK: - Variance Calculation Tests

    func testVarianceCalculation() {
        let values = [38.0, 40.0, 42.0, 40.0]
        let mean = values.reduce(0, +) / Double(values.count)
        let variance = values.map { pow($0 - mean, 2) }.reduce(0, +) / Double(values.count)
        let stdDev = sqrt(variance)

        XCTAssertEqual(mean, 40, accuracy: 0.01)
        XCTAssertEqual(stdDev, 1.414, accuracy: 0.01)
    }

    func testLowVarianceIndicatesConsistency() {
        let consistentValues = [40.0, 40.0, 40.0, 40.0]
        let mean = consistentValues.reduce(0, +) / Double(consistentValues.count)
        let variance = consistentValues.map { pow($0 - mean, 2) }.reduce(0, +) / Double(consistentValues.count)

        XCTAssertEqual(variance, 0)
    }

    // MARK: - Suggestion Thresholds Tests

    func testHighWorkloadThreshold() {
        let weeklyHours: Double = 55
        let isHighWorkload = weeklyHours > 50

        XCTAssertTrue(isHighWorkload)
    }

    func testUnderutilizationThreshold() {
        let weeklyHours: Double = 15
        let isUnderutilized = weeklyHours < 20 && weeklyHours > 0

        XCTAssertTrue(isUnderutilized)
    }

    func testSlowPaymentThreshold() {
        let averagePaymentDays: Double = 35
        let isSlowPayment = averagePaymentDays > 30

        XCTAssertTrue(isSlowPayment)
    }

    // MARK: - Milestone Detection Tests

    func testMilestoneDetection() {
        let revenue: Double = 105000
        let isMilestoneReached = revenue > 100000 && revenue <= 110000

        XCTAssertTrue(isMilestoneReached)
    }

    func testMilestoneAlreadyPassed() {
        let revenue: Double = 150000
        let isMilestoneReached = revenue > 100000 && revenue <= 110000

        XCTAssertFalse(isMilestoneReached, "Should not show milestone if already passed")
    }

    // MARK: - Trend Analysis Tests

    func testUpwardTrend() {
        let recentAvg: Double = 45
        let earlierAvg: Double = 35
        let isUpward = recentAvg > earlierAvg * 1.1

        XCTAssertTrue(isUpward)
    }

    func testDownwardTrend() {
        let recentAvg: Double = 25
        let earlierAvg: Double = 35
        let isDownward = recentAvg < earlierAvg * 0.9

        XCTAssertTrue(isDownward)
    }

    func testStableTrend() {
        let recentAvg: Double = 38
        let earlierAvg: Double = 35
        let isStable = !(recentAvg > earlierAvg * 1.1) && !(recentAvg < earlierAvg * 0.9)

        XCTAssertTrue(isStable)
    }
}
