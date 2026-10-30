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
- All tests: `make test-all` is an informational summary; it currently exits 0 even when a constituent suite fails. Run unit/integration/E2E targets separately for a reliable gate.
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
- Both apps use `biome lint --write .`; lint is not read-only and does not replace typecheck. `pnpm --filter casbin-client typecheck` generates React Router types before running TypeScript.

### Worktree Isolated Dev
- Create worktree: `make wt-new <branch> [base]` (creates `.worktrees/<slug>` with isolated port offsets in `.env.local`).
- Inside worktree: `make dev-up`, `make dev-down`, `make dev-reset`, `make migrate-up-local`.

## 3. Hard Architectural & Safety Invariants

- Tenant isolation: Multi-tenancy is first-class. Every repository query must scope by tenant ID (`org_id` / `tenant_id`). Cross-tenant data access is strictly forbidden (`exception.ErrForbidden`).
- Redis-backed sessions: Auth requires valid Redis session verification in middleware/usecase. JWT cryptographic signature validity alone is insufficient.
- Casbin transactions: Policy mutations inside DB transactions must use transactional enforcer patterns (`txEnforcer`).
- QueryBuilder security: Field allowlists are strict; never permit dynamic filtering or sorting on sensitive fields (passwords, tokens, internal keys).
- Logging & Audit: Every controller, usecase, and transaction path must log. Transactional side effects must emit audit logs via non-blocking `tryAudit` pattern.
- Mocking standard: All Go unit test mocks MUST be generated using Mockery (`make mocks` via `.mockery.yml`). Handwritten mock/stub structs (e.g. `type stub... struct`) or ad-hoc fake implementations in `*_test.go` are strictly forbidden. When an interface is required for unit tests, register it in `.mockery.yml` and run `make mocks`. Follow [mockery-testing](.agents/skills/mockery-testing/SKILL.md).
- Tool usage restriction: NEVER generate Python or shell scripts (`.py`, `.sh`, etc.) to edit codebase files or apply schema changes. Always use standard agent editing tools directly.

## 4. QMS Domain Specifics

- Master ticket: `queues` table is the sole master ticket record.
- Ticket forwarding: Uses `queue_journeys`, NEVER creates a second queue master ticket row.
- Visit history: `visit_journeys` is internal readable history/projection.
- Configuration hierarchy: tenant, branch, branch-service, counter. `QueueConfigResolver` implements the queue's `SettingsResolver` interface and resolves the most specific override first; service overrides belong to `branch_services`.
- TDD Requirement: Queue feature/fix changes require table-driven tests covering four mandatory cases:
  1. Positive (happy path)
  2. Negative (invalid inputs / state transitions)
  3. Edge (e.g. queue date before daily reset time mapped to previous business date)
  4. Vulnerability / Security (e.g. cross-tenant or mismatched branch/service access rejected)

## 5. Harness & Task Routing

- `CLAUDE.md` imports this file. Canonical skills live in `.agents/skills/`; `.claude/skills` is a relative symlink, not another copy. Personal settings and worktrees remain ignored. Shared Claude hooks, permissions, and plugins are not required.
- Start with [project memory routing](llm/README.md), then load only the skill/workflow matching the task. Runtime facts come from live code, wiring/config, then tests; approved requirements define intended changes. Historical plans are evidence, not current execution authority.
- Queue registration, forwarding, reset/settings, caller, scanner, signage, or device identity: [qms-domain](.agents/skills/qms-domain/SKILL.md).
- Go module logic: [go-service](.agents/skills/go-service/SKILL.md). Endpoints: [api-endpoint](.agents/skills/api-endpoint/SKILL.md). Auth/session/tenant/Casbin: [auth-tenant-casbin](.agents/skills/auth-tenant-casbin/SKILL.md).
- Frontend ownership: [frontend-surface](.agents/skills/frontend-surface/SKILL.md). Producer/consumer contracts: [cross-stack-change](.agents/skills/cross-stack-change/SKILL.md). Current QMS pages belong to `apps/web`; `apps/client` remains active but has no QMS routes. Check actual consumers instead of automatically mirroring features.
- Bug diagnosis: [systematic-debugging](.agents/skills/systematic-debugging/SKILL.md). Feature/fix tests: [test-driven-development](.agents/skills/test-driven-development/SKILL.md). Interface mocking: [mockery-testing](.agents/skills/mockery-testing/SKILL.md).
- Review findings: [review-findings](.agents/skills/review-findings/SKILL.md). Delegation: [dispatch-prompt](.agents/skills/dispatch-prompt/SKILL.md). A recommendation is not owner approval. Preserve existing work; hiding a feature or unmounting a route does not authorize deleting its implementation.
- Before claiming completion: [verification-before-completion](.agents/skills/verification-before-completion/SKILL.md). Harness-only checks: `pnpm harness:check`; no Docker or application build required.
- Route groups are mounted independently in `internal/router/router.go`. `tenantAuthorized` is not a child of `authenticated`; inspect exact middleware and token/session validation before changing protection. Caller/signage add QMS-client authentication beneath tenant-authorized routes.

## 6. Deeper Context References

- Architecture intent: [QMS architecture](documentation/QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md). Verify implementation against live code before treating design as runtime fact.
- Runtime facts: [architecture](llm/cache/architecture.md), [backend map](llm/cache/backend-map.md), [domain rules](llm/cache/domain-rules.md), [API contracts](llm/cache/api-contracts.md).
- Historical rebuild context: [brief](llm/research/queue-management-rebuild-brief.md), [roadmap](llm/plans/roadmap/queue-management-rebuild.md).
- Conventions: [Go](llm/conventions/golang.md), [testing](llm/conventions/testing.md), [database](llm/conventions/database.md).
