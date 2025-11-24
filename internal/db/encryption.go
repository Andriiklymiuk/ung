package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Encryption parameters
	keySize    = 32 // AES-256
	saltSize   = 32
	nonceSize  = 12 // GCM standard nonce size
	pbkdf2Iter = 100000
)

// EncryptDatabase encrypts a database file using AES-256-GCM with password-derived key
func EncryptDatabase(inputPath, outputPath string, password string) error {
	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha256.New)

	// Read input file
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write encrypted file: [salt][ciphertext]
	output := append(salt, ciphertext...)
	if err := os.WriteFile(outputPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// DecryptDatabase decrypts a database file encrypted with EncryptDatabase
func DecryptDatabase(inputPath, outputPath string, password string) error {
	// Read encrypted file
	encrypted, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Extract salt
	if len(encrypted) < saltSize {
		return fmt.Errorf("encrypted file too short")
	}
	salt := encrypted[:saltSize]
	ciphertext := encrypted[saltSize:]

	// Derive key from password
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	// Write decrypted file
	if err := os.WriteFile(outputPath, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// IsEncrypted checks if a file appears to be encrypted
// This is a heuristic check - it verifies the file size and tries to read the salt
func IsEncrypted(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// Encrypted files should be at least saltSize + nonceSize bytes
	if info.Size() < int64(saltSize+nonceSize) {
		return false
	}

	// Try to read the salt to verify file format
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	if len(data) < saltSize {
		return false
	}

	// Check if first saltSize bytes look like random data
	// by checking they're not all zeros and not all printable ASCII
	salt := data[:saltSize]
	allZeros := true
	allPrintable := true

	for _, b := range salt {
		if b != 0 {
			allZeros = false
		}
		if b < 32 || b > 126 {
			allPrintable = false
		}
	}

	// If all zeros or all printable ASCII, probably not encrypted
	return !allZeros && !allPrintable
}
