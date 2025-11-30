//
//  AppState.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import Combine
import LocalAuthentication
import Security
import SwiftUI
import UserNotifications

// MARK: - App State Enum
enum AppStatus: Equatable {
    case loading
    case notInitialized
    case ready
}

// MARK: - Pomodoro State
struct PomodoroState: Equatable {
    var isActive: Bool = false
    var isBreak: Bool = false
    var isPaused: Bool = false
    var secondsRemaining: Int = 25 * 60
    var sessionsCompleted: Int = 0
    var workMinutes: Int = 25
    var breakMinutes: Int = 5
    var longBreakMinutes: Int = 15
    var sessionsUntilLongBreak: Int = 4

    var formattedTime: String {
        let minutes = secondsRemaining / 60
        let seconds = secondsRemaining % 60
        return String(format: "%02d:%02d", minutes, seconds)
    }

    var progress: Double {
        let total =
            isBreak
            ? (sessionsCompleted % sessionsUntilLongBreak == 0 ? longBreakMinutes : breakMinutes) * 60
            : workMinutes * 60
        return Double(total - secondsRemaining) / Double(total)
    }

    var statusText: String {
        if !isActive { return "Ready to focus" }
        if isPaused { return "Paused" }
        if isBreak {
            return sessionsCompleted % sessionsUntilLongBreak == 0 ? "Long Break" : "Short Break"
        }
        return "Focus Time"
    }
}

// MARK: - Data Models (View-friendly versions)
struct ActiveSession: Equatable {
    let id: Int
    let project: String
    let client: String
    let startTime: Date
    var elapsedSeconds: Int

    var formattedDuration: String {
        let hours = elapsedSeconds / 3600
        let minutes = (elapsedSeconds % 3600) / 60
        let seconds = elapsedSeconds % 60
        return String(format: "%d:%02d:%02d", hours, minutes, seconds)
    }
}

struct DashboardMetrics: Equatable {
    var totalRevenue: Double = 0
    var monthlyRevenue: Double = 0
    var pendingAmount: Double = 0
    var overdueAmount: Double = 0
    var weeklyHours: Double = 0
    var weeklyTarget: Double = 40
    var trackingStreak: Int = 0
}

struct RecentInvoice: Identifiable, Equatable {
    let id: Int
    let invoiceNum: String
    let client: String
    let amount: String
    let status: String
}

struct RecentSession: Identifiable, Equatable {
    let id: Int
    let project: String
    let duration: String
    let date: String
}

struct RecentExpense: Identifiable, Equatable {
    let id: Int
    let description: String
    let amount: String
    let category: String
    var date: String = ""
}

struct SetupStatus: Equatable {
    var hasCompany: Bool = false
    var hasClient: Bool = false
    var hasContract: Bool = false

    var isComplete: Bool {
        hasCompany && hasClient && hasContract
    }

    var nextStep: String {
        if !hasCompany { return "Create company profile" }
        if !hasClient { return "Add your first client" }
        if !hasContract { return "Create a contract" }
        return "Setup complete!"
    }
}

struct Client: Identifiable, Equatable, Hashable {
    let id: Int
    let name: String
    var email: String = ""
    var address: String = ""
    var taxId: String = ""
}

struct Contract: Identifiable, Equatable {
    let id: Int
    let name: String
    let clientName: String
    var rate: Double = 0
    var price: Double = 0
    var type: String = "hourly"
    var currency: String = "USD"
    var notes: String = ""
}

// MARK: - Keychain Manager
class KeychainManager: @unchecked Sendable {
    static let shared = KeychainManager()
    private let service = "com.ung.ung"
    private let passwordKey = "database_password"
    private let encryptionPasswordKey = "encryption_password"

    // MARK: - Generic Password Operations

    private func saveItem(_ key: String, value: String) -> Bool {
        guard let data = value.data(using: .utf8) else { return false }

        // Delete existing
        let deleteQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
        ]
        SecItemDelete(deleteQuery as CFDictionary)

