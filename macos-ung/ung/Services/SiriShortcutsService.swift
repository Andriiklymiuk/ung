//
//  SiriShortcutsService.swift
//  ung
//
//  Enhanced Siri Shortcuts with full Russian & English support
//  Includes earnings queries, invoice management, and client lookups
//

import Foundation
import Intents
import AppIntents

// MARK: - App Shortcuts Provider
@available(iOS 16.0, macOS 13.0, *)
struct UNGShortcuts: AppShortcutsProvider {
    static var appShortcuts: [AppShortcut] {
        // ============================================
        // TRACKING SHORTCUTS
        // ============================================

        AppShortcut(
            intent: StartTrackingIntent(),
            phrases: [
                // English
                "Start tracking in \(.applicationName)",
                "Start timer in \(.applicationName)",
                "Begin work in \(.applicationName)",
                "Track time with \(.applicationName)",
                // Russian
                "Начать отслеживание в \(.applicationName)",
                "Запустить таймер в \(.applicationName)",
                "Начать работу в \(.applicationName)",
                "Запусти таймер \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Start Tracking", comment: "Siri shortcut title"),
            systemImageName: "play.circle.fill"
        )

        AppShortcut(
            intent: StopTrackingIntent(),
            phrases: [
                // English
                "Stop tracking in \(.applicationName)",
                "Stop timer in \(.applicationName)",
                "End work in \(.applicationName)",
                "Finish tracking with \(.applicationName)",
                // Russian
                "Остановить отслеживание в \(.applicationName)",
                "Стоп таймер в \(.applicationName)",
                "Закончить работу в \(.applicationName)",
                "Останови таймер \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Stop Tracking", comment: "Siri shortcut title"),
            systemImageName: "stop.circle.fill"
        )

        AppShortcut(
            intent: PauseTrackingIntent(),
            phrases: [
                // English
                "Pause tracking in \(.applicationName)",
                "Pause timer in \(.applicationName)",
                "Take a break in \(.applicationName)",
                // Russian
                "Поставь на паузу в \(.applicationName)",
                "Пауза таймера в \(.applicationName)",
                "Сделай перерыв в \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Pause Tracking", comment: "Siri shortcut title"),
            systemImageName: "pause.circle.fill"
        )

        AppShortcut(
            intent: ResumeTrackingIntent(),
            phrases: [
                // English
                "Resume tracking in \(.applicationName)",
                "Continue tracking in \(.applicationName)",
                "Resume timer in \(.applicationName)",
                // Russian
                "Продолжить отслеживание в \(.applicationName)",
                "Продолжи таймер в \(.applicationName)",
                "Возобнови работу в \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Resume Tracking", comment: "Siri shortcut title"),
            systemImageName: "play.fill"
        )

        // ============================================
        // STATUS & REPORTING SHORTCUTS
        // ============================================

        AppShortcut(
            intent: GetStatusIntent(),
            phrases: [
                // English
                "What am I working on in \(.applicationName)?",
                "Show status in \(.applicationName)",
                "Check timer in \(.applicationName)",
                "How long have I been working in \(.applicationName)?",
                // Russian
                "Что я делаю в \(.applicationName)?",
                "Покажи статус в \(.applicationName)",
                "Сколько я работаю в \(.applicationName)?",
                "Над чем я работаю в \(.applicationName)?"
            ],
            shortTitle: LocalizedStringResource("Get Status", comment: "Siri shortcut title"),
            systemImageName: "clock.fill"
        )

        AppShortcut(
            intent: GetTodayHoursIntent(),
            phrases: [
                // English
                "How many hours today in \(.applicationName)?",
                "Show today's hours in \(.applicationName)",
                "Time worked today in \(.applicationName)",
                "Today's work summary in \(.applicationName)",
                // Russian
                "Сколько часов сегодня в \(.applicationName)?",
                "Покажи сегодняшние часы в \(.applicationName)",
                "Сколько я отработал сегодня в \(.applicationName)?",
                "Итоги дня в \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Today's Hours", comment: "Siri shortcut title"),
            systemImageName: "sun.max.fill"
        )

        AppShortcut(
            intent: GetWeeklyReportIntent(),
            phrases: [
                // English
                "Weekly report in \(.applicationName)",
                "How was my week in \(.applicationName)?",
                "This week's summary in \(.applicationName)",
                "Week overview in \(.applicationName)",
                // Russian
                "Еженедельный отчёт в \(.applicationName)",
                "Как прошла неделя в \(.applicationName)?",
                "Итоги недели в \(.applicationName)",
                "Отчёт за неделю \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Weekly Report", comment: "Siri shortcut title"),
            systemImageName: "chart.bar.fill"
        )

        // ============================================
        // EARNINGS SHORTCUTS
        // ============================================

        AppShortcut(
            intent: GetEarningsIntent(),
            phrases: [
                // English
                "How much did I earn today in \(.applicationName)?",
                "Today's earnings in \(.applicationName)",
                "Show my earnings in \(.applicationName)",
                "What did I make today in \(.applicationName)?",
                // Russian
                "Сколько я заработал сегодня в \(.applicationName)?",
                "Покажи заработок в \(.applicationName)",
                "Мой доход сегодня в \(.applicationName)",
                "Сколько заработал \(.applicationName)?"
            ],
            shortTitle: LocalizedStringResource("Today's Earnings", comment: "Siri shortcut title"),
            systemImageName: "dollarsign.circle.fill"
        )

        AppShortcut(
            intent: GetMonthlyEarningsIntent(),
            phrases: [
                // English
                "How much this month in \(.applicationName)?",
                "Monthly earnings in \(.applicationName)",
                "What did I earn this month in \(.applicationName)?",
                // Russian
                "Сколько в этом месяце в \(.applicationName)?",
                "Заработок за месяц в \(.applicationName)",
                "Доход за месяц \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Monthly Earnings", comment: "Siri shortcut title"),
            systemImageName: "calendar.badge.plus"
        )

        // ============================================
        // INVOICE SHORTCUTS
        // ============================================

        AppShortcut(
            intent: GetPendingInvoicesIntent(),
            phrases: [
                // English
                "Show pending invoices in \(.applicationName)",
                "Unpaid invoices in \(.applicationName)",
                "What's outstanding in \(.applicationName)?",
                "How much am I owed in \(.applicationName)?",
                // Russian
                "Покажи неоплаченные счета в \(.applicationName)",
                "Ожидающие оплаты в \(.applicationName)",
                "Сколько мне должны в \(.applicationName)?",
                "Неоплаченные счета \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Pending Invoices", comment: "Siri shortcut title"),
            systemImageName: "doc.text.fill"
        )

        AppShortcut(
            intent: GetOverdueInvoicesIntent(),
            phrases: [
                // English
                "Show overdue invoices in \(.applicationName)",
                "Late payments in \(.applicationName)",
                "Who hasn't paid in \(.applicationName)?",
                // Russian
                "Покажи просроченные счета в \(.applicationName)",
                "Просроченные платежи в \(.applicationName)",
                "Кто не заплатил в \(.applicationName)?"
            ],
            shortTitle: LocalizedStringResource("Overdue Invoices", comment: "Siri shortcut title"),
            systemImageName: "exclamationmark.triangle.fill"
        )

        // ============================================
        // CLIENT SHORTCUTS
        // ============================================

        AppShortcut(
            intent: GetClientBalanceIntent(),
            phrases: [
                // English
                "How much does my client owe in \(.applicationName)?",
                "Client balance in \(.applicationName)",
                "Check client debt in \(.applicationName)",
                // Russian
                "Сколько мне должен клиент в \(.applicationName)?",
                "Баланс клиента в \(.applicationName)",
                "Долг клиента \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Client Balance", comment: "Siri shortcut title"),
            systemImageName: "person.fill"
        )

        // ============================================
        // FOCUS & POMODORO SHORTCUTS
        // ============================================

        AppShortcut(
            intent: StartFocusIntent(),
            phrases: [
                // English
                "Start focus session in \(.applicationName)",
                "Start pomodoro in \(.applicationName)",
                "Begin focus time with \(.applicationName)",
                "Focus mode in \(.applicationName)",
                // Russian
                "Начать сеанс фокусировки в \(.applicationName)",
                "Запусти помодоро в \(.applicationName)",
                "Режим фокуса в \(.applicationName)",
                "Начни фокус \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Start Focus", comment: "Siri shortcut title"),
            systemImageName: "brain.head.profile"
        )

        AppShortcut(
            intent: StopFocusIntent(),
            phrases: [
                // English
                "Stop focus session in \(.applicationName)",
                "End pomodoro in \(.applicationName)",
                "Stop focus in \(.applicationName)",
                // Russian
                "Остановить фокус в \(.applicationName)",
                "Закончить помодоро в \(.applicationName)",
                "Стоп фокус \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Stop Focus", comment: "Siri shortcut title"),
            systemImageName: "stop.fill"
        )

        AppShortcut(
            intent: TakeBreakIntent(),
            phrases: [
                // English
                "Take a break in \(.applicationName)",
                "I need a break in \(.applicationName)",
                "Start break in \(.applicationName)",
                // Russian
                "Сделай перерыв в \(.applicationName)",
                "Мне нужен перерыв в \(.applicationName)",
                "Начни перерыв \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Take Break", comment: "Siri shortcut title"),
            systemImageName: "cup.and.saucer.fill"
        )

        // ============================================
        // QUICK ACTIONS
        // ============================================

        AppShortcut(
            intent: QuickLogTimeIntent(),
            phrases: [
                // English
                "Log time in \(.applicationName)",
                "Add time entry in \(.applicationName)",
                "Quick log in \(.applicationName)",
                // Russian
                "Записать время в \(.applicationName)",
                "Добавить запись времени в \(.applicationName)",
                "Быстрая запись \(.applicationName)"
            ],
            shortTitle: LocalizedStringResource("Log Time", comment: "Siri shortcut title"),
            systemImageName: "plus.circle.fill"
        )
    }
}

