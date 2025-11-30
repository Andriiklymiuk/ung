package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptDecryptFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "test.db")
	encryptedPath := filepath.Join(tmpDir, "test.db.encrypted")
	decryptedPath := filepath.Join(tmpDir, "test.db.decrypted")

	// Create test data
	testData := []byte("This is test database content with sensitive data")
	if err := os.WriteFile(inputPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	password := "SecureTestPassword123!"

	// Test encryption
	t.Run("Encrypt file", func(t *testing.T) {
		err := EncryptFile(inputPath, encryptedPath, password)
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

		if bytes.Equal(encryptedData, testData) {
			t.Fatal("Encrypted data should not match original data")
		}

		// Verify encrypted file is larger (salt + nonce + ciphertext + tag)
		if len(encryptedData) <= len(testData) {
			t.Fatal("Encrypted file should be larger than original")
		}
	})

	// Test decryption
	t.Run("Decrypt file", func(t *testing.T) {
		err := DecryptFile(encryptedPath, decryptedPath, password)
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

		if !bytes.Equal(decryptedData, testData) {
			t.Fatalf("Decrypted data does not match original.\nExpected: %s\nGot: %s",
				string(testData), string(decryptedData))
		}
	})

	// Test wrong password
	t.Run("Decrypt with wrong password", func(t *testing.T) {
		wrongPath := filepath.Join(tmpDir, "wrong.db")
		err := DecryptFile(encryptedPath, wrongPath, "WrongPassword123!")

		if err == nil {
			t.Fatal("Expected decryption to fail with wrong password")
		}

		if !strings.Contains(err.Error(), "decryption failed") {
			t.Fatalf("Expected 'decryption failed' error, got: %v", err)
		}
	})

	// Test large file
	t.Run("Encrypt large file", func(t *testing.T) {
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
		if err := EncryptFile(largePath, largeEncPath, password); err != nil {
			t.Fatalf("Failed to encrypt large file: %v", err)
		}

		// Decrypt
		if err := DecryptFile(largeEncPath, largeDecPath, password); err != nil {
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

		if !bytes.Equal(decryptedLargeData, largeData) {
			t.Fatal("Large file data mismatch after decryption")
		}
	})
}

func TestEncryptDecryptStream(t *testing.T) {
	password := "StreamTestPassword123!"

	t.Run("Stream encrypt/decrypt round trip", func(t *testing.T) {
		testData := []byte("Stream encryption test data with multiple lines\nLine 2\nLine 3")

		// Encrypt
		var encrypted bytes.Buffer
		err := EncryptStream(bytes.NewReader(testData), &encrypted, password)
		if err != nil {
			t.Fatalf("Stream encryption failed: %v", err)
		}

		// Encrypted data should be different and larger
		if bytes.Equal(encrypted.Bytes(), testData) {
			t.Fatal("Encrypted stream data should differ from original")
		}

		if encrypted.Len() <= len(testData) {
			t.Fatal("Encrypted stream should be larger than original")
		}

		// Decrypt
		var decrypted bytes.Buffer
		err = DecryptStream(bytes.NewReader(encrypted.Bytes()), &decrypted, password)
		if err != nil {
			t.Fatalf("Stream decryption failed: %v", err)
		}

		// Verify match
		if !bytes.Equal(decrypted.Bytes(), testData) {
			t.Fatalf("Decrypted stream does not match original.\nExpected: %s\nGot: %s",
				string(testData), decrypted.String())
		}
	})

	t.Run("Stream decrypt with wrong password", func(t *testing.T) {
		testData := []byte("Secret stream data")

		var encrypted bytes.Buffer
		if err := EncryptStream(bytes.NewReader(testData), &encrypted, password); err != nil {
			t.Fatalf("Stream encryption failed: %v", err)
		}

		var decrypted bytes.Buffer
		err := DecryptStream(bytes.NewReader(encrypted.Bytes()), &decrypted, "WrongPassword!")

		if err == nil {
			t.Fatal("Expected decryption to fail with wrong password")
		}
	})

	t.Run("Stream empty data", func(t *testing.T) {
		testData := []byte{}

		var encrypted bytes.Buffer
		if err := EncryptStream(bytes.NewReader(testData), &encrypted, password); err != nil {
			t.Fatalf("Stream encryption of empty data failed: %v", err)
		}

		var decrypted bytes.Buffer
		if err := DecryptStream(bytes.NewReader(encrypted.Bytes()), &decrypted, password); err != nil {
			t.Fatalf("Stream decryption of empty data failed: %v", err)
		}

		if decrypted.Len() != 0 {
			t.Fatalf("Decrypted empty stream should be empty, got %d bytes", decrypted.Len())
		}
	})
}

func TestFileErrors(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Encrypt nonexistent file", func(t *testing.T) {
		err := EncryptFile(
			filepath.Join(tmpDir, "nonexistent.db"),
			filepath.Join(tmpDir, "out.encrypted"),
			"password",
		)

		if err == nil {
			t.Fatal("Expected error when encrypting nonexistent file")
		}
	})

	t.Run("Decrypt file too short", func(t *testing.T) {
		shortPath := filepath.Join(tmpDir, "short.db")
		if err := os.WriteFile(shortPath, []byte("short"), 0600); err != nil {
			t.Fatalf("Failed to create short file: %v", err)
		}

		err := DecryptFile(
			shortPath,
			filepath.Join(tmpDir, "out.db"),
			"password",
		)

		if err == nil {
			t.Fatal("Expected error when decrypting file too short")
		}

		if !strings.Contains(err.Error(), "too short") {
			t.Fatalf("Expected 'too short' error, got: %v", err)
		}
	})

	t.Run("Decrypt corrupted file", func(t *testing.T) {
		inputPath := filepath.Join(tmpDir, "valid.db")
		encryptedPath := filepath.Join(tmpDir, "valid.encrypted")
		corruptedPath := filepath.Join(tmpDir, "corrupted.encrypted")

		testData := []byte("Valid test data for corruption test")
		if err := os.WriteFile(inputPath, testData, 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		password := "CorruptionTest123!"
		if err := EncryptFile(inputPath, encryptedPath, password); err != nil {
			t.Fatalf("Encryption failed: %v", err)
		}

		// Corrupt the file
		encryptedData, _ := os.ReadFile(encryptedPath)
		if len(encryptedData) > 50 {
			encryptedData[50] ^= 0xFF // Flip bits
		}
		if err := os.WriteFile(corruptedPath, encryptedData, 0600); err != nil {
			t.Fatalf("Failed to write corrupted file: %v", err)
		}

		err := DecryptFile(
			corruptedPath,
			filepath.Join(tmpDir, "out.db"),
			password,
		)

		if err == nil {
			t.Fatal("Expected error when decrypting corrupted file")
		}
	})

	t.Run("Encrypt to readonly directory", func(t *testing.T) {
		// Skip if running as root
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

		err := EncryptFile(
			inputPath,
			filepath.Join(readonlyDir, "out.encrypted"),
			"password",
		)

		if err == nil {
			t.Fatal("Expected error when writing to readonly directory")
		}
	})
}

