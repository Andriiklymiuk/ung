//
//  SmartSearchView.swift
//  ung
//
//  Global smart search with natural language support
//

import SwiftUI

// MARK: - Smart Search Bar
struct SmartSearchBar: View {
    @Binding var searchText: String
    @Binding var isSearching: Bool
    @State private var suggestions: [String] = []
    @FocusState private var isFocused: Bool
    @Environment(\.colorScheme) var colorScheme

    let onSearch: (ParsedQuery) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Search field
            HStack(spacing: Design.Spacing.sm) {
                Image(systemName: "magnifyingglass")
                    .foregroundColor(Design.Colors.textTertiary)

                TextField("Search invoices, sessions, expenses...", text: $searchText)
                    .textFieldStyle(.plain)
                    .font(Design.Typography.bodyMedium)
                    .focused($isFocused)
                    .onSubmit {
                        performSearch()
                    }
                    .onChange(of: searchText) { _, newValue in
                        updateSuggestions(for: newValue)
                    }

                if !searchText.isEmpty {
                    Button(action: {
                        searchText = ""
                        suggestions = []
                    }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(Design.Colors.textTertiary)
                    }
                    .buttonStyle(.plain)
                }

                // Search hint
                Text("⌘K")
                    .font(.system(size: 10, weight: .medium, design: .monospaced))
                    .foregroundColor(Design.Colors.textTertiary)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 3)
                    .background(
                        RoundedRectangle(cornerRadius: 4)
                            .fill(Design.Colors.backgroundTertiary(colorScheme))
                    )
            }
            .padding(Design.Spacing.sm)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.md)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
                    .shadow(color: Design.Shadow.sm.color, radius: Design.Shadow.sm.radius, y: Design.Shadow.sm.y)
            )

            // Suggestions dropdown
            if isFocused && !suggestions.isEmpty {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(suggestions, id: \.self) { suggestion in
                        Button(action: {
                            searchText = suggestion
                            performSearch()
                        }) {
                            HStack {
                                Image(systemName: "magnifyingglass")
                                    .font(.system(size: 12))
                                    .foregroundColor(Design.Colors.textTertiary)
                                Text(suggestion)
                                    .font(Design.Typography.bodySmall)
                                    .foregroundColor(Design.Colors.textPrimary)
                                Spacer()
                            }
                            .padding(.horizontal, Design.Spacing.sm)
                            .padding(.vertical, Design.Spacing.xs)
                            .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                        .background(Color.clear)
                    }
                }
                .padding(.vertical, Design.Spacing.xs)
                .background(
                    RoundedRectangle(cornerRadius: Design.Radius.md)
                        .fill(Design.Colors.surfaceElevated(colorScheme))
                        .shadow(color: Design.Shadow.md.color, radius: Design.Shadow.md.radius, y: Design.Shadow.md.y)
                )
                .padding(.top, 4)
            }
        }
        .onAppear {
            updateSuggestions(for: "")
        }
    }

    private func updateSuggestions(for query: String) {
        suggestions = SmartSearchService.shared.getSuggestions(for: query)
    }

    private func performSearch() {
        let parsed = SmartSearchService.shared.parse(searchText)
        onSearch(parsed)
        isFocused = false
    }
}

// MARK: - Search Results View
struct SearchResultsView: View {
    let results: [SearchResult]
    let isLoading: Bool
    let query: ParsedQuery
    let onResultTap: (SearchResult) -> Void
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.md) {
            // Query summary
            if !query.isEmpty {
                queryBadges
            }

            if isLoading {
                ProgressView()
                    .frame(maxWidth: .infinity)
                    .padding()
            } else if results.isEmpty {
                emptyState
            } else {
                resultsList
            }
        }
    }

    private var queryBadges: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: Design.Spacing.xs) {
                if query.category != .all {
                    badge(query.category.rawValue.capitalized, icon: "folder", color: Design.Colors.primary)
                }

                if let dateRange = query.dateRange {
                    let formatter = DateFormatter()
                    formatter.dateStyle = .short
                    badge("\(formatter.string(from: dateRange.start)) - \(formatter.string(from: dateRange.end))",
                          icon: "calendar", color: Design.Colors.info)
                }

                if let status = query.statusFilter {
                    badge(status.rawValue.capitalized, icon: "tag", color: Design.Colors.warning)
                }

                if let amount = query.amountFilter {
                    let op: String
                    switch amount.op {
                    case .greaterThan: op = ">"
                    case .lessThan: op = "<"
                    case .equals: op = "="
                    case .between(let low, let high): op = "\(Int(low))-\(Int(high))"
                    }
                    badge("\(op)$\(Int(amount.value))", icon: "dollarsign.circle", color: Design.Colors.success)
                }

                if let client = query.clientName {
                    badge(client, icon: "person", color: Design.Colors.purple)
                }
            }
        }
    }

    private func badge(_ text: String, icon: String, color: Color) -> some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.system(size: 10))
            Text(text)
                .font(Design.Typography.labelSmall)
        }
        .foregroundColor(color)
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(
            Capsule()
                .fill(color.opacity(0.15))
        )
    }

    private var emptyState: some View {
        VStack(spacing: Design.Spacing.md) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 40))
                .foregroundColor(Design.Colors.textTertiary)
            Text("No results found")
                .font(Design.Typography.bodyMedium)
                .foregroundColor(Design.Colors.textSecondary)
            Text("Try a different search term or filter")
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textTertiary)
        }
        .frame(maxWidth: .infinity)
        .padding(Design.Spacing.xl)
    }

    private var resultsList: some View {
        LazyVStack(spacing: Design.Spacing.xs) {
            ForEach(results) { result in
                Button(action: { onResultTap(result) }) {
                    SearchResultRow(result: result)
                }
                .buttonStyle(.plain)
            }
        }
    }
}

