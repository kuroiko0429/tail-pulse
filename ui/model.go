package ui

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"

	"github.com/kuroiko0429/tail-pulse/config"
	"github.com/kuroiko0429/tail-pulse/network"
	"github.com/kuroiko0429/tail-pulse/tailscale"
)

type TabMode int

const (
	TabDevices TabMode = iota
	TabExitNodes
	TabServe
	TabLogs
	TabDaemon
	tabCount
)

func (t TabMode) String() string {
	switch t {
	case TabExitNodes:
		return "Exit Nodes"
	case TabServe:
		return "Serve"
	case TabLogs:
		return "Logs"
	case TabDaemon:
		return "Daemon"
	default:
		return "Devices"
	}
}

type tickMsg time.Time
type clearNotifMsg struct{}
type pingResultMsg struct {
	host    string
	latency float64
	ports   map[int]bool
}
type portUpdateMsg struct {
	host string
	port string
}
type logLineMsg string
type serveStatusMsg string

type Model struct {
	cfg        config.Config
	status     tailscale.Status
	peers      []tailscale.PeerStatus
	netInfo    map[string]*network.NodeNetInfo
	sshPorts   map[string]string
	lastOnline map[string]bool

	cursor      int
	search      textinput.Model
	isSearching bool
	isDetail    bool
	activeTab   TabMode

	notifMsg       string
	sshTarget      string
	hackAnimActive bool
	hackTicks      int
	err            error
	width          int
	height         int

	// Logs tab
	logs       []string
	logChan    chan string
	logsScroll int

	// Serve tab
	serveStatus string

	// File transfer
	isFileTransfer bool
	fileInput      textinput.Model
}

func NewModel(cfg config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Search nodes..."
	ti.CharLimit = 156
	ti.Width = 30

	fi := textinput.New()
	fi.Placeholder = "/path/to/file"
	fi.CharLimit = 256
	fi.Width = 50

	return Model{
		cfg:        cfg,
		netInfo:    make(map[string]*network.NodeNetInfo),
		sshPorts:   make(map[string]string),
		lastOnline: make(map[string]bool),
		search:     ti,
		fileInput:  fi,
		logChan:    make(chan string, 100),
	}
}

func (m Model) Init() tea.Cmd {
	go streamLogs(m.logChan)
	return tea.Batch(tea.ClearScreen, textinput.Blink, tick(), waitForLog(m.logChan))
}

