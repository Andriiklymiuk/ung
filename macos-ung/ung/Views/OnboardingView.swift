//
//  OnboardingView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

enum OnboardingState {
    case notInstalled
    case notInitialized
}

struct OnboardingView: View {
    @EnvironmentObject var appState: AppState
    let state: OnboardingState

    @State private var showFeatures = false
    @State private var isInstallingHomebrew = false
    @State private var isInitializing = false

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                // Header
                headerSection

                // Main content based on state
                switch state {
                case .notInstalled:
                    notInstalledContent
                case .notInitialized:
                    notInitializedContent
                }

                // Features section (collapsible)
                featuresSection

                // Footer
                footerSection
            }
            .padding(16)
        }
        .frame(maxHeight: 500)
    }

    // MARK: - Header
    private var headerSection: some View {
        VStack(spacing: 12) {
            // Logo
            ZStack {
                RoundedRectangle(cornerRadius: 16)
                    .fill(
                        LinearGradient(
                            colors: [.blue.opacity(0.15), .purple.opacity(0.15)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .frame(width: 64, height: 64)

                Image(systemName: "clock.badge.checkmark")
                    .font(.system(size: 28, weight: .medium))
                    .foregroundStyle(
                        LinearGradient(
                            colors: [.blue, .purple],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
            }

            VStack(spacing: 4) {
                Text("Welcome to UNG")
                    .font(.system(size: 18, weight: .bold, design: .rounded))

                Text("Time tracking & invoicing for freelancers")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
            }

            // Status badge
            statusBadge
        }
    }

    private var statusBadge: some View {
        HStack(spacing: 6) {
            Circle()
                .fill(state == .notInstalled ? Color.orange : Color.blue)
                .frame(width: 8, height: 8)

            Text(state == .notInstalled ? "CLI Not Installed" : "Setup Required")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(state == .notInstalled ? .orange : .blue)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 6)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(state == .notInstalled ? Color.orange.opacity(0.15) : Color.blue.opacity(0.15))
        )
    }

    // MARK: - Not Installed Content
    private var notInstalledContent: some View {
        VStack(spacing: 12) {
            Text("Install UNG CLI")
                .font(.system(size: 14, weight: .semibold))
                .frame(maxWidth: .infinity, alignment: .leading)

            // Homebrew (Recommended)
            InstallOptionButton(
                icon: "cup.and.saucer.fill",
                title: "Homebrew",
                subtitle: "Recommended",
                isRecommended: true,
                isLoading: isInstallingHomebrew
            ) {
                installViaHomebrew()
            }

            // Go install
            InstallOptionButton(
                icon: "chevron.left.forwardslash.chevron.right",
                title: "Go Install",
                subtitle: "go install github.com/...",
                isRecommended: false,
                isLoading: false
            ) {
                openGoInstall()
            }

            // Manual download
            InstallOptionButton(
                icon: "arrow.down.circle.fill",
                title: "Download Binary",
                subtitle: "Manual installation",
                isRecommended: false,
                isLoading: false
            ) {
                openDownloadPage()
            }

            // Recheck button
            Button(action: { appState.checkStatus() }) {
                HStack(spacing: 6) {
                    Image(systemName: "arrow.clockwise")
                        .font(.system(size: 12, weight: .medium))
                    Text("Recheck Installation")
                        .font(.system(size: 12, weight: .medium))
                }
                .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
            .padding(.top, 4)
        }
    }

    // MARK: - Not Initialized Content
    private var notInitializedContent: some View {
        VStack(spacing: 12) {
            // Success badge
            HStack(spacing: 6) {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundColor(.green)
                    .font(.system(size: 12))
                Text("CLI Installed")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.green)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 6)
            .background(
                RoundedRectangle(cornerRadius: 12)
                    .fill(Color.green.opacity(0.15))
            )

            Text("Choose Setup Type")
                .font(.system(size: 14, weight: .semibold))
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.top, 4)

            // Global setup (Recommended)
            SetupOptionButton(
                icon: "globe",
                title: "Global Setup",
                description: "Data stored in ~/.ung/\nAccessible from anywhere",
                isRecommended: true,
                isLoading: isInitializing
            ) {
                Task {
                    isInitializing = true
                    await appState.initializeGlobal()
                    isInitializing = false
                }
            }

            // Local setup
            SetupOptionButton(
                icon: "folder.fill",
                title: "Project Setup",
                description: "Data stored in .ung/\nIsolated to this folder",
                isRecommended: false,
                isLoading: false
            ) {
                Task {
                    await appState.initializeLocal()
                }
            }

            // Help text
            Text("Choose Global for most use cases. Project setup is useful for separate business tracking.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.top, 4)
        }
    }

    // MARK: - Features Section
    private var featuresSection: some View {
        VStack(spacing: 8) {
            Button(action: { withAnimation(.spring(response: 0.3)) { showFeatures.toggle() } }) {
                HStack {
                    Text("Features")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                    Spacer()
                    Image(systemName: showFeatures ? "chevron.up" : "chevron.down")
                        .font(.system(size: 10, weight: .medium))
                        .foregroundColor(.secondary)
                }
            }
            .buttonStyle(.plain)

            if showFeatures {
                VStack(spacing: 8) {
                    FeatureRow(icon: "clock.fill", title: "Time Tracking", color: .blue)
                    FeatureRow(icon: "doc.text.fill", title: "Invoice Generation", color: .green)
                    FeatureRow(icon: "person.2.fill", title: "Client Management", color: .purple)
                    FeatureRow(icon: "dollarsign.circle.fill", title: "Expense Tracking", color: .orange)
                    FeatureRow(icon: "chart.bar.fill", title: "Reports & Analytics", color: .pink)
                    FeatureRow(icon: "lock.fill", title: "100% Local & Private", color: .gray)
                }
                .transition(.opacity.combined(with: .move(edge: .top)))
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack(spacing: 16) {
            Button("Documentation") {
                if let url = URL(string: "https://github.com/Andriiklymiuk/ung") {
                    NSWorkspace.shared.open(url)
                }
            }
            .buttonStyle(.plain)
            .font(.system(size: 11))
            .foregroundColor(.blue)

            Spacer()

            Text("v1.0")
                .font(.system(size: 10))
                .foregroundColor(.secondary)
        }
    }

    // MARK: - Actions
    private func installViaHomebrew() {
        isInstallingHomebrew = true
        // Open Terminal with Homebrew command
        let script = """
        tell application "Terminal"
            activate
            do script "brew tap Andriiklymiuk/ung && brew install ung"
        end tell
        """
        if let appleScript = NSAppleScript(source: script) {
            var error: NSDictionary?
            appleScript.executeAndReturnError(&error)
        }
        isInstallingHomebrew = false
    }

    private func openGoInstall() {
        let script = """
        tell application "Terminal"
            activate
            do script "go install github.com/Andriiklymiuk/ung@latest"
        end tell
        """
        if let appleScript = NSAppleScript(source: script) {
            var error: NSDictionary?
            appleScript.executeAndReturnError(&error)
        }
    }

    private func openDownloadPage() {
        if let url = URL(string: "https://github.com/Andriiklymiuk/ung/releases") {
            NSWorkspace.shared.open(url)
        }
    }
}

// MARK: - Supporting Views
struct InstallOptionButton: View {
    let icon: String
    let title: String
    let subtitle: String
    let isRecommended: Bool
    let isLoading: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.system(size: 16))
                    .foregroundColor(isRecommended ? .blue : .secondary)
                    .frame(width: 24)

                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 6) {
                        Text(title)
                            .font(.system(size: 13, weight: .medium))
                            .foregroundColor(.primary)

                        if isRecommended {
                            Text("Recommended")
                                .font(.system(size: 9, weight: .semibold))
                                .foregroundColor(.blue)
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(Color.blue.opacity(0.15))
                                .cornerRadius(4)
                        }
                    }

                    Text(subtitle)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                Spacer()

                if isLoading {
                    ProgressView()
                        .scaleEffect(0.7)
                } else {
                    Image(systemName: "chevron.right")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.secondary)
                }
            }
            .padding(12)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color(nsColor: .controlBackgroundColor))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(isRecommended ? Color.blue.opacity(0.3) : Color.clear, lineWidth: 1)
                    )
            )
        }
        .buttonStyle(.plain)
    }
}

