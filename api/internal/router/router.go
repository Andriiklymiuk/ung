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
	clientController *controllers.ClientController,
	companyController *controllers.CompanyController,
	contractController *controllers.ContractController,
	expenseController *controllers.ExpenseController,
	trackingController *controllers.TrackingController,
	dashboardController *controllers.DashboardController,
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
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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

				// Clients
				r.Route("/clients", func(r chi.Router) {
					r.Get("/", clientController.List)
					r.Post("/", clientController.Create)
					r.Get("/{id}", clientController.Get)
					r.Put("/{id}", clientController.Update)
					r.Delete("/{id}", clientController.Delete)
				})

				// Invoices
				r.Route("/invoices", func(r chi.Router) {
					r.Get("/", invoiceController.List)
					r.Post("/", invoiceController.Create)
					r.Get("/{id}", invoiceController.Get)
					r.Put("/{id}", invoiceController.Update)
					r.Delete("/{id}", invoiceController.Delete)
					r.Patch("/{id}/status", invoiceController.UpdateStatus)
				})

				// Companies
				r.Route("/companies", func(r chi.Router) {
					r.Get("/", companyController.List)
					r.Post("/", companyController.Create)
					r.Get("/{id}", companyController.Get)
					r.Put("/{id}", companyController.Update)
					r.Delete("/{id}", companyController.Delete)
				})

				// Contracts
				r.Route("/contracts", func(r chi.Router) {
					r.Get("/", contractController.List)
					r.Post("/", contractController.Create)
					r.Get("/{id}", contractController.Get)
					r.Put("/{id}", contractController.Update)
					r.Delete("/{id}", contractController.Delete)
				})

				// Expenses
				r.Route("/expenses", func(r chi.Router) {
					r.Get("/", expenseController.List)
					r.Post("/", expenseController.Create)
					r.Get("/{id}", expenseController.Get)
					r.Put("/{id}", expenseController.Update)
					r.Delete("/{id}", expenseController.Delete)
				})

				// Time Tracking
				r.Route("/tracking", func(r chi.Router) {
					r.Get("/", trackingController.List)
					r.Post("/", trackingController.Create)
					r.Get("/active", trackingController.Active)
					r.Post("/start", trackingController.Start)
					r.Get("/{id}", trackingController.Get)
					r.Post("/{id}/stop", trackingController.Stop)
					r.Put("/{id}", trackingController.Update)
					r.Delete("/{id}", trackingController.Delete)
				})

				// Dashboard
				r.Route("/dashboard", func(r chi.Router) {
					r.Get("/revenue", dashboardController.GetRevenue)
					r.Get("/summary", dashboardController.GetSummary)
				})
			})
		})
	})

	return r
}
