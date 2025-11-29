//
//  GoalSuggestionsView.swift
//  ung
//
//  AI-powered goal suggestions with insights
//

import SwiftUI

// MARK: - Goal Suggestions Section
struct GoalSuggestionsSection: View {
    @State private var suggestions: [GoalSuggestion] = []
    @State private var insights: [GoalInsight] = []
    @State private var isLoading = true
    @State private var expandedSuggestion: GoalSuggestion.ID?
    @Environment(\.colorScheme) var colorScheme

    let onApplySuggestion: (GoalSuggestion) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.md) {
            // Header
            HStack {
                Image(systemName: "sparkles")
                    .foregroundColor(Design.Colors.warning)
                Text("Smart Suggestions")
                    .font(Design.Typography.headingSmall)

                Spacer()

                if isLoading {
                    ProgressView()
                        .scaleEffect(0.7)
                } else {
                    Button(action: loadSuggestions) {
                        Image(systemName: "arrow.clockwise")
                            .foregroundColor(Design.Colors.textTertiary)
                    }
                    .buttonStyle(.plain)
                }
            }

            if isLoading {
                loadingState
            } else if suggestions.isEmpty && insights.isEmpty {
                emptyState
            } else {
                // Insights cards
                if !insights.isEmpty {
                    insightsSection
                }

                // Goal suggestions
                if !suggestions.isEmpty {
                    suggestionsSection
                }
            }
        }
        .onAppear {
            loadSuggestions()
        }
    }

    private var loadingState: some View {
        VStack(spacing: Design.Spacing.sm) {
            ForEach(0..<3, id: \.self) { _ in
                RoundedRectangle(cornerRadius: Design.Radius.sm)
                    .fill(Design.Colors.backgroundTertiary(colorScheme))
                    .frame(height: 60)
                    .shimmer()
            }
        }
    }

    private var emptyState: some View {
        VStack(spacing: Design.Spacing.sm) {
            Image(systemName: "chart.bar.doc.horizontal")
                .font(.system(size: 32))
                .foregroundColor(Design.Colors.textTertiary)
            Text("Not enough data yet")
                .font(Design.Typography.bodyMedium)
                .foregroundColor(Design.Colors.textSecondary)
            Text("Track more time and invoices to get personalized suggestions")
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textTertiary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(Design.Spacing.lg)
    }

    private var insightsSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
            Text("Insights")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textTertiary)
                .textCase(.uppercase)

            ForEach(insights) { insight in
                InsightCard(insight: insight)
            }
        }
    }

    private var suggestionsSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
            Text("Recommended Goals")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textTertiary)
                .textCase(.uppercase)

            ForEach(suggestions) { suggestion in
                SuggestionCard(
                    suggestion: suggestion,
                    isExpanded: expandedSuggestion == suggestion.id,
                    onTap: {
                        withAnimation(Design.Animation.smooth) {
                            expandedSuggestion = expandedSuggestion == suggestion.id ? nil : suggestion.id
                        }
                    },
                    onApply: {
                        onApplySuggestion(suggestion)
                    }
                )
            }
        }
    }

    private func loadSuggestions() {
        isLoading = true
        Task {
            async let suggestionsTask = SmartGoalsService.shared.generateSuggestions()
            async let insightsTask = SmartGoalsService.shared.generateInsights()

            let (newSuggestions, newInsights) = await (suggestionsTask, insightsTask)

            await MainActor.run {
                withAnimation {
                    suggestions = newSuggestions
                    insights = newInsights
                    isLoading = false
                }
            }
        }
    }
}

// MARK: - Insight Card
struct InsightCard: View {
    let insight: GoalInsight
    @Environment(\.colorScheme) var colorScheme

    private var iconColor: Color {
        switch insight.type {
        case .positive: return Design.Colors.success
        case .warning: return Design.Colors.warning
        case .info: return Design.Colors.info
        case .achievement: return Design.Colors.primary
        }
    }

    private var backgroundColor: Color {
        switch insight.type {
        case .positive: return Design.Colors.successLight
        case .warning: return Design.Colors.warningLight
        case .info: return Design.Colors.infoLight
        case .achievement: return Design.Colors.primaryLight
        }
    }

    var body: some View {
        HStack(alignment: .top, spacing: Design.Spacing.sm) {
            // Icon
            ZStack {
                Circle()
                    .fill(backgroundColor)
                    .frame(width: 32, height: 32)
                Image(systemName: insight.icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(iconColor)
            }

            // Content
            VStack(alignment: .leading, spacing: 4) {
                Text(insight.title)
                    .font(Design.Typography.labelMedium)
                    .foregroundColor(Design.Colors.textPrimary)

                Text(insight.message)
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textSecondary)

                if let action = insight.actionSuggestion {
                    Text(action)
                        .font(Design.Typography.labelSmall)
                        .foregroundColor(iconColor)
                        .padding(.top, 2)
                }
            }

            Spacer()
        }
        .padding(Design.Spacing.sm)
        .background(
            RoundedRectangle(cornerRadius: Design.Radius.md)
                .fill(Design.Colors.surfaceElevated(colorScheme))
        )
    }
}

// MARK: - Suggestion Card
struct SuggestionCard: View {
    let suggestion: GoalSuggestion
    let isExpanded: Bool
    let onTap: () -> Void
    let onApply: () -> Void
    @Environment(\.colorScheme) var colorScheme

    private var trendColor: Color {
        switch suggestion.trend {
        case .up: return Design.Colors.success
        case .down: return Design.Colors.error
        case .stable: return Design.Colors.textTertiary
        }
    }