        // Add new
        let addQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlocked,
        ]

        let status = SecItemAdd(addQuery as CFDictionary, nil)
        return status == errSecSuccess
    }

    private func getItem(_ key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne,
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
            let data = result as? Data,
            let value = String(data: data, encoding: .utf8)
        else {
            return nil
        }

        return value
    }

    private func deleteItem(_ key: String) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
        ]

        let status = SecItemDelete(query as CFDictionary)
        return status == errSecSuccess || status == errSecItemNotFound
    }

    private func hasItem(_ key: String) -> Bool {
        return getItem(key) != nil
    }

    // MARK: - Database Password (Legacy)

    func savePassword(_ password: String) -> Bool {
        return saveItem(passwordKey, value: password)
    }

    func getPassword() -> String? {
        return getItem(passwordKey)
    }

    func deletePassword() -> Bool {
        return deleteItem(passwordKey)
    }

    func hasPassword() -> Bool {
        return hasItem(passwordKey)
    }

    // MARK: - Encryption Password

    /// Save the database encryption password to keychain
    func saveEncryptionPassword(_ password: String) -> Bool {
        return saveItem(encryptionPasswordKey, value: password)
    }

    /// Get the database encryption password from keychain
    func getEncryptionPassword() -> String? {
        return getItem(encryptionPasswordKey)
    }

    /// Delete the database encryption password from keychain
    func deleteEncryptionPassword() -> Bool {
        return deleteItem(encryptionPasswordKey)
    }

    /// Check if encryption password is stored in keychain
    func hasEncryptionPassword() -> Bool {
        return hasItem(encryptionPasswordKey)
    }
}

// MARK: - App State
@MainActor
class AppState: ObservableObject {
    // Current status
    @Published var status: AppStatus = .loading

    // Tracking state
    @Published var isTracking: Bool = false
    @Published var activeSession: ActiveSession?

    // Pomodoro state
    @Published var pomodoroState: PomodoroState = PomodoroState()

    // Dashboard data
    @Published var metrics: DashboardMetrics = DashboardMetrics()
    @Published var setupStatus: SetupStatus = SetupStatus()
    @Published var onboardingSkipped: Bool = false
    @Published var hasSeenWelcome: Bool = false
    @Published var recentInvoices: [RecentInvoice] = []
    @Published var recentSessions: [RecentSession] = []
    @Published var recentExpenses: [RecentExpense] = []

    // Entity counts
    @Published var clientCount: Int = 0
    @Published var contractCount: Int = 0
    @Published var invoiceCount: Int = 0

    // Available clients and contracts for tracking
    @Published var clients: [Client] = []
    @Published var contracts: [Contract] = []

    // UI State
    @Published var secureMode: Bool = false
    @Published var isRefreshing: Bool = false
    @Published var showError: Bool = false
    @Published var errorMessage: String = ""
    @Published var showToast: Bool = false
    @Published var toastMessage: String = ""
    @Published var toastType: ToastType = .info
    @Published var showGlobalSearch: Bool = false

    enum ToastType {
        case success, info, warning, error
    }

    // App Lock
    @Published var isLocked: Bool = false
    @Published var appLockEnabled: Bool = false
    @Published var useTouchID: Bool = true
    @Published var isStartupLock: Bool = false

    // Settings
    @Published var hasStoredPassword: Bool = false
    @Published var databaseEncrypted: Bool = false
    @Published var hasEncryptionPassword: Bool = false
    @Published var encryptionStatus: EncryptionStatus = .disabled
    @Published var showPasswordPrompt: Bool = false
    @Published var pendingPasswordAction: PasswordAction?

    enum PasswordAction {
        case unlock
        case enable
        case disable
        case change
    }

    // iCloud Sync
    @Published var iCloudEnabled: Bool = false
    @Published var iCloudAvailable: Bool = false
    @Published var syncStatus: SyncStatus = .idle
    @Published var showSyncBanner: Bool = false

    // Main window navigation
    @Published var selectedTab: SidebarTab = .next {
        willSet {
            if newValue != selectedTab {
                objectWillChange.send()
            }
        }
    }

    // Services - using native database instead of CLI
    let database = DatabaseService.shared
    let keychain = KeychainManager.shared

    // Performance optimization
    private let refreshDebouncer = Debouncer(delay: 0.3)
    private let refreshThrottler = Throttler(minimumInterval: 0.5)
    private var lastRefreshTime: Date?
    private var pendingRefreshTask: Task<Void, Never>?

    // Timers
    private var trackingTimer: Timer?
    private var pomodoroTimer: Timer?
    private var cancellables = Set<AnyCancellable>()

    init() {
        hasStoredPassword = keychain.hasPassword()
        hasEncryptionPassword = keychain.hasEncryptionPassword()
        loadAppLockSettings()
        loadICloudSettings()
        loadEncryptionSettings()
        loadOnboardingStatus()
        checkStatus()
    }

