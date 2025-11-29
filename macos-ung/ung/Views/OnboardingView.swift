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
                .frame(height: Design.Spacing.xl)

            // App icon
            appIcon

            Spacer()
                .frame(height: Design.Spacing.md)

            // Title
            Text("Database Error")
                .font(Design.Typography.headingMedium)
                .accessibleHeader(label: "Database Error")

            Spacer()
                .frame(height: Design.Spacing.xxs)

            // Subtitle
            Text("Failed to initialize database")
                .font(Design.Typography.bodyMedium)
                .foregroundColor(Design.Colors.textSecondary)

            Spacer()
                .frame(height: Design.Spacing.lg)

            // Error info
            VStack(spacing: Design.Spacing.md) {
                Text("The app couldn't create or access the database. This might be due to permissions or disk space issues.")
                    .font(Design.Typography.bodyMedium)
                    .foregroundColor(Design.Colors.textSecondary)
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
                .accessibleButton(label: isRetrying ? "Retrying database initialization" : "Retry database initialization")
            }

            Spacer()

            // Footer
            footerSection
                .padding(.bottom, Design.Spacing.lg)
        }
        .padding(.horizontal, Design.Spacing.xl)
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
            .foregroundColor(Design.Colors.warning)
            .frame(width: 80, height: 80)
            .accessibilityLabel("Warning icon")
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
            .foregroundColor(Design.Colors.brand)
            .font(Design.Typography.bodySmall)
            .accessibleButton(label: "Learn more about UNG")

            Spacer()

            Text("v1.0")
                .font(Design.Typography.bodySmall)
                .foregroundColor(Design.Colors.textSecondary)
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