    private var confidenceColor: Color {
        switch suggestion.confidence {
        case .high: return Design.Colors.success
        case .medium: return Design.Colors.warning
        case .low: return Design.Colors.textTertiary
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Main content
            Button(action: onTap) {
                HStack(spacing: Design.Spacing.sm) {
                    // Icon
                    ZStack {
                        Circle()
                            .fill(Design.Colors.primaryLight)
                            .frame(width: 40, height: 40)
                        Image(systemName: suggestion.icon)
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundColor(Design.Colors.primary)
                    }

                    // Content
                    VStack(alignment: .leading, spacing: 2) {
                        Text(suggestion.title)
                            .font(Design.Typography.labelMedium)
                            .foregroundColor(Design.Colors.textPrimary)

                        HStack(spacing: Design.Spacing.xs) {
                            Text(formatValue(suggestion.currentValue))
                                .font(Design.Typography.bodySmall)
                                .foregroundColor(Design.Colors.textTertiary)

                            Image(systemName: "arrow.right")
                                .font(.system(size: 10))
                                .foregroundColor(Design.Colors.textTertiary)

                            Text(formatValue(suggestion.suggestedValue))
                                .font(Design.Typography.labelMedium)
                                .foregroundColor(Design.Colors.primary)
                        }
                    }

                    Spacer()

                    // Trend indicator
                    HStack(spacing: 4) {
                        Image(systemName: suggestion.trend.icon)
                            .font(.system(size: 12, weight: .bold))
                        Text(suggestion.confidence.rawValue)
                            .font(Design.Typography.labelSmall)
                    }
                    .foregroundColor(trendColor)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(
                        Capsule()
                            .fill(trendColor.opacity(0.15))
                    )

                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .font(.system(size: 12))
                        .foregroundColor(Design.Colors.textTertiary)
                }
                .padding(Design.Spacing.sm)
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)

            // Expanded content
            if isExpanded {
                VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                    Divider()

                    // Reasoning
                    Text(suggestion.reasoning)
                        .font(Design.Typography.bodySmall)
                        .foregroundColor(Design.Colors.textSecondary)

                    // Comparison
                    HStack(spacing: Design.Spacing.lg) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text("Current")
                                .font(Design.Typography.labelSmall)
                                .foregroundColor(Design.Colors.textTertiary)
                            Text(formatValue(suggestion.currentValue))
                                .font(Design.Typography.headingSmall)
                                .foregroundColor(Design.Colors.textSecondary)
                        }

                        VStack(alignment: .leading, spacing: 2) {
                            Text("Suggested")
                                .font(Design.Typography.labelSmall)
                                .foregroundColor(Design.Colors.textTertiary)
                            Text(formatValue(suggestion.suggestedValue))
                                .font(Design.Typography.headingSmall)
                                .foregroundColor(Design.Colors.primary)
                        }

                        VStack(alignment: .leading, spacing: 2) {
                            Text("Change")
                                .font(Design.Typography.labelSmall)
                                .foregroundColor(Design.Colors.textTertiary)
                            let change = ((suggestion.suggestedValue - suggestion.currentValue) / suggestion.currentValue) * 100
                            Text(String(format: "%+.0f%%", change))
                                .font(Design.Typography.headingSmall)
                                .foregroundColor(change >= 0 ? Design.Colors.success : Design.Colors.error)
                        }
                    }

                    // Apply button
                    Button(action: onApply) {
                        HStack {
                            Image(systemName: "checkmark.circle.fill")
                            Text("Apply This Goal")
                        }
                        .frame(maxWidth: .infinity)
                    }
                    .buttonStyle(DSPrimaryButtonStyle())
                    .padding(.top, Design.Spacing.xs)
                }
                .padding(Design.Spacing.sm)
                .transition(.opacity.combined(with: .move(edge: .top)))
            }
        }
        .background(
            RoundedRectangle(cornerRadius: Design.Radius.md)
                .fill(Design.Colors.surfaceElevated(colorScheme))
        )
    }

    private func formatValue(_ value: Double) -> String {
        switch suggestion.type {
        case .weeklyHours:
            return String(format: "%.0fh/week", value)
        case .monthlyIncome, .yearlyIncome:
            return Formatters.formatCurrency(value)
        case .hourlyRate:
            return Formatters.formatCurrency(value) + "/hr"
        case .clientsPerMonth, .projectsPerMonth:
            return String(format: "%.0f", value)
        }
    }
}

// MARK: - Shimmer Effect
struct ShimmerModifier: ViewModifier {
    @State private var phase: CGFloat = 0

    func body(content: Content) -> some View {
        content
            .overlay(
                LinearGradient(
                    gradient: Gradient(stops: [
                        .init(color: .clear, location: 0),
                        .init(color: .white.opacity(0.3), location: 0.5),
                        .init(color: .clear, location: 1)
                    ]),
                    startPoint: .leading,
                    endPoint: .trailing
                )
                .rotationEffect(.degrees(30))
                .offset(x: phase)
                .mask(content)
            )
            .onAppear {
                withAnimation(.linear(duration: 1.5).repeatForever(autoreverses: false)) {
                    phase = 400
                }
            }
    }
}

extension View {
    func shimmer() -> some View {
        modifier(ShimmerModifier())
    }
}

#Preview {
    GoalSuggestionsSection(onApplySuggestion: { _ in })
        .padding()
        .frame(width: 400)
}
