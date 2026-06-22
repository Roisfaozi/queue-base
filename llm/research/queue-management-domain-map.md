# QMS Domain Map Addendum

This note supplements existing starter domain map guidance.

## New QMS domain slices

- tenants
- branches
- services
- counters
- queues
- queue_journeys
- visit_journeys
- settings
- scanner

## Rule

- `queues` = master ticket identity
- `queue_journeys` = forward and service-step history
- `visit_journeys` = readable projection/history
- forward does not create second master queue row

## QMS Testing Categories

Every domain slice should be backed by tests for:

- positive flow
- negative validation
- edge timing/state/config case
- vulnerability or tenant-isolation case
