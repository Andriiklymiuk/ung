//
//  MainWindowView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import LocalAuthentication
import SwiftUI

struct MainWindowView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        Group {
            // Show lock screen if app is locked
            if appState.isLocked {
                LockScreenView()
                    .frame(minWidth: 400, minHeight: 300)
                    .background(backgroundColor)
            } else {
                switch appState.status {
                case .loading:
                    LoadingView()
                        .frame(width: 300, height: 180)
                        .background(backgroundColor)
                case .notInitialized:
                    // Database initialization failed - show error view
                    DatabaseErrorView()
                        .frame(width: 340, height: 300)
                        .background(backgroundColor)
                case .ready:
                    mainContent
                    #if os(macOS)
                        .frame(minWidth: 600, minHeight: 400)
                    #endif
                }
            }
        }
        .task {
            // Only refresh if ready and not already loaded
            guard appState.status == .ready, appState.metrics.totalRevenue == 0 else { return }
            await appState.refreshDashboard()
        }
        .overlay(alignment: .bottom) {
            if appState.showToast {
                DSToast(
                    message: appState.toastMessage,
                    type: mapToastType(appState.toastType),
                    isPresented: $appState.showToast
                )
                .padding(.bottom, 16)
                .padding(.horizontal, 16)
                .transition(
                    .move(edge: .bottom).combined(with: .scale(scale: 0.9)).combined(with: .opacity)
                )
                .animation(.spring(response: 0.4, dampingFraction: 0.7), value: appState.showToast)
            }
        }
    }

    private func mapToastType(_ type: AppState.ToastType) -> DSToast.ToastType {
        switch type {
        case .success: return .success
        case .info: return .info
        case .warning: return .warning
        case .error: return .error
        }
    }

    @ViewBuilder
    private var mainContent: some View {
        #if os(iOS)
        // iOS: Use TabView for navigation
        iOSTabView()
        #else
        // macOS: Use sidebar navigation
        HStack(spacing: 0) {
            SidebarView()
            ContentAreaView()
        }
        .background(backgroundColor)
        #endif
    }

    private var backgroundColor: Color {
        #if os(macOS)
        colorScheme == .dark
            ? Color(nsColor: .windowBackgroundColor)
            : Color(nsColor: .windowBackgroundColor)
        #else
        Color(uiColor: .systemBackground)
        #endif
    }
}

// MARK: - iOS Tab View
#if os(iOS)
struct iOSTabView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        TabView(selection: $appState.selectedTab) {
            DashboardTabView()
                .tabItem {
                    Label(SidebarTab.dashboard.shortLabel, systemImage: SidebarTab.dashboard.icon)
                }
                .tag(SidebarTab.dashboard)

            TrackingTabView()
                .tabItem {
                    Label(SidebarTab.tracking.shortLabel, systemImage: SidebarTab.tracking.icon)
                }
                .tag(SidebarTab.tracking)

            ClientsTabView()
                .tabItem {
                    Label(SidebarTab.clients.shortLabel, systemImage: SidebarTab.clients.icon)
                }
                .tag(SidebarTab.clients)

            InvoicesTabView()
                .tabItem {
                    Label(SidebarTab.invoices.shortLabel, systemImage: SidebarTab.invoices.icon)
                }
                .tag(SidebarTab.invoices)

            MoreTabView()
                .tabItem {
                    Label("More", systemImage: "ellipsis")
                }
                .tag(SidebarTab.settings)
        }
    }
}

// MARK: - iOS Tab Views
struct DashboardTabView: View {
    var body: some View {
        NavigationStack {
            MainDashboardContent()
                .navigationTitle("Dashboard")
        }
    }
}

struct TrackingTabView: View {
    @State private var showAddSheet = false

    var body: some View {
        NavigationStack {
            TrackingContent(showAddSheet: $showAddSheet)
                .navigationTitle("Time Tracking")
        }
    }
}

struct ClientsTabView: View {
    @State private var showAddSheet = false