struct SetupOptionButton: View {
    let icon: String
    let title: String
    let description: String
    let isRecommended: Bool
    let isLoading: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                ZStack {
                    RoundedRectangle(cornerRadius: 8)
                        .fill(isRecommended ? Color.blue.opacity(0.15) : Color.gray.opacity(0.15))
                        .frame(width: 36, height: 36)

                    Image(systemName: icon)
                        .font(.system(size: 16))
                        .foregroundColor(isRecommended ? .blue : .secondary)
                }

                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 6) {
                        Text(title)
                            .font(.system(size: 13, weight: .medium))
                            .foregroundColor(.primary)

                        if isRecommended {
                            Text("Recommended")
                                .font(.system(size: 9, weight: .semibold))
                                .foregroundColor(.blue)
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(Color.blue.opacity(0.15))
                                .cornerRadius(4)
                        }
                    }

                    Text(description)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }

                Spacer()

                if isLoading {
                    ProgressView()
                        .scaleEffect(0.7)
                } else {
                    Image(systemName: "chevron.right")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.secondary)
                }
            }
            .padding(12)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color(nsColor: .controlBackgroundColor))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(isRecommended ? Color.blue.opacity(0.3) : Color.clear, lineWidth: 1)
                    )
            )
        }
        .buttonStyle(.plain)
    }
}

struct FeatureRow: View {
    let icon: String
    let title: String
    let color: Color

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: icon)
                .font(.system(size: 12))
                .foregroundColor(color)
                .frame(width: 20)

            Text(title)
                .font(.system(size: 12))
                .foregroundColor(.primary)

            Spacer()
        }
    }
}

#Preview("Not Installed") {
    OnboardingView(state: .notInstalled)
        .environmentObject(AppState())
        .frame(width: 320)
}

#Preview("Not Initialized") {
    OnboardingView(state: .notInitialized)
        .environmentObject(AppState())
        .frame(width: 320)
}
