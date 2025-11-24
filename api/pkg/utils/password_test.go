package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("Generated hash is empty")
	}

	if hash == password {
		t.Fatal("Hash should not equal plain password")
	}

	// Bcrypt hashes start with "$2a$" or "$2b$"
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("Hash doesn't look like a bcrypt hash: %s", hash)
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "SamePassword"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password (first): %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password (second): %v", err)
	}

	// Due to salt, same password should produce different hashes
	if hash1 == hash2 {
		t.Error("Same password produced identical hashes (salt not working)")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	password := "CorrectPassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Error("CheckPassword failed for correct password")
	}
}

func TestCheckPassword_Incorrect(t *testing.T) {
	password := "CorrectPassword"
	wrongPassword := "WrongPassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if CheckPassword(hash, wrongPassword) {
		t.Error("CheckPassword succeeded for incorrect password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "RealPassword"
	emptyPassword := ""

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if CheckPassword(hash, emptyPassword) {
		t.Error("CheckPassword succeeded for empty password")
	}
}

func TestCheckPassword_CaseSensitive(t *testing.T) {
	password := "Password123"
	uppercasePassword := "PASSWORD123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if CheckPassword(hash, uppercasePassword) {
		t.Error("CheckPassword should be case-sensitive")
	}
}

func TestHashPassword_EmptyString(t *testing.T) {
	// Bcrypt should still work with empty strings
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("Failed to hash empty password: %v", err)
	}

	if !CheckPassword(hash, "") {
		t.Error("CheckPassword failed for empty password")
	}

	if CheckPassword(hash, "nonempty") {
		t.Error("CheckPassword succeeded with wrong password for empty hash")
	}
}

func TestHashPassword_LongPassword(t *testing.T) {
	// Bcrypt can handle passwords up to 72 bytes
	validLongPassword := strings.Repeat("a", 72)

	hash, err := HashPassword(validLongPassword)
	if err != nil {
		t.Fatalf("Failed to hash long password: %v", err)
	}

	if !CheckPassword(hash, validLongPassword) {
		t.Error("CheckPassword failed for long password")
	}
}

func TestHashPassword_TooLongPassword(t *testing.T) {
	// Bcrypt has a limit of 72 bytes
	tooLongPassword := strings.Repeat("a", 100)

	_, err := HashPassword(tooLongPassword)
	if err == nil {
		t.Error("Expected error for password exceeding 72 bytes")
	}
}

func TestHashPassword_SpecialCharacters(t *testing.T) {
	specialPassword := "P@ssw0rd!#$%^&*()_+-={}[]|:;<>?,./~`"

	hash, err := HashPassword(specialPassword)
	if err != nil {
		t.Fatalf("Failed to hash password with special characters: %v", err)
	}

	if !CheckPassword(hash, specialPassword) {
		t.Error("CheckPassword failed for password with special characters")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	password := "TestPassword"
	invalidHash := "not-a-valid-bcrypt-hash"

	if CheckPassword(invalidHash, password) {
		t.Error("CheckPassword should fail for invalid hash")
	}
}
