# QMS Typed Configuration Alignment Plan

## Purpose

This plan tracks the rebuild path from the current QMS foundation into the latest typed-configuration design.

Primary design sources:

- `documentation/New Design — Typed Configuration Architecture for QMS.md`
- `documentation/QMS NEW Design Diagrams.md`
- `documentation/QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md`
- `documentation/task-overview.md`
- `llm/research/queue-management-rebuild-brief.md`
- `llm/research/queue-management-domain-map.md`
- `llm/research/queue-management-implementation-design.md`
- `llm/plans/roadmap/queue-management-rebuild.md`
- `llm/plans/roadmap/queue-management-feature-map.md`

## Current Verified Progress

### Completed Foundation

| Area | Status | Evidence | Notes |
| --- | --- | --- | --- |
| Queue master and journey schema | Completed for old normalized design | `db/migrations/000027_create_queues_and_journeys.up.sql` | Has `queues`, `queue_counters`, `queue_journeys`, `visit_journeys`. |
| Queue module | Completed foundation | `internal/modules/queue/*` | Register, list, forward, transition, journey reads, stats exist. |
| Queue transaction and row lock | Completed foundation | `internal/modules/queue/repository/queue_repository.go` | Uses row lock around `queue_counters`. |
| Scanner module | Completed foundation | `internal/modules/scanner/*` | Scanner orchestration and relation validation exist. |
| Branch module | Completed old foundation | `internal/modules/organization/*branch*` | Tenant-scoped branch CRUD exists, but latest profile fields missing. |
| Service module | Completed old foundation | `internal/modules/service/*` | Service CRUD exists, pharmacy flags exist, but typed design fields missing. |
| Counter module | Completed old foundation | `internal/modules/counter/*` | Counter CRUD exists, but `branch_service_id` and display fields missing. |
| Generic settings module | Completed old foundation | `internal/modules/settings/*` | Must not remain core QMS config under latest design. |
| Web dashboard QMS pages | Partial | `apps/web/src/app/[locale]/dashboard/*` | Existing pages target old APIs and generic settings behavior. |
| QMS tests | Partial | `internal/modules/*/*_test.go`, `tests/integration`, `tests/e2e` | Coverage exists, but latest typed-config behavior not covered. |

### Not Yet Aligned With Latest Design

| Area | Gap | Required Change |
| --- | --- | --- |
| Tenant profile | `organizations` lacks tenant profile fields from design | Add/align tenant profile fields or introduce explicit tenant model while preserving org tenant boundary. |
| Branch profile | `branches` lacks address, city, province, phone, logo, running text, timezone | Add typed profile fields and activation validation. |
| Service profile | `services` lacks `type` and `default_estimated_duration`; still has JSON `settings` | Add typed fields and move behavior config to typed settings tables. |
| Branch-service activation | Missing `branch_services` table and module behavior | Add table/repository/usecase and require active relation for queue creation/forwarding. |
| Counter profile | `counters` lacks `branch_service_id` and `display_name`; still has JSON `settings` | Link counter to `branch_services`; add display name. |
| Core QMS settings | Generic `settings` still powers `queue_reset_time` | Add typed settings tables and typed effective resolver. |
| Frontend queue settings | UI calls missing/old generic endpoint | Redesign around typed profile/settings/effective values. |

## Scope

### In Scope

- Database migrations for latest typed QMS design.
- Backend entities, models, repositories, usecases, routes, and API contracts for typed settings.
- Queue/scanner validation changes that depend on branch-service and counter typed relations.
- Web dashboard sync for profile, overrides, and effective configuration.
- Tests following QMS TDD rule: positive, negative, edge, vulnerability/security.
- Progress, error, and fix recording after every implementation slice.

### Non-Goals

- Removing generic `settings` completely.
- Adding complex signage theme/display tables before MVP requires them.
- Medifrans, Bithealth, or external integration work unless explicitly reintroduced.
- Frontend-only guards as security substitute.
- Broad unrelated refactors outside QMS alignment.

## Execution Rules

1. Start each slice by adding or updating progress record in `llm/tasks/qms-typed-config-progress.md`.
2. Start implementation with failing tests when feasible.
3. Use live code as truth; docs define target design only.
4. Keep tenant and branch ownership checks first-class in repositories and usecases.
5. Use typed tables for core QMS behavior; generic `settings` only for non-critical future dynamic config.
6. Record every encountered error, root cause, and fix in `llm/tasks/lessons.md` under QMS typed-config section.
7. Do not mark a slice complete without verification command and result.