// MARK: - Search Result Row
struct SearchResultRow: View {
    let result: SearchResult
    @Environment(\.colorScheme) var colorScheme
    @State private var isHovered = false

    private var categoryColor: Color {
        switch result.type {
        case .invoices: return Design.Colors.primary
        case .sessions: return Design.Colors.info
        case .clients: return Design.Colors.purple
        case .expenses: return Design.Colors.warning
        case .contracts: return Design.Colors.success
        case .all: return Design.Colors.textSecondary
        }
    }

    var body: some View {
        HStack(spacing: Design.Spacing.sm) {
            // Icon
            ZStack {
                Circle()
                    .fill(categoryColor.opacity(0.15))
                    .frame(width: 36, height: 36)
                Image(systemName: result.icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(categoryColor)
            }

            // Content
            VStack(alignment: .leading, spacing: 2) {
                Text(result.title)
                    .font(Design.Typography.bodyMedium)
                    .foregroundColor(Design.Colors.textPrimary)
                    .lineLimit(1)

                Text(result.subtitle)
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textSecondary)
                    .lineLimit(1)
            }

            Spacer()

            // Amount or detail
            if let amount = result.amount {
                Text(Formatters.formatCurrency(amount))
                    .font(Design.Typography.labelMedium)
                    .foregroundColor(Design.Colors.textPrimary)
            } else if let detail = result.detail {
                Text(detail)
                    .font(Design.Typography.labelSmall)
                    .foregroundColor(Design.Colors.textTertiary)
            }

            // Date
            if let date = result.date {
                Text(Formatters.shortDate.string(from: date))
                    .font(Design.Typography.labelSmall)
                    .foregroundColor(Design.Colors.textTertiary)
            }

            Image(systemName: "chevron.right")
                .font(.system(size: 12))
                .foregroundColor(Design.Colors.textTertiary)
                .opacity(isHovered ? 1 : 0.5)
        }
        .padding(Design.Spacing.sm)
        .background(
            RoundedRectangle(cornerRadius: Design.Radius.sm)
                .fill(isHovered ? Design.Colors.backgroundTertiary(colorScheme) : Color.clear)
        )
        .onHover { hovering in
            withAnimation(Design.Animation.quick) {
                isHovered = hovering
            }
        }
    }
}

// MARK: - Global Search Overlay (Cmd+K)
struct GlobalSearchOverlay: View {
    @Binding var isPresented: Bool
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme

    @State private var searchText = ""
    @State private var results: [SearchResult] = []
    @State private var isLoading = false
    @State private var parsedQuery = ParsedQuery()