    var body: some View {
        NavigationStack {
            ClientsContent(showAddSheet: $showAddSheet)
                .navigationTitle("Clients")
        }
    }
}

struct InvoicesTabView: View {
    @State private var showAddSheet = false

    var body: some View {
        NavigationStack {
            InvoicesContent(showAddSheet: $showAddSheet)
                .navigationTitle("Invoices")
        }
    }
}

struct MoreTabView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        NavigationStack {
            List {
                Section("Features") {
                    NavigationLink {
                        ContractsContent(showAddSheet: .constant(false))
                            .navigationTitle("Contracts")
                    } label: {
                        Label("Contracts", systemImage: SidebarTab.contracts.icon)
                    }

                    NavigationLink {
                        ExpensesContent(showAddSheet: .constant(false))
                            .navigationTitle("Expenses")
                    } label: {
                        Label("Expenses", systemImage: SidebarTab.expenses.icon)
                    }

                    NavigationLink {
                        PomodoroContent()
                            .navigationTitle("Focus Timer")
                    } label: {
                        Label("Focus Timer", systemImage: SidebarTab.pomodoro.icon)
                    }

                    NavigationLink {
                        ReportsContent()
                            .navigationTitle("Reports")
                    } label: {
                        Label("Reports", systemImage: SidebarTab.reports.icon)
                    }
                }

                Section("Settings") {
                    NavigationLink {
                        SettingsContent()
                            .navigationTitle("Settings")
                    } label: {
                        Label("Settings", systemImage: SidebarTab.settings.icon)
                    }
                }
            }
            .navigationTitle("More")
        }
    }
}
#endif

