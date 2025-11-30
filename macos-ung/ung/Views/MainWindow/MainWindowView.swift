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
                    // Show walkthrough for first-time users
                    if !appState.hasSeenWelcome {
                        WelcomeWalkthroughView()
                    } else {
                        mainContent
                    }
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
        .overlay(alignment: .top) {
            // iCloud sync banner at TOP of screen
            if appState.showSyncBanner {
                iCloudSyncBanner
                    .transition(.move(edge: .top).combined(with: .opacity))
                    .animation(.spring(response: 0.4, dampingFraction: 0.7), value: appState.showSyncBanner)
            }
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
        #if os(macOS)
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in
            // Trigger iCloud sync when app becomes active (from background)
            appState.checkICloudSync()
        }
        #else
        .onReceive(NotificationCenter.default.publisher(for: UIApplication.didBecomeActiveNotification)) { _ in
            // Trigger iCloud sync when app becomes active (from background)
            appState.checkICloudSync()
        }
        #endif
    }

    // MARK: - iCloud Sync Banner
    private var iCloudSyncBanner: some View {
        HStack(spacing: Design.Spacing.xs) {
            if case .syncing = appState.syncStatus {
                ProgressView()
                    .scaleEffect(0.7)
                    #if os(macOS)
                    .controlSize(.small)
                    #endif
            } else {
                Image(systemName: "checkmark.icloud.fill")
                    .foregroundColor(Design.Colors.success)
            }

            Text(syncStatusText)
                .font(Design.Typography.labelMedium)
                .foregroundColor(.white)
        }
        .padding(.horizontal, Design.Spacing.md)
        .padding(.vertical, Design.Spacing.xs)
        .background(
            Capsule()
                .fill(Design.Colors.brand.opacity(0.9))
                .shadow(color: Design.Shadow.md.color, radius: Design.Shadow.md.radius, y: Design.Shadow.md.y)
        )
        .padding(.top, Design.Spacing.xs)
        .accessibilityElement(children: .combine)
        .accessibilityLabel("iCloud sync: \(syncStatusText)")
    }

    private var syncStatusText: String {
        switch appState.syncStatus {
        case .syncing:
            return "Syncing with iCloud..."
        case .completed:
            return "iCloud sync complete"
        case .error(let message):
            return "Sync error: \(message)"
        case .idle:
            return ""
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
        Design.Colors.windowBackground
    }
}

// MARK: - iOS Tab View
#if os(iOS)
struct iOSTabView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.horizontalSizeClass) var horizontalSizeClass

    var body: some View {
        if horizontalSizeClass == .regular {
            // iPad: Use sidebar navigation like macOS
            iPadSplitView()
        } else {
            // iPhone: Use tab bar
            iPhoneTabView()
        }
    }
}

// MARK: - iPhone Tab View
struct iPhoneTabView: View {
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

// MARK: - iPad Split View
struct iPadSplitView: View {
    @EnvironmentObject var appState: AppState
    @State private var columnVisibility: NavigationSplitViewVisibility = .all

    var body: some View {
        NavigationSplitView(columnVisibility: $columnVisibility) {
            // Sidebar
            iPadSidebar()
        } detail: {
            // Content
            iPadDetailView()
        }
        .navigationSplitViewStyle(.balanced)
    }
}

struct iPadSidebar: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        List(selection: $appState.selectedTab) {
            Section("Main") {
                ForEach([SidebarTab.dashboard, .tracking, .clients, .invoices], id: \.self) { tab in
                    Label(tab.rawValue, systemImage: tab.icon)
                        .tag(tab)
                }
            }

            Section("More") {
                ForEach([SidebarTab.contracts, .expenses, .pomodoro, .reports], id: \.self) { tab in
                    Label(tab.rawValue, systemImage: tab.icon)
                        .tag(tab)
                }
            }

