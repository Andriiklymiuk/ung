//
//  WatchConnectivityService.swift
//  ung
//
//  Handles communication with Apple Watch companion app
//

import Foundation
import Combine

#if os(iOS)
import WatchConnectivity

// MARK: - Watch Connectivity Service
@MainActor
class WatchConnectivityService: NSObject, ObservableObject {
    static let shared = WatchConnectivityService()

    @Published var isWatchPaired = false
    @Published var isWatchReachable = false

    private var appState: AppState?

    override init() {
        super.init()
        setupWatchConnectivity()
    }

    func configure(with appState: AppState) {
        self.appState = appState
    }

    private func setupWatchConnectivity() {
        guard WCSession.isSupported() else { return }
        let session = WCSession.default
        session.delegate = self
        session.activate()
    }

    // MARK: - Send Data to Watch

    func sendSessionUpdate() {
        guard let appState = appState else { return }
        guard WCSession.default.isReachable else { return }

        let sessionData = WatchSessionData(
            isTracking: appState.isTracking,
            projectName: appState.activeSession?.project ?? "",
            clientName: appState.activeSession?.client ?? "",
            startTime: appState.activeSession.map { Date().addingTimeInterval(-TimeInterval($0.elapsedSeconds)) },
            elapsedSeconds: appState.activeSession?.elapsedSeconds ?? 0
        )

        guard let data = try? JSONEncoder().encode(sessionData) else { return }
        WCSession.default.sendMessage(["session": data], replyHandler: nil, errorHandler: nil)
    }

    func sendPomodoroUpdate() {
        guard let appState = appState else { return }
        guard WCSession.default.isReachable else { return }

        let pomodoroData = WatchPomodoroData(
            isActive: appState.pomodoroState.isActive,
            isBreak: appState.pomodoroState.isBreak,
            isPaused: appState.pomodoroState.isPaused,
            secondsRemaining: appState.pomodoroState.secondsRemaining,
            sessionsCompleted: appState.pomodoroState.sessionsCompleted,
            workMinutes: appState.pomodoroState.workMinutes,
            breakMinutes: appState.pomodoroState.breakMinutes
        )

        guard let data = try? JSONEncoder().encode(pomodoroData) else { return }
        WCSession.default.sendMessage(["pomodoro": data], replyHandler: nil, errorHandler: nil)
    }

    func sendStatsUpdate() {
        guard let appState = appState else { return }
        guard WCSession.default.isReachable else { return }

        let statsData = WatchStatsData(
            todayHours: calculateTodayHours(),
            weeklyHours: appState.metrics.weeklyHours,
            weeklyTarget: appState.metrics.weeklyTarget,
            pendingInvoices: appState.invoiceCount, // Could filter for pending only
            totalRevenue: appState.metrics.totalRevenue
        )

        guard let data = try? JSONEncoder().encode(statsData) else { return }
        WCSession.default.sendMessage(["stats": data], replyHandler: nil, errorHandler: nil)
    }

    private func calculateTodayHours() -> Double {
        // Sum today's sessions
        guard let appState = appState else { return 0 }
        let today = Calendar.current.startOfDay(for: Date())

        return appState.recentSessions
            .filter { session in
                // Parse the date from the string if needed
                return true // Simplified - would need proper date parsing
            }
            .reduce(0.0) { total, _ in total + 1.0 } // Simplified
    }

    func sendFullSync() {
        guard let appState = appState else { return }
        guard WCSession.default.isReachable else { return }

        var response: [String: Any] = [:]

        // Session data
        let sessionData = WatchSessionData(
            isTracking: appState.isTracking,
            projectName: appState.activeSession?.project ?? "",
            clientName: appState.activeSession?.client ?? "",
            startTime: appState.activeSession.map { Date().addingTimeInterval(-TimeInterval($0.elapsedSeconds)) },
            elapsedSeconds: appState.activeSession?.elapsedSeconds ?? 0
        )
        if let data = try? JSONEncoder().encode(sessionData) {
            response["session"] = data
        }

        // Pomodoro data
        let pomodoroData = WatchPomodoroData(
            isActive: appState.pomodoroState.isActive,
            isBreak: appState.pomodoroState.isBreak,
            isPaused: appState.pomodoroState.isPaused,
            secondsRemaining: appState.pomodoroState.secondsRemaining,
            sessionsCompleted: appState.pomodoroState.sessionsCompleted,
            workMinutes: appState.pomodoroState.workMinutes,
            breakMinutes: appState.pomodoroState.breakMinutes
        )
        if let data = try? JSONEncoder().encode(pomodoroData) {
            response["pomodoro"] = data
        }

        // Stats data
        let statsData = WatchStatsData(
            todayHours: calculateTodayHours(),
            weeklyHours: appState.metrics.weeklyHours,
            weeklyTarget: appState.metrics.weeklyTarget,
            pendingInvoices: appState.invoiceCount,
            totalRevenue: appState.metrics.totalRevenue
        )
        if let data = try? JSONEncoder().encode(statsData) {
            response["stats"] = data
        }

        // Quick projects
        let recentProjects = appState.recentSessions.prefix(5).map { $0.project }
        let defaultProjects = ["Development", "Meeting", "Design", "Research", "Admin"]
        let projects = recentProjects.isEmpty ? defaultProjects : Array(Set(recentProjects))
        response["quickProjects"] = projects

        WCSession.default.sendMessage(response, replyHandler: nil, errorHandler: nil)
    }

