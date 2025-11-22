package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DatabasePath string `yaml:"database_path"`
	InvoicesDir  string `yaml:"invoices_dir"`
}

var currentConfig *Config

// Load loads configuration from local (.ung.yaml) or global (~/.ung/config.yaml)
func Load() (*Config, error) {
	if currentConfig != nil {
		return currentConfig, nil
	}

	// Try local config first
	localConfig := ".ung.yaml"
	if _, err := os.Stat(localConfig); err == nil {
		cfg, err := loadFromFile(localConfig)
		if err == nil {
			currentConfig = cfg
			return currentConfig, nil
		}
	}

	// Try global config
	home, err := os.UserHomeDir()
	if err != nil {
		return getDefaultConfig(), nil
	}

	globalConfig := filepath.Join(home, ".ung", "config.yaml")
	if _, err := os.Stat(globalConfig); err == nil {
		cfg, err := loadFromFile(globalConfig)
		if err == nil {
			currentConfig = cfg
			return currentConfig, nil
		}
	}

	// Return default config
	currentConfig = getDefaultConfig()
	return currentConfig, nil
}

// loadFromFile loads config from a YAML file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand paths
	if cfg.DatabasePath != "" {
		cfg.DatabasePath = expandPath(cfg.DatabasePath)
	}
	if cfg.InvoicesDir != "" {
		cfg.InvoicesDir = expandPath(cfg.InvoicesDir)
	}

	return &cfg, nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return &Config{
		DatabasePath: filepath.Join(home, ".ung", "ung.db"),
		InvoicesDir:  filepath.Join(home, ".ung", "invoices"),
	}
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// Save saves the current configuration to a file
func Save(cfg *Config, global bool) error {
	var path string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".ung", "config.yaml")
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
	} else {
		path = ".ung.yaml"
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDatabasePath returns the configured database path
func GetDatabasePath() string {
	cfg, _ := Load()
	return cfg.DatabasePath
}

// GetInvoicesDir returns the configured invoices directory
func GetInvoicesDir() string {
	cfg, _ := Load()
	return cfg.InvoicesDir
}
