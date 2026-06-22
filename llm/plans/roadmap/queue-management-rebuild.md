# Queue Management Rebuild Roadmap

## Goal

Rebuild queue-management application on top of current starter repo without losing starter strengths around middleware, transactions, workers, realtime, and workspace structure.

## Constraints

- Current runtime code is still generic starter, not queue app.
- Queue target comes from `documentation/task-overview.md`, `documentation/project-diagram.md`, and legacy evidence in `code_context.txt`.
- Live code remains source of truth for what already exists in this repo.
- New queue-domain facts should stay in `llm/research/` or `llm/plans/` until implemented and verified.

## Phase map

### Phase 0 — Rebuild framing

Owner files:

- `llm/research/queue-management-rebuild-brief.md`
- `llm/plans/roadmap/queue-management-rebuild.md`
- `llm/tasks/todo.md`

Outcome:

- shared target vocabulary
- starter-vs-target mismatch documented
- execution order locked before code churn

Verification:

- confirm referenced files exist
- confirm brief matches live starter wiring in `internal/config/app.go` and `internal/router/router.go`

### Phase 1 — Domain and schema design

Owner files:

- `llm/research/*queue*`
- `llm/plans/*queue*`
- future `db/migrations/*`
- target module `model` and `entity` packages

Work:

- choose final module names
- define queue, queue_counter, scanner_client, branch/menu/station/device, appointment/patient relationships
- map business-day and state-transition rules into explicit models
- define idempotency key inputs for queue registration and forward flows

Dependencies:

- phase 0 complete

Verification:

- schema review against docs
- migration naming and up/down pairing review

### Phase 2 — Core queue registration

Owner files:

- `internal/modules/queue/*`
- possible shared transaction helpers under `pkg/tx`
- route registration in `internal/router/router.go` or queue route files

Work:

- add queue registration endpoint/usecase
- calculate business day at `04:00 Asia/Jakarta`
- detect active duplicates before insert
- lock queue counter row and increment atomically
- return existing queue idempotently when duplicate exists

Dependencies:

- phase 1 schema ready

Verification:

- narrow package tests for usecase/repository
- integration test for concurrent registration and duplicate prevention

### Phase 3 — Forward queue orchestration

Owner files:

- `internal/modules/queue_forward/*` or forward slice inside queue module
- worker/audit/webhook helpers if side effects are emitted

Work:

- lock source queue row
- validate source state and target station/menu/device
- enforce pharmacy rule
- check destination duplicate
- create destination queue and update source queue in one transaction
- trigger notifications only after commit or safe enqueue

Dependencies:

- phase 2 core queue registration available

Verification:

- usecase tests for invalid transitions and duplicate destination flow
- integration test for transaction rollback and post-commit side effect order

### Phase 4 — Scanner auth and check-in

Owner files:

- `internal/modules/scanner/*`
- `internal/middleware/*` if scanner auth becomes middleware
- scanner route files

Work:

- validate `x-client-id` + `x-api-key` against repo truth
- verify branch/station/menu relation
- detect existing active queue
- delegate to queue registration or forward usecase
- avoid embedding forward logic in scanner handler

Dependencies:

- phase 2 and phase 3 complete

Verification:

- middleware/service tests for invalid credentials
- integration tests for existing queue, forward, and new registration cases

### Phase 5 — External integration and consumer sync

Owner files:

- integration adapters
- frontend proxy files
- shared API types
- docs/playbooks

Work:

- add internal notification adapter only if still needed after queue core stabilizes
- define stable request/response contracts
- update `apps/web`, `apps/client`, and `packages/api-types` when consumers exist
- add manual and automated verification playbooks

Dependencies:

- backend queue contracts stable

Verification:

- contract-level tests
- frontend typecheck/build after consumer sync
- manual playbook for scanner and queue flows

## Cross-cutting decisions to preserve

- use starter transaction manager instead of ad hoc transaction code
- keep controller thin
- keep route protection centralized
- keep side-effect order explicit
- keep business rules in named services/usecases, not regex and handler conditionals
- prefer explicit state machine helpers for queue status transitions

## Immediate next tasks

1. Freeze target naming for queue, forward, scanner, menu/service-point, and patient/customer domains.
2. Draft schema and API contract notes before coding.
3. Build queue registration slice first because scanner and forward both depend on it.
