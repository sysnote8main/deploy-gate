# deploy-gate

A simple Go-based webhook server for safely triggering deployments from GitHub Actions without exposing the Docker Socket.

`deploy-gate` receives GitHub Webhooks, verifies their signatures, and records deployment requests in a queue file.

## Features

- GitHub HMAC-SHA256 signature verification
- No Docker Socket access required
- File-based deployment queue
- Single binary deployment
- Uses only the Go standard library

## Architecture

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

`deploy-gate` does not perform deployments itself.

Its responsibilities are intentionally limited to:

1. Receiving webhook requests
2. Verifying signatures
3. Recording deployment requests

Actual deployment logic is expected to be handled by a separate process.

## Requirements

- Go 1.24 or later
- Linux
- GitHub Webhooks

## Configuration

Configuration is provided through environment variables.

| Variable        | Required | Description                                           |
| --------------- | -------- | ----------------------------------------------------- |
| `DEPLOY_SECRET` | Yes      | GitHub Webhook secret used for signature verification |
| `QUEUE_DIR`     | Yes      | Directory where the queue file will be written        |

Example:

```env
DEPLOY_SECRET=replace_me
QUEUE_DIR=/queue
```

## Build

```bash
go build -o deploy-gate ./cmd/deploy-gate
```

## Run

```bash
DEPLOY_SECRET=replace_me \
QUEUE_DIR=/queue \
./deploy-gate
```

The server listens on `:9000` by default.

## Docker

Build:

```bash
docker build -t deploy-gate .
```

Run:

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

Accepts webhook requests from GitHub.

Signature header:

```http
X-Hub-Signature-256: sha256=<signature>
```

Responses:

| Status | Description                 |
| ------ | --------------------------- |
| 204    | Deployment request queued   |
| 403    | Invalid method or signature |
| 500    | Failed to write queue file  |

## Project Structure

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

## Security

`deploy-gate` does not require access to the Docker Socket.

Exposing the Docker Socket through a webhook endpoint effectively grants remote control over containers and, in many cases, the host system itself.

To reduce the attack surface, `deploy-gate` only verifies webhook signatures and writes deployment requests to a queue file. Deployment execution is intentionally delegated to a separate process.

## License

MIT License
