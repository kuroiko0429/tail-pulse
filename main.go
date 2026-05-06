package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

// --- Data Structures ---

type TailscaleStatus struct {
	Peer map[string]PeerStatus `json:"Peer"`
	Self PeerStatus            `json:"Self"`
}

type PeerStatus struct {
	DNSName             string    `json:"DNSName"`
	HostName            string    `json:"HostName"`
	TailscaleIPs        []string  `json:"TailscaleIPs"`
	OS                  string    `json:"OS"`
	Online              bool      `json:"Online"`
	Active              bool      `json:"Active"`
	Relay               string    `json:"Relay"`
	CurAddr             string    `json:"CurAddr"`
	LastSeen            time.Time `json:"LastSeen"`
	ExitNodeOption      bool      `json:"ExitNodeOption"`
	ExitNode            bool      `json:"ExitNode"`
	Tags                []string  `json:"Tags"`
	TailscaleSSHEnabled bool      `json:"TailscaleSSHEnabled"`
	PrimaryRoutes       []string  `json:"PrimaryRoutes"`
	KeyExpiry           time.Time `json:"KeyExpiry"`
	TxBytes             int64     `json:"TxBytes"`
	RxBytes             int64     `json:"RxBytes"`
}

type pingInfo struct {
	latency []float64
	current float64
	canSSH  bool
	derp    string // DERP region or "Direct"
}

// --- Dynamic Style Definitions ---

var (
	titleStyle, headerStyle, onlineStyle, offlineStyle, relayStyle, directStyle, footerStyle, selectedStyle, notifyStyle, pingStyle, highlightStyle, activeTabStyle, inactiveTabStyle lipgloss.Style
	colHost, colIP, colOS, colStatus, colSSH, colConn, colPing lipgloss.Style
	logErrStyle, logWarnStyle, logInfoStyle lipgloss.Style

	iconSSHDefault   = "󱘖"
	iconTailscaleSSH = "󰣀"
	
	red, darkGrey, yellow lipgloss.Color
)

func initStyles(t Theme) {
	cyan := lipgloss.Color(t.Cyan)
	darkGrey = lipgloss.Color(t.DarkGrey)
	red = lipgloss.Color(t.Red)
	white := lipgloss.Color(t.White)
	green := lipgloss.Color(t.Green)
	yellow = lipgloss.Color(t.Yellow)
	searchHighlight := lipgloss.Color(t.Highlight)
	tabActive := lipgloss.Color(t.TabActive)
	tabInactive := lipgloss.Color(t.TabInactive)

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(white).Background(red).Padding(0, 1).MarginBottom(1)
	headerStyle = lipgloss.NewStyle().Foreground(cyan).Bold(true).Underline(true)

	onlineStyle = lipgloss.NewStyle().Foreground(green)
	offlineStyle = lipgloss.NewStyle().Foreground(darkGrey)
	relayStyle = lipgloss.NewStyle().Foreground(yellow)
	directStyle = lipgloss.NewStyle().Foreground(cyan)
	footerStyle = lipgloss.NewStyle().Foreground(darkGrey)
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#2a2827")).Foreground(white) 
	notifyStyle = lipgloss.NewStyle().Foreground(white).Background(cyan).Padding(0, 1).Bold(true)
	pingStyle = lipgloss.NewStyle().Foreground(cyan)
	highlightStyle = lipgloss.NewStyle().Foreground(searchHighlight).Bold(true)

	activeTabStyle = lipgloss.NewStyle().Background(tabActive).Foreground(white).Padding(0, 2).Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Background(tabInactive).Foreground(white).Padding(0, 2)

	colHost = lipgloss.NewStyle().Width(22).MaxHeight(1)
	colIP = lipgloss.NewStyle().Width(16).MaxHeight(1)
	colOS = lipgloss.NewStyle().Width(6).MaxHeight(1)
	colStatus = lipgloss.NewStyle().Width(12).MaxHeight(1)
	colSSH = lipgloss.NewStyle().Width(6).MaxHeight(1)
	colConn = lipgloss.NewStyle().Width(15).MaxHeight(1)
	colPing = lipgloss.NewStyle().Width(25).MaxHeight(1)

	logErrStyle = lipgloss.NewStyle().Foreground(red)
	logWarnStyle = lipgloss.NewStyle().Foreground(yellow)
	logInfoStyle = lipgloss.NewStyle().Foreground(darkGrey)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit { return fmt.Sprintf("%d B", b) }
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit { div *= unit; exp++ }
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// --- Enums ---
type SortMode int

const (
	SortByName SortMode = iota
	SortByIP
	SortByOS
	SortByPing
)

func (s SortMode) String() string {
	switch s {
	case SortByName: return "Name"
	case SortByIP: return "IP"
	case SortByOS: return "OS"
	case SortByPing: return "Ping"
	default: return "Unknown"
	}
}

type TabMode int

const (
	TabDevices TabMode = iota
	TabExitNodes
	TabServe
	TabLogs
	TabDaemon
)

// --- Bubble Tea Model ---

type model struct {
	config         Config
	status         TailscaleStatus
	err            error
	errMessage     string
	cursor         int
	peers          []PeerStatus
	notifMsg       string
	sshTarget      string
	pings          map[string]*pingInfo
	sshPorts       map[string]string
	showOnlineOnly bool
	sortMode       SortMode

	// Window resize & scrolling
	width         int
	height        int
	viewportStart int
	logsScroll    int

	// Search
	searchInput textinput.Model
	isSearching bool

	// Tabs & Details
	activeTab      TabMode
	isDetailView   bool
	fileInput      textinput.Model
	isFileTransfer bool

	// Notifications Tracker
	lastOnline map[string]bool
	
	// Logs Streamer
	logs     []string
	logChan  chan string
	
	// Serve State
	serveStatus string
}

type tickStatusMsg time.Time
type tickConnMsg time.Time
type clearNotifMsg struct{}
type showNotifMsg string
type pingResultMsg struct {
	host    string
	latency float64
	derp    string
	canSSH  bool
}
type portUpdateMsg struct {
	host string
	port string
}
type logLineMsg string
type serveStatusMsg string

func initialModel() model {
	cfg := loadConfig()
	initStyles(cfg.Theme)

	ti := textinput.New()
	ti.Placeholder = "Search hostname or IP..."
	ti.CharLimit = 156
	ti.Width = 40

	fi := textinput.New()
	fi.Placeholder = "/path/to/file.txt"
	fi.CharLimit = 256
	fi.Width = 60

	m := model{
		config:      cfg,
		pings:       make(map[string]*pingInfo),
		sshPorts:    make(map[string]string),
		lastOnline:  make(map[string]bool),
		logChan:     make(chan string, 100),
		searchInput: ti,
		fileInput:   fi,
		sortMode:    getSortMode(cfg.DefaultSort),
		activeTab:   TabDevices,
	}

	go streamLogs(m.logChan)

	return m
}

func streamLogs(sub chan<- string) {
	cmd := exec.Command("journalctl", "-u", "tailscaled", "-n", "100", "-f")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sub <- fmt.Sprintf("Failed to stream logs: %v", err)
		return
	}
	cmd.Start()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		sub <- scanner.Text()
	}
}

