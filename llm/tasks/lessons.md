# Lessons

## QMS MVP Design Drift

- `documentation/New Design Document — QMS MVP Operatio.md` is now the highest-priority QMS design source.
- Earlier typed-config notes that still mention `service_queue_settings` should be treated as stale where they conflict with `branch_service_queue_settings`.
- New MVP scope also introduces `qms_clients`, `qms_client_credentials`, and `operator_counter_assignments`; future implementation plans and audits must include them.
- Caller/signage/client-binding flows are now first-class QMS runtime concerns, not optional future notes.

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
- Frontend progress is not current unless checked against live backend routes; `queue-config` currently targets old generic settings behavior and must be realigned with typed effective configuration APIs.
- Branch/service/counter validation must use typed ownership relations: tenant owns branch and service, branch enables service through `branch_services`, and counter points to `branch_service_id`.
- When writing Go test assertions that call methods on composite literals inside short variable declarations, wrap the composite literal in parentheses first, for example `(BranchService{}).TableName()`.
- For new entity files, avoid carrying template imports like `soft_delete` unless the struct really uses soft-delete fields; narrow compile checks catch this fast.
- Strict `assert.Equal` on UnixMilli timestamps in repository tests can flake when using SQLite memory DBs if operations straddle a millisecond boundary; use `assert.InDelta(t, expected, actual, 5)` instead.
- If sandboxed runs block `git commit` due to read-only `.git/index.lock` across worktrees, record progress in the durable tracker and proceed without forcing source-control mutations.
- When scaffolding new route domains, finalize route ownership first, then write controller and registration separately; avoid placeholder route helpers that duplicate groups or create temporary noop handlers.
- Avoid adding duplicate imports for same package with different aliases; staticcheck `ST1019` fails fast. Reuse one alias for all interfaces from same package.
- Remove placeholder-only statements before compile checks; a bare `nil` in Go switch branches causes `nil is not used`. Use `return nil` when branch intentionally has no typed value.
- For frontend contract sync, re-read exact local API type shape before patching; stale inferred field order caused the first `qms.ts` patch to fail.
- Do not hide a missing backend list endpoint behind `.catch(() => ({ data: [] }))`; render error state and wire UI to live contract instead.
- Keep QMS phase tracker chronological with implementation intent; if a later consumer slice lands first, immediately add the missing backend phase record so progress does not look skipped.
- Queue usecase must not call legacy config keys such as `prefix` or `numbering`; compatibility belongs in resolver/settings layer, while queue core uses typed design keys only.
- When adding a new shared type for cross-package resolver interfaces, define it in the `model` package, not in a non-model package, to keep import graphs clean and avoid unused imports.
- Effective config source metadata (`_source`, `_inherited`) gives the frontend enough data to render inheritance chain without needing a separate "resolve each key" flow.
- When adding adjacent TypeScript interfaces, verify the new interface is not nested inside another interface block; `tsc --noEmit` catches this as TS1131/TS1109 syntax errors.
- QMS docs sync must replace legacy generic-settings queue flow with typed effective-config flow; leaving `reset_time`/`prefix`/`numbering` in the core docs makes the design look older than runtime.
- Temporary patch files like `patch.txt` should not be kept in worktree; use direct `apply_patch` so docs slices stay clean.
- When the MVP operational design supersedes earlier typed-config drafts, rewrite the roadmap fully instead of layering patch notes on the old plan; mixed plan generations hide true dependencies and cause overlapping worktree write sets.
## 2026-07-02 — QMS typed-config MVP phase 1

- `branch_service_queue_settings` cannot be treated like old `service_queue_settings`; runtime resolution now needs `tenant_id + branch_id + branch_service_id` as stable lookup key.
- When router composition root adds a new module argument, immediately update `internal/router/router_test.go` or build will pass while package test fails.
- For sandboxed Go builds on this machine, use `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache` instead of `make build` to avoid read-only Go cache failures.
## 2026-07-02 — Signage skeleton follow-up

- For staged QMS rebuild work, route skeletons can land first, but progress log must mark auth/credential validation as pending so placeholder handlers are not mistaken for complete feature delivery.
## 2026-07-03 — QMS client auth wiring

- When adding new auth middleware path beside JWT and API key, prefer route-local middleware for dedicated surfaces like caller/signage instead of broad global wiring to avoid changing unrelated auth behavior.
- Placeholder credential flows should fail closed in usecase/controller paths once middleware lands; leaving dummy IDs after route guard wiring hides real auth bugs.
## 2026-07-03 — Caller boundary scoping

