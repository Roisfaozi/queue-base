## Audit: TDT Migration Remaining

**Generated:** 2026-06-30  
**Branch:** `test/table-testing`

### State: Completed ✅

- `internal/modules/auth/test/*` (13 files) — all TDT with `name`, `category`, `run`
- `internal/modules/user/test/*` (6 files) — all TDT
- `internal/modules/auth/repository/*_test.go` (2 files) — all TDT
- `internal/modules/user/repository/*_test.go` (2 files) — all TDT

### State: Needs Fix (missing `category`) — None. All TDT-compliant. ✅

### State: Not Yet Migrated (needs full TDT wrap)

#### Unit Tests (Completed — Phase 2)

- `internal/modules/organization/test/organization_usecase_test.go` — ✅ TDT unified (25+ test cases, all with `category`)
- `internal/modules/organization/test/organization_member_usecase_test.go` — ✅ TDT unified

#### Integration Tests

| File                                                                    | Test Count (est.) | Priority |
| ----------------------------------------------------------------------- | ----------------- | -------- |
| `tests/integration/modules/api_key_integration_test.go`                 | 3                 | DONE     |
| `tests/integration/modules/audit_integration_test.go`                   | 2                 | Medium   |
| `tests/integration/modules/data_isolation_integration_test.go`          | 5+                | DONE     |
| `tests/integration/modules/organization_integration_test.go`            | 4                 | DONE     |
| `tests/integration/modules/project_integration_test.go`                 | 2                 | Medium   |
| `tests/integration/modules/stats_integration_test.go`                   | 3                 | Low      |
| `tests/integration/modules/tus_integration_test.go`                     | 3                 | Low      |
| `tests/integration/modules/webhook_integration_test.go`                 | 2                 | Medium   |
| `tests/integration/modules/worker_integration_test.go`                  | 2                 | Medium   |
| `tests/integration/scenarios/admin_security_test.go`                    | 2                 | High     |
| `tests/integration/scenarios/adv_rate_limit_test.go`                    | 2                 | Low      |
| `tests/integration/scenarios/api_key_lifecycle_test.go`                 | 3                 | Medium   |
| `tests/integration/scenarios/concurrent_session_test.go`                | 2                 | Medium   |
| `tests/integration/scenarios/delete_user_integrity_test.go`             | 2                 | Medium   |
| `tests/integration/scenarios/exception_handling_test.go`                | 3                 | Medium   |
| `tests/integration/scenarios/password_recovery_test.go`                 | 1                 | Low      |
| `tests/integration/scenarios/rate_limit_integration_test.go`            | 2                 | Low      |
| `tests/integration/scenarios/rbac_orchestration_test.go`                | 2                 | Medium   |
| `tests/integration/scenarios/realtime_test.go`                          | 1                 | Low      |
| `tests/integration/scenarios/role_hierarchy_test.go`                    | 1                 | Low      |
| `tests/integration/scenarios/transaction_integrity_test.go`             | 2                 | Medium   |
| `tests/integration/scenarios/transactional_enforcer_regression_test.go` | 2                 | Medium   |
| `tests/integration/scenarios/user_lifecycle_test.go`                    | 3                 | DONE     |
| `tests/integration/scenarios/worker_integration_test.go`                | 2                 | Low      |

#### E2E Tests

| File                                           | Test Count (est.) | Priority |
| ---------------------------------------------- | ----------------- | -------- |
| `tests/e2e/api/auth_e2e_test.go`               | 4                 | DONE     |
| `tests/e2e/api/email_verification_e2e_test.go` | 2                 | Medium   |
| `tests/e2e/api/export_audit_e2e_test.go`       | 1                 | Low      |
| `tests/e2e/api/multi_tenancy_e2e_test.go`      | 2                 | DONE     |
| `tests/e2e/api/organization_e2e_test.go`       | 2                 | Medium   |
| `tests/e2e/api/project_e2e_test.go`            | 2                 | Medium   |
| `tests/e2e/api/qms_queue_e2e_test.go`          | 3                 | DONE     |
| `tests/e2e/api/stats_e2e_test.go`              | 2                 | Low      |
| `tests/e2e/api/tenant_isolation_e2e_test.go`   | 2                 | DONE     |
| `tests/e2e/api/user_e2e_test.go`               | 4                 | DONE     |
| `tests/e2e/modules/api_key_e2e_test.go`        | 2                 | Medium   |
| `tests/e2e/modules/qms_scanner_e2e_test.go`    | 2                 | Medium   |
| `tests/e2e/modules/tus_e2e_test.go`            | 2                 | Low      |
| `tests/e2e/modules/webhook_e2e_test.go`        | 2                 | Medium   |
| `tests/e2e/realtime/sse_e2e_test.go`           | 1                 | Low      |

#### Unit Tests (Modules)

| File                                                                     | Test Count (est.) | Priority |
| ------------------------------------------------------------------------ | ----------------- | -------- |
| `internal/modules/api_key/test/api_key_controller_test.go`               | 4                 | DONE     |
| `internal/modules/api_key/test/api_key_usecase_test.go`                  | 3                 | DONE     |
| `internal/modules/audit/test/audit_controller_test.go`                   | 2                 | DONE     |
| `internal/modules/audit/test/audit_repository_test.go`                   | 2                 | Low      |
| `internal/modules/audit/test/audit_usecase_test.go`                      | 3                 | DONE     |
| `internal/modules/auth/test/setup_test.go`                               | 0 (helper)        | —        |
| `internal/modules/organization/test/organization_controller_test.go`     | 4                 | High     |
| `internal/modules/organization/test/organization_member_usecase_test.go` | 3                 | DONE     |
| `internal/modules/organization/test/organization_usecase_test.go`        | 5                 | DONE     |
| `internal/modules/organization/test/reader_test.go`                      | 1                 | Low      |
| `internal/modules/project/test/project_usecase_test.go`                  | 3                 | Medium   |
| `internal/modules/role/usecase/role_usecase_test.go`                     | 3                 | DONE     |
| `internal/modules/permission/usecase/permission_usecase_test.go`         | 3                 | High     |
| `internal/modules/stats/test/stats_usecase_test.go`                      | 2                 | Low      |
| `internal/modules/webhook/test/webhook_controller_test.go`               | 3                 | Medium   |
| `internal/modules/webhook/test/webhook_repository_test.go`               | 2                 | Low      |
| `internal/modules/webhook/test/webhook_usecase_test.go`                  | 3                 | Medium   |

### Implementation Plan

1. **Phase 1 — Fix TDT with missing `category` (DONE):**
   - Remaining Files: qms_queue_integration, qms_scanner_integration (Already compliant)
2. **Phase 2 — Priority (high risk / core functionality):**
   - integration: data_isolation, organization, api_key, user_lifecycle (DONE)
   - e2e: auth_e2e (DONE), user_e2e (DONE), multi_tenancy (DONE), tenant_isolation (DONE), qms_queue (DONE)
   - unit: organization (controller, usecase, member)

3. **Phase 3 — Remaining medium priority**

4. **Phase 4 — Low priority (specialized / optional services)**
