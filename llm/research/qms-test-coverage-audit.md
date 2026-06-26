# QMS Test Coverage Audit

Based on the QMS architecture domain rules (Queue Master, Scanner, Settings Inheritance, Counter, Service, Organization), the following is the current test coverage map and gap analysis.

## Module-by-Module Coverage

### Queue (`internal/modules/queue`)

- **Usecase:** ✅ Table-driven (`queue_usecase_test.go` — 9 methods, full category coverage)
- **Repository:** ✅ Exists (`queue_repository_test.go`)
- **Handler:** ✅ Exists (`queue_controller_test.go`)
- **Integration:** ✅ Table-driven (`qms_queue_integration_test.go` — 4 phases)

### Scanner (`internal/modules/scanner`)

- **Usecase:** ✅ Table-driven (`scanner_usecase_test.go` — `TestScannerCheckIn` with 14 cases)
- **Repository:** ❌ Missing (No repository folder)
- **Handler:** ✅ Exists (`scanner_controller_test.go`)
- **Integration:** ✅ Exists (`qms_scanner_integration_test.go` — 5 table-driven cases, positive/negative/vulnerability coverage)
- **E2E:** ✅ Exists (`qms_scanner_e2e_test.go` — API-key-scoped check-in flow, 4 subtests: positive, negative, invalid-key, forward)

### Settings (`internal/modules/settings`)

- **Usecase:** ✅ Table-driven (`settings_usecase_test.go` — 16 inheritance-chain cases)
- **Repository:** ❌ Missing
- **Handler:** ✅ Exists (`settings_controller_test.go`)

### Counter (`internal/modules/counter`)

- **Usecase:** ⚠️ Legacy format (no table, no vuln/edge coverage)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing

### Service (`internal/modules/service`)

- **Usecase:** ⚠️ Legacy format (6 single-case functions, no vuln/edge coverage)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing

### Organization / Branch (`internal/modules/organization`)

- **Branch Usecase:** ⚠️ Legacy format
- **Branch Repository:** ❌ Missing
- **Branch Handler:** ❌ Missing

---

## Priority Matrix

| Priority | Package         | Gap             | Reason                                                               |
| -------- | --------------- | --------------- | -------------------------------------------------------------------- |
| **P2**   | Counter Usecase | Old-style tests | Needs table-driven rewrite + category coverage.                      |
| **P2**   | Service Usecase | Old-style tests | Needs table-driven rewrite + category coverage.                      |
| **P2**   | Branch Usecase  | Old-style tests | Needs table-driven rewrite + category coverage.                      |
| **P3**   | Repositories    | No tests        | Queue, Scanner, Settings repo queries warrant unit-style repo tests. |
