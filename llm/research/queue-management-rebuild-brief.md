# Queue Management Rebuild Brief

## Purpose

This repo is still a generic multi-tenant starter in live runtime code, but the current product goal is a rebuild of a hospital/clinic queue-management system using:

- `documentation/task-overview.md`
- `documentation/project-diagram.md`
- `code_context.txt`

This file exists so future agent sessions do not mistake current starter runtime for final product intent.

## Source classification

### Verified from live starter code

- Runtime stack today is Go + Gin + GORM + Redis + Casbin + Asynq + TUS, wired from `internal/config/app.go`.
- Current starter route strata are defined in `internal/router/router.go`.
- Current domain modules are generic starter modules: auth, user, organization, role, permission, project, audit, stats, webhook, access, and api_key.
- No queue-management module is wired yet in `internal/config/app.go`.
- No queue or scanner route group is wired yet in `internal/router/router.go`.

### Verified from provided rebuild docs

From `documentation/task-overview.md` and `documentation/project-diagram.md`, target system must support:

- patient/customer queue registration with idempotent duplicate prevention
- transactional queue-number generation using row locking on queue counters
- forward queue flow with source locking, destination duplicate checks, and post-commit notification
- scanner check-in flow with real credential validation, station/menu relation checks, and delegation into queue/forward services
- pharmacy transition validation based on explicit menu semantics, not regex text matching
- queue-state transition rules centralized in service/state-machine logic
- business-day calculation around `04:00 Asia/Jakarta`

### Legacy evidence only

`code_context.txt` is legacy implementation evidence. Use it to recover naming, route patterns, and old module placement, but do not treat it as runtime truth for this starter repo.

## Key rebuild mismatch

Current starter shape and target product do not line up yet.

### Starter owns today

- generic auth / org / permission / project platform
- generic admin/user management surfaces
- generic upload, webhook, realtime, worker infrastructure

### Target rebuild needs next

- queue domain modules
- scanner / external integration boundary
- hospital branch/station/menu/device modeling
- patient appointment and forward-flow orchestration
- queue-specific transaction and idempotency logic

## Recommended target module split

Use docs as product target, but keep starter repo layering:

- `internal/modules/queue`
  - queue registration
  - queue counter allocation
  - duplicate detection
  - queue read/list/detail
- `internal/modules/queue_forward`
  - forward orchestration
  - source/destination validation
  - source status transition
  - post-commit notification trigger
- `internal/modules/scanner`
  - scanner credential validation
  - branch/station/menu/device relation checks
  - check-in orchestration that delegates to queue or forward services
- `internal/modules/menu` or `internal/modules/service_point`
  - branch/menu/station/loket/device metadata
  - pharmacy classification source of truth
- `internal/modules/patient` or `internal/modules/customer`
  - patient identity / appointment lookup
  - active queue search inputs
  - external doctor/schedule/appointment/patient/scanner adapters if still needed

Exact names can move, but business rules should not stay buried inside handlers.

## Non-negotiable architecture rules for rebuild

- handler/controller only binds request and returns response
- queue, forward, scanner, pharmacy, and transition logic live in usecase/service layer
- queue number generation must happen inside DB transaction with counter lock
- duplicate detection must happen before new queue insert
- notification/webhook side effects must happen after durable commit or safe outbox enqueue
- scanner auth must validate `x-client-id` and `x-api-key` against repository truth
- pharmacy rules must use explicit code/flag, not regex on menu names
- route protection must stay in router/middleware, not frontend-only guards

## First implementation priorities

1. Define target bounded contexts and naming in starter repo.
2. Add DB schema plan for queue counters, customers/queues, scanner clients, service points, and forward history.
3. Implement queue registration usecase with transaction + idempotency first.
4. Implement forward usecase with source lock and post-commit side effect handling.
5. Implement scanner auth + check-in orchestration only after queue/forward core exists.
6. Re-map frontend/API consumers after backend contract stabilizes.

## Guardrails for future sessions

- Do not rewrite starter auth/org/permission platform blindly unless target requirement proves it.
- Reuse starter cross-cutting infra where it helps: middleware, tx manager, worker, storage, realtime, config, logging.
- Do not copy old legacy package layout verbatim if it fights current repo conventions.
- When docs and old code disagree, prefer docs for target product and starter live code for current runtime constraints.
