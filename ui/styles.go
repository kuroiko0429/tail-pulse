package ui

import "github.com/charmbracelet/lipgloss"

var (
	cyan     = lipgloss.Color("#0ff")
	red      = lipgloss.Color("#f00")
	green    = lipgloss.Color("#0f0")
	yellow   = lipgloss.Color("#ff0")
	darkGrey = lipgloss.Color("#333")
	white    = lipgloss.Color("#fff")
	bg       = lipgloss.Color("#000")

	cyanStyle = lipgloss.NewStyle().Foreground(cyan)

	// Base
	baseStyle = lipgloss.NewStyle().Foreground(white).Background(bg)

	// Titles
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

	// Status
	onlineStyle  = lipgloss.NewStyle().Foreground(green)
	offlineStyle = lipgloss.NewStyle().Foreground(darkGrey)
	relayStyle   = lipgloss.NewStyle().Foreground(yellow)
	directStyle  = lipgloss.NewStyle().Foreground(cyan)
	loadingStyle = lipgloss.NewStyle().Foreground(yellow).Blink(true)

	// Components
	selectedStyle = lipgloss.NewStyle().Background(darkGrey).Foreground(white).Bold(true)
	detailStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(cyan).Padding(1, 2)
	notifyStyle   = lipgloss.NewStyle().Foreground(bg).Background(cyan).Padding(0, 1).Bold(true)
	footerStyle   = lipgloss.NewStyle().Foreground(darkGrey)

	// Tabs
	activeTabStyle   = lipgloss.NewStyle().Background(cyan).Foreground(bg).Padding(0, 2).Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Background(darkGrey).Foreground(white).Padding(0, 2)

	// Logs
	logErrStyle  = lipgloss.NewStyle().Foreground(red)
	logWarnStyle = lipgloss.NewStyle().Foreground(yellow)
	logInfoStyle = lipgloss.NewStyle().Foreground(darkGrey)

	// Columns
	colHost   = lipgloss.NewStyle().Width(20).MaxHeight(1)
	colIP     = lipgloss.NewStyle().Width(16).MaxHeight(1)
	colOS     = lipgloss.NewStyle().Width(6).MaxHeight(1)
	colStatus = lipgloss.NewStyle().Width(10).MaxHeight(1)
	colPorts  = lipgloss.NewStyle().Width(12).MaxHeight(1)
	colConn   = lipgloss.NewStyle().Width(12).MaxHeight(1)
	colPing   = lipgloss.NewStyle().Width(25).MaxHeight(1)
)
