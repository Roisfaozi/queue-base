# Architecture Snapshot

## Important split

For this repo, architecture understanding now has two layers:

1. current starter runtime architecture from live code
2. target queue-management architecture from rebuild docs

Do not collapse them into one story.

## Current starter runtime architecture

Primary runtime composition lives in:

- `cmd/api/main.go`
- `internal/config/app.go`
- `internal/router/router.go`
- `internal/config/config.go`

Current composition root wires:

- logger
- validator
- GORM database connection
- Redis client
- Asynq distributor, processor, and scheduler
- JWT manager
- SSE manager
- WebSocket manager, presence manager, and ticket manager
- Casbin enforcer
- storage provider
- TUS handler and hook registry
- SSO providers
- generic starter modules: auth, user, organization, permission, role, access, project, stats, audit, webhook, api_key

Current route strata are:

- public auth/public user/public organization flows
- authenticated flows with API-key + token + session + user status
- tenant-authorized flows with organization + Casbin
- admin/authorized flows with `admin:manage`
- upload flow with token + user status only

## Target queue-management architecture from docs

Rebuild docs describe a different business architecture:

- queue registration service
- queue forward service
- scanner check-in service
- pharmacy validation service
- queue state machine
- queue/customer repositories
- queue counter repository
- internal notification boundary after forward commit

Core product rule from docs: handlers must stay thin while queue, forward, scanner, pharmacy, and transition logic move into service/usecase layer.

## Architecture consequences for rebuild

### Reuse from starter

Keep using starter infra where it already solves cross-cutting concerns:

- middleware layering
- transaction manager
- background workers
- realtime channels
- storage and upload infra
- config loading and env handling
- validation, logging, and telemetry

### Add for target

Need new queue-domain bounded contexts with clear ownership for:

- queue registration and numbering
- forward transaction orchestration
- scanner auth/check-in
- service-point metadata such as branch/menu/station/device
- patient/customer + appointment lookup inputs

## High-risk architectural boundaries

For queue rebuild, treat these as hard boundaries:

- queue number generation under concurrency
- source/destination queue state transitions
- post-commit notification timing
- scanner credential validation
- pharmacy-rule source of truth
- business-day calculation around `04:00 Asia/Jakarta`
- external integration adapters vs internal domain rules

## Truth order for future sessions

When rebuilding queue app:

1. live starter code for current capabilities and extension points
2. `documentation/task-overview.md` for target product rules
3. `documentation/project-diagram.md` for target decomposition and flow sequencing
4. `code_context.txt` for legacy evidence only when naming or old behavior matters

## QMS Rebuild Addendum

When task is QMS rebuild, add these architecture constraints to the existing starter architecture:

- tenant resolves before business logic
- branch always validated under tenant
- queue creation generates one `queues` master row and first `queue_journey`
- forwarding appends `queue_journeys` and preserves `ticket_no` / `queue_no`
- `visit_journeys` is internal readable event stream
- queue counters are scoped by tenant + branch + queue_date + prefix/menu strategy
- settings flow follows tenant -> branch -> service -> counter inheritance
