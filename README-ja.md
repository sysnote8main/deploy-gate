# deploy-gate

Docker Socketを公開せずに、GitHub Actionsから安全にデプロイを行うためのシンプルなGo製Webhookサーバーです。

GitHub Webhookを受信し、署名を検証した上でデプロイ要求をキューファイルへ記録します。

## 特徴

- GitHub HMAC-SHA256署名検証
- Docker Socket不要
- ファイルベースのデプロイキュー
- 単一バイナリで動作
- 標準ライブラリのみ使用

## アーキテクチャ

```text
GitHub Actions
      │
      ▼
 deploy-gate
      │
      ▼
 deploy.queue
      │
      ▼
 deploy worker
      │
      ▼
 docker compose up -d
```

`deploy-gate` 自身はデプロイを実行しません。

責務は以下の3つだけです。

1. Webhookを受信する
2. 署名を検証する
3. デプロイ要求を記録する

実際のデプロイ処理は別プロセスへ委譲することを想定しています。

## 動作要件

- Go 1.24以降
- Linux
- GitHub Webhook

## 設定

環境変数で設定します。

| 変数名        | 必須 | 説明                                 |
| ------------- | ---- | ------------------------------------ |
| DEPLOY_SECRET | ○    | GitHub Webhook Secret                |
| QUEUE_DIR     | ○    | キューファイルを書き込むディレクトリ |

例:

```env
DEPLOY_SECRET=replace_me
QUEUE_DIR=/queue
```

## ビルド

```bash
go build -o deploy-gate ./cmd/deploy-gate
```

## 実行

```bash
DEPLOY_SECRET=replace_me \
QUEUE_DIR=/queue \
./deploy-gate
```

起動後は `:9000` で待ち受けます。

## Docker

ビルド:

```bash
docker build -t deploy-gate .
```

起動:

```bash
docker run \
  -e DEPLOY_SECRET=replace_me \
  -e QUEUE_DIR=/queue \
  -v $(pwd)/queue:/queue \
  -p 9000:9000 \
  deploy-gate
```

## API

### POST /deploy

GitHub Webhookから送信されるリクエストを受け付けます。

署名ヘッダ:

```http
X-Hub-Signature-256: sha256=<signature>
```

レスポンス:

| Status | Description                 |
| ------ | --------------------------- |
| 204    | Deployment request queued   |
| 403    | Invalid method or signature |
| 500    | Queue write failure         |

## プロジェクト構成

```text
deploy-gate/
├── cmd/
│   └── deploy-gate/
│       └── main.go
├── internal/
│   ├── queue/
│   │   └── file.go
│   ├── signature/
│   │   └── hmac.go
│   └── webhook/
│       └── deploy.go
├── Dockerfile
├── compose.yml
├── go.mod
└── README.md
```

## セキュリティ

deploy-gate は Docker Socket を利用しません。

Docker Socket をWebhook経由で公開すると、コンテナ操作やホストへのアクセスが可能となり、実質的にサーバーの管理権限を外部へ公開することになります。

deploy-gate は署名検証後にキューファイルを作成するだけの構成とすることで、攻撃面を最小限に抑えることを目的としています。

## ライセンス

MIT License
