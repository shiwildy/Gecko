package service

import (
	"encoding/json"
	"fmt"
	"gecko/internal/shared"
	"os"
)

const (
	geckoConfigPath = `C:\Gecko\gecko-config.json`
)

// Config struct holds all dynamic configurations for Gecko.
type Config struct {
	ApachePort       string `json:"apache_port"`
	ApacheSSLPort    string `json:"apache_ssl_port"`
	MySQLPort        string `json:"mysql_port"`
	PostgresPort     string `json:"postgres_port"`
	PostgresPassword string `json:"postgres_password"` // Field untuk menyimpan password
	DevelopmentMode  bool   `json:"development_mode"`
}

var globalConfig *Config

// LoadConfig reads the configuration from gecko-config.json.
func LoadConfig() (*Config, error) {
	if _, err := os.Stat(geckoConfigPath); os.IsNotExist(err) {
		fmt.Printf("%sConfig file not found. Creating a default one at %s...%s\n", shared.ColorYellow, geckoConfigPath, shared.ColorReset)
		defaultConfig := &Config{
			ApachePort:      "80",
			ApacheSSLPort:   "443",
			MySQLPort:       "3306",
			PostgresPort:    "5432",
			PostgresPassword: "", // Default kosong
			DevelopmentMode: false,
		}
		if err := SaveConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		globalConfig = defaultConfig
		return defaultConfig, nil
	}

	file, err := os.ReadFile(geckoConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

    // Penanganan jika file config lama
	if config.PostgresPort == "" {
		config.PostgresPort = "5432"
        SaveConfig(&config)
	}

	globalConfig = &config
	return &config, nil
}

// SaveConfig writes the provided config struct to gecko-config.json.
func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(geckoConfigPath, data, 0644)
}

// GetConfig returns the currently loaded configuration.
func GetConfig() (*Config, error) {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig, nil
}