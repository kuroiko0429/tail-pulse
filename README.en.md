# Tail-Pulse

A Watch Dogs-inspired TUI monitoring tool for keeping an eye on all devices in your Tailscale network.

## Features

- **Real-time monitoring**: Fetches Tailscale status every 3 seconds and displays the state of all nodes.
- **Connectivity checks**:
  - Automatic Ping (latency) measurement to each node.
  - Sparkline mini-graphs to visualize recent connection quality.
  - Port scanning: SSH(22) / Web(80,443) / RDP(3389) / VNC(5900) auto-detection.
- **UI / Design**:
  - Stylish TUI built with `Charmbracelet Bubble Tea / Lip Gloss`.
  - Per-OS icon display.
  - Direct vs. relay connection indicator.
  - Node detail panel (port scan results, DNSName, tags, routes).
  - Hack animation on SSH connect.
- **SSH integration**:
  - Select a node and press `Enter` to SSH in immediately.
  - One-key copy of IP address or Taildrop command to clipboard.
- **Node search**: Press `/` to filter nodes by hostname or IP in real time.

## Usage

### Keybindings

| Key | Action |
| :--- | :--- |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `/` | Search nodes |
| `d` | Toggle detail panel |
| `c` | Copy selected node's Tailscale IP |
| `t` | Copy Taildrop command (`tailscale file cp <file> <hostname>:`) |
| `Enter` | SSH into selected node |
| `q` / `Ctrl+C` | Quit |

### Install

**Using go install** (Go 1.21+)

```bash
go install github.com/kuroiko0429/tail-pulse@latest
```

**Using a prebuilt binary**

Download the binary for your OS from the [Releases](https://github.com/kuroiko0429/tail-pulse/releases) page and make it executable.

```bash
chmod +x tail-pulse
./tail-pulse
```

**Running from source**

```bash
git clone https://github.com/kuroiko0429/tail-pulse.git
cd tail-pulse
go run main.go
```

## Requirements

### System

- **Tailscale CLI**: `tailscale` must be available in your PATH.
- **Clipboard tool**: One of the following is required for copy functionality:
  - `wl-copy` (Wayland)
  - `pbcopy` (macOS)
- **Nerd Fonts**: Recommended for correct icon rendering.

### Go packages

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Styling and layout
- `github.com/charmbracelet/bubbles` — TUI components

## License

MIT