// MARK: - macOS Sidebar View
#if os(macOS)
struct SidebarView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme
    @State private var showMoreMenu = false
    @State private var biometricType: LABiometryType = .none

    private let itemHeight: CGFloat = 60
    private let fixedHeaderHeight: CGFloat = 96

    var body: some View {
        GeometryReader { geometry in
            let availableHeight =
                geometry.size.height - fixedHeaderHeight
                - (appState.isTracking || appState.pomodoroState.isActive ? 100 : 0)
            let maxVisibleItems = max(1, Int(availableHeight / itemHeight))
            let allTabs = SidebarTab.allCases
            let needsOverflow = allTabs.count > maxVisibleItems
            let visibleTabs = needsOverflow ? Array(allTabs.prefix(maxVisibleItems - 1)) : allTabs
            let hiddenTabs = needsOverflow ? Array(allTabs.suffix(from: maxVisibleItems - 1)) : []

            VStack(spacing: 0) {
                appHeader

                if appState.isTracking || appState.pomodoroState.isActive {
                    activeStatusBanner
                }

                VStack(spacing: 4) {
                    ForEach(visibleTabs) { tab in
                        SidebarItem(tab: tab)
                    }

                    if needsOverflow {
                        MoreTabsButton(hiddenTabs: hiddenTabs, showMenu: $showMoreMenu)
                    }
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 12)

                Spacer()

                touchIDButton
            }
        }
        .frame(width: 80)
        .background(.ultraThinMaterial)
        .onAppear {
            checkBiometrics()
        }
    }

    private func checkBiometrics() {
        let context = LAContext()
        var error: NSError?
        if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
            biometricType = context.biometryType
        }
    }

    private var appHeader: some View {
        VStack(spacing: 4) {
            ZStack(alignment: .bottomTrailing) {
                if let appIcon = NSImage(named: "AppIcon") {
                    Image(nsImage: appIcon)
                        .resizable()
                        .frame(width: 40, height: 40)
                        .cornerRadius(10)
                } else {
                    ZStack {
                        RoundedRectangle(cornerRadius: 10)
                            .fill(Color.blue)
                            .frame(width: 40, height: 40)

                        Image(systemName: "clock.badge.checkmark")
                            .font(.system(size: 18, weight: .semibold))
                            .foregroundColor(.white)
                    }
                }

                if appState.secureMode {
                    ZStack {
                        Circle()
                            .fill(Color.green)
                            .frame(width: 16, height: 16)

                        Image(systemName: "eye.slash.fill")
                            .font(.system(size: 8, weight: .bold))
                            .foregroundColor(.white)
                    }
                    .offset(x: 4, y: 4)
                }
            }
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 16)
        .background(.thinMaterial)
        .onTapGesture(count: 2) {
            appState.secureMode.toggle()
            let message =
                appState.secureMode
                ? "Secure Mode Enabled - Sensitive data hidden"
                : "Secure Mode Disabled - All data visible"
            appState.showToastMessage(message, type: appState.secureMode ? .warning : .info)
        }
        .help(
            appState.secureMode
                ? "Secure Mode: On (double-click to toggle)" : "Double-click to enable Secure Mode")
    }

    private var activeStatusBanner: some View {
        VStack(spacing: 4) {
            ZStack {
                Circle()
                    .fill(
                        appState.isTracking
                            ? Color.red.opacity(0.15)
                            : (appState.pomodoroState.isBreak
                                ? Color.green.opacity(0.15) : Color.orange.opacity(0.15))
                    )
                    .frame(width: 44, height: 44)

                Image(systemName: appState.isTracking ? "clock.fill" : "brain.head.profile")
                    .font(.system(size: 18))
                    .foregroundColor(
                        appState.isTracking
                            ? Color.red : (appState.pomodoroState.isBreak ? Color.green : Color.orange)
                    )
            }

            if appState.isTracking, let session = appState.activeSession {
                Text(session.formattedDuration)
                    .font(.system(size: 10, weight: .medium, design: .monospaced))
                    .foregroundColor(.primary)
            } else if appState.pomodoroState.isActive {
                Text(appState.pomodoroState.formattedTime)
                    .font(.system(size: 10, weight: .medium, design: .monospaced))
                    .foregroundColor(.primary)
            }

            Button(action: {
                if appState.isTracking {
                    Task { await appState.stopTracking() }
                } else {
                    appState.stopPomodoro()
                }
            }) {
                Image(systemName: "stop.fill")
                    .font(.system(size: 10))
                    .foregroundColor(.white)
                    .padding(6)
                    .background(Circle().fill(Color.red.opacity(0.8)))
            }
            .buttonStyle(.plain)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 8)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(
                    appState.isTracking
                        ? Color.red.opacity(0.1)
                        : (appState.pomodoroState.isBreak
                            ? Color.green.opacity(0.1) : Color.orange.opacity(0.1)))
        )
        .padding(.horizontal, 8)
        .padding(.bottom, 8)
    }

    @ViewBuilder
    private var touchIDButton: some View {
        if appState.appLockEnabled {
            Button(action: {
                appState.lockApp()
                appState.showToastMessage("App Locked", type: .warning)
            }) {
                Image(
                    systemName: biometricType == .touchID
                        ? "touchid" : (biometricType == .faceID ? "faceid" : "lock.fill")
                )
                .font(.system(size: 20))
                .foregroundColor(.secondary)
                .frame(width: 44, height: 44)
            }
            .buttonStyle(.plain)
            .help(
                biometricType == .touchID
                    ? "Lock app (Touch ID required to unlock)"
                    : (biometricType == .faceID
                        ? "Lock app (Face ID required to unlock)"
                        : "Lock app (Password required to unlock)")
            )
            .padding(.horizontal, 8)
            .padding(.bottom, 12)
        }
    }
}

// MARK: - Sidebar Item
struct SidebarItem: View {
    @EnvironmentObject var appState: AppState
    let tab: SidebarTab
    @Environment(\.colorScheme) var colorScheme

    var isSelected: Bool {
        appState.selectedTab == tab
    }

