package services

import (
	"errors"
	"fmt"

	"ung/api/internal/models"
	"ung/api/internal/repository"
	"ung/api/pkg/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	userDataDir string
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, userDataDir string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		userDataDir: userDataDir,
	}
}

// Register creates a new user account
func (s *AuthService) Register(email, password, name string) (*models.User, string, string, error) {
	// Check if user already exists
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, "", "", errors.New("user already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		PlanType:     "free",
		Active:       true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Create user database
	dbPath := fmt.Sprintf("%s/user_%d/ung.db", s.userDataDir, user.ID)
	user.DBPath = dbPath
	if err := s.userRepo.Update(user); err != nil {
		return nil, "", "", fmt.Errorf("failed to update user: %w", err)
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// Login authenticates a user
func (s *AuthService) Login(email, password string) (*models.User, string, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	// Check password
	if !utils.CheckPassword(user.PasswordHash, password) {
		return nil, "", "", errors.New("invalid credentials")
	}

	if !user.Active {
		return nil, "", "", errors.New("account disabled")
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// RefreshToken generates new tokens from refresh token
func (s *AuthService) RefreshToken(refreshTokenStr string) (string, string, error) {
	// Validate refresh token
	claims, err := utils.ValidateToken(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Generate new tokens
	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Email, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(claims.UserID, claims.Email, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
