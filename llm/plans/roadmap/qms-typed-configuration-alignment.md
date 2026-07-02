# QMS Typed Configuration & MVP Alignment Plan

> **Design**: `documentation/New Design Document — QMS MVP Operatio.md`
> **Audit**: `llm/research/typed-config-design-coverage.md`
> **Rebased**: 2026-07-02 (replaces older generic-settings drafts)

## Dependencies Map

```text
Phase 1 (schema/entities)        → base layer, no dependency
       └── Phase 2A (branch-service + counter relation) → depends Phase 1
       └── Phase 2B (typed config resolver)              → depends Phase 1
              └── Phase 3 (queue hardening)              → depends Phase 2B
       └── Phase 2C (caller/signage/client-binding)      → depends Phase 1
              └── Phase 4A (caller endpoint)             → depends Phase 2C
              └── Phase 4B (signage endpoint)            → depends Phase 2C
       └── Phase 2D (audit/logging)                      → depends Phase 1
Phase 5 (frontend contract sync)     → depends Phase 4A, 4B, 3
Phase 6 (activation rules)           → depends Phase 1
Phase 7 (tests)         → can run parallel after each phase
```

## Parallel Worktree Slices

The following slices have **no overlapping write sets** and can be executed in parallel agents:

| Slice | Write Scope | Depends On |
|---|---|---|
| A — Schema + Entities | `db/migrations/*`, `internal/modules/*/entity/*` | — |
| B — Branch-Service + Counter Relations | `internal/modules/service/`, `internal/modules/counter/` | A |
| C — Typed Config Resolver | `internal/modules/settings/`, `pkg/settings/` | A |
| D — Client Binding + Auth Middleware | `internal/modules/auth/`, `internal/middleware/` | A |
| E — Audit/Logging | `internal/modules/audit/`, `internal/modules/*/usecase/` | A (D if caller logging) |

---

## Phase A — Schema/Entity Migration

### Step A1 — Migration SQL

**Owner Path**: `db/migrations/000033_qms_mvp_schema_alignment.up.sql`, `.down.sql`

**Work**:
- Add `branch_service_queue_settings` (tenant_id, branch_service_id, nullable overrides, unique(branch_service_id)).
- Add `qms_clients` table (id, tenant_id, branch_id, client_type VARCHAR(20), name, is_active, created_at, updated_at, deleted_at).
- Add `qms_client_credentials` table (id, client_id FK, client_secret_hashed, scope, last_used_at, expires_at, created_at).
- Add `operator_counter_assignments` table (id, tenant_id, user_id, counter_id FK, assigned_at, is_primary).
- Add composite indexes: `qms_client_credentials(client_id)`, `operator_counter_assignments(user_id, counter_id)`.

**Dependencies**: None

**Test Matrix**:
- Positive: `make migrate-up`, then `make migrate-down` rolls back clean.
- Negative: duplicate migration errors handled (no re-run on existing).
- Edge: fresh database migration runs clean without seed data.
- Vulnerability: migration does not expose plaintext secrets; `client_secret_hashed` is NOT NULL.

**Commit Strategy**:
- Commit 1: `db/migrations/000033_qms_mvp_schema_alignment.up.sql`
- Commit 2: `db/migrations/000033_qms_mvp_schema_alignment.down.sql`

### Step A2 — Entity Definitions

**Owner Path**:
- New: `internal/modules/qmsclient/entity/qms_client_entity.go`
- New: `internal/modules/qmsclient/entity/qms_client_credential_entity.go`
- New: `internal/modules/qmsclient/entity/operator_counter_assignment_entity.go`
- Update: `internal/modules/settings/entity/qms_queue_settings_entity.go` (add BranchServiceQueueSetting row, rename/alias service scope)
- Update: `internal/modules/scanner/entity/` (add client reference field if needed)

**Work**: Define GORM structs matching Step A1 tables. Use `soft_delete` only when true soft-delete is needed.

**Dependencies**: Step A1

