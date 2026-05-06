# Tail-Pulse

Tailscaleネットワーク内の全デバイスをリアルタイム監視する、Watch Dogs風TUIツール。

## 機能

- **リアルタイム監視**: 2秒ごとにTailscale statusを取得し全ノードの状態を表示
- **接続チェック**: `tailscale ping` によるレイテンシ測定 + スパークライン可視化
- **SSH検出**: ノードごとのSSHポートを自動検出 (`ssh -G`)
- **デスクトップ通知**: ノードのオンライン/オフライン変化を即座に通知
- **Live Logs**: journalctl経由でtailscaledのログをリアルタイムストリーム
- **ファイル転送**: `tailscale file cp/get` によるTaildropの送受信
- **Wake-on-LAN**: ローカルUDP送信 または SSHプロキシ経由のWoL対応
- **Exit Node管理**: Exit Nodeタブから即座に切り替え
- **Serve状態確認**: `tailscale serve status` の表示
- **Daemon制御**: up/down/shields-up/shields-downをTUI内から実行
- **検索・ソート・フィルタ**: ホスト名・IPで絞り込み、Name/IP/OS/Pingでソート
- **テーマ設定**: `~/.config/tailpuls/config.yml` でGruvboxカラーをカスタマイズ可能

## スクリーンショット

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

## インストール

```bash
git clone https://github.com/kuroiko0429/tail-pulse
cd tail-pulse
go build -o tail-pulse .
```

## 使い方

```bash
./tail-pulse
# または
go run .
```

`~/.local/bin` にPATHを通しておくと、どこからでも呼び出せて便利です:

```bash
mkdir -p ~/.local/bin
mv tail-pulse ~/.local/bin/
# ~/.bashrc or ~/.zshrc or ~/.config/fish/config.fish に追加:
# export PATH="$HOME/.local/bin:$PATH"
```

## キーバインド

### メイン画面 (Devicesタブ)

| キー | アクション |
| :--- | :--- |
| `j` / `↓` | 下に移動 |
| `k` / `↑` | 上に移動 |
| `PgDn` / `PgUp` | 10行ずつ移動 |
| `Tab` / `Shift+Tab` | タブ切り替え |
| `Enter` | 詳細パネルを開く |
| `/` | ホスト名・IPで検索 |
| `o` | フィルタ切り替え (ALL / ONLINE ONLY) |
| `s` | ソート切り替え (Name / IP / OS / Ping) |
| `r` | 接続状態を今すぐ更新 |
| `c` | 選択ノードのTailscale IPをコピー |
| `S` | SSHコマンドをコピー (`ssh <IP>`) |
| `E` | Exit NodeとしてON (Exit Nodesタブのみ) |
| `q` / `Ctrl+C` | 終了 |

### 詳細パネル

| キー | アクション |
| :--- | :--- |
| `s` | SSH接続 |
| `f` | ファイルを送信 (Taildrop) |
| `g` | ファイルを受信 (Taildrop) |
| `a` | サブネットルートを承認 |
| `w` | Wake-on-LAN |
| `Esc` / `q` / `Backspace` | 戻る |

### Logsタブ

| キー | アクション |
| :--- | :--- |
| `j` / `k` / `PgDn` / `PgUp` | スクロール |

### Daemonタブ

| キー | アクション |
| :--- | :--- |
| `u` | `tailscale up` |
| `d` | `tailscale down` |
| `s` | Shields UP |
| `S` | Shields DOWN |

## 設定ファイル

初回起動時に `~/.config/tailpuls/config.yml` が自動生成されます。

```yaml
ping_interval: 15           # 接続チェックの間隔（秒）
default_sort: Name          # 起動時のソート (Name / IP / OS / Ping)

# ノードごとのSSHポート（デフォルト: 22）
ports:
  my-server: "2222"

# Wake-on-LAN用のMACアドレス
mac_addresses:
  my-desktop: "AA:BB:CC:DD:EE:FF"

# WoLをSSH経由で送る場合のプロキシホスト名
wol_proxy: ""

# カラーテーマ (デフォルト: Gruvbox)
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

## 依存関係

### システム要件

- **Tailscale CLI**: `tailscale` コマンドがPATHに通っていること
- **systemd**: Logsタブは `journalctl` を使用
- **Clipboardツール** (コピー機能): `wl-copy` (Wayland) / `xclip` (X11) / `pbcopy` (macOS)
- **Nerd Fonts**: アイコン表示に推奨

### Goパッケージ

- `github.com/charmbracelet/bubbletea` — TUIフレームワーク
- `github.com/charmbracelet/lipgloss` — スタイリング
- `github.com/charmbracelet/bubbles` — TUIコンポーネント
- `github.com/gen2brain/beeep` — デスクトップ通知
- `gopkg.in/yaml.v3` — 設定ファイル

## ライセンス

MIT
