//
//  NextContent.swift
//  ung
//
//  "What's next?" - The daily driver view for freelancers
//

import SwiftUI

struct NextContent: View {
    @EnvironmentObject var appState: AppState
    @StateObject private var nextState = NextState()
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        ScrollView(showsIndicators: false) {
            VStack(spacing: 20) {
                // Header
                headerSection

                // Main cards
                VStack(spacing: 16) {
                    // WORK card
                    workCard

                    // BILL card
                    if nextState.pendingInvoiceAmount > 0 || nextState.unbilledHours > 0 {
                        billCard
                    }

                    // HUNT card (only if few active gigs)
                    if nextState.activeGigsCount < 3 && nextState.topJobMatch != nil {
                        huntCard
                    }

                    // GOAL card
                    if nextState.currentGoal != nil {
                        goalCard
                    }
                }

                Spacer(minLength: 20)
            }
            .padding(24)
        }
        .onAppear {
            Task {
                await nextState.loadData(appState: appState)
            }
        }
    }

    // MARK: - Header

    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(greeting)
                    .font(.system(size: 28, weight: .bold))
                    .foregroundColor(.primary)

                Text("What's your next move?")
                    .font(.system(size: 15))
                    .foregroundColor(.secondary)
            }

            Spacer()

            // Quick stats
            HStack(spacing: 20) {
                StatBadge(
                    value: appState.formatHours(appState.metrics.weeklyHours),
                    label: "This week",
                    icon: "clock.fill",
                    color: .blue
                )

                StatBadge(
                    value: appState.formatCurrency(appState.metrics.monthlyRevenue),
                    label: "This month",
                    icon: "dollarsign.circle.fill",
                    color: .green
                )
            }
        }
    }

    private var greeting: String {
        let hour = Calendar.current.component(.hour, from: Date())
        if hour < 12 { return "Good morning" }
        if hour < 17 { return "Good afternoon" }
        return "Good evening"
    }

    // MARK: - Work Card

    private var workCard: some View {
        NextCard(
            title: "WORK",
            icon: "hammer.fill",
            color: .blue,
            action: { appState.selectedTab = .tracking }
        ) {
            if let session = appState.activeSession {
                // Active tracking
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Circle()
                            .fill(Color.green)
                            .frame(width: 8, height: 8)
                        Text("Tracking now")
                            .font(.system(size: 12, weight: .medium))
                            .foregroundColor(.green)
                    }

                    Text(session.project)
                        .font(.system(size: 18, weight: .semibold))

                    if !session.client.isEmpty {
                        Text(session.client)
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text(session.formattedDuration)
                            .font(.system(size: 24, weight: .bold, design: .monospaced))
                            .foregroundColor(.blue)

                        Spacer()

                        Button("Stop") {
                            Task { await appState.stopTracking() }
                        }
                        .buttonStyle(.borderedProminent)
                        .tint(.red)
                    }
                }
            } else if let gig = nextState.activeGig {
                // Active gig to continue
                VStack(alignment: .leading, spacing: 12) {
                    Text("Continue working on")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)

                    Text(gig.name)
                        .font(.system(size: 18, weight: .semibold))

                    if gig.totalHoursTracked > 0 {
                        Text("\(String(format: "%.1f", gig.totalHoursTracked))h tracked")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)
                    }

                    Button("Start Timer") {
                        Task {
                            await appState.startTracking(project: gig.name, clientId: gig.clientId.map { Int($0) })
                        }
                    }
                    .buttonStyle(.borderedProminent)
                }
            } else {
                // No active work
                VStack(alignment: .leading, spacing: 12) {
                    Text("Ready to work?")
                        .font(.system(size: 16, weight: .medium))

                    Text("Start tracking time on a project")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)

                    Button("Start Tracking") {
                        appState.selectedTab = .tracking
                    }
                    .buttonStyle(.borderedProminent)
                }
            }
        }
    }

    // MARK: - Bill Card

    private var billCard: some View {
        NextCard(
            title: "BILL",
            icon: "dollarsign.circle.fill",
            color: .orange,
            action: { appState.selectedTab = .invoices }
        ) {
            VStack(alignment: .leading, spacing: 12) {
                if nextState.pendingInvoiceAmount > 0 {
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Pending payment")
                                .font(.system(size: 12))
                                .foregroundColor(.secondary)
                            Text(appState.formatCurrency(nextState.pendingInvoiceAmount))
                                .font(.system(size: 20, weight: .bold))
                                .foregroundColor(.orange)
                        }

                        Spacer()

                        if nextState.overdueCount > 0 {
                            Text("\(nextState.overdueCount) overdue")
                                .font(.system(size: 12, weight: .medium))
                                .foregroundColor(.red)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 4)
                                .background(Color.red.opacity(0.1))
                                .clipShape(Capsule())
                        }
                    }
                }

                if nextState.unbilledHours > 0 {
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Unbilled work")
                                .font(.system(size: 12))
                                .foregroundColor(.secondary)
                            Text("\(String(format: "%.1f", nextState.unbilledHours))h ready to invoice")
                                .font(.system(size: 14, weight: .medium))
                        }

                        Spacer()

                        Button("Create Invoice") {
                            appState.selectedTab = .invoices
                        }
                        .buttonStyle(.bordered)
                        .controlSize(.small)
                    }
                }
            }
        }
    }

    // MARK: - Hunt Card

    private var huntCard: some View {
        NextCard(
            title: "HUNT",
            icon: "magnifyingglass",
            color: .purple,
            action: { appState.selectedTab = .hunter }
        ) {
            if let job = nextState.topJobMatch {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Text("Top match")
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)

                        Spacer()

                        if let score = job.matchScore, score > 0 {
                            HStack(spacing: 4) {
                                Image(systemName: "star.fill")
                                    .font(.system(size: 10))
                                Text("\(Int(score))%")
                                    .font(.system(size: 12, weight: .semibold))
                            }
                            .foregroundColor(.green)
                        }
                    }

                    Text(job.title)
                        .font(.system(size: 16, weight: .semibold))
                        .lineLimit(2)

                    if let company = job.company {
                        Text(company)
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)
                    }

                    Button("View Jobs") {
                        appState.selectedTab = .hunter
                    }
                    .buttonStyle(.bordered)
                    .controlSize(.small)
                }
            }
        }
    }

    // MARK: - Goal Card

    private var goalCard: some View {
        NextCard(
            title: "GOAL",
            icon: "target",
            color: .green,
            action: { appState.selectedTab = .reports }
        ) {
            if let goal = nextState.currentGoal {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Text(goal.periodLabel)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)

                        Spacer()

                        Text(goal.statusText)
                            .font(.system(size: 12, weight: .medium))
                            .foregroundColor(goal.isOnTrack ? .green : .orange)
                    }

                    // Progress bar
                    GeometryReader { geo in
                        ZStack(alignment: .leading) {
                            RoundedRectangle(cornerRadius: 4)
                                .fill(Color.gray.opacity(0.2))

                            RoundedRectangle(cornerRadius: 4)
                                .fill(goal.isOnTrack ? Color.green : Color.orange)
                                .frame(width: geo.size.width * min(goal.progress, 1.0))
                        }
                    }
                    .frame(height: 8)

                    HStack {
                        Text(appState.formatCurrency(goal.current))
                            .font(.system(size: 16, weight: .semibold))

                        Text("of \(appState.formatCurrency(goal.target))")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)

                        Spacer()

                        Text("\(Int(goal.progress * 100))%")
                            .font(.system(size: 14, weight: .medium))
                            .foregroundColor(goal.isOnTrack ? .green : .orange)
                    }
                }
            }
        }
    }
}

