package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptDecryptDatabase(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "test.db")
	encryptedPath := filepath.Join(tmpDir, "test.db.encrypted")
	decryptedPath := filepath.Join(tmpDir, "test.db.decrypted")

	// Create test database file
	testData := []byte("This is test database content with sensitive data")
	if err := os.WriteFile(inputPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	password := "SecureTestPassword123!"

	// Test encryption
	t.Run("Encrypt database", func(t *testing.T) {
		err := EncryptDatabase(inputPath, encryptedPath, password)
		if err != nil {
			t.Fatalf("Encryption failed: %v", err)
		}

		// Verify encrypted file exists
		if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
			t.Fatal("Encrypted file was not created")
		}

		// Verify encrypted file is different from original
		encryptedData, err := os.ReadFile(encryptedPath)
		if err != nil {
			t.Fatalf("Failed to read encrypted file: %v", err)
		}

		if string(encryptedData) == string(testData) {
			t.Fatal("Encrypted data should not match original data")
		}

		// Verify encrypted file is larger (salt + nonce + ciphertext)
		if len(encryptedData) <= len(testData) {
			t.Fatal("Encrypted file should be larger than original")
		}
	})

	// Test decryption
	t.Run("Decrypt database", func(t *testing.T) {
		err := DecryptDatabase(encryptedPath, decryptedPath, password)
		if err != nil {
			t.Fatalf("Decryption failed: %v", err)
		}

		// Verify decrypted file exists
		if _, err := os.Stat(decryptedPath); os.IsNotExist(err) {
			t.Fatal("Decrypted file was not created")
		}

		// Verify decrypted data matches original
		decryptedData, err := os.ReadFile(decryptedPath)
		if err != nil {
			t.Fatalf("Failed to read decrypted file: %v", err)
		}

		if string(decryptedData) != string(testData) {
			t.Fatalf("Decrypted data does not match original.\nExpected: %s\nGot: %s",
				string(testData), string(decryptedData))
		}
	})

	// Test wrong password
	t.Run("Decrypt with wrong password", func(t *testing.T) {
		wrongPath := filepath.Join(tmpDir, "wrong.db")
		err := DecryptDatabase(encryptedPath, wrongPath, "WrongPassword123!")

		if err == nil {
			t.Fatal("Expected decryption to fail with wrong password")
		}

		// Verify error message mentions decryption failure
		if !strings.Contains(err.Error(), "decryption failed") {
			t.Fatalf("Expected 'decryption failed' error, got: %v", err)
		}
	})

	// Test encrypting large file
	t.Run("Encrypt large database", func(t *testing.T) {
		largePath := filepath.Join(tmpDir, "large.db")
		largeEncPath := filepath.Join(tmpDir, "large.db.encrypted")
		largeDecPath := filepath.Join(tmpDir, "large.db.decrypted")

		// Create 1MB file
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		if err := os.WriteFile(largePath, largeData, 0600); err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}

		// Encrypt
		if err := EncryptDatabase(largePath, largeEncPath, password); err != nil {
			t.Fatalf("Failed to encrypt large file: %v", err)
		}

		// Decrypt
		if err := DecryptDatabase(largeEncPath, largeDecPath, password); err != nil {
			t.Fatalf("Failed to decrypt large file: %v", err)
		}

		// Verify data integrity
		decryptedLargeData, err := os.ReadFile(largeDecPath)
		if err != nil {
			t.Fatalf("Failed to read decrypted large file: %v", err)
		}

		if len(decryptedLargeData) != len(largeData) {
			t.Fatalf("Decrypted data size mismatch. Expected: %d, Got: %d",
				len(largeData), len(decryptedLargeData))
		}

		// Check a few bytes
		for i := 0; i < 100; i++ {
			if decryptedLargeData[i] != largeData[i] {
				t.Fatalf("Data mismatch at byte %d", i)
			}
		}
	})

	// Test IsEncrypted
	t.Run("IsEncrypted check", func(t *testing.T) {
		if !IsEncrypted(encryptedPath) {
			t.Error("Encrypted file should be detected as encrypted")
		}

		if IsEncrypted(inputPath) {
			t.Error("Plain file should not be detected as encrypted")
		}

		nonexistentPath := filepath.Join(tmpDir, "nonexistent.db")
		if IsEncrypted(nonexistentPath) {
			t.Error("Nonexistent file should not be detected as encrypted")
		}
	})
}

func TestEncryptDatabaseErrors(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Encrypt nonexistent file", func(t *testing.T) {
		err := EncryptDatabase(
			filepath.Join(tmpDir, "nonexistent.db"),
			filepath.Join(tmpDir, "out.encrypted"),
			"password",
		)

		if err == nil {
			t.Fatal("Expected error when encrypting nonexistent file")
		}
	})

	t.Run("Decrypt invalid file", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.db")
		if err := os.WriteFile(invalidPath, []byte("short"), 0600); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		err := DecryptDatabase(
			invalidPath,
			filepath.Join(tmpDir, "out.db"),
			"password",
		)

		if err == nil {
			t.Fatal("Expected error when decrypting invalid file")
		}
	})

	t.Run("Encrypt to readonly directory", func(t *testing.T) {
		// Skip if running as root (root can write to readonly directories)
		if os.Geteuid() == 0 {
			t.Skip("Skipping readonly directory test when running as root")
		}

		readonlyDir := filepath.Join(tmpDir, "readonly")
		if err := os.MkdirAll(readonlyDir, 0400); err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}

		inputPath := filepath.Join(tmpDir, "test.db")
		if err := os.WriteFile(inputPath, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := EncryptDatabase(
			inputPath,
			filepath.Join(readonlyDir, "out.encrypted"),
			"password",
		)

		if err == nil {
			t.Fatal("Expected error when writing to readonly directory")
		}
	})
}
