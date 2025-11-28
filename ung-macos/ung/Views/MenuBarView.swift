//
//  MenuBarView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct MenuBarView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(spacing: 0) {
            switch appState.status {
            case .loading:
                LoadingView()
            case .cliNotInstalled:
                OnboardingView(state: .notInstalled)
            case .notInitialized:
                OnboardingView(state: .notInitialized)
            case .ready:
                DashboardView()
            }
        }
        .frame(width: 320)
        .background(Color(nsColor: .windowBackgroundColor))
    }
}

#Preview {
    MenuBarView()
        .environmentObject(AppState())
}
