//
//  AppState.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI
import Combine

// MARK: - App State Enum
enum AppStatus: Equatable {
    case loading
    case cliNotInstalled
    case notInitialized
    case ready
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

struct Client: Identifiable, Equatable {
    let id: Int
    let name: String
}

struct Contract: Identifiable, Equatable {
    let id: Int
    let name: String
    let clientName: String
}

// MARK: - App State
@MainActor
class AppState: ObservableObject {
    // Current status
    @Published var status: AppStatus = .loading

    // Tracking state
    @Published var isTracking: Bool = false
    @Published var activeSession: ActiveSession?

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

    // Settings
    @Published var useGlobalDatabase: Bool = true

    // Services
    let cliService = CLIService()

    // Timer for active tracking
    private var trackingTimer: Timer?
    private var cancellables = Set<AnyCancellable>()

    init() {
        checkStatus()
    }

    // MARK: - Status Check
    func checkStatus() {
        status = .loading

        Task {
            // Check if CLI is installed
            let isInstalled = await cliService.isCliInstalled()

            if !isInstalled {
                status = .cliNotInstalled
                return
            }

            // Check if initialized
            let isInitialized = await cliService.isInitialized()

            if !isInitialized {
                status = .notInitialized
                return
            }

            status = .ready
            await refreshDashboard()
        }
    }

    // MARK: - Dashboard Refresh
    func refreshDashboard() async {
        guard status == .ready else { return }
        isRefreshing = true

        async let metricsTask = loadMetrics()
        async let setupTask = loadSetupStatus()
        async let trackingTask = checkActiveTracking()
        async let invoicesTask = loadRecentInvoices()
        async let sessionsTask = loadRecentSessions()
        async let expensesTask = loadRecentExpenses()
        async let countsTask = loadCounts()
        async let clientsTask = loadClients()
        async let contractsTask = loadContracts()

        await metricsTask
        await setupTask
        await trackingTask
        await invoicesTask
        await sessionsTask
        await expensesTask
        await countsTask
        await clientsTask
        await contractsTask

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
            Task { @MainActor in
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
