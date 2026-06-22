# Queue Management Feature Map

## Goal

Translate all provided docs into build-ready feature slices that fit current starter architecture.

## Feature slices

### Slice A — Core domain foundation

Includes:

- business day helper
- queue state machine
- domain errors
- queue counter entity
- queue status enum
- explicit pharmacy classification fields

Target files:

- `internal/modules/queue/*`
- `internal/modules/queue_forward/*`
- shared domain helper package if needed
- future migration files in `db/migrations/*`

### Slice B — Queue registration

Includes:

- POST queue registration
- duplicate active queue check
- counter lock and increment
- idempotent existing-queue return
- queue estimate replacement

Target files:

- `internal/modules/queue/delivery/http/*`
- `internal/modules/queue/usecase/*`
- `internal/modules/queue/repository/*`

### Slice C — Forward orchestration

Includes:

- source lock
- destination validation
- pharmacy validator
- destination duplicate check
- source done/continued update
- post-commit notification trigger

Target files:

- `internal/modules/queue_forward/*`
- worker or webhook integration helpers if side effects are async

### Slice D — Scanner flow

Includes:

- credential validation
- scanner login/check-in
- branch/station/menu/device relation checks
- active queue lookup
- delegate to queue or forward usecases

Target files:

- `internal/modules/scanner/*`
- `internal/middleware/*` if auth becomes middleware

### Slice E — Platform compatibility

Includes:

- Casbin matcher alignment
- CORS hardening
- order sanitization
- debug route protection
- registration error handling
- role policy dedupe

Target files:

- `internal/config/*`
- `internal/router/router.go`
- current auth/role/permission modules

### Slice F — Integration adapters

Includes:

- scanner route surface
- internal notification dispatch
- optional internal appointment/patient adapter only if later required

Target files:

- `internal/modules/scanner/*`
- `internal/modules/queue_forward/*`

### Slice G — Consumer sync

Includes:

- frontend proxy updates if API contracts move
- shared API types
- docs/playbooks
- verification commands

Target files:

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`
- `llm/test-playbooks/*`

## Dependency order

1. Slice A
2. Slice B
3. Slice C
4. Slice D
5. Slice E
6. Slice F
7. Slice G

## Implementation rule

Do not let scanner code invent queue rules.
Queue and forward core own business truth; scanner only orchestrates.