**Test Matrix**:
- Positive: `make build` passes; entity struct compiles.
- Negative: none at entity layer.
- Edge: struct field types match migration column types (e.g. nullable fields use `*string`/`*bool`/`sql.NullInt64`).
- Vulnerability: `ClientSecretHashed` is never serialized via JSON tag (shall have `json:"-"`).

**Commit Strategy**:
- Commit 1: entity files for new QMS client/credential/assignment tables.
- Commit 2: update `qms_queue_settings_entity.go` to add BranchServiceQueueSetting.

---

## Phase B — Branch-Service + Counter Relation

**Owner Path**:
- `internal/modules/service/usecase/branch_service_usecase.go`
- `internal/modules/service/repository/branch_service_repository.go`
- `internal/modules/service/delivery/http/branch_service_routes.go`
- `internal/modules/counter/usecase/counter_usecase.go`
- `internal/modules/counter/repository/counter_repository.go`
- `internal/modules/counter/delivery/http/counter_routes.go`

**Work**:
- Add `POST /branches/{id}/services` if missing (activate service for branch).
- Add `DELETE /branches/{id}/services/{bs_id}` (deactivate branch service; cascade check on queues).
- Add `GET /services` filterable by branch_id.
- Validate `branch_service_id` on counter create/update: must match same branch.

**Dependencies**: Phase A

**Test Matrix**:
- Positive: branch service create returns service with `is_active=true`.
- Negative: deactive branch service with active ongoing queue returns 409.
- Edge: counter optional `branch_service_id` — unset meaning general counter.
- Vulnerability: cross-tenant branch service assign returns 403.

**Commit Strategy**:
- Commit 1: branch service active/deactivate routes + repository.
- Commit 2: counter branch-service validation update.
- Commit 3: branch-service list endpoint.

---

## Phase C — Typed Config Resolver

**Owner Path**:
- `internal/modules/settings/queue_settings_resolver.go`
- `internal/modules/settings/model/settings_model.go`
- `internal/modules/settings/delivery/http/settings_controller.go`

**Work**:
- Add `BranchServiceQueueSetting` resolver path under `resolveTypedDetailed`:
  - Override chain: counter → branch_service → branch → tenant
- Update `GET /settings/effective` to support new `branch_service_queue_settings` override.
- Add `branch_service_id` query parameter to effective config endpoint.
- Remove generic settings fallback when resolving core QMS keys.

**Dependencies**: Phase A

**Test Matrix**:
- Positive: branch service override applies; counter inherits from branch service.
- Negative: branch service with no row inherits branch setting.
- Edge: `branch_service_id` and `counter_id` both specified — counter wins.
- Vulnerability: missing tenant context returns 400.

**Commit Strategy**:
- Commit 1: resolver cascade logic + model.
- Commit 2: effective config controller update.
- Commit 3: remove generic fallback for QMS keys.

---

## Phase D — Client Binding + Auth Middleware

### Step D1 — QMS Client Credential Auth

**Owner Path**:
- New: `internal/modules/qmsclient/usecase/qms_client_usecase.go`
- New: `internal/modules/qmsclient/repository/qms_client_repository.go`
- New: `internal/middleware/qms_client_middleware.go`
- Update: `internal/router/router.go`

**Work**:
- Implement `Authenticate(clientID, clientSecret)` which checks `qms_client_credentials` hashed secret.
- Implement `ResolveClientContext` that returns `(tenantID, branchID, clientType)` for the session.
- Add middleware `QMSClientAuth()` that sets request context with `clientInfo`.
- Add caller/scanner route group that uses QMS client middleware before tenant middleware, allowing client pre-resolve.

**Dependencies**: Phase A

**Test Matrix**:
- Positive: valid clientID+secret returns client context with tenant+branch.
- Negative: wrong secret returns 401.
- Edge: expired credential returns 401.
- Vulnerability: missing client_type header returns 400.

### Step D2 — Operator Counter Assignment