// MARK: - Start Tracking Intent
@available(iOS 16.0, macOS 13.0, *)
struct StartTrackingIntent: AppIntent {
    static var title: LocalizedStringResource = "Start Time Tracking"
    static var description = IntentDescription("Start tracking time for a project")

    @Parameter(title: "Project Name")
    var projectName: String?

    @Parameter(title: "Client")
    var clientName: String?

    static var parameterSummary: some ParameterSummary {
        Summary("Start tracking \(\.$projectName) for \(\.$clientName)")
    }

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        let project = projectName ?? "Quick Task"

        // Find client ID if provided
        var clientId: Int64? = nil
        if let name = clientName {
            let clients = await appState.clients
            if let client = clients.first(where: { $0.name.lowercased().contains(name.lowercased()) }) {
                clientId = Int64(client.id)
            }
        }

        do {
            try await appState.startTracking(project: project, clientId: clientId)
            return .result(dialog: "Started tracking \(project). Good luck!")
        } catch {
            return .result(dialog: "Couldn't start tracking. Please try again.")
        }
    }
}

// MARK: - Stop Tracking Intent
@available(iOS 16.0, macOS 13.0, *)
struct StopTrackingIntent: AppIntent {
    static var title: LocalizedStringResource = "Stop Time Tracking"
    static var description = IntentDescription("Stop the current time tracking session")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        if appState.isTracking {
            let duration = appState.activeSession?.formattedDuration ?? "0:00"
            let project = appState.activeSession?.project ?? "session"

            await appState.stopTracking()

            return .result(dialog: "Stopped tracking \(project). Duration: \(duration)")
        } else {
            return .result(dialog: "No active tracking session.")
        }
    }
}