func waitForLog(sub chan string) tea.Cmd {
	return func() tea.Msg { return logLineMsg(<-sub) }
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickStatus(),
		triggerConnCheck(),
		tickConn(m.config.PingInterval),
		waitForLog(m.logChan),
		fetchServe(),
	)
}

func tickStatus() tea.Cmd {
	return tea.Every(2*time.Second, func(t time.Time) tea.Msg { return tickStatusMsg(t) })
}

func tickConn(interval int) tea.Cmd {
	if interval <= 0 { interval = 15 }
	return tea.Every(time.Duration(interval)*time.Second, func(t time.Time) tea.Msg { return tickConnMsg(t) })
}

func triggerConnCheck() tea.Cmd {
	return func() tea.Msg { return tickConnMsg(time.Now()) }
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd = exec.Command("wl-copy")
	} else if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else {
		return fmt.Errorf("No clipboard tool found")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func getSparkline(latencies []float64) string {
	if len(latencies) == 0 { return "" }
	bars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var spark strings.Builder
	maxVisible := 8
	start := 0
	if len(latencies) > maxVisible { start = len(latencies) - maxVisible }
	for _, l := range latencies[start:] {
		idx := int(l / 25)
		if idx >= len(bars) { idx = len(bars) - 1 }
		if l <= 0 { idx = 0 }
		spark.WriteString(bars[idx])
	}
	return spark.String()
}

func getSSHPortCmd(host string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("ssh", "-G", host).Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(strings.ToLower(line), "port ") {
					port := strings.TrimSpace(strings.SplitN(line, " ", 2)[1])
					return portUpdateMsg{host: host, port: port}
				}
			}
		}
		return portUpdateMsg{host: host, port: "22"} 
	}
}

func sendFile(path, host string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("tailscale", "file", "cp", path, host+":")
		err := cmd.Run()
		if err != nil { return showNotifMsg("󰚌 Failed: " + err.Error()) }
		return showNotifMsg("󰄬 File sent to " + host)
	}
}

func getFileWait() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("tailscale", "file", "get", ".")
		err := cmd.Run()
		if err != nil { return showNotifMsg("󰚌 " + err.Error()) }
		return showNotifMsg("󰄬 File received in current directory.")
	}
}