**Owner Path**:
- New: `internal/modules/qmsclient/usecase/operator_assignment_usecase.go`
- New: `internal/modules/qmsclient/repository/operator_assignment_repository.go`

**Work**:
- Resolve operator assignment by `user_id` → `(counterID, branch_service_id)`.
- Expose `ResolveOperatorCounter(ctx) → (counter, branchService)` for caller endpoint.

**Dependencies**: Phase A

**Test Matrix**:
- Positive: operator with assignment returns counter+branch service.
- Negative: operator with no assignment returns 404.
- Edge: operator with multiple assignments returns primary only.
- Vulnerability: cross-tenant assignment returns 403.

### Step D3 — Route Registration

**Owner Path**: `internal/router/router.go`

**Work**:
- Add new route group under `/caller` (QMS client auth middleware, tenant middleware).
- Add new route group under `/signage` (QMS client auth middleware, tenant middleware, but client_type=signage only).
- Don't forget the admin-level `/api/v1/qms/clients` CRUD for provisionining.

**Dependencies**: Steps D1, D2

**Test Matrix**: route registration compiles; admin CRUD accessible via `admin:manage` scope.

**Commit Strategy**:
- Commit 1: auth middleware + repository.
- Commit 2: operator assignment.
- Commit 3: route registration.

---

## Phase E — Audit/Logging

**Owner Path**:
- `internal/modules/audit/usecase/audit_usecase.go`
- `internal/modules/qmsclient/usecase/*`
- `internal/modules/settings/usecase/settings_usecase.go`
- `internal/modules/queue/usecase/queue_usecase.go`

**Work**:
- Add QMS client events: `QMS_CLIENT_CREATE`, `QMS_CLIENT_CREDENTIAL_CREATE`, `OPERATOR_ASSIGNMENT_CREATE`.
- Add caller audit: `CALLER_CALL`, `CALLER_SERVE`, `CALLER_COMPLETE`, `CALLER_SKIP`, `CALLER_CANCEL`.
- Ensure `tryAudit` is present in all operational queue usecases (register, forward, transition).
- Add correlation ID to error log context in QMS client middleware.

**Dependencies**: Phase D (for caller audit) or Phase A (for schema audit)

**Test Matrix**:
- Positive: queue register emits `QUEUE_REGISTER` audit entry.
- Edge: audit logging failure does not fail the business transaction.
- Vulnerability: audit does not log `client_secret_hashed`.

**Commit Strategy**:
- Commit 1: audit events for new schema tables.
- Commit 2: caller audit events.
- Commit 3: correlation ID in middleware logs.

---

## Phase F — Queue Hardening

**Owner Path**:
- `internal/modules/queue/usecase/queue_usecase.go`
- `internal/modules/queue/repository/queue_repository.go`
- `internal/modules/queue/model/queue_model.go`

**Work**:
- Add `allow_recall` check in transition: if `call` action and queue status is `calling` and `allow_recall=false`, reject.
- Add `allow_skip` guard in skip transition.
- Add `allow_cancel` guard in cancel transition.
- Add auto-call-next logic after `complete` transition (optional, depends on config).

**Dependencies**: Phase C (resolver)

**Test Matrix**:
- Positive: queue call succeeds when `allow_recall=true`.
- Negative: recall fails when `allow_recall=false`.
- Edge: repeated `call` on already-calling queue with `allow_recall=true` updates journey timestamp.
- Vulnerability: cross-branch queue transition returns 403.

**Commit Strategy**:
- Commit 1: allow_* guards.
- Commit 2: auto-call-next.

---

## Phase G — Caller Endpoint

**Owner Path**:
- New: `internal/modules/caller/delivery/http/caller_controller.go`
- New: `internal/modules/caller/delivery/http/caller_routes.go`
- New: `internal/modules/caller/usecase/caller_usecase.go`
- New: `internal/modules/caller/model/caller_model.go`
- New: `internal/modules/caller/module.go`
- Update: `internal/config/app.go` (wire caller module)
- Update: `internal/router/router.go` (register caller routes under /caller)

