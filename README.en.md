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
  - Select a node and press `Enter` to open its detail screen, then `s` to SSH in (SSH port auto-detected via `ssh -G`).
  - One-key copy of IP address, SSH command, or Taildrop command to clipboard.
- **Node search**: Press `/` to filter nodes by hostname or IP in real time.
- **File transfer**: Send/receive files over Taildrop (from the detail screen: `f` to send, `g` to receive).
- **Wake-on-LAN**: Send a magic packet to a configured MAC address (with SSH-proxy support for nodes on a different LAN, via `w` in the detail screen).
- **Desktop notifications**: Get notified when a node goes online/offline.
- **Multiple tabs**:
  - `Devices`: node list (default)
  - `Exit Nodes`: browse and select exit node candidates
  - `Serve`: shows `tailscale serve/funnel` status
  - `Logs`: live `journalctl -u tailscaled` output, color-coded by severity
  - `Daemon`: run `tailscale up/down` and toggle Shields Up/Down

## Usage

### Keybindings

**Devices / Exit Nodes tabs (list view)**

| Key | Action |
| :--- | :--- |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `PgUp` / `PgDn` | Jump 10 rows |
| `/` | Search nodes |
| `o` | Toggle online-only filter |
| `r` | Manual refresh |
| `s` | Cycle sort order (Name → IP → OS → Ping) |
| `Enter` | Open the selected node's detail screen |
| `c` | Copy selected node's Tailscale IP |
| `S` | Copy the SSH command for the selected node |
| `t` | Copy Taildrop command (`tailscale file cp <file> <hostname>:`) |
| `E` | (Exit Nodes tab only) set selected node as exit node |
| `Tab` / `Shift+Tab` | Switch tabs |
| `q` / `Ctrl+C` | Quit |

**Detail screen** (`Enter` from the list)

| Key | Action |
| :--- | :--- |
| `s` | SSH into this node |
| `f` | Send a file to this node (Taildrop) |
| `g` | Receive pending Taildrop files |
| `a` | Accept subnet routes (`tailscale up --accept-routes`) |
| `w` | Send a Wake-on-LAN packet |
| `Esc` / `q` / `Backspace` | Back to the list |

**Logs tab**: `j/k`, `PgUp/PgDn` to scroll
**Serve tab**: `r` to refresh
**Daemon tab**: `u`=up, `d`=down, `s`=Shields Up, `S`=Shields Down

### Install

**Using go install** (Go 1.21+)

```bash
go install github.com/kuroiko0429/tail-pulse@latest
```

**Using a prebuilt binary**

Download the binary for your OS from the [Releases](https://github.com/kuroiko0429/tail-pulse/releases) page (example for Linux amd64):

```bash
curl -LO https://github.com/kuroiko0429/tail-pulse/releases/latest/download/tail-pulse-linux-amd64
install -Dm755 tail-pulse-linux-amd64 ~/.local/bin/tail-pulse
tail-pulse
```

If `~/.local/bin` isn't on your PATH yet, add it first (e.g. `export PATH="$HOME/.local/bin:$PATH"` in `~/.bashrc` or `~/.zshrc`).

For other platforms, swap the filename: `tail-pulse-linux-arm64` / `tail-pulse-darwin-arm64`.

**Running from source**

```bash
git clone https://github.com/kuroiko0429/tail-pulse.git
cd tail-pulse
go run main.go
```

## Configuration

`~/.config/tail-pulse/config.yaml` is generated automatically on first run.

```yaml
theme:
  cyan: "#83a598"
  dark_grey: "#928374"
  red: "#fb4934"
  white: "#ebdbb2"
  green: "#8ec07c"
  yellow: "#fabd2f"
  background: "#282828"
  tab_active: "#83a598"
  tab_inactive: "#3c3836"
  highlight: "#d3869b"
show_ping: true
cyber_glitch: true
ping_interval: 15       # seconds
ports: {}                # hostname -> SSH port override
mac_addresses: {}        # hostname -> MAC address (for Wake-on-LAN)
wol_proxy: ""             # hostname to relay WoL packets through for nodes on another LAN
```

Defaults to a Gruvbox theme. Override the hex codes under `theme` for a different palette.

## Requirements

### System

- **Tailscale CLI**: `tailscale` must be available in your PATH.
- **Clipboard tool**: One of the following is required for copy functionality:
  - `wl-copy` (Wayland)
  - `pbcopy` (macOS)
  - `xclip` (X11)
- **journalctl**: The Logs tab requires a systemd environment with `journalctl -u tailscaled`.
- **Nerd Fonts**: Recommended for correct icon rendering.

### Go packages

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Styling and layout
- `github.com/charmbracelet/bubbles` — TUI components
- `github.com/gen2brain/beeep` — Desktop notifications

## License

MIT
