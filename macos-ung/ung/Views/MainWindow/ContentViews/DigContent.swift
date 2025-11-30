//
//  DigContent.swift
//  ung
//
//  Dig - Idea Analysis & Incubation View
//  Transform raw ideas into actionable plans with AI-powered analysis
//

import SwiftUI

struct DigContent: View {
    @EnvironmentObject var appState: AppState
    @StateObject private var digState = DigState()

    @State private var ideaText = ""
    @State private var showingSettings = false
    @State private var selectedTab: DigTab = .overview

    enum DigTab: String, CaseIterable {
        case overview = "Overview"
        case analysis = "Analysis"
        case execution = "Execution"
        case marketing = "Marketing"
        case revenue = "Revenue"
        case alternatives = "Alternatives"

        var icon: String {
            switch self {
            case .overview: return "sparkles"
            case .analysis: return "brain.head.profile"
            case .execution: return "hammer.fill"
            case .marketing: return "megaphone.fill"
            case .revenue: return "chart.line.uptrend.xyaxis"
            case .alternatives: return "arrow.triangle.branch"
            }
        }
    }

    var body: some View {
        HSplitView {
            // Left sidebar - Sessions list
            sessionsSidebar
                .frame(minWidth: 200, idealWidth: 250, maxWidth: 300)

            // Main content
            if digState.currentSession != nil {
                sessionDetailView
            } else {
                newIdeaView
            }
        }
        .task {
            await digState.loadSessions(appState: appState)
        }
        .sheet(isPresented: $showingSettings) {
            APIKeySettingsView(digState: digState)
        }
    }

    // MARK: - Sessions Sidebar