**Work**:
- `POST /caller/action`
  - Accepts `{action, queue_id, destination_service_id?}`.
  - Resolves operator session from `operator_counter_assignments`.
  - Routes to `queueUseCase.TransitionQueue` (for call/serve/complete/skip/cancel) or `queueUseCase.ForwardQueue` (for forward).
  - Returns updated queue + journey.

**Dependencies**: Phase D (client binding + operator assignment), Phase F (allow_* guards)

**Test Matrix**:
- Positive: caller with valid assignment emits journey transition.
- Negative: caller with no operator assignment returns 403.
- Edge: forward action includes destination_service_id + auto-resolved counter.
- Vulnerability: caller scoped to different branch cannot act on queues from another branch.

**Commit Strategy**:
- Commit 1: caller model + usecase.
- Commit 2: caller controller + routes.
- Commit 3: module wiring in app.go + router.go.

---

## Phase H — Signage Endpoint

**Owner Path**:
- New: `internal/modules/signage/delivery/http/signage_controller.go`
- New: `internal/modules/signage/delivery/http/signage_routes.go`
- New: `internal/modules/signage/usecase/signage_usecase.go`
- New: `internal/modules/signage/model/signage_model.go`
- New: `internal/modules/signage/module.go`
- Update: `internal/config/app.go`
- Update: `internal/router/router.go`

**Work**:
- `GET /signage/me`:
  - Returns `{tenant: {name, logo}, branch: {name, running_text, logo}}`.
  - Logo fallback: if branch logo is nil, return tenant logo.
- `GET /signage/current-calls`:
  - Returns active call journeys for the signage branch, grouped by counter.
  - Journey status = `calling` or `serving`, limited to last N entries.

**Dependencies**: Phase D (client binding), Phase A (qms_client to resolve branch context)

**Test Matrix**:
- Positive: signage client returns branch display info.
- Negative: signage client with wrong client_type (caller) returns 403.
- Edge: branch logo nil → tenant logo returned in response.
- Vulnerability: signage client scoped to branch A cannot see branch B calls.

**Commit Strategy**:
- Commit 1: signage model + usecase.
- Commit 2: signage controller + routes.
- Commit 3: module wiring.

---

## Phase I — Activation Rules

**Owner Path**:
- `internal/modules/organization/usecase/organization_usecase.go`
- `internal/modules/organization/usecase/branch_usecase.go`
- `internal/modules/organization/model/organization_model.go`
- `internal/modules/organization/model/branch_model.go`

**Work**:
- Tenant activate: reject if `address`, `city`, `province`, `phone`, `logo_asset_id` empty.
- Branch activate: reject if `address`, `city`, `province`, `phone`, `running_text` empty.
- Logo fallback resolver: when serving branch profile, if `logo_asset_id` is nil/empty, inherit from tenant.

**Dependencies**: Phase A

**Test Matrix**:
- Positive: activation succeeds with all required fields.
- Negative: activation fails with missing address.
- Edge: branch activation with no logo (tenant has logo) succeeds, and profile returns tenant logo.
- Vulnerability: cross-tenant profile leak via fallback is guarded.

**Commit Strategy**:
- Commit 1: tenant activation guard.
- Commit 2: branch activation guard + logo fallback.
- Commit 3: logo fallback in service/branch profile response.

---

## Phase J — Frontend Contract Sync

**Owner Path**:
- `apps/web/src/lib/api/qms.ts`
- `apps/web/src/components/dashboard/caller/`
- `apps/web/src/components/dashboard/signage/`
- `apps/web/src/components/dashboard/counters/`
- `packages/api-types/*`

**Work**:
- Add TypeScript types for `BranchServiceQueueSetting`, `QmsClient`, `OperatorAssignment`.
- Add caller UI page (or rewire existing queue page) to use `POST /caller/action`.
- Add signage UI page with `GET /signage/me` + `GET /signage/current-calls`.
- Sync counter dialog to allow `branch_service_id` binding.
- Remove all generic `GET /settings` usage from core queue/settings pages.

