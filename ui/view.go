package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kuroiko0429/tail-pulse/tailscale"
)

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if m.hackAnimActive {
		return m.renderHackAnim()
	}

	if m.isFileTransfer {
		return titleStyle.Render(" SEND FILE VIA TAILDROP ") + "\n\n" + m.fileInput.View() + "\n\n" + footerStyle.Render("[Enter]:Send  [Esc]:Cancel")
	}

	if m.isDetail {
		if p, ok := m.selectedPeer(); ok {
			return m.viewDetail(p)
		}
		m.isDetail = false
	}

	switch m.activeTab {
	case TabServe:
		return m.viewServe()
	case TabLogs:
		return m.viewLogs()
	case TabDaemon:
		return m.viewDaemon()
	}

	var s strings.Builder

	s.WriteString(m.viewTabs())
	s.WriteString(titleStyle.Render("󰒄 CTOS // TAILNET_MONITOR // v2.0.0"))
	s.WriteString("\n")

	// Search bar
	if m.isSearching || m.search.Value() != "" {
		s.WriteString(m.search.View() + "\n\n")
	} else {
		s.WriteString(footerStyle.Render(fmt.Sprintf("Sort: %v (press 's') | [/]search [Enter]detail [c]copy-ip [S]copy-ssh [t]taildrop-cmd [Tab]switch", m.sortMode)) + "\n\n")
	}

	s.WriteString(m.renderList())
	s.WriteString("\n")
	if m.notifMsg != "" {
		s.WriteString(notifyStyle.Render(m.notifMsg) + "\n")
	}

	return s.String()
}