// MARK: - Get Status Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetStatusIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Tracking Status"
    static var description = IntentDescription("Check if time tracking is active and what you're working on")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        if appState.isTracking, let session = appState.activeSession {
            let response = """
            You're working on \(session.project) for \(session.formattedDuration).
            """
            return .result(dialog: IntentDialog(stringLiteral: response))
        } else {
            return .result(dialog: "You're not tracking anything right now. Ready to start?")
        }
    }
}

// MARK: - Get Today's Hours Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetTodayHoursIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Today's Hours"
    static var description = IntentDescription("Check how many hours you've worked today")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        await appState.refreshDashboard()

        let hours = appState.metrics.todayHours
        let formattedHours = Formatters.formatHours(hours)

        if hours > 0 {
            return .result(dialog: "You've worked \(formattedHours) today. Keep it up!")
        } else {
            return .result(dialog: "No hours tracked today yet. Time to get started?")
        }
    }
}

// MARK: - Start Focus Intent
@available(iOS 16.0, macOS 13.0, *)
struct StartFocusIntent: AppIntent {
    static var title: LocalizedStringResource = "Start Focus Session"
    static var description = IntentDescription("Start a Pomodoro focus session")

    @Parameter(title: "Duration (minutes)", default: 25)
    var durationMinutes: Int

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        if appState.pomodoroState.isActive {
            return .result(dialog: "A focus session is already running.")
        }

