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

type Config struct {
	ApachePort    string `json:"apache_port"`
	ApacheSSLPort string `json:"apache_ssl_port"`
	MySQLPort     string `json:"mysql_port"`
}

var globalConfig *Config

func LoadConfig() (*Config, error) {
	if _, err := os.Stat(geckoConfigPath); os.IsNotExist(err) {
		fmt.Printf("%sConfig file not found. Creating a default one at %s...%s\n", shared.ColorYellow, geckoConfigPath, shared.ColorReset)
		defaultConfig := &Config{
			ApachePort:    "80",
			ApacheSSLPort: "443",
			MySQLPort:     "3306",
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

	if config.ApacheSSLPort == "" {
		config.ApacheSSLPort = "443"
		fmt.Println("Updating config file with default SSL port 443...")
		SaveConfig(&config)
	}

	globalConfig = &config
	return &config, nil
}

func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(geckoConfigPath, data, 0644)
}

func GetConfig() (*Config, error) {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig, nil
}