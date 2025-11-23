package main

import (
	"fmt"
	"log"
	"net/http"

	"ung/api/internal/config"
	"ung/api/internal/controllers"
	"ung/api/internal/database"
	"ung/api/internal/middleware"
	"ung/api/internal/repository"
	"ung/api/internal/router"
	"ung/api/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting UNG API server...")
	log.Printf("Environment: %s", cfg.Env)
	log.Printf("Port: %s", cfg.Port)

	// Initialize API database
	apiDB, err := database.InitAPIDatabase(cfg.APIDatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize API database: %v", err)
	}
	log.Printf("API database initialized: %s", cfg.APIDatabasePath)

	// Initialize repositories
	userRepo := repository.NewUserRepository(apiDB)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.UserDataDir)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	invoiceController := controllers.NewInvoiceController()

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(apiDB, cfg.JWTSecret)
	tenantMiddleware := middleware.TenantMiddleware()

	// Setup router
	r := router.SetupRouter(
		authController,
		invoiceController,
		authMiddleware,
		tenantMiddleware,
	)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server listening on %s", addr)
	log.Printf("API endpoints available at http://localhost:%s/api/v1", cfg.Port)
	log.Printf("Health check: http://localhost:%s/health", cfg.Port)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