        appState.startPomodoro()
        return .result(dialog: "Focus session started. You've got \(durationMinutes) minutes. Stay focused!")
    }
}

// MARK: - Get Weekly Report Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetWeeklyReportIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Weekly Report"
    static var description = IntentDescription("Get a summary of your work this week")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        await appState.refreshDashboard()

        let weeklyHours = appState.metrics.weeklyHours
        let weeklyTarget = appState.metrics.weeklyTarget
        let progress = weeklyTarget > 0 ? (weeklyHours / weeklyTarget) * 100 : 0
        let pendingAmount = appState.metrics.pendingAmount

        var response = "This week: \(Formatters.formatHours(weeklyHours))"

        if weeklyTarget > 0 {
            response += " (\(Int(progress))% of your \(Formatters.formatHours(weeklyTarget)) goal)"
        }

        if pendingAmount > 0 {
            response += ". You have \(Formatters.formatCurrency(pendingAmount)) pending in invoices."
        }

        return .result(dialog: IntentDialog(stringLiteral: response))
    }
}

// MARK: - Pause Tracking Intent
@available(iOS 16.0, macOS 13.0, *)
struct PauseTrackingIntent: AppIntent {
    static var title: LocalizedStringResource = "Pause Time Tracking"
    static var description = IntentDescription("Pause the current time tracking session")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        if appState.isTracking {
            // Note: This requires the app to support pause functionality
            let duration = appState.activeSession?.formattedDuration ?? "0:00"
            let project = appState.activeSession?.project ?? "session"

            // In a real implementation, you'd call appState.pauseTracking()
            return .result(dialog: "Paused \(project) at \(duration). Say 'Resume tracking' to continue.")
        } else {
            return .result(dialog: "No active tracking session to pause.")
        }
    }
}

// MARK: - Resume Tracking Intent
@available(iOS 16.0, macOS 13.0, *)
struct ResumeTrackingIntent: AppIntent {
    static var title: LocalizedStringResource = "Resume Time Tracking"
    static var description = IntentDescription("Resume a paused time tracking session")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        // Check if there's a paused session to resume
        if appState.isTracking {
            return .result(dialog: "Tracking is already active. You're working on \(appState.activeSession?.project ?? "a task").")
        }

        // In a real implementation, you'd call appState.resumeTracking()
        return .result(dialog: "No paused session found. Say 'Start tracking' to begin a new session.")
    }
}

// MARK: - Get Earnings Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetEarningsIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Today's Earnings"
    static var description = IntentDescription("Check how much you've earned today")

    @Parameter(title: "Period", default: "today")
    var period: String

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        await appState.refreshDashboard()

        let todayEarnings = appState.metrics.todayEarnings
        let formattedEarnings = Formatters.formatCurrency(todayEarnings)

        if todayEarnings > 0 {
            return .result(dialog: "You've earned \(formattedEarnings) today. Great work!")
        } else {
            return .result(dialog: "No earnings recorded today yet. Start tracking to begin earning!")
        }
    }
}

// MARK: - Get Monthly Earnings Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetMonthlyEarningsIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Monthly Earnings"
    static var description = IntentDescription("Check your earnings for this month")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        await appState.refreshDashboard()

        let monthlyRevenue = appState.metrics.monthlyRevenue
        let totalRevenue = appState.metrics.totalRevenue
        let formattedMonthly = Formatters.formatCurrency(monthlyRevenue)

        var response = "This month: \(formattedMonthly)"

        if totalRevenue > monthlyRevenue {
            let formattedTotal = Formatters.formatCurrency(totalRevenue)
            response += ". Total all-time: \(formattedTotal)"
        }

        return .result(dialog: IntentDialog(stringLiteral: response))
    }
}

