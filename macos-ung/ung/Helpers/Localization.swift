//
//  Localization.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import Foundation

// MARK: - Localization Helper
/// Simple localization helper for easy string localization.
/// Usage: L10n.welcomeTitle or "Welcome".localized
extension String {
  /// Returns the localized version of the string
  var localized: String {
    NSLocalizedString(self, comment: "")
  }

  /// Returns a localized string with format arguments
  func localized(with arguments: CVarArg...) -> String {
    String(format: self.localized, arguments: arguments)
  }
}

// MARK: - Localization Keys
/// Centralized localization keys for type-safe access
enum L10n {
  // MARK: - Common
  static let ok = "OK".localized
  static let cancel = "Cancel".localized
  static let save = "Save".localized
  static let delete = "Delete".localized
  static let edit = "Edit".localized
  static let done = "Done".localized
  static let close = "Close".localized
  static let reset = "Reset".localized
  static let search = "Search".localized
  static let loading = "Loading...".localized
  static let error = "Error".localized
  static let success = "Success".localized

  // MARK: - Navigation
  enum Navigation {
    static let home = "Home".localized
    static let track = "Track".localized
    static let clients = "Clients".localized
    static let contracts = "Contracts".localized
    static let invoices = "Invoices".localized
    static let expenses = "Expenses".localized
    static let focus = "Focus".localized
    static let reports = "Reports".localized
    static let settings = "Settings".localized
  }

  // MARK: - Onboarding
  enum Onboarding {
    static let welcomeTitle = "Welcome to UNG".localized
    static let welcomeSubtitle = "Let's get you set up in 3 easy steps".localized
    static let stepCompany = "Set Up Your Company".localized
    static let stepCompanyDesc = "Add your business details like name, address, and logo".localized
    static let stepClient = "Add Your First Client".localized
    static let stepClientDesc = "Who will you be working with?".localized
    static let stepContract = "Create a Contract".localized
    static let stepContractDesc = "Define your working terms and rates".localized
    static let quickActions = "Quick Actions".localized
    static let completeStep = "Complete step %d".localized
    static let setUp = "Set Up".localized

    static func ofComplete(_ completed: Int, _ total: Int) -> String {
      return "%d of %d complete".localized(with: completed, total)
    }
  }

  // MARK: - Dashboard
  enum Dashboard {
    static let totalRevenue = "Total Revenue".localized
    static let thisMonth = "This Month".localized
    static let pending = "Pending".localized
    static let hoursThisWeek = "Hours This Week".localized
    static let quickActions = "Quick Actions".localized
    static let recentActivity = "Recent Activity".localized
    static let weeklyProgress = "Weekly Progress".localized
    static let setupChecklist = "Setup Checklist".localized
    static let unpaidInvoices = "Unpaid Invoices".localized
    static let noRecentActivity = "No recent activity".localized
    static let viewInvoices = "View Invoices".localized
    static let overdue = "Overdue".localized
  }

  // MARK: - Quick Actions
  enum QuickActions {
    static let startTracking = "Start Tracking".localized
    static let trackYourTime = "Track your time".localized
    static let createInvoice = "Create Invoice".localized
    static let billYourClients = "Bill your clients".localized
    static let logExpense = "Log Expense".localized
    static let trackSpending = "Track spending".localized
    static let focusMode = "Focus Mode".localized
    static let startPomodoro = "Start Pomodoro".localized
    static let viewReports = "View Reports".localized
  }

  // MARK: - Tracking
  enum Tracking {
    static let startSession = "Start Session".localized
    static let stopSession = "Stop Session".localized
    static let pauseSession = "Pause Session".localized
    static let resumeSession = "Resume Session".localized
    static let noActiveSessions = "No active sessions".localized
    static let recentSessions = "Recent Sessions".localized
    static let project = "Project".localized
    static let client = "Client".localized
    static let duration = "Duration".localized
    static let date = "Date".localized
  }

  // MARK: - Pomodoro
  enum Pomodoro {
    static let readyToFocus = "Ready to focus".localized
    static let paused = "Paused".localized
    static let focusTime = "Focus Time".localized
    static let shortBreak = "Short Break".localized
    static let longBreak = "Long Break".localized
    static let startFocusSession = "Start Focus Session".localized
    static let stayFocused = "Stay focused".localized
    static let takeABreak = "Take a break".localized
    static let todaysFocus = "Today's Focus".localized
    static let sessions = "Sessions".localized
    static let focusTimeLabel = "Focus Time".localized
    static let cycles = "Cycles".localized
    static let timerSettings = "Timer Settings".localized
    static let workDuration = "Work Duration".localized
    static let shortBreakLabel = "Short Break".localized
    static let longBreakLabel = "Long Break".localized
    static let sessionsUntilLongBreak = "Sessions Until Long Break".localized
    static let pause = "Pause".localized
    static let resume = "Resume".localized
    static let skip = "Skip".localized
    static let stop = "Stop".localized
    static let startFocus = "Start Focus".localized
  }

