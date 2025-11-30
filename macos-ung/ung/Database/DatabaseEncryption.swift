//
//  DatabaseEncryption.swift
//  ung
//
//  AES-256-GCM encryption/decryption compatible with the Go CLI.
//  File format: [32-byte salt][12-byte nonce][ciphertext with auth tag]
//
//  This implementation matches the Go CLI encryption in internal/db/encryption.go
//  to ensure cross-platform database compatibility.
//

import CommonCrypto
import CryptoKit
import Foundation

// MARK: - Encryption Constants

enum EncryptionConstants {
    static let keySize = 32  // AES-256
    static let saltSize = 32
    static let nonceSize = 12  // GCM standard nonce size
    static let pbkdf2Iterations: UInt32 = 100_000
    static let encryptedFileExtension = ".encrypted"
}

// MARK: - Encryption Errors

enum EncryptionError: LocalizedError {
    case invalidPassword
    case encryptionFailed(String)
    case decryptionFailed(String)
    case fileTooShort
    case fileReadFailed(String)
    case fileWriteFailed(String)
    case keyDerivationFailed

    var errorDescription: String? {
        switch self {
        case .invalidPassword:
            return "Invalid password"
        case .encryptionFailed(let message):
            return "Encryption failed: \(message)"
        case .decryptionFailed(let message):
            return "Decryption failed: \(message)"
        case .fileTooShort:
            return "Encrypted file is too short or corrupted"
        case .fileReadFailed(let message):
            return "Failed to read file: \(message)"
        case .fileWriteFailed(let message):
            return "Failed to write file: \(message)"
        case .keyDerivationFailed:
            return "Failed to derive encryption key"
        }
    }
}

// MARK: - Database Encryption Service