func (m Model) viewTabs() string {
	names := []string{"Devices", "Exit Nodes", "Serve", "Logs", "Daemon"}
	var rendered []string
	for i, name := range names {
		if int(m.activeTab) == i {
			rendered = append(rendered, activeTabStyle.Render(name))
		} else {
			rendered = append(rendered, inactiveTabStyle.Render(name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...) + "\n"
}

func (m Model) viewServe() string {
	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString(titleStyle.Render("󰒄 TAILSCALE SERVE & FUNNEL"))
	s.WriteString("\n\n")
	if m.serveStatus == "" {
		s.WriteString("Fetching serve status...\n")
	} else {
		s.WriteString(m.serveStatus)
	}
	s.WriteString(footerStyle.Render("\n[r]refresh [Tab]switch [q]quit\n"))
	return s.String()
}

func (m Model) viewDaemon() string {
	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString(titleStyle.Render("󰒄 TAILSCALE DAEMON STATUS"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("%-14s: %s\n", "HostName", m.status.Self.HostName))
	ips := "N/A"
	if len(m.status.Self.TailscaleIPs) > 0 {
		ips = strings.Join(m.status.Self.TailscaleIPs, ", ")
	}
	s.WriteString(fmt.Sprintf("%-14s: %s\n", "Local IPs", ips))
	s.WriteString("\n")
	if m.notifMsg != "" {
		s.WriteString(notifyStyle.Render(m.notifMsg) + "\n\n")
	}
	s.WriteString(footerStyle.Render("[u]up [d]down [s]shields-up [S]shields-down [Tab]switch [q]quit\n"))
	return s.String()
}

func (m Model) viewLogs() string {
	var s strings.Builder
	s.WriteString(m.viewTabs())
	s.WriteString(titleStyle.Render("󰒄 TAILSCALED LIVE LOGS"))
	s.WriteString("\n\n")

	headerHeight := 5
	footerHeight := 2
	viewHeight := m.height
	if viewHeight == 0 {
		viewHeight = 24
	}
	visibleRows := viewHeight - headerHeight - footerHeight
	if visibleRows < 1 {
		visibleRows = 3
	}

	start := m.logsScroll - visibleRows
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > len(m.logs) {
		end = len(m.logs)
	}

	for i := start; i < end; i++ {
		line := m.logs[i]
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "error") || strings.Contains(lower, "fail") || strings.Contains(lower, "dropped"):
			s.WriteString(logErrStyle.Render(line) + "\n")
		case strings.Contains(lower, "derp") || strings.Contains(lower, "fallback"):
			s.WriteString(logWarnStyle.Render(line) + "\n")
		default:
			s.WriteString(logInfoStyle.Render(line) + "\n")
		}
	}

	s.WriteString(footerStyle.Render("\n[j/k, PgUp/PgDn]scroll [Tab]switch [q]quit\n"))
	return s.String()
}

func (m Model) renderHackAnim() string {
	chars := []rune("!@#$%^&*()_+-=[]{}|;':,./<>?")
	var sb strings.Builder
	sb.WriteString("\n  [ !! ] INITIATING SECURE SHELL BYPASS...\n\n")

	for i := 0; i < m.hackTicks; i++ {
		sb.WriteString("  > Decrypting key " + randomString(chars, 16) + " [OK]\n")
	}
	sb.WriteString("\n  CONNECTING TO " + m.sshTarget + " ...\n")
	return lipgloss.NewStyle().Foreground(green).Render(sb.String())
}

func randomString(charset []rune, length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getSparkline(latencies []float64) string {
	if len(latencies) == 0 {
		return ""
	}
	bars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var spark strings.Builder

	start := 0
	if len(latencies) > 10 {
		start = len(latencies) - 10
	}

	for _, l := range latencies[start:] {
		idx := int(l / 25)
		if idx >= len(bars) {
			idx = len(bars) - 1
		}
		if l <= 0 {
			idx = 0
		}
		spark.WriteString(bars[idx])
	}
	return spark.String()
}

func getOSIcon(osName string) string {
	switch strings.ToLower(osName) {
	case "linux":
		return ""
	case "windows":
		return ""
	case "macos", "darwin":
		return ""
	case "android":
		return ""
	case "ios":
		return "🍎"
	default:
		return "󰌢"
	}
}

const (
	iconSSHDefault   = "󱘖"
	iconTailscaleSSH = "󰣀"
)

func (m Model) renderList() string {
	var s strings.Builder

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		"  ",
		colHost.Render("HOSTNAME"),
		colIP.Render("IP"),
		colOS.Render("OS"),
		colStatus.Render("STATUS"),
		colSSH.Render("SSH"),
		colConn.Render("CONN_TYPE"),
		colPing.Render("PING"),
	)
	s.WriteString(headerStyle.Render(header) + "\n")

	for i, p := range m.peers {
		cursorStr := "  "
		rowBg := lipgloss.Color("")
		if m.cursor == i {
			cursorStr = lipgloss.NewStyle().Foreground(red).Render("󰁔 ")
			rowBg = darkGrey
		}

		statusIcon := "󰄱"
		statusText := "OFFLINE"
		rStyle := offlineStyle
		if p.Online || p.Active {
			statusIcon = "󰄬"
			statusText = "ONLINE"
			rStyle = onlineStyle
		}

		if !p.Online && m.cfg.CyberGlitch && rand.Float32() < 0.05 {
			statusText = randomString([]rune("01"), 3)
		}

		name := p.HostName
		if p.IsSelf {
			name = ">> " + name
		} else {
			name = "  " + name
		}

		ip := "n/a"
		if len(p.TailscaleIPs) > 0 {
			ip = p.TailscaleIPs[0]
		}

		activeSSHIcon := iconSSHDefault
		if p.TailscaleSSHEnabled {
			activeSSHIcon = iconTailscaleSSH
		}
		info := m.netInfo[p.HostName]
		sshIcon := lipgloss.NewStyle().Foreground(darkGrey).Render(activeSSHIcon)
		if info != nil && info.OpenPorts[22] {
			sshIcon = lipgloss.NewStyle().Foreground(yellow).Render(activeSSHIcon)
		}

		connType := "----"
		cStyle := offlineStyle
		if p.Active {
			if p.Relay != "" {
				connType = "󰇚 " + p.Relay
				cStyle = relayStyle
			} else if p.CurAddr != "" {
				connType = "󰄘 Direct"
				cStyle = directStyle
			}
		}

		pingDisp := "---"
		if info != nil {
			if info.Latency > 0 {
				spark := getSparkline(info.LatencyHist)
				pingDisp = fmt.Sprintf("%3.0fms %s", info.Latency, spark)
			} else if p.Online || p.Active {
				pingDisp = "timeout"
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			cursorStr,
			colHost.Render(name),
			colIP.Render(ip),
			colOS.Render(getOSIcon(p.OS)),
			rStyle.Render(colStatus.Render(statusIcon+" "+statusText)),
			colSSH.Render(sshIcon),
			cStyle.Render(colConn.Render(connType)),
			cyanStyle.Render(colPing.Render(pingDisp)),
		)

		if rowBg != "" {
			s.WriteString(lipgloss.NewStyle().Background(rowBg).Render(row) + "\n")
		} else {
			s.WriteString(row + "\n")
		}
	}

	return s.String()
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (m Model) viewDetail(p tailscale.PeerStatus) string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("󰒄 CTOS // DETAILED_INFO // " + p.HostName))
	s.WriteString("\n\n")

	ipList := "N/A"
	if len(p.TailscaleIPs) > 0 {
		ipList = strings.Join(p.TailscaleIPs, ", ")
	}
	routeStr := "None"
	if len(p.PrimaryRoutes) > 0 {
		routeStr = strings.Join(p.PrimaryRoutes, ", ")
	}
	macStr := "Not Configured"
	if m.cfg.MacAddresses[p.HostName] != "" {
		macStr = m.cfg.MacAddresses[p.HostName]
	}

	s.WriteString(fmt.Sprintf("%-16s: %s\n", "DNSName", p.DNSName))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Tailnet IPs", ipList))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "OS", p.OS))
	s.WriteString(fmt.Sprintf("%-16s: %t\n", "Online", p.Online))
	s.WriteString(fmt.Sprintf("%-16s: %t\n", "TailscaleSSH", p.TailscaleSSHEnabled))

	expiryStr := "N/A"
	if !p.KeyExpiry.IsZero() {
		expiryStr = p.KeyExpiry.Format("2006-01-02 15:04")
		if time.Until(p.KeyExpiry) < 14*24*time.Hour {
			expiryStr = logErrStyle.Render(expiryStr + " (EXPIRES SOON)")
		}
	}
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Key Expiry", expiryStr))

	s.WriteString(fmt.Sprintf("%-16s: %s\n", "Subnet Routes", routeStr))
	s.WriteString(fmt.Sprintf("%-16s: Tx: %s / Rx: %s\n", "Bandwidth", formatBytes(p.TxBytes), formatBytes(p.RxBytes)))
	s.WriteString(fmt.Sprintf("%-16s: %s\n", "WoL MAC Address", macStr))

	if len(p.Tags) > 0 {
		s.WriteString(fmt.Sprintf("%-16s: %s\n", "Tags", strings.Join(p.Tags, ", ")))
	} else {
		s.WriteString(fmt.Sprintf("%-16s: %s\n", "Tags", "None"))
	}

	if info := m.netInfo[p.HostName]; info != nil && len(info.OpenPorts) > 0 {
		s.WriteString("\n" + cyanStyle.Render("[ Port Scan Results ]") + "\n")
		for _, port := range []int{22, 80, 443, 3389, 5900} {
			state := "CLOSED"
			color := darkGrey
			if info.OpenPorts[port] {
				state = "OPEN"
				color = green
			}
			s.WriteString(fmt.Sprintf(" %4d: %s\n", port, lipgloss.NewStyle().Foreground(color).Render(state)))
		}
	}
	s.WriteString("\n\n")

	if m.notifMsg != "" {
		s.WriteString(notifyStyle.Render(m.notifMsg) + "\n")
	} else {
		s.WriteString("\n")
	}

	s.WriteString(footerStyle.Render("\n 󰌌 [s]:SSH | [g]:Get File | [f]:Send File | [a]:Accept Routes | [w]:WoL | [Esc]:Back\n"))
	return s.String()
}

func (m Model) GetSSHTarget() string {
	return m.sshTarget
}
