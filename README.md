# Soulwi Backend Showcase (Go)

Public showcase of the Soulwi backend service, focused on Go architecture and API implementation quality.

## What this repository contains

- Go API service (`gin`) with layered architecture.
- PostgreSQL persistence via `gorm`.
- Background jobs and notification flows.
- Auth middleware (JWT + optional Firebase verification).

This repo is intentionally scoped to backend code for recruiter review.

## Tech stack

- Go 1.23+
- Gin (`github.com/gin-gonic/gin`)
- GORM + PostgreSQL
- Firebase Admin SDK (optional for local run)

## Project structure

- `cmd/server/main.go` - application entrypoint.
- `internal/transport/router` - route registration and grouping.
- `internal/handler` - HTTP handlers.
- `internal/usecase` - business logic.
- `internal/repository` - data access layer.
- `internal/model` + `internal/migration` - schema and migrations.
- `internal/di/di.go` - dependency wiring.

## Quick start

### 1. Prepare environment

```bash
cp .env.example .env
```

Minimal local values are already prefilled in `.env.example` for PostgreSQL.

Optional for full auth/integration features:

- `FIREBASE_CREDS_FILE` or `FIREBASE_CREDS_JSON`
- `FIREBASE_WEB_API_KEY`
- `OPENAI_KEY`
- `TG_BOT_TOKEN`

If Firebase credentials are not provided, Firebase-dependent endpoints return `503`, but the server still starts and core infra can be reviewed.

### 2. Start PostgreSQL

```bash
docker compose -f compose.yaml up -d db
```

### 3. Run the API

```bash
go run ./cmd/server
```

Service starts on `http://localhost:${API_PORT}` (`8000` in `.env.example`).

### 4. Health check

```bash
curl http://localhost:8000/health
```

Expected response:

```json
{"status":"ok"}
```

## Development commands

```bash
make deps        # download modules
make run         # run server
make build       # build binary
make lint        # go fmt + go vet
```

## Tests

```bash
go test ./...
```

## Security and showcase defaults

- Development/debug auth endpoints are disabled by default.
- Enable them only for local debugging with:

```bash
ENABLE_DEV_ROUTES=true
```

- No secrets are committed; use `.env` locally.

## Notes for reviewers

- Main focus for review is Go backend design and maintainability.
- Feature areas include chat, prompts, users, notes, todos, subscriptions, notifications, and cron-triggered flows.
