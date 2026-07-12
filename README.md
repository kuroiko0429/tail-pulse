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
  - ノードを選択して `Enter` で即座にSSH接続（SSHポートは`ssh -G`で自動検出）。
  - ワンキーでIPアドレスやTaildropコマンドをクリップボードにコピー。
- **ノード検索**: `/` キーでホスト名・IPによるリアルタイム絞り込み。
- **ファイル転送**: Taildrop経由でのファイル送受信（`T`で送信、`g`で受信）。
- **Wake-on-LAN**: 設定したMACアドレスにマジックパケットを送信（同一LAN外はSSHプロキシ経由にも対応）。
- **デスクトップ通知**: ノードのオンライン/オフライン切り替わりを通知。
- **複数タブ**:
  - `Devices`: ノード一覧（デフォルト）
  - `Exit Nodes`: Exit Node候補の一覧・切り替え
  - `Serve`: `tailscale serve/funnel`のステータス表示
  - `Logs`: `journalctl -u tailscaled`のライブログ（エラー/警告を色分け）
  - `Daemon`: `tailscale up/down`・Shields up/downの実行

## 使い方

### キーバインド

**Devices / Exit Nodes タブ**

| キー | アクション |
| :--- | :--- |
| `j` / `↓` | 下に移動 |
| `k` / `↑` | 上に移動 |
| `/` | ノード検索 |
| `d` | 詳細パネルの表示/非表示 |
| `c` | 選択中のノードのTailscale IPをコピー |
| `t` | Taildropコマンド (`tailscale file cp <file> <hostname>:`) をコピー |
| `T` | 選択中のノードにファイルを送信（Taildrop） |
| `g` | 保留中のTaildropファイルを受信 |
| `w` | 選択中のノードにWake-on-LANパケットを送信 |
| `a` | サブネットルートを許可（`tailscale up --accept-routes`） |
| `E` | (Exit Nodesタブのみ) 選択中のノードをExit Nodeに設定 |
| `Enter` | 選択中のノードへSSH接続を実行 |
| `Tab` / `Shift+Tab` | タブ切り替え |
| `q` / `Ctrl+C` | 終了 |

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
theme: cyberpunk
show_ping: true
cyber_glitch: true
ping_interval: 15       # 秒
ports: {}                # ホスト名 -> SSHポートの上書き
mac_addresses: {}        # ホスト名 -> MACアドレス(Wake-on-LAN用)
wol_proxy: ""             # 別LANのノードを起こす際に踏み台にするホスト名
```

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