func executeDaemonCmd(args ...string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("tailscale", args...)
		err := cmd.Run()
		if err != nil { return showNotifMsg(fmt.Sprintf("󰚌 Error: %v", err)) }
		return showNotifMsg(fmt.Sprintf("󰄬 Executed '%s'", strings.Join(args, " ")))
	}
}

func fetchServe() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tailscale", "serve", "status").Output()
		if err != nil { return serveStatusMsg("Serve Status Error: " + err.Error()) }
		return serveStatusMsg(string(out))
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case logLineMsg:
		m.logs = append(m.logs, string(msg))
		if len(m.logs) > 1000 { m.logs = m.logs[1:] }
		if m.activeTab == TabLogs { m.logsScroll = len(m.logs) }
		return m, waitForLog(m.logChan)

	case serveStatusMsg:
		m.serveStatus = string(msg)
		return m, nil

	case tea.KeyMsg:
		if m.isFileTransfer {
			switch msg.Type {
			case tea.KeyEsc:
				m.isFileTransfer = false
				m.fileInput.Blur()
				return m, nil
			case tea.KeyEnter:
				m.isFileTransfer = false
				path := m.fileInput.Value()
				m.fileInput.Blur()
				m.fileInput.SetValue("")
				if len(m.peers) > m.cursor {
					cmds = append(cmds, sendFile(path, m.peers[m.cursor].HostName))
					m.notifMsg = "󰑐 Sending file..."
				}
				return m, tea.Batch(cmds...)
			default:
				var cmd tea.Cmd
				m.fileInput, cmd = m.fileInput.Update(msg)
				return m, cmd
			}
		}

		if m.isSearching {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				m.isSearching = false
				m.searchInput.Blur()
				if m.cursor >= len(m.peers) && len(m.peers) > 0 { m.cursor = len(m.peers) - 1 } else if len(m.peers) == 0 { m.cursor = 0 }
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.peers = m.filterAndSortPeers()
				if m.cursor >= len(m.peers) && len(m.peers) > 0 { m.cursor = len(m.peers) - 1 } else if len(m.peers) == 0 { m.cursor = 0 }
				return m, cmd
			}
		}

		if m.isDetailView {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.isDetailView = false
				return m, nil
			case "s":
				if len(m.peers) > m.cursor {
					p := m.peers[m.cursor]
					m.sshTarget = m.getSSHCommandTarget(p)
					return m, tea.Quit
				}
			case "a":
				if len(m.peers) > m.cursor {
					p := m.peers[m.cursor]
					cmds = append(cmds, executeDaemonCmd("up", "--accept-routes"))
					cmds = append(cmds, func()tea.Msg{ return showNotifMsg("󰄬 Accepting Routes (" + p.HostName + " etc..)") })
					return m, tea.Batch(cmds...)
				}
			case "g":
				cmds = append(cmds, getFileWait())
				return m, tea.Batch(cmds...)
			case "f":
				m.isFileTransfer = true
				m.fileInput.Focus()
				return m, textinput.Blink
			case "w": // Wake-on-LAN (Local or via Proxy)
				if len(m.peers) > m.cursor {
					p := m.peers[m.cursor]
					macStr := m.config.MacAddresses[p.HostName]
					if macStr == "" {
						cmds = append(cmds, func() tea.Msg { return showNotifMsg("󰚌 No MAC defined in config.yml for " + p.HostName) })
					} else {
						if m.config.WolProxy != "" && m.config.WolProxy != m.status.Self.HostName {
							// Execute Python code over SSH to send magic packet
							pyCode := fmt.Sprintf(`import socket; socket.socket(socket.AF_INET, socket.SOCK_DGRAM).sendto(b'\xff'*6 + bytes.fromhex('%s'.replace(':', '').replace('-', '')*16), ('255.255.255.255', 9))`, macStr)
							cmds = append(cmds, func() tea.Msg {
								cmd := exec.Command("ssh", "-o", "ConnectTimeout=3", m.config.WolProxy, "python3", "-c", pyCode)
								if err := cmd.Run(); err != nil {
									return showNotifMsg("󰚌 WoL Proxy error: " + err.Error())
								}
								return showNotifMsg("󰄬 Proxy-Sent Magic Packet to " + p.HostName)
							})
						} else {
							// Local UDP
							err := sendMagicPacket(macStr)
							if err != nil { cmds = append(cmds, func() tea.Msg { return showNotifMsg("󰚌 WoL error: " + err.Error()) }) } else { cmds = append(cmds, func() tea.Msg { return showNotifMsg("󰄬 Local-Sent Magic Packet to " + p.HostName) }) }
						}
					}
					return m, tea.Batch(cmds...)
				}
			}
			return m, nil
		}

		if m.activeTab == TabLogs {
			switch msg.String() {
			case "q", "ctrl+c": return m, tea.Quit
			case "tab": m.activeTab = TabDaemon; return m, nil
			case "shift+tab": m.activeTab = TabServe; return m, nil
			case "up", "k": if m.logsScroll > 0 { m.logsScroll-- }
			case "down", "j": if m.logsScroll < len(m.logs) { m.logsScroll++ }
			case "pgup": m.logsScroll -= 20; if m.logsScroll < 0 { m.logsScroll = 0 }
			case "pgdown": m.logsScroll += 20; if m.logsScroll > len(m.logs) { m.logsScroll = len(m.logs) }
			}
			return m, nil
		}

		if m.activeTab == TabServe {
			switch msg.String() {
			case "q", "ctrl+c": return m, tea.Quit
			case "tab": m.activeTab = TabLogs; return m, fetchServe()
			case "shift+tab": m.activeTab = TabExitNodes; return m, fetchServe()
			case "r": return m, fetchServe()
			}
			return m, nil
		}

		if m.activeTab == TabDaemon {
			switch msg.String() {
			case "q", "ctrl+c": return m, tea.Quit
			case "tab": m.activeTab = TabDevices; m.peers = m.filterAndSortPeers(); return m, nil
			case "shift+tab": m.activeTab = TabLogs; return m, nil
			case "u": cmds = append(cmds, executeDaemonCmd("up"))
			case "d": cmds = append(cmds, executeDaemonCmd("down"))
			case "s": cmds = append(cmds, executeDaemonCmd("set", "--shields-up")) // Shields Up
			case "S": cmds = append(cmds, executeDaemonCmd("set", "--shields-up=false")) // Shields Down
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "ctrl+c": return m, tea.Quit
		case "up", "k": if m.cursor > 0 { m.cursor-- }
		case "down", "j": if m.cursor < len(m.peers)-1 { m.cursor++ }
		case "pgup": m.cursor -= 10; if m.cursor < 0 { m.cursor = 0 }
		case "pgdown": m.cursor += 10; if m.cursor >= len(m.peers) { m.cursor = len(m.peers) - 1 }
		case "tab":
			m.activeTab = TabMode((int(m.activeTab) + 1) % 5)
			m.peers = m.filterAndSortPeers()
			if m.cursor >= len(m.peers) && len(m.peers) > 0 { m.cursor = len(m.peers) - 1 } else if len(m.peers) == 0 { m.cursor = 0 }
			if m.activeTab == TabServe { cmds = append(cmds, fetchServe()) }
		case "shift+tab":
			m.activeTab = TabMode((int(m.activeTab) + 4) % 5) 
			m.peers = m.filterAndSortPeers()
			if m.cursor >= len(m.peers) && len(m.peers) > 0 { m.cursor = len(m.peers) - 1 } else if len(m.peers) == 0 { m.cursor = 0 }
			if m.activeTab == TabServe { cmds = append(cmds, fetchServe()) }
		case "o": m.showOnlineOnly = !m.showOnlineOnly; m.peers = m.filterAndSortPeers(); if m.cursor >= len(m.peers) && len(m.peers) > 0 { m.cursor = len(m.peers) - 1 } else if len(m.peers) == 0 { m.cursor = 0 }
		case "s": m.sortMode = SortMode((int(m.sortMode) + 1) % 4); m.peers = m.filterAndSortPeers()
		case "/": m.isSearching = true; m.searchInput.Focus(); return m, textinput.Blink
		case "r": cmds = append(cmds, triggerConnCheck()); m.notifMsg = "󰑐 Refreshing network status..."; cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} }))
		case "c":
			if len(m.peers) > m.cursor {
				p := m.peers[m.cursor]
				if len(p.TailscaleIPs) > 0 { _ = copyToClipboard(p.TailscaleIPs[0]); m.notifMsg = "󰅍 COPIED IP: " + p.TailscaleIPs[0]; cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} })) }
			}
		case "S": 
			if len(m.peers) > m.cursor {
				p := m.peers[m.cursor]
				sshCmd := "ssh " + m.getSSHCommandTarget(p)
				_ = copyToClipboard(sshCmd); m.notifMsg = "󰆍 COPIED CMD: " + sshCmd; cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} }))
			}
		case "enter": if len(m.peers) > m.cursor { m.isDetailView = true; return m, nil }
		case "E":
			if m.activeTab == TabExitNodes && len(m.peers) > m.cursor {
				p := m.peers[m.cursor]
				if len(p.TailscaleIPs) > 0 { cmds = append(cmds, executeDaemonCmd("set", "--exit-node="+p.TailscaleIPs[0])) }
			}
		}
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case showNotifMsg:
		m.notifMsg = string(msg)
		return m, tea.Tick(time.Second*4, func(t time.Time) tea.Msg { return clearNotifMsg{} })

	case clearNotifMsg:
		m.notifMsg = ""
		return m, nil

	case portUpdateMsg:
		m.sshPorts[msg.host] = msg.port
		return m, nil

	case pingResultMsg:
		if m.pings[msg.host] == nil { m.pings[msg.host] = &pingInfo{} }
		m.pings[msg.host].current = msg.latency
		m.pings[msg.host].derp = msg.derp
		m.pings[msg.host].canSSH = msg.canSSH
		m.pings[msg.host].latency = append(m.pings[msg.host].latency, msg.latency)
		if len(m.pings[msg.host].latency) > 20 { m.pings[msg.host].latency = m.pings[msg.host].latency[1:] }
		if m.sortMode == SortByPing && !m.isDetailView && !m.isSearching && (m.activeTab == TabDevices || m.activeTab == TabExitNodes) {
			m.peers = m.filterAndSortPeers()
		}
		return m, nil

	case tickStatusMsg:
		out, err := exec.Command("tailscale", "status", "--json").Output()
		if err != nil {
			m.err = err; m.errMessage = "Failed to run 'tailscale status'. Is Tailscale installed and running?"
			cmds = append(cmds, tickStatus()); return m, tea.Batch(cmds...)
		}
		var newStatus TailscaleStatus
		if err := json.Unmarshal(out, &newStatus); err != nil {
			m.err = err; m.errMessage = "Failed to parse JSON."
			cmds = append(cmds, tickStatus()); return m, tea.Batch(cmds...)
		}
		
		for _, p := range newStatus.Peer {
			prev, known := m.lastOnline[p.HostName]
			if known {
				if prev && !p.Online { _ = beeep.Notify("Tailpuls Alert", p.HostName+" has gone OFFLINE", "") } else if !prev && p.Online { _ = beeep.Notify("Tailpuls Alert", p.HostName+" is now ONLINE", "") }
			}
			m.lastOnline[p.HostName] = p.Online
		}

		m.status = newStatus
		m.err = nil
		m.errMessage = ""
		if !m.isDetailView && !m.isSearching && (m.activeTab == TabDevices || m.activeTab == TabExitNodes) {
			m.peers = m.filterAndSortPeers()
			if m.cursor >= len(m.peers) { if len(m.peers) > 0 { m.cursor = len(m.peers) - 1 } else { m.cursor = 0 } }
		}
		cmds = append(cmds, tickStatus())
		return m, tea.Batch(cmds...)

	case tickConnMsg:
		for _, p := range m.peers {
			if len(p.TailscaleIPs) > 0 && (p.Online || p.Active) {
				host := p.HostName
				ip := p.TailscaleIPs[0]
				port := m.config.Ports[host]
				if port == "" {
					if _, val := m.sshPorts[host]; !val { cmds = append(cmds, getSSHPortCmd(host)) }
					if p, ok := m.sshPorts[host]; ok && p != "" { port = p } else { port = "22" }
				}
				cmds = append(cmds, func() tea.Msg {
					start := time.Now()
					// 1. Tailing ping provides DERP stats.
					lat := 0.0
					derpStr := ""
					pCmd := exec.Command("tailscale", "ping", "-c", "1", "--timeout=1s", ip)
					out, errPing := pCmd.Output()
					if errPing == nil {
						lat = float64(time.Since(start).Milliseconds())
						outStr := string(out)
						if strings.Contains(outStr, "via DERP") {
							parts := strings.Split(outStr, "via DERP")
							if len(parts) > 1 {
								reg := strings.SplitN(parts[1], ")", 2)[0]
								derpStr = "DERP" + reg + ")"
							} else { derpStr = "DERP" }
						} else if strings.Contains(outStr, "via ") { derpStr = "Direct" }
					}

					timeout := time.Second * 1
					conn, sshErr := net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
					canSSH := false
					if sshErr == nil { canSSH = true; conn.Close() }
					return pingResultMsg{host: host, latency: lat, derp: derpStr, canSSH: canSSH}
				})
			}
		}
		cmds = append(cmds, tickConn(m.config.PingInterval))
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m model) getSSHCommandTarget(p PeerStatus) string {
	if len(p.TailscaleIPs) == 0 { return "" }
	ip := p.TailscaleIPs[0]
	port := m.config.Ports[p.HostName]
	if port == "" { port = m.sshPorts[p.HostName] }
	if port != "" && port != "22" { return fmt.Sprintf("%s -p %s", ip, port) }
	return ip
}

