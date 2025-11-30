//
//  DatabaseEncryptionTests.swift
//  ungTests
//
//  Security tests for AES-256-GCM database encryption.
//  Tests encryption/decryption, cross-platform compatibility, and error handling.
//

import XCTest
@testable import ung

final class DatabaseEncryptionTests: XCTestCase {

    var tempDirectory: URL!
    let testPassword = "SecureTestPassword123!"

    override func setUpWithError() throws {
        // Create temp directory for each test
        tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent("ungEncryptionTests-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: tempDirectory, withIntermediateDirectories: true)
    }

    override func tearDownWithError() throws {
        // Clean up temp directory
        if let tempDir = tempDirectory {
            try? FileManager.default.removeItem(at: tempDir)
        }
    }

    // MARK: - Basic Encryption/Decryption Tests

    func testEncryptDecryptRoundTrip() async throws {
        let inputPath = tempDirectory.appendingPathComponent("test.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("test.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("test.db.decrypted").path

        // Create test data
        let testData = "This is test database content with sensitive data".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared

        // Encrypt
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Verify encrypted file exists
        XCTAssertTrue(FileManager.default.fileExists(atPath: encryptedPath))

        // Decrypt
        try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: testPassword)

        // Verify decrypted data matches original
        let decryptedData = try Data(contentsOf: URL(fileURLWithPath: decryptedPath))
        XCTAssertEqual(decryptedData, testData)
    }

    func testEncryptedDataDiffersFromOriginal() async throws {
        let inputPath = tempDirectory.appendingPathComponent("original.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("original.db.encrypted").path

        let testData = "Sensitive information that should be encrypted".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        let encryptedData = try Data(contentsOf: URL(fileURLWithPath: encryptedPath))

        // Encrypted data should be different
        XCTAssertNotEqual(encryptedData, testData)

        // Encrypted file should be larger (salt + nonce + tag overhead)
        XCTAssertGreaterThan(encryptedData.count, testData.count)
    }

    // MARK: - Wrong Password Tests

    func testDecryptWithWrongPassword() async throws {
        let inputPath = tempDirectory.appendingPathComponent("secure.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("secure.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("secure.db.wrong").path

        let testData = "Secret data".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Try decrypting with wrong password
        do {
            try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: "WrongPassword123!")
            XCTFail("Should have thrown error with wrong password")
        } catch let error as EncryptionError {
            if case .decryptionFailed = error {
                // Expected error
            } else {
                XCTFail("Expected decryptionFailed error, got: \(error)")
            }
        }
    }

    func testDecryptWithEmptyPassword() async throws {
        let inputPath = tempDirectory.appendingPathComponent("data.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("data.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("data.db.decrypted").path

        let testData = "Test data".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Try decrypting with empty password
        do {
            try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: "")
            XCTFail("Should have thrown error with empty password")
        } catch {
            // Expected - either decryption failed or key derivation failed
        }
    }

    // MARK: - Large File Tests

    func testEncryptLargeFile() async throws {
        let inputPath = tempDirectory.appendingPathComponent("large.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("large.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("large.db.decrypted").path

        // Create 1MB file
        var largeData = Data(count: 1024 * 1024)
        for i in 0..<largeData.count {
            largeData[i] = UInt8(i % 256)
        }
        try largeData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared

        // Encrypt
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Decrypt
        try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: testPassword)

        // Verify integrity
        let decryptedData = try Data(contentsOf: URL(fileURLWithPath: decryptedPath))
        XCTAssertEqual(decryptedData.count, largeData.count)
        XCTAssertEqual(decryptedData, largeData)
    }

    // MARK: - isEncrypted Detection Tests

    func testIsEncryptedDetectsEncryptedFile() async throws {
        let inputPath = tempDirectory.appendingPathComponent("detect.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("detect.db.encrypted").path

        let testData = "Test content".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        let isEncrypted = await service.isEncrypted(path: encryptedPath)
        XCTAssertTrue(isEncrypted, "Encrypted file should be detected as encrypted")
    }

    func testIsEncryptedRejectsPlainFile() async throws {
        let plainPath = tempDirectory.appendingPathComponent("plain.db").path

        // Create a plain text file
        let plainData = "This is plain text content that is not encrypted".data(using: .utf8)!
        try plainData.write(to: URL(fileURLWithPath: plainPath))

        let service = DatabaseEncryptionService.shared
        let isEncrypted = await service.isEncrypted(path: plainPath)
        XCTAssertFalse(isEncrypted, "Plain file should not be detected as encrypted")
    }

    func testIsEncryptedRejectsNonexistentFile() async {
        let nonexistentPath = tempDirectory.appendingPathComponent("nonexistent.db").path

        let service = DatabaseEncryptionService.shared
        let isEncrypted = await service.isEncrypted(path: nonexistentPath)
        XCTAssertFalse(isEncrypted, "Nonexistent file should not be detected as encrypted")
    }

    func testIsEncryptedRejectsShortFile() async throws {
        let shortPath = tempDirectory.appendingPathComponent("short.db").path

        // Create a file that's too short to be encrypted
        let shortData = Data([0x01, 0x02, 0x03, 0x04, 0x05])
        try shortData.write(to: URL(fileURLWithPath: shortPath))

        let service = DatabaseEncryptionService.shared
        let isEncrypted = await service.isEncrypted(path: shortPath)
        XCTAssertFalse(isEncrypted, "Short file should not be detected as encrypted")
    }

    func testIsEncryptedRejectsSQLiteFile() async throws {
        let sqlitePath = tempDirectory.appendingPathComponent("sqlite.db").path

        // Create a file with SQLite header
        var sqliteData = "SQLite format 3\0".data(using: .utf8)!
        // Pad to make it look like a real file
        sqliteData.append(Data(repeating: 0, count: 100))
        try sqliteData.write(to: URL(fileURLWithPath: sqlitePath))

        let service = DatabaseEncryptionService.shared
        let isEncrypted = await service.isEncrypted(path: sqlitePath)
        XCTAssertFalse(isEncrypted, "SQLite file should not be detected as encrypted")
    }

    // MARK: - Error Handling Tests

    func testEncryptNonexistentFile() async {
        let nonexistentPath = tempDirectory.appendingPathComponent("missing.db").path
        let outputPath = tempDirectory.appendingPathComponent("output.encrypted").path

        let service = DatabaseEncryptionService.shared

        do {
            try await service.encryptDatabase(inputPath: nonexistentPath, outputPath: outputPath, password: testPassword)
            XCTFail("Should have thrown error for nonexistent file")
        } catch let error as EncryptionError {
            if case .fileReadFailed = error {
                // Expected
            } else {
                XCTFail("Expected fileReadFailed error, got: \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    func testDecryptInvalidFile() async throws {
        let invalidPath = tempDirectory.appendingPathComponent("invalid.encrypted").path
        let outputPath = tempDirectory.appendingPathComponent("output.db").path

        // Create a file with invalid encrypted content (too short)
        let invalidData = Data([0x01, 0x02, 0x03])
        try invalidData.write(to: URL(fileURLWithPath: invalidPath))

        let service = DatabaseEncryptionService.shared

        do {
            try await service.decryptDatabase(inputPath: invalidPath, outputPath: outputPath, password: testPassword)
            XCTFail("Should have thrown error for invalid file")
        } catch let error as EncryptionError {
            if case .fileTooShort = error {
                // Expected
            } else {
                XCTFail("Expected fileTooShort error, got: \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    func testDecryptCorruptedFile() async throws {
        let inputPath = tempDirectory.appendingPathComponent("valid.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("valid.db.encrypted").path
        let corruptedPath = tempDirectory.appendingPathComponent("corrupted.encrypted").path
        let outputPath = tempDirectory.appendingPathComponent("output.db").path

        let testData = "Valid test data".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Corrupt the encrypted file by modifying some bytes
        var encryptedData = try Data(contentsOf: URL(fileURLWithPath: encryptedPath))
        if encryptedData.count > 50 {
            encryptedData[50] ^= 0xFF // Flip bits
        }
        try encryptedData.write(to: URL(fileURLWithPath: corruptedPath))

        do {
            try await service.decryptDatabase(inputPath: corruptedPath, outputPath: outputPath, password: testPassword)
            XCTFail("Should have thrown error for corrupted file")
        } catch let error as EncryptionError {
            if case .decryptionFailed = error {
                // Expected - GCM authentication should fail
            } else {
                XCTFail("Expected decryptionFailed error, got: \(error)")
            }
        } catch {
            // CryptoKit may throw its own error, which is also acceptable
        }
    }

    // MARK: - File Permission Tests

    func testEncryptedFileHasRestrictedPermissions() async throws {
        let inputPath = tempDirectory.appendingPathComponent("perms.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("perms.db.encrypted").path

        let testData = "Sensitive data".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Check file permissions
        let attributes = try FileManager.default.attributesOfItem(atPath: encryptedPath)
        if let permissions = attributes[.posixPermissions] as? Int {
            // 0600 = owner read/write only
            XCTAssertEqual(permissions, 0o600, "Encrypted file should have 0600 permissions")
        }
    }

    // MARK: - Path Utilities Tests

    func testEncryptedPathGeneration() async {
        let service = DatabaseEncryptionService.shared
        let originalPath = "/path/to/database.db"
        let encrypted = await service.encryptedPath(for: originalPath)
        XCTAssertEqual(encrypted, "/path/to/database.db.encrypted")
    }

    func testDecryptedPathGeneration() async {
        let service = DatabaseEncryptionService.shared
        let originalPath = "/path/to/database.db"
        let decrypted = await service.decryptedPath(for: originalPath)
        XCTAssertEqual(decrypted, "/path/to/database.db.decrypted")
    }

    // MARK: - Encryption Constants Tests

    func testEncryptionConstantsMatch() {
        // Verify constants match Go CLI implementation
        XCTAssertEqual(EncryptionConstants.keySize, 32, "AES-256 requires 32-byte key")
        XCTAssertEqual(EncryptionConstants.saltSize, 32, "Salt should be 32 bytes")
        XCTAssertEqual(EncryptionConstants.nonceSize, 12, "GCM nonce should be 12 bytes")
        XCTAssertEqual(EncryptionConstants.pbkdf2Iterations, 100_000, "PBKDF2 should use 100,000 iterations")
    }

    // MARK: - Empty File Tests

    func testEncryptEmptyFile() async throws {
        let inputPath = tempDirectory.appendingPathComponent("empty.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("empty.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("empty.db.decrypted").path

        // Create empty file
        try Data().write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared

        // Encrypt empty file
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)

        // Decrypt
        try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: testPassword)

        // Verify decrypted file is also empty
        let decryptedData = try Data(contentsOf: URL(fileURLWithPath: decryptedPath))
        XCTAssertEqual(decryptedData.count, 0)
    }

    // MARK: - Password Variations Tests

    func testPasswordWithSpecialCharacters() async throws {
        let inputPath = tempDirectory.appendingPathComponent("special.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("special.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("special.db.decrypted").path

        let testData = "Test data".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let specialPassword = "P@$$w0rd!#$%^&*()_+-=[]{}|;':\",./<>?"
        let service = DatabaseEncryptionService.shared

        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: specialPassword)
        try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: specialPassword)

        let decryptedData = try Data(contentsOf: URL(fileURLWithPath: decryptedPath))
        XCTAssertEqual(decryptedData, testData)
    }

    func testPasswordWithUnicode() async throws {
        let inputPath = tempDirectory.appendingPathComponent("unicode.db").path
        let encryptedPath = tempDirectory.appendingPathComponent("unicode.db.encrypted").path
        let decryptedPath = tempDirectory.appendingPathComponent("unicode.db.decrypted").path

        let testData = "Test data with unicode".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let unicodePassword = "пароль密码パスワード🔐"
        let service = DatabaseEncryptionService.shared

        try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: unicodePassword)
        try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: unicodePassword)

        let decryptedData = try Data(contentsOf: URL(fileURLWithPath: decryptedPath))
        XCTAssertEqual(decryptedData, testData)
    }

    // MARK: - Concurrent Access Tests

    func testConcurrentEncryption() async throws {
        let service = DatabaseEncryptionService.shared

        // Create multiple files
        var tasks: [Task<Void, Error>] = []

        for i in 0..<5 {
            let inputPath = tempDirectory.appendingPathComponent("concurrent\(i).db").path
            let encryptedPath = tempDirectory.appendingPathComponent("concurrent\(i).db.encrypted").path
            let decryptedPath = tempDirectory.appendingPathComponent("concurrent\(i).db.decrypted").path

            let testData = "Test data for file \(i)".data(using: .utf8)!
            try testData.write(to: URL(fileURLWithPath: inputPath))

            let task = Task {
                try await service.encryptDatabase(inputPath: inputPath, outputPath: encryptedPath, password: testPassword)
                try await service.decryptDatabase(inputPath: encryptedPath, outputPath: decryptedPath, password: testPassword)

                let decryptedData = try Data(contentsOf: URL(fileURLWithPath: decryptedPath))
                XCTAssertEqual(decryptedData, testData)
            }
            tasks.append(task)
        }

        // Wait for all tasks
        for task in tasks {
            try await task.value
        }
    }

    // MARK: - Salt Uniqueness Tests

    func testSaltIsUniquePerEncryption() async throws {
        let inputPath = tempDirectory.appendingPathComponent("salt.db").path
        let encrypted1Path = tempDirectory.appendingPathComponent("salt1.encrypted").path
        let encrypted2Path = tempDirectory.appendingPathComponent("salt2.encrypted").path

        let testData = "Same data encrypted twice".data(using: .utf8)!
        try testData.write(to: URL(fileURLWithPath: inputPath))

        let service = DatabaseEncryptionService.shared

        try await service.encryptDatabase(inputPath: inputPath, outputPath: encrypted1Path, password: testPassword)
        try await service.encryptDatabase(inputPath: inputPath, outputPath: encrypted2Path, password: testPassword)

        let encrypted1 = try Data(contentsOf: URL(fileURLWithPath: encrypted1Path))
        let encrypted2 = try Data(contentsOf: URL(fileURLWithPath: encrypted2Path))

        // Extract salts (first 32 bytes)
        let salt1 = encrypted1.prefix(32)
        let salt2 = encrypted2.prefix(32)

        // Salts should be different (random)
        XCTAssertNotEqual(salt1, salt2, "Each encryption should use a unique salt")

        // The entire ciphertext should also be different
        XCTAssertNotEqual(encrypted1, encrypted2, "Encrypted data should differ due to random salt/nonce")
    }
}

// MARK: - EncryptionStatus Tests

final class EncryptionStatusTests: XCTestCase {

    func testDisabledStatus() {
        let status = EncryptionStatus.disabled
        XCTAssertFalse(status.isEnabled)
        XCTAssertEqual(status.description, "Database is not encrypted")
    }

    func testEnabledStatus() {
        let status = EncryptionStatus.enabled
        XCTAssertTrue(status.isEnabled)
        XCTAssertEqual(status.description, "Database is encrypted")
    }

    func testCheckingStatus() {
        let status = EncryptionStatus.checking
        XCTAssertFalse(status.isEnabled)
        XCTAssertEqual(status.description, "Checking encryption status...")
    }

    func testErrorStatus() {
        let status = EncryptionStatus.error("Test error message")
        XCTAssertFalse(status.isEnabled)
        XCTAssertEqual(status.description, "Error: Test error message")
    }

    func testStatusEquality() {
        XCTAssertEqual(EncryptionStatus.disabled, EncryptionStatus.disabled)
        XCTAssertEqual(EncryptionStatus.enabled, EncryptionStatus.enabled)
        XCTAssertNotEqual(EncryptionStatus.disabled, EncryptionStatus.enabled)
        XCTAssertEqual(EncryptionStatus.error("msg"), EncryptionStatus.error("msg"))
        XCTAssertNotEqual(EncryptionStatus.error("msg1"), EncryptionStatus.error("msg2"))
    }
}
