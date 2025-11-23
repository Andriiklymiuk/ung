package config

import "os"

// Config holds bot configuration
type Config struct {
	BotToken   string
	APIURL     string
	WebAppURL  string
	Debug      bool
}

// Load loads configuration from environment variables
func Load() *Config {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		panic("TELEGRAM_BOT_TOKEN is required")
	}

	apiURL := os.Getenv("UNG_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	webAppURL := os.Getenv("WEB_APP_URL")
	if webAppURL == "" {
		webAppURL = "https://ung.app"
	}

	debug := os.Getenv("DEBUG") == "true"

	return &Config{
		BotToken:  botToken,
		APIURL:    apiURL,
		WebAppURL: webAppURL,
		Debug:     debug,
	}
}
