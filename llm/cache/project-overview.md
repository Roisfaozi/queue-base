# Project Overview

## Current repo state

This repository is currently a hybrid starter monorepo with a Go API backend as the operational core and multiple JavaScript/TypeScript workspace packages for web clients and shared frontend libraries.

## Rebuild target context

Current product goal is not generic starter maintenance. Current goal is rebuilding a hospital/clinic queue-management application using:

- `documentation/task-overview.md`
- `code_context.txt` as legacy evidence only

Important distinction:

- live code describes what starter repo can do today
- rebuild docs describe what queue product must do next
- `code_context.txt` describes old implementation shape and bugs, not current runtime truth here

## Repository shape today

- Root workspace package name: `casbin-monorepo` from `package.json`.
- Root package manager: `pnpm@10.8.0` from `package.json`.
- Root build orchestrator: `turbo` from root scripts in `package.json`.
- Primary backend module: Go module `github.com/Roisfaozi/queue-base` from `go.mod`.
- Backend entrypoint: `cmd/api/main.go`.
- Secondary Go entrypoint: `cmd/gen/main.go` for module scaffolding.
- Frontend app packages present:
  - `apps/web` with `next dev --turbo` / `next build`
  - `apps/client` with `react-router dev` / `react-router build`

## Confirmed backend stack today

- HTTP framework: Gin.
- ORM / database access: GORM.
- Database drivers present: MySQL, Postgres, SQLite.
- Authorization library: Casbin.
- Cache / session / worker backend: Redis.
- Background jobs: Asynq.
- Upload protocol: TUS (`tusd`).
- Observability: OpenTelemetry.
- API documentation: Swagger (`swag`).

These are confirmed from `go.mod`, `package.json`, and runtime wiring.

## Confirmed queue-target needs from rebuild docs

Target product must eventually cover:

- patient/customer queue registration
- queue forward orchestration
- scanner check-in and credential validation
- branch/menu/station/device relations
- pharmacy flow validation
- transaction-safe queue number allocation
- business-day handling around `04:00 Asia/Jakarta`
- external integration-oriented flows

Those target capabilities are not yet represented as first-class starter modules.

## Current module gap

Current starter modules are platform-oriented:

- `access`
- `api_key`
- `audit`
- `auth`
- `organization`
- `permission`
- `project`
- `role`
- `stats`
- `user`
- `webhook`

Queue-domain modules still need to be designed and added.

## Canonical development commands confirmed

Root workspace:

- `pnpm build`
- `pnpm lint`
- `pnpm test`
- `pnpm typecheck`
- `pnpm go:run`
- `pnpm go:test`
- `pnpm go:test-integration`
- `pnpm go:test-e2e`

Backend via Makefile:

- `make run`
- `make build`
- `make test`
- `make test-unit`
- `make test-integration`
- `make test-e2e`
- `make test-all`
- `make docs`

## Scope boundary

For queue rebuild work, use `llm/research/queue-management-rebuild-brief.md` and `llm/plans/roadmap/queue-management-rebuild.md` together with live code.

## QMS Rebuild Addendum

For current QMS rebuild task, apply these additional target rules on top of existing starter overview:

- tenant-first queue platform
- branch under tenant
- one queue master row per ticket/day
- forwarding through `queue_journeys`
- `visit_journeys` as readable internal history
- settings inheritance: tenant -> branch -> service -> counter

Use `documentation/QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md` as target-product truth and keep this file's existing starter facts as runtime truth.