    private var sessionsSidebar: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("Ideas")
                    .font(Design.Typography.headingSmall)
                Spacer()
                Button(action: { digState.currentSession = nil }) {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 20))
                        .foregroundColor(Design.Colors.brand)
                }
                .buttonStyle(.plain)
                .help("New Idea")
            }
            .padding()

            Divider()

            // Sessions list
            if digState.sessions.isEmpty {
                VStack(spacing: Design.Spacing.md) {
                    Image(systemName: "lightbulb")
                        .font(.system(size: 40))
                        .foregroundColor(.secondary)
                    Text("No ideas yet")
                        .font(Design.Typography.bodyMedium)
                        .foregroundColor(.secondary)
                    Text("Type your idea and let AI analyze it")
                        .font(Design.Typography.bodySmall)
                        .foregroundColor(.tertiary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .padding()
            } else {
                ScrollView {
                    LazyVStack(spacing: 2) {
                        ForEach(digState.sessions) { session in
                            SessionRow(
                                session: session,
                                isSelected: digState.currentSession?.id == session.id,
                                onSelect: {
                                    Task {
                                        await digState.loadSession(id: session.id!, appState: appState)
                                    }
                                },
                                onDelete: {
                                    Task {
                                        await digState.deleteSession(id: session.id!, appState: appState)
                                    }
                                }
                            )
                        }
                    }
                    .padding(.horizontal, Design.Spacing.sm)
                }
            }

            Divider()

            // Settings button
            Button(action: { showingSettings = true }) {
                HStack {
                    Image(systemName: "key.fill")
                    Text("API Keys")
                    Spacer()
                    if digState.hasOpenAIKey || digState.hasClaudeKey {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundColor(.green)
                    }
                }
                .padding()
            }
            .buttonStyle(.plain)
        }
        .background(Design.Colors.cardBackground)
    }

    // MARK: - New Idea View

    private var newIdeaView: some View {
        VStack(spacing: Design.Spacing.xl) {
            Spacer()

            // Hero section
            VStack(spacing: Design.Spacing.md) {
                Image(systemName: "lightbulb.max.fill")
                    .font(.system(size: 60))
                    .foregroundStyle(
                        LinearGradient(
                            colors: [.yellow, .orange],
                            startPoint: .top,
                            endPoint: .bottom
                        )
                    )

                Text("Dig Into Your Idea")
                    .font(.system(size: 32, weight: .bold))

                Text("Get comprehensive analysis from multiple perspectives:\nFirst Principles, Design, Marketing, Technical, and Financial")
                    .font(Design.Typography.bodyMedium)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
            }

            // Input area
            VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                Text("What's your idea?")
                    .font(Design.Typography.labelMedium)
                    .foregroundColor(.secondary)

                TextEditor(text: $ideaText)
                    .font(Design.Typography.bodyMedium)
                    .frame(height: 120)
                    .padding(Design.Spacing.sm)
                    .background(Design.Colors.inputBackground)
                    .cornerRadius(Design.Radius.md)
                    .overlay(
                        RoundedRectangle(cornerRadius: Design.Radius.md)
                            .stroke(Design.Colors.border, lineWidth: 1)
                    )

                Text("Be as detailed as possible - the more context, the better the analysis")
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(.tertiary)
            }
            .frame(maxWidth: 600)

            // Analyze button
            Button(action: startAnalysis) {
                HStack(spacing: Design.Spacing.sm) {
                    if digState.isAnalyzing {
                        ProgressView()
                            .scaleEffect(0.8)
                            .progressViewStyle(CircularProgressViewStyle(tint: .white))
                    } else {
                        Image(systemName: "sparkles")
                    }
                    Text(digState.isAnalyzing ? "Analyzing..." : "Analyze Idea")
                        .fontWeight(.semibold)
                }
                .foregroundColor(.white)
                .padding(.horizontal, Design.Spacing.xl)
                .padding(.vertical, Design.Spacing.md)
                .background(
                    LinearGradient(
                        colors: ideaText.isEmpty ? [.gray] : [Design.Colors.brand, Design.Colors.brand.opacity(0.8)],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                )
                .cornerRadius(Design.Radius.lg)
                .shadow(color: Design.Colors.brand.opacity(0.3), radius: 10, y: 5)
            }
            .buttonStyle(.plain)
            .disabled(ideaText.isEmpty || digState.isAnalyzing)

            // Analysis progress
            if digState.isAnalyzing, let progress = digState.progress {
                AnalysisProgressView(progress: progress)
                    .frame(maxWidth: 500)
            }

            Spacer()

            // Features preview
            featuresPreview
        }
        .padding(Design.Spacing.xl)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Design.Colors.windowBackground)
    }

    private var featuresPreview: some View {
        HStack(spacing: Design.Spacing.xl) {
            FeatureCard(icon: "atom", title: "First Principles", description: "Elon-style analysis", color: .purple)
            FeatureCard(icon: "paintbrush.fill", title: "Design", description: "UX perspective", color: .pink)
            FeatureCard(icon: "megaphone.fill", title: "Marketing", description: "Go-to-market", color: .orange)
            FeatureCard(icon: "gearshape.2.fill", title: "Technical", description: "Architecture", color: .blue)
            FeatureCard(icon: "chart.line.uptrend.xyaxis", title: "Financial", description: "Revenue model", color: .green)
        }
        .padding()
    }

    // MARK: - Session Detail View

    private var sessionDetailView: some View {
        VStack(spacing: 0) {
            // Header
            sessionHeader

            Divider()

            // Tab bar
            tabBar

            Divider()

            // Content based on selected tab
            ScrollView {
                switch selectedTab {
                case .overview:
                    overviewTab
                case .analysis:
                    analysisTab
                case .execution:
                    executionTab
                case .marketing:
                    marketingTab
                case .revenue:
                    revenueTab
                case .alternatives:
                    alternativesTab
                }
            }
        }
        .background(Design.Colors.windowBackground)
    }

    private var sessionHeader: some View {
        HStack(spacing: Design.Spacing.md) {
            VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
                Text(digState.currentSession?.title ?? "Untitled Idea")
                    .font(Design.Typography.headingMedium)

                if let score = digState.currentSession?.overallScore {
                    HStack(spacing: Design.Spacing.sm) {
                        ScoreBadge(score: score)

                        if let recommendation = digState.currentSession?.recommendation {
                            RecommendationBadge(recommendation: recommendation)
                        }
                    }
                }
            }

            Spacer()

            // Export button
            Menu {
                Button("Export as JSON") {
                    Task { await exportSession(format: "json") }
                }
                Button("Export as Markdown") {
                    Task { await exportSession(format: "markdown") }
                }
            } label: {
                Label("Export", systemImage: "square.and.arrow.up")
            }

            // Generate images button
            if digState.currentSession?.marketing != nil {
                Button(action: { Task { await digState.generateImages(appState: appState) } }) {
                    Label("Generate Images", systemImage: "photo.fill")
                }
            }
        }
        .padding()
    }

    private var tabBar: some View {
        HStack(spacing: 0) {
            ForEach(DigTab.allCases, id: \.self) { tab in
                Button(action: { selectedTab = tab }) {
                    HStack(spacing: Design.Spacing.xs) {
                        Image(systemName: tab.icon)
                        Text(tab.rawValue)
                    }
                    .font(Design.Typography.labelMedium)
                    .foregroundColor(selectedTab == tab ? Design.Colors.brand : .secondary)
                    .padding(.vertical, Design.Spacing.sm)
                    .padding(.horizontal, Design.Spacing.md)
                    .background(
                        selectedTab == tab
                            ? Design.Colors.brand.opacity(0.1)
                            : Color.clear
                    )
                    .cornerRadius(Design.Radius.sm)
                }
                .buttonStyle(.plain)
            }
            Spacer()
        }
        .padding(.horizontal)
        .padding(.vertical, Design.Spacing.xs)
    }

    // MARK: - Tab Content Views

    private var overviewTab: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.lg) {
            // Original Idea
            SectionCard(title: "Original Idea", icon: "lightbulb") {
                Text(digState.currentSession?.rawIdea ?? "")
                    .font(Design.Typography.bodyMedium)
            }

            // Refined Idea
            if let refinedIdea = digState.currentSession?.refinedIdea, !refinedIdea.isEmpty {
                SectionCard(title: "Refined Idea", icon: "sparkles") {
                    Text(refinedIdea)
                        .font(Design.Typography.bodyMedium)
                }
            }

            // Quick scores
            if let analyses = digState.currentSession?.analyses, !analyses.isEmpty {
                SectionCard(title: "Perspective Scores", icon: "chart.bar") {
                    HStack(spacing: Design.Spacing.md) {
                        ForEach(analyses) { analysis in
                            VStack(spacing: Design.Spacing.xs) {
                                ZStack {
                                    Circle()
                                        .fill(analysis.perspective.color.opacity(0.2))
                                        .frame(width: 60, height: 60)

                                    VStack(spacing: 2) {
                                        Image(systemName: analysis.perspective.icon)
                                            .font(.system(size: 16))
                                            .foregroundColor(analysis.perspective.color)
                                        if let score = analysis.score {
                                            Text("\(Int(score))")
                                                .font(.system(size: 14, weight: .bold))
                                        }
                                    }
                                }

                                Text(analysis.perspective.displayName)
                                    .font(Design.Typography.labelSmall)
                                    .foregroundColor(.secondary)
                            }
                        }
                    }
                }
            }
        }
        .padding()
    }

    private var analysisTab: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.lg) {
            if let analyses = digState.currentSession?.analyses {
                ForEach(analyses) { analysis in
                    AnalysisCard(analysis: analysis)
                }
            }
        }
        .padding()
    }

    private var executionTab: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.lg) {
            if let plan = digState.currentSession?.executionPlan {
                // Summary
                if let summary = plan.summary {
                    SectionCard(title: "Executive Summary", icon: "doc.text") {
                        Text(summary)
                    }
                }

                // MVP Scope
                if let mvp = plan.mvpScope {
                    SectionCard(title: "MVP Scope", icon: "1.circle.fill") {
                        Text(mvp)
                    }
                }

                // Tech Stack
                if let techStack = plan.techStack {
                    SectionCard(title: "Recommended Tech Stack", icon: "server.rack") {
                        Text(techStack)
                            .font(.system(.body, design: .monospaced))
                    }
                }

                // LLM Prompt
                if let llmPrompt = plan.llmPrompt {
                    SectionCard(title: "LLM-Ready Prompt", icon: "text.bubble") {
                        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                            Text(llmPrompt)
                                .font(.system(.body, design: .monospaced))
                                .textSelection(.enabled)

                            Button(action: { copyToClipboard(llmPrompt) }) {
                                Label("Copy to Clipboard", systemImage: "doc.on.doc")
                            }
                        }
                    }
                }
            } else {
                emptyStateView("No execution plan yet")
            }
        }
        .padding()
    }

    private var marketingTab: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.lg) {
            if let marketing = digState.currentSession?.marketing {
                // Value Proposition
                if let vp = marketing.valueProposition {
                    SectionCard(title: "Value Proposition", icon: "star.fill") {
                        Text(vp)
                            .font(.system(size: 18, weight: .medium))
                    }
                }

                // Taglines
                if !marketing.taglinesArray.isEmpty {
                    SectionCard(title: "Taglines", icon: "quote.bubble") {
                        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                            ForEach(marketing.taglinesArray, id: \.self) { tagline in
                                HStack {
                                    Text("\"" + tagline + "\"")
                                        .font(.system(size: 16, weight: .medium, design: .serif))
                                        .italic()
                                    Spacer()
                                    Button(action: { copyToClipboard(tagline) }) {
                                        Image(systemName: "doc.on.doc")
                                    }
                                    .buttonStyle(.plain)
                                }
                                .padding()
                                .background(Design.Colors.cardBackground)
                                .cornerRadius(Design.Radius.sm)
                            }
                        }
                    }
                }

                // Elevator Pitch
                if let pitch = marketing.elevatorPitch {
                    SectionCard(title: "Elevator Pitch", icon: "person.wave.2") {
                        Text(pitch)
                    }
                }

                // Headlines
                if !marketing.headlinesArray.isEmpty {
                    SectionCard(title: "Headlines", icon: "textformat.size") {
                        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                            ForEach(marketing.headlinesArray, id: \.self) { headline in
                                Text(headline)
                                    .font(Design.Typography.headingSmall)
                                    .padding()
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                    .background(Design.Colors.cardBackground)
                                    .cornerRadius(Design.Radius.sm)
                            }
                        }
                    }
                }

                // Launch Strategy
                if let launch = marketing.launchStrategy {
                    SectionCard(title: "Launch Strategy", icon: "paperplane.fill") {
                        Text(launch)
                    }
                }
            } else {
                emptyStateView("No marketing materials yet")
            }
        }
        .padding()
    }

    private var revenueTab: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.lg) {
            if let revenue = digState.currentSession?.revenueProjection {
                // Recommended Pricing
                if let price = revenue.recommendedPrice {
                    SectionCard(title: "Recommended Pricing", icon: "dollarsign.circle.fill") {
                        VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                            Text(price)
                                .font(.system(size: 24, weight: .bold))
                                .foregroundColor(.green)

                            if let rationale = revenue.pricingRationale {
                                Text(rationale)
                                    .font(Design.Typography.bodySmall)
                                    .foregroundColor(.secondary)
                            }
                        }
                    }
                }

                // Market Size
                if let marketSize = revenue.marketSize {
                    SectionCard(title: "Market Size", icon: "chart.pie.fill") {
                        Text(marketSize)
                            .font(.system(.body, design: .monospaced))
                    }
                }

                // Break Even
                if let breakEven = revenue.breakEvenAnalysis {
                    SectionCard(title: "Break-Even Analysis", icon: "equal.circle.fill") {
                        Text(breakEven)
                    }
                }

                // Key Metrics
                if let metrics = revenue.keyMetrics {
                    SectionCard(title: "Key Metrics to Track", icon: "gauge") {
                        Text(metrics)
                            .font(.system(.body, design: .monospaced))
                    }
                }
            } else {
                emptyStateView("No revenue projections yet")
            }
        }
        .padding()
    }

    private var alternativesTab: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.lg) {
            if let alternatives = digState.currentSession?.alternatives, !alternatives.isEmpty {
                ForEach(alternatives) { alternative in
                    AlternativeCard(alternative: alternative)
                }
            } else {
                emptyStateView("No alternative ideas generated")
            }
        }
        .padding()
    }

    // MARK: - Helper Views

    private func emptyStateView(_ message: String) -> some View {
        VStack(spacing: Design.Spacing.md) {
            Image(systemName: "doc.text")
                .font(.system(size: 40))
                .foregroundColor(.secondary)
            Text(message)
                .font(Design.Typography.bodyMedium)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding()
    }

    // MARK: - Actions

    private func startAnalysis() {
        guard !ideaText.isEmpty else { return }
        Task {
            await digState.startAnalysis(idea: ideaText, appState: appState)
            ideaText = ""
        }
    }

    private func exportSession(format: String) async {
        guard let data = await digState.exportSession(format: format, appState: appState) else { return }

        let panel = NSSavePanel()
        panel.allowedContentTypes = format == "json" ? [.json] : [.plainText]
        panel.nameFieldStringValue = "dig-analysis.\(format == "json" ? "json" : "md")"

        if panel.runModal() == .OK, let url = panel.url {
            try? data.write(to: url)
        }
    }

    private func copyToClipboard(_ text: String) {
        #if os(macOS)
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(text, forType: .string)
        #endif
    }
}

