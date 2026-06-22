# Todo Addendum

This note tracks the current QMS rebuild direction while leaving starter repo workflow intact.

## QMS items

- keep starter conventions as runtime truth
- add tenant/branch domain rules before queue logic
- model `queues` + `queue_journeys` + `visit_journeys`
- model settings inheritance in one reusable resolver
- keep forward journey-only; never create second master queue row

## Execution slices

1. tenant and branch foundation
   - owner: `internal/modules/organization`, `internal/middleware/tenant_middleware.go`, `internal/config/app.go`, `internal/router/router.go`
   - outcome: tenant context first, branch always validated under tenant, repo queries stay scoped
   - verify: module wiring review + targeted tenant/organization tests

2. queue master and journey model
   - owner: new `internal/modules/queue` package, plus `queue_journeys` and `visit_journeys` subpackages
   - outcome: normalized queue aggregate, one queue row per ticket/day, forward appends journey only
   - verify: unit tests for numbering, duplicate guard, and forward transition

3. settings inheritance
   - owner: new `internal/modules/settings` package and queue-domain consumers
   - outcome: tenant -> branch -> service -> counter override resolution with one helper
   - verify: resolver tests for precedence and fallback

4. scanner orchestration and route exposure
   - owner: `internal/router/router.go`, scanner controller/usecase, queue route groups
   - outcome: thin handlers, domain logic in usecase, route contract ready for consumers
   - verify: route-level tests and request flow tests

5. migration and persistence alignment
   - owner: `db/migrations/*`, queue repositories, module wiring
   - outcome: schema matches normalized queue design and transaction boundaries
   - verify: migration review + repository tests with transaction coverage

## QMS TDD Reminder

For every implementation slice:

- start with failing tests when feasible
- cover positive / negative / edge / vulnerability cases
- do not mark feature done on happy-path only
