---
name: qms-domain
description: Use when changing queue registration, numbering, forwarding, journey transitions, business-date reset, inherited queue settings, caller/scanner/signage flows, operator assignments, or QMS device authentication in queue-base.
---

# QMS Domain

## Boundary

Use for QMS behavior across modules, persistence, and operational consumers. Pure IAM work uses `auth-tenant-casbin`; pure rendering uses `frontend-surface` and `ui`. Add `api-endpoint` for route changes, `database-transactions` for atomicity, and `cross-stack-change` for contracts. This skill does not authorize changing product scope or mirroring QMS into another app.

## Read order

1. [AGENTS.md](../../../AGENTS.md), [domain rules](../../../llm/cache/domain-rules.md), and the affected tests.
2. [App wiring](../../../internal/config/app.go) and [router](../../../internal/router/router.go) for the mounted production path.
3. [Queue entities](../../../internal/modules/queue/entity/queue_entity.go), [usecase](../../../internal/modules/queue/usecase/queue_usecase.go), [repository](../../../internal/modules/queue/repository/queue_repository.go), and [module adapters](../../../internal/modules/queue/module.go) for lifecycle/numbering/forwarding.
4. [QueueConfigResolver](../../../internal/modules/queue_config/queue_config_resolver.go) for settings/reset changes.
5. [QMS client middleware](../../../internal/middleware/qms_client_middleware.go) and the affected caller/scanner/signage routes for device flows.
6. [Testing conventions](../../../llm/conventions/testing.md) and [verification anti-patterns](../../../llm/conventions/verification-anti-patterns.md).

Read only sources relevant to the branch; verify code rather than treating old roadmap status as proof.

## Workflow

1. Name the operation, owning module, trusted tenant/branch/device context, current queue/journey state, and actual frontend consumer. Current QMS pages belong to `apps/web`; `apps/client` remains active but has no QMS routes.
2. Trace router, middleware, controller, constructor, usecase, repository, and side effects. Route groups are independent. Caller/signage add QMS-client authentication beneath tenant-authorized routes; scanner is not assumed public. Check persisted credentials and client type, not header presence alone.
3. Define the behavior delta and preservation constraints before editing. Keep one `queues` master; forward appends `queue_journeys`, preserving ticket identity. `visit_journeys` stays internal history/projection. Preserve atomic counter allocation, duplicate guards, valid transitions, and current relation validation.
4. For settings, distinguish `SettingsResolver` interface from `QueueConfigResolver` implementation. Check counter, branch-service, branch, tenant precedence and nullable fallback. Service overrides belong to `branch_services`. For reset, inspect current timezone/fallback; test before, at, and after reset rather than using calendar date.
5. Add failing-first table-driven coverage, make the smallest coherent change at the owning layer, then rerun focused checks. Keep transaction decisions in usecase/repository seams and cross-module adapters in `module.go`. Trace broadcasts/audit/workers relative to commit.

## Required cases

- **Positive:** valid registration, forwarding, transition, or override resolution.
- **Negative:** invalid payload/state or invalid device credential/type.
- **Edge:** reset boundary, duplicate/retry, nullable override, or concurrent allocation as applicable.
- **Security/vulnerability:** foreign tenant, mismatched branch-service/counter, or forged device/actor context rejected without a partial write or leak.

Mocked usecase tests do not prove row locks, rollback, or concurrent numbering. Add repository/integration proof when those semantics change. Preserve existing `tryAudit`, transaction manager, and `txEnforcer` behavior; do not import a different product's audit API or database convention.

## Verification

Start with `make test-custom RUN='<TestName>' PKG=./internal/modules/queue/...` or the affected QMS module. Use `make test-integration-custom RUN='<TestName>' PKG=./tests/integration/...` for real persistence and `make test-e2e-custom RUN='<TestName>' PKG=./tests/e2e/...` for mounted HTTP/security behavior. Docker is required for MySQL/Redis Testcontainers; report skipped execution.

For changed web consumers run `pnpm --filter casbin-web typecheck`; expand to client/shared checks only when actual consumers are affected. Both app lint scripts write files. `make test-all` always exits 0 and is not an acceptance gate.

## Stop conditions

Stop when intended transition, scope ownership, transaction boundary, or device protection is unclear. Report code/spec conflicts before changing behavior. Ask before destructive data/schema actions or product ownership changes. Do not repair unrelated runtime issues discovered during harness work.

## Completion

Report the operation and owners, preserved master/journey identity, four test categories, commands and actual results, skipped checks with blockers, and remaining unproved production boundaries.
