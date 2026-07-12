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
  - ノードを選択して `Enter` で即座にSSH接続。
  - ワンキーでIPアドレスやTaildropコマンドをクリップボードにコピー。
- **ノード検索**: `/` キーでホスト名・IPによるリアルタイム絞り込み。

## 使い方

### キーバインド

| キー | アクション |
| :--- | :--- |
| `j` / `↓` | 下に移動 |
| `k` / `↑` | 上に移動 |
| `/` | ノード検索 |
| `d` | 詳細パネルの表示/非表示 |
| `c` | 選択中のノードのTailscale IPをコピー |
| `t` | Taildropコマンド (`tailscale file cp <file> <hostname>:`) をコピー |
| `Enter` | 選択中のノードへSSH接続を実行 |
| `q` / `Ctrl+C` | 終了 |

### 実行方法

```bash
go run main.go
```

## 依存関係

### システム要件
- **Tailscale CLI**: `tailscale` コマンドがパスに通っている必要があります。
- **Clipboardツール**: コピー機能を使用するために以下のいずれかが必要です。
  - `wl-copy` (Wayland)
  - `pbcopy` (macOS)
- **Nerd Fonts**: アイコンを正しく表示するために推奨されます。

### Goパッケージ
- `github.com/charmbracelet/bubbletea`: TUIフレームワーク
- `github.com/charmbracelet/lipgloss`: スタイリング・レイアウト
- `github.com/charmbracelet/bubbles`: TUIコンポーネント

## 作者 / ライセンス
MIT License