            Section {
                Label(SidebarTab.settings.rawValue, systemImage: SidebarTab.settings.icon)
                    .tag(SidebarTab.settings)
            }
        }
        .listStyle(.sidebar)
        .navigationTitle("UNG")
        .toolbar {
            // iCloud sync indicator
            ToolbarItem(placement: .navigation) {
                if appState.iCloudEnabled {
                    iCloudStatusIcon
                }
            }

            ToolbarItem(placement: .primaryAction) {
                if appState.isTracking {
                    HStack(spacing: Design.Spacing.xxs) {
                        Circle()
                            .fill(Design.Colors.error)
                            .frame(width: 8, height: 8)
                        Text(appState.activeSession?.formattedDuration ?? "")
                            .font(Design.Typography.monoSmall)
                    }
                    .accessibilityLabel("Tracking: \(appState.activeSession?.formattedDuration ?? "")")
                }
            }
        }
    }

    @ViewBuilder
    private var iCloudStatusIcon: some View {
        Group {
            if case .syncing = appState.syncStatus {
                ProgressView()
                    .scaleEffect(0.6)
            } else if case .error = appState.syncStatus {
                Image(systemName: "exclamationmark.icloud.fill")
                    .foregroundColor(Design.Colors.warning)
            } else {
                Image(systemName: "checkmark.icloud.fill")
                    .foregroundColor(Design.Colors.success)
            }
        }
        .font(.system(size: Design.IconSize.sm))
        .accessibilityLabel("iCloud sync status")
    }
}

struct iPadDetailView: View {
    @EnvironmentObject var appState: AppState
    @State private var showAddSheet = false

    var body: some View {
        Group {
            switch appState.selectedTab {
            case .next:
                NextContent()
            case .gigs:
                GigsContent()
            case .dig:
                DigContent()
            case .dashboard:
                MainDashboardContent()
            case .tracking:
                TrackingContent(showAddSheet: $showAddSheet)
            case .clients:
                ClientsContent(showAddSheet: $showAddSheet)
            case .contracts:
                ContractsContent(showAddSheet: $showAddSheet)
            case .invoices:
                InvoicesContent(showAddSheet: $showAddSheet)
            case .expenses:
                ExpensesContent(showAddSheet: $showAddSheet)
            case .pomodoro:
                PomodoroContent()
            case .hunter:
                HunterContent()
            case .reports:
                ReportsContent()
            case .settings:
                SettingsContent()
            }
        }
        .navigationTitle(appState.selectedTab.rawValue)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                if let action = primaryAction {
                    Button(action: action.action) {
                        Label(action.title, systemImage: action.icon)
                    }
                    .tint(action.color)
                }
            }

