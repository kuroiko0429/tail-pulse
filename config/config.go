package config

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
	Background  string `yaml:"background"`
	TabActive   string `yaml:"tab_active"`
	TabInactive string `yaml:"tab_inactive"`
	Highlight   string `yaml:"highlight"`
}

// GruvboxTheme is the default palette, matching the Gruvbox dark colorscheme
// used across the rest of this user's setup.
var GruvboxTheme = Theme{
	Cyan:        "#83a598",
	DarkGrey:    "#928374",
	Red:         "#fb4934",
	White:       "#ebdbb2",
	Green:       "#8ec07c",
	Yellow:      "#fabd2f",
	Background:  "#282828",
	TabActive:   "#83a598",
	TabInactive: "#3c3836",
	Highlight:   "#d3869b",
}

type Config struct {
	Theme        Theme             `yaml:"theme"`
	ShowPing     bool              `yaml:"show_ping"`
	CyberGlitch  bool              `yaml:"cyber_glitch"`
	PingInterval int               `yaml:"ping_interval"`
	Ports        map[string]string `yaml:"ports"`
	MacAddresses map[string]string `yaml:"mac_addresses"`
	WolProxy     string            `yaml:"wol_proxy"`
}

var DefaultConfig = Config{
	Theme:        GruvboxTheme,
	ShowPing:     true,
	CyberGlitch:  true,
	PingInterval: 15,
	Ports:        map[string]string{},
	MacAddresses: map[string]string{},
	WolProxy:     "",
}

func fillThemeDefaults(t *Theme) {
	if t.Cyan == "" {
		t.Cyan = GruvboxTheme.Cyan
	}
	if t.DarkGrey == "" {
		t.DarkGrey = GruvboxTheme.DarkGrey
	}
	if t.Red == "" {
		t.Red = GruvboxTheme.Red
	}
	if t.White == "" {
		t.White = GruvboxTheme.White
	}
	if t.Green == "" {
		t.Green = GruvboxTheme.Green
	}
	if t.Yellow == "" {
		t.Yellow = GruvboxTheme.Yellow
	}
	if t.Background == "" {
		t.Background = GruvboxTheme.Background
	}
	if t.TabActive == "" {
		t.TabActive = GruvboxTheme.TabActive
	}
	if t.TabInactive == "" {
		t.TabInactive = GruvboxTheme.TabInactive
	}
	if t.Highlight == "" {
		t.Highlight = GruvboxTheme.Highlight
	}
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

	conf := DefaultConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return DefaultConfig
	}
	if conf.Ports == nil {
		conf.Ports = map[string]string{}
	}
	if conf.MacAddresses == nil {
		conf.MacAddresses = map[string]string{}
	}
	if conf.PingInterval <= 0 {
		conf.PingInterval = 15
	}
	fillThemeDefaults(&conf.Theme)
	return conf
}
