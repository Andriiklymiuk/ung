package db

import (
	"fmt"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	// KeychainService is the service name used in the OS keychain
	KeychainService = "ung-database"
	// KeychainUser is the account name used in the OS keychain
	KeychainUser = "database-password"
)

// KeychainAvailable checks if keychain functionality is available on the current platform
func KeychainAvailable() bool {
	// go-keyring supports macOS, Windows, and Linux (with Secret Service)
	switch runtime.GOOS {
	case "darwin", "windows", "linux":
		return true
	default:
		return false
	}
}

// GetKeychainPlatformName returns a human-readable name for the keychain on the current platform
func GetKeychainPlatformName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "windows":
		return "Windows Credential Manager"
	case "linux":
		return "Secret Service (libsecret)"
	default:
		return "system keychain"
	}
}

// SavePasswordToKeychain saves the database password to the OS keychain
func SavePasswordToKeychain(password string) error {
	if !KeychainAvailable() {
		return fmt.Errorf("keychain not available on %s", runtime.GOOS)
	}

	err := keyring.Set(KeychainService, KeychainUser, password)
	if err != nil {
		return fmt.Errorf("failed to save password to %s: %w", GetKeychainPlatformName(), err)
	}

	return nil
}

// GetPasswordFromKeychain retrieves the database password from the OS keychain
func GetPasswordFromKeychain() (string, error) {
	if !KeychainAvailable() {
		return "", fmt.Errorf("keychain not available on %s", runtime.GOOS)
	}

	password, err := keyring.Get(KeychainService, KeychainUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", nil // Password not stored, not an error
		}
		return "", fmt.Errorf("failed to get password from %s: %w", GetKeychainPlatformName(), err)
	}

	return password, nil
}

// DeletePasswordFromKeychain removes the database password from the OS keychain
func DeletePasswordFromKeychain() error {
	if !KeychainAvailable() {
		return fmt.Errorf("keychain not available on %s", runtime.GOOS)
	}

	err := keyring.Delete(KeychainService, KeychainUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil // Already deleted or never saved, not an error
		}
		return fmt.Errorf("failed to delete password from %s: %w", GetKeychainPlatformName(), err)
	}

	return nil
}

// HasPasswordInKeychain checks if a password is stored in the keychain
func HasPasswordInKeychain() bool {
	if !KeychainAvailable() {
		return false
	}

	_, err := keyring.Get(KeychainService, KeychainUser)
	return err == nil
}