func (m model) filterAndSortPeers() []PeerStatus {
	var pList []PeerStatus
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))

	if m.activeTab != TabExitNodes {
		if !m.showOnlineOnly || m.status.Self.Online || true {
			match := true
			if query != "" {
				hostMatch := strings.Contains(strings.ToLower(m.status.Self.HostName), query)
				ipMatch := len(m.status.Self.TailscaleIPs) > 0 && strings.Contains(m.status.Self.TailscaleIPs[0], query)
				match = hostMatch || ipMatch
			}
			if match && m.status.Self.HostName != "" { pList = append(pList, m.status.Self) }
		}
	}

	for _, p := range m.status.Peer {
		if m.showOnlineOnly && !p.Online { continue }
		if m.activeTab == TabExitNodes && !p.ExitNodeOption && !p.ExitNode { continue }
		if query != "" {
			if !strings.Contains(strings.ToLower(p.HostName), query) && (len(p.TailscaleIPs) == 0 || !strings.Contains(p.TailscaleIPs[0], query)) { continue }
		}
		pList = append(pList, p)
	}

	sort.Slice(pList, func(i, j int) bool {
		switch m.sortMode {
		case SortByIP:
			ip1, ip2 := "", ""
			if len(pList[i].TailscaleIPs) > 0 { ip1 = pList[i].TailscaleIPs[0] }
			if len(pList[j].TailscaleIPs) > 0 { ip2 = pList[j].TailscaleIPs[0] }
			return ip1 < ip2
		case SortByOS:
			if pList[i].OS == pList[j].OS { return pList[i].HostName < pList[j].HostName }
			return pList[i].OS < pList[j].OS
		case SortByPing:
			valI, valJ := 0.0, 0.0
			if info, ok := m.pings[pList[i].HostName]; ok { valI = info.current }
			if info, ok := m.pings[pList[j].HostName]; ok { valJ = info.current }
			if valI == valJ { return pList[i].HostName < pList[j].HostName }
			if valI == 0 { return false }
			if valJ == 0 { return true }
			return valI < valJ
		default: 
			return pList[i].HostName < pList[j].HostName
		}
	})

	return pList
}