  // MARK: - Settings
  enum Settings {
    static let display = "Display".localized
    static let secureMode = "Secure Mode".localized
    static let hideSensitiveAmounts = "Hide sensitive amounts".localized
    static let globalDatabase = "Global Database".localized
    static let useGlobalDb = "Use ~/.ung/ung.db".localized
    static let data = "Data".localized
    static let importDatabase = "Import Database".localized
    static let importFromSQL = "Import from SQL file".localized
    static let exportDatabase = "Export Database".localized
    static let exportToSQL = "Export to SQL/CSV".localized
    static let backupData = "Backup Data".localized
    static let createFullBackup = "Create full backup".localized
    static let database = "Database".localized
    static let location = "Location".localized
    static let resetDatabase = "Reset Database".localized
    static let resetDatabaseWarning =
      "Delete all data (cannot be undone)".localized
    static let about = "About".localized
    static let timeTrackingAndInvoicing = "Time tracking & invoicing".localized
    static let documentation = "Documentation".localized
    static let reportIssue = "Report Issue".localized
    static let checkForUpdates = "Check for Updates".localized
    static let youAreUpToDate = "You're up to date!".localized
    static let newVersionAvailable = "New version available".localized
    static let currentVersion = "Current version".localized
    static let latestVersion = "Latest version".localized
    static let downloadUpdate = "Download Update".localized
    static let checkingForUpdates = "Checking for updates...".localized
    static let updateError = "Could not check for updates".localized
  }

  // MARK: - Invoices
  enum Invoices {
    static let totalInvoices = "Total Invoices".localized
    static let createInvoice = "Create Invoice".localized
    static let searchInvoices = "Search invoices...".localized
    static let noInvoicesYet = "No Invoices Yet".localized
    static let createFirstInvoice = "Create your first invoice from tracked time".localized
    static let all = "All".localized
    static let paid = "Paid".localized
    static let sent = "Sent".localized
    static let pendingStatus = "Pending".localized
    static let overdueStatus = "Overdue".localized
    static let markAsPaid = "Mark as Paid".localized
    static let markAsSent = "Mark as Sent".localized
    static let generatePDF = "Generate PDF".localized
    static let sendEmail = "Send Email".localized
  }

  // MARK: - Expenses
  enum Expenses {
    static let totalExpenses = "Total Expenses".localized
    static let logExpense = "Log Expense".localized
    static let searchExpenses = "Search expenses...".localized
    static let noExpensesYet = "No Expenses Yet".localized
    static let trackFirstExpense = "Track your first business expense".localized
    static let description = "Description".localized
    static let amount = "Amount".localized
    static let category = "Category".localized
    static let categories = "Categories".localized
  }

  // MARK: - Clients
  enum Clients {
    static let totalClients = "Total Clients".localized
    static let addClient = "Add Client".localized
    static let searchClients = "Search clients...".localized
    static let noClientsYet = "No Clients Yet".localized
    static let addFirstClient = "Add your first client to get started".localized
    static let clientName = "Client Name".localized
    static let email = "Email".localized
    static let phone = "Phone".localized
    static let address = "Address".localized
  }

  // MARK: - Contracts
  enum Contracts {
    static let totalContracts = "Total Contracts".localized
    static let newContract = "New Contract".localized
    static let searchContracts = "Search contracts...".localized
    static let noContractsYet = "No Contracts Yet".localized
    static let createFirstContract = "Create your first contract".localized
    static let contractName = "Contract Name".localized
    static let hourlyRate = "Hourly Rate".localized
    static let fixedRate = "Fixed Rate".localized
    static let rateType = "Rate Type".localized
    static let hourly = "Hourly".localized
    static let fixed = "Fixed".localized
    static let daily = "Daily".localized
  }

  // MARK: - Reports
  enum Reports {
    static let analyticsAndInsights = "Analytics and insights".localized
    static let period = "Period".localized
    static let week = "Week".localized
    static let month = "Month".localized
    static let quarter = "Quarter".localized
    static let year = "Year".localized
    static let hoursTracked = "Hours Tracked".localized
    static let activeClients = "Active Clients".localized
    static let averageRate = "Average Rate".localized
    static let exportReport = "Export Report".localized
  }

  // MARK: - Errors
  enum Errors {
    static let somethingWentWrong = "Something went wrong".localized
    static let tryAgain = "Please try again".localized
    static let networkError = "Network error".localized
    static let fileNotFound = "File not found".localized
    static let importFailed = "Import failed".localized
    static let exportFailed = "Export failed".localized
    static let cliNotInstalled = "CLI not installed".localized
    static let databaseNotInitialized = "Database not initialized".localized
  }

  // MARK: - Alerts
  enum Alerts {
    static let areYouSure = "Are you sure?".localized
    static let deleteConfirmation = "This action cannot be undone".localized
    static let importSuccessful = "Import Successful".localized
    static let importSuccessfulMessage =
      "Database imported successfully. Your data has been updated."
      .localized
    static let resetDatabaseTitle = "Reset Database?".localized
    static let resetDatabaseMessage =
      "This will permanently delete all your data including clients, contracts, invoices, expenses, and tracking sessions. This action cannot be undone."
      .localized
    static let templateSaved = "Template Saved".localized
    static let templateSavedMessage = "Your template has been saved successfully.".localized
  }

  // MARK: - Time Formatting
  enum Time {
    static let hours = "hours".localized
    static let minutes = "minutes".localized
    static let seconds = "seconds".localized
    static let hourAbbrev = "h".localized
    static let minuteAbbrev = "m".localized
    static let secondAbbrev = "s".localized
    static let dayStreak = "%d day streak".localized
  }

  // MARK: - Misc
  enum Misc {
    static let of = "of".localized
    static let noResults = "No Results".localized
    static let tryDifferentFilters = "Try different filters".localized
    static let complete = "Complete".localized
    static let incomplete = "Incomplete".localized
    static let locked = "Locked".localized
    static let enabled = "Enabled".localized
    static let disabled = "Disabled".localized
    static let quitApp = "Quit UNG".localized
    static let openApp = "Open App".localized
    static let expand = "Expand".localized
  }
}

// MARK: - App Version Helper
enum AppVersion {
  static var current: String {
    Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0"
  }

  static var build: String {
    Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "1"
  }

  static var fullVersion: String {
    "\(current) (\(build))"
  }
}
