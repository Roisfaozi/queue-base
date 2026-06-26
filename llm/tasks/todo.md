# Todo Addendum

This note tracks the current QMS rebuild direction while leaving starter repo workflow intact.

## QMS items

- keep starter conventions as runtime truth
- add tenant/branch domain rules before queue logic
- model `queues` + `queue_journeys` + `visit_journeys`
- model settings inheritance in one reusable resolver
- keep forward journey-only; never create second master queue row

## Execution slices

1. tenant and branch foundation (Completed)
   - owner: `internal/modules/organization`, `internal/middleware/tenant_middleware.go`, `internal/config/app.go`, `internal/router/router.go`
   - outcome: tenant context first, branch always validated under tenant, repo queries stay scoped
   - verify: module wiring review + targeted tenant/organization tests

2. queue master and journey model (Completed)
   - owner: new `internal/modules/queue` package, plus `queue_journeys` and `visit_journeys` subpackages
   - outcome: normalized queue aggregate, one queue row per ticket/day, forward appends journey only
   - verify: unit tests for numbering, duplicate guard, and forward transition

3. settings inheritance (Completed)
   - owner: new `internal/modules/settings` package and queue-domain consumers
   - outcome: tenant -> branch -> service -> counter override resolution with one helper
   - verify: resolver tests for precedence and fallback

4. scanner orchestration and route exposure (Completed)
   - owner: `internal/router/router.go`, scanner controller/usecase, queue route groups
   - outcome: thin handlers, domain logic in usecase, route contract ready for consumers
   - verify: route-level tests and request flow tests

5. migration and persistence alignment (Completed)
   - owner: `db/migrations/*`, queue repositories, module wiring
   - outcome: schema matches normalized queue design and transaction boundaries
   - verify: migration review + repository tests with transaction coverage

6. frontend dashboard integration (Completed)
   - owner: `apps/web/src/app/[locale]/dashboard/*`
   - outcome: UI for Services, Counters, Queue Settings, and Queues live board under tenant scope.
   - verify: typecheck, biome lint, and component functionality.

7. caller flow and operational route behavior (Completed)
   - owner: `internal/modules/queue/*`, `internal/modules/scanner/*`, route layer, and QMS route-focused tests
   - outcome: branch-scoped queue register/list, journey reads, transitions, forward behavior, and scanner request contract aligned
   - verify: targeted controller/usecase tests plus QMS integration and E2E lifecycle checks

8. dashboard/admin support surface (Completed)
   - owner: `apps/web/src/app/[locale]/dashboard/queues/*`, `apps/web/src/components/dashboard/queues/*`, QMS API helper layer
   - outcome: queue dashboard consumer stays aligned with backend branch scope and active queue routes
   - verify: `apps/web` typecheck plus pre-commit frontend lint/typecheck hooks

9. final hardening + integration + E2E + security audit (Completed for current QMS scope)
   - owner: QMS queue/scanner/settings modules, targeted tests, and frontend consumer alignment
   - outcome: positive, negative, edge, and security-oriented QMS route checks covered in current queue/scanner slice work
   - verify: targeted Go package tests, `TestQMSQueueIntegration_LifecycleAndSettingsGuard`, and `TestQMSQueueE2E_LifecycleAndScannerGuard`

## QMS TDD Reminder

For every implementation slice:

- start with failing tests when feasible
- cover positive / negative / edge / vulnerability cases
- do not mark feature done on happy-path only
