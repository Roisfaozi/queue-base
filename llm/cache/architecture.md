# Architecture Snapshot

Verified against the working tree on 2026-10-28. Approved specifications describe intended changes; the sources below describe current runtime.

## Composition and ownership

[cmd/api/main.go](../../cmd/api/main.go) starts the API. [internal/config/app.go](../../internal/config/app.go) constructs infrastructure and modules; [internal/router/router.go](../../internal/router/router.go) installs middleware and routes. There is no separate QMS binary in this repository.

Module ownership remains: thin HTTP delivery and validation, usecase business orchestration/transaction decisions, repository persistence and scoped queries, `module.go` cross-module adapters. Preserve existing transaction manager and transactional Casbin patterns. Do not import another module's usecase directly to bypass its constructor seam.

App wiring includes platform IAM/audit/webhook/stats and live QMS modules: queue, queue_config, service, counter, operator_assignment, qms_client, caller, signage, scanner, and organization branches. Queue receives its configuration resolver and realtime broadcasters through constructor wiring.

## Route strata

The router builds independent groups under `/api/v1`, not a single inheritance chain:

- `public`: public registration/auth/invitation flows and applicable rate limits.
- `authenticated`: API-key authentication, token validation, automatic API-key scope, `RequireUserSession`, user status, and applicable rate limit.
- `tenantAuthorized`: independently installs API-key authentication, token validation, automatic scope, user status, required organization, Casbin, and applicable rate limit. It does not inherit the authenticated group's explicit `RequireUserSession` call. Trace token/session behavior in middleware before claiming protection; valid JWT parsing alone is not authentication proof.
- `authorized`: explicit `admin:manage`, token/authentication, user status, optional organization, Casbin, and applicable rate limit.
- upload: separate TUS path with its own token/user-status handling.

QMS module route helpers mount under `tenantAuthorized`. Caller and signage additionally install [QMS client middleware](../../internal/middleware/qms_client_middleware.go) and client-type restrictions. Scanner check-in also mounts under the central tenant-authorized group. Device identity is not evidence that an endpoint is public.

## QMS aggregate and persistence

[Queue entities](../../internal/modules/queue/entity/queue_entity.go) define the master `queues`, progression `queue_journeys`, and internal readable `visit_journeys` history. Forwarding preserves master ticket identity and appends a journey.

[Queue repository](../../internal/modules/queue/repository/queue_repository.go) owns the registration transaction: scoped counter lock, number allocation, queue, initial journey, and visit persistence. [Queue usecase](../../internal/modules/queue/usecase/queue_usecase.go) owns lifecycle decisions, business date, configuration consumption, and side-effect decisions. [Queue module](../../internal/modules/queue/module.go) supplies cross-module relation validation.

[QueueConfigResolver](../../internal/modules/queue_config/queue_config_resolver.go) implements the queue's `SettingsResolver` interface. It checks counter, branch-service, branch, then tenant. Service-level configuration is tied to `branch_services`; do not invent an independent global service override.

## Consumers and verification

`apps/web` currently owns QMS pages. `apps/client` remains active but has no QMS routes. Both use app-specific proxies and shared `packages/*` where actual reuse exists. A changed shared contract requires consumer inspection, not automatic feature duplication.

High-risk seams: concurrent numbering, duplicate guard, state transitions, cross-tenant/branch/service/counter relations, reset boundary, device credentials, and post-transaction side effects. Follow [qms-domain](../../.agents/skills/qms-domain/SKILL.md) and [verification guidance](../conventions/verification-anti-patterns.md). Static inspection identifies wiring; it does not prove runtime authentication or transaction correctness.
