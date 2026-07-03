# Typed Configuration Design — Implementation Coverage

> **Source**: `documentation/New Design Document — QMS MVP Operatio.md`
> **Generated**: 2026-07-02
> **Status**: Live-runtime audit against MVP operational design, with detailed gap register

---

## 4. NEW Relationship Design

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| tenants → tenant_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:54`, `internal/modules/settings/entity/qms_queue_settings_entity.go:3` |
| tenants → branches | ✅ done | `internal/modules/organization/entity/organization_entity.go:20` |
| branches → branch_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:71`, `internal/modules/settings/entity/qms_queue_settings_entity.go:20` |
| branches → branch_services → counters | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:29`, `db/migrations/000032_align_qms_typed_configuration.up.sql:49` |
| counters → counter_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:78`, `internal/modules/settings/entity/qms_queue_settings_entity.go:54` |
| services → service_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:84`, `internal/modules/settings/entity/qms_queue_settings_entity.go:38` |

---

## 5. Effective Configuration Resolution

### 5.1 Queue Config Resolution Order

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| tenant→branch→service→counter override chain | ✅ done | `internal/modules/settings/queue_settings_resolver.go:32-101` |
| queue_reset_time example | ✅ done | `internal/modules/settings/queue_settings_resolver.go:105-113`, test: `internal/modules/settings/usecase/settings_usecase_test.go:232` |
| ticket_prefix example | ✅ done | `internal/modules/settings/queue_settings_resolver.go:146` |
| estimated_duration example | ✅ done | `internal/modules/settings/queue_settings_resolver.go:142` |
| auto_call_next example | ✅ done | `auto_call_next` now exists in typed entities and resolver (`internal/modules/settings/entity/qms_queue_settings_entity.go:3`, `internal/modules/settings/queue_settings_resolver.go:12`, `internal/modules/settings/queue_settings_resolver.go:150`) |

### 5.2 Effective Config Response

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| effective config API exists | ✅ done | `internal/modules/settings/delivery/http/settings_routes.go:11` → `GET /settings/effective` |
| tenant.branch.queue nested shape | ⚠️ partial | Current response is flat `ResolvedQueueSetting` array; doc wants nested `{tenant:{}, branch:{}, queue:{}}` shape |
| source metadata (`source`, `is_overridden`) | ✅ done | `internal/modules/settings/model/settings_model.go:82-90` (`Source`, `Inherited` fields) |
| `can_override` / `can_reset` fields | ❌ missing | Not in current response model |
| `effective_logo_asset_id` in response | ❌ missing | Not returned by effective config endpoint |

---

## 6. API Design

### 6.1 Tenant Profile

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/tenant/profile` | ❌ missing | No dedicated endpoint. Org data accessible via `GET /api/v1/organizations/:id` |
| `PATCH /api/v1/tenant/profile` | ❌ missing | No dedicated endpoint. Update via `PUT /api/v1/organizations/:id` |
| Request body fields (`name`, `legal_name`, `address`, etc.) | ⚠️ partial | Entity fields exist in `internal/modules/organization/entity/organization_entity.go:15-26` but no typed profile endpoint |

### 6.2 Tenant Queue Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/tenant/queue-settings` | ❌ missing | No dedicated endpoint. Must read via generic `GET /api/v1/settings/:id` or `GET /api/v1/settings/resolve` |
| `PATCH /api/v1/tenant/queue-settings` | ❌ missing | No dedicated PATCH endpoint. Must write via `POST /api/v1/settings` (generic) |

### 6.3 Branch Profile

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/branches/{branch_id}/profile` | ❌ missing | Branch CRUD exists at `/api/v1/branches/:id` via `internal/modules/organization/delivery/http/branch_routes.go` but no typed profile endpoint |
| `PATCH /api/v1/branches/{branch_id}/profile` | ❌ missing | Same — uses generic branch update |

### 6.4 Branch Queue Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/branches/{branch_id}/queue-settings` | ❌ missing | Must use generic settings resolve |
| `PATCH /api/v1/branches/{branch_id}/queue-settings` | ❌ missing | Must use generic settings POST |
| `DELETE .../queue-settings/{field}` | ❌ missing | No reset-to-inherit API |

### 6.5 Service

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `POST /api/v1/services` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:11` |
| `GET /api/v1/services` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:12` |
| `GET /api/v1/services/{service_id}` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:13` |
| `PUT /api/v1/services/{service_id}` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:14` (doc says PATCH, actual is PUT) |
| `DELETE /api/v1/services/{service_id}` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:15` |
| Request body includes `type`, `default_estimated_duration`, pharmacy flags | ✅ done | `internal/modules/service/model/service_model.go:25-41` |

### 6.6 Branch Service

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/branches/{branch_id}/services` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:24` |
| `POST /api/v1/branches/{branch_id}/services` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:23` |
| `POST .../services/{service_id}/enable` | ⚠️ partial | Uses generic POST instead of dedicated enable/disable verbs |
| `POST .../services/{service_id}/disable` | ❌ missing | No disable verb; uses `is_active` field |
| `PATCH .../services/{service_id}` | ✅ done | `internal/modules/service/delivery/http/service_routes.go:25` as PUT |

### 6.7 Service Queue Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/services/{service_id}/queue-settings` | ❌ missing | No dedicated typed endpoint |
| `PATCH /api/v1/services/{service_id}/queue-settings` | ❌ missing | Must use generic settings |
| `DELETE .../queue-settings/{field}` | ❌ missing | No reset-to-inherit API |