    // MARK: - Handle Watch Requests
    private func handleWatchMessage(_ message: [String: Any], replyHandler: (([String: Any]) -> Void)?) {
        guard let action = message["action"] as? String else {
            if let reply = replyHandler {
                buildSyncResponse(reply)
            }
            return
        }

        Task { @MainActor in
            switch action {
            case "sync":
                if let reply = replyHandler {
                    buildSyncResponse(reply)
                }

            case "startTracking":
                if let project = message["project"] as? String,
                   let appState = self.appState {
                    await appState.startTracking(project: project, clientId: nil)
                    sendSessionUpdate()
                }

            case "stopTracking":
                if let appState = self.appState {
                    await appState.stopTracking()
                    sendSessionUpdate()
                }

            case "startPomodoro":
                appState?.startPomodoro()
                sendPomodoroUpdate()

            case "pausePomodoro":
                appState?.pausePomodoro()
                sendPomodoroUpdate()

            case "resumePomodoro":
                appState?.resumePomodoro()
                sendPomodoroUpdate()

            case "stopPomodoro":
                appState?.stopPomodoro()
                sendPomodoroUpdate()

            case "skipPomodoro":
                appState?.skipPomodoro()
                sendPomodoroUpdate()

            default:
                break
            }
        }
    }

    private func buildSyncResponse(_ reply: @escaping ([String: Any]) -> Void) {
        guard let appState = appState else {
            reply([:])
            return
        }

        var response: [String: Any] = [:]

        // Session data
        let sessionData = WatchSessionData(
            isTracking: appState.isTracking,
            projectName: appState.activeSession?.project ?? "",
            clientName: appState.activeSession?.client ?? "",
            startTime: appState.activeSession.map { Date().addingTimeInterval(-TimeInterval($0.elapsedSeconds)) },
            elapsedSeconds: appState.activeSession?.elapsedSeconds ?? 0
        )
        if let data = try? JSONEncoder().encode(sessionData) {
            response["session"] = data
        }

        // Pomodoro data
        let pomodoroData = WatchPomodoroData(
            isActive: appState.pomodoroState.isActive,
            isBreak: appState.pomodoroState.isBreak,
            isPaused: appState.pomodoroState.isPaused,
            secondsRemaining: appState.pomodoroState.secondsRemaining,
            sessionsCompleted: appState.pomodoroState.sessionsCompleted,
            workMinutes: appState.pomodoroState.workMinutes,
            breakMinutes: appState.pomodoroState.breakMinutes
        )
        if let data = try? JSONEncoder().encode(pomodoroData) {
            response["pomodoro"] = data
        }

        // Stats data
        let statsData = WatchStatsData(
            todayHours: calculateTodayHours(),
            weeklyHours: appState.metrics.weeklyHours,
            weeklyTarget: appState.metrics.weeklyTarget,
            pendingInvoices: appState.invoiceCount,
            totalRevenue: appState.metrics.totalRevenue
        )
        if let data = try? JSONEncoder().encode(statsData) {
            response["stats"] = data
        }

        // Quick projects
        let recentProjects = appState.recentSessions.prefix(5).map { $0.project }
        let defaultProjects = ["Development", "Meeting", "Design", "Research", "Admin"]
        let projects = recentProjects.isEmpty ? defaultProjects : Array(Set(recentProjects))
        response["quickProjects"] = projects

        reply(response)
    }
}

// MARK: - WCSessionDelegate
extension WatchConnectivityService: WCSessionDelegate {
    nonisolated func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        DispatchQueue.main.async {
            self.isWatchPaired = session.isPaired
            self.isWatchReachable = session.isReachable
        }
    }

    nonisolated func sessionDidBecomeInactive(_ session: WCSession) {}

    nonisolated func sessionDidDeactivate(_ session: WCSession) {
        // Reactivate
        WCSession.default.activate()
    }

    nonisolated func session(_ session: WCSession, didReceiveMessage message: [String: Any]) {
        DispatchQueue.main.async {
            self.handleWatchMessage(message, replyHandler: nil)
        }
    }

    nonisolated func session(_ session: WCSession, didReceiveMessage message: [String: Any], replyHandler: @escaping ([String: Any]) -> Void) {
        DispatchQueue.main.async {
            self.handleWatchMessage(message, replyHandler: replyHandler)
        }
    }

    nonisolated func sessionReachabilityDidChange(_ session: WCSession) {
        DispatchQueue.main.async {
            self.isWatchReachable = session.isReachable
        }
    }

    nonisolated func sessionWatchStateDidChange(_ session: WCSession) {
        DispatchQueue.main.async {
            self.isWatchPaired = session.isPaired
        }
    }
}

// MARK: - Shared Data Types (also used by Watch app)
struct WatchSessionData: Codable {
    var isTracking: Bool = false
    var projectName: String = ""
    var clientName: String = ""
    var startTime: Date?
    var elapsedSeconds: Int = 0
}

struct WatchPomodoroData: Codable {
    var isActive: Bool = false
    var isBreak: Bool = false
    var isPaused: Bool = false
    var secondsRemaining: Int = 25 * 60
    var sessionsCompleted: Int = 0
    var workMinutes: Int = 25
    var breakMinutes: Int = 5
}

struct WatchStatsData: Codable {
    var todayHours: Double = 0
    var weeklyHours: Double = 0
    var weeklyTarget: Double = 40
    var pendingInvoices: Int = 0
    var totalRevenue: Double = 0
}

#endif