**Dependencies**: Phase G, H, F, B

**Test Matrix**:
- Positive: `pnpm typecheck` passes across workspace.
- Negative: missing API returns error state, not empty data.
- Edge: empty signage feed shows "no active calls".
- Vulnerability: no UI-only auth; backend enforces client type and branch scope.

**Commit Strategy**:
- Commit 1: shared types.
- Commit 2: caller UI.
- Commit 3: signage UI.
- Commit 4: counter dialog update.

---

## Phase K — Tests (Integration + E2E)

**Owner Path**:
- `tests/integration/modules/qms_caller_integration_test.go` (new)
- `tests/integration/modules/qms_signage_integration_test.go` (new)
- `tests/integration/modules/qms_client_credential_integration_test.go` (new)
- `tests/e2e/api/qms_caller_e2e_test.go` (new)
- `tests/e2e/api/qms_signage_e2e_test.go` (new)
- `internal/modules/caller/usecase/caller_usecase_test.go`
- `internal/modules/signage/usecase/signage_usecase_test.go`
- `internal/modules/qmsclient/usecase/qms_client_usecase_test.go`

**Work**:
- Integration tests covering full request lifecycle with DB + Redis + Casbin.
- E2E tests for caller flow (register → call → serve → complete).
- E2E tests for signage: `/signage/me` and `/signage/current-calls`.

**Dependencies**: Phase G, H (can run after each phase's unit tests)

**Test Matrix**:
- Positive: full queue flow completes via caller endpoint.
- Negative: caller action with no assignment returns 403.
- Edge: signage returns empty current-calls when no active calls.
- Vulnerability: signage client type=caller returns 403 on signage endpoint.

**Can run parallel** with Phase J (no write overlap).

**Commit Strategy**:
- Commit 1: caller integration tests.
- Commit 2: signage integration tests.
- Commit 3: client credential integration tests.
- Commit 4: E2E caller flow.
- Commit 5: E2E signage flow.

---

## Cross-File Write Sets Per Agent

Each agent worktree must claim exactly one slice from the table below to avoid conflicts:

| Agent | Claimed Slice | Write Set |
|---|---|---|
| Agent 1 | Phase A — Schema/Entities | `db/migrations/`, `internal/**/entity/` |
| Agent 2 | Phase B — Branch-Svc + Counter | `internal/modules/service/`, `internal/modules/counter/` |
| Agent 3 | Phase C — Typed Config Resolver | `internal/modules/settings/` |
| Agent 4 | Phase D — Client Binding + Auth | `internal/modules/qmsclient/`, `internal/middleware/` |
| Agent 5 | Phase E — Audit/Logging | `internal/modules/audit/`, `internal/modules/*/usecase/` |
| Agent 6 | Phase F — Queue Hardening | `internal/modules/queue/` |
| Agent 7 | Phase G — Caller Endpoint | `internal/modules/caller/`, `internal/router/router.go`, `internal/config/app.go` |
| Agent 8 | Phase H — Signage Endpoint | `internal/modules/signage/`, `internal/router/router.go`, `internal/config/app.go` |
| Agent 9 | Phase I — Activation Rules | `internal/modules/organization/` |
| Agent 10 | Phase J — Frontend | `apps/web/`, `packages/api-types/` |
| Agent 11 | Phase K — Tests | `tests/`, `internal/modules/*/usecase/*_test.go` |

## Execution Flow

```text
Round 1 (parallel):  A, B, C, D, I
Round 2 (parallel):  E (after A), F (after C), G (after D), H (after D)
Round 3 (parallel):  J (after G+H+F), K (after G+H)
```

## Record Update

After each phase, append entry to `llm/tasks/qms-typed-config-progress.md` with:
- slice name + status
- owner paths
- design source
- work done summary
- test matrix per category
- verification command + result
- errors + fixes + lesson reference

If a fixable error surfaces, append to `llm/tasks/lessons.md` with:
- exact symptom
- root cause
- fix
- prevention rule
