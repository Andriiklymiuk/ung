//
//  OnboardingView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

// MARK: - Database Error View
struct DatabaseErrorView: View {
    @EnvironmentObject var appState: AppState
    @State private var isRetrying = false

    var body: some View {
        VStack(spacing: 0) {
            Spacer()
                .frame(height: 32)

            // App icon
            appIcon

            Spacer()
                .frame(height: 16)

            // Title
            Text("Database Error")
                .font(.system(size: 22, weight: .bold))

            Spacer()
                .frame(height: 6)

            // Subtitle
            Text("Failed to initialize database")
                .font(.system(size: 13))
                .foregroundColor(.secondary)

            Spacer()
                .frame(height: 24)

            // Error info
            VStack(spacing: 16) {
                Text("The app couldn't create or access the database. This might be due to permissions or disk space issues.")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)

                // Retry button
                Button(action: retry) {
                    HStack {
                        if isRetrying {
                            ProgressView()
                                .scaleEffect(0.7)
                        }
                        Text(isRetrying ? "Retrying..." : "Retry")
                    }
                    .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.large)
                .disabled(isRetrying)
            }

            Spacer()

            // Footer
            footerSection
                .padding(.bottom, 20)
        }
        .padding(.horizontal, 32)
    }

    // MARK: - App Icon
    private var appIcon: some View {
        Group {
            #if os(macOS)
            if let appIconImage = NSImage(named: "AppIcon") {
                Image(nsImage: appIconImage)
                    .resizable()
                    .frame(width: 80, height: 80)
            } else {
                fallbackIcon
            }
            #else
            if let appIconImage = UIImage(named: "AppIcon") {
                Image(uiImage: appIconImage)
                    .resizable()
                    .frame(width: 80, height: 80)
            } else {
                fallbackIcon
            }
            #endif
        }
    }

    private var fallbackIcon: some View {
        Image(systemName: "exclamationmark.triangle")
            .font(.system(size: 48, weight: .light))
            .foregroundColor(.orange)
            .frame(width: 80, height: 80)
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack {
            Button("Learn More") {
                #if os(macOS)
                if let url = URL(string: "https://andriiklymiuk.github.io/ung/") {
                    NSWorkspace.shared.open(url)
                }
                #else
                if let url = URL(string: "https://andriiklymiuk.github.io/ung/") {
                    UIApplication.shared.open(url)
                }
                #endif
            }
            .buttonStyle(.plain)
            .foregroundColor(.accentColor)
            .font(.system(size: 12))

            Spacer()

            Text("v1.0")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }

    // MARK: - Actions
    private func retry() {
        isRetrying = true
        Task {
            await appState.initializeDatabase()
            isRetrying = false
        }
    }
}

#Preview("Database Error") {
    DatabaseErrorView()
        .environmentObject(AppState())
        .frame(width: 340, height: 300)
}
