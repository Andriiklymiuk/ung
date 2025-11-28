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
  case cliNotInstalled
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

// MARK: - Data Models
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
}

struct Contract: Identifiable, Equatable {
  let id: Int
  let name: String
  let clientName: String
  var rate: Double = 0
  var type: String = "hourly"
}

// MARK: - Keychain Manager
class KeychainManager {
  static let shared = KeychainManager()
  private let service = "com.ung.ung"
  private let passwordKey = "database_password"

  func savePassword(_ password: String) -> Bool {
    guard let data = password.data(using: .utf8) else { return false }

    // Delete existing
    let deleteQuery: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrService as String: service,
      kSecAttrAccount as String: passwordKey,
    ]
    SecItemDelete(deleteQuery as CFDictionary)

    // Add new
    let addQuery: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrService as String: service,
      kSecAttrAccount as String: passwordKey,
      kSecValueData as String: data,
      kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlocked,
    ]

    let status = SecItemAdd(addQuery as CFDictionary, nil)
    return status == errSecSuccess
  }

  func getPassword() -> String? {
    let query: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrService as String: service,
      kSecAttrAccount as String: passwordKey,
      kSecReturnData as String: true,
      kSecMatchLimit as String: kSecMatchLimitOne,
    ]

    var result: AnyObject?
    let status = SecItemCopyMatching(query as CFDictionary, &result)

    guard status == errSecSuccess,
      let data = result as? Data,
      let password = String(data: data, encoding: .utf8)
    else {
      return nil
    }

    return password
  }

  func deletePassword() -> Bool {
    let query: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrService as String: service,
      kSecAttrAccount as String: passwordKey,
    ]

    let status = SecItemDelete(query as CFDictionary)
    return status == errSecSuccess || status == errSecItemNotFound
  }

  func hasPassword() -> Bool {
    return getPassword() != nil
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

  // Main window navigation - optimize with equatable to reduce redraws
  @Published var selectedTab: SidebarTab = .dashboard {
    willSet {
      // Only update if actually changing
      if newValue != selectedTab {
        objectWillChange.send()
      }
    }
  }

  // Services
  let cliService = CLIService()
  let keychain = KeychainManager.shared

  // Timers
  private var trackingTimer: Timer?
  private var pomodoroTimer: Timer?
  private var cancellables = Set<AnyCancellable>()

  init() {
    hasStoredPassword = keychain.hasPassword()
    loadAppLockSettings()
    checkStatus()
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
      // Biometrics not available, try password
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
      // Biometrics failed, fall back to password
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
      // Check if CLI is installed
      print("[AppState] Checking CLI installation...")
      let isInstalled = await cliService.isCliInstalled()
      print("[AppState] CLI installed: \(isInstalled)")

      if !isInstalled {
        print("[AppState] Setting status to .cliNotInstalled")
        status = .cliNotInstalled
        return
      }

      // Check if initialized
      print("[AppState] Checking initialization...")
      let isInitialized = await cliService.isInitialized()
      print("[AppState] Initialized: \(isInitialized)")

      if !isInitialized {
        print("[AppState] Setting status to .notInitialized")
        status = .notInitialized
        return
      }

      print("[AppState] Setting status to .ready")
      status = .ready
      await refreshDashboard()
    }
  }

  // MARK: - Toast
  func showToastMessage(_ message: String, type: ToastType = .info) {
    toastMessage = message
    toastType = type
    showToast = true

    // Auto-hide after 3 seconds
    Task {
      try? await Task.sleep(nanoseconds: 3_000_000_000)
      showToast = false
    }
  }

  // MARK: - Dashboard Refresh
  func refreshDashboard() async {
    guard status == .ready else { return }
    isRefreshing = true

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

    isRefreshing = false
  }

  // MARK: - Load Data
  private func loadMetrics() async {
    if let result = await cliService.getDashboardMetrics() {
      metrics = result
    }
  }

  private func loadSetupStatus() async {
    setupStatus = await cliService.getSetupStatus()
  }

  private func checkActiveTracking() async {
    if let session = await cliService.getActiveSession() {
      activeSession = session
      isTracking = true
      startTrackingTimer()
    } else {
      activeSession = nil
      isTracking = false
      stopTrackingTimer()
    }
  }

  private func loadRecentInvoices() async {
    recentInvoices = await cliService.getRecentInvoices()
  }

  private func loadRecentSessions() async {
    recentSessions = await cliService.getRecentSessions()
  }

  private func loadRecentExpenses() async {
    recentExpenses = await cliService.getRecentExpenses()
  }

  private func loadCounts() async {
    let counts = await cliService.getEntityCounts()
    clientCount = counts.clients
    contractCount = counts.contracts
    invoiceCount = counts.invoices
  }

  private func loadClients() async {
    clients = await cliService.getClients()
  }

  private func loadContracts() async {
    contracts = await cliService.getContracts()
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
    let success = await cliService.startTracking(project: project, clientId: clientId)
    if success {
      await checkActiveTracking()
    }
  }

  func stopTracking() async {
    let success = await cliService.stopTracking()
    if success {
      isTracking = false
      activeSession = nil
      stopTrackingTimer()
      await refreshDashboard()
    }
  }

  // MARK: - Pomodoro
  func startPomodoro() {
    pomodoroState.isActive = true
    pomodoroState.isPaused = false
    pomodoroState.isBreak = false
    pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
    startPomodoroTimer()
  }

  func pausePomodoro() {
    pomodoroState.isPaused = true
    stopPomodoroTimer()
  }

  func resumePomodoro() {
    pomodoroState.isPaused = false
    startPomodoroTimer()
  }

  func stopPomodoro() {
    pomodoroState.isActive = false
    pomodoroState.isPaused = false
    pomodoroState.isBreak = false
    pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
    stopPomodoroTimer()
  }

  func skipPomodoro() {
    if pomodoroState.isBreak {
      // Skip break, start new work session
      pomodoroState.isBreak = false
      pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
    } else {
      // Skip work, start break
      pomodoroState.sessionsCompleted += 1
      pomodoroState.isBreak = true
      let isLongBreak = pomodoroState.sessionsCompleted % pomodoroState.sessionsUntilLongBreak == 0
      pomodoroState.secondsRemaining =
        (isLongBreak ? pomodoroState.longBreakMinutes : pomodoroState.breakMinutes) * 60
    }
  }

  private func startPomodoroTimer() {
    stopPomodoroTimer()
    pomodoroTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
      Task { @MainActor [weak self] in
        guard let self = self else { return }
        if self.pomodoroState.secondsRemaining > 0 {
          self.pomodoroState.secondsRemaining -= 1
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
    // Play notification sound
    NSSound.beep()

    if pomodoroState.isBreak {
      // Break completed, start new work session
      pomodoroState.isBreak = false
      pomodoroState.secondsRemaining = pomodoroState.workMinutes * 60
    } else {
      // Work completed, start break
      pomodoroState.sessionsCompleted += 1
      pomodoroState.isBreak = true
      let isLongBreak = pomodoroState.sessionsCompleted % pomodoroState.sessionsUntilLongBreak == 0
      pomodoroState.secondsRemaining =
        (isLongBreak ? pomodoroState.longBreakMinutes : pomodoroState.breakMinutes) * 60

      // Send notification
      sendPomodoroNotification(isLongBreak: isLongBreak)
    }
  }

  private func sendPomodoroNotification(isLongBreak: Bool) {
    let content = UNMutableNotificationContent()
    content.title = "Pomodoro Complete!"
    content.body =
      isLongBreak ? "Great work! Take a long break." : "Nice focus session! Take a short break."
    content.sound = .default

    let request = UNNotificationRequest(
      identifier: UUID().uuidString, content: content, trigger: nil)
    UNUserNotificationCenter.current().add(request)
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

  // MARK: - Initialization
  func initializeGlobal() async {
    let success = await cliService.initializeGlobal()
    if success {
      checkStatus()
    }
  }

  func initializeLocal() async {
    let success = await cliService.initializeLocal()
    if success {
      checkStatus()
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
  case dashboard = "Dashboard"
  case tracking = "Time Tracking"
  case clients = "Clients"
  case contracts = "Contracts"
  case invoices = "Invoices"
  case expenses = "Expenses"
  case pomodoro = "Focus Timer"
  case reports = "Reports"
  case settings = "Settings"

  var id: String { rawValue }

  var icon: String {
    switch self {
    case .dashboard: return "square.grid.2x2"
    case .tracking: return "clock"
    case .clients: return "person.2"
    case .contracts: return "doc.text"
    case .invoices: return "doc.plaintext"
    case .expenses: return "dollarsign.circle"
    case .pomodoro: return "brain.head.profile"
    case .reports: return "chart.bar"
    case .settings: return "gearshape"
    }
  }

  var iconFilled: String {
    switch self {
    case .dashboard: return "square.grid.2x2.fill"
    case .tracking: return "clock.fill"
    case .clients: return "person.2.fill"
    case .contracts: return "doc.text.fill"
    case .invoices: return "doc.plaintext.fill"
    case .expenses: return "dollarsign.circle.fill"
    case .pomodoro: return "brain.head.profile.fill"
    case .reports: return "chart.bar.fill"
    case .settings: return "gearshape.fill"
    }
  }

  /// Short label for sidebar display
  var shortLabel: String {
    switch self {
    case .dashboard: return "Home"
    case .tracking: return "Track"
    case .clients: return "Clients"
    case .contracts: return "Contracts"
    case .invoices: return "Invoices"
    case .expenses: return "Expenses"
    case .pomodoro: return "Focus"
    case .reports: return "Reports"
    case .settings: return "Settings"
    }
  }

  /// Single accent color for selected state (Slack-style)
  static let accentColor = Color.accentColor
}
