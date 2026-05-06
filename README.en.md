# Tail-Pulse

A Watch Dogs-inspired TUI monitoring tool for your Tailscale network.

## Features

- **Real-time monitoring**: Polls Tailscale status every 2 seconds and displays all nodes
- **Connectivity checks**: Latency measurement via `tailscale ping` + sparkline visualization
- **SSH detection**: Automatic SSH port discovery per node (`ssh -G`)
- **Desktop notifications**: Instant alerts when nodes go online or offline
- **Live Logs**: Real-time `tailscaled` log streaming via journalctl
- **File transfer**: Send and receive files via Taildrop (`tailscale file cp/get`)
- **Wake-on-LAN**: Local UDP broadcast or via SSH proxy
- **Exit Node management**: Switch exit nodes directly from the TUI
- **Serve status**: View `tailscale serve status` output
- **Daemon control**: Run up/down/shields-up/shields-down from within the TUI
- **Search, sort, filter**: Filter by hostname or IP; sort by Name/IP/OS/Ping
- **Themeable**: Customize colors via `~/.config/tailpuls/config.yml` (Gruvbox by default)

## Screenshot

```
[ Devices ] [ Exit Nodes ] [ Serve ] [ Logs ] [ Daemon ]
󰒄 CTOS // TAILNET_MONITOR // v4.0.0

Filter: ALL (press 'o') | Sort: Name (press 's')
  HOSTNAME              IP              OS    STATUS       SSH   CONN_TYPE      PING
  >> SV1-cachy          100.107.227.39        󰄬 ONLINE     󱘖     ----            14ms
󰁔   thincentre          100.106.198.93        󰄬 ONLINE     󱘖     󰇚 tok           20ms  Direct
    thinkarch-server    100.71.188.88         󰄬 ONLINE     󱘖     󰇚 tok           19ms  Direct
    llm-server          100.78.153.4          󰄱 OFFLINE    󱘖     ----           ---
```

## Installation

```bash
git clone https://github.com/kuroiko0429/tail-pulse
cd tail-pulse
go build -o tail-pulse .
```

## Usage

```bash
./tail-pulse
# or
go run .
```

For easy access from anywhere, move the binary to `~/.local/bin`:

```bash
mkdir -p ~/.local/bin
mv tail-pulse ~/.local/bin/
# Add to ~/.bashrc, ~/.zshrc, or ~/.config/fish/config.fish:
# export PATH="$HOME/.local/bin:$PATH"
```

## Keybindings

### Main View (Devices tab)

| Key | Action |
| :--- | :--- |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `PgDn` / `PgUp` | Move 10 rows |
| `Tab` / `Shift+Tab` | Switch tab |
| `Enter` | Open detail panel |
| `/` | Search by hostname or IP |
| `o` | Toggle filter (ALL / ONLINE ONLY) |
| `s` | Cycle sort (Name / IP / OS / Ping) |
| `r` | Refresh connection status now |
| `c` | Copy selected node's Tailscale IP |
| `S` | Copy SSH command (`ssh <IP>`) |
| `E` | Use as exit node (Exit Nodes tab only) |
| `q` / `Ctrl+C` | Quit |

### Detail Panel

| Key | Action |
| :--- | :--- |
| `s` | SSH connect |
| `f` | Send file (Taildrop) |
| `g` | Receive file (Taildrop) |
| `a` | Accept subnet routes |
| `w` | Wake-on-LAN |
| `Esc` / `q` / `Backspace` | Back |

### Logs Tab

| Key | Action |
| :--- | :--- |
| `j` / `k` / `PgDn` / `PgUp` | Scroll |

### Daemon Tab

| Key | Action |
| :--- | :--- |
| `u` | `tailscale up` |
| `d` | `tailscale down` |
| `s` | Shields UP |
| `S` | Shields DOWN |

## Configuration

`~/.config/tailpuls/config.yml` is auto-generated on first run.

```yaml
ping_interval: 15           # connection check interval in seconds
default_sort: Name          # startup sort: Name / IP / OS / Ping

# per-host SSH ports (default: 22)
ports:
  my-server: "2222"

# MAC addresses for Wake-on-LAN
mac_addresses:
  my-desktop: "AA:BB:CC:DD:EE:FF"

# SSH proxy host for WoL (leave empty for local UDP)
wol_proxy: ""

# color theme (default: Gruvbox)
theme:
  cyan: "#83a598"
  dark_grey: "#928374"
  red: "#fb4934"
  white: "#ebdbb2"
  green: "#8ec07c"
  yellow: "#fabd2f"
  tab_active: "#83a598"
  tab_inactive: "#3c3836"
  highlight: "#d3869b"
```

## Requirements

### System

- **Tailscale CLI**: `tailscale` must be available in PATH
- **systemd**: Logs tab uses `journalctl`
- **Clipboard tool**: `wl-copy` (Wayland) / `xclip` (X11) / `pbcopy` (macOS)
- **Nerd Fonts**: Recommended for icon rendering

### Go packages

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Styling
- `github.com/charmbracelet/bubbles` — TUI components
- `github.com/gen2brain/beeep` — Desktop notifications
- `gopkg.in/yaml.v3` — Config file

## License

MIT
