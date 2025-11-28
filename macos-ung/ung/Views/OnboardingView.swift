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

  @State private var isInitializing = false
  @State private var copiedCommand = false

  private let brewCommand = "brew tap Andriiklymiuk/ung && brew install ung"

  var body: some View {
    VStack(spacing: 0) {
      Spacer()
        .frame(height: 32)

      // App icon
      appIcon

      Spacer()
        .frame(height: 16)

      // Title
      Text("Welcome to UNG")
        .font(.system(size: 22, weight: .bold))

      Spacer()
        .frame(height: 6)

      // Subtitle
      Text("Time tracking & invoicing for freelancers")
        .font(.system(size: 13))
        .foregroundColor(.secondary)

      Spacer()
        .frame(height: 24)

      // Main content
      switch state {
      case .notInstalled:
        notInstalledContent
      case .notInitialized:
        notInitializedContent
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
      if let appIconImage = NSImage(named: "AppIcon") {
        Image(nsImage: appIconImage)
          .resizable()
          .frame(width: 80, height: 80)
      } else {
        Image(systemName: "clock.badge.checkmark")
          .font(.system(size: 48, weight: .light))
          .foregroundColor(.accentColor)
          .frame(width: 80, height: 80)
      }
    }
  }

  // MARK: - Not Installed Content
  private var notInstalledContent: some View {
    VStack(spacing: 16) {
      Text("Install UNG CLI to get started")
        .font(.system(size: 13))
        .foregroundColor(.secondary)
        .multilineTextAlignment(.center)

      // Command box
      VStack(spacing: 8) {
        HStack {
          Text(brewCommand)
            .font(.system(size: 11, design: .monospaced))
            .foregroundColor(.primary)
            .lineLimit(1)

          Spacer()

          Button(action: copyCommand) {
            Image(systemName: copiedCommand ? "checkmark" : "doc.on.doc")
              .font(.system(size: 12))
              .foregroundColor(copiedCommand ? .green : .secondary)
          }
          .buttonStyle(.plain)
        }
        .padding(12)
        .background(
          RoundedRectangle(cornerRadius: 8)
            .fill(Color(nsColor: .textBackgroundColor))
            .overlay(
              RoundedRectangle(cornerRadius: 8)
                .stroke(Color.primary.opacity(0.1), lineWidth: 1)
            )
        )

        Text("Run this command in Terminal")
          .font(.system(size: 11))
          .foregroundColor(Color(nsColor: .tertiaryLabelColor))
      }

      // Copy button
      Button(action: copyCommand) {
        Text(copiedCommand ? "Copied!" : "Copy Command")
          .frame(maxWidth: .infinity)
      }
      .buttonStyle(.borderedProminent)
      .controlSize(.large)

      // Recheck button
      Button("Recheck Installation") {
        appState.checkStatus()
      }
      .buttonStyle(.plain)
      .foregroundColor(.accentColor)
      .font(.system(size: 12))
      .padding(.top, 4)
    }
  }

  // MARK: - Not Initialized Content
  private var notInitializedContent: some View {
    VStack(spacing: 16) {
      // Success indicator
      HStack(spacing: 6) {
        Image(systemName: "checkmark.circle.fill")
          .foregroundColor(.green)
          .font(.system(size: 14))
        Text("CLI installed successfully")
          .font(.system(size: 13))
          .foregroundColor(.secondary)
      }

      Text("Choose where to store your data:")
        .font(.system(size: 13))
        .foregroundColor(.secondary)

      VStack(spacing: 10) {
        // Global setup
        SetupButton(
          title: "Global Setup",
          subtitle: "Store data in ~/.ung (recommended)",
          isLoading: isInitializing
        ) {
          Task {
            isInitializing = true
            await appState.initializeGlobal()
            isInitializing = false
          }
        }

        // Local setup
        SetupButton(
          title: "Project Setup",
          subtitle: "Store data in current folder",
          isLoading: false
        ) {
          Task {
            await appState.initializeLocal()
          }
        }
      }
    }
  }

  // MARK: - Footer
  private var footerSection: some View {
    HStack {
      Button("Learn More") {
        if let url = URL(string: "https://andriiklymiuk.github.io/ung/") {
          NSWorkspace.shared.open(url)
        }
      }
      .buttonStyle(.plain)
      .foregroundColor(.accentColor)
      .font(.system(size: 12))

      Spacer()

      Text("v1.0")
        .font(.system(size: 11))
        .foregroundColor(Color(nsColor: .tertiaryLabelColor))
    }
  }

  // MARK: - Actions
  private func copyCommand() {
    NSPasteboard.general.clearContents()
    NSPasteboard.general.setString(brewCommand, forType: .string)
    copiedCommand = true

    // Reset after delay
    DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
      copiedCommand = false
    }
  }
}

// MARK: - Setup Button
struct SetupButton: View {
  let title: String
  let subtitle: String
  let isLoading: Bool
  let action: () -> Void

  @State private var isHovered = false

  var body: some View {
    Button(action: action) {
      HStack {
        VStack(alignment: .leading, spacing: 2) {
          Text(title)
            .font(.system(size: 13, weight: .medium))
            .foregroundColor(.primary)

          Text(subtitle)
            .font(.system(size: 11))
            .foregroundColor(.secondary)
        }

        Spacer()

        if isLoading {
          ProgressView()
            .scaleEffect(0.6)
        } else {
          Image(systemName: "chevron.right")
            .font(.system(size: 12, weight: .medium))
            .foregroundColor(Color(nsColor: .tertiaryLabelColor))
        }
      }
      .padding(.horizontal, 14)
      .padding(.vertical, 12)
      .background(
        RoundedRectangle(cornerRadius: 8)
          .fill(isHovered ? Color.primary.opacity(0.05) : Color.clear)
          .overlay(
            RoundedRectangle(cornerRadius: 8)
              .stroke(Color.primary.opacity(0.1), lineWidth: 1)
          )
      )
    }
    .buttonStyle(.plain)
    .onHover { hovering in
      isHovered = hovering
    }
    .disabled(isLoading)
  }
}

#Preview("Not Installed") {
  OnboardingView(state: .notInstalled)
    .environmentObject(AppState())
    .frame(width: 340, height: 400)
}

#Preview("Not Initialized") {
  OnboardingView(state: .notInitialized)
    .environmentObject(AppState())
    .frame(width: 340, height: 400)
}