func TestGenerateSalt(t *testing.T) {
	t.Run("Salt has correct size", func(t *testing.T) {
		salt, err := GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}

		if len(salt) != saltSize {
			t.Fatalf("Salt should be %d bytes, got %d", saltSize, len(salt))
		}
	})

	t.Run("Salts are unique", func(t *testing.T) {
		salt1, err := GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}

		salt2, err := GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}

		if bytes.Equal(salt1, salt2) {
			t.Fatal("Generated salts should be unique")
		}
	})

	t.Run("Salt is not all zeros", func(t *testing.T) {
		salt, err := GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}

		allZeros := true
		for _, b := range salt {
			if b != 0 {
				allZeros = false
				break
			}
		}

		if allZeros {
			t.Fatal("Salt should not be all zeros")
		}
	})
}

func TestDeriveKey(t *testing.T) {
	t.Run("Key has correct size", func(t *testing.T) {
		salt, _ := GenerateSalt()
		key := DeriveKey("password", salt)

		if len(key) != keySize {
			t.Fatalf("Key should be %d bytes, got %d", keySize, len(key))
		}
	})

	t.Run("Same password and salt produce same key", func(t *testing.T) {
		salt, _ := GenerateSalt()
		password := "consistent-password"

		key1 := DeriveKey(password, salt)
		key2 := DeriveKey(password, salt)

		if !bytes.Equal(key1, key2) {
			t.Fatal("Same password and salt should produce same key")
		}
	})

	t.Run("Different passwords produce different keys", func(t *testing.T) {
		salt, _ := GenerateSalt()

		key1 := DeriveKey("password1", salt)
		key2 := DeriveKey("password2", salt)

		if bytes.Equal(key1, key2) {
			t.Fatal("Different passwords should produce different keys")
		}
	})

	t.Run("Different salts produce different keys", func(t *testing.T) {
		salt1, _ := GenerateSalt()
		salt2, _ := GenerateSalt()
		password := "same-password"

		key1 := DeriveKey(password, salt1)
		key2 := DeriveKey(password, salt2)

		if bytes.Equal(key1, key2) {
			t.Fatal("Different salts should produce different keys")
		}
	})

	t.Run("Empty password works", func(t *testing.T) {
		salt, _ := GenerateSalt()
		key := DeriveKey("", salt)

		if len(key) != keySize {
			t.Fatalf("Key should be %d bytes even with empty password, got %d", keySize, len(key))
		}
	})

	t.Run("Unicode password works", func(t *testing.T) {
		salt, _ := GenerateSalt()
		key := DeriveKey("пароль密码パスワード🔐", salt)

		if len(key) != keySize {
			t.Fatalf("Key should be %d bytes with unicode password, got %d", keySize, len(key))
		}
	})
}

