package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Theme struct {
	Cyan        string `yaml:"cyan"`
	DarkGrey    string `yaml:"dark_grey"`
	Red         string `yaml:"red"`
	White       string `yaml:"white"`
	Green       string `yaml:"green"`
	Yellow      string `yaml:"yellow"`
	TabActive   string `yaml:"tab_active"`
	TabInactive string `yaml:"tab_inactive"`
	Highlight   string `yaml:"highlight"`
}

type Config struct {
	PingInterval int               `yaml:"ping_interval"`
	DefaultSort  string            `yaml:"default_sort"`
	Ports        map[string]string `yaml:"ports"`
	MacAddresses map[string]string `yaml:"mac_addresses"`
	WolProxy     string            `yaml:"wol_proxy"`
	Theme        Theme             `yaml:"theme"`
}

func loadConfig() Config {
	cfg := Config{
		PingInterval: 15,
		DefaultSort:  "Name",
		Ports:        make(map[string]string),
		MacAddresses: make(map[string]string),
		WolProxy:     "",
		Theme: Theme{
			Cyan:        "#83a598",
			DarkGrey:    "#928374",
			Red:         "#fb4934",
			White:       "#ebdbb2",
			Green:       "#8ec07c",
			Yellow:      "#fabd2f",
			TabActive:   "#83a598",
			TabInactive: "#3c3836",
			Highlight:   "#d3869b",
		},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	configDir := filepath.Join(home, ".config", "tailpuls")
	configPath := filepath.Join(configDir, "config.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
		data, _ := yaml.Marshal(cfg)
		os.WriteFile(configPath, data, 0644)
		return cfg
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		yaml.Unmarshal(data, &cfg)
	}

	if cfg.Theme.Cyan == "" { cfg.Theme.Cyan = "#83a598" }
	if cfg.Theme.DarkGrey == "" { cfg.Theme.DarkGrey = "#928374" }
	if cfg.Theme.Red == "" { cfg.Theme.Red = "#fb4934" }
	if cfg.Theme.White == "" { cfg.Theme.White = "#ebdbb2" }
	if cfg.Theme.Green == "" { cfg.Theme.Green = "#8ec07c" }
	if cfg.Theme.Yellow == "" { cfg.Theme.Yellow = "#fabd2f" }
	if cfg.Theme.TabActive == "" { cfg.Theme.TabActive = "#83a598" }
	if cfg.Theme.TabInactive == "" { cfg.Theme.TabInactive = "#3c3836" }
	if cfg.Theme.Highlight == "" { cfg.Theme.Highlight = "#d3869b" }

	return cfg
}

func getSortMode(s string) SortMode {
	switch s {
	case "IP":
		return SortByIP
	case "OS":
		return SortByOS
	case "Ping":
		return SortByPing
	default:
		return SortByName
	}
}
