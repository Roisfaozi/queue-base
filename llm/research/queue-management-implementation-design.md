# QMS Implementation Design Addendum

This note augments existing starter implementation-design guidance.

## Design rule

Preserve starter repo conventions and add QMS domain rules on top.

## QMS-specific architecture

- tenant-first boundary
- branch-under-tenant
- one queue master row per ticket/day
- forwarding through `queue_journeys`
- `visit_journeys` as readable internal history
- settings inheritance: tenant -> branch -> service -> counter

## Ownership

- `queues` owns master ticket state
- `queue_journeys` owns forward/service-step history
- `visit_journeys` owns readable event stream
- `settings` owns overrides and effective configuration