// MARK: - Next Card Component

struct NextCard<Content: View>: View {
    let title: String
    let icon: String
    let color: Color
    let action: () -> Void
    @ViewBuilder let content: () -> Content
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: 16) {
                // Header
                HStack(spacing: 8) {
                    Image(systemName: icon)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(color)

                    Text(title)
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(color)

                    Spacer()

                    Image(systemName: "chevron.right")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }

                content()
            }
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
                    .shadow(
                        color: Design.Shadow.sm.color,
                        radius: Design.Shadow.sm.radius,
                        y: Design.Shadow.sm.y
                    )
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Stat Badge

struct StatBadge: View {
    let value: String
    let label: String
    let icon: String
    let color: Color
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .trailing, spacing: 4) {
            HStack(spacing: 4) {
                Image(systemName: icon)
                    .font(.system(size: 12))
                    .foregroundColor(color)
                Text(value)
                    .font(.system(size: 16, weight: .semibold))
            }

            Text(label)
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }
}

// MARK: - Next State

@MainActor
class NextState: ObservableObject {
    @Published var activeGig: Gig?
    @Published var activeGigsCount: Int = 0
    @Published var pendingInvoiceAmount: Double = 0
    @Published var overdueCount: Int = 0
    @Published var unbilledHours: Double = 0
    @Published var topJobMatch: HunterJob?
    @Published var currentGoal: GoalProgress?

    struct GoalProgress {
        let target: Double
        let current: Double
        let periodLabel: String
        let daysRemaining: Int

        var progress: Double {
            guard target > 0 else { return 0 }
            return current / target
        }

        var isOnTrack: Bool {
            // Simple pace check
            let calendar = Calendar.current
            let dayOfMonth = calendar.component(.day, from: Date())
            let expectedProgress = Double(dayOfMonth) / 30.0
            return progress >= expectedProgress * 0.8 // Within 80% of expected pace
        }

        var statusText: String {
            if progress >= 1.0 { return "Goal reached!" }
            if isOnTrack { return "On track" }
            return "\(daysRemaining) days left"
        }
    }

    func loadData(appState: AppState) async {
        let db = DatabaseService.shared

        // Load active gig
        do {
            if let gig = try await db.getActiveGig() {
                activeGig = gig
            }
            activeGigsCount = try await db.getActiveGigsCount()
        } catch {
            print("[NextState] Failed to load gigs: \(error)")
        }

        // Load billing info
        pendingInvoiceAmount = appState.metrics.pendingAmount
        overdueCount = 0 // Would need to query

        // Load unbilled hours (sessions without invoice)
        do {
            unbilledHours = try await db.getUnbilledHours()
        } catch {
            print("[NextState] Failed to load unbilled hours: \(error)")
        }

        // Load top job match
        do {
            if let job = try await db.getTopJobMatch() {
                topJobMatch = job
            }
        } catch {
            print("[NextState] Failed to load top job: \(error)")
        }

        // Load current goal
        do {
            if let goal = try await db.getCurrentGoal() {
                let current = appState.metrics.monthlyRevenue
                let calendar = Calendar.current
                let daysRemaining = calendar.range(of: .day, in: .month, for: Date())?.count ?? 30
                    - calendar.component(.day, from: Date())

                currentGoal = GoalProgress(
                    target: goal.amount,
                    current: current,
                    periodLabel: goal.period.capitalized,
                    daysRemaining: daysRemaining
                )
            }
        } catch {
            print("[NextState] Failed to load goal: \(error)")
        }
    }
}

#Preview {
    NextContent()
        .environmentObject(AppState())
        .frame(width: 600, height: 700)
}
