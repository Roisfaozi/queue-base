# Typed Configuration Design — Implementation Coverage

> **Source**: `documentation/New Design Document — QMS MVP Operatio.md`
> **Generated**: 2026-07-02
> **Last Rebased**: 2026-07-06 (latest live audit in this thread)
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
| services → branch_service_queue_settings | ✅ done | `db/migrations/000033_qms_mvp_alignment.up.sql:3`, `internal/modules/settings/entity/qms_queue_settings_entity.go:39` |

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
| `GET /api/v1/tenant/queue-settings` | ⚠️ deferred | Skipped for MVP. Read effective queue config through `GET /api/v1/settings/effective`. |
| `PATCH /api/v1/tenant/queue-settings` | ✅ done | Generic settings fallback removed; endpoints deleted. |

### 6.3 Branch Profile

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/branches/{branch_id}/profile` | ❌ missing | Branch CRUD exists at `/api/v1/branches/:id` via `internal/modules/organization/delivery/http/branch_routes.go` but no typed profile endpoint |
| `PATCH /api/v1/branches/{branch_id}/profile` | ❌ missing | Same — uses generic branch update |
| Branch CRUD UI | ✅ done | `apps/web/src/app/[locale]/dashboard/branches/page.tsx`, `apps/web/src/app/[locale]/dashboard/branches/_components/branches-content.tsx` |
| Branch activation frontend validation | ✅ done | Branches page uses shared branch activation contract before sending active status updates |

### 6.4 Branch Queue Settings

| Sub-Point | Status | Evidence |
|-----------|--------|----------|
| `GET /api/v1/branches/{branch_id}/queue-settings` | ⚠️ deferred | Skipped for MVP. Read effective queue config through `GET /api/v1/settings/effective`. |
| `PATCH /api/v1/branches/{branch_id}/queue-settings` | ✅ done | Generic settings fallback removed; endpoints deleted. |
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
| `GET /api/v1/services/{service_id}/queue-settings` | ⚠️ deferred | Skipped for MVP. Read effective queue config through `GET /api/v1/settings/effective`. |
| `PATCH /api/v1/services/{service_id}/queue-settings` | ✅ done | Generic settings fallback removed; endpoints deleted. |
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
| Dedicated typed endpoints | ✅ done | Generic settings fallback removed; endpoints deleted. |

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
| Step 3 — Branch Profile | ⚠️ partial | Setup wizard handles branch draft status block and links to Branch CRUD UI; inline embedded branch form not built |
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
| cannot activate without address/city/province/phone/running_text/timezone | ✅ done | Branch create guard added in `branch_usecase.go` |
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
| branch_service_queue_settings | ✅ done | `db/migrations/000033_qms_mvp_alignment.up.sql:3` |
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
| 4.5 No generic settings anywhere | ✅ done | Generic settings module and table fully removed. Resolver uses strict typed path. | `internal/modules/settings/queue_settings_resolver.go:32` |
| 4.6 Profile data is not settings | done | Tenant/branch profile fields live on main organization/branch entities. | `internal/modules/organization/entity/organization_entity.go:18`, `internal/modules/organization/entity/organization_entity.go:21`, `internal/modules/organization/entity/branch_entity.go:16` |
| 4.7 Behavior config uses typed tables | ✅ done | Typed tenant, branch, branch-service, and counter queue settings exist. | `db/migrations/000032_align_qms_typed_configuration.up.sql:54`, `db/migrations/000033_qms_mvp_alignment.up.sql:3`, `db/migrations/000032_align_qms_typed_configuration.up.sql:108` |

### 16.2 Tenant, Branch, Services, Counters — Sections 7 to 15

| Section | Status | Gap / Finding | Live Evidence |
|---|---|---|---|
| 7 Tenant design | partial | Tenant profile fields and status exist, but activation completeness rule is not enforced. | `internal/modules/organization/entity/organization_entity.go:15`, `internal/modules/organization/entity/organization_entity.go:30`, `internal/modules/organization/usecase/organization_usecase.go:55` |
| 8 `tenant_queue_settings` | done | Table/entity include defaults and tenant unique constraint. | `db/migrations/000032_align_qms_typed_configuration.up.sql:54`, `db/migrations/000032_align_qms_typed_configuration.up.sql:67`, `internal/modules/settings/entity/qms_queue_settings_entity.go:3` |
| 9 Branch design | partial | Branch profile fields exist; branch logo field exists; explicit branch-logo-fallback-to-tenant resolver is missing. | `db/migrations/000032_align_qms_typed_configuration.up.sql:14`, `internal/modules/organization/entity/branch_entity.go:22`, `internal/modules/organization/model/branch_model.go:20` |
| 10 `branch_queue_settings` | done | Nullable override fields and tenant+branch unique key exist. | `db/migrations/000032_align_qms_typed_configuration.up.sql:71`, `db/migrations/000032_align_qms_typed_configuration.up.sql:85`, `internal/modules/settings/entity/qms_queue_settings_entity.go:20` |
| 11 Services | partial | Service type/duration/pharmacy flags exist; audio/narrative fallback rules are not implemented. | `db/migrations/000032_align_qms_typed_configuration.up.sql:25`, `internal/modules/service/entity/service_entity.go:13`, `internal/modules/service/entity/service_entity.go:17` |
| 12 Branch services | done | Branch-service table and CRUD/usecase exist with tenant/branch/service binding. | `db/migrations/000032_align_qms_typed_configuration.up.sql:29`, `internal/modules/service/usecase/branch_service_usecase.go:31`, `internal/modules/service/repository/branch_service_repository.go:1` |
| 13 `branch_service_queue_settings` | done | Design table exists after MVP alignment migration and entity uses branch-service scope. | `db/migrations/000033_qms_mvp_alignment.up.sql:3`, `internal/modules/settings/entity/qms_queue_settings_entity.go:38` |
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
| 29 `qms_clients` | partial | Migration/entity/repository/auth middleware and minimal admin create/write paths exist; full list/update/delete deferred. | `db/migrations/000033_qms_mvp_alignment.up.sql:35`, `internal/modules/qms_client/usecase/qms_client_admin_usecase.go:36`, `internal/middleware/qms_client_middleware.go:20`, `internal/router/router.go:240` |
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
| `branch_service_queue_settings` | done | Migration/entity/resolver scope keyed by tenant, branch, and branch-service exists. | `db/migrations/000033_qms_mvp_alignment.up.sql:3`, `internal/modules/settings/queue_settings_resolver.go:73` |
| `qms_clients` | partial | Client table/domain/auth and minimal admin create endpoints exist; full CRUD deferred. | `internal/modules/qms_client/usecase/qms_client_admin_usecase.go:36` |
| `qms_client_credentials` | missing | Add hashed credential table and auth middleware/resolver. | `documentation/New Design Document — QMS MVP Operatio.md:1889` |
| `operator_counter_assignments` | partial | Assignment table/domain and caller validation exist; admin CRUD deferred. | `db/migrations/000033_qms_mvp_alignment.up.sql:61`, `internal/modules/caller/usecase/caller_usecase.go:225` |
| Caller action endpoint | done | `/api/v1/caller/queue-journeys/{journey_id}/action` implemented with caller-bound context enforcement. | `internal/modules/caller/delivery/http/caller_routes.go:15`, `internal/modules/caller/usecase/caller_usecase.go:64` |
| Signage endpoint | done | `/signage/me`, `/signage/current-calls`, and `/signage/queues` implemented with client binding checks. | `internal/modules/signage/delivery/http/signage_routes.go:10`, `internal/modules/signage/usecase/signage_usecase.go:21` |
| Effective config response | done | Effective config includes typed fields and source metadata for inheritance. | `internal/modules/settings/delivery/http/settings_controller.go:63`, `internal/modules/settings/model/settings_model.go:187` |
| Branch logo fallback | done | Signage `GetMe` falls back to tenant branding when branch branding is null. | `internal/modules/signage/usecase/signage_usecase.go:80` |
| Activation rules tenant/branch | deferred | Explicitly skipped for MVP; no activation completeness guard yet. | `documentation/New Design Document — QMS MVP Operatio.md` |

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
| Phase A: typed tables active, old read-only | ✅ done | Typed tables active, generic write endpoints already removed |
| Phase B: old no longer read by QMS core | ✅ done | `QueueSettingsResolver` reads typed tables only (`internal/modules/settings/queue_settings_resolver.go:32-101`) |
| Phase C: drop old settings | ✅ done | Only `GET /settings/effective` remains; write endpoints already removed from router |

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
| Generic settings only for non-core | ✅ done | Generic write removed; QMS core reads typed tables only |

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
| Section 6 — API Design | 18 | 5 | 7 |
| Section 7 — Setup Wizard | 1 | 1 | 5 |
| Section 8 — Validation Rules | 6 | 1 | 2 |
| Section 9 — Queue Config Usage | 3 | 1 | 0 |
| Section 10 — Audit Log | 12 | 0 | 0 |
| Section 11 — Error Logging | 3 | 0 | 0 |
| Section 12 — Migration Strategy | 6 | 1 | 4 |
| Section 13 — Testing | 16 | 1 | 3 |
| Section 14 — Architecture Decision | 6 | 1 | 0 |
| Section 15 — Table List | 14 | 0 | 0 |

---

## Detailed Gap Register

> **Note**: Gaps below reflect `2026-07-06` runtime. Items resolved since initial 2026-07-02 generation: caller/signage endpoint CRUD, QMS client admin CRUD, operator assignment CRUD, branch CRUD UI, setup wizard shell, table-driven caller/signage/operator_assignment tests, WS event pipeline, frontend contract sync.

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

### Gap D — Tenant/Branch Activation Rules Missing [RESOLVED]

**Design target**

- Tenant and branch cannot become active before required profile completeness rules pass.

**Current runtime**

- Backend guards now enforce profile completeness on update.
- Frontend shared validation added via `@casbin/api-types`.

**Evidence**

- Tenant activation guard: `internal/modules/organization/usecase/organization_usecase.go`
- Branch activation guard: `internal/modules/organization/usecase/branch_usecase.go:103`
- Test: `TestUpdateBranch/Positive_ActivateWithRequiredFieldsInRequest` in `branch_usecase_test.go`.

**Gap impact**

- (Resolved) Incomplete tenants/branches cannot be marked active on backend.

**Needed change**

- Monitor if `apps/web` needs a dedicated branch form component, as it does not currently exist.

### Gap E — Branch Logo Fallback Missing [RESOLVED]

**Design target**

- Branch logo optional; when empty, effective logo should fallback to tenant logo.

**Current runtime**

- Signage `GetMe` now correctly implements tenant logo fallback when branch logo is null.

**Evidence**

- `internal/modules/signage/usecase/signage_usecase.go:80`

**Gap impact**

- (Resolved) Signage feed has logo.

**Needed change**

- None.

### Gap F — Caller Domain Missing [RESOLVED]

**Design target**

- `qms_clients`, `qms_client_credentials`, `operator_counter_assignments`.
- Caller context auto-resolves tenant/branch/branch_service/counter.
- Single action endpoint: `POST /api/v1/caller/queue-journeys/{journey_id}/action`.

**Current runtime**

- Caller domain is fully implemented and tested.
- Admin endpoints for QMS clients and operator assignments added (Jalur C).

**Evidence**

- `internal/modules/caller/delivery/http/caller_routes.go:15`
- `internal/modules/qms_client/delivery/http/qms_client_routes.go:9`
- `internal/modules/operator_assignment/delivery/http/operator_assignment_routes.go:9`

**Gap impact**

- (Resolved) MVP caller flow is fully supported on backend.

**Needed change**

- None.

### Gap G — Signage Domain Missing [RESOLVED]

**Design target**

- Signage authenticates by client credential.
- Signage feed resolves branch/service/counter and returns running_text, effective logo, and current calls.

**Current runtime**

- Signage feed is fully implemented and tested.

**Evidence**

- `internal/modules/signage/delivery/http/signage_routes.go:9`

**Gap impact**

- (Resolved) Signage flow supported.

**Needed change**

- None.

### Gap H — Queue Journey State Machine Partial vs MVP [PARTIAL]

**Design target**

- Typed journey timestamps replace old generic state timestamps.
- Repeated `call` becomes internal recall.
- Skipped journey can be called again.
- Forward/complete/skip/cancel follow strict state machine.

**Current runtime**

- Queue transitions exist.
- Caller single-action endpoint now fronts queue transitions.
- Repeated call/recall semantics and typed journey flow are substantially covered in backend paths, but not yet fully closed by final E2E proof.

**Evidence**

- Queue runtime: `internal/modules/queue/usecase/queue_usecase.go`
- Caller runtime: `internal/modules/caller/usecase/caller_usecase.go`
- Caller tests: `internal/modules/caller/usecase/caller_usecase_test.go`

**Gap impact**

- Backend mostly aligned.
- Remaining risk is proof depth, not missing main runtime path.

**Needed change**

- Add final integration/E2E proof for caller action lifecycle when Docker slice resumes.

### Gap I — Estimate, Audio, Narrative, and Recall Coverage Missing [PARTIAL]

**Design target**

- Effective config and runtime should support estimated duration, audio fallback, narrative fallback, `allow_recall`.

**Current runtime**

- Duration exists in runtime.
- Signage now exposes `audio_id` and `audio_en` on service/current-call payloads.
- `auto_call_next` is present in typed config path.
- Narrative fields and full effective-config narrative/audio projection remain incomplete.

**Evidence**

- `internal/modules/signage/model/signage_model.go`
- `internal/modules/signage/usecase/signage_usecase.go`
- `internal/modules/settings/queue_settings_resolver.go`

**Gap impact**

- Signage is mostly covered.
- Effective config/dashboard contract still not fully covers narrative/audio/fallback story.

**Needed change**

- Finish service audio/narrative migration + resolver exposure if product still needs it in dashboard/effective-config contract.

### Gap J — Migration Strategy Incomplete For MVP [MOSTLY RESOLVED]

**Design target**

- Remove generic settings first.
- Add typed config tables.
- Add tenant/branch profile fields.
- Update services.
- Create branch service.
- Create qms clients and assignments.

**Current runtime**

- Main migration wave already added for typed config alignment and QMS client binding.
- Remaining concern is cleanup consistency and any still-missing service audio/narrative columns.

**Evidence**

- `db/migrations/000032_align_qms_typed_configuration.up.sql`
- `db/migrations/000033_qms_mvp_alignment.up.sql`
- `db/migrations/000034_qms_client_binding.up.sql`

**Gap impact**

- Core MVP schema is reachable now.
- Remaining migration work is additive cleanup, not blocker for core flow.

**Needed change**

- Audit if service audio/narrative columns still need dedicated follow-up migration.

### Gap K — Coverage Matrix Outdated vs MVP [RESOLVED IN PART, DOC NEEDS CONTINUED MAINTENANCE]

**Design target**

- Test matrix now includes caller, signage, client credential security, estimate logic, timestamp behavior, and branch-service scoped config.

**Current runtime**

- Coverage doc has been partially rebased, but drift can reappear after each feature slice.
- Unit/usecase coverage now exists for caller, signage, QMS client credential auth, operator assignment, typed config, and queue transitions.
- Integration/E2E remains deferred.

**Evidence**

- Current coverage file scope: this file
- Current test playbooks still mention old settings inheritance/service scope patterns.

**Gap impact**

- Team can incorrectly assume Docker-backed integration/E2E is done when only narrow backend/frontend checks have passed.

**Needed change**

- Keep this audit updated after each slice.
- Resume Docker-backed integration/E2E when environment scope allows it.

---

## 17. What Needs Work Next

Priority order from current runtime (2026-07-06):

1. **Typed write API for config**
   - Expose first-class create/update endpoints for typed queue settings, or explicitly document that only read/effective config is supported for MVP.

2. **Remove legacy JSON config from core entities**
   - Drop or quarantine `services.settings` and `counters.settings` so core queue behavior cannot drift back to generic config.

3. **Finish UI-facing effective config contract**
   - If dashboard/caller/signage need one payload, expand `/settings/effective` into a richer contract with tenant/branch/service/counter presentation data.

4. **Rebase cached overview docs**
   - Update `llm/cache/project-overview.md` so it no longer says queue modules are missing.

5. **Close proof gaps**
   - Run narrow integration/E2E proof for caller, signage, and queue journey lifecycle when Docker-backed validation is available.