func tick() tea.Cmd {
	return tea.Every(time.Second*3, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func hackTick() tea.Cmd {
	return tea.Every(time.Millisecond*50, func(t time.Time) tea.Msg {
		return "hack_tick"
	})
}

func streamLogs(sub chan<- string) {
	cmd := exec.Command("journalctl", "-u", "tailscaled", "-n", "100", "-f")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sub <- fmt.Sprintf("Failed to stream logs: %v", err)
		return
	}
	if err := cmd.Start(); err != nil {
		sub <- fmt.Sprintf("Failed to start journalctl: %v", err)
		return
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		sub <- scanner.Text()
	}
}

func waitForLog(sub chan string) tea.Cmd {
	return func() tea.Msg { return logLineMsg(<-sub) }
}

func fetchServeStatus() tea.Cmd {
	return func() tea.Msg {
		out, err := tailscale.ServeStatus()
		if err != nil {
			return serveStatusMsg("Serve Status Error: " + err.Error())
		}
		return serveStatusMsg(out)
	}
}

func resolveSSHPort(host string) tea.Cmd {
	return func() tea.Msg {
		return portUpdateMsg{host: host, port: tailscale.GetSSHPort(host)}
	}
}

func runDaemonCmd(notifyPrefix string, args ...string) tea.Cmd {
	return func() tea.Msg {
		if err := tailscale.DaemonCmd(args...); err != nil {
			return showNotif("󰚌 " + err.Error())
		}
		return showNotif("󰄬 " + notifyPrefix)
	}
}

func showNotif(msg string) tea.Msg {
	return notifMsgUpdate(msg)
}

type notifMsgUpdate string

func (m *Model) filterPeers() {
	var list []tailscale.PeerStatus
	query := strings.ToLower(m.search.Value())

	if m.activeTab == TabExitNodes {
		for _, p := range m.status.Peer {
			if !p.ExitNodeOption && !p.ExitNode {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(p.HostName), query) {
				continue
			}
			list = append(list, p)
		}
	} else {
		if m.status.Self.HostName != "" {
			if query == "" || strings.Contains(strings.ToLower(m.status.Self.HostName), query) {
				list = append(list, m.status.Self)
			}
		}
		for _, p := range m.status.Peer {
			if query == "" || strings.Contains(strings.ToLower(p.HostName), query) || strings.Contains(strings.ToLower(strings.Join(p.TailscaleIPs, " ")), query) {
				list = append(list, p)
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].HostName < list[j].HostName
	})
	m.peers = list

	if m.cursor >= len(m.peers) && len(m.peers) > 0 {
		m.cursor = len(m.peers) - 1
	} else if len(m.peers) == 0 {
		m.cursor = 0
	}
}

func (m *Model) selectedPeer() (tailscale.PeerStatus, bool) {
	if len(m.peers) > m.cursor {
		return m.peers[m.cursor], true
	}
	return tailscale.PeerStatus{}, false
}

func (m *Model) sshTargetFor(p tailscale.PeerStatus) string {
	if len(p.TailscaleIPs) == 0 {
		return ""
	}
	ip := p.TailscaleIPs[0]
	port := m.cfg.Ports[p.HostName]
	if port == "" {
		port = m.sshPorts[p.HostName]
	}
	if port != "" && port != "22" {
		return fmt.Sprintf("%s -p %s", ip, port)
	}
	return ip
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case string:
		if msg == "hack_tick" {
			m.hackTicks++
			if m.hackTicks > 15 {
				return m, tea.Quit
			}
			return m, hackTick()
		}

	case logLineMsg:
		m.logs = append(m.logs, string(msg))
		if len(m.logs) > 1000 {
			m.logs = m.logs[1:]
		}
		if m.activeTab == TabLogs {
			m.logsScroll = len(m.logs)
		}
		return m, waitForLog(m.logChan)

	case serveStatusMsg:
		m.serveStatus = string(msg)

	case notifMsgUpdate:
		m.notifMsg = string(msg)
		return m, tea.Tick(time.Second*4, func(t time.Time) tea.Msg { return clearNotifMsg{} })

	case portUpdateMsg:
		m.sshPorts[msg.host] = msg.port

	case tea.KeyMsg:
		if m.hackAnimActive {
			return m, nil
		}

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
				if p, ok := m.selectedPeer(); ok && path != "" {
					host := p.HostName
					cmds = append(cmds, func() tea.Msg {
						if err := tailscale.FileCp(path, host); err != nil {
							return notifMsgUpdate("󰚌 Failed: " + err.Error())
						}
						return notifMsgUpdate("󰄬 File sent to " + host)
					})
				}
				return m, tea.Batch(cmds...)
			default:
				m.fileInput, cmd = m.fileInput.Update(msg)
				return m, cmd
			}
		}

		if m.isSearching {
			switch msg.String() {
			case "enter", "esc":
				m.isSearching = false
				m.search.Blur()
			default:
				m.search, cmd = m.search.Update(msg)
				m.filterPeers()
				return m, cmd
			}
			return m, nil
		}

		if m.activeTab == TabLogs {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.activeTab = (m.activeTab + 1) % tabCount
				m.filterPeers()
				return m, nil
			case "shift+tab":
				m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
				m.filterPeers()
				return m, nil
			case "up", "k":
				if m.logsScroll > 0 {
					m.logsScroll--
				}
			case "down", "j":
				if m.logsScroll < len(m.logs) {
					m.logsScroll++
				}
			case "pgup":
				m.logsScroll -= 20
				if m.logsScroll < 0 {
					m.logsScroll = 0
				}
			case "pgdown":
				m.logsScroll += 20
				if m.logsScroll > len(m.logs) {
					m.logsScroll = len(m.logs)
				}
			}
			return m, nil
		}

		if m.activeTab == TabServe {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.activeTab = (m.activeTab + 1) % tabCount
				m.filterPeers()
				return m, fetchServeStatus()
			case "shift+tab":
				m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
				m.filterPeers()
				return m, fetchServeStatus()
			case "r":
				return m, fetchServeStatus()
			}
			return m, nil
		}

		if m.activeTab == TabDaemon {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.activeTab = (m.activeTab + 1) % tabCount
				m.filterPeers()
				return m, nil
			case "shift+tab":
				m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
				m.filterPeers()
				return m, nil
			case "u":
				return m, runDaemonCmd("tailscale up", "up")
			case "d":
				return m, runDaemonCmd("tailscale down", "down")
			case "s":
				return m, runDaemonCmd("Shields Up", "set", "--shields-up")
			case "S":
				return m, runDaemonCmd("Shields Down", "set", "--shields-up=false")
			}
			return m, nil
		}

		// Devices / Exit Nodes tabs
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.peers)-1 {
				m.cursor++
			}
		case "tab":
			m.activeTab = (m.activeTab + 1) % tabCount
			m.filterPeers()
			if m.activeTab == TabServe {
				return m, fetchServeStatus()
			}
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			m.filterPeers()
			if m.activeTab == TabServe {
				return m, fetchServeStatus()
			}
		case "/":
			m.isSearching = true
			m.search.Focus()
			if m.search.Value() != "" {
				m.search.SetValue("")
				m.filterPeers()
			}
			return m, textinput.Blink
		case "d":
			m.isDetail = !m.isDetail
		case "enter":
			if p, ok := m.selectedPeer(); ok && len(p.TailscaleIPs) > 0 {
				m.sshTarget = m.sshTargetFor(p)
				m.hackAnimActive = true
				m.hackTicks = 0
				if m.cfg.CyberGlitch {
					return m, hackTick()
				}
				return m, tea.Quit
			}
		case "c":
			if p, ok := m.selectedPeer(); ok && len(p.TailscaleIPs) > 0 {
				copyToClipboard(p.TailscaleIPs[0])
				m.notifMsg = "󰅍 COPIED IP"
				return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} })
			}
		case "t":
			if p, ok := m.selectedPeer(); ok {
				cmdStr := fmt.Sprintf("tailscale file cp <file> %s:", p.HostName)
				copyToClipboard(cmdStr)
				m.notifMsg = "󰅍 COPIED TAILDROP CMD"
				return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} })
			}
		case "T":
			m.isFileTransfer = true
			m.fileInput.Focus()
			return m, textinput.Blink
		case "g":
			return m, func() tea.Msg {
				if err := tailscale.FileGet(); err != nil {
					return notifMsgUpdate("󰚌 " + err.Error())
				}
				return notifMsgUpdate("󰄬 File(s) received in current directory")
			}
		case "a":
			if p, ok := m.selectedPeer(); ok {
				name := p.HostName
				return m, runDaemonCmd("Accepting routes ("+name+")", "up", "--accept-routes")
			}
		case "w":
			if p, ok := m.selectedPeer(); ok {
				mac := m.cfg.MacAddresses[p.HostName]
				if mac == "" {
					m.notifMsg = "󰚌 No MAC defined in config for " + p.HostName
					return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg { return clearNotifMsg{} })
				}
				name := p.HostName
				if m.cfg.WolProxy != "" && m.cfg.WolProxy != m.status.Self.HostName {
					proxy := m.cfg.WolProxy
					return m, func() tea.Msg {
						if err := network.WakeOnLanViaProxy(proxy, mac); err != nil {
							return notifMsgUpdate("󰚌 WoL Proxy error: " + err.Error())
						}
						return notifMsgUpdate("󰄬 Proxy-sent Magic Packet to " + name)
					}
				}
				return m, func() tea.Msg {
					if err := network.WakeOnLan(mac); err != nil {
						return notifMsgUpdate("󰚌 WoL error: " + err.Error())
					}
					return notifMsgUpdate("󰄬 Sent Magic Packet to " + name)
				}
			}
		case "E":
			if m.activeTab == TabExitNodes {
				if p, ok := m.selectedPeer(); ok && len(p.TailscaleIPs) > 0 {
					return m, runDaemonCmd("Exit node set to "+p.HostName, "set", "--exit-node="+p.TailscaleIPs[0])
				}
			}
		}

	case clearNotifMsg:
		m.notifMsg = ""

	case pingResultMsg:
		if m.netInfo[msg.host] == nil {
			m.netInfo[msg.host] = &network.NodeNetInfo{}
		}
		info := m.netInfo[msg.host]
		info.Latency = msg.latency
		info.OpenPorts = msg.ports
		info.LatencyHist = append(info.LatencyHist, msg.latency)
		if len(info.LatencyHist) > 20 {
			info.LatencyHist = info.LatencyHist[1:]
		}

	case tickMsg:
		status, err := tailscale.GetStatus()
		if err != nil {
			m.err = err
		} else {
			for _, p := range status.Peer {
				prev, known := m.lastOnline[p.HostName]
				if known {
					if prev && !p.Online {
						_ = beeep.Notify("Tail-Pulse", p.HostName+" has gone OFFLINE", "")
					} else if !prev && p.Online {
						_ = beeep.Notify("Tail-Pulse", p.HostName+" is now ONLINE", "")
					}
				}
				m.lastOnline[p.HostName] = p.Online
			}

			m.status = status
			m.err = nil
			m.filterPeers()

			for _, p := range m.peers {
				if len(p.TailscaleIPs) > 0 && (p.Online || p.Active) {
					host := p.HostName
					ip := p.TailscaleIPs[0]
					if _, cached := m.sshPorts[host]; !cached {
						cmds = append(cmds, resolveSSHPort(host))
					}
					cmds = append(cmds, func() tea.Msg {
						lat := network.Ping(ip)
						ports := network.CheckPorts(ip)
						return pingResultMsg{host: host, latency: lat, ports: ports}
					})
				}
			}
		}
		cmds = append(cmds, tick())
	}

	return m, tea.Batch(cmds...)
}

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch {
	case lookPath("wl-copy"):
		cmd = exec.Command("wl-copy")
	case lookPath("pbcopy"):
		cmd = exec.Command("pbcopy")
	case lookPath("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		return
	}
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}

func lookPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}
