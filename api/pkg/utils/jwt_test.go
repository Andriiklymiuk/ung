package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessToken(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	secret := "test-secret"

	token, err := GenerateAccessToken(userID, email, secret)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Validate the token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate generated token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	if claims.Issuer != "ung-api" {
		t.Errorf("Expected Issuer 'ung-api', got %s", claims.Issuer)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uint(2)
	email := "refresh@example.com"
	secret := "test-secret"

	token, err := GenerateRefreshToken(userID, email, secret)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Validate the token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate generated token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	// Check expiration is approximately 7 days
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	if claims.ExpiresAt.Time.Before(expectedExpiry.Add(-1 * time.Minute)) ||
		claims.ExpiresAt.Time.After(expectedExpiry.Add(1 * time.Minute)) {
		t.Errorf("Token expiry not within expected range (7 days)")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	userID := uint(3)
	email := "validate@example.com"
	secret := "test-secret"

	token, err := GenerateAccessToken(userID, email, secret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	userID := uint(4)
	email := "invalid@example.com"
	secret := "test-secret"
	wrongSecret := "wrong-secret"

	token, err := GenerateAccessToken(userID, email, secret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = ValidateToken(token, wrongSecret)
	if err == nil {
		t.Fatal("Expected validation to fail with wrong secret")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	userID := uint(5)
	email := "expired@example.com"
	secret := "test-secret"

	// Create an expired token manually
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "ung-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ValidateToken(tokenString, secret)
	if err == nil {
		t.Fatal("Expected validation to fail for expired token")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	secret := "test-secret"
	malformedToken := "not.a.valid.token"

	_, err := ValidateToken(malformedToken, secret)
	if err == nil {
		t.Fatal("Expected validation to fail for malformed token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	secret := "test-secret"

	_, err := ValidateToken("", secret)
	if err == nil {
		t.Fatal("Expected validation to fail for empty token")
	}
}

func TestTokenExpiry_AccessToken(t *testing.T) {
	userID := uint(6)
	email := "expiry@example.com"
	secret := "test-secret"

	token, err := GenerateAccessToken(userID, email, secret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check that expiry is approximately 15 minutes
	expectedExpiry := time.Now().Add(15 * time.Minute)
	if claims.ExpiresAt.Time.Before(expectedExpiry.Add(-1 * time.Minute)) ||
		claims.ExpiresAt.Time.After(expectedExpiry.Add(1 * time.Minute)) {
		t.Errorf("Access token expiry not within expected range (15 minutes)")
	}
}