## Phase Plan

### Phase 0 — Baseline Record and Contract Freeze

Owner paths:

- `llm/tasks/qms-typed-config-progress.md`
- `llm/tasks/lessons.md`
- `documentation/New Design — Typed Configuration Architecture for QMS.md`
- `documentation/QMS NEW Design Diagrams.md`

Tasks:

- Record current baseline status from live code.
- Freeze target naming and ownership for tenant, branch, service, branch service, counter, settings.
- Decide whether existing `organizations` is renamed conceptually as tenant or wrapped by tenant-facing API.

Verification:

- Documentation review only.
- Confirm file paths exist with `rg --files`.

Exit criteria:

- Progress record has baseline entry.
- No implementation begins before baseline is recorded.

### Phase 1 — Typed Schema and Entity Alignment

Owner paths:

- `db/migrations/*`
- `internal/modules/organization/entity/*`
- `internal/modules/service/entity/*`
- `internal/modules/counter/entity/*`
- new typed settings entities/repositories if created under `internal/modules/settings/*` or dedicated QMS config module

Tasks:

- Add tenant profile fields required by design.
- Add branch profile fields: address, city, province, postal code, phone, email, logo asset, running text, timezone.
- Add service fields: type, default estimated duration.
- Add `branch_services` table.
- Add counter fields: branch_service_id, display_name.
- Add typed settings tables: `tenant_queue_settings`, `branch_queue_settings`, `service_queue_settings`, `counter_queue_settings`.
- Keep down migrations paired with up migrations.

Tests:

- Positive: migration creates tables and columns.
- Negative: required FK/unique constraints reject invalid duplicate/ownership data.
- Edge: nullable override fields allow inheritance.
- Vulnerability: cross-tenant branch/service/counter relation is rejected by composite constraints or repository validation.

