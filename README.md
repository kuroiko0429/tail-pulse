# Tail-Pulse

Tailscale内の全デバイスの稼働状況を監視する、Watch Dogs風デザインのTUIモニタリングツールです。

## 機能

- **リアルタイム・モニタリング**: Tailscale statusを3秒ごとに取得し、全ノードの状態を表示。
- **コネクティビティ・チェック**: 
  - 各ノードへのPing (レイテンシ) を自動測定。
  - スパークライン（ミニグラフ）による直近の通信品質の可視化。
  - ポートスキャン: SSH(22) / Web(80,443) / RDP(3389) / VNC(5900) の自動検出。
- **デザイン・UI**:
  - `Charmbracelet Bubble Tea / Lip Gloss` を使用したスタイリッシュなTUI。
  - OSごとのアイコン表示。
  - ダイレクト接続/リレー接続の判別表示。
  - ノード詳細パネル（ポートスキャン結果・DNSName・タグ・ルート表示）。
  - SSH接続時のハックアニメーション演出。
- **SSH連携**:
  - ノードを選択して `Enter` で詳細画面を開き、`s`でSSH接続（SSHポートは`ssh -G`で自動検出）。
  - ワンキーでIPアドレスやSSHコマンド・Taildropコマンドをクリップボードにコピー。
- **ノード検索**: `/` キーでホスト名・IPによるリアルタイム絞り込み。
- **ファイル転送**: Taildrop経由でのファイル送受信（詳細画面で`f`送信・`g`受信）。
- **Wake-on-LAN**: 設定したMACアドレスにマジックパケットを送信（同一LAN外はSSHプロキシ経由にも対応、詳細画面の`w`）。
- **デスクトップ通知**: ノードのオンライン/オフライン切り替わりを通知。
- **複数タブ**:
  - `Devices`: ノード一覧（デフォルト）
  - `Exit Nodes`: Exit Node候補の一覧・切り替え
  - `Serve`: `tailscale serve/funnel`のステータス表示
  - `Logs`: `journalctl -u tailscaled`のライブログ（エラー/警告を色分け）
  - `Daemon`: `tailscale up/down`・Shields up/downの実行

## 使い方

### キーバインド

**Devices / Exit Nodes タブ（一覧画面）**

| キー | アクション |
| :--- | :--- |
| `j` / `↓` | 下に移動 |
| `k` / `↑` | 上に移動 |
| `/` | ノード検索 |
| `Enter` | 選択中のノードの詳細画面を開く |
| `c` | 選択中のノードのTailscale IPをコピー |
| `S` | 選択中のノードへのSSHコマンドをコピー |
| `t` | Taildropコマンド (`tailscale file cp <file> <hostname>:`) をコピー |
| `E` | (Exit Nodesタブのみ) 選択中のノードをExit Nodeに設定 |
| `Tab` / `Shift+Tab` | タブ切り替え |
| `q` / `Ctrl+C` | 終了 |

**詳細画面**（一覧で`Enter`）

| キー | アクション |
| :--- | :--- |
| `s` | SSH接続を実行 |
| `f` | 選択中のノードにファイルを送信（Taildrop） |
| `g` | 保留中のTaildropファイルを受信 |
| `a` | サブネットルートを許可（`tailscale up --accept-routes`） |
| `w` | Wake-on-LANパケットを送信 |
| `Esc` / `q` / `Backspace` | 一覧画面に戻る |

**Logsタブ**: `j/k`・`PgUp/PgDn`でスクロール
**Serveタブ**: `r`で再取得
**Daemonタブ**: `u`=up, `d`=down, `s`=Shields Up, `S`=Shields Down

### インストール

**go installを使う場合**（Go 1.21以上）

```bash
go install github.com/kuroiko0429/tail-pulse@latest
```

**ビルド済みバイナリを使う場合**

[Releases](https://github.com/kuroiko0429/tail-pulse/releases)からOSに合ったバイナリをダウンロードして実行権限を付与するだけ。

```bash
chmod +x tail-pulse
./tail-pulse
```

**ソースから実行する場合**

```bash
git clone https://github.com/kuroiko0429/tail-pulse.git
cd tail-pulse
go run main.go
```

## 設定

初回起動時に `~/.config/tail-pulse/config.yaml` が自動生成されます。

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
ping_interval: 15       # 秒
ports: {}                # ホスト名 -> SSHポートの上書き
mac_addresses: {}        # ホスト名 -> MACアドレス(Wake-on-LAN用)
wol_proxy: ""             # 別LANのノードを起こす際に踏み台にするホスト名
```

デフォルトはGruvboxテーマ。`theme`以下の色コードを書き換えれば好きな配色にできる。

## 依存関係

### システム要件
- **Tailscale CLI**: `tailscale` コマンドがパスに通っている必要があります。
- **Clipboardツール**: コピー機能を使用するために以下のいずれかが必要です。
  - `wl-copy` (Wayland)
  - `pbcopy` (macOS)
  - `xclip` (X11)
- **journalctl**: Logsタブは`journalctl -u tailscaled`が使える環境（systemd）が前提です。
- **Nerd Fonts**: アイコンを正しく表示するために推奨されます。

### Goパッケージ
- `github.com/charmbracelet/bubbletea`: TUIフレームワーク
- `github.com/charmbracelet/lipgloss`: スタイリング・レイアウト
- `github.com/charmbracelet/bubbles`: TUIコンポーネント
- `github.com/gen2brain/beeep`: デスクトップ通知

## 作者 / ライセンス
MIT License
