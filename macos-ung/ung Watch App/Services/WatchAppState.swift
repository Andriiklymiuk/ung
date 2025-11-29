//
//  WatchAppState.swift
//  ung Watch App
//
//  State management and WatchConnectivity for Watch app
//

import Foundation
import WatchConnectivity
import Combine

// MARK: - Watch Session Data
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

// MARK: - Watch App State
@MainActor
class WatchAppState: NSObject, ObservableObject {
    // Tracking State
    @Published var sessionData = WatchSessionData()
    @Published var pomodoroData = WatchPomodoroData()
    @Published var statsData = WatchStatsData()

    // Connection Status
    @Published var isConnected = false
    @Published var lastSyncTime: Date?

    // Timers
    private var trackingTimer: Timer?
    private var pomodoroTimer: Timer?

    // Quick Projects (cached for offline use)
    @Published var quickProjects: [String] = ["Development", "Meeting", "Design", "Research", "Admin"]

    override init() {
        super.init()
        setupWatchConnectivity()
    }

    // MARK: - Watch Connectivity
    private func setupWatchConnectivity() {
        guard WCSession.isSupported() else { return }
        let session = WCSession.default
        session.delegate = self
        session.activate()
    }

    func requestSync() {
        guard WCSession.default.isReachable else { return }
        WCSession.default.sendMessage(["action": "sync"], replyHandler: { [weak self] response in
            DispatchQueue.main.async {
                self?.handleSyncResponse(response)
            }
        }, errorHandler: { error in
            print("Watch sync error: \(error)")
        })
    }

    private func handleSyncResponse(_ response: [String: Any]) {
        if let sessionJSON = response["session"] as? Data,
           let session = try? JSONDecoder().decode(WatchSessionData.self, from: sessionJSON) {
            sessionData = session
            if session.isTracking {
                startTrackingTimer()
            }
        }

        if let pomodoroJSON = response["pomodoro"] as? Data,
           let pomodoro = try? JSONDecoder().decode(WatchPomodoroData.self, from: pomodoroJSON) {
            pomodoroData = pomodoro
            if pomodoro.isActive && !pomodoro.isPaused {
                startPomodoroTimer()
            }
        }

        if let statsJSON = response["stats"] as? Data,
           let stats = try? JSONDecoder().decode(WatchStatsData.self, from: statsJSON) {
            statsData = stats
        }

        if let projects = response["quickProjects"] as? [String] {
            quickProjects = projects
        }

        lastSyncTime = Date()
    }

    // MARK: - Time Tracking Actions
    func startTracking(project: String) {
        sessionData.isTracking = true
        sessionData.projectName = project
        sessionData.startTime = Date()
        sessionData.elapsedSeconds = 0
        startTrackingTimer()

        // Send to iPhone
        sendAction("startTracking", data: ["project": project])
    }

    func stopTracking() {
        sessionData.isTracking = false
        stopTrackingTimer()

        // Send to iPhone
        sendAction("stopTracking", data: [:])
    }

