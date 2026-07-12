package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/kuroiko0429/tail-pulse/config"
)

var (
	cyan     lipgloss.Color
	red      lipgloss.Color
	green    lipgloss.Color
	yellow   lipgloss.Color
	darkGrey lipgloss.Color
	white    lipgloss.Color
	bg       lipgloss.Color

	cyanStyle lipgloss.Style

	// Base
	baseStyle lipgloss.Style

	// Titles
	titleStyle  lipgloss.Style
	headerStyle lipgloss.Style

	// Status
	onlineStyle  lipgloss.Style
	offlineStyle lipgloss.Style
	relayStyle   lipgloss.Style
	directStyle  lipgloss.Style
	loadingStyle lipgloss.Style

	// Components
	selectedStyle lipgloss.Style
	detailStyle   lipgloss.Style
	notifyStyle   lipgloss.Style
	footerStyle   lipgloss.Style

	// Tabs
	activeTabStyle   lipgloss.Style
	inactiveTabStyle lipgloss.Style

	// Logs
	logErrStyle  lipgloss.Style
	logWarnStyle lipgloss.Style
	logInfoStyle lipgloss.Style

	// Columns
	colHost   = lipgloss.NewStyle().Width(20).MaxHeight(1)
	colIP     = lipgloss.NewStyle().Width(16).MaxHeight(1)
	colOS     = lipgloss.NewStyle().Width(6).MaxHeight(1)
	colStatus = lipgloss.NewStyle().Width(10).MaxHeight(1)
	colPorts  = lipgloss.NewStyle().Width(12).MaxHeight(1)
	colConn   = lipgloss.NewStyle().Width(12).MaxHeight(1)
	colPing   = lipgloss.NewStyle().Width(25).MaxHeight(1)
)

// InitStyles derives all lipgloss styles from the configured theme. Must be
// called before the first render (NewModel does this).
func InitStyles(t config.Theme) {
	cyan = lipgloss.Color(t.Cyan)
	darkGrey = lipgloss.Color(t.DarkGrey)
	red = lipgloss.Color(t.Red)
	white = lipgloss.Color(t.White)
	green = lipgloss.Color(t.Green)
	yellow = lipgloss.Color(t.Yellow)
	bg = lipgloss.Color(t.Background)
	tabActive := lipgloss.Color(t.TabActive)
	tabInactive := lipgloss.Color(t.TabInactive)

	cyanStyle = lipgloss.NewStyle().Foreground(cyan)
	baseStyle = lipgloss.NewStyle().Foreground(white).Background(bg)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(bg).
		Background(cyan).
		Padding(0, 1).
		MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Underline(true)

	onlineStyle = lipgloss.NewStyle().Foreground(green)
	offlineStyle = lipgloss.NewStyle().Foreground(darkGrey)
	relayStyle = lipgloss.NewStyle().Foreground(yellow)
	directStyle = lipgloss.NewStyle().Foreground(cyan)
	loadingStyle = lipgloss.NewStyle().Foreground(yellow).Blink(true)

	selectedStyle = lipgloss.NewStyle().Background(darkGrey).Foreground(white).Bold(true)
	detailStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(cyan).Padding(1, 2)
	notifyStyle = lipgloss.NewStyle().Foreground(bg).Background(cyan).Padding(0, 1).Bold(true)
	footerStyle = lipgloss.NewStyle().Foreground(darkGrey)

	activeTabStyle = lipgloss.NewStyle().Background(tabActive).Foreground(white).Padding(0, 2).Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Background(tabInactive).Foreground(white).Padding(0, 2)

	logErrStyle = lipgloss.NewStyle().Foreground(red)
	logWarnStyle = lipgloss.NewStyle().Foreground(yellow)
	logInfoStyle = lipgloss.NewStyle().Foreground(darkGrey)
}
