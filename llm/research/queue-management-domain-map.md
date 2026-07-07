# QMS Domain Map Addendum

Latest design priority:

- `documentation/New Design Document — QMS MVP Operatio.md`

This note supplements existing starter domain map guidance.

## New QMS domain slices

- tenants
- branches
- services
- branch_services
- branch_service_queue_settings
- counters
- qms_clients
- qms_client_credentials
- operator_counter_assignments
- queues
- queue_journeys
- visit_journeys
- typed queue settings
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