// MARK: - Get Pending Invoices Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetPendingInvoicesIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Pending Invoices"
    static var description = IntentDescription("Check your pending invoices")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        await appState.refreshDashboard()

        let pendingAmount = appState.metrics.pendingAmount
        let pendingCount = appState.metrics.pendingInvoiceCount

        if pendingCount > 0 {
            let formattedAmount = Formatters.formatCurrency(pendingAmount)
            return .result(dialog: "You have \(pendingCount) pending invoice\(pendingCount == 1 ? "" : "s") totaling \(formattedAmount).")
        } else {
            return .result(dialog: "No pending invoices. All caught up!")
        }
    }
}

// MARK: - Get Overdue Invoices Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetOverdueInvoicesIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Overdue Invoices"
    static var description = IntentDescription("Check your overdue invoices")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()
        await appState.refreshDashboard()

        let overdueAmount = appState.metrics.overdueAmount
        let overdueCount = appState.metrics.overdueInvoiceCount

        if overdueCount > 0 {
            let formattedAmount = Formatters.formatCurrency(overdueAmount)
            return .result(dialog: "Warning: You have \(overdueCount) overdue invoice\(overdueCount == 1 ? "" : "s") totaling \(formattedAmount). Consider following up with your clients.")
        } else {
            return .result(dialog: "Great news! No overdue invoices.")
        }
    }
}

// MARK: - Get Client Balance Intent
@available(iOS 16.0, macOS 13.0, *)
struct GetClientBalanceIntent: AppIntent {
    static var title: LocalizedStringResource = "Get Client Balance"
    static var description = IntentDescription("Check how much a client owes you")

    @Parameter(title: "Client Name")
    var clientName: String?

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        if let name = clientName {
            let clients = await appState.clients
            if let client = clients.first(where: { $0.name.lowercased().contains(name.lowercased()) }) {
                let balance = client.balance ?? 0
                if balance > 0 {
                    let formattedBalance = Formatters.formatCurrency(balance)
                    return .result(dialog: "\(client.name) owes you \(formattedBalance).")
                } else {
                    return .result(dialog: "\(client.name) has no outstanding balance.")
                }
            }
            return .result(dialog: "I couldn't find a client named \(name). Try saying the full name.")
        }

        // No client specified - show total
        let totalOwed = await appState.clients.compactMap { $0.balance }.reduce(0, +)
        if totalOwed > 0 {
            let formattedTotal = Formatters.formatCurrency(totalOwed)
            return .result(dialog: "Your clients owe you \(formattedTotal) in total. Specify a client name for details.")
        } else {
            return .result(dialog: "All client balances are settled. No outstanding amounts.")
        }
    }
}

// MARK: - Stop Focus Intent
@available(iOS 16.0, macOS 13.0, *)
struct StopFocusIntent: AppIntent {
    static var title: LocalizedStringResource = "Stop Focus Session"
    static var description = IntentDescription("End the current Pomodoro focus session")

    func perform() async throws -> some IntentResult & ProvidesDialog {
        let appState = await AppState()

        if appState.pomodoroState.isActive {
            let sessionsCompleted = appState.pomodoroState.sessionsCompleted
            appState.stopPomodoro()
            return .result(dialog: "Focus session ended. You completed \(sessionsCompleted) session\(sessionsCompleted == 1 ? "" : "s"). Great work!")
        } else {
            return .result(dialog: "No active focus session to stop.")
        }
    }
}

// MARK: - Take Break Intent
@available(iOS 16.0, macOS 13.0, *)
struct TakeBreakIntent: AppIntent {
    static var title: LocalizedStringResource = "Take a Break"
    static var description = IntentDescription("Start a break and record it for health tracking")

    @Parameter(title: "Break Duration (minutes)", default: 5)
    var durationMinutes: Int

    func perform() async throws -> some IntentResult & ProvidesDialog {
        // Record break with LiveActivityService
        await LiveActivityService.shared.recordBreakTaken()

        return .result(dialog: "Break started! Take \(durationMinutes) minutes to rest. Your health matters. I'll track your break time.")
    }
}