// MARK: - Supporting Views

struct SessionRow: View {
    let session: DigSession
    let isSelected: Bool
    let onSelect: () -> Void
    let onDelete: () -> Void

    var body: some View {
        Button(action: onSelect) {
            HStack(spacing: Design.Spacing.sm) {
                // Status indicator
                Circle()
                    .fill(statusColor)
                    .frame(width: 8, height: 8)

                VStack(alignment: .leading, spacing: 2) {
                    Text(session.title ?? "Untitled")
                        .font(Design.Typography.labelMedium)
                        .lineLimit(1)

                    if let score = session.overallScore {
                        Text("Score: \(Int(score))")
                            .font(Design.Typography.labelSmall)
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                if session.status == .analyzing {
                    ProgressView()
                        .scaleEffect(0.6)
                }
            }
            .padding(Design.Spacing.sm)
            .background(isSelected ? Design.Colors.brand.opacity(0.1) : Color.clear)
            .cornerRadius(Design.Radius.sm)
        }
        .buttonStyle(.plain)
        .contextMenu {
            Button(role: .destructive, action: onDelete) {
                Label("Delete", systemImage: "trash")
            }
        }
    }

    private var statusColor: Color {
        switch session.status {
        case .completed: return .green
        case .analyzing: return .orange
        case .failed: return .red
        case .pending: return .gray
        }
    }
}

struct FeatureCard: View {
    let icon: String
    let title: String
    let description: String
    let color: Color

    var body: some View {
        VStack(spacing: Design.Spacing.sm) {
            Image(systemName: icon)
                .font(.system(size: 24))
                .foregroundColor(color)

            Text(title)
                .font(Design.Typography.labelMedium)

            Text(description)
                .font(Design.Typography.labelSmall)
                .foregroundColor(.secondary)
        }
        .frame(width: 100)
        .padding()
        .background(Design.Colors.cardBackground)
        .cornerRadius(Design.Radius.md)
    }
}

struct AnalysisProgressView: View {
    let progress: DigProgressResponse

    var body: some View {
        VStack(spacing: Design.Spacing.md) {
            ProgressView(value: Double(progress.progress), total: 100)
                .progressViewStyle(LinearProgressViewStyle(tint: Design.Colors.brand))

            HStack {
                Image(systemName: "brain.head.profile")
                    .foregroundColor(Design.Colors.brand)
                Text(progress.message)
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(.secondary)
            }
        }
        .padding()
        .background(Design.Colors.cardBackground)
        .cornerRadius(Design.Radius.md)
    }
}

struct ScoreBadge: View {
    let score: Double

    var color: Color {
        switch score {
        case 75...: return .green
        case 55..<75: return .yellow
        case 40..<55: return .orange
        default: return .red
        }
    }

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: "star.fill")
                .font(.system(size: 10))
            Text("\(Int(score))")
                .font(.system(size: 12, weight: .bold))
        }
        .foregroundColor(.white)
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(color)
        .cornerRadius(Design.Radius.xs)
    }
}

