package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	BaseURL     string `json:"base_url"`
	FrontendURL string `json:"frontend_url"`
}

var (
	DefaultBaseURL     = "https://ctf.rootaccess.live/api"
	DefaultFrontendURL = "https://ctf.rootaccess.live"
)

const (
	ConfigDir  = ".config/rootaccess"
	ConfigFile = "config.json"
)

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ConfigDir, ConfigFile)
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{BaseURL: DefaultBaseURL, FrontendURL: DefaultFrontendURL}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	if cfg.FrontendURL == "" {
		cfg.FrontendURL = DefaultFrontendURL
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (c *Config) Clear() error {
	c.Token = ""
	c.Username = ""
	return c.Save()
}
