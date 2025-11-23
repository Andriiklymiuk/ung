package config

import (
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	Port            string
	Env             string
	APIDatabasePath string
	UserDataDir     string
	JWTSecret       string
	CORSOrigins     []string
}

// Load loads configuration from environment variables
func Load() *Config {
	home, _ := os.UserHomeDir()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	apiDBPath := os.Getenv("API_DATABASE_PATH")
	if apiDBPath == "" {
		apiDBPath = filepath.Join(home, ".ung", "api.db")
	}

	userDataDir := os.Getenv("USER_DATA_DIR")
	if userDataDir == "" {
		userDataDir = filepath.Join(home, ".ung", "users")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-this-super-secret-key-in-production"
	}

	corsOrigins := []string{"http://localhost:3000", "https://ung.app"}

	return &Config{
		Port:            port,
		Env:             env,
		APIDatabasePath: apiDBPath,
		UserDataDir:     userDataDir,
		JWTSecret:       jwtSecret,
		CORSOrigins:     corsOrigins,
	}
}