struct RecommendationBadge: View {
    let recommendation: DigRecommendation

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: recommendation.icon)
                .font(.system(size: 10))
            Text(recommendation.displayName)
                .font(.system(size: 12, weight: .medium))
        }
        .foregroundColor(recommendation.color)
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(recommendation.color.opacity(0.1))
        .cornerRadius(Design.Radius.xs)
    }
}

struct SectionCard<Content: View>: View {
    let title: String
    let icon: String
    @ViewBuilder let content: Content

    var body: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.md) {
            HStack {
                Image(systemName: icon)
                    .foregroundColor(Design.Colors.brand)
                Text(title)
                    .font(Design.Typography.headingSmall)
            }

            content
                .font(Design.Typography.bodyMedium)
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Design.Colors.cardBackground)
        .cornerRadius(Design.Radius.md)
    }
}

struct AnalysisCard: View {
    let analysis: DigAnalysis

    @State private var isExpanded = false

    var body: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.md) {
            // Header
            Button(action: { withAnimation { isExpanded.toggle() } }) {
                HStack {
                    Image(systemName: analysis.perspective.icon)
                        .foregroundColor(analysis.perspective.color)
                        .frame(width: 30)

                    Text(analysis.perspective.displayName)
                        .font(Design.Typography.headingSmall)

                    Spacer()

                    if let score = analysis.score {
                        ScoreBadge(score: score)
                    }

                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .foregroundColor(.secondary)
                }
            }
            .buttonStyle(.plain)

            // Summary
            if let summary = analysis.summary {
                Text(summary)
                    .font(Design.Typography.bodyMedium)
                    .foregroundColor(.secondary)
            }

            // Expanded content
            if isExpanded {
                Divider()

                // Strengths
                if !analysis.strengthsArray.isEmpty {
                    SWOTSection(title: "Strengths", items: analysis.strengthsArray, color: .green)
                }

                // Weaknesses
                if !analysis.weaknessesArray.isEmpty {
                    SWOTSection(title: "Weaknesses", items: analysis.weaknessesArray, color: .red)
                }

                // Opportunities
                if !analysis.opportunitiesArray.isEmpty {
                    SWOTSection(title: "Opportunities", items: analysis.opportunitiesArray, color: .blue)
                }

                // Recommendations
                if !analysis.recommendationsArray.isEmpty {
                    SWOTSection(title: "Recommendations", items: analysis.recommendationsArray, color: .purple)
                }
            }
        }
        .padding()
        .background(Design.Colors.cardBackground)
        .cornerRadius(Design.Radius.md)
    }
}

