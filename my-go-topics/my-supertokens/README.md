# my-supertokens

SuperTokens email+password authentication example using the layered architecture pattern from this repo.

## Quick start

```bash
# 1. Start SuperTokens core + Postgres
docker compose up -d

# 2. Run the server (waits for core to be up on :3567)
go run ./cmd/server/main.go
```

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/register` | — | Register a new user |
| `POST` | `/auth/login` | — | Login, creates a session cookie |
| `POST` | `/auth/logout` | session | Revoke session |
| `GET`  | `/api/me` | session | Return current user ID |

SuperTokens also auto-mounts its own routes under `/auth/*` (token refresh, signout SDK flow, etc.).

### Register

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"Password123!"}'
```

### Login

```bash
curl -sc cookies.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"Password123!"}'
```

### Protected route

```bash
curl -sb cookies.txt http://localhost:8080/api/me
```

## Project layout

```
cmd/server/main.go          — wires context, config, starts server
internal/app/app.go         — dependency injection, SuperTokens init
internal/domain/            — User model, AuthService interface, errors
internal/repository/auth.go — SuperTokens implementation of AuthService
internal/http/handler.go    — HTTP handlers (register, login, logout, me)
internal/http/router.go     — Chi router with SuperTokens middleware
internal/config/config.go   — config struct with defaults
docker-compose.yml          — SuperTokens core + PostgreSQL
```
