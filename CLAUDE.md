# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Structure

This is a Go learning and experimentation monorepo. Each sub-project under `my-go-topics/` has its own `go.mod` and is independent. The root `go.mod` is used only for the Makefile-driven Kubernetes/GORM experiments.

Key sub-projects (most to least advanced):
- `my-go-topics/my-http-server-002/` — most advanced: OAPI codegen, OpenAPI validation, Viper config, graceful shutdown
- `my-go-topics/my-go-rest-foundation-004/` — signal handling, DDD patterns
- `my-go-topics/my-go-rest-foundation-002/` — Chi router, slog middleware
- `my-go-topics/my-go-rest-foundation/` — simplest complete REST API

## Commands

Each sub-project is independent. Run commands from within the specific project directory.

```bash
# Build
go build ./...

# Run
go run ./cmd/server/main.go

# Test all
go test ./...

# Test single package
go test ./internal/http/...

# Test single function
go test -run TestFunctionName ./internal/http/...

# Lint (from repo root, applies to root module)
make lint          # runs golangci-lint with --fix
make lint-install  # installs golangci-lint v2.5.0

# Docker
make docker-dev-build
make docker-dev-run
```

## Architecture Pattern

All REST projects use the same layered structure:

```
cmd/server/main.go          → wires context, config, starts server
internal/app.go             → dependency injection: connects domain, repo, HTTP
internal/domain/            → business logic, models, interfaces (no external deps)
internal/http/              → Chi router, handlers, middleware
internal/repository/        → data access implementations
internal/config/            → Viper-based YAML config
```

The `internal/app.go` is the single wiring point — all constructors flow through `app.New()`.

## Key Patterns

**HTTP Routing**: Chi v5 (`github.com/go-chi/chi/v5`) or `oapi-codegen` generated strict handlers.

**OAPI codegen** (in `my-http-server-002/`): OpenAPI spec → generated Go interfaces/types. Implement the generated `StrictServerInterface`. Validation via `kin-openapi`.

**Logging**: `log/slog` with `slogctx` (`github.com/veqryn/slog-context`) for context-propagated loggers. Request ID is injected per-request in middleware.

**Error handling**: Domain errors defined in `internal/domain/errors.go`, wrapped with `fmt.Errorf("%w")` or `samber/oops`. HTTP layer maps domain errors to status codes + structured JSON error responses.

**Middleware chain order** (chi — last `.Use()` runs first around handler):
```
Panic Recovery → Logging → OAPI Validation → Request ID → Handler
```

**Config**: Viper reads YAML file + environment variable overrides. Config structs live in `internal/config/`.

**Signal handling**: `signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)` → cancels root context → triggers `server.Shutdown()`.

## Codegen

For projects using `oapi-codegen`, regenerate after editing the OpenAPI spec:
```bash
go generate ./...
```
The `//go:generate` directive is in the file that imports the generated code.
