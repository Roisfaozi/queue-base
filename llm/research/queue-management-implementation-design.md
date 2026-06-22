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

## QMS TDD and Verification

Preferred implementation loop:

1. write failing test
2. implement minimal passing behavior
3. refactor with tests green

For each feature/update, cover:

- positive case
- negative case
- edge case
- vulnerability/security case

Examples for QMS:

- positive: queue create under valid tenant/branch
- negative: invalid service/counter under branch rejected
- edge: reset-time boundary around `04:00`
- vulnerability: cross-tenant queue lookup blocked