struct SWOTSection: View {
    let title: String
    let items: [String]
    let color: Color

    var body: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            Text(title)
                .font(Design.Typography.labelMedium)
                .foregroundColor(color)

            ForEach(items, id: \.self) { item in
                HStack(alignment: .top, spacing: Design.Spacing.xs) {
                    Circle()
                        .fill(color)
                        .frame(width: 6, height: 6)
                        .padding(.top, 6)
                    Text(item)
                        .font(Design.Typography.bodySmall)
                }
            }
        }
    }
}

struct AlternativeCard: View {
    let alternative: DigAlternative

    var body: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.md) {
            HStack {
                Text(alternative.alternativeIdea)
                    .font(Design.Typography.headingSmall)

                Spacer()

                if let score = alternative.viabilityScore {
                    ScoreBadge(score: score)
                }
            }

            if let rationale = alternative.rationale {
                Text(rationale)
                    .font(Design.Typography.bodyMedium)
                    .foregroundColor(.secondary)
            }

            HStack(spacing: Design.Spacing.md) {
                if let effort = alternative.effortLevel {
                    Label(effort.capitalized, systemImage: "hammer")
                        .font(Design.Typography.labelSmall)
                        .foregroundColor(.secondary)
                }

                if let potential = alternative.potential {
                    Label(potential.replacingOccurrences(of: "_", with: " ").capitalized, systemImage: "arrow.up.right")
                        .font(Design.Typography.labelSmall)
                        .foregroundColor(alternative.potentialColor)
                }
            }
        }
        .padding()
        .background(Design.Colors.cardBackground)
        .cornerRadius(Design.Radius.md)
    }
}

