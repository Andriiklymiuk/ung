//
//  SmartGoalsService.swift
//  ung
//
//  AI-powered goal suggestions based on user's historical data
//  Analyzes patterns and provides personalized recommendations
//

import Foundation

// MARK: - Goal Suggestion Types

struct GoalSuggestion: Identifiable {
    let id = UUID()
    let type: GoalType
    let title: String
    let description: String
    let suggestedValue: Double
    let currentValue: Double
    let confidence: ConfidenceLevel
    let reasoning: String
    let icon: String
    let trend: TrendDirection

    enum GoalType {
        case weeklyHours
        case monthlyIncome
        case yearlyIncome
        case hourlyRate
        case clientsPerMonth
        case projectsPerMonth
    }

    enum ConfidenceLevel: String {
        case high = "High"
        case medium = "Medium"
        case low = "Low"

        var color: String {
            switch self {
            case .high: return "green"
            case .medium: return "orange"
            case .low: return "gray"
            }
        }
    }

    enum TrendDirection {
        case up
        case down
        case stable

        var icon: String {
            switch self {
            case .up: return "arrow.up.right"
            case .down: return "arrow.down.right"
            case .stable: return "arrow.right"
            }
        }
    }
}

struct GoalInsight: Identifiable {
    let id = UUID()
    let title: String
    let message: String
    let type: InsightType
    let actionSuggestion: String?
    let icon: String

    enum InsightType {
        case positive
        case warning
        case info
        case achievement
    }
}

// MARK: - Smart Goals Service

class SmartGoalsService {
    static let shared = SmartGoalsService()

    private let database = DatabaseService.shared

    private init() {}

    // MARK: - Generate Goal Suggestions

    func generateSuggestions() async -> [GoalSuggestion] {
        var suggestions: [GoalSuggestion] = []

        do {
            // Gather historical data
            let historicalData = try await gatherHistoricalData()

            // Generate weekly hours suggestion
            if let weeklyHoursSuggestion = generateWeeklyHoursSuggestion(from: historicalData) {
                suggestions.append(weeklyHoursSuggestion)
            }

            // Generate monthly income suggestion
            if let monthlyIncomeSuggestion = generateMonthlyIncomeSuggestion(from: historicalData) {
                suggestions.append(monthlyIncomeSuggestion)
            }

            // Generate yearly income suggestion
            if let yearlyIncomeSuggestion = generateYearlyIncomeSuggestion(from: historicalData) {
                suggestions.append(yearlyIncomeSuggestion)
            }

            // Generate hourly rate suggestion
            if let rateSuggestion = generateRateSuggestion(from: historicalData) {
                suggestions.append(rateSuggestion)
            }

        } catch {
            print("[SmartGoalsService] Error generating suggestions: \(error)")
        }

        return suggestions
    }

    // MARK: - Generate Insights

    func generateInsights() async -> [GoalInsight] {
        var insights: [GoalInsight] = []

        do {
            let data = try await gatherHistoricalData()

            // Check for positive trends
            if data.revenueGrowthRate > 0.1 {
                insights.append(GoalInsight(
                    title: "Revenue Growing",
                    message: "Your revenue has grown \(Int(data.revenueGrowthRate * 100))% compared to last period. Keep it up!",
                    type: .positive,
                    actionSuggestion: "Consider raising your rates",
                    icon: "chart.line.uptrend.xyaxis"
                ))
            }

            // Check for warnings
            if data.averageWeeklyHours > 50 {
                insights.append(GoalInsight(
                    title: "High Workload",
                    message: "You're averaging \(Int(data.averageWeeklyHours)) hours/week. Consider work-life balance.",
                    type: .warning,
                    actionSuggestion: "Try setting boundaries or increasing your rate",
                    icon: "exclamationmark.triangle"
                ))
            }

            // Check for underutilization
            if data.averageWeeklyHours < 20 && data.averageWeeklyHours > 0 {
                insights.append(GoalInsight(
                    title: "Capacity Available",
                    message: "You're working \(Int(data.averageWeeklyHours)) hours/week. You have room for more clients.",
                    type: .info,
                    actionSuggestion: "Consider marketing or reaching out to past clients",
                    icon: "clock.badge.checkmark"
                ))
            }

            // Check for payment patterns
            if data.averagePaymentDays > 30 {
                insights.append(GoalInsight(
                    title: "Slow Payments",
                    message: "Clients take \(Int(data.averagePaymentDays)) days on average to pay invoices.",
                    type: .warning,
                    actionSuggestion: "Consider shorter payment terms or late fees",
                    icon: "calendar.badge.exclamationmark"
                ))
            }

            // Milestone achievements
            if data.totalRevenue > 100000 && data.totalRevenue <= 110000 {
                insights.append(GoalInsight(
                    title: "Milestone Reached!",
                    message: "You've earned over $100,000. That's a major achievement!",
                    type: .achievement,
                    actionSuggestion: nil,
                    icon: "star.fill"
                ))
            }

            // Consistency check
            if data.hoursVariance < 5 {
                insights.append(GoalInsight(
                    title: "Consistent Schedule",
                    message: "Your work hours are very consistent. Great discipline!",
                    type: .positive,
                    actionSuggestion: nil,
                    icon: "checkmark.seal.fill"
                ))
            }

        } catch {
            print("[SmartGoalsService] Error generating insights: \(error)")
        }

        return insights
    }