func getOSIcon(osName string) string {
	switch strings.ToLower(osName) {
	case "linux": return ""
	case "windows": return ""
	case "macos", "darwin": return ""
	case "android": return ""
	case "ios": return "🍎"
	default: return "󰌢"
	}
}

func (m model) viewTabs() string {
	tabs := []string{"Devices", "Exit Nodes", "Serve", "Logs", "Daemon"}
	var rendered []string
	for i, t := range tabs {
		if m.activeTab == TabMode(i) { rendered = append(rendered, activeTabStyle.Render(t)) } else { rendered = append(rendered, inactiveTabStyle.Render(t)) }
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...) + "\n"
}

func (m model) viewDetail() string {
	if len(m.peers) <= m.cursor { return "Invalid peer selected." }
	p := m.peers[m.cursor]

	var s strings.Builder
	s.WriteString(titleStyle.Render("󰒄 CTOS // DETAILED_INFO // " + p.HostName))
	s.WriteString("\n\n")

	ipList := "N/A"
	if len(p.TailscaleIPs) > 0 { ipList = strings.Join(p.TailscaleIPs, ", ") }
	routeStr := "None"
	if len(p.PrimaryRoutes) > 0 { routeStr = strings.Join(p.PrimaryRoutes, ", ") }
	
	macStr := "Not Configured"
	if m.config.MacAddresses[p.HostName] != "" { macStr = m.config.MacAddresses[p.HostName] }

	s.WriteString(fmt.Sprintf("%-16s: %s\n", "DNSName", p.DNSName))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Tailnet IPs", ipList))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "OS", p.OS))
	s.WriteString(fmt.Sprintf("%-16s: %t\n", "Online", p.Online))
	s.WriteString(fmt.Sprintf("%-16s: %t\n", "TailscaleSSH", p.TailscaleSSHEnabled))

	expiryStr := p.KeyExpiry.Format("2006-01-02 15:04")
	if time.Until(p.KeyExpiry) < 14*24*time.Hour && !p.KeyExpiry.IsZero() {
		expiryStr = logErrStyle.Render(expiryStr + " (EXPIRES SOON)")
	}
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Key Expiry", expiryStr))

	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Subnet Routes", routeStr))
	s.WriteString(fmt.Sprintf("%-16s: Tx: %s / Rx: %s\n", "Bandwidth", formatBytes(p.TxBytes), formatBytes(p.RxBytes)))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "WoL MAC Address", macStr))

	if len(p.Tags) > 0 {
		s.WriteString(fmt.Sprintf("%-16s: %s\n", "Tags", strings.Join(p.Tags, ", ")))
	} else { s.WriteString(fmt.Sprintf("%-16s: %s\n", "Tags", "None")) }
	s.WriteString("\n\n")

	if m.isFileTransfer {
		s.WriteString(highlightStyle.Render("File Path to Send: ") + m.fileInput.View() + "\n")
	} else if m.notifMsg != "" {
		s.WriteString(notifyStyle.Render(m.notifMsg) + "\n")
	} else { s.WriteString("\n") }

	s.WriteString(footerStyle.Render("\n 󰌌 [s]:SSH | [g]:Get File | [f]:Send File | [a]:Accept Routes | [w]:WoL | [Esc]:Back\n"))
	return s.String()
}