### 6.8 Counter

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `POST /api/v1/branches/{branch_id}/counters` | ⚠️ partial | Doc wants nested under `/branches/`, actual is `POST /api/v1/counters` (`internal/modules/counter/delivery/http/counter_routes.go:9`) |
| `GET /api/v1/branches/{branch_id}/counters` | ⚠️ partial | Same — actual is `GET /api/v1/counters` |
| `GET /api/v1/branches/{branch_id}/counters/{counter_id}` | ⚠️ partial | Actual `GET /api/v1/counters/:id` |
| `PATCH /api/v1/branches/{branch_id}/counters/{counter_id}` | ⚠️ partial | Actual `PUT /api/v1/counters/:id` |
| `DELETE /api/v1/branches/{branch_id}/counters/{counter_id}` | ⚠️ partial | Actual `DELETE /api/v1/counters/:id` |
| Request body includes `branch_service_id` | ✅ done | `internal/modules/counter/model/counter_model.go:13` |

### 6.9 Counter Queue Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Dedicated typed endpoints | ❌ missing | All missing — must use generic settings |

### 6.10 Effective Config

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/branches/{branch_id}/effective-config` | ⚠️ partial | Actual endpoint is `GET /api/v1/settings/effective` with query params, not branch-prefixed path |
| `GET .../services/{service_id}/effective-config` | ⚠️ partial | Same — uses query params: `?branch_id=&service_id=` |
| `GET .../counters/{counter_id}/effective-config` | ⚠️ partial | Same — uses `?branch_id=&service_id=&counter_id=` |

---

## 7. Setup Wizard UX

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Step 1 — Tenant Profile wizard | ❌ missing | No wizard at all; profile editing via generic org CRUD |
| Step 2 — Tenant Queue Default | ❌ missing | No wizard; manual settings CRUD |
| Step 3 — Branch Profile | ❌ missing | No wizard; branch CRUD exists via admin forms |
| Step 4 — Branch Queue Override | ❌ missing | No wizard; queue-settings UI exists but standalone |
| Step 5 — Service Setup | ❌ missing | No wizard; service dialog exists standalone (`apps/web/src/components/dashboard/services/service-dialog.tsx:1`) |
| Step 6 — Enable Service for Branch | ❌ missing | No wizard; branch-service CRUD via admin forms |
| Step 7 — Counter Setup | ❌ missing | No wizard; counter dialog exists standalone (`apps/web/src/components/dashboard/counters/counter-dialog.tsx:1`) |

---

## 8. Validation Rules

### 8.1 Tenant Validation

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| cannot activate without address/city/province/phone/logo/timezone | ❌ missing | No activation guard in org usecase (`internal/modules/organization/usecase/organization_usecase.go`) |
| tenant status enum (draft/active/inactive/suspended) | ⚠️ partial | Entity has `OrgStatusActive/Inactive/Suspended/Draft` but activation rule not enforced |

### 8.2 Branch Validation

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| cannot activate without address/city/province/phone/running_text/timezone | ❌ missing | Branch entity has fields but no activation guard (`internal/modules/organization/usecase/branch_usecase.go:32-65`) |
| can activate without logo if tenant logo exists | ❌ missing | No logo fallback logic |
| effective logo fallback (branch→tenant) | ❌ missing | Not implemented in any endpoint |

### 8.3 Branch Service Validation

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| service not enabled → cannot queue | ✅ done | `internal/modules/scanner/usecase/relation_validator.go:40` |
| service not enabled → cannot forward | ✅ done | `internal/modules/queue/usecase/queue_usecase.go:385` (via validator) |

### 8.4 Counter Validation

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| counter from different branch rejected | ✅ done | `internal/modules/scanner/usecase/relation_validator.go:117` |
| counter with invalid branch_service rejected | ✅ done | `internal/modules/counter/usecase/counter_usecase.go:155` |
| cross-tenant branch service rejected | ✅ done | `internal/modules/service/usecase/branch_service_usecase.go:161` |

---

## 9. Queue Config Usage

### 9.1 Queue Date

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| queue_date before reset_time uses previous business date | ✅ done | `internal/modules/queue/usecase/queue_usecase.go:61-68` |

### 9.2 Ticket Prefix

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| prefix from typed config resolver | ✅ done | `internal/modules/settings/queue_settings_resolver.go:146` |
| numbering strategy (daily_branch_sequence, sequential, random) | ✅ done | `internal/modules/queue/usecase/queue_usecase.go:141-173` |

### 9.3 Estimated Duration

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| estimated_duration from typed resolver | ✅ done | Resolver supports `default_estimated_duration` key |
| duration used in queue flow | ⚠️ partial | Entity has field but runtime propagation not fully wired in all call/journey flows |

---

## 10. Audit Log Design

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| SERVICE_CREATE / UPDATE / DELETE | ✅ done | `internal/modules/service/usecase/service_usecase.go:165` |
| SERVICE_CREATE test | ✅ done | `internal/modules/service/usecase/service_usecase_test.go:33` |
| BRANCH_SERVICE_CREATE / UPDATE / DELETE | ✅ done | `internal/modules/service/usecase/branch_service_usecase.go:133` |
| BRANCH_SERVICE test | ✅ done | `internal/modules/service/usecase/branch_service_usecase_test.go:1` |
| COUNTER_CREATE / UPDATE / DELETE | ✅ done | `internal/modules/counter/usecase/counter_usecase.go:185` |
| COUNTER test | ✅ done | `internal/modules/counter/usecase/counter_usecase_test.go:307` |
| SETTING_CREATE / UPDATE / DELETE | ✅ done | `internal/modules/settings/usecase/settings_usecase.go:188` |
| SETTING test | ✅ done | `internal/modules/settings/usecase/settings_usecase_test.go:256` |
| QUEUE_REGISTER / FORWARD / transition audit | ✅ done | `internal/modules/queue/usecase/queue_usecase.go:275` |
| Integration audit visibility | ✅ done | `tests/integration/modules/qms_audit_integration_test.go:1` |
| E2E audit visibility | ✅ done | `tests/e2e/api/qms_audit_e2e_test.go:1` |
| Audit metadata (old_values, new_values, IP, user_agent) | ✅ done | `internal/modules/audit/usecase/audit_usecase.go:50-90` |

---

## 11. Error Logging Design

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Controller logging (`logError`) | ✅ done | `internal/modules/service/delivery/http/service_controller.go:158`, `internal/modules/service/delivery/http/branch_service_controller.go:80`, `internal/modules/counter/delivery/http/counter_controller.go:74`, `internal/modules/settings/delivery/http/settings_controller.go:78` |
| Usecase logging | ✅ done | Usecases use `audit.LogActivity` and return errors |
| Sensitive data rule documented | ⚠️ partial | Documented in design doc and `AGENTS.md` but no centralized filter in logging middleware |
| Required log context | ⚠️ partial | Standard logrus context propagation present but no explicit audit/log correlation ID |

---

## 12. Migration Strategy

### Step 1 — Add New Columns

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Add address/city/province/postal_code/phone/email/logo_asset_id/timezone to tenants | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:3-10` |
| Add address/city/province/postal_code/phone/email/logo_asset_id/running_text/timezone to branches | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:17-23` |

### Step 2 — Create Typed Settings Tables

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| tenant_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:54` |
| branch_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:71` |
| service_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:84` |
| counter_queue_settings | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:78` |
| branch_services | ✅ done | `db/migrations/000032_align_qms_typed_configuration.up.sql:29` |

### Step 3 — Backfill From Generic Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Backfill script | ❌ missing | No seed/backfill SQL script in `db/seeds/` |
| Mapping: settings → typed tables | ❌ missing | No automated migration for existing data |

---

## 16. Final MVP Operational Design Gap Audit — Sections 4 to 42

> Source audited: `documentation/New Design Document — QMS MVP Operatio.md`
> Scope: live backend/frontend/schema evidence only. Legacy docs are not runtime truth.
> Status terms: `done` means live code covers design intent; `partial` means foundation exists but endpoint/shape/rule is incomplete; `missing` means no live implementation found.

### 16.1 Core Architecture Principles — Section 4

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 4.1 Tenant first | done | QMS entities/usecases resolve and write `tenant_id`; queue list/stats require tenant context. | `internal/modules/queue/entity/queue_entity.go:22`, `internal/modules/queue/usecase/queue_usecase.go:60`, `internal/modules/queue/repository/queue_repository.go:168` |
| 4.2 Branch under tenant | done | Branch entity and branch-owned queue queries are tenant+branch scoped. | `internal/modules/organization/entity/branch_entity.go:13`, `internal/modules/queue/repository/queue_repository.go:168`, `internal/modules/queue/repository/queue_repository.go:210` |
| 4.3 One queue record per ticket/visit | done | Queue master has one `ticket_no`, one `queue_no`, and `current_journey_id`; journey/history separated. | `internal/modules/queue/entity/queue_entity.go:22`, `internal/modules/queue/entity/queue_entity.go:38`, `internal/modules/queue/entity/queue_entity.go:51` |
| 4.4 Forwarding uses queue journey | done | Forward updates current queue and appends next `queue_journey`; no new queue master row. | `internal/modules/queue/usecase/queue_usecase.go:351`, `internal/modules/queue/usecase/queue_usecase.go:377`, `internal/modules/queue/repository/queue_repository.go:240` |
| 4.5 No generic settings anywhere | partial | Typed resolver exists, but legacy `settings` table/module still exists and resolver supports fallback path. | `internal/modules/settings/entity/settings_entity.go:3`, `internal/modules/settings/queue_settings_resolver.go:47`, `internal/modules/settings/queue_settings_resolver.go:54` |
| 4.6 Profile data is not settings | done | Tenant/branch profile fields live on main organization/branch entities. | `internal/modules/organization/entity/organization_entity.go:18`, `internal/modules/organization/entity/organization_entity.go:21`, `internal/modules/organization/entity/branch_entity.go:16` |
| 4.7 Behavior config uses typed tables | partial | Typed tables exist, but design-named `branch_service_queue_settings` is absent; live code uses `service_queue_settings` and `counter_queue_settings`. | `db/migrations/000032_align_qms_typed_configuration.up.sql:54`, `db/migrations/000032_align_qms_typed_configuration.up.sql:90`, `db/migrations/000032_align_qms_typed_configuration.up.sql:108` |

### 16.2 Tenant, Branch, Services, Counters — Sections 7 to 15

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 7 Tenant design | partial | Tenant profile fields and status exist, but activation completeness rule is not enforced. | `internal/modules/organization/entity/organization_entity.go:15`, `internal/modules/organization/entity/organization_entity.go:30`, `internal/modules/organization/usecase/organization_usecase.go:55` |
| 8 `tenant_queue_settings` | done | Table/entity include defaults and tenant unique constraint. | `db/migrations/000032_align_qms_typed_configuration.up.sql:54`, `db/migrations/000032_align_qms_typed_configuration.up.sql:67`, `internal/modules/settings/entity/qms_queue_settings_entity.go:3` |
| 9 Branch design | partial | Branch profile fields exist; branch logo field exists; explicit branch-logo-fallback-to-tenant resolver is missing. | `db/migrations/000032_align_qms_typed_configuration.up.sql:14`, `internal/modules/organization/entity/branch_entity.go:22`, `internal/modules/organization/model/branch_model.go:20` |
| 10 `branch_queue_settings` | done | Nullable override fields and tenant+branch unique key exist. | `db/migrations/000032_align_qms_typed_configuration.up.sql:71`, `db/migrations/000032_align_qms_typed_configuration.up.sql:85`, `internal/modules/settings/entity/qms_queue_settings_entity.go:20` |
| 11 Services | partial | Service type/duration/pharmacy flags exist; audio/narrative fallback rules are not implemented. | `db/migrations/000032_align_qms_typed_configuration.up.sql:25`, `internal/modules/service/entity/service_entity.go:13`, `internal/modules/service/entity/service_entity.go:17` |
| 12 Branch services | done | Branch-service table and CRUD/usecase exist with tenant/branch/service binding. | `db/migrations/000032_align_qms_typed_configuration.up.sql:29`, `internal/modules/service/usecase/branch_service_usecase.go:31`, `internal/modules/service/repository/branch_service_repository.go:1` |
| 13 `branch_service_queue_settings` | missing | Design table by this name is absent; current schema has `service_queue_settings`, not branch-service-specific settings. | `documentation/New Design Document — QMS MVP Operatio.md:778`, `db/migrations/000032_align_qms_typed_configuration.up.sql:90`, `internal/modules/settings/entity/qms_queue_settings_entity.go:38` |
| 14 Counters | done | Counter has `branch_service_id`, display name, status, and branch-service validation. | `db/migrations/000032_align_qms_typed_configuration.up.sql:48`, `internal/modules/counter/entity/counter_entity.go:14`, `internal/modules/counter/usecase/counter_usecase.go:151` |
| 15 `counter_queue_settings` | done | Counter settings table/entity exist with nullable override fields. | `db/migrations/000032_align_qms_typed_configuration.up.sql:108`, `db/migrations/000032_align_qms_typed_configuration.up.sql:122`, `internal/modules/settings/entity/qms_queue_settings_entity.go:57` |

### 16.3 Queue, Journeys, Estimate, Counter Sequence — Sections 16 to 23

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 16 Queues | done | Queue master row has tenant, branch, date, ticket, number, status, current journey. | `internal/modules/queue/entity/queue_entity.go:22`, `internal/modules/queue/model/queue_model.go:7`, `db/migrations/000027_create_queues_and_journeys.up.sql:1` |
| 17 Queue journeys | done | Journey entity has queue, tenant, branch, service, counter, seq, status; forwarding enforces active journey guard. | `internal/modules/queue/entity/queue_entity.go:38`, `internal/modules/queue/repository/queue_repository.go:246`, `internal/modules/queue/repository/queue_repository.go:248` |
| 18 Mapping concept | done | Queue points to current journey and list endpoints filter journeys by branch/service/counter. | `internal/modules/queue/entity/queue_entity.go:32`, `internal/modules/queue/delivery/http/queue_routes.go:22`, `internal/modules/queue/delivery/http/queue_routes.go:23` |
| 19 Operational actions | done | Generic queue transition exists and caller-specific single action endpoint is implemented. | `internal/modules/queue/model/queue_model.go:38`, `internal/modules/caller/delivery/http/caller_routes.go:10`, `internal/modules/caller/usecase/caller_usecase.go:92` |
| 21 Queue-left and estimate response | partial | Queue stats and active journeys exist, but doc-level estimate response shape and estimate endpoint are absent. | `internal/modules/queue/usecase/queue_usecase.go:78`, `internal/modules/queue/usecase/queue_usecase.go:107`, `internal/modules/queue/model/queue_model.go:70` |
| 22 Visit journeys | done | Visit journey entity and event writes exist for register/forward/transition. | `internal/modules/queue/entity/queue_entity.go:51`, `internal/modules/queue/usecase/queue_usecase.go:367`, `internal/modules/queue/usecase/queue_usecase.go:415` |
| 23 `queue_counters` | partial | Atomic create/numbering exists in repository flow, but no explicit `queue_counters` table found. | `internal/modules/queue/usecase/queue_usecase.go:149`, `internal/modules/queue/repository/queue_repository.go:232`, `internal/modules/queue/repository/queue_repository.go:240` |

### 16.4 Caller, Signage, Client Credential Binding — Sections 28 to 34

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 28 `operator_counter_assignments` | partial | Migration/entity and runtime enforcement exist, but no admin CRUD/write path exists yet. | `db/migrations/000033_qms_mvp_alignment.up.sql:61`, `internal/modules/operator_assignment/entity/operator_counter_assignment_entity.go:13`, `internal/modules/caller/usecase/caller_usecase.go:225` |
| 29 `qms_clients` | partial | Migration/entity/repository/auth middleware and route bindings exist, but no admin CRUD/write path exists yet. | `db/migrations/000033_qms_mvp_alignment.up.sql:35`, `internal/modules/qms_client/entity/qms_client_entity.go:26`, `internal/middleware/qms_client_middleware.go:20`, `internal/router/router.go:240` |
| 30 `qms_client_credentials` | partial | Credential table, hash verification, and expiry enforcement exist, but no admin CRUD/write path exists yet. | `db/migrations/000033_qms_mvp_alignment.up.sql:49`, `internal/modules/qms_client/entity/qms_client_entity.go:38`, `internal/modules/qms_client/usecase/qms_client_authenticator.go:34` |
| 31 Caller login context binding | done | Two-step client credential + human operator session binding exists with operator assignment enforcement. | `internal/modules/caller/usecase/caller_usecase.go:33`, `internal/modules/caller/usecase/caller_usecase.go:225`, `internal/middleware/qms_client_middleware.go:20` |
| 32 Caller endpoints | done | `/caller/login`, `/caller/me`, and `/caller/queue-journeys/:journey_id/action` are implemented. | `internal/modules/caller/delivery/http/caller_routes.go:10`, `internal/modules/caller/delivery/http/caller_controller.go:1`, `internal/router/router.go:240` |
| 34 Signage endpoints | done | `/signage/me`, `/signage/current-calls`, and `/signage/queues` are implemented. | `internal/modules/signage/delivery/http/signage_routes.go:10`, `internal/modules/signage/usecase/signage_usecase.go:21`, `internal/router/router.go:241` |

### 16.5 Typed Behavior Consumption — Section 36

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 36.1 Queue reset time | done | Queue stats/register use resolver-provided reset time for business date. | `internal/modules/queue/usecase/queue_usecase.go:90`, `internal/modules/queue/usecase/queue_usecase.go:149`, `internal/modules/queue/usecase/queue_usecase.go:160` |
| 36.2 Ticket prefix | done | Ticket prefix resolved through settings resolver. | `internal/modules/queue/usecase/queue_usecase.go:161`, `internal/modules/settings/queue_settings_resolver.go:146` |
| 36.3 Estimated duration | partial | Fields exist; resolver handling is incomplete for nullable typed fields and no full estimate response wiring exists. | `db/migrations/000032_align_qms_typed_configuration.up.sql:59`, `db/migrations/000032_align_qms_typed_configuration.up.sql:94`, `internal/modules/settings/queue_settings_resolver.go:155` |
| 36.4 Audio | partial | Service audio schema and signage exposure exist, but effective config resolver does not project audio fallback yet. | `db/migrations/000035_add_service_audio_columns.up.sql:2`, `internal/modules/service/entity/service_entity.go:23`, `internal/modules/signage/usecase/signage_usecase.go:125` |
| 36.5 Narrative | partial | Service narrative schema exists, but runtime signage/effective resolver does not project narrative fields yet. | `db/migrations/000035_add_service_audio_columns.up.sql:4`, `internal/modules/service/entity/service_entity.go:25`, `internal/modules/settings/model/settings_model.go:76` |
| 36.6 Auto call next | done | `auto_call_next` wired in schema/entity/resolver and surfaced in effective config. | `documentation/New Design Document — QMS MVP Operatio.md:2248`, `internal/modules/settings/entity/qms_queue_settings_entity.go:3`, `internal/modules/settings/queue_settings_resolver.go:12`, `internal/modules/settings/delivery/http/settings_controller.go:63` |
| 36.7 Allow recall | partial | `allow_recall` columns and queue guard exist, but effective config does not expose per-source recall metadata beyond flat bool fields. | `db/migrations/000032_align_qms_typed_configuration.up.sql:62`, `internal/modules/queue/usecase/queue_usecase.go:470`, `internal/modules/settings/model/settings_model.go:68` |

### 16.6 Audit, Logging, UI, Tests — Sections 38 to 42

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 38 Audit setup events | partial | Service/branch-service/counter/settings/queue audit calls exist; client/caller/signage setup audit cannot exist because domains missing. | `internal/modules/service/usecase/service_usecase.go:165`, `internal/modules/service/usecase/branch_service_usecase.go:133`, `internal/modules/counter/usecase/counter_usecase.go:180` |
| 38 Queue events | done | Queue register/forward/transition emits visit journeys and audit. | `internal/modules/queue/usecase/queue_usecase.go:275`, `internal/modules/queue/usecase/queue_usecase.go:385`, `internal/modules/queue/usecase/queue_usecase.go:415` |
| 38 Audit metadata | partial | Audit supports values and request metadata in audit module, but no caller/signage client metadata yet. | `internal/modules/audit/usecase/audit_usecase.go:50`, `internal/modules/scanner/usecase/scanner_usecase.go:160`, `internal/modules/queue/usecase/queue_usecase.go:385` |
| 39 Error logging context | partial | Tenant/branch guards exist; no central QMS correlation/client context for caller/signage. | `internal/modules/queue/usecase/queue_usecase.go:60`, `internal/modules/scanner/usecase/scanner_usecase.go:75`, `internal/modules/settings/delivery/http/settings_controller.go:46` |
| 40.1 Dashboard/manage | partial | Backend CRUD/settings/stat endpoints exist; MVP setup wizard/typed manage surface is incomplete. | `internal/modules/settings/delivery/http/settings_routes.go:8`, `internal/modules/service/delivery/http/service_routes.go:8`, `internal/modules/counter/delivery/http/counter_routes.go:8` |
| 40.2 Queue operation | done | Register/list/forward/transition/visit journey routes exist. | `internal/modules/queue/delivery/http/queue_routes.go:11`, `internal/modules/queue/delivery/http/queue_routes.go:15`, `internal/modules/queue/delivery/http/queue_routes.go:16` |
| 40.3 Caller | missing | No caller module, route, context, page, or endpoint found. | `documentation/New Design Document — QMS MVP Operatio.md:2502`; `rg 'caller|Caller|/caller' internal apps` returned no implementation |
| 40.4 Signage | missing | No signage module, route, context, page, or endpoint found. | `documentation/New Design Document — QMS MVP Operatio.md:2550`; `rg 'signage|Signage|/signage' internal apps` returned no implementation |
| 42 Tests | partial | Queue/service/counter/settings tests exist; caller/signage/client credential/operator assignment tests are missing. | `internal/modules/settings/queue_settings_resolver_test.go:1`, `internal/modules/counter/usecase/counter_usecase_test.go:1`, `tests/e2e/api/qms_queue_e2e_test.go:1` |

### 16.7 Focus Gap Register

| Focus Area | Status | Required Next Work | Evidence |
|---|---|---|---|
| `branch_service_queue_settings` | missing | Add migration/entity/repository/resolver scope keyed by `tenant_id + branch_service_id`, or explicitly revise design to use `service_queue_settings`. | `documentation/New Design Document — QMS MVP Operatio.md:778`, `db/migrations/000032_align_qms_typed_configuration.up.sql:90` |
| `qms_clients` | missing | Add client table/domain for caller/signage/scanner/kiosk binding. | `documentation/New Design Document — QMS MVP Operatio.md:1824` |
| `qms_client_credentials` | missing | Add hashed credential table and auth middleware/resolver. | `documentation/New Design Document — QMS MVP Operatio.md:1889` |
| `operator_counter_assignments` | missing | Add assignment table/domain and validate operator/counter scope at caller login. | `documentation/New Design Document — QMS MVP Operatio.md:1773` |
| Caller action endpoint | missing | Add `/caller/action` endpoint resolving queue operations from bound caller context. | `documentation/New Design Document — QMS MVP Operatio.md:2002`, `internal/modules/queue/delivery/http/queue_routes.go:16` |
| Signage endpoint | missing | Add `/signage/me` and `/signage/current-calls` using client credential branch context. | `documentation/New Design Document — QMS MVP Operatio.md:2093`, `documentation/New Design Document — QMS MVP Operatio.md:2126` |
| Effective config response | partial | Existing endpoint resolves queue values, but response shape lacks full nested metadata and logo fallback. | `internal/modules/settings/delivery/http/settings_routes.go:12`, `internal/modules/settings/delivery/http/settings_controller.go:63` |
| Branch logo fallback | missing | Fields exist, but no resolver returns branch logo fallback to tenant logo. | `internal/modules/organization/entity/organization_entity.go:27`, `internal/modules/organization/entity/branch_entity.go:22` |
| Activation rules tenant/branch | partial | Status fields exist, but activation completeness validation is missing. | `internal/modules/organization/entity/organization_entity.go:30`, `internal/modules/organization/entity/branch_entity.go:25`, `internal/modules/organization/usecase/branch_usecase.go:132` |

### 16.8 Recommended MVP Implementation Order

1. Fix design/schema mismatch for `branch_service_queue_settings` before adding caller/signage, because effective config inheritance depends on this scope.
2. Add `qms_clients` and `qms_client_credentials` with hashed credential auth and tenant/branch/client-type constraints.
3. Add `operator_counter_assignments` and caller session context resolution.
4. Add `/caller/action` over existing queue transition/forward logic with context filtering.
5. Add `/signage/me` and `/signage/current-calls`, including branch logo fallback and running text.
6. Add activation validation for tenant/branch profile completeness.
7. Extend effective config response with nested metadata, `can_override`, `can_reset`, and `effective_logo_asset_id`.

### Step 4 — Backfill Defaults

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Default values for tenant_queue_settings | ⚠️ partial | Defaults exist in SQL but no backfill for existing tenants without row |
| Default running_text for existing branches | ❌ missing | No backfill |

### Step 5 — Deprecate Generic Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Phase A: typed tables active, old read-only | ⚠️ partial | Typed tables active, generic still writable (`/api/v1/settings POST`) |
| Phase B: old no longer read by QMS core | ✅ done | `QueueSettingsResolver` reads typed tables only (`internal/modules/settings/queue_settings_resolver.go:32-101`) |
| Phase C: drop old settings | ❌ missing | Generic `/settings` endpoints still active |

---

## 17. Testing Requirements

### Tenant Tests

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| cannot activate without address | ❌ missing | No tenant activation tests |
| cannot activate without city | ❌ missing | Same |
| cannot activate without province | ❌ missing | Same |
| cannot activate without phone | ❌ missing | Same |
| cannot activate without logo | ❌ missing | Same |
| profile update writes audit log | ❌ missing | No test |
| queue settings update writes audit log | ❌ missing | No test |

### Branch Tests

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| activation rules (address/city/province/phone/running_text) | ❌ missing | No branch activation tests |
| can activate without logo if tenant logo exists | ❌ missing | No logo fallback test |
| effective logo fallback | ❌ missing | No logo fallback test |
| running_text update audit | ❌ missing | No test |
| queue settings update audit | ❌ missing | No test |

### Effective Config Tests

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| branch inherits tenant value | ✅ done | `internal/modules/settings/usecase/settings_usecase_test.go:232` |
| branch overrides tenant value | ✅ done | `internal/modules/settings/usecase/settings_usecase_test.go:240` |
| branch ticket prefix inherits/overrides | ✅ done | `internal/modules/settings/usecase/settings_usecase_test.go:232` |
| service overrides branch/tenant | ✅ done | `internal/modules/settings/usecase/settings_usecase_test.go:248` |
| counter overrides service/branch/tenant | ✅ done | `internal/modules/settings/usecase/settings_usecase_test.go:256` |
| reset branch override → tenant value | ⚠️ partial | Resolver handles nullable but no explicit test |
| response includes source metadata | ✅ done | `ResolvedQueueSetting.Source` and `Inherited` tested |

### Relation Tests

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| service not enabled → cannot queue | ✅ done | `internal/modules/scanner/usecase/relation_validator_test.go:131` |
| service not enabled → cannot forward | ✅ done | Queue integration tests (train of logic) |
| counter from different branch rejected | ✅ done | `internal/modules/scanner/usecase/relation_validator_test.go:181` |
| counter with invalid branch_service rejected | ✅ done | `internal/modules/counter/usecase/counter_usecase_test.go:155` (validation path) |
| cross-tenant branch service rejected | ✅ done | `internal/modules/service/usecase/branch_service_usecase.go:161` |

---

## 18. NEW Architecture Decision

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| Use typed domain tables | ✅ done | `internal/modules/settings/entity/qms_queue_settings_entity.go:1` |
| Keep profile data on main entity | ✅ done | `internal/modules/organization/entity/organization_entity.go:15-26`, `internal/modules/organization/entity/branch_entity.go:11-26` |
| Keep running text on branch | ✅ done | `internal/modules/organization/entity/branch_entity.go:21` |
| Use branch services | ✅ done | `internal/modules/service/entity/service_entity.go:37`, `internal/modules/service/usecase/branch_service_usecase.go:1` |
| Counter references `branch_service_id` | ✅ done | `internal/modules/counter/entity/counter_entity.go:14` |
| Keep effective config API | ✅ done | `internal/modules/settings/delivery/http/settings_routes.go:11` |
| Generic settings only for non-core | ⚠️ partial | Generic still writable; no read guard for core QMS |

---

## 19. NEW Recommended Table List for MVP

| Table | Status | Evidence |
|-------|--------|----------|
| `tenants` (`organizations`) | ✅ done | `internal/modules/organization/entity/organization_entity.go:13` |
| `tenant_queue_settings` | ✅ done | `internal/modules/settings/entity/qms_queue_settings_entity.go:3` |
| `branches` | ✅ done | `internal/modules/organization/entity/branch_entity.go:9` |
| `branch_queue_settings` | ✅ done | `internal/modules/settings/entity/qms_queue_settings_entity.go:20` |
| `services` | ✅ done | `internal/modules/service/entity/service_entity.go:16` |
| `branch_services` | ✅ done | `internal/modules/service/entity/service_entity.go:37` |
| `service_queue_settings` | ✅ done | `internal/modules/settings/entity/qms_queue_settings_entity.go:38` |
| `counters` | ✅ done | `internal/modules/counter/entity/counter_entity.go:9` |
| `counter_queue_settings` | ✅ done | `internal/modules/settings/entity/qms_queue_settings_entity.go:54` |
| `queues` | ✅ done | `internal/modules/queue/entity/queue_entity.go` |
| `queue_journeys` | ✅ done | `internal/modules/queue/entity/queue_entity.go` |
| `visit_journeys` | ✅ done | `internal/modules/queue/entity/queue_entity.go` |
| `queue_counters` | ✅ done | `internal/modules/queue/repository/queue_repository.go`, `db/migrations/000032_align_qms_typed_configuration.up.sql` |
| `audit_logs` | ✅ done | `internal/modules/audit/entity/audit_log.go` |

---

## Summary

| Category | ✅ Done | ⚠️ Partial | ❌ Missing |
|----------|---------|------------|------------|
| Section 4 — Relationship Design | 7 | 0 | 0 |
| Section 5 — Effective Config | 4 | 1 | 2 |
| Section 6 — API Design | 7 | 9 | 14 |
| Section 7 — Setup Wizard | 0 | 0 | 7 |
| Section 8 — Validation Rules | 4 | 1 | 4 |
| Section 9 — Queue Config Usage | 3 | 1 | 0 |
| Section 10 — Audit Log | 12 | 0 | 0 |
| Section 11 — Error Logging | 1 | 2 | 0 |
| Section 12 — Migration Strategy | 5 | 2 | 5 |
| Section 13 — Testing | 8 | 1 | 11 |
| Section 14 — Architecture Decision | 6 | 1 | 0 |
| Section 15 — Table List | 14 | 0 | 0 |

---

## Detailed Gap Register

This section expands the raw `partial/missing` markers into concrete implementation gaps against the newer MVP operational design.

### Gap A — Typed Config Scope Drift

**Design target**

- `branch_service_queue_settings` should exist and replace service-global queue override.
- Effective config should resolve `tenant -> branch -> branch_service -> counter`.

**Current runtime**

- Runtime still uses `service_queue_settings` instead of `branch_service_queue_settings`.
- Resolver still resolves `tenant -> branch -> service -> counter`.

**Evidence**

- Current table/entity: `db/migrations/000032_align_qms_typed_configuration.up.sql:90`
- Current entity: `internal/modules/settings/entity/qms_queue_settings_entity.go:38`
- Current resolver chain: `internal/modules/settings/queue_settings_resolver.go:32`

**Gap impact**

- Cannot model branch-specific service duration/audio/narrative/recall behavior correctly.
- Service override leaks across branches under same tenant.

**Needed change**

- Replace `service_queue_settings` with `branch_service_queue_settings` in schema, entities, resolver, API, tests, and frontend contracts.

### Gap B — Generic Settings Not Fully Removed

**Design target**

- Generic settings removed fully from QMS core, and in latest MVP migration step, removed first.

**Current runtime**

- `/api/v1/settings` CRUD still active.
- Generic settings still exist as runtime write surface.

**Evidence**

- Active routes: `internal/modules/settings/delivery/http/settings_routes.go:8`
- Legacy usecase still present: `internal/modules/settings/usecase/settings_usecase.go:21`

**Gap impact**

- Mixed configuration model remains possible.
- Future contributors may keep wiring core QMS behavior back into generic settings.

**Needed change**

- Freeze generic settings for QMS use, remove from QMS docs/playbooks, then deprecate or isolate the module.

### Gap C — Effective Config Response Too Thin

**Design target**

- Effective config should include branch/tenant presentation data and queue behavior details.
- Latest MVP also expects audio, narrative, auto_call_next, allow_recall.

**Current runtime**

- Effective config endpoint is flat and mostly queue-reset/prefix/strategy oriented.
- No effective logo fallback, no audio/narrative, no `allow_recall` metadata.

**Evidence**

- Route: `internal/modules/settings/delivery/http/settings_routes.go:11`
- Response model: `internal/modules/settings/model/settings_model.go:47`
- Resolver fields: `internal/modules/settings/queue_settings_resolver.go:146`

**Gap impact**

- Dashboard/caller/signage cannot rely on one complete effective config contract.
- UI must stitch values from separate APIs or cannot render final operational context correctly.

**Needed change**

- Expand effective config to include:
  - tenant summary
  - branch summary
  - effective logo fallback
  - queue reset / prefix / duration / forward / skip / recall / cancel
  - audio and narrative
  - `auto_call_next`

### Gap D — Tenant/Branch Activation Rules Missing

**Design target**

- Tenant and branch cannot become active before required profile completeness rules pass.

**Current runtime**

- Status enums exist, but no activation guard enforcing profile completeness.

**Evidence**

- Tenant entity fields: `internal/modules/organization/entity/organization_entity.go:13`
- Branch entity fields: `internal/modules/organization/entity/branch_entity.go:9`
- Branch usecase lacks activation validation: `internal/modules/organization/usecase/branch_usecase.go:27`

**Gap impact**

- Incomplete tenants/branches can be marked active.
- Caller/signage/dashboard flows can operate on operationally invalid branch data.

**Needed change**

- Add pre-activation validation in tenant and branch usecases.
- Add audit and tests for failed and successful activation.

### Gap E — Branch Logo Fallback Missing

**Design target**

- Branch logo optional; when empty, effective logo should fallback to tenant logo.

**Current runtime**

- Fields exist, but no effective fallback logic exposed in API.

**Evidence**

- Branch fields: `internal/modules/organization/entity/branch_entity.go:19`
- Tenant fields: `internal/modules/organization/entity/organization_entity.go:24`
- No effective-logo response in settings/controller path.

**Gap impact**

- Signage and dashboard cannot render guaranteed logo source from one contract.

**Needed change**

- Add effective-logo computation and response fields.
- Add tests for branch-logo-present and branch-logo-null fallback cases.

### Gap F — Caller Domain Missing

**Design target**

- `qms_clients`, `qms_client_credentials`, `operator_counter_assignments`.
- Caller context auto-resolves tenant/branch/branch_service/counter.
- Single action endpoint: `POST /api/v1/caller/queue-journeys/{journey_id}/action`.

**Current runtime**

- Old scanner/client API-key patterns exist.
- No `qms_clients` domain.
- No caller action endpoint.
- No operator assignment domain.

**Evidence**

- New MVP headings: `documentation/New Design Document — QMS MVP Operatio.md:1771`, `documentation/New Design Document — QMS MVP Operatio.md:1824`, `documentation/New Design Document — QMS MVP Operatio.md:1889`, `documentation/New Design Document — QMS MVP Operatio.md:2002`
- Current router has no `/caller/*` namespace: `internal/router/router.go:223`

**Gap impact**

- MVP caller flow is not implemented.
- Current scanner/API-key flow is not enough to satisfy bound caller operational UX.

**Needed change**

- Add new backend domains and endpoints for caller login, bound context, queue list, and action dispatch.

### Gap G — Signage Domain Missing

**Design target**

- Signage authenticates by client credential.
- Signage feed resolves branch/service/counter and returns running_text, effective logo, and current calls.

**Current runtime**

- No `signage/me` or `signage/current-calls` endpoints.

**Evidence**

- Design API: `documentation/New Design Document — QMS MVP Operatio.md:2093`
- No signage module/route in current backend router.

**Gap impact**

- MVP signage flow not implemented.

**Needed change**

- Add signage auth and feed endpoints plus view contract.

### Gap H — Queue Journey State Machine Partial vs MVP

**Design target**

- Typed journey timestamps replace old generic state timestamps.
- Repeated `call` becomes internal recall.
- Skipped journey can be called again.
- Forward/complete/skip/cancel follow strict state machine.

**Current runtime**

- Queue transitions exist, but latest MVP-specific caller semantics and recall semantics are not fully proven against new caller API.

**Evidence**

- Current queue usecase: `internal/modules/queue/usecase/queue_usecase.go`
- Existing tests centered on queue and scanner, not caller state machine contract.

**Gap impact**

- Core queue foundation exists, but MVP user-facing action semantics remain only partially aligned.

**Needed change**

- Rework and expand tests around typed state machine and caller single-action endpoint.

### Gap I — Estimate, Audio, Narrative, and Recall Coverage Missing

**Design target**

- Effective config and runtime should support estimated duration, audio fallback, narrative fallback, `allow_recall`.

**Current runtime**

- Duration partially exists.
- Audio/narrative fallback not implemented in effective resolver.
- `allow_recall` runtime guard exists, but response metadata is still partial.

**Evidence**

- Service entity has duration field: `internal/modules/service/entity/service_entity.go:23`
- No audio/narrative fields in effective resolver path.
- `auto_call_next` now exists in typed entities and effective config path.

**Gap impact**

- Caller/signage/dashboard cannot reflect full operational behavior promised by MVP design.

**Needed change**

- Expand schema and resolver, then add unit/integration/e2e tests.

### Gap J — Migration Strategy Incomplete For MVP

**Design target**

- Remove generic settings first.
- Add typed config tables.
- Add tenant/branch profile fields.
- Update services.
- Create branch service.
- Create qms clients and assignments.

**Current runtime**

- Typed table migration only covers earlier typed-config subset.
- No migration for `branch_service_queue_settings`, `qms_clients`, `qms_client_credentials`, `operator_counter_assignments`.

**Evidence**

- Current migration file: `db/migrations/000032_align_qms_typed_configuration.up.sql:1`
- Missing new MVP tables in `db/migrations/`

**Gap impact**

- Latest MVP cannot be reached incrementally from current schema set.

**Needed change**

- Create new migration wave aligned to sections 13, 28, 29, 30, and 41 of MVP doc.

### Gap K — Coverage Matrix Outdated vs MVP

**Design target**

- Test matrix now includes caller, signage, client credential security, estimate logic, timestamp behavior, and branch-service scoped config.

**Current runtime**

- Coverage doc still centered on older typed-config sections 4–15.
- Test suites do not yet cover new MVP sections 28–42.

**Evidence**

- Current coverage file scope: this file
- Current test playbooks still mention old settings inheritance/service scope patterns.

**Gap impact**

- Team can incorrectly assume coverage is adequate while major MVP domains are still untested.

**Needed change**

- Expand this audit and downstream playbooks to cover:
  - qms client credential auth
  - operator assignment
  - caller single action endpoint
  - signage feed
  - estimate time math
  - typed timestamp behavior