    var body: some View {
        Button(action: {
            appState.selectedTab = tab
        }) {
            VStack(spacing: 2) {
                ZStack {
                    Image(systemName: isSelected ? tab.iconFilled : tab.icon)
                        .font(.system(size: 18, weight: isSelected ? .semibold : .regular))
                        .foregroundColor(isSelected ? SidebarTab.accentColor : .secondary)
                        .frame(width: 36, height: 36)

                    if tab == .invoices && appState.metrics.overdueAmount > 0 {
                        VStack {
                            HStack {
                                Spacer()
                                Circle()
                                    .fill(Color.red)
                                    .frame(width: 8, height: 8)
                            }
                            Spacer()
                        }
                        .frame(width: 36, height: 36)
                    }
                }

                Text(tab.shortLabel)
                    .font(.system(size: 9, weight: isSelected ? .medium : .regular))
                    .foregroundColor(isSelected ? SidebarTab.accentColor : .secondary)
                    .lineLimit(1)
            }
            .contentShape(Rectangle())
            .frame(width: 64, height: 56)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(
                        isSelected
                            ? SidebarTab.accentColor.opacity(colorScheme == .dark ? 0.15 : 0.1)
                            : Color.clear
                    )
            )
        }
        .buttonStyle(.plain)
        .help(tab.rawValue)
    }
}

// MARK: - More Tabs Button
struct MoreTabsButton: View {
    @EnvironmentObject var appState: AppState
    let hiddenTabs: [SidebarTab]
    @Binding var showMenu: Bool
    @Environment(\.colorScheme) var colorScheme

    var hasSelectedHidden: Bool {
        hiddenTabs.contains(appState.selectedTab)
    }

    var body: some View {
        Menu {
            ForEach(hiddenTabs) { tab in
                Button(action: { appState.selectedTab = tab }) {
                    Label(tab.rawValue, systemImage: tab.icon)
                }
            }
        } label: {
            VStack(spacing: 2) {
                ZStack {
                    Image(systemName: "ellipsis")
                        .font(.system(size: 18, weight: hasSelectedHidden ? .semibold : .regular))
                        .foregroundColor(hasSelectedHidden ? SidebarTab.accentColor : .secondary)
                        .frame(width: 36, height: 36)
                }

                Text("More")
                    .font(.system(size: 9, weight: hasSelectedHidden ? .medium : .regular))
                    .foregroundColor(hasSelectedHidden ? SidebarTab.accentColor : .secondary)
                    .lineLimit(1)
            }
            .contentShape(Rectangle())
            .frame(width: 64, height: 56)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(
                        hasSelectedHidden
                            ? SidebarTab.accentColor.opacity(colorScheme == .dark ? 0.15 : 0.1)
                            : Color.clear
                    )
            )
        }
        .menuStyle(.button)
        .buttonStyle(.plain)
        .menuIndicator(.hidden)
        .help("More options")
    }
}
#endif

