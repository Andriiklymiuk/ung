package controllers

import (
	"encoding/json"
	"net/http"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
	"ung/api/internal/services"
)

// AuthController handles authentication endpoints
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register handles POST /api/v1/auth/register
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := c.authService.Register(req.Email, req.Password, req.Name)
	if err != nil {
		RespondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	RespondJSON(w, map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusCreated)
}

// Login handles POST /api/v1/auth/login
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		RespondError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	RespondJSON(w, map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusOK)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		RespondError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	RespondJSON(w, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusOK)
}

// GetProfile handles GET /api/v1/auth/me
func (c *AuthController) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	RespondJSON(w, user, http.StatusOK)
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.StandardResponse{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.StandardResponse{
		Success: false,
		Error:   message,
	}

	json.NewEncoder(w).Encode(response)
}
