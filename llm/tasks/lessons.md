# Lessons

## Phase 1

- Root repo is hybrid: Go backend core plus active `apps/web` and `apps/client` frontends.
- `package.json`, `go.mod`, `Makefile`, `.env.example`, Docker Compose, and CI are enough to ground toolchain analysis before deeper architecture work.
- Env ownership can be mapped concretely from `internal/config/config.go`, `apps/web`, and `apps/client` code usage.

## Phase 2

- `internal/config/app.go` is the highest-value file for runtime truth.
- `internal/router/router.go` is the clearest single file for route strata, middleware layering, and upload/realtime exposure.
- `internal/modules/*/module.go` files are the best compact view of real dependency boundaries.

## Phase 3

- Organization is the tenant backbone; many auth/permission behaviors depend on org/member context.
- Auth correctness depends on JWT plus Redis-backed session behavior, not JWT parsing alone.
- API key and Casbin layering must be reviewed together for protected routes.
- Sentinel guidance matters for WebSocket origin validation and reflection-based query safety.

## Phase 4

- Conventions in this repo are driven more by live module patterns and Makefile/CI than by style docs alone.
- Frontend apps are active; `apps/client` lint now runs Biome; `typecheck` remains the separate TypeScript gate.
- Integration/E2E validation expectations are strong because auth, tenant, worker, upload, and realtime flows are infrastructure-heavy.

## Phase 5+

- Proxy behavior in `apps/web` and `apps/client` is part of the real API contract surface and should be audited with backend route changes.
- Existing `documentation/llm/*` docs are helpful, but live code remains authoritative when there is drift.
- `documentation/api/AI_STREAMING_CONTRACT.md` currently reads as supporting/planned contract documentation, not confirmed live backend routing.

## TDT Migration Phase 1-3
- Refactored Repositories (P1), Usecases (P2), and Controllers (P3) to Table-Driven Testing (TDT) format.
- Targets included `settings`, `counter`, `service`, and `branch` (organization) modules.
- Migration adheres strictly to standard TDT structure (`t.Run` with `[]struct`).
- Maintains existing test coverage. All tests pass locally.

## QMS Typed Configuration Alignment

- Latest QMS design from `documentation/New Design — Typed Configuration Architecture for QMS.md` supersedes old generic-settings progress for core QMS configuration.
- Do not mark generic `settings` inheritance as complete for latest core QMS behavior; typed tables are required for tenant, branch, service, and counter queue configuration.
- Keep old `settings` only for experimental flags, temporary feature toggles, non-critical UI preferences, or future dynamic config.
- Before starting implementation, append a progress entry to `llm/tasks/qms-typed-config-progress.md` with owner paths, design source, tests, verification, errors, fixes, and next step.
- Every QMS typed-config implementation slice must use TDD when feasible and cover positive, negative, edge, and vulnerability/security cases.
- Record every fixed error as a lesson: exact symptom, root cause, fix, and prevention rule. Fixed errors must not disappear from project memory.
- Frontend progress is not current unless checked against live backend routes; `queue-settings` currently targets old generic settings behavior and must be realigned with typed effective configuration APIs.
- Branch/service/counter validation must use typed ownership relations: tenant owns branch and service, branch enables service through `branch_services`, and counter points to `branch_service_id`.
- When writing Go test assertions that call methods on composite literals inside short variable declarations, wrap the composite literal in parentheses first, for example `(BranchService{}).TableName()`.
- For new entity files, avoid carrying template imports like `soft_delete` unless the struct really uses soft-delete fields; narrow compile checks catch this fast.
- Strict `assert.Equal` on UnixMilli timestamps in repository tests can flake when using SQLite memory DBs if operations straddle a millisecond boundary; use `assert.InDelta(t, expected, actual, 5)` instead.
- If sandboxed runs block `git commit` due to read-only `.git/index.lock` across worktrees, record progress in the durable tracker and proceed without forcing source-control mutations.