// MARK: - Content Area
struct ContentAreaView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.colorScheme) var colorScheme
    @State private var showGlobalSearch = false
    @State private var globalSearchQuery = ""

    @State private var showAddClient = false
    @State private var showAddContract = false
    @State private var showCreateInvoice = false
    @State private var showAddExpense = false
    @State private var showStartTracking = false

    var body: some View {
        ZStack {
            VStack(spacing: 0) {
                contentHeader
                Divider()
                contentView
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(colorScheme == .dark ? Color(white: 0.16) : Color.white)

            if showGlobalSearch {
                GlobalSearchOverlay(
                    searchQuery: $globalSearchQuery,
                    isPresented: $showGlobalSearch
                )
                .transition(.identity)
            }
        }
    }

    @ViewBuilder
    private var contentView: some View {
        switch appState.selectedTab {
        case .dashboard:
            MainDashboardContent()
        case .tracking:
            TrackingContent(showAddSheet: $showStartTracking)
        case .clients:
            ClientsContent(showAddSheet: $showAddClient)
        case .contracts:
            ContractsContent(showAddSheet: $showAddContract)
        case .invoices:
            InvoicesContent(showAddSheet: $showCreateInvoice)
        case .expenses:
            ExpensesContent(showAddSheet: $showAddExpense)
        case .pomodoro:
            PomodoroContent()
        case .reports:
            ReportsContent()
        case .settings:
            SettingsContent()
        }
    }

    private var contentHeader: some View {
        HStack(spacing: 12) {
            VStack(alignment: .leading, spacing: 2) {
                Text(appState.selectedTab.rawValue)
                    .font(.system(size: 20, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                Text(headerSubtitle)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Button(action: { Task { await appState.refreshDashboard() } }) {
                Image(systemName: appState.isRefreshing ? "arrow.clockwise" : "arrow.clockwise")
                    .font(.system(size: 14))
                    .foregroundColor(.secondary)
                    .rotationEffect(.degrees(appState.isRefreshing ? 360 : 0))
                    .animation(
                        appState.isRefreshing
                            ? .linear(duration: 1).repeatForever(autoreverses: false) : .default,
                        value: appState.isRefreshing)
            }
            .buttonStyle(.plain)
            .disabled(appState.isRefreshing)

            if let action = primaryAction {
                Button(action: action.action) {
                    HStack(spacing: 6) {
                        Image(systemName: action.icon)
                            .font(.system(size: 12, weight: .semibold))
                        Text(action.title)
                            .font(.system(size: 13, weight: .semibold))
                    }
                    .foregroundColor(.white)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 10)
                    .background(
                        RoundedRectangle(cornerRadius: 10)
                            .fill(
                                LinearGradient(
                                    colors: [action.color, action.color.opacity(0.8)],
                                    startPoint: .top,
                                    endPoint: .bottom
                                )
                            )
                            .shadow(color: action.color.opacity(0.3), radius: 4, y: 2)
                    )
                }
                .buttonStyle(.plain)
            }
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 16)
    }

    private var headerSubtitle: String {
        switch appState.selectedTab {
        case .dashboard: return "Overview of your business"
        case .tracking: return "Manage your time sessions"
        case .clients: return "\(appState.clientCount) clients"
        case .contracts:
            if appState.clients.isEmpty {
                return "Add a client first to create contracts"
            }
            return "\(appState.contractCount) contracts"
        case .invoices:
            if !appState.setupStatus.hasCompany {
                return "Set up your company to create invoices"
            }
            if appState.clients.isEmpty {
                return "Add a client first to create invoices"
            }
            if appState.contracts.isEmpty {
                return "Create a contract first to invoice clients"
            }
            return "\(appState.invoiceCount) invoices"
        case .expenses: return "Track your business expenses"
        case .pomodoro: return "Focus sessions: \(appState.pomodoroState.sessionsCompleted)"
        case .reports: return "Analytics and insights"
        case .settings: return "Configure your preferences"
        }
    }

    private var primaryAction: (title: String, icon: String, color: Color, action: () -> Void)? {
        switch appState.selectedTab {
        case .tracking:
            return appState.isTracking
                ? nil : ("Start Tracking", "play.fill", .green, { showStartTracking = true })
        case .clients:
            return ("New Client", "plus", .purple, { showAddClient = true })
        case .contracts:
            if appState.clients.isEmpty {
                return (
                    "Add Client First", "person.badge.plus", .purple, { appState.selectedTab = .clients }
                )
            }
            return ("New Contract", "plus", .indigo, { showAddContract = true })
        case .invoices:
            if !appState.setupStatus.hasCompany {
                return ("Set Up Company", "building.2", .blue, { appState.selectedTab = .settings })
            }
            if appState.clients.isEmpty {
                return (
                    "Add Client First", "person.badge.plus", .purple, { appState.selectedTab = .clients }
                )
            }
            if appState.contracts.isEmpty {
                return (
                    "Add Contract First", "doc.badge.plus", .indigo, { appState.selectedTab = .contracts }
                )
            }
            return ("New Invoice", "plus", .teal, { showCreateInvoice = true })
        case .expenses:
            return ("New Expense", "plus", .orange, { showAddExpense = true })
        case .pomodoro:
            return appState.pomodoroState.isActive
                ? nil : ("Start Focus", "play.fill", .red, { appState.startPomodoro() })
        default:
            return nil
        }
    }
}

// MARK: - Global Search Overlay
#if os(macOS)
struct GlobalSearchOverlay: View {
    @EnvironmentObject var appState: AppState
    @Binding var searchQuery: String
    @Binding var isPresented: Bool
    @FocusState private var isFocused: Bool
    @Environment(\.colorScheme) var colorScheme

    struct SearchResult: Identifiable {
        let id = UUID()
        let type: SearchResultType
        let title: String
        let subtitle: String
        let icon: String
        let color: Color
        let action: () -> Void
    }

    enum SearchResultType: String {
        case client, contract, invoice, expense, session, action
    }

    var searchResults: [SearchResult] {
        guard !searchQuery.isEmpty else {
            return quickActions
        }

        var results: [SearchResult] = []

        results += appState.clients
            .filter {
                $0.name.localizedCaseInsensitiveContains(searchQuery)
                    || $0.email.localizedCaseInsensitiveContains(searchQuery)
            }
            .prefix(3)
            .map { client in
                SearchResult(
                    type: .client,
                    title: client.name,
                    subtitle: client.email.isEmpty ? "Client" : client.email,
                    icon: "person.fill",
                    color: Design.Colors.purple,
                    action: {
                        appState.selectedTab = .clients
                        isPresented = false
                    }
                )
            }

        results += appState.contracts
            .filter {
                $0.name.localizedCaseInsensitiveContains(searchQuery)
                    || $0.clientName.localizedCaseInsensitiveContains(searchQuery)
            }
            .prefix(3)
            .map { contract in
                SearchResult(
                    type: .contract,
                    title: contract.name,
                    subtitle: "\(contract.clientName) - $\(Int(contract.rate))/hr",
                    icon: "doc.text.fill",
                    color: Design.Colors.indigo,
                    action: {
                        appState.selectedTab = .contracts
                        isPresented = false
                    }
                )
            }

        return results
    }

    var quickActions: [SearchResult] {
        [
            SearchResult(
                type: .action, title: "Start Tracking", subtitle: "Begin a new time session",
                icon: "play.fill", color: Design.Colors.success,
                action: {
                    appState.selectedTab = .tracking
                    isPresented = false
                }),
            SearchResult(
                type: .action, title: "Create Invoice", subtitle: "Bill your clients",
                icon: "doc.text.fill", color: Design.Colors.teal,
                action: {
                    appState.selectedTab = .invoices
                    isPresented = false
                }),
            SearchResult(
                type: .action, title: "Add Client", subtitle: "Register a new client",
                icon: "person.badge.plus", color: Design.Colors.purple,
                action: {
                    appState.selectedTab = .clients
                    isPresented = false
                }),
        ]
    }

    var body: some View {
        ZStack {
            Color.black.opacity(0.5)
                .ignoresSafeArea()
                .onTapGesture { isPresented = false }
                .drawingGroup()

            VStack(spacing: 0) {
                HStack(spacing: Design.Spacing.sm) {
                    Image(systemName: "magnifyingglass")
                        .font(.system(size: 16))
                        .foregroundColor(Design.Colors.textSecondary)

                    SearchTextField(
                        text: $searchQuery,
                        placeholder: "Search clients, invoices, or type a command...",
                        isFocused: $isFocused
                    )
                    .frame(height: 24)

                    if !searchQuery.isEmpty {
                        Button(action: { searchQuery = "" }) {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundColor(Design.Colors.textTertiary)
                        }
                        .buttonStyle(.plain)
                    }

                    Text("ESC")
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                        .foregroundColor(Design.Colors.textTertiary)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 3)
                        .background(RoundedRectangle(cornerRadius: 4).fill(Color.secondary.opacity(0.15)))
                }
                .padding(Design.Spacing.md)

                Divider()

                ScrollView {
                    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
                        if searchQuery.isEmpty {
                            Text("Quick Actions")
                                .font(Design.Typography.labelSmall)
                                .foregroundColor(Design.Colors.textTertiary)
                                .padding(.horizontal, Design.Spacing.md)
                                .padding(.top, Design.Spacing.sm)
                        } else if searchResults.isEmpty {
                            VStack(spacing: Design.Spacing.sm) {
                                Image(systemName: "magnifyingglass")
                                    .font(.system(size: 24))
                                    .foregroundColor(Design.Colors.textTertiary)
                                Text("No results found")
                                    .font(Design.Typography.bodyMedium)
                                    .foregroundColor(Design.Colors.textSecondary)
                            }
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, Design.Spacing.xl)
                        }

                        ForEach(searchResults) { result in
                            SearchResultRow(result: result)
                        }
                    }
                    .padding(.vertical, Design.Spacing.sm)
                }
                .frame(maxHeight: 400)
            }
            .frame(width: 560)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.lg)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
                    .shadow(color: .black.opacity(0.2), radius: 16, y: 8)
            )
            .drawingGroup()
            .onAppear { isFocused = true }
            .onExitCommand { isPresented = false }
        }
    }
}

