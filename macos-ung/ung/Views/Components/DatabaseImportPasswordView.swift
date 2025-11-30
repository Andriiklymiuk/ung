//
//  DatabaseImportPasswordView.swift
//  ung
//
//  Password prompt view for importing encrypted database files.
//

import SwiftUI

struct DatabaseImportPasswordView: View {
    let fileURL: URL?
    let onImport: (String) -> Void
    let onCancel: () -> Void

    @EnvironmentObject var appState: AppState
    @State private var password = ""
    @State private var isVerifying = false
    @State private var errorMessage: String?
    @State private var saveToKeychain = true

    var body: some View {
        VStack(spacing: Design.Spacing.lg) {
            // Header
            VStack(spacing: Design.Spacing.sm) {
                Image(systemName: "lock.doc.fill")
                    .font(.system(size: 48))
                    .foregroundColor(.orange)

                Text("Encrypted Database")
                    .font(Design.Typography.headingMedium)

                if let url = fileURL {
                    Text(url.lastPathComponent)
                        .font(Design.Typography.bodySmall)
                        .foregroundColor(Design.Colors.textSecondary)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }
            }
            .padding(.top, Design.Spacing.md)

            // Description
            VStack(alignment: .leading, spacing: Design.Spacing.xs) {
                HStack(spacing: Design.Spacing.xs) {
                    Image(systemName: "info.circle.fill")
                        .foregroundColor(.blue)
                    Text("This database is encrypted")
                        .font(Design.Typography.labelMedium)
                }

                Text("Enter the password used to encrypt this database. This is the same password used in the UNG CLI or another UNG app.")
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(Design.Colors.textSecondary)
            }
            .padding(Design.Spacing.md)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.md)
                    .fill(Color.blue.opacity(0.1))
            )

            // Password Field
            VStack(alignment: .leading, spacing: Design.Spacing.xs) {
                Text("Password")
                    .font(Design.Typography.labelMedium)
                    .foregroundColor(Design.Colors.textSecondary)

                SecureField("Enter encryption password", text: $password)
                    .textFieldStyle(.roundedBorder)
                    .onSubmit {
                        if !password.isEmpty {
                            performImport()
                        }
                    }

                if let error = errorMessage {
                    HStack(spacing: Design.Spacing.xs) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundColor(.red)
                        Text(error)
                            .font(Design.Typography.bodySmall)
                            .foregroundColor(.red)
                    }
                }
            }

            // Save to Keychain option
            Toggle(isOn: $saveToKeychain) {
                HStack(spacing: Design.Spacing.xs) {
                    Image(systemName: "key.icloud")
                        .foregroundColor(.secondary)
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Save password to Keychain")
                            .font(Design.Typography.labelMedium)
                        Text("Automatically unlock on future launches")
                            .font(Design.Typography.bodySmall)
                            .foregroundColor(Design.Colors.textSecondary)
                    }
                }
            }
            .toggleStyle(.checkbox)

            // Info about merge
            HStack(spacing: Design.Spacing.xs) {
                Image(systemName: "arrow.triangle.merge")
                    .foregroundColor(.green)
                Text("Data will be merged with your existing records")
                    .font(Design.Typography.bodySmall)
                    .foregroundColor(.green)
            }
            .padding(Design.Spacing.sm)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.sm)
                    .fill(Color.green.opacity(0.1))
            )

            Spacer()

            // Action Buttons
            HStack(spacing: Design.Spacing.md) {
                Button("Cancel") {
                    onCancel()
                }
                .buttonStyle(.bordered)
                .keyboardShortcut(.escape)

                Button(action: performImport) {
                    if isVerifying {
                        ProgressView()
                            .scaleEffect(0.8)
                    } else {
                        Text("Import & Merge")
                    }
                }
                .buttonStyle(.borderedProminent)
                .tint(.orange)
                .disabled(password.isEmpty || isVerifying)
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(Design.Spacing.lg)
        .frame(width: 420, height: 520)
    }

    private func performImport() {
        guard !password.isEmpty else { return }

        isVerifying = true
        errorMessage = nil

        Task {
            // Verify password first
            let isValid = await verifyPassword()

            await MainActor.run {
                isVerifying = false

                if isValid {
                    // Save to keychain if requested
                    if saveToKeychain {
                        _ = appState.saveEncryptionPasswordToKeychain(password)
                    }
                    onImport(password)
                } else {
                    errorMessage = "Incorrect password. Please try again."
                    password = ""
                }
            }
        }
    }

    private func verifyPassword() async -> Bool {
        guard let url = fileURL else { return false }

        // Try to decrypt a small portion to verify password
        let encryption = DatabaseEncryptionService.shared
        let tempPath = FileManager.default.temporaryDirectory
            .appendingPathComponent("verify_\(UUID().uuidString).db")
            .path

        do {
            try await encryption.decryptDatabase(
                inputPath: url.path,
                outputPath: tempPath,
                password: password
            )
            // Clean up
            try? FileManager.default.removeItem(atPath: tempPath)
            return true
        } catch {
            try? FileManager.default.removeItem(atPath: tempPath)
            return false
        }
    }
}

// MARK: - Preview

#Preview {
    DatabaseImportPasswordView(
        fileURL: URL(fileURLWithPath: "/path/to/ung.db.encrypted"),
        onImport: { _ in },
        onCancel: {}
    )
    .environmentObject(AppState())
}
