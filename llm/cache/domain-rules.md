# Domain Rules

Verified against the working tree on 2026-10-28. Preserve these QMS boundaries; compare any new requirement with live implementation before editing.

## Ticket identity and progression

- `queues` is the sole master ticket record. Registration creates the master and initial journey together. Forwarding appends `queue_journeys` and preserves master ticket identity; it never creates a second master ticket for the destination.
- `visit_journeys` is internal readable history/projection, not an external integration authority.
- Validate the current queue and journey state before transitions. Keep related writes atomic and preserve the production duplicate/idempotency checks.
- Number allocation locks the counter within the registration transaction. Scope counters by tenant, branch, business date, and prefix according to the current repository strategy; do not generate numbers in a handler or assume additional numbering strategies are implemented.

Sources: [entities](../../internal/modules/queue/entity/queue_entity.go), [repository](../../internal/modules/queue/repository/queue_repository.go), [usecase](../../internal/modules/queue/usecase/queue_usecase.go).

## Scope and trusted identity

- Every business query and mutation carries tenant context. A branch ID is valid only under its tenant; service/counter IDs also require their branch relationships.
- Forwarding validates destination branch, branch-service relation, and counter ownership through [module adapters](../../internal/modules/queue/module.go), not ID existence alone.
- Actors and device scope come from verified identity/context. Caller/signage add QMS-client authentication beneath tenant-authorized routes; scanner credentials must be validated through the actual mounted path. Headers alone are not authentication.
- Keep Redis session verification, user status, API-key scope, and Casbin boundaries intact. Route groups are independently mounted; see [architecture](architecture.md).

Sources: [router](../../internal/router/router.go), [device middleware](../../internal/middleware/qms_client_middleware.go).

## Typed settings and business date

- `QueueConfigResolver` implements `SettingsResolver`. Resolve overrides from counter, branch-service, branch, then tenant. Nullable overrides and fallback are meaningful; service settings join `branch_services` under tenant/branch scope.
- Business date is not necessarily calendar date. Current queue usecase defaults to a `04:00` reset in `Asia/Jakarta`; inspect effective settings and fallback before changing behavior. Test times immediately before, at, and after reset.
- Preserve explicit pharmacy policy and journey-history checks where implemented. Do not replace business semantics with display-name regex matching or import clinical stock/billing rules.

Sources: [resolver](../../internal/modules/queue_config/queue_config_resolver.go), [queue usecase](../../internal/modules/queue/usecase/queue_usecase.go).

## Platform seams

Keep strict query/filter allowlists, contextual logging, existing audit/worker ownership, transactional enforcer use, and retry-safe side effects. Preserve `tryAudit` behavior where present. Do not substitute an audit API or durability policy from a different product. Trace notifications, SSE/WebSocket broadcasts, and queued work relative to commit before changing transaction behavior.

## Required QMS proof

Feature/fix changes require table-driven positive, negative, edge, and vulnerability/security cases. Include cross-tenant or mismatched branch/service/counter access, invalid transitions, and reset-boundary cases where relevant. Concurrent number allocation needs persistence-level proof, not only a mocked usecase test.

Use [qms-domain](../../.agents/skills/qms-domain/SKILL.md), [testing conventions](../conventions/testing.md), and the actual unit/integration/E2E targets. A handler smoke test alone does not prove the domain or transaction boundary.
