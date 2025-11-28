//
//  LockScreenView.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import LocalAuthentication
import SwiftUI

struct LockScreenView: View {
  @EnvironmentObject var appState: AppState
  @Environment(\.colorScheme) var colorScheme
  @State private var isAuthenticating = false
  @State private var authError: String?
  @State private var showError = false
  @State private var biometricType: LABiometryType = .none

  var body: some View {
    ZStack {
      // Background with blur effect
      backgroundView

      // Lock content
      VStack(spacing: Design.Spacing.xl) {
        Spacer()

        // App icon and lock indicator
        lockIconView

        // Title
        VStack(spacing: Design.Spacing.xs) {
          Text("UNG is Locked")
            .font(Design.Typography.headingLarge)
            .foregroundColor(Design.Colors.textPrimary)

          Text("Authenticate to access your data")
            .font(Design.Typography.bodyMedium)
            .foregroundColor(Design.Colors.textSecondary)
        }

        Spacer()

        // Unlock button
        unlockButton

        // Error message
        if showError, let error = authError {
          Text(error)
            .font(Design.Typography.bodySmall)
            .foregroundColor(.red)
            .padding(.top, Design.Spacing.xs)
        }

        Spacer()
      }
      .padding(Design.Spacing.xxl)
    }
    .frame(minWidth: 400, minHeight: 300)
    .onAppear {
      checkBiometricType()
    }
  }

  private var backgroundView: some View {
    ZStack {
      // Gradient background
      LinearGradient(
        colors: colorScheme == .dark
          ? [Color(white: 0.1), Color(white: 0.15)]
          : [Color(white: 0.95), Color(white: 0.98)],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
      )
      .ignoresSafeArea()

      // Subtle pattern overlay
      Circle()
        .fill(
          RadialGradient(
            colors: [Color.blue.opacity(0.1), Color.clear],
            center: .center,
            startRadius: 0,
            endRadius: 300
          )
        )
        .frame(width: 600, height: 600)
        .blur(radius: 60)
        .offset(x: -100, y: -100)

      Circle()
        .fill(
          RadialGradient(
            colors: [Color.purple.opacity(0.08), Color.clear],
            center: .center,
            startRadius: 0,
            endRadius: 250
          )
        )
        .frame(width: 500, height: 500)
        .blur(radius: 50)
        .offset(x: 150, y: 150)
    }
  }

  private var lockIconView: some View {
    ZStack {
      // Outer ring with animation
      Circle()
        .stroke(
          LinearGradient(
            colors: [Color.blue.opacity(0.3), Color.purple.opacity(0.2)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
          ),
          lineWidth: 3
        )
        .frame(width: 120, height: 120)

      // Inner circle
      Circle()
        .fill(
          LinearGradient(
            colors: colorScheme == .dark
              ? [Color(white: 0.18), Color(white: 0.14)]
              : [Color.white, Color(white: 0.95)],
            startPoint: .top,
            endPoint: .bottom
          )
        )
        .frame(width: 100, height: 100)
        .shadow(color: .black.opacity(0.1), radius: 10, y: 5)

      // Icon
      VStack(spacing: 4) {
        if isAuthenticating {
          ProgressView()
            .scaleEffect(1.2)
        } else {
          Image(systemName: biometricIcon)
            .font(.system(size: 36, weight: .light))
            .foregroundColor(.blue)
        }
      }
    }
  }

  private var biometricIcon: String {
    switch biometricType {
    case .touchID:
      return "touchid"
    case .faceID:
      return "faceid"
    default:
      return "lock.fill"
    }
  }

  private var unlockButton: some View {
    Button(action: {
      Task {
        await authenticate()
      }
    }) {
      HStack(spacing: Design.Spacing.sm) {
        if isAuthenticating {
          ProgressView()
            .scaleEffect(0.8)
            .tint(.white)
        } else {
          Image(systemName: biometricIcon)
            .font(.system(size: 16, weight: .medium))
        }

        Text(unlockButtonText)
          .font(Design.Typography.labelLarge)
      }
      .foregroundColor(.white)
      .frame(width: 200, height: 48)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.md)
          .fill(
            LinearGradient(
              colors: [Color.blue, Color.blue.opacity(0.8)],
              startPoint: .top,
              endPoint: .bottom
            )
          )
          .shadow(color: Color.blue.opacity(0.3), radius: 8, y: 4)
      )
    }
    .buttonStyle(.plain)
    .disabled(isAuthenticating)
    .scaleEffect(isAuthenticating ? 0.98 : 1.0)
    .animation(Design.Animation.quick, value: isAuthenticating)
  }

  private var unlockButtonText: String {
    if isAuthenticating {
      return "Authenticating..."
    }

    switch biometricType {
    case .touchID:
      return "Unlock with Touch ID"
    case .faceID:
      return "Unlock with Face ID"
    default:
      return "Unlock"
    }
  }

  private func checkBiometricType() {
    let context = LAContext()
    var error: NSError?
    if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
      biometricType = context.biometryType
    }
  }

  private func authenticate() async {
    guard !isAuthenticating else { return }

    isAuthenticating = true
    showError = false
    authError = nil

    var success = false

    if appState.useTouchID {
      success = await appState.authenticateWithBiometrics()
      if !success {
        // Fall back to password
        success = await appState.authenticateWithPassword()
      }
    } else {
      success = await appState.authenticateWithPassword()
    }

    isAuthenticating = false

    if !success {
      authError = "Authentication failed. Please try again."
      showError = true
    }
  }
}

#Preview {
  LockScreenView()
    .environmentObject(AppState())
    .frame(width: 500, height: 400)
}