    // MARK: - Encryption Settings
    private func loadEncryptionSettings() {
        databaseEncrypted = UserDefaults.standard.bool(forKey: "databaseEncryptionEnabled")
        Task {
            encryptionStatus = await database.checkEncryptionStatus()
            databaseEncrypted = encryptionStatus.isEnabled
        }
    }

    /// Check if database needs password to unlock
    func checkNeedsPassword() async -> Bool {
        return await database.needsPassword
    }

    /// Enable database encryption
    func enableEncryption(password: String) async {
        do {
            try await database.enableEncryption(password: password)
            // Save password to keychain
            _ = keychain.saveEncryptionPassword(password)
            hasEncryptionPassword = true
            databaseEncrypted = true
            encryptionStatus = .enabled
            showToastMessage("Database encryption enabled", type: .success)
        } catch {
            showError("Failed to enable encryption: \(error.localizedDescription)")
        }
    }

    /// Disable database encryption
    func disableEncryption(password: String) async {
        do {
            try await database.disableEncryption(password: password)
            // Remove password from keychain
            _ = keychain.deleteEncryptionPassword()
            hasEncryptionPassword = false
            databaseEncrypted = false
            encryptionStatus = .disabled
            showToastMessage("Database encryption disabled", type: .success)
        } catch {
            showError("Failed to disable encryption: \(error.localizedDescription)")
        }
    }

    /// Change encryption password
    func changeEncryptionPassword(currentPassword: String, newPassword: String) async {
        do {
            try await database.changePassword(currentPassword: currentPassword, newPassword: newPassword)
            // Update keychain
            _ = keychain.saveEncryptionPassword(newPassword)
            showToastMessage("Encryption password changed", type: .success)
        } catch {
            showError("Failed to change password: \(error.localizedDescription)")
        }
    }

    /// Unlock encrypted database with password
    func unlockDatabase(password: String) async -> Bool {
        let verified = await database.verifyPassword(password)
        if verified {
            await database.setPassword(password)
            do {
                try await database.initialize()
                // Save password to keychain for convenience
                _ = keychain.saveEncryptionPassword(password)
                hasEncryptionPassword = true
                showToastMessage("Database unlocked", type: .success)
                await refreshDashboard()
                return true
            } catch {
                showError("Failed to unlock database: \(error.localizedDescription)")
                return false
            }
        } else {
            showError("Incorrect password")
            return false
        }
    }

    /// Save encryption password to keychain
    func saveEncryptionPasswordToKeychain(_ password: String) -> Bool {
        let success = keychain.saveEncryptionPassword(password)
        if success {
            hasEncryptionPassword = true
        }
        return success
    }

    /// Clear encryption password from keychain
    func clearEncryptionPasswordFromKeychain() -> Bool {
        let success = keychain.deleteEncryptionPassword()
        if success {
            hasEncryptionPassword = false
        }
        return success
    }

    /// Get stored encryption password
    func getStoredEncryptionPassword() -> String? {
        return keychain.getEncryptionPassword()
    }

    /// Import an encrypted database from a file URL
    func importEncryptedDatabase(from url: URL, password: String) async throws {
        // Close current database
        await database.close()

        // Import the encrypted database
        try await database.importEncryptedDatabase(from: url, password: password)

        // Save password to keychain
        _ = keychain.saveEncryptionPassword(password)
        hasEncryptionPassword = true
        databaseEncrypted = true
        encryptionStatus = .enabled

        // Reinitialize
        checkStatus()
    }

    /// Export the current database (encrypted if encryption is enabled)
    func exportDatabase(to url: URL, includeEncryption: Bool = true) async throws {
        try await database.exportDatabase(to: url)
    }

    // MARK: - iCloud Settings
    private func loadICloudSettings() {
        iCloudEnabled = UserDefaults.standard.bool(forKey: "iCloudSyncEnabled")
        Task {
            iCloudAvailable = await database.isICloudAvailable
        }
    }

    func setICloudEnabled(_ enabled: Bool) async {
        do {
            try await database.setICloudEnabled(enabled)
            iCloudEnabled = enabled
            await refreshDashboard()
            showToastMessage(enabled ? "iCloud sync enabled" : "iCloud sync disabled", type: .success)
        } catch {
            showError("Failed to \(enabled ? "enable" : "disable") iCloud sync: \(error.localizedDescription)")
        }
    }

    func checkICloudSync() {
        guard iCloudEnabled else { return }

        Task {
            showSyncBanner = true
            syncStatus = .syncing

            let status = await database.triggerSync()
            syncStatus = status

            // Keep banner visible for a moment
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            showSyncBanner = false

            if case .completed = status {
                await refreshDashboard()
            }
        }
    }