struct SearchResultRow: View {
    let result: GlobalSearchOverlay.SearchResult
    @State private var isHovered = false

    var body: some View {
        Button(action: result.action) {
            HStack(spacing: Design.Spacing.sm) {
                ZStack {
                    RoundedRectangle(cornerRadius: Design.Radius.xs)
                        .fill(result.color.opacity(0.15))
                        .frame(width: 32, height: 32)

                    Image(systemName: result.icon)
                        .font(.system(size: 14))
                        .foregroundColor(result.color)
                }

                VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
                    Text(result.title)
                        .font(Design.Typography.labelMedium)
                        .foregroundColor(Design.Colors.textPrimary)

                    Text(result.subtitle)
                        .font(Design.Typography.bodySmall)
                        .foregroundColor(Design.Colors.textSecondary)
                }

                Spacer()

                if result.type != .action {
                    Text(result.type.rawValue.capitalized)
                        .font(Design.Typography.labelSmall)
                        .foregroundColor(Design.Colors.textTertiary)
                        .padding(.horizontal, Design.Spacing.xs)
                        .padding(.vertical, 2)
                        .background(RoundedRectangle(cornerRadius: 4).fill(Color.secondary.opacity(0.1)))
                }

                Image(systemName: "return")
                    .font(.system(size: 10))
                    .foregroundColor(Design.Colors.textTertiary)
                    .opacity(isHovered ? 1 : 0)
            }
            .padding(.horizontal, Design.Spacing.md)
            .padding(.vertical, Design.Spacing.sm)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.sm)
                    .fill(isHovered ? Color.secondary.opacity(0.1) : Color.clear)
            )
        }
        .buttonStyle(.plain)
        .onHover { hovering in isHovered = hovering }
    }
}

