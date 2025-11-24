package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Encryption parameters
	keySize        = 32 // AES-256
	saltSize       = 32
	nonceSize      = 12 // GCM standard nonce size
	pbkdf2Iter     = 100000
	encryptedExt   = ".encrypted"
	plaintextExt   = ".db"
)

// EncryptionConfig holds encryption parameters
type EncryptionConfig struct {
	Password string
	Salt     []byte
}

// EncryptFile encrypts a file using AES-256-GCM with password-derived key
func EncryptFile(inputPath, outputPath string, password string) error {
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

// DecryptFile decrypts a file encrypted with EncryptFile
func DecryptFile(inputPath, outputPath string, password string) error {
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

// EncryptStream encrypts data from reader to writer
func EncryptStream(reader io.Reader, writer io.Writer, password string) error {
	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Write salt first
	if _, err := writer.Write(salt); err != nil {
		return fmt.Errorf("failed to write salt: %w", err)
	}

	// Derive key
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate and write nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Read all data (for simplicity - could be chunked for large files)
	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	// Encrypt and write
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	if _, err := writer.Write(ciphertext); err != nil {
		return fmt.Errorf("failed to write ciphertext: %w", err)
	}

	return nil
}

// DecryptStream decrypts data from reader to writer
func DecryptStream(reader io.Reader, writer io.Writer, password string) error {
	// Read salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return fmt.Errorf("failed to read salt: %w", err)
	}

	// Derive key
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Read remaining data
	ciphertext, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read ciphertext: %w", err)
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
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write plaintext
	if _, err := writer.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write plaintext: %w", err)
	}

	return nil
}

// GenerateSalt generates a random salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// DeriveKey derives an encryption key from password and salt
func DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha256.New)
}