func (m model) viewDaemon() string {
	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("󰒄 TAILSCALE DAEMON STATUS"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("%-16s: %s\n", "HostName", m.status.Self.HostName))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Backend State", "Running")) 
	
	ips := "N/A"
	if len(m.status.Self.TailscaleIPs) > 0 { ips = strings.Join(m.status.Self.TailscaleIPs, ", ") }
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Local IPs", ips))

	s.WriteString("\n\n")
	if m.notifMsg != "" {
		s.WriteString(notifyStyle.Render(m.notifMsg) + "\n")
	} else { s.WriteString("\n") }

	s.WriteString(footerStyle.Render("\n 󰌌 [u]:Tailscale Up | [d]:Tailscale Down | [s]:Shields UP | [S]:Shields DOWN | [Tab]:Switch Tab | [q]:Quit\n"))
	return s.String()
}

func (m model) viewServe() string {
	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("󰒄 TAILSCALE SERVE & FUNNEL"))
	s.WriteString("\n\n")
	if m.serveStatus == "" { s.WriteString("Fetching serve status...\n") } else { s.WriteString(m.serveStatus) }
	s.WriteString(footerStyle.Render("\n 󰌌 [r]:Refresh | [Tab]:Switch Tab | [q]:Quit\n"))
	return s.String()
}

