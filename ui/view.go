package ui

import (
	"fmt"
	"math/rand"
	"strings"

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

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("󰒄 CTOS // TAILNET_MONITOR_ADVANCED // v2.0.0"))
	s.WriteString("\n")

	// Search bar
	if m.isSearching || m.search.Value() != "" {
		s.WriteString(m.search.View() + "\n\n")
	} else {
		s.WriteString(footerStyle.Render("Press '/' to search, 'd' for details, 't' for taildrop cmd") + "\n\n")
	}

	// Main Layout
	var left string
	var right string

	left = m.renderList()

	if m.isDetail && len(m.peers) > m.cursor {
		right = m.renderDetail(m.peers[m.cursor])
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right))
	} else {
		s.WriteString(left)
	}

	s.WriteString("\n")
	if m.notifMsg != "" {
		s.WriteString(notifyStyle.Render(m.notifMsg) + "\n")
	}

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
		if l <= 0 { idx = 0 } 
		spark.WriteString(bars[idx])
	}
	return spark.String()
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

func (m Model) renderList() string {
	var s strings.Builder

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		"  ",
		colHost.Render("HOSTNAME"),
		colIP.Render("IP"),
		colOS.Render("OS"),
		colStatus.Render("STATUS"),
		colPorts.Render("PORTS"),
		colConn.Render("CONN"),
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
		statusText := "OFF"
		rStyle := offlineStyle
		if p.Online || p.Active {
			statusIcon = "󰄬"
			statusText = "ON"
			rStyle = onlineStyle
		}

		if !p.Online && m.cfg.CyberGlitch && rand.Float32() < 0.05 {
			statusText = randomString([]rune("01"), 3)
		}

		name := p.HostName
		if p.IsSelf {
			name = ">> " + name
		}

		ip := "n/a"
		if len(p.TailscaleIPs) > 0 {
			ip = p.TailscaleIPs[0]
		}

		connType := "----"
		cStyle := offlineStyle
		if p.Active {
			if p.Relay != "" {
				connType = "󰇚 " + p.Relay
				cStyle = relayStyle
			} else if p.CurAddr != "" {
				connType = "󰄘 Dir"
				cStyle = directStyle
			}
		}

		pingDisp := "---"
		portsDisp := "-"
		
		info := m.netInfo[p.HostName]
		if info != nil {
			if info.Latency > 0 {
				spark := getSparkline(info.LatencyHist)
				pingDisp = fmt.Sprintf("%3.0fms %s", info.Latency, spark)
			} else if p.Online || p.Active {
				pingDisp = "timeout"
			}
			
			var open []string
			if info.OpenPorts[22] { open = append(open, "22") }
			if info.OpenPorts[80] || info.OpenPorts[443] { open = append(open, "Web") }
			if info.OpenPorts[3389] { open = append(open, "RDP") }
			if len(open) > 0 {
				portsDisp = strings.Join(open, ",")
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			cursorStr,
			colHost.Render(name),
			colIP.Render(ip),
			colOS.Render(getOSIcon(p.OS)),
			rStyle.Render(colStatus.Render(statusIcon+" "+statusText)),
			colPorts.Render(portsDisp),
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

func (m Model) renderDetail(p tailscale.PeerStatus) string {
	var sb strings.Builder
	
	sb.WriteString(titleStyle.Render(" NODE DETAILS ") + "\n\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", cyanStyle.Render("Hostname:"), p.HostName))
	sb.WriteString(fmt.Sprintf("%s %s\n", cyanStyle.Render("DNS Name:"), p.DNSName))
	sb.WriteString(fmt.Sprintf("%s %s\n", cyanStyle.Render("OS:      "), p.OS))
	
	if len(p.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("%s %s\n", cyanStyle.Render("Tags:    "), strings.Join(p.Tags, ", ")))
	}
	
	sb.WriteString(fmt.Sprintf("%s %v\n", cyanStyle.Render("Exit Node:"), p.ExitNodeOption))
	
	if len(p.PrimaryRoutes) > 0 {
		sb.WriteString(fmt.Sprintf("%s %s\n", cyanStyle.Render("Routes:  "), strings.Join(p.PrimaryRoutes, ", ")))
	}

	info := m.netInfo[p.HostName]
	if info != nil {
		sb.WriteString("\n" + cyanStyle.Render("[ Port Scan Results ]") + "\n")
		for port, open := range info.OpenPorts {
			state := "CLOSED"
			color := darkGrey
			if open {
				state = "OPEN"
				color = green
			}
			sb.WriteString(fmt.Sprintf(" %4d: %s\n", port, lipgloss.NewStyle().Foreground(color).Render(state)))
		}
	}

	return detailStyle.Render(sb.String())
}

func (m Model) GetSSHTarget() string {
	return m.sshTarget
}