- When designing multi-layered hybrid auth (machine credential + human session), boundary enforcement must check both: the machine hardware context (qms_client binding) and the human operational context (operator assignment). Failing to check either introduces privilege bypass.

## 2026-07-03 — Signage context binding and GORM order-by footgun

- Usecase validation must happen at each boundary method, not only in middleware; signage `GetCurrentCalls`/`GetQueues` lacked tenant/branch context matching against client binding, allowing cross-tenant reads when middleware already validated.
- GORM `First(&dest).Error` appends `ORDER BY <table>.<pk>` even for raw `Select` on joined columns; with SQLite memory DB the missing column from an aliased `SELECT` expression became `ORDER BY branch_services.service_name` which fails. `Take` avoids this.
## 2026-07-03 — Caller login/me reuse existing auth stack

- problem: Caller login needed hybrid machine-plus-human auth without duplicating token/session code.
- root cause: no caller-specific session stack existed, but repo already had stable `auth` login/session flow.
- fix: reuse `authModule.AuthUseCase.Login`, then layer caller binding and operator assignment checks in caller usecase.
- lesson: for QMS caller login, keep machine binding in `qms_client` middleware/usecase and keep human session issuance in shared auth usecase; do not fork token logic unless caller session semantics truly diverge.
## 2026-07-03 — Do not add audit events for non-existent write paths

- plan originally listed `QMS_CLIENT_CREATE`, `QMS_CLIENT_CREDENTIAL_CREATE`, and `OPERATOR_ASSIGNMENT_CREATE` before repo had controller/usecase write path for those resources.
- adding audit events without a real mutation boundary creates fake completeness and dead code; Jalur C later closed that gap with real write paths.
## 2026-07-03 — Syncing design coverage with real implementation

- stale coverage docs cause agents to misreport missing features that were already built but not documented.
- `auto_call_next` queue transition behavior was already built in queue usecase (via `tryCallNextQueue`), but typed schema storage and effective config exposure was genuinely missing.
- always verify actual code paths, test files, and DB schemas before declaring a gap missing or completed.

## 2026-07-03 — Jalur C admin API minimal slice

- operator assignment had only entity + caller-side enforcement; backend UI needed real admin API to avoid manual DB seeding.
- qms client admin CRUD is enough as soft-deactivate + list/update; hard delete is unnecessary for MVP.
- when adding a new admin surface, keep audit on mutate-only paths and avoid introducing repository layers if direct GORM in usecase already matches repo style.

## 2026-07-06 — Reuse SSE for QMS queue realtime

- problem: caller/signage state needed live updates, but adding a new queue-specific realtime system would duplicate infrastructure.
- fix: reuse existing `pkg/sse.Manager` via a tiny `EventBroadcaster` interface on queue usecase.
- lesson: for QMS queue transitions, emit mutate-only SSE events from usecase after successful repository mutation; let frontend refresh state from existing read endpoints.

## 2026-07-06 — QMS logging belongs in usecase, not controller only

- problem: QMS admin and signage controllers already had logger injection, but usecases had no structured logs, so business-path failures were invisible beyond HTTP layer.
- root cause: module constructors passed `logrus.Logger` only to controllers and skipped usecase injection.
- fix: inject logger into usecase constructors and log start/error/success at business methods with tenant/client/resource context.
- affected paths:
  - `internal/modules/signage/usecase/signage_usecase.go`
  - `internal/modules/qms_client/usecase/qms_client_admin_usecase.go`
  - `internal/modules/operator_assignment/usecase/operator_assignment_usecase.go`

## 2026-07-06 — Rebase QMS docs immediately after runtime slices

- problem:
  - QMS progress and coverage docs drifted after logging, branch draft guard, and wizard blocker slices.
- root cause:
  - runtime slices were committed before handoff docs were rebased.
- lesson:
  - for QMS rebuild work, update `llm/tasks/qms-typed-config-progress.md` and `llm/research/typed-config-design-coverage.md` in same slice or immediately after, otherwise later audits will over-report missing gaps.
- action:
  - treat doc rebase as mandatory close-out step for each QMS runtime slice.

## 2026-07-07 — Queue Usecase Constructor Returns Interface

- context: adding unit coverage for internal `attachQueueEstimate` helper.
- error: `NewQueueUseCase(...)` returns `QueueUseCase`, so concrete helper methods are not callable on the returned value.
- fix: in same-package tests that need an unexported helper, type assert once to `*queueUseCase` after construction.
- prevention: prefer public behavior tests first; only use concrete type assertion for focused internal helper coverage that avoids wider fixture setup.
