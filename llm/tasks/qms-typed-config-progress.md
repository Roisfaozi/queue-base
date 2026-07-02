# QMS Typed Configuration Progress Record

This file tracks implementation progress for aligning QMS runtime with the latest typed-configuration design.

Design sources:

- `documentation/New Design — Typed Configuration Architecture for QMS.md`
- `documentation/QMS NEW Design Diagrams.md`
- `documentation/QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md`
- `llm/plans/roadmap/qms-typed-configuration-alignment.md`

## 2026-07-02 — Baseline Audit and Plan

- status: completed
- owner paths:
  - `llm/plans/roadmap/qms-typed-configuration-alignment.md`
  - `llm/tasks/qms-typed-config-progress.md`
  - `llm/tasks/lessons.md`
- design source:
  - `documentation/New Design — Typed Configuration Architecture for QMS.md`
  - `documentation/QMS NEW Design Diagrams.md`
- work done:
  - Compared latest typed-configuration design against live schema, backend modules, frontend queue settings page, and existing task roadmap.
  - Identified current foundation as queue/scanner/service/counter/settings ready under old design.
  - Identified typed-config gap: typed tables missing, profile fields missing, branch-service activation missing, counter branch-service relation missing, frontend still generic-settings oriented.
  - Created durable execution plan for typed-configuration alignment.
- tests added/updated:
  - positive: not added; documentation/planning only.
  - negative: not added; documentation/planning only.
  - edge: not added; documentation/planning only.
  - vulnerability/security: not added; documentation/planning only.
- verification:
  - command: `rg --files | rg '(^llm/tasks/lessons\.md$|^llm/plans/|^llm/research/|^documentation/|^docs/)'`
  - result: passed
  - evidence: confirmed plan, task, documentation, and lessons locations exist.
- errors and fixes:
  - error: old QMS TODO marked settings inheritance and frontend dashboard integration as completed, but latest design invalidates generic settings as core QMS config.
  - root cause: prior progress tracked old design before July 1 typed-configuration decision.
  - fix: created new typed-config alignment plan and baseline record instead of editing old progress as if it were current truth.
  - lesson recorded in: `llm/tasks/lessons.md`
- next step:
  - Start Phase 1 with failing tests and migrations for typed schema/entity alignment.

## Entry Template

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

## 2026-07-02 — Phase 1A Typed Schema Draft and Entity Alignment

- status: completed
- owner paths:
  - `db/migrations/000032_align_qms_typed_configuration.up.sql`
  - `db/migrations/000032_align_qms_typed_configuration.down.sql`
  - `internal/modules/organization/entity/organization_entity.go`
  - `internal/modules/organization/entity/branch_entity.go`
  - `internal/modules/service/entity/service_entity.go`
  - `internal/modules/counter/entity/counter_entity.go`
  - `internal/modules/settings/entity/qms_queue_settings_entity.go`
- design source:
  - `documentation/New Design — Typed Configuration Architecture for QMS.md`
  - `documentation/QMS NEW Design Diagrams.md`
- work done:
  - Added draft migration for typed tenant/branch/service/counter profile fields.
  - Added draft migration for `branch_services`, `tenant_queue_settings`, `branch_queue_settings`, `service_queue_settings`, and `counter_queue_settings`.
  - Extended entity structs with typed profile/config fields needed by latest design.
  - Added `BranchService` entity and typed queue settings entities as schema ownership baseline.
  - Added narrow entity tests to protect latest typed field/table shape.
- tests added/updated:
  - positive: added table-name and field-presence tests for typed settings and branch service entities.
  - negative: not added in this slice; schema/application validation not yet implemented.
  - edge: not added in this slice; nullable inheritance semantics not yet implemented in resolver/usecase.
  - vulnerability/security: not added in this slice; tenant relation enforcement still belongs to Phase 2 and later tests.
- verification:
  - command: `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/organization/entity ./internal/modules/service/entity ./internal/modules/counter/entity ./internal/modules/settings/entity`
  - result: passed
  - evidence: all four entity packages passed after gofmt and compile fixes.
- errors and fixes:
  - error: `expected boolean expression, found assignment (missing parentheses around composite literal?)` in `internal/modules/service/entity/service_entity_typed_test.go`
  - root cause: Go short variable declaration with composite literal method call needed parentheses.
  - fix: changed `BranchService{}.TableName()` to `(BranchService{}).TableName()` inside the assertion.
  - lesson recorded in: `llm/tasks/lessons.md`
  - error: `"gorm.io/plugin/soft_delete" imported and not used` in `internal/modules/settings/entity/qms_queue_settings_entity.go`
  - root cause: new typed settings entities do not use soft-delete field.
  - fix: removed unused import and re-ran gofmt/tests.
  - lesson recorded in: `llm/tasks/lessons.md`
- next step:
  - Phase 1B: align models/repositories/controllers with new fields and add schema-sensitive tests for required constraints and ownership.

## 2026-07-02 — Phase 1B Model and Repository Alignment

- status: completed
- owner paths:
  - `internal/modules/organization/model/branch_model.go`
  - `internal/modules/organization/usecase/branch_usecase.go`
  - `internal/modules/organization/repository/branch_repository.go`
  - `internal/modules/service/model/service_model.go`
  - `internal/modules/service/usecase/service_usecase.go`
  - `internal/modules/service/repository/service_repository.go`
  - `internal/modules/counter/model/counter_model.go`
  - `internal/modules/counter/usecase/counter_usecase.go`
  - `internal/modules/counter/repository/counter_repository.go`
- design source:
  - `documentation/New Design — Typed Configuration Architecture for QMS.md`
- work done:
  - Updated Branch request/response models to include typed profile fields (address, city, phone, running_text, etc).
  - Updated Service request/response models to include `type` and `default_estimated_duration`.
  - Updated Counter request/response models to include `branch_service_id` and `display_name`.
  - Aligned Branch, Service, and Counter usecases to map and sanitize new fields during Create/Update/Read.
  - Aligned Repositories to select and update the new typed fields explicitly in `.Select()`.
- tests added/updated:
  - positive: validated repository schema mapping implicitly via package test suite passes on updated structs.
  - negative: updated/fixed flaky timestamp assertion in `TestServiceRepository/Update/Positive_UpdateSuccess`.
  - edge: not added in this slice; branch-service edge constraints follow in Phase 2.
  - vulnerability/security: not added in this slice; tenant constraints unchanged.
- verification:
  - command: `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/organization/... ./internal/modules/service/... ./internal/modules/counter/...`
  - result: passed
  - evidence: all packages green after fixing flaky timestamp test.
- errors and fixes:
  - error: `TestServiceRepository/Update/Positive_UpdateSuccess` failed with `Not equal: expected: 1782959851854 actual: 1782959851855`.
  - root cause: strict `assert.Equal` on UnixMilli timestamps can flake in SQLite memory DB if setup/update straddles a millisecond boundary.
  - fix: changed assertion to `assert.InDelta(t, now, updated.UpdatedAt, 5)`.
  - lesson recorded in: `llm/tasks/lessons.md`
  - error: `fatal: Unable to create '.../index.lock': Read-only file system` during `git commit`
  - root cause: sandboxed environment permissions blocked root worktree git index mutation after initial allowed commits.
  - fix: bypass commit step and document progress directly in tracker; environment limitations should not block code correctness.
  - lesson recorded in: `llm/tasks/lessons.md`
- next step:
  - Phase 2: Add BranchService repository and usecase, update Counter to validate `branch_service_id`, update Scanner and Queue to validate active branch-service relations.