func (m model) viewLogs() string {
	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("󰒄 TAILSCALED LIVE LOGS"))
	s.WriteString("\n\n")

	headerHeight := 6
	footerHeight := 2
	viewHeight := m.height
	if viewHeight == 0 { viewHeight = 24 }
	visibleRows := viewHeight - headerHeight - footerHeight
	if visibleRows < 1 { visibleRows = 3 }

	start := m.logsScroll - visibleRows
	if start < 0 { start = 0 }
	end := start + visibleRows
	if end > len(m.logs) { end = len(m.logs) }

	for i := start; i < end; i++ {
		line := m.logs[i]
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") || strings.Contains(lower, "fail") || strings.Contains(lower, "dropped") {
			s.WriteString(logErrStyle.Render(line) + "\n")
		} else if strings.Contains(lower, "derp") || strings.Contains(lower, "fallback") {
			s.WriteString(logWarnStyle.Render(line) + "\n")
		} else { s.WriteString(logInfoStyle.Render(line) + "\n") }
	}

	for empty := 0; empty < visibleRows-(end-start); empty++ { s.WriteString("\n") }
	s.WriteString(footerStyle.Render("\n 󰌌 [j/k/PgDn/PgUp]:Scroll | [Tab]:Switch Tab | [q]:Quit\n"))
	return s.String()
}

