package ui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	
	"github.com/kuroiko0429/tail-pulse/config"
	"github.com/kuroiko0429/tail-pulse/network"
	"github.com/kuroiko0429/tail-pulse/tailscale"
)

type tickMsg time.Time
type clearNotifMsg struct{}
type pingResultMsg struct {
	host    string
	latency float64
	ports   map[int]bool
}

type Model struct {
	cfg            config.Config
	status         tailscale.Status
	peers          []tailscale.PeerStatus
	netInfo        map[string]*network.NodeNetInfo
	
	cursor         int
	search         textinput.Model
	isSearching    bool
	isDetail       bool
	
	notifMsg       string
	sshTarget      string
	hackAnimActive bool
	hackTicks      int
	err            error
	width          int
	height         int
}

func NewModel(cfg config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Search nodes..."
	ti.CharLimit = 156
	ti.Width = 30

	return Model{
		cfg:     cfg,
		netInfo: make(map[string]*network.NodeNetInfo),
		search:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, textinput.Blink, tick())
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

func (m *Model) filterPeers() {
	var list []tailscale.PeerStatus
	query := strings.ToLower(m.search.Value())

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

	case tea.KeyMsg:
		if m.hackAnimActive {
			return m, nil
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
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if !m.isSearching {
				return m, tea.Quit
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.peers)-1 {
				m.cursor++
			}
		case "/":
			if !m.isSearching {
				m.isSearching = true
				m.search.Focus()
				if m.search.Value() != "" {
					m.search.SetValue("")
					m.filterPeers()
				}
				return m, textinput.Blink
			}
		case "d":
			m.isDetail = !m.isDetail
		case "enter":
			if len(m.peers) > m.cursor {
				p := m.peers[m.cursor]
				if len(p.TailscaleIPs) > 0 {
					m.sshTarget = p.TailscaleIPs[0]
					m.hackAnimActive = true
					m.hackTicks = 0
					if m.cfg.CyberGlitch {
						return m, hackTick()
					}
					return m, tea.Quit
				}
			}
		case "c": // IP
			if len(m.peers) > m.cursor && len(m.peers[m.cursor].TailscaleIPs) > 0 {
				copyToClipboard(m.peers[m.cursor].TailscaleIPs[0])
				m.notifMsg = "󰅍 COPIED IP"
				return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} })
			}
		case "t": // copy taildrop command
			if len(m.peers) > m.cursor {
				cmdStr := fmt.Sprintf("tailscale file cp <file> %s:", m.peers[m.cursor].HostName)
				copyToClipboard(cmdStr)
				m.notifMsg = "󰅍 COPIED TAILDROP CMD"
				return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return clearNotifMsg{} })
			}
		case "n": // notify test
			_ = exec.Command("notify-send", "Tail-Pulse", "Notification test.").Run()
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
			m.status = status
			m.filterPeers()
			
			// Fire async network checks
			for _, p := range m.peers {
				if len(p.TailscaleIPs) > 0 && (p.Online || p.Active) {
					host := p.HostName
					ip := p.TailscaleIPs[0]
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
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
	}
}