    var body: some View {
        ZStack {
            // Background blur
            Color.black.opacity(0.3)
                .ignoresSafeArea()
                .onTapGesture {
                    isPresented = false
                }

            // Search panel
            VStack(spacing: 0) {
                // Search bar
                SmartSearchBar(
                    searchText: $searchText,
                    isSearching: $isLoading,
                    onSearch: performSearch
                )
                .padding(Design.Spacing.md)

                // Results
                if !searchText.isEmpty || !results.isEmpty {
                    Divider()

                    ScrollView {
                        SearchResultsView(
                            results: results,
                            isLoading: isLoading,
                            query: parsedQuery,
                            onResultTap: navigateToResult
                        )
                        .padding(Design.Spacing.md)
                    }
                    .frame(maxHeight: 400)
                } else {
                    // Quick actions
                    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
                        Text("Quick Search")
                            .font(Design.Typography.labelSmall)
                            .foregroundColor(Design.Colors.textTertiary)
                            .textCase(.uppercase)

                        quickSearchButton("unpaid invoices", icon: "doc.text", color: Design.Colors.warning)
                        quickSearchButton("hours this week", icon: "clock", color: Design.Colors.info)
                        quickSearchButton("expenses last month", icon: "dollarsign.circle", color: Design.Colors.error)
                    }
                    .padding(Design.Spacing.md)
                }
            }
            .frame(width: 500)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.lg)
                    .fill(Design.Colors.backgroundSecondary(colorScheme))
                    .shadow(color: Design.Shadow.lg.color, radius: Design.Shadow.lg.radius, y: Design.Shadow.lg.y)
            )
            .padding(.top, 100)
        }
        .onAppear {
            // Reset state
            searchText = ""
            results = []
        }
    }

    private func quickSearchButton(_ text: String, icon: String, color: Color) -> some View {
        Button(action: {
            searchText = text
            performSearch(SmartSearchService.shared.parse(text))
        }) {
            HStack {
                Image(systemName: icon)
                    .foregroundColor(color)
                Text(text)
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textPrimary)
                Spacer()
            }
            .padding(Design.Spacing.xs)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    private func performSearch(_ query: ParsedQuery) {
        parsedQuery = query
        isLoading = true

        Task {
            // Simulate search - in production, this would query the database
            results = await executeSearch(query)
            isLoading = false
        }
    }

    private func executeSearch(_ query: ParsedQuery) async -> [SearchResult] {
        var results: [SearchResult] = []

        // Search based on category
        switch query.category {
        case .invoices, .all:
            let invoices = appState.recentInvoices
            for invoice in invoices {
                if matchesQuery(invoice, query: query) {
                    results.append(SearchResult(
                        type: .invoices,
                        title: invoice.invoiceNum,
                        subtitle: invoice.client.isEmpty ? "Invoice" : invoice.client,
                        detail: invoice.status,
                        amount: Double(invoice.amount.replacingOccurrences(of: "$", with: "").replacingOccurrences(of: ",", with: "")) ?? 0,
                        deepLink: "ung://invoices/\(invoice.id)"
                    ))
                }
            }
            if query.category == .invoices { break }
            fallthrough

        case .sessions:
            let sessions = appState.recentSessions
            for session in sessions {
                if matchesQuery(session, query: query) {
                    results.append(SearchResult(
                        type: .sessions,
                        title: session.project,
                        subtitle: session.duration,
                        detail: session.date,
                        deepLink: "ung://sessions/\(session.id)"
                    ))
                }
            }
            if query.category == .sessions { break }
            fallthrough

        case .expenses:
            let expenses = appState.recentExpenses
            for expense in expenses {
                if matchesQuery(expense, query: query) {
                    results.append(SearchResult(
                        type: .expenses,
                        title: expense.description,
                        subtitle: expense.category,
                        amount: Double(expense.amount.replacingOccurrences(of: "$", with: "").replacingOccurrences(of: ",", with: "")) ?? 0,
                        deepLink: "ung://expenses/\(expense.id)"
                    ))
                }
            }
            if query.category == .expenses { break }
            fallthrough

        case .clients:
            let clients = appState.clients
            for client in clients {
                if matchesQuery(client, query: query) {
                    results.append(SearchResult(
                        type: .clients,
                        title: client.name,
                        subtitle: client.email,
                        deepLink: "ung://clients/\(client.id)"
                    ))
                }
            }
            if query.category == .clients { break }
            fallthrough

        case .contracts:
            let contracts = appState.contracts
            for contract in contracts {
                if matchesQuery(contract, query: query) {
                    results.append(SearchResult(
                        type: .contracts,
                        title: contract.name,
                        subtitle: contract.clientName,
                        detail: contract.type,
                        deepLink: "ung://contracts/\(contract.id)"
                    ))
                }
            }
        }

        // Apply limit
        if let limit = query.limit {
            return Array(results.prefix(limit))
        }

        return results
    }

    private func matchesQuery(_ item: Any, query: ParsedQuery) -> Bool {
        // For now, simple text matching
        // In production, would apply all filters
        if query.searchText.isEmpty { return true }

        let searchLower = query.searchText.lowercased()

        if let invoice = item as? RecentInvoice {
            return invoice.invoiceNum.lowercased().contains(searchLower) ||
                   invoice.client.lowercased().contains(searchLower)
        }

        if let session = item as? RecentSession {
            return session.project.lowercased().contains(searchLower)
        }

        if let expense = item as? RecentExpense {
            return expense.description.lowercased().contains(searchLower) ||
                   expense.category.lowercased().contains(searchLower)
        }

        if let client = item as? Client {
            return client.name.lowercased().contains(searchLower) ||
                   client.email.lowercased().contains(searchLower)
        }

        if let contract = item as? Contract {
            return contract.name.lowercased().contains(searchLower) ||
                   contract.clientName.lowercased().contains(searchLower)
        }

        return false
    }

    private func navigateToResult(_ result: SearchResult) {
        isPresented = false

        // Navigate based on result type
        switch result.type {
        case .invoices:
            appState.selectedTab = .invoices
        case .sessions:
            appState.selectedTab = .tracking
        case .clients:
            appState.selectedTab = .clients
        case .expenses:
            appState.selectedTab = .expenses
        case .contracts:
            appState.selectedTab = .contracts
        case .all:
            break
        }
    }
}

#Preview {
    SmartSearchBar(
        searchText: .constant(""),
        isSearching: .constant(false),
        onSearch: { _ in }
    )
    .padding()
    .frame(width: 400)
}