func TestFilePermissions(t *testing.T) {
	// Skip if running as root
	if os.Geteuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "perms.db")
	encryptedPath := filepath.Join(tmpDir, "perms.encrypted")

	testData := []byte("Permission test data")
	if err := os.WriteFile(inputPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := EncryptFile(inputPath, encryptedPath, "password"); err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	info, err := os.Stat(encryptedPath)
	if err != nil {
		t.Fatalf("Failed to stat encrypted file: %v", err)
	}

	// Check permissions (0600)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Encrypted file should have 0600 permissions, got %o", perm)
	}
}

func TestSaltUniqueness(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "salt.db")
	encrypted1Path := filepath.Join(tmpDir, "salt1.encrypted")
	encrypted2Path := filepath.Join(tmpDir, "salt2.encrypted")

	testData := []byte("Same data encrypted twice")
	if err := os.WriteFile(inputPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	password := "SamePassword123!"

	if err := EncryptFile(inputPath, encrypted1Path, password); err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	if err := EncryptFile(inputPath, encrypted2Path, password); err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	encrypted1, _ := os.ReadFile(encrypted1Path)
	encrypted2, _ := os.ReadFile(encrypted2Path)

	// Extract salts (first 32 bytes)
	salt1 := encrypted1[:saltSize]
	salt2 := encrypted2[:saltSize]

	if bytes.Equal(salt1, salt2) {
		t.Fatal("Each encryption should use a unique salt")
	}

	if bytes.Equal(encrypted1, encrypted2) {
		t.Fatal("Encrypted data should differ due to random salt/nonce")
	}
}

func TestPasswordVariations(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		password string
	}{
		{"Short password", "short"},
		{"Long password", strings.Repeat("a", 256)},
		{"Special characters", "P@$$w0rd!#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"Unicode", "пароль密码パスワード🔐"},
		{"Spaces", "password with spaces"},
		{"Newlines", "password\nwith\nnewlines"},
		{"Tabs", "password\twith\ttabs"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(tmpDir, tc.name+".db")
			encryptedPath := filepath.Join(tmpDir, tc.name+".encrypted")
			decryptedPath := filepath.Join(tmpDir, tc.name+".decrypted")

			testData := []byte("Test data for " + tc.name)
			if err := os.WriteFile(inputPath, testData, 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			if err := EncryptFile(inputPath, encryptedPath, tc.password); err != nil {
				t.Fatalf("Encryption with %s failed: %v", tc.name, err)
			}

			if err := DecryptFile(encryptedPath, decryptedPath, tc.password); err != nil {
				t.Fatalf("Decryption with %s failed: %v", tc.name, err)
			}

			decryptedData, _ := os.ReadFile(decryptedPath)
			if !bytes.Equal(decryptedData, testData) {
				t.Fatalf("Data mismatch with %s password", tc.name)
			}
		})
	}
}

func TestEncryptionConstants(t *testing.T) {
	// Verify constants match CLI implementation for cross-platform compatibility
	t.Run("Key size is 32 bytes (AES-256)", func(t *testing.T) {
		if keySize != 32 {
			t.Fatalf("Key size should be 32 for AES-256, got %d", keySize)
		}
	})

	t.Run("Salt size is 32 bytes", func(t *testing.T) {
		if saltSize != 32 {
			t.Fatalf("Salt size should be 32, got %d", saltSize)
		}
	})

	t.Run("Nonce size is 12 bytes (GCM standard)", func(t *testing.T) {
		if nonceSize != 12 {
			t.Fatalf("Nonce size should be 12 for GCM, got %d", nonceSize)
		}
	})

	t.Run("PBKDF2 iterations is 100000", func(t *testing.T) {
		if pbkdf2Iter != 100000 {
			t.Fatalf("PBKDF2 iterations should be 100000, got %d", pbkdf2Iter)
		}
	})
}

func TestEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "empty.db")
	encryptedPath := filepath.Join(tmpDir, "empty.encrypted")
	decryptedPath := filepath.Join(tmpDir, "empty.decrypted")

	// Create empty file
	if err := os.WriteFile(inputPath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	password := "EmptyFileTest123!"

	if err := EncryptFile(inputPath, encryptedPath, password); err != nil {
		t.Fatalf("Encryption of empty file failed: %v", err)
	}

	if err := DecryptFile(encryptedPath, decryptedPath, password); err != nil {
		t.Fatalf("Decryption of empty file failed: %v", err)
	}

	decryptedData, _ := os.ReadFile(decryptedPath)
	if len(decryptedData) != 0 {
		t.Fatalf("Decrypted empty file should be empty, got %d bytes", len(decryptedData))
	}
}
