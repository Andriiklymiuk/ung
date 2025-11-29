//
//  ungWatchApp.swift
//  ung Watch App
//
//  Apple Watch companion app for ung time tracking
//

import SwiftUI
import WatchConnectivity

@main
struct ungWatchApp: App {
    @StateObject private var watchState = WatchAppState()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(watchState)
        }
    }
}

// MARK: - Content View
struct ContentView: View {
    @EnvironmentObject var watchState: WatchAppState

    var body: some View {
        TabView {
            // Time Tracking Tab
            TimeTrackingView()

            // Pomodoro Tab
            PomodoroWatchView()

            // Quick Stats Tab
            QuickStatsView()
        }
        .tabViewStyle(.verticalPage)
    }
}

#Preview {
    ContentView()
        .environmentObject(WatchAppState())
}
