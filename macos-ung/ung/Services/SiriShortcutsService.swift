//
//  SiriShortcutsService.swift
//  ung
//
//  Siri Shortcuts integration with multi-language support (English, Russian)
//

import Foundation
import Intents
import AppIntents

// MARK: - App Shortcuts Provider
@available(iOS 16.0, macOS 13.0, *)
struct UNGShortcuts: AppShortcutsProvider {
    static var appShortcuts: [AppShortcut] {
        // English shortcuts
        AppShortcut(
            intent: StartTrackingIntent(),
            phrases: [
                "Start tracking in \(.applicationName)",
                "Start timer in \(.applicationName)",
                "Begin work in \(.applicationName)",
                "Track time with \(.applicationName)"
            ],
            shortTitle: "Start Tracking",
            systemImageName: "play.circle.fill"
        )

        AppShortcut(
            intent: StopTrackingIntent(),
            phrases: [
                "Stop tracking in \(.applicationName)",
                "Stop timer in \(.applicationName)",
                "End work in \(.applicationName)",
                "Finish tracking with \(.applicationName)"
            ],
            shortTitle: "Stop Tracking",
            systemImageName: "stop.circle.fill"
        )

        AppShortcut(
            intent: GetStatusIntent(),
            phrases: [
                "What am I working on in \(.applicationName)?",
                "Show status in \(.applicationName)",
                "Check timer in \(.applicationName)",
                "How long have I been working in \(.applicationName)?"
            ],
            shortTitle: "Get Status",
            systemImageName: "clock.fill"
        )

        AppShortcut(
            intent: GetTodayHoursIntent(),
            phrases: [
                "How many hours today in \(.applicationName)?",
                "Show today's hours in \(.applicationName)",
                "Time worked today in \(.applicationName)"
            ],
            shortTitle: "Today's Hours",
            systemImageName: "sun.max.fill"
        )

        AppShortcut(
            intent: StartFocusIntent(),
            phrases: [
                "Start focus session in \(.applicationName)",
                "Start pomodoro in \(.applicationName)",
                "Begin focus time with \(.applicationName)"
            ],
            shortTitle: "Start Focus",
            systemImageName: "brain.head.profile"
        )

        AppShortcut(
            intent: GetWeeklyReportIntent(),
            phrases: [
                "Weekly report in \(.applicationName)",
                "How was my week in \(.applicationName)?",
                "This week's summary in \(.applicationName)"
            ],
            shortTitle: "Weekly Report",
            systemImageName: "chart.bar.fill"
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