actor DatabaseEncryptionService {
    static let shared = DatabaseEncryptionService()

    private init() {}

    // MARK: - Public API

    /// Encrypt a database file using AES-256-GCM with password-derived key
    /// Compatible with Go CLI encryption format
    func encryptDatabase(inputPath: String, outputPath: String, password: String) throws {
        // Generate random salt
        var salt = Data(count: EncryptionConstants.saltSize)
        let saltResult = salt.withUnsafeMutableBytes { buffer in
            SecRandomCopyBytes(kSecRandomDefault, EncryptionConstants.saltSize, buffer.baseAddress!)
        }
        guard saltResult == errSecSuccess else {
            throw EncryptionError.encryptionFailed("Failed to generate salt")
        }

        // Derive key from password using PBKDF2
        guard let key = deriveKey(password: password, salt: salt) else {
            throw EncryptionError.keyDerivationFailed
        }

        // Read input file
        guard let plaintext = FileManager.default.contents(atPath: inputPath) else {
            throw EncryptionError.fileReadFailed(inputPath)
        }

        // Generate nonce
        var nonce = Data(count: EncryptionConstants.nonceSize)
        let nonceResult = nonce.withUnsafeMutableBytes { buffer in
            SecRandomCopyBytes(kSecRandomDefault, EncryptionConstants.nonceSize, buffer.baseAddress!)
        }
        guard nonceResult == errSecSuccess else {
            throw EncryptionError.encryptionFailed("Failed to generate nonce")
        }

        // Encrypt using AES-256-GCM
        let symmetricKey = SymmetricKey(data: key)
        let gcmNonce = try AES.GCM.Nonce(data: nonce)
        let sealedBox = try AES.GCM.seal(plaintext, using: symmetricKey, nonce: gcmNonce)

        // Combine nonce + ciphertext + tag (this is what sealedBox.combined gives us)
        guard let combined = sealedBox.combined else {
            throw EncryptionError.encryptionFailed("Failed to get combined ciphertext")
        }

        // Write encrypted file: [salt][nonce + ciphertext + tag]
        var output = Data()
        output.append(salt)
        output.append(combined)

        // Write with restricted permissions (0600)
        let fileURL = URL(fileURLWithPath: outputPath)
        try output.write(to: fileURL, options: .atomic)

        // Set file permissions to 0600
        try FileManager.default.setAttributes(
            [.posixPermissions: 0o600],
            ofItemAtPath: outputPath
        )
    }

    /// Decrypt a database file encrypted with encryptDatabase
    /// Compatible with Go CLI encryption format
    func decryptDatabase(inputPath: String, outputPath: String, password: String) throws {
        // Read encrypted file
        guard let encrypted = FileManager.default.contents(atPath: inputPath) else {
            throw EncryptionError.fileReadFailed(inputPath)
        }

        // Verify minimum size: salt + nonce + at least some ciphertext
        let minimumSize = EncryptionConstants.saltSize + EncryptionConstants.nonceSize + 16  // 16 = GCM tag size
        guard encrypted.count >= minimumSize else {
            throw EncryptionError.fileTooShort
        }

        // Extract salt
        let salt = encrypted.prefix(EncryptionConstants.saltSize)
        let ciphertextWithNonceAndTag = encrypted.suffix(from: EncryptionConstants.saltSize)

        // Derive key from password
        guard let key = deriveKey(password: password, salt: salt) else {
            throw EncryptionError.keyDerivationFailed
        }

        // Decrypt using AES-256-GCM
        let symmetricKey = SymmetricKey(data: key)
        do {
            let sealedBox = try AES.GCM.SealedBox(combined: ciphertextWithNonceAndTag)
            let plaintext = try AES.GCM.open(sealedBox, using: symmetricKey)

            // Write decrypted file
            let fileURL = URL(fileURLWithPath: outputPath)
            try plaintext.write(to: fileURL, options: .atomic)

            // Set file permissions to 0600
            try FileManager.default.setAttributes(
                [.posixPermissions: 0o600],
                ofItemAtPath: outputPath
            )
        } catch {
            throw EncryptionError.decryptionFailed("Wrong password or corrupted file")
        }
    }

    /// Check if a file appears to be encrypted
    /// Uses heuristics similar to the Go CLI implementation
    func isEncrypted(path: String) -> Bool {
        guard let attributes = try? FileManager.default.attributesOfItem(atPath: path),
              let fileSize = attributes[.size] as? Int64
        else {
            return false
        }

        // Encrypted files should be at least saltSize + nonceSize bytes
        let minimumSize = EncryptionConstants.saltSize + EncryptionConstants.nonceSize
        guard fileSize >= Int64(minimumSize) else {
            return false
        }

        // Read first saltSize bytes to check if they look like random data
        guard let fileHandle = FileHandle(forReadingAtPath: path) else {
            return false
        }
        defer { try? fileHandle.close() }

        guard let saltData = try? fileHandle.read(upToCount: EncryptionConstants.saltSize),
              saltData.count == EncryptionConstants.saltSize
        else {
            return false
        }

        // Check if first saltSize bytes look like random data
        // by checking they're not all zeros and not all printable ASCII
        var allZeros = true
        var allPrintable = true

        for byte in saltData {
            if byte != 0 {
                allZeros = false
            }
            if byte < 32 || byte > 126 {
                allPrintable = false
            }
        }

        // If all zeros or all printable ASCII, probably not encrypted
        return !allZeros && !allPrintable
    }

    /// Get the encrypted database path for a given database path
    func encryptedPath(for databasePath: String) -> String {
        return databasePath + EncryptionConstants.encryptedFileExtension
    }

    /// Get the decrypted (working) database path
    func decryptedPath(for databasePath: String) -> String {
        return databasePath + ".decrypted"
    }

    // MARK: - Private Methods

    /// Derive a key from password using PBKDF2 with SHA256
    /// Matches the Go implementation using golang.org/x/crypto/pbkdf2
    private func deriveKey(password: String, salt: Data) -> Data? {
        guard let passwordData = password.data(using: .utf8) else {
            return nil
        }

        var derivedKey = Data(count: EncryptionConstants.keySize)

        let result = derivedKey.withUnsafeMutableBytes { derivedKeyBuffer in
            salt.withUnsafeBytes { saltBuffer in
                passwordData.withUnsafeBytes { passwordBuffer in
                    CCKeyDerivationPBKDF(
                        CCPBKDFAlgorithm(kCCPBKDF2),
                        passwordBuffer.baseAddress?.assumingMemoryBound(to: Int8.self),
                        passwordData.count,
                        saltBuffer.baseAddress?.assumingMemoryBound(to: UInt8.self),
                        salt.count,
                        CCPseudoRandomAlgorithm(kCCPRFHmacAlgSHA256),
                        EncryptionConstants.pbkdf2Iterations,
                        derivedKeyBuffer.baseAddress?.assumingMemoryBound(to: UInt8.self),
                        EncryptionConstants.keySize
                    )
                }
            }
        }

        return result == kCCSuccess ? derivedKey : nil
    }
}

// MARK: - Encryption Status

enum EncryptionStatus: Equatable {
    case disabled
    case enabled
    case checking
    case error(String)

    var isEnabled: Bool {
        if case .enabled = self { return true }
        return false
    }

    var description: String {
        switch self {
        case .disabled:
            return "Database is not encrypted"
        case .enabled:
            return "Database is encrypted"
        case .checking:
            return "Checking encryption status..."
        case .error(let message):
            return "Error: \(message)"
        }
    }
}