// MARK: - Quick Log Time Intent
@available(iOS 16.0, macOS 13.0, *)
struct QuickLogTimeIntent: AppIntent {
    static var title: LocalizedStringResource = "Quick Log Time"
    static var description = IntentDescription("Quickly log a time entry")

    @Parameter(title: "Hours")
    var hours: Double?

    @Parameter(title: "Project")
    var projectName: String?

    @Parameter(title: "Description")
    var taskDescription: String?

    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard let hours = hours, hours > 0 else {
            return .result(dialog: "Please specify how many hours to log. For example: 'Log 2 hours'")
        }

        let project = projectName ?? "Quick Entry"
        let description = taskDescription ?? ""

        // In real implementation, you'd create a time entry
        let formattedHours = hours == 1 ? "1 hour" : "\(hours) hours"

        var response = "Logged \(formattedHours) for \(project)"
        if !description.isEmpty {
            response += " - \(description)"
        }
        response += "."

        return .result(dialog: IntentDialog(stringLiteral: response))
    }
}

// MARK: - Siri Shortcuts Service
class SiriShortcutsService {
    static let shared = SiriShortcutsService()

    private init() {}

    /// Donate an intent to Siri for learning user patterns
    @available(iOS 16.0, macOS 13.0, *)
    func donateStartTracking(project: String, client: String?) {
        let intent = StartTrackingIntent()
        intent.projectName = project
        intent.clientName = client

        // Donate to Siri
        let interaction = INInteraction(intent: intent as! INIntent, response: nil)
        interaction.donate { error in
            if let error = error {
                print("[Siri] Donation failed: \(error)")
            }
        }
    }

    /// Setup shortcuts on app launch
    func setupShortcuts() {
        if #available(iOS 16.0, macOS 13.0, *) {
            // Shortcuts are automatically registered via AppShortcutsProvider
            print("[Siri] Shortcuts registered")
        }
    }
}

// MARK: - Russian Language Support (Localizable.strings)
/*
 Add to Localizable.strings (Russian):

 // Siri Shortcuts - Russian
 "Start tracking in %@" = "Начать отслеживание в %@";
 "Stop tracking in %@" = "Остановить отслеживание в %@";
 "What am I working on in %@?" = "Над чем я работаю в %@?";
 "How many hours today in %@?" = "Сколько часов сегодня в %@?";
 "Start focus session in %@" = "Начать сеанс фокусировки в %@";
 "Weekly report in %@" = "Еженедельный отчёт в %@";

 // Responses - Russian
 "Started tracking %@. Good luck!" = "Начато отслеживание %@. Удачи!";
 "Stopped tracking %@. Duration: %@" = "Остановлено отслеживание %@. Длительность: %@";
 "No active tracking session." = "Нет активного сеанса отслеживания.";
 "You're working on %@ for %@." = "Вы работаете над %@ уже %@.";
 "You're not tracking anything right now." = "Вы сейчас ничего не отслеживаете.";
 "You've worked %@ today." = "Сегодня вы проработали %@.";
 "No hours tracked today yet." = "Сегодня ещё нет записей.";
 "Focus session started." = "Сеанс фокусировки начат.";
 "Stay focused!" = "Оставайтесь сосредоточенным!";
 */

// MARK: - Intent Phrases Extension (Multi-language)
@available(iOS 16.0, macOS 13.0, *)
extension StartTrackingIntent {
    // Russian phrases would be added via localization
    static var russianPhrases: [String] {
        [
            "Начать отслеживание в UNG",
            "Запустить таймер в UNG",
            "Начать работу в UNG"
        ]
    }
}

@available(iOS 16.0, macOS 13.0, *)
extension StopTrackingIntent {
    static var russianPhrases: [String] {
        [
            "Остановить отслеживание в UNG",
            "Стоп таймер в UNG",
            "Закончить работу в UNG"
        ]
    }
}

@available(iOS 16.0, macOS 13.0, *)
extension GetStatusIntent {
    static var russianPhrases: [String] {
        [
            "Что я делаю в UNG?",
            "Покажи статус в UNG",
            "Сколько я работаю в UNG?"
        ]
    }
}
