package middleware

import (
	"context"
	"net/http"
	"strings"

	"gorm.io/gorm"
	"ung/api/internal/models"
	"ung/api/internal/repository"
	"ung/api/pkg/utils"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
	TenantDBKey    contextKey = "tenantDB"
)

// AuthMiddleware validates JWT tokens and adds user to context
func AuthMiddleware(apiDB *gorm.DB, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				respondError(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			// Validate JWT
			claims, err := utils.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				respondError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Get user from database
			userRepo := repository.NewUserRepository(apiDB)
			user, err := userRepo.GetByID(claims.UserID)
			if err != nil {
				respondError(w, "User not found", http.StatusUnauthorized)
				return
			}

			if !user.Active {
				respondError(w, "Account disabled", http.StatusForbidden)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser retrieves user from request context
func GetUser(r *http.Request) *models.User {
	if user, ok := r.Context().Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// respondError sends an error response
func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"success":false,"error":"` + message + `"}`))
}