    // MARK: - App Lock Settings
    private func loadAppLockSettings() {
        appLockEnabled = UserDefaults.standard.bool(forKey: "appLockEnabled")
        useTouchID = UserDefaults.standard.object(forKey: "useTouchID") as? Bool ?? true
        if appLockEnabled {
            isLocked = true
            isStartupLock = true
        }
    }

    func setAppLockEnabled(_ enabled: Bool) {
        appLockEnabled = enabled
        UserDefaults.standard.set(enabled, forKey: "appLockEnabled")
        if enabled {
            isLocked = true
        }
    }

    func setUseTouchID(_ enabled: Bool) {
        useTouchID = enabled
        UserDefaults.standard.set(enabled, forKey: "useTouchID")
    }

    // MARK: - Authentication
    func authenticateWithBiometrics() async -> Bool {
        let context = LAContext()
        var error: NSError?

        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            return await authenticateWithPassword()
        }

        do {
            let success = try await context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: "Unlock UNG to access your data"
            )
            if success {
                isLocked = false
                isStartupLock = false
            }
            return success
        } catch {
            return false
        }
    }

    func authenticateWithPassword() async -> Bool {
        let context = LAContext()
        var error: NSError?

        guard context.canEvaluatePolicy(.deviceOwnerAuthentication, error: &error) else {
            return false
        }

        do {
            let success = try await context.evaluatePolicy(
                .deviceOwnerAuthentication,
                localizedReason: "Enter your password to unlock UNG"
            )
            if success {
                isLocked = false
                isStartupLock = false
            }
            return success
        } catch {
            return false
        }
    }

    func lockApp() {
        if appLockEnabled {
            isLocked = true
            isStartupLock = false
        }
    }

    // MARK: - Status Check
    func checkStatus() {
        print("[AppState] Starting status check...")
        status = .loading

        Task {
            // Check if database is encrypted and needs password
            let needsPassword = await database.needsPassword

            if needsPassword {
                // Try to use stored password from keychain
                if let storedPassword = keychain.getEncryptionPassword() {
                    await database.setPassword(storedPassword)
                    do {
                        try await database.initialize()
                        print("[AppState] Database initialized with stored password")
                        encryptionStatus = .enabled
                        databaseEncrypted = true
                        status = .ready
                        await refreshDashboard()
                    } catch {
                        print("[AppState] Failed to initialize with stored password: \(error)")
                        // Password might be wrong, prompt user
                        showPasswordPrompt = true
                        pendingPasswordAction = .unlock
                        status = .notInitialized
                    }
                } else {
                    // No stored password, prompt user
                    print("[AppState] Encrypted database found, password required")
                    showPasswordPrompt = true
                    pendingPasswordAction = .unlock
                    encryptionStatus = .enabled
                    databaseEncrypted = true
                    status = .notInitialized
                }
            } else {
                // Auto-initialize the database - no CLI required!
                do {
                    try await database.initialize()
                    print("[AppState] Database initialized successfully")
                    encryptionStatus = await database.checkEncryptionStatus()
                    databaseEncrypted = encryptionStatus.isEnabled
                    status = .ready
                    await refreshDashboard()
                } catch {
                    print("[AppState] Failed to initialize database: \(error)")
                    status = .notInitialized
                }
            }
        }
    }

    // MARK: - Toast
    func showToastMessage(_ message: String, type: ToastType = .info) {
        toastMessage = message
        toastType = type
        showToast = true

        Task {
            try? await Task.sleep(nanoseconds: 3_000_000_000)
            showToast = false
        }
    }

    // MARK: - Onboarding

    /// Skip onboarding and show main dashboard
    /// User can always complete setup later from Settings
    func skipOnboarding() {
        onboardingSkipped = true
        UserDefaults.standard.set(true, forKey: "onboardingSkipped")
    }

    /// Load all onboarding-related status from UserDefaults
    func loadOnboardingStatus() {
        onboardingSkipped = UserDefaults.standard.bool(forKey: "onboardingSkipped")
        hasSeenWelcome = UserDefaults.standard.bool(forKey: "hasSeenWelcome")
    }

    /// Reset onboarding skip (useful if user wants to see onboarding again)
    func resetOnboardingSkip() {
        onboardingSkipped = false
        UserDefaults.standard.set(false, forKey: "onboardingSkipped")
    }

    /// Complete the welcome walkthrough and proceed to main app
    func completeWelcomeWalkthrough() {
        hasSeenWelcome = true
        UserDefaults.standard.set(true, forKey: "hasSeenWelcome")
    }

    /// Reset welcome walkthrough (useful for testing or re-showing)
    func resetWelcomeWalkthrough() {
        hasSeenWelcome = false
        UserDefaults.standard.set(false, forKey: "hasSeenWelcome")
    }

    // MARK: - Dashboard Refresh

    /// Debounced refresh - use this for UI-triggered refreshes
    func requestRefresh() {
        refreshDebouncer.debounce { [weak self] in
            Task { @MainActor in
                await self?.refreshDashboard()
            }
        }
    }

    /// Core refresh implementation with throttling
    func refreshDashboard() async {
        guard status == .ready else { return }

        // Throttle: Skip if we refreshed recently (within 0.5s)
        if let lastTime = lastRefreshTime,
           Date().timeIntervalSince(lastTime) < 0.5 {
            return
        }

        // Cancel any pending refresh
        pendingRefreshTask?.cancel()

        isRefreshing = true
        lastRefreshTime = Date()

        // Use PerformanceMonitor in debug to track slow operations
        #if DEBUG
        let start = CFAbsoluteTimeGetCurrent()
        #endif

        await withTaskGroup(of: Void.self) { group in
            group.addTask { await self.loadMetrics() }
            group.addTask { await self.loadSetupStatus() }
            group.addTask { await self.checkActiveTracking() }
            group.addTask { await self.loadRecentInvoices() }
            group.addTask { await self.loadRecentSessions() }
            group.addTask { await self.loadRecentExpenses() }
            group.addTask { await self.loadCounts() }
            group.addTask { await self.loadClients() }
            group.addTask { await self.loadContracts() }
        }

        #if DEBUG
        let elapsed = CFAbsoluteTimeGetCurrent() - start
        if elapsed > 0.1 {
            print("⚠️ Dashboard refresh took \(String(format: "%.3f", elapsed))s")
        }
        #endif

        isRefreshing = false

        // Sync data to widgets
        syncWidgetData()
    }

    // MARK: - Widget Data Sync
    private func syncWidgetData() {
        #if os(iOS)
        // Update stats for widgets
        WidgetDataManager.shared.updateStats(
            todayHours: getTodayHours(),
            weeklyHours: metrics.weeklyHours,
            weeklyTarget: metrics.weeklyTarget,
            pendingInvoices: invoiceCount,
            pendingAmount: metrics.pendingAmount
        )

        // Update tracking status
        if let session = activeSession {
            WidgetDataManager.shared.updateTrackingStatus(
                isTracking: true,
                project: session.project,
                client: session.client,
                startTime: session.startTime
            )
        } else {
            WidgetDataManager.shared.updateTrackingStatus(
                isTracking: false,
                project: "",
                client: "",
                startTime: nil
            )
        }

        // Update pomodoro status
        WidgetDataManager.shared.updatePomodoro(
            active: pomodoroState.isActive,
            isBreak: pomodoroState.isBreak,
            secondsRemaining: pomodoroState.secondsRemaining,
            sessionsCompleted: pomodoroState.sessionsCompleted
        )
        #endif
    }

    private func getTodayHours() -> Double {
        // Calculate hours tracked today from recent sessions
        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())
        var todayHours: Double = 0

        for session in recentSessions {
            // Parse the duration string (format: "Xh Xm")
            let components = session.duration.components(separatedBy: " ")
            if components.count >= 1 {
                if let hours = Int(components[0].replacingOccurrences(of: "h", with: "")) {
                    todayHours += Double(hours)
                }
                if components.count >= 2, let minutes = Int(components[1].replacingOccurrences(of: "m", with: "")) {
                    todayHours += Double(minutes) / 60.0
                }
            }
        }

        // Add current active session if any
        if let session = activeSession {
            todayHours += Double(session.elapsedSeconds) / 3600.0
        }

        return todayHours
    }

    // MARK: - Load Data (using native database)
    private func loadMetrics() async {
        do {
            let result = try await database.getDashboardMetrics()
            metrics.totalRevenue = result.totalRevenue
            metrics.pendingAmount = result.pendingAmount
            metrics.overdueAmount = result.overdueAmount
            metrics.weeklyHours = result.weeklyHours
        } catch {
            print("[AppState] Failed to load metrics: \(error)")
        }
    }

    private func loadSetupStatus() async {
        do {
            let company = try await database.getCompany()
            setupStatus.hasCompany = company != nil

            let clientCount = try await database.getClientCount()
            setupStatus.hasClient = clientCount > 0

            let contractCount = try await database.getContractCount()
            setupStatus.hasContract = contractCount > 0
        } catch {
            print("[AppState] Failed to load setup status: \(error)")
        }
    }

    private func checkActiveTracking() async {
        do {
            // Optimized: Single query with JOIN instead of N+1
            if let result = try await database.getActiveSessionWithClient() {
                activeSession = ActiveSession(
                    id: Int(result.session.id ?? 0),
                    project: result.session.projectName ?? "Active Session",
                    client: result.clientName ?? "",
                    startTime: result.session.startTime,
                    elapsedSeconds: result.session.calculatedDuration
                )
                isTracking = true
                startTrackingTimer()
            } else {
                activeSession = nil
                isTracking = false
                stopTrackingTimer()
            }
        } catch {
            print("[AppState] Failed to check active tracking: \(error)")
            activeSession = nil
            isTracking = false
        }
    }

    private func loadRecentInvoices() async {
        do {
            let invoices = try await database.getRecentInvoices(limit: 3)
            // Use pre-configured cached formatter for performance
            recentInvoices = invoices.map { invoice in
                RecentInvoice(
                    id: Int(invoice.id ?? 0),
                    invoiceNum: invoice.invoiceNum,
                    client: "",
                    amount: Formatters.currency.string(from: NSNumber(value: invoice.amount)) ?? "$0.00",
                    status: invoice.status
                )
            }
        } catch {
            print("[AppState] Failed to load recent invoices: \(error)")
        }
    }

    private func loadRecentSessions() async {
        do {
            let sessions = try await database.getRecentSessions(limit: 5)
            // Use pre-configured cached formatter for performance
            recentSessions = sessions.compactMap { session in
                guard let id = session.id else { return nil }
                return RecentSession(
                    id: Int(id),
                    project: session.projectName ?? "Session",
                    duration: session.formattedDuration,
                    date: Formatters.shortDate.string(from: session.startTime)
                )
            }
        } catch {
            print("[AppState] Failed to load recent sessions: \(error)")
        }
    }

    private func loadRecentExpenses() async {
        do {
            let expenses = try await database.getRecentExpenses(limit: 5)
            // Use pre-configured cached formatter for performance
            recentExpenses = expenses.compactMap { expense in
                guard let id = expense.id else { return nil }
                return RecentExpense(
                    id: Int(id),
                    description: expense.description,
                    amount: Formatters.currency.string(from: NSNumber(value: expense.amount)) ?? "$0.00",
                    category: expense.category,
                    date: Formatters.shortDate.string(from: expense.date)
                )
            }
        } catch {
            print("[AppState] Failed to load recent expenses: \(error)")
        }
    }

    private func loadCounts() async {
        do {
            clientCount = try await database.getClientCount()
            contractCount = try await database.getContractCount()
            invoiceCount = try await database.getInvoiceCount()
        } catch {
            print("[AppState] Failed to load counts: \(error)")
        }
    }

    private func loadClients() async {
        do {
            let dbClients = try await database.getClients()
            clients = dbClients.compactMap { client in
                guard let id = client.id else { return nil }
                return Client(
                    id: Int(id),
                    name: client.name,
                    email: client.email,
                    address: client.address ?? "",
                    taxId: client.taxId ?? ""
                )
            }
        } catch {
            print("[AppState] Failed to load clients: \(error)")
        }
    }

    private func loadContracts() async {
        do {
            let dbContracts = try await database.getContracts()
            contracts = dbContracts.compactMap { contract in
                guard let id = contract.id else { return nil }
                return Contract(
                    id: Int(id),
                    name: contract.name,
                    clientName: "",
                    rate: contract.hourlyRate ?? 0,
                    price: contract.fixedPrice ?? 0,
                    type: contract.contractType,
                    currency: contract.currency,
                    notes: contract.notes ?? ""
                )
            }
        } catch {
            print("[AppState] Failed to load contracts: \(error)")
        }
    }

    // MARK: - Tracking Timer
    private func startTrackingTimer() {
        stopTrackingTimer()
        trackingTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                guard let self = self, var session = self.activeSession else { return }
                session.elapsedSeconds += 1
                self.activeSession = session
            }
        }
    }

    private func stopTrackingTimer() {
        trackingTimer?.invalidate()
        trackingTimer = nil
    }

    // MARK: - Tracking Actions
    func startTracking(project: String, clientId: Int?) async {
        do {
            _ = try await database.startTracking(
                projectName: project,
                clientId: clientId != nil ? Int64(clientId!) : nil
            )
            await checkActiveTracking()

            // Start Live Activity for real-time display
            #if os(iOS)
            if let session = activeSession {
                LiveActivityService.shared.startTrackingActivity(
                    project: session.project,
                    client: session.client,
                    startTime: session.startTime
                )
            }
            #endif

            showToastMessage("Started tracking: \(project)", type: .success)
        } catch {
            showError("Failed to start tracking: \(error.localizedDescription)")
        }
    }

    func stopTracking() async {
        guard let session = activeSession else { return }
        do {
            _ = try await database.stopTracking(sessionId: Int64(session.id))
            isTracking = false
            activeSession = nil
            stopTrackingTimer()

            // End Live Activity
            #if os(iOS)
            LiveActivityService.shared.endTrackingActivity()
            #endif

            await refreshDashboard()
            showToastMessage("Tracking stopped", type: .success)
        } catch {
            showError("Failed to stop tracking: \(error.localizedDescription)")
        }
    }

    // MARK: - Pomodoro
    func startPomodoro() {
        pomodoroState.isActive = true
        pomodoroState.isPaused = false
        pomodoroState.isBreak = false
        pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
        startPomodoroTimer()

        // Start Live Activity
        #if os(iOS)
        LiveActivityService.shared.startPomodoroActivity(
            sessionsCompleted: pomodoroState.sessionsCompleted,
            workMinutes: pomodoroState.workMinutes,
            breakMinutes: pomodoroState.breakMinutes,
            secondsRemaining: pomodoroState.secondsRemaining,
            isBreak: false
        )
        #endif
    }

    func pausePomodoro() {
        pomodoroState.isPaused = true

        // Update Live Activity
        #if os(iOS)
        LiveActivityService.shared.updatePomodoroActivity(
            secondsRemaining: pomodoroState.secondsRemaining,
            isBreak: pomodoroState.isBreak,
            isPaused: true
        )
        #endif
        stopPomodoroTimer()
    }

    func resumePomodoro() {
        pomodoroState.isPaused = false
        startPomodoroTimer()

        // Update Live Activity
        #if os(iOS)
        LiveActivityService.shared.updatePomodoroActivity(
            secondsRemaining: pomodoroState.secondsRemaining,
            isBreak: pomodoroState.isBreak,
            isPaused: false
        )
        #endif
    }

    func stopPomodoro() {
        pomodoroState.isActive = false
        pomodoroState.isPaused = false
        pomodoroState.isBreak = false
        pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
        stopPomodoroTimer()

        // End Live Activity
        #if os(iOS)
        LiveActivityService.shared.endPomodoroActivity()
        #endif
    }

    func skipPomodoro() {
        if pomodoroState.isBreak {
            pomodoroState.isBreak = false
            pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
        } else {
            pomodoroState.sessionsCompleted += 1
            pomodoroState.isBreak = true
            let isLongBreak = pomodoroState.sessionsCompleted % pomodoroState.sessionsUntilLongBreak == 0
            pomodoroState.secondsRemaining =
                (isLongBreak ? pomodoroState.longBreakMinutes : pomodoroState.breakMinutes) * 60
        }

        // Update Live Activity
        #if os(iOS)
        LiveActivityService.shared.updatePomodoroActivity(
            secondsRemaining: pomodoroState.secondsRemaining,
            isBreak: pomodoroState.isBreak,
            isPaused: false
        )
        #endif
    }

    private func startPomodoroTimer() {
        stopPomodoroTimer()
        pomodoroTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                guard let self = self else { return }
                if self.pomodoroState.secondsRemaining > 0 {
                    self.pomodoroState.secondsRemaining -= 1

                    // Update Live Activity every second
                    #if os(iOS)
                    LiveActivityService.shared.updatePomodoroActivity(
                        secondsRemaining: self.pomodoroState.secondsRemaining,
                        isBreak: self.pomodoroState.isBreak,
                        isPaused: false
                    )
                    #endif
                } else {
                    self.pomodoroCompleted()
                }
            }
        }
    }

    private func stopPomodoroTimer() {
        pomodoroTimer?.invalidate()
        pomodoroTimer = nil
    }

    private func pomodoroCompleted() {
        #if os(macOS)
        NSSound.beep()
        #endif

        let isLongBreak = pomodoroState.sessionsCompleted % pomodoroState.sessionsUntilLongBreak == 0

        if pomodoroState.isBreak {
            // Break just finished, back to work
            pomodoroState.isBreak = false
            pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
            // Notify that break is complete
            NotificationService.shared.schedulePomodoroComplete(isBreak: true, isLongBreak: isLongBreak, delay: 0.5)
        } else {
            // Work session finished, start break
            pomodoroState.sessionsCompleted += 1
            pomodoroState.isBreak = true
            let isNowLongBreak = pomodoroState.sessionsCompleted % pomodoroState.sessionsUntilLongBreak == 0
            pomodoroState.secondsRemaining =
                (isNowLongBreak ? pomodoroState.longBreakMinutes : pomodoroState.breakMinutes) * 60
            // Notify that work session is complete
            NotificationService.shared.schedulePomodoroComplete(isBreak: false, isLongBreak: isNowLongBreak, delay: 0.5)
        }

        // Update Live Activity with new state
        #if os(iOS)
        LiveActivityService.shared.updatePomodoroActivity(
            secondsRemaining: pomodoroState.secondsRemaining,
            isBreak: pomodoroState.isBreak,
            isPaused: false
        )
        #endif
    }

    // MARK: - Password Management
    func savePassword(_ password: String) -> Bool {
        let success = keychain.savePassword(password)
        if success {
            hasStoredPassword = true
        }
        return success
    }

    func getStoredPassword() -> String? {
        return keychain.getPassword()
    }

    func clearPassword() -> Bool {
        let success = keychain.deletePassword()
        if success {
            hasStoredPassword = false
        }
        return success
    }

    // MARK: - Initialization (now automatic)
    func initializeDatabase() async {
        do {
            try await database.initialize()
            checkStatus()
        } catch {
            showError("Failed to initialize database: \(error.localizedDescription)")
        }
    }

    // MARK: - Error Handling
    func showError(_ message: String) {
        errorMessage = message
        showError = true
    }

    // MARK: - Utility
    func formatCurrency(_ amount: Double) -> String {
        if secureMode {
            return "****"
        }
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        return formatter.string(from: NSNumber(value: amount)) ?? "$0.00"
    }

    func formatHours(_ hours: Double) -> String {
        if secureMode {
            return "**:**"
        }
        let h = Int(hours)
        let m = Int((hours - Double(h)) * 60)
        return String(format: "%dh %02dm", h, m)
    }
}

