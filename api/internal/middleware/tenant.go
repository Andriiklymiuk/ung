package middleware

import (
	"context"
	"net/http"

	"gorm.io/gorm"
	"ung/api/internal/database"
)

// TenantMiddleware opens user's database and adds to context
func TenantMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// User already set by AuthMiddleware
			user := GetUser(r)
			if user == nil {
				respondError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Open user's specific database
			tenantDB, err := database.InitUserDatabase(user.DBPath)
			if err != nil {
				respondError(w, "Database error", http.StatusInternalServerError)
				return
			}

			// Add tenant DB to context
			ctx := context.WithValue(r.Context(), TenantDBKey, tenantDB)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTenantDB retrieves tenant database from request context
func GetTenantDB(r *http.Request) *gorm.DB {
	if db, ok := r.Context().Value(TenantDBKey).(*gorm.DB); ok {
		return db
	}
	return nil
}