func (m model) View() string {
	if m.err != nil { return fmt.Sprintf("󰚌 Error connecting to Tailscale:\n\n%v\n\n%s\n\nPress 'q' to quit.", m.err, m.errMessage) }
	if m.status.Self.HostName == "" && len(m.status.Peer) == 0 { return "󱚽 Waiting for Tailscale status details..." }
	if m.isDetailView { return m.viewDetail() }

	switch m.activeTab {
	case TabServe: return m.viewServe()
	case TabLogs: return m.viewLogs()
	case TabDaemon: return m.viewDaemon()
	}

	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString(titleStyle.Render("󰒄 CTOS // TAILNET_MONITOR // v4.0.0"))
	s.WriteString("\n")

	filterStatus := "ALL"
	if m.showOnlineOnly { filterStatus = "ONLINE ONLY" }
	searchHint := ""
	if m.searchInput.Value() != "" { searchHint = fmt.Sprintf(" | Search: '%s'", m.searchInput.Value()) }
	s.WriteString(fmt.Sprintf("Filter: %s (press 'o') | Sort: %v (press 's') %s\n", filterStatus, m.sortMode, searchHint))

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		"  ", colHost.Render("HOSTNAME"), colIP.Render("IP"), colOS.Render("OS"),
		colStatus.Render("STATUS"), colSSH.Render("SSH"), colConn.Render("CONN_TYPE"), colPing.Render("PING"),
	)
	s.WriteString(headerStyle.Render(header) + "\n")

	headerHeight := 6
	footerHeight := 3
	if m.isSearching { footerHeight = 4 }
	viewHeight := m.height
	if viewHeight == 0 { viewHeight = 24 }
	visibleRows := viewHeight - headerHeight - footerHeight
	if visibleRows < 1 { visibleRows = 3 }

	if m.cursor < m.viewportStart { m.viewportStart = m.cursor }
	if m.cursor >= m.viewportStart+visibleRows { m.viewportStart = m.cursor - visibleRows + 1 }
	end := m.viewportStart + visibleRows
	if end > len(m.peers) { end = len(m.peers) }

	for i := m.viewportStart; i < end; i++ {
		p := m.peers[i]
		cursorStr := "  "
		if m.cursor == i { cursorStr = lipgloss.NewStyle().Foreground(red).Render("󰁔 ") }

		statusIcon := "󰄱"
		statusText := "OFFLINE"
		rowStyle := offlineStyle
		if p.Online || p.Active {
			statusIcon = "󰄬"
			statusText = "ONLINE"
			rowStyle = onlineStyle
		}

		name := p.HostName
		if p.HostName == m.status.Self.HostName { name = ">> " + name } else { name = "  " + name }

		ip := "n/a"
		if len(p.TailscaleIPs) > 0 { ip = p.TailscaleIPs[0] }

		activeSSHIcon := iconSSHDefault
		if p.TailscaleSSHEnabled { activeSSHIcon = iconTailscaleSSH }
		
		sshIcon := lipgloss.NewStyle().Foreground(darkGrey).Render(activeSSHIcon)
		if info, ok := m.pings[p.HostName]; ok && info.canSSH {
			sshIcon = lipgloss.NewStyle().Foreground(yellow).Render(activeSSHIcon)
		}

		connType := "----"
		connStyle := offlineStyle
		if p.Active {
			if p.Relay != "" {
				connType = fmt.Sprintf("󰇚 %s", p.Relay)
				connStyle = relayStyle
			} else if p.CurAddr != "" {
				connType = "󰄘 Direct"
				connStyle = directStyle
			}
		}

		pingDisp := "---"
		if info, ok := m.pings[p.HostName]; ok {
			spark := getSparkline(info.latency)
			if info.derp != "" {
				pingDisp = fmt.Sprintf("%3.0fms%s %s", info.current, spark, info.derp)
			} else if info.current > 0 {
				pingDisp = fmt.Sprintf("%3.0fms%s", info.current, spark)
			} else if p.Online || p.Active { pingDisp = "timeout" }
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			cursorStr, colHost.Render(name), colIP.Render(ip), colOS.Render(getOSIcon(p.OS)),
			rowStyle.Render(colStatus.Render(statusIcon+" "+statusText)), colSSH.Render(sshIcon),
			connStyle.Render(colConn.Render(connType)), pingStyle.Render(colPing.Render(pingDisp)),
		)

		if m.cursor == i { s.WriteString(selectedStyle.Render(row) + "\n") } else { s.WriteString(row + "\n") }
	}

	for empty := 0; empty < visibleRows-(end-m.viewportStart); empty++ { s.WriteString("\n") }
	if m.isSearching {
		s.WriteString("\n" + highlightStyle.Render("Search: ") + m.searchInput.View())
	} else if m.notifMsg != "" {
		s.WriteString("\n" + notifyStyle.Render(m.notifMsg))
	} else { s.WriteString("\n") }

	scrollIndicator := " "
	if end < len(m.peers) { scrollIndicator = " ↓" }
	if m.viewportStart > 0 { scrollIndicator = " ↑" + scrollIndicator }

	var exitOpt string
	if m.activeTab == TabExitNodes { exitOpt = " | [E]:Use ExitNode" }
	footer := fmt.Sprintf("\n 󰌌 [j/k]:Move%s | [/]:Search | [Enter]:Detail View%s | [Tab]:Switch Tab | [q]:Quit\n", scrollIndicator, exitOpt)
	s.WriteString(footerStyle.Render(footer))

	return s.String()
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen()) 
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Fatal: %v", err)
		os.Exit(1)
	}

	updatedModel := finalModel.(model)
	if updatedModel.sshTarget != "" {
		fmt.Printf("󰆍 Connecting to %s via SSH...\n", updatedModel.sshTarget)
		parts := strings.Split(updatedModel.sshTarget, " ")
		cmdArgs := parts
		if len(parts) > 1 { cmdArgs = parts[0:] }
		
		cmd := exec.Command("ssh", cmdArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}
}