    // MARK: - Historical Data Analysis

    private struct HistoricalData {
        var averageWeeklyHours: Double = 0
        var hoursVariance: Double = 0
        var weeklyHoursTrend: [Double] = []

        var averageMonthlyIncome: Double = 0
        var monthlyIncomeTrend: [Double] = []

        var averageYearlyIncome: Double = 0
        var revenueGrowthRate: Double = 0

        var averageHourlyRate: Double = 0
        var effectiveRate: Double = 0

        var totalRevenue: Double = 0
        var totalHours: Double = 0

        var averagePaymentDays: Double = 0
        var clientCount: Int = 0
        var projectCount: Int = 0
    }

    private func gatherHistoricalData() async throws -> HistoricalData {
        var data = HistoricalData()

        // Get sessions from last 12 weeks for hours analysis
        let sessions = try await database.getWeeklySessions()
        let calendar = Calendar.current

        // Calculate weekly hours
        var weeklyHours: [Int: Double] = [:]
        for session in sessions {
            let week = calendar.component(.weekOfYear, from: session.startTime)
            let hours = Double(session.calculatedDuration) / 3600.0
            weeklyHours[week, default: 0] += hours
        }

        if !weeklyHours.isEmpty {
            let hoursArray = Array(weeklyHours.values)
            data.averageWeeklyHours = hoursArray.reduce(0, +) / Double(hoursArray.count)
            data.weeklyHoursTrend = hoursArray.sorted()

            // Calculate variance
            let mean = data.averageWeeklyHours
            let variance = hoursArray.map { pow($0 - mean, 2) }.reduce(0, +) / Double(hoursArray.count)
            data.hoursVariance = sqrt(variance)
        }

        // Get invoices for income analysis
        let invoices = try await database.getRecentInvoices(limit: 100)

        // Group by month
        var monthlyIncome: [String: Double] = [:]
        let monthFormatter = DateFormatter()
        monthFormatter.dateFormat = "yyyy-MM"

        for invoice in invoices where invoice.status == "paid" {
            let monthKey = monthFormatter.string(from: invoice.issuedDate)
            monthlyIncome[monthKey, default: 0] += invoice.amount
        }

        if !monthlyIncome.isEmpty {
            let incomeArray = Array(monthlyIncome.values)
            data.averageMonthlyIncome = incomeArray.reduce(0, +) / Double(incomeArray.count)
            data.monthlyIncomeTrend = incomeArray.sorted()
            data.totalRevenue = incomeArray.reduce(0, +)

            // Calculate growth rate (last 3 months vs previous 3 months)
            if incomeArray.count >= 6 {
                let recent = incomeArray.suffix(3).reduce(0, +)
                let previous = incomeArray.dropLast(3).suffix(3).reduce(0, +)
                if previous > 0 {
                    data.revenueGrowthRate = (recent - previous) / previous
                }
            }
        }

        // Calculate yearly income projection
        data.averageYearlyIncome = data.averageMonthlyIncome * 12

        // Calculate effective hourly rate
        let totalTrackedHours = sessions.reduce(0.0) { $0 + Double($1.calculatedDuration) / 3600.0 }
        data.totalHours = totalTrackedHours
        if totalTrackedHours > 0 {
            data.effectiveRate = data.totalRevenue / totalTrackedHours
        }

        // Get contract rates for average rate
        let contracts = try await database.getContracts()
        if !contracts.isEmpty {
            let rates = contracts.compactMap { $0.hourlyRate }.filter { $0 > 0 }
            if !rates.isEmpty {
                data.averageHourlyRate = rates.reduce(0, +) / Double(rates.count)
            }
        }

        // Client and project counts
        data.clientCount = try await database.getClientCount()
        data.projectCount = contracts.count

        return data
    }

    // MARK: - Individual Suggestion Generators

    private func generateWeeklyHoursSuggestion(from data: HistoricalData) -> GoalSuggestion? {
        guard data.averageWeeklyHours > 0 else { return nil }

        // Suggest 10% increase if trending up, maintain if stable
        let trend: GoalSuggestion.TrendDirection
        let suggestedHours: Double
        let reasoning: String

        if data.weeklyHoursTrend.count >= 4 {
            let recentAvg = data.weeklyHoursTrend.suffix(2).reduce(0, +) / 2
            let earlierAvg = data.weeklyHoursTrend.prefix(2).reduce(0, +) / 2

            if recentAvg > earlierAvg * 1.1 {
                trend = .up
                suggestedHours = min(data.averageWeeklyHours * 1.1, 50) // Cap at 50 hours
                reasoning = "Your hours have been increasing. This goal matches your momentum."
            } else if recentAvg < earlierAvg * 0.9 {
                trend = .down
                suggestedHours = data.averageWeeklyHours // Maintain current
                reasoning = "Your hours have decreased. This goal helps stabilize."
            } else {
                trend = .stable
                suggestedHours = data.averageWeeklyHours * 1.05 // Slight increase
                reasoning = "Your hours are consistent. A small stretch goal."
            }
        } else {
            trend = .stable
            suggestedHours = data.averageWeeklyHours
            reasoning = "Based on your recent work pattern."
        }

        return GoalSuggestion(
            type: .weeklyHours,
            title: "Weekly Hours Goal",
            description: "Target hours per week",
            suggestedValue: round(suggestedHours),
            currentValue: round(data.averageWeeklyHours),
            confidence: data.weeklyHoursTrend.count >= 8 ? .high : .medium,
            reasoning: reasoning,
            icon: "clock.fill",
            trend: trend
        )
    }

    private func generateMonthlyIncomeSuggestion(from data: HistoricalData) -> GoalSuggestion? {
        guard data.averageMonthlyIncome > 0 else { return nil }

        let trend: GoalSuggestion.TrendDirection
        let suggestedIncome: Double
        let reasoning: String

        if data.revenueGrowthRate > 0.1 {
            trend = .up
            suggestedIncome = data.averageMonthlyIncome * 1.15
            reasoning = "You're on a growth trajectory. Aim for 15% above average."
        } else if data.revenueGrowthRate < -0.1 {
            trend = .down
            suggestedIncome = data.averageMonthlyIncome
            reasoning = "Focus on stabilizing at your current average."
        } else {
            trend = .stable
            suggestedIncome = data.averageMonthlyIncome * 1.1
            reasoning = "A 10% increase is a healthy stretch goal."
        }

        return GoalSuggestion(
            type: .monthlyIncome,
            title: "Monthly Income Goal",
            description: "Target revenue per month",
            suggestedValue: round(suggestedIncome / 100) * 100, // Round to nearest 100
            currentValue: round(data.averageMonthlyIncome),
            confidence: data.monthlyIncomeTrend.count >= 6 ? .high : .medium,
            reasoning: reasoning,
            icon: "dollarsign.circle.fill",
            trend: trend
        )
    }

