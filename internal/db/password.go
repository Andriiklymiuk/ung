package db

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

var passwordCache string

// GetDatabasePassword retrieves the database password from cache, environment, or prompt
func GetDatabasePassword() (string, error) {
	// Check cache first
	if passwordCache != "" {
		return passwordCache, nil
	}

	// Check environment variable
	if envPassword := os.Getenv("UNG_DB_PASSWORD"); envPassword != "" {
		passwordCache = envPassword
		return passwordCache, nil
	}

	// Prompt user
	fmt.Print("Enter database password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after password input

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	password := string(passwordBytes)
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Cache for this session
	passwordCache = password
	return password, nil
}

// SetDatabasePassword sets the password cache (useful for testing or automation)
func SetDatabasePassword(password string) {
	passwordCache = password
}

// ClearPasswordCache clears the cached password
func ClearPasswordCache() {
	passwordCache = ""
}
