# deploy-gate

GitHub Webhookから安全にローカルのデプロイスクリプトを実行するための、シンプルなGo製Webhookサーバーです。

`deploy-gate` はGitHub Webhookの署名を検証し、リクエストパスに応じて設定済みのローカルスクリプトを実行します。Webhook経由でDocker Socketを公開せずにデプロイを起動することを目的としています。

## 特徴

- GitHub HMAC-SHA256署名検証
- パスごとのデプロイルーティング
- 設定ファイルによるスクリプト指定
- 単一バイナリで動作
- 標準ライブラリのみ使用
- `deploy-gate` 自体はDocker Socket不要

## アーキテクチャ

```text
GitHub Webhook
      │
      ▼
 deploy-gate
      │
      ├─ /deploy/bot       → deploy-bot.sh
      │
      └─ /deploy/dashboard → deploy-dashboard.sh
```

`deploy-gate` の責務は以下です。

1. Webhookを受信する
2. GitHub署名を検証する
3. 設定済みのルートを選択する
4. 対応するスクリプトを実行する

実際のデプロイ処理は、各ルートに設定したスクリプト側で実装します。

## 動作要件

- Linux
- GitHub Webhook

Goはソースからビルドする場合のみ必要です。ビルド済みバイナリを使用する場合、実行環境にGoは不要です。

## 設定

`deploy-gate` は環境変数とJSON設定ファイルで設定します。

### 環境変数

| 変数名          | 必須 | 説明                   |
| --------------- | ---- | ---------------------- |
| `DEPLOY_SECRET` | ○    | GitHub Webhook Secret  |
| `DEPLOY_CONFIG` | ○    | JSON設定ファイルのパス |

例:

```env
DEPLOY_SECRET=replace_me
DEPLOY_CONFIG=/etc/deploy-gate/config.json
```

### 設定ファイル

例:

```json
{
  "routes": [
    {
      "path": "/deploy/bot",
      "script": "/opt/deploy-gate/scripts/deploy-bot.sh"
    },
    {
      "path": "/deploy/dashboard",
      "script": "/opt/deploy-gate/scripts/deploy-dashboard.sh"
    }
  ]
}
```

各ルートで、HTTPパスと実行するローカルスクリプトを対応付けます。

スクリプトのパスは絶対パスで指定する必要があります。

## ビルド

```bash
go build -o bin/deploy-gate ./cmd/deploy-gate
```

## 実行

```bash
DEPLOY_SECRET=replace_me \
DEPLOY_CONFIG=/etc/deploy-gate/config.json \
./bin/deploy-gate
```

起動後は `:9000` で待ち受けます。

## systemd設定例

```ini
[Unit]
Description=deploy-gate
After=network.target

[Service]
Type=simple
Environment=DEPLOY_SECRET=replace_me
Environment=DEPLOY_CONFIG=/etc/deploy-gate/config.json
ExecStart=/usr/local/bin/deploy-gate
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

## API

### POST 設定済みルート

GitHub Webhookから送信されるリクエストを受け付けます。

例:

```text
POST /deploy/bot
POST /deploy/dashboard
```

署名ヘッダ:

```http
X-Hub-Signature-256: sha256=<signature>
```

レスポンス:

| Status | Description                |
| ------ | -------------------------- |
| 204    | スクリプト実行成功         |
| 403    | メソッド不正または署名不正 |
| 500    | スクリプト実行失敗         |

## プロジェクト構成

```text
deploy-gate/
├── cmd/
│   └── deploy-gate/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── deploy/
│   │   └── run.go
│   ├── signature/
│   │   └── hmac.go
│   └── webhook/
│       └── deploy.go
├── go.mod
└── README.md
```

## セキュリティ

`deploy-gate` 自体はDocker Socketを必要としません。

Docker SocketをWebhook経由で公開すると、コンテナ操作やホストへのアクセスが可能となり、実質的にサーバーの管理権限を外部へ公開することになります。

`deploy-gate` は署名検証後、明示的に設定されたローカルスクリプトのみを実行します。スクリプトは小さく、監査しやすく、必要なデプロイ処理だけを行うようにしてください。

## ライセンス

MIT License
