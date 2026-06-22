# Domain Rules

## Current baseline

Live starter still enforces generic platform rules around auth, organization, Casbin, API keys, uploads, realtime, workers, and query safety.

Those rules still matter during rebuild because queue features will run inside this same infrastructure.

## Queue-management target rules from rebuild docs

### Queue registration

- registering queue must be idempotent for same active patient/appointment/menu/business-day combination
- active duplicate should return existing queue instead of blindly creating new row
- next queue number must be generated inside DB transaction with queue-counter row lock
- business day is not plain calendar day; it rolls around `04:00 Asia/Jakarta`

### Forward queue

- source queue must be locked before validation and mutation
- source state transition must be explicit and valid
- destination branch/menu/device relation must be validated before creation
- destination duplicate queue must be checked before creating new queue
- source queue completion/continued flags must update in same transaction as destination queue
- notification or integration side effects must happen only after commit or via safe outbox/worker path

### Scanner auth and check-in

- `x-client-id` and `x-api-key` presence alone is insufficient
- credential pair must be validated against persisted scanner/client record
- scanner request must be tied to correct branch/station/menu/device relation
- scanner handler should only classify flow, then call queue or forward usecase

### Pharmacy rule

- pharmacy flow cannot rely on regex like `(?i)resep`
- pharmacy semantics should come from explicit code such as `PENERIMAAN_RESEP` or explicit flag field
- source-to-target pharmacy transition rules must be centralized in one validator/service

### State machine rule

- queue status transitions should be centralized, not scattered through handlers
- invalid transitions should fail closed with explicit error path

## Repo-level rules still inherited from starter

- keep auth/session checks in middleware or usecase boundary
- keep tenant/org context first-class if queue app stays multi-tenant
- keep query/filter fields explicit and fail closed
- keep worker side effects retry-safe
- keep upload/realtime changes isolated from queue core unless requirement proves need

## Build-order rule

For queue rebuild, build in this dependency order:

1. domain names and schema
2. queue registration
3. forward transaction flow
4. scanner auth and check-in
5. external integration and consumer sync

Do not start from scanner or frontend first because they depend on queue core semantics.

## QMS Rebuild Addendum

Additional QMS domain rules:

- every business query and mutation carries `tenant_id`
- `branch_id` is never valid without tenant ownership validation
- one queue record must keep one `ticket_no` and one `queue_no` per day
- forward must append `queue_journeys`, not duplicate master queue rows
- `visit_journeys` is internal history, not external integration logic
- pharmacy validation may consult `queue_journeys` history
- queue reset time may inherit from tenant and be overridden by branch

## QMS TDD Addendum

For QMS rebuild features, TDD is mandatory by default.

Every new feature or update should add tests covering:

- positive behavior
- negative behavior
- edge behavior
- vulnerability/security behavior

Do not treat handler smoke tests alone as enough for queue domain work. At minimum, protect usecase/domain logic and add route/repository tests where boundary risk exists.
