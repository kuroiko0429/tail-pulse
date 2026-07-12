package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme       string `yaml:"theme"`
	ShowPing    bool   `yaml:"show_ping"`
	CyberGlitch bool   `yaml:"cyber_glitch"`
}

var DefaultConfig = Config{
	Theme:       "cyberpunk",
	ShowPing:    true,
	CyberGlitch: true,
}

func LoadConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig
	}

	configDir := filepath.Join(homeDir, ".config", "tail-pulse")
	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err == nil {
			data, _ := yaml.Marshal(DefaultConfig)
			_ = os.WriteFile(configPath, data, 0644)
		}
		return DefaultConfig
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return DefaultConfig
	}
	return conf
}