    private func startTrackingTimer() {
        stopTrackingTimer()
        trackingTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            DispatchQueue.main.async {
                self?.sessionData.elapsedSeconds += 1
            }
        }
    }

    private func stopTrackingTimer() {
        trackingTimer?.invalidate()
        trackingTimer = nil
    }

    // MARK: - Pomodoro Actions
    func startPomodoro() {
        pomodoroData.isActive = true
        pomodoroData.isBreak = false
        pomodoroData.isPaused = false
        pomodoroData.secondsRemaining = pomodoroData.workMinutes * 60
        startPomodoroTimer()

        sendAction("startPomodoro", data: [:])
    }

    func pausePomodoro() {
        pomodoroData.isPaused = true
        stopPomodoroTimer()

        sendAction("pausePomodoro", data: [:])
    }

    func resumePomodoro() {
        pomodoroData.isPaused = false
        startPomodoroTimer()

        sendAction("resumePomodoro", data: [:])
    }

    func stopPomodoro() {
        pomodoroData.isActive = false
        pomodoroData.isPaused = false
        stopPomodoroTimer()

        sendAction("stopPomodoro", data: [:])
    }

    func skipPomodoro() {
        if pomodoroData.isBreak {
            // End break, start work
            pomodoroData.isBreak = false
            pomodoroData.secondsRemaining = pomodoroData.workMinutes * 60
        } else {
            // End work, start break
            pomodoroData.sessionsCompleted += 1
            pomodoroData.isBreak = true
            pomodoroData.secondsRemaining = pomodoroData.breakMinutes * 60
        }

        sendAction("skipPomodoro", data: [:])
    }

    private func startPomodoroTimer() {
        stopPomodoroTimer()
        pomodoroTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            DispatchQueue.main.async {
                guard let self = self else { return }
                if self.pomodoroData.secondsRemaining > 0 {
                    self.pomodoroData.secondsRemaining -= 1
                } else {
                    self.handlePomodoroComplete()
                }
            }
        }
    }

    private func stopPomodoroTimer() {
        pomodoroTimer?.invalidate()
        pomodoroTimer = nil
    }

    private func handlePomodoroComplete() {
        // Haptic feedback
        WKInterfaceDevice.current().play(.notification)

        if pomodoroData.isBreak {
            // Break finished, start work
            pomodoroData.isBreak = false
            pomodoroData.secondsRemaining = pomodoroData.workMinutes * 60
        } else {
            // Work finished, start break
            pomodoroData.sessionsCompleted += 1
            pomodoroData.isBreak = true
            pomodoroData.secondsRemaining = pomodoroData.breakMinutes * 60
        }
    }

    // MARK: - Helper Methods
    private func sendAction(_ action: String, data: [String: Any]) {
        guard WCSession.default.isReachable else {
            // Queue for later if not reachable
            return
        }

        var message: [String: Any] = ["action": action]
        message.merge(data) { _, new in new }

        WCSession.default.sendMessage(message, replyHandler: nil) { error in
            print("Watch send error: \(error)")
        }
    }

    // MARK: - Formatters
    var formattedTrackingTime: String {
        let hours = sessionData.elapsedSeconds / 3600
        let minutes = (sessionData.elapsedSeconds % 3600) / 60
        let seconds = sessionData.elapsedSeconds % 60
        return String(format: "%d:%02d:%02d", hours, minutes, seconds)
    }

    var formattedPomodoroTime: String {
        let minutes = pomodoroData.secondsRemaining / 60
        let seconds = pomodoroData.secondsRemaining % 60
        return String(format: "%02d:%02d", minutes, seconds)
    }

    var pomodoroProgress: Double {
        let total = pomodoroData.isBreak
            ? pomodoroData.breakMinutes * 60
            : pomodoroData.workMinutes * 60
        return Double(total - pomodoroData.secondsRemaining) / Double(total)
    }

    var weeklyProgress: Double {
        guard statsData.weeklyTarget > 0 else { return 0 }
        return min(statsData.weeklyHours / statsData.weeklyTarget, 1.0)
    }
}

// MARK: - WCSessionDelegate
extension WatchAppState: WCSessionDelegate {
    nonisolated func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        DispatchQueue.main.async {
            self.isConnected = activationState == .activated
            if activationState == .activated {
                self.requestSync()
            }
        }
    }

    nonisolated func session(_ session: WCSession, didReceiveMessage message: [String: Any]) {
        DispatchQueue.main.async {
            self.handleSyncResponse(message)
        }
    }

    nonisolated func sessionReachabilityDidChange(_ session: WCSession) {
        DispatchQueue.main.async {
            self.isConnected = session.isReachable
            if session.isReachable {
                self.requestSync()
            }
        }
    }
}

// MARK: - WKInterfaceDevice Extension
import WatchKit
