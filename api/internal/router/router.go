package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"ung/api/internal/controllers"
	ungMiddleware "ung/api/internal/middleware"
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
	settingsController *controllers.SettingsController,
	rateController *controllers.RateController,
	goalController *controllers.GoalController,
	subscriptionController *ungMiddleware.SubscriptionController,
	recurringController *controllers.RecurringController,
	reportController *controllers.ReportController,
	pomodoroController *controllers.PomodoroController,
	templateController *controllers.TemplateController,
	searchController *controllers.SearchController,
	exportController *controllers.ExportController,
	authMiddleware func(http.Handler) http.Handler,
	tenantMiddleware func(http.Handler) http.Handler,
	subscriptionMiddleware func(http.Handler) http.Handler,
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
					r.Get("/profit", dashboardController.GetProfit)
				})

				// Settings
				r.Route("/settings", func(r chi.Router) {
					r.Get("/", settingsController.Get)
					r.Put("/", settingsController.Update)
					r.Get("/working-hours", settingsController.GetWorkingHours)
				})

				// Rate Calculator
				r.Route("/rate", func(r chi.Router) {
					r.Post("/calculate", rateController.Calculate)
					r.Get("/analyze", rateController.Analyze)
					r.Get("/compare", rateController.Compare)
				})

				// Income Goals
				r.Route("/goals", func(r chi.Router) {
					r.Get("/", goalController.List)
					r.Post("/", goalController.Create)
					r.Get("/status", goalController.Status)
					r.Get("/{id}", goalController.Get)
					r.Put("/{id}", goalController.Update)
					r.Delete("/{id}", goalController.Delete)
				})

				// Recurring Invoices
				r.Route("/recurring", func(r chi.Router) {
					r.Get("/", recurringController.List)
					r.Post("/", recurringController.Create)
					r.Post("/generate", recurringController.Generate)
					r.Get("/{id}", recurringController.Get)
					r.Put("/{id}", recurringController.Update)
					r.Delete("/{id}", recurringController.Delete)
					r.Post("/{id}/pause", recurringController.Pause)
					r.Post("/{id}/resume", recurringController.Resume)
					r.Post("/{id}/generate", recurringController.GenerateSingle)
				})

				// Reports
				r.Route("/reports", func(r chi.Router) {
					r.Get("/weekly", reportController.Weekly)
					r.Get("/monthly", reportController.Monthly)
					r.Get("/revenue", reportController.Revenue)
					r.Get("/clients", reportController.Clients)
					r.Get("/overdue", reportController.Overdue)
					r.Get("/unpaid", reportController.Unpaid)
				})

				// Pomodoro Timer
				r.Route("/pomodoro", func(r chi.Router) {
					r.Get("/", pomodoroController.List)
					r.Get("/active", pomodoroController.Active)
					r.Get("/stats", pomodoroController.Stats)
					r.Post("/start", pomodoroController.Start)
					r.Get("/{id}", pomodoroController.Get)
					r.Post("/{id}/stop", pomodoroController.Stop)
					r.Post("/{id}/complete", pomodoroController.Complete)
					r.Delete("/{id}", pomodoroController.Delete)
				})

				// Templates
				r.Route("/templates", func(r chi.Router) {
					r.Get("/", templateController.List)
					r.Post("/", templateController.Create)
					r.Get("/default", templateController.GetDefault)
					r.Get("/{id}", templateController.Get)
					r.Put("/{id}", templateController.Update)
					r.Delete("/{id}", templateController.Delete)
					r.Post("/{id}/preview", templateController.Preview)
				})

				// Search
				r.Route("/search", func(r chi.Router) {
					r.Get("/", searchController.Search)
					r.Get("/invoices", searchController.SearchInvoices)
					r.Get("/clients", searchController.SearchClients)
					r.Get("/contracts", searchController.SearchContracts)
				})

				// Export
				r.Route("/export", func(r chi.Router) {
					r.Get("/all", exportController.ExportAllJSON)
					r.Get("/invoices/csv", exportController.ExportInvoicesCSV)
					r.Get("/expenses/csv", exportController.ExportExpensesCSV)
					r.Get("/tracking/csv", exportController.ExportTrackingCSV)
					r.Get("/clients/csv", exportController.ExportClientsCSV)
				})

				// Import
				r.Route("/import", func(r chi.Router) {
					r.Post("/all", exportController.ImportAllJSON)
					r.Post("/clients/csv", exportController.ImportClientsCSV)
					r.Post("/expenses/csv", exportController.ImportExpensesCSV)
				})
			})

			// Subscription routes (no tenant DB needed)
			r.Route("/subscription", func(r chi.Router) {
				r.Use(subscriptionMiddleware)
				r.Get("/", subscriptionController.GetStatus)
				r.Post("/verify", subscriptionController.VerifyPurchase)
			})
		})
	})

	return r
}