            ToolbarItem(placement: .automatic) {
                Button(action: { Task { await appState.refreshDashboard() } }) {
                    Image(systemName: "arrow.clockwise")
                }
                .disabled(appState.isRefreshing)
            }
        }
    }

    private var primaryAction: (title: String, icon: String, color: Color, action: () -> Void)? {
        switch appState.selectedTab {
        case .tracking:
            return appState.isTracking
                ? nil : ("Start", "play.fill", .green, { showAddSheet = true })
        case .clients:
            return ("Add", "plus", .purple, { showAddSheet = true })
        case .contracts:
            return appState.clients.isEmpty ? nil : ("Add", "plus", .indigo, { showAddSheet = true })
        case .invoices:
            return (appState.setupStatus.hasCompany && !appState.clients.isEmpty && !appState.contracts.isEmpty)
                ? ("Create", "plus", .teal, { showAddSheet = true }) : nil
        case .expenses:
            return ("Add", "plus", .orange, { showAddSheet = true })
        case .pomodoro:
            return appState.pomodoroState.isActive
                ? nil : ("Start", "play.fill", .red, { appState.startPomodoro() })
        default:
            return nil
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

                Section("Tools") {
                    NavigationLink {
                        IncomeGoalsView()
                            .navigationTitle("Income Goals")
                    } label: {
                        Label("Income Goals", systemImage: "target")
                    }

                    NavigationLink {
                        RateCalculatorView()
                            .navigationTitle("Rate Calculator")
                    } label: {
                        Label("Rate Calculator", systemImage: "dollarsign.circle")
                    }

                    NavigationLink {
                        CSVExportView()
                            .navigationTitle("Export Data")
                    } label: {
                        Label("Export Data", systemImage: "square.and.arrow.up")
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
                // App icon with sync indicator
                ZStack(alignment: .topLeading) {
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

                    // iCloud sync indicator (top-left corner)
                    if appState.iCloudEnabled {
                        syncStatusBadge
                            .offset(x: -6, y: -6)
                    }
                }

                // Secure mode badge (bottom-right corner)
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
        .help(syncStatusHelp)
    }

    @ViewBuilder
    private var syncStatusBadge: some View {
        ZStack {
            Circle()
                .fill(syncBadgeColor)
                .frame(width: 18, height: 18)
                .shadow(color: syncBadgeColor.opacity(0.4), radius: 3, x: 0, y: 1)

            if case .syncing = appState.syncStatus {
                // Spinning cloud icon when syncing
                Image(systemName: "arrow.triangle.2.circlepath")
                    .font(.system(size: 9, weight: .bold))
                    .foregroundColor(.white)
                    .rotationEffect(.degrees(syncRotation))
                    .onAppear {
                        withAnimation(.linear(duration: 1).repeatForever(autoreverses: false)) {
                            syncRotation = 360
                        }
                    }
            } else if case .error = appState.syncStatus {
                Image(systemName: "exclamationmark.icloud.fill")
                    .font(.system(size: 9, weight: .bold))
                    .foregroundColor(.white)
            } else {
                Image(systemName: "checkmark.icloud.fill")
                    .font(.system(size: 9, weight: .bold))
                    .foregroundColor(.white)
            }
        }
    }

    @State private var syncRotation: Double = 0

    private var syncBadgeColor: Color {
        switch appState.syncStatus {
        case .syncing:
            return Color.blue
        case .error:
            return Color.orange
        default:
            return Color.green
        }
    }

    private var syncStatusHelp: String {
        if appState.iCloudEnabled {
            switch appState.syncStatus {
            case .syncing:
                return "Syncing with iCloud..."
            case .error(let msg):
                return "Sync error: \(msg)"
            default:
                return "iCloud sync enabled. Double-click to toggle Secure Mode"
            }
        }
        return appState.secureMode
            ? "Secure Mode: On (double-click to toggle)"
            : "Double-click to enable Secure Mode"
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

            if appState.showGlobalSearch {
                GlobalSearchOverlay(
                    isPresented: $appState.showGlobalSearch
                )
                .transition(.opacity)
            }
        }
    }

    @ViewBuilder
    private var contentView: some View {
        switch appState.selectedTab {
        case .next:
            NextContent()
        case .gigs:
            GigsContent()
        case .dig:
            DigContent()
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
        case .hunter:
            HunterContent()
        case .reports:
            ReportsContent()
        case .settings:
            SettingsContent()
        }
    }

    private var contentHeader: some View {
        HStack(spacing: Design.Spacing.sm) {
            VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
                Text(appState.selectedTab.rawValue)
                    .font(Design.Typography.headingMedium)
                    .foregroundColor(Design.Colors.textPrimary)
                    .accessibleHeader(label: appState.selectedTab.rawValue)

                Text(headerSubtitle)
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textSecondary)
            }

            Spacer()

            Button(action: { Task { await appState.refreshDashboard() } }) {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: Design.IconSize.sm))
                    .foregroundColor(Design.Colors.textSecondary)
                    .rotationEffect(.degrees(appState.isRefreshing ? 360 : 0))
                    .animation(
                        appState.isRefreshing
                            ? .linear(duration: 1).repeatForever(autoreverses: false) : .default,
                        value: appState.isRefreshing)
            }
            .buttonStyle(.plain)
            .disabled(appState.isRefreshing)
            .accessibleButton(label: "Refresh", hint: "Refreshes dashboard data")

            if let action = primaryAction {
                Button(action: action.action) {
                    HStack(spacing: Design.Spacing.xxs) {
                        Image(systemName: action.icon)
                            .font(.system(size: Design.IconSize.xs, weight: .semibold))
                        Text(action.title)
                            .font(Design.Typography.labelMedium)
                    }
                    .foregroundColor(.white)
                    .padding(.horizontal, Design.Spacing.md)
                    .padding(.vertical, Design.Spacing.xs)
                    .background(
                        RoundedRectangle(cornerRadius: Design.Radius.sm)
                            .fill(
                                LinearGradient(
                                    colors: [action.color, action.color.opacity(0.8)],
                                    startPoint: .top,
                                    endPoint: .bottom
                                )
                            )
                            .shadow(color: action.color.opacity(0.3), radius: Design.Shadow.sm.radius, y: Design.Shadow.sm.y)
                    )
                }
                .buttonStyle(.plain)
                .accessibleButton(label: action.title)
            }
        }
        .padding(.horizontal, Design.Spacing.lg)
        .padding(.vertical, Design.Spacing.md)
    }

    private var headerSubtitle: String {
        switch appState.selectedTab {
        case .next: return "What's your next move?"
        case .gigs: return "Manage your gigs and tasks"
        case .dig: return "Analyze and incubate your ideas"
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
        case .hunter: return "Find and apply to freelance jobs"
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

#Preview {
    MainWindowView()
        .environmentObject(AppState())
        .frame(width: 1100, height: 700)
}
