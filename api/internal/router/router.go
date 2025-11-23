package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"ung/api/internal/controllers"
)

// SetupRouter creates and configures the Chi router
func SetupRouter(
	authController *controllers.AuthController,
	invoiceController *controllers.InvoiceController,
	authMiddleware func(http.Handler) http.Handler,
	tenantMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authController.Register)
			r.Post("/auth/login", authController.Login)
			r.Post("/auth/refresh", authController.RefreshToken)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// Auth endpoints
			r.Get("/auth/me", authController.GetProfile)

			// Tenant-specific routes (require user database)
			r.Group(func(r chi.Router) {
				r.Use(tenantMiddleware)

				// Invoices
				r.Route("/invoices", func(r chi.Router) {
					r.Get("/", invoiceController.List)
					r.Get("/{id}", invoiceController.Get)
					// Add POST, PUT, DELETE as needed
				})

				// Add other resources (clients, contracts, etc.)
			})
		})
	})

	return r
}