// MARK: - API Key Settings View

struct APIKeySettingsView: View {
    @ObservedObject var digState: DigState
    @Environment(\.dismiss) var dismiss

    @State private var openAIKey = ""
    @State private var claudeKey = ""
    @State private var showOpenAIKey = false
    @State private var showClaudeKey = false

    var body: some View {
        VStack(spacing: Design.Spacing.lg) {
            Text("AI API Keys")
                .font(Design.Typography.headingMedium)

            Text("Enter your own API keys to analyze ideas directly without using the server.")
                .font(Design.Typography.bodySmall)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)

            VStack(alignment: .leading, spacing: Design.Spacing.md) {
                // OpenAI
                VStack(alignment: .leading, spacing: Design.Spacing.xs) {
                    HStack {
                        Text("OpenAI API Key")
                            .font(Design.Typography.labelMedium)
                        if digState.hasOpenAIKey {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundColor(.green)
                        }
                    }

                    HStack {
                        if showOpenAIKey {
                            TextField("sk-...", text: $openAIKey)
                        } else {
                            SecureField("sk-...", text: $openAIKey)
                        }

                        Button(action: { showOpenAIKey.toggle() }) {
                            Image(systemName: showOpenAIKey ? "eye.slash" : "eye")
                        }
                        .buttonStyle(.plain)

                        if !openAIKey.isEmpty {
                            Button("Save") {
                                digState.saveOpenAIKey(openAIKey)
                                openAIKey = ""
                            }
                        }

                        if digState.hasOpenAIKey {
                            Button(role: .destructive, action: digState.deleteOpenAIKey) {
                                Image(systemName: "trash")
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .textFieldStyle(.roundedBorder)
                }

                Divider()

                // Claude
                VStack(alignment: .leading, spacing: Design.Spacing.xs) {
                    HStack {
                        Text("Claude API Key")
                            .font(Design.Typography.labelMedium)
                        if digState.hasClaudeKey {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundColor(.green)
                        }
                    }

                    HStack {
                        if showClaudeKey {
                            TextField("sk-ant-...", text: $claudeKey)
                        } else {
                            SecureField("sk-ant-...", text: $claudeKey)
                        }

                        Button(action: { showClaudeKey.toggle() }) {
                            Image(systemName: showClaudeKey ? "eye.slash" : "eye")
                        }
                        .buttonStyle(.plain)

                        if !claudeKey.isEmpty {
                            Button("Save") {
                                digState.saveClaudeKey(claudeKey)
                                claudeKey = ""
                            }
                        }

                        if digState.hasClaudeKey {
                            Button(role: .destructive, action: digState.deleteClaudeKey) {
                                Image(systemName: "trash")
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .textFieldStyle(.roundedBorder)
                }
            }
            .padding()
            .background(Design.Colors.cardBackground)
            .cornerRadius(Design.Radius.md)

            Text("Keys are stored securely in your system Keychain")
                .font(Design.Typography.labelSmall)
                .foregroundColor(.tertiary)

            Button("Done") { dismiss() }
                .buttonStyle(.borderedProminent)
        }
        .padding(Design.Spacing.xl)
        .frame(width: 450)
    }
}

#Preview {
    DigContent()
        .environmentObject(AppState())
        .frame(width: 1000, height: 700)
}
