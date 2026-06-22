# Queue Management Implementation Design

## Summary

Build queue-management application on current starter by adding new bounded contexts instead of mutating all existing platform modules into queue logic.

Recommended ownership:

- `queue` owns queue registration, numbering, estimate, business day, and read paths
- `queue_forward` owns forward transaction, destination duplicate check, and post-commit side effects
- `scanner` owns credential validation and check-in orchestration only
- `service_point` owns branch/menu/loket/device/station metadata and pharmacy classification
- `customer` owns patient/customer lookup and queue history inputs

## Files read

- `AGENTS.md`
- `llm/cache/project-overview.md`
- `llm/cache/architecture.md`
- `llm/cache/backend-map.md`
- `llm/cache/module-map.md`
- `llm/cache/domain-rules.md`
- `internal/config/app.go`
- `internal/router/router.go`
- `documentation/task-overview.md`
- `documentation/project-diagram.md`
- `code_context.txt`

## Current runtime path

### Startup and composition

- `cmd/api/main.go` boots app
- `internal/config/app.go` wires logger, DB, Redis, worker, JWT, SSE, WS, Casbin, storage, and current modules
- queue-specific modules are not yet wired here

### Route ownership today

- `internal/router/router.go` defines public, authenticated, tenant-authorized, authorized, and upload route strata
- no queue-specific or scanner-specific routes are registered yet

### Current reusable platform boundaries

- auth/session via middleware and auth module
- tenant/org via organization module and tenant middleware
- Casbin and API-key enforcement
- worker/realtime/upload/storage infrastructure

## Proposed ownership

## Option A — Single mega `customer` module

Put queue, forward, scanner, pharmacy, and patient lookup inside one large `customer` module.

### Pros

- fewer initial folders
- closer to some legacy layout

### Cons

- mixes registration, forward, scanner, and metadata concerns
- harder testing boundaries
- scanner/integration logic risks leaking into queue core
- grows into handler-heavy god module

## Option B — Split queue + forward + scanner + service_point + customer

Separate business capabilities into focused modules.

### Pros

- matches provided architecture better
- keeps handlers thin and usecases clear
- easier transaction and concurrency testing
- scanner remains orchestration, not domain owner
- easier future frontend/API ownership mapping

### Cons

- more initial wiring work
- more constructor dependencies to define carefully

## Recommendation

Choose **Option B**.

It best matches `documentation/project-diagram.md` and avoids repeating old mixed-logic problems.

## Contract and data model impact

### New core entities

Recommended minimum initial entities:

- `queue_records`
  - patient/customer identity reference
  - appointment reference if any
  - branch id
  - menu id
  - station/device target fields
  - business date
  - queue number
  - status
  - continued / forwarded metadata
  - estimate fields if persisted
- `queue_counters`
  - branch id
  - menu id
  - business date
  - current number
  - unique composite key on `(branch_id, menu_id, business_date)`
- `scanner_clients`
  - client id/device id
  - api key hash or secret
  - branch id
  - active status
- `service_points` or split metadata tables
  - branch
  - menu
  - submenu
  - locket/counter
  - device
  - pharmacy classification flag/code
- `queue_transitions` or audit/history table if state history must be queryable

### Initial route contracts

Recommended first backend contracts:

- `POST /api/v1/queues`
- `GET /api/v1/queues`
- `GET /api/v1/queues/:id`
- `POST /api/v1/queues/:id/forward`
- `POST /api/v1/scanner/login` or scanner token exchange equivalent
- `POST /api/v1/scanner/check-in`

Exact final paths can move, but queue core should not be hidden under unrelated starter resources.

### Shared helper additions

Need shared domain helpers for:

- business day calculation around `04:00 Asia/Jakarta`
- queue state machine transitions
- domain error typing and response mapping

## Auth, tenant, Casbin, API-key implications

### Auth

Normal admin/operator queue management can reuse existing JWT/session middleware.

### Scanner auth

Scanner flow should not reuse weak header-presence checks. It needs either:

1. scanner-specific middleware validating `x-client-id` + `x-api-key`, or
2. scanner credential exchange endpoint that returns short-lived scanner token

Recommended path:

- validate headers against repository truth on scanner login/check-in boundary
- issue short-lived scanner token only if scanner requires repeated calls after login

### Tenant/org

If final product stays multi-tenant, queue tables and queries must carry organization scope at repository boundary.

If product is single-tenant hospital app, organization layer may remain only for platform/admin surfaces. This needs confirmation before migrations.

### Casbin

Use starter Casbin for admin/operator HTTP routes.

Do not force scanner machine-to-machine flows through identical human role paths if scanner credential model differs.

### API keys

Existing API-key module can still protect integration/admin endpoints, but scanner credentials should remain domain-specific rather than overloading generic API-key semantics unless that truly matches product rules.

## Verification strategy

### Phase 1 — schema and helpers

- migration review
- unit tests for business day helper
- unit tests for queue state machine

### Phase 2 — queue registration

- usecase tests for duplicate detection
- repository tests for counter fetch/update path
- integration test for concurrent queue creation

### Phase 3 — forward

- usecase tests for invalid status transitions
- usecase tests for pharmacy-rule acceptance/rejection
- integration test for destination duplicate and transaction rollback

### Phase 4 — scanner

- middleware/service tests for invalid credentials
- integration tests for existing queue vs forward vs new registration

### Phase 5 — contract and consumer sync

- route tests
- frontend proxy/type sync checks if frontend consumer exists

## Rejected shortcuts

- putting queue logic directly into handlers
- using regex menu-name matching for pharmacy truth
- sending notification before commit
- letting scanner module implement forward logic itself
- using `code_context.txt` layout as mandatory folder truth

## Open questions

- needs confirmation: final table names should preserve legacy names like `customers` or use clearer queue-oriented names
- needs confirmation: scanner token model vs one-request header auth model
- needs confirmation: whether frontend for queue operations belongs in `apps/web`, `apps/client`, or both
