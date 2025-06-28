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
	ApachePort      string `json:"apache_port"`
	ApacheSSLPort   string `json:"apache_ssl_port"`
	MySQLPort       string `json:"mysql_port"`
	DevelopmentMode bool   `json:"development_mode"`
}

var globalConfig *Config

func LoadConfig() (*Config, error) {
	if _, err := os.Stat(geckoConfigPath); os.IsNotExist(err) {
		fmt.Printf("%sConfig file not found. Creating a default one at %s...%s\n", shared.ColorYellow, geckoConfigPath, shared.ColorReset)
		defaultConfig := &Config{
			ApachePort:      "80",
			ApacheSSLPort:   "443",
			MySQLPort:       "3306",
			DevelopmentMode: false, // set disable as default
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
		var rawConfig map[string]interface{}
		if json.Unmarshal(file, &rawConfig) == nil {
			if _, ok := rawConfig["development_mode"]; !ok {
				rawConfig["development_mode"] = false
			}
			config.ApachePort = rawConfig["apache_port"].(string)
			config.ApacheSSLPort = rawConfig["apache_ssl_port"].(string)
			config.MySQLPort = rawConfig["mysql_port"].(string)
			config.DevelopmentMode = rawConfig["development_mode"].(bool)
			SaveConfig(&config)
		} else {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
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