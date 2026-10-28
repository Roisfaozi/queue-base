# AGENTS.md

Repository operational guide for AI coding agents.

## 1. System Overview

- Hybrid monorepo: Go backend operational core, TypeScript frontends (`apps/web` Next.js, `apps/client` React Router v7), root workspace managed via `pnpm` and `turbo`.
- Execution lifecycle: `cmd/api/main.go` -> `internal/config/app.go` (dependency injection & module wiring) -> `internal/router/router.go` (middleware layering: public, auth, tenant+Casbin, admin, upload).
- Module pattern (`internal/modules/<name>/`):
  - `entity/`: GORM models
  - `model/`: Request/Response DTOs
  - `repository/`: Data access, every query scoped by tenant ID
  - `usecase/`: Business logic, DB transactions, audit logging
  - `delivery/http/`: Thin Gin handlers and route registration
  - `module.go`: Constructor returning `*Module` consumed by `app.go`
- Cross-module dependencies: Wired exclusively in `module.go` (e.g. `RelationValidator`), never imported directly across usecase packages.
- Shared infrastructure: `pkg/` (tx manager, querybuilder, jwt, sse, ws, tus, storage, exception).
- Background jobs: Asynq workers in `internal/worker`.
- Active QMS modules: `queue`, `queue_config`, `service`, `counter`, `caller`, `scanner`, `signage`, `operator_assignment`, `qms_client` (live alongside starter modules `auth`, `organization`, `user`, `role`, `permission`, `access`, `api_key`, `audit`, `webhook`, `stats`).

## 2. Developer Commands

### Backend (Go)
- Run unit tests: `make test` (serial `-p 1 -parallel 1 -count=1`, output filtered to PASS or FAIL detail).
- Single unit test: `make test-custom RUN='<TestName>' PKG=./internal/modules/<pkg>/usecase/...`
- Unfiltered verbose test run: `make test-local TEST_PKG=./internal/modules/<pkg>/...`
- Integration tests (requires Docker): `make test-integration`
- Single integration test: `make test-integration-custom RUN='<TestName>' PKG=./tests/integration/...` (tag `integration`).
- E2E tests (requires Docker): `make test-e2e`
- Single E2E test: `make test-e2e-custom RUN='<TestName>' PKG=./tests/e2e/...` (tag `e2e`).
- All tests: `make test-all`
- Lint: `make lint` (`golangci-lint`)
- Interface mocks: `make mocks` (mockery with `.mockery.yml`; run after interface changes).
- Swagger docs: `make docs` (regenerates `docs/swagger.*`; note `make run` auto-runs `make docs`).
- Scaffold module: `make gen-module` (uses `cmd/gen`).
- Environment readiness: `make doctor`
- CI workflow (`.github/workflows/ci.yml`): Runs `golangci-lint`, `make test-coverage`, `make test-integration`, `make test-e2e`, `make bench`, and `go build ./cmd/api/main.go`.

### Database & Migrations
- Migration files: `db/migrations/<timestamp>_<name>.up.sql` and `.down.sql` (always paired).
- Worktree local migrations: `make migrate-up-local`, `make migrate-down-local` (reads from `.env.local`).

### Frontend (pnpm + turbo)
- Typecheck: `pnpm typecheck`
- Lint: `pnpm lint` (runs `biome lint --write`, modifies files in-place).
- Build: `pnpm build`
- Format: `pnpm format` (`biome format --write .`)
- Dev servers: `pnpm dev:web` (`apps/web`), `pnpm dev:client` (`apps/client`).
- Note: `apps/client` lint script is placeholder-only; verify via `pnpm typecheck` or `pnpm --filter casbin-client typecheck`.

### Worktree Isolated Dev
- Create worktree: `make wt-new <branch> [base]` (creates `.worktrees/<slug>` with isolated port offsets in `.env.local`).
- Inside worktree: `make dev-up`, `make dev-down`, `make dev-reset`, `make migrate-up-local`.

## 3. Hard Architectural & Safety Invariants

- Tenant isolation: Multi-tenancy is first-class. Every repository query must scope by tenant ID (`org_id` / `tenant_id`). Cross-tenant data access is strictly forbidden (`exception.ErrForbidden`).
- Redis-backed sessions: Auth requires valid Redis session verification in middleware/usecase. JWT cryptographic signature validity alone is insufficient.
- Casbin transactions: Policy mutations inside DB transactions must use transactional enforcer patterns (`txEnforcer`).
- QueryBuilder security: Field allowlists are strict; never permit dynamic filtering or sorting on sensitive fields (passwords, tokens, internal keys).
- Logging & Audit: Every controller, usecase, and transaction path must log. Transactional side effects must emit audit logs via non-blocking `tryAudit` pattern.
- Tool usage restriction: NEVER generate Python or shell scripts (`.py`, `.sh`, etc.) to edit codebase files or apply schema changes. Always use standard agent editing tools directly.

## 4. QMS Domain Specifics

- Master ticket: `queues` table is the sole master ticket record.
- Ticket forwarding: Uses `queue_journeys`, NEVER creates a second queue master ticket row.
- Visit history: `visit_journeys` is internal readable history/projection.
- Configuration hierarchy: Tenant -> Branch -> Service -> Counter resolved via `SettingsResolver`.
- TDD Requirement: Queue feature/fix changes require table-driven tests covering four mandatory cases:
  1. Positive (happy path)
  2. Negative (invalid inputs / state transitions)
  3. Edge (e.g. queue date before daily reset time mapped to previous business date)
  4. Vulnerability / Security (e.g. cross-tenant or mismatched branch/service access rejected)

## 5. Deeper Context References

- Architecture details: `documentation/QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md`
- Verified system cache: `llm/cache/` (`architecture.md`, `backend-map.md`, `domain-rules.md`, `api-contracts.md`)
- Queue domain & rebuild plans: `llm/research/queue-management-rebuild-brief.md`, `llm/plans/roadmap/queue-management-rebuild.md`
- Coding conventions: `llm/conventions/golang.md`, `llm/conventions/testing.md`, `llm/conventions/database.md`
