//
//  ungApp.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

@main
struct ungApp: App {
    @StateObject private var appState = AppState()

    var body: some Scene {
        MenuBarExtra {
            MenuBarView()
                .environmentObject(appState)
        } label: {
            Label {
                Text("UNG")
            } icon: {
                if appState.isTracking {
                    Image(systemName: "record.circle.fill")
                        .symbolRenderingMode(.palette)
                        .foregroundStyle(.red, .primary)
                } else {
                    Image(systemName: "clock.badge.checkmark")
                        .symbolRenderingMode(.hierarchical)
                }
            }
        }
        .menuBarExtraStyle(.window)
    }
}