// MARK: - Sidebar Tab
enum SidebarTab: String, CaseIterable, Identifiable {
    case next = "Next"
    case gigs = "Gigs"
    case dashboard = "Dashboard"
    case tracking = "Time Tracking"
    case clients = "Clients"
    case contracts = "Contracts"
    case invoices = "Invoices"
    case expenses = "Expenses"
    case pomodoro = "Focus Timer"
    case hunter = "Job Hunter"
    case reports = "Reports"
    case settings = "Settings"

    var id: String { rawValue }

    var icon: String {
        switch self {
        case .next: return "arrow.right.circle"
        case .gigs: return "rectangle.3.group"
        case .dashboard: return "square.grid.2x2"
        case .tracking: return "clock"
        case .clients: return "person.2"
        case .contracts: return "doc.text"
        case .invoices: return "doc.plaintext"
        case .expenses: return "dollarsign.circle"
        case .pomodoro: return "brain.head.profile"
        case .hunter: return "magnifyingglass"
        case .reports: return "chart.bar"
        case .settings: return "gearshape"
        }
    }

    var iconFilled: String {
        switch self {
        case .next: return "arrow.right.circle.fill"
        case .gigs: return "rectangle.3.group.fill"
        case .dashboard: return "square.grid.2x2.fill"
        case .tracking: return "clock.fill"
        case .clients: return "person.2.fill"
        case .contracts: return "doc.text.fill"
        case .invoices: return "doc.plaintext.fill"
        case .expenses: return "dollarsign.circle.fill"
        case .pomodoro: return "brain.head.profile.fill"
        case .hunter: return "magnifyingglass.circle.fill"
        case .reports: return "chart.bar.fill"
        case .settings: return "gearshape.fill"
        }
    }

    var shortLabel: String {
        switch self {
        case .next: return "Next"
        case .gigs: return "Gigs"
        case .dashboard: return "Stats"
        case .tracking: return "Track"
        case .clients: return "Clients"
        case .contracts: return "Contracts"
        case .invoices: return "Invoices"
        case .expenses: return "Expenses"
        case .pomodoro: return "Focus"
        case .hunter: return "Jobs"
        case .reports: return "Reports"
        case .settings: return "Settings"
        }
    }

    static var accentColor: Color { Design.Colors.primary }
}