Verification:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/organization ./internal/modules/service ./internal/modules/counter ./internal/modules/settings`
- Migration test/harness if available.

Exit criteria:

- Schema can support latest ERD.
- Entity structs match new columns.
- Existing tests updated without weakening tenant scope.

### Phase 2 — Branch-Service Activation Domain

Owner paths:

- new or existing `internal/modules/service/*`
- `internal/modules/counter/*`
- `internal/modules/scanner/usecase/relation_validator.go`
- `internal/modules/queue/usecase/*`

Tasks:

- Add branch-service repository/usecase/model/controller contracts.
- Enforce `unique tenant_id + branch_id + service_id`.
- Make counters point to `branch_service_id`.
- Update scanner relation validator to validate tenant -> branch -> branch_service -> counter.
- Update queue register and forward to reject inactive/missing branch service.

Tests:

- Positive: active branch service can register/forward queue.
- Negative: inactive branch service fails.
- Edge: custom_name nullable still resolves to service name.
- Vulnerability: service from another tenant or branch fails closed.

Verification:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/service ./internal/modules/counter ./internal/modules/scanner ./internal/modules/queue`

Exit criteria:

- Queue creation and forwarding no longer accept raw service IDs without branch activation validation.

### Phase 3 — Typed Effective Configuration Resolver

Owner paths:

- `internal/modules/settings/*` or new QMS config module
- `internal/modules/queue/usecase/queue_usecase.go`
- `internal/modules/queue/model/*`
- `internal/modules/settings/delivery/http/*`
- `documentation/api/qms/SETTINGS_API.md`

Tasks:

- Implement typed resolver for tenant -> branch -> service -> counter.
- Expose effective config API returning source metadata per field.
- Replace queue reset and ticket prefix lookup from generic settings with typed resolver.
- Keep generic settings endpoints but remove them from core QMS consumers.

Tests:

- Positive: effective config resolves tenant default.
- Negative: missing tenant default returns clear error.
- Edge: branch null override inherits tenant value, service override wins when set.
- Vulnerability: cross-tenant scope IDs cannot influence effective config.

Verification:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/settings ./internal/modules/queue`

Exit criteria:

- No core queue path reads generic `settings` for reset time, ticket prefix, or operational behavior.

### Phase 4 — Queue Domain Hardening Against New Model

Owner paths:

- `internal/modules/queue/*`
- `internal/modules/scanner/*`
- `tests/integration/modules/qms_queue_integration_test.go`
- `tests/e2e/api/qms_queue_e2e_test.go`

Tasks:

- Ensure forward appends `queue_journeys` only and keeps same queue master row.
- Ensure queue number and ticket number remain stable across journeys.
- Ensure status transitions follow typed setting flags: allow forward, skip, recall, cancel.
- Ensure pharmacy flow uses service flags, not string matching.

Tests:

- Positive: register then forward creates one queue with multiple journeys.
- Negative: disabled allow_forward rejects forward.
- Edge: business date honors typed queue_reset_time.
- Vulnerability: cross-tenant forward/service/counter relation fails.

Verification:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/queue ./internal/modules/scanner`
- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./tests/integration/modules -run TestQMSQueueIntegration`
- E2E only if Docker/test harness available.

Exit criteria:

- Latest queue business rules are enforced by backend, not UI assumptions.

### Phase 5 — Frontend Contract and UI Alignment

Owner paths:

- `apps/web/src/lib/api/qms.ts`
- `apps/web/src/app/[locale]/dashboard/queue-settings/*`
- `apps/web/src/app/[locale]/dashboard/services/*`
- `apps/web/src/app/[locale]/dashboard/counters/*`
- `packages/*` if shared API types exist

Tasks:

- Replace generic settings UI with typed settings UI.
- Show profile data and queue behavior config separately.
- Show inherited/effective values and override source per field.
- Update service/counter pages for branch service and counter branch_service_id.
- Ensure dashboard uses available backend endpoints only.

Tests:

- Positive: settings page renders effective config data.
- Negative: missing branch/service selection shows actionable empty state.
- Edge: inherited values display source labels correctly.
- Vulnerability: UI does not offer cross-tenant IDs and backend still rejects forged IDs.

Verification:

- `pnpm --filter web typecheck`
- `pnpm --filter web lint`
- `pnpm --filter web build` if route changes are broad.

Exit criteria:

- UI no longer calls `GET /settings` generic list for QMS core settings.

### Phase 6 — Documentation, API, and Handoff Sync

Owner paths:

- `documentation/api/qms/*`
- `documentation/QMS_FEATURE_AND_E2E_GUIDE.md`
- `documentation/guides/QMS_MANUAL_TEST_FLOW.md`
- `llm/tasks/qms-typed-config-progress.md`
- `llm/tasks/lessons.md`

Tasks:

- Update API docs for typed config and branch-service contracts.
- Update manual test flow.
- Update progress record with final status.
- Add lessons from actual errors/fixes.

Verification:

- Docs path and endpoint references checked against route files.
- Manual test flow references only existing endpoints.

Exit criteria:

- Reviewer can see what changed, what passed, what failed/blocked, and what remains.

## Progress Record Format

Every implementation session must append an entry to `llm/tasks/qms-typed-config-progress.md`:

```md
## YYYY-MM-DD — <slice name>

- status: planned | in_progress | completed | blocked
- owner paths:
  - `path/to/file`
- design source:
  - `documentation/...`
- work done:
  - ...
- tests added/updated:
  - positive: ...
  - negative: ...
  - edge: ...
  - vulnerability/security: ...
- verification:
  - command: `...`
  - result: passed | failed | blocked
  - evidence: ...
- errors and fixes:
  - error: ...
  - root cause: ...
  - fix: ...
  - lesson recorded in: `llm/tasks/lessons.md`
- next step:
  - ...
```

## Error and Fix Recording Rule

When any error occurs:

1. Record the exact error text or symptom.
2. Record the root cause after investigation.
3. Record the fix applied.
4. Record prevention rule in `llm/tasks/lessons.md`.
5. Link the lesson from the progress entry.

Do not hide fixed errors. Fixed errors are useful project memory.

## Verification Summary Matrix

| Phase | Narrow Verification | Broader Verification | Required Before Complete |
| --- | --- | --- | --- |
| Phase 1 | Module tests for affected entities/repos | Migration harness if available | Yes |
| Phase 2 | Service/counter/scanner/queue package tests | Integration route tests | Yes |
| Phase 3 | Settings + queue package tests | API contract tests | Yes |
| Phase 4 | Queue/scanner package tests | QMS integration and E2E | Yes |
| Phase 5 | Web typecheck/lint | Web build | Yes for typecheck/lint |
| Phase 6 | Docs route/path review | Manual smoke flow | Yes |

## Approval Gates

- Migration shape must be reviewed before broad implementation.
- Any destructive data migration or table removal needs explicit user approval.
- Generic `settings` may stay for non-core use; do not delete without separate plan.
- Cross-stack API changes require backend and frontend sync in same phase or explicit temporary compatibility adapter.