struct SearchTextField: NSViewRepresentable {
    @Binding var text: String
    var placeholder: String
    var isFocused: FocusState<Bool>.Binding

    func makeNSView(context: Context) -> NSTextField {
        let textField = NSTextField()
        textField.placeholderString = placeholder
        textField.isBordered = false
        textField.drawsBackground = false
        textField.focusRingType = .none
        textField.font = NSFont.systemFont(ofSize: 14)
        textField.textColor = NSColor.labelColor
        textField.delegate = context.coordinator
        textField.cell?.sendsActionOnEndEditing = false
        return textField
    }

    func updateNSView(_ nsView: NSTextField, context: Context) {
        if nsView.stringValue != text {
            nsView.stringValue = text
        }
        if isFocused.wrappedValue && nsView.window?.firstResponder != nsView {
            DispatchQueue.main.async {
                nsView.window?.makeFirstResponder(nsView)
            }
        }
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, NSTextFieldDelegate {
        var parent: SearchTextField

        init(_ parent: SearchTextField) {
            self.parent = parent
        }

        func controlTextDidChange(_ notification: Notification) {
            if let textField = notification.object as? NSTextField {
                parent.text = textField.stringValue
            }
        }
    }
}
#endif

#Preview {
    MainWindowView()
        .environmentObject(AppState())
        .frame(width: 1100, height: 700)
}