    private func generateYearlyIncomeSuggestion(from data: HistoricalData) -> GoalSuggestion? {
        guard data.averageYearlyIncome > 0 else { return nil }

        let suggestedIncome = data.averageYearlyIncome * 1.2 // 20% growth target
        let trend: GoalSuggestion.TrendDirection = data.revenueGrowthRate > 0 ? .up : .stable

        return GoalSuggestion(
            type: .yearlyIncome,
            title: "Yearly Income Goal",
            description: "Annual revenue target",
            suggestedValue: round(suggestedIncome / 1000) * 1000, // Round to nearest 1000
            currentValue: round(data.averageYearlyIncome),
            confidence: data.monthlyIncomeTrend.count >= 12 ? .high : .low,
            reasoning: "Based on your monthly average, aim for 20% annual growth.",
            icon: "chart.bar.fill",
            trend: trend
        )
    }

    private func generateRateSuggestion(from data: HistoricalData) -> GoalSuggestion? {
        let currentRate = data.averageHourlyRate > 0 ? data.averageHourlyRate : data.effectiveRate

        guard currentRate > 0 else { return nil }

        // Suggest rate based on utilization and market factors
        let suggestedRate: Double
        let reasoning: String
        let trend: GoalSuggestion.TrendDirection

        if data.averageWeeklyHours > 40 {
            // High demand - suggest rate increase
            suggestedRate = currentRate * 1.15
            reasoning = "High demand for your work. You can charge more."
            trend = .up
        } else if data.effectiveRate > currentRate {
            // Effective rate is higher - projects are more profitable
            suggestedRate = data.effectiveRate
            reasoning = "Your effective rate is higher than your listed rate."
            trend = .up
        } else {
            // Standard suggestion
            suggestedRate = currentRate * 1.05
            reasoning = "A modest rate increase to keep up with inflation."
            trend = .stable
        }

        return GoalSuggestion(
            type: .hourlyRate,
            title: "Hourly Rate Goal",
            description: "Target rate per hour",
            suggestedValue: round(suggestedRate),
            currentValue: round(currentRate),
            confidence: data.totalHours > 100 ? .high : .medium,
            reasoning: reasoning,
            icon: "banknote.fill",
            trend: trend
        )
    }
}

// MARK: - Goal Achievement Tracking

extension SmartGoalsService {
    func checkGoalProgress(weeklyHoursGoal: Double, monthlyIncomeGoal: Double) async -> [GoalInsight] {
        var insights: [GoalInsight] = []

        do {
            let data = try await gatherHistoricalData()

            // Check weekly hours progress
            let hoursProgress = data.averageWeeklyHours / weeklyHoursGoal
            if hoursProgress >= 1.0 {
                insights.append(GoalInsight(
                    title: "Hours Goal Met!",
                    message: "You're hitting \(Int(hoursProgress * 100))% of your weekly hours goal.",
                    type: .achievement,
                    actionSuggestion: "Consider increasing your goal",
                    icon: "checkmark.circle.fill"
                ))
            } else if hoursProgress >= 0.8 {
                insights.append(GoalInsight(
                    title: "Almost There",
                    message: "You're at \(Int(hoursProgress * 100))% of your weekly hours goal.",
                    type: .positive,
                    actionSuggestion: nil,
                    icon: "chart.line.uptrend.xyaxis"
                ))
            } else if hoursProgress < 0.5 {
                insights.append(GoalInsight(
                    title: "Hours Behind",
                    message: "You're at \(Int(hoursProgress * 100))% of your weekly hours goal.",
                    type: .warning,
                    actionSuggestion: "Review your goal or schedule more work",
                    icon: "exclamationmark.circle"
                ))
            }

            // Check monthly income progress
            let incomeProgress = data.averageMonthlyIncome / monthlyIncomeGoal
            if incomeProgress >= 1.0 {
                insights.append(GoalInsight(
                    title: "Income Goal Exceeded!",
                    message: "You're earning \(Int(incomeProgress * 100))% of your monthly goal.",
                    type: .achievement,
                    actionSuggestion: "Great work! Consider saving or investing the extra",
                    icon: "star.fill"
                ))
            }

        } catch {
            print("[SmartGoalsService] Error checking progress: \(error)")
        }

        return insights
    }
}
