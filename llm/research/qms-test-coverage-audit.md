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

- **Usecase:** ✅ Table-driven (`counter_usecase_test.go` — full category coverage)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing

### Service (`internal/modules/service`)

- **Usecase:** ✅ Table-driven (`service_usecase_test.go` — full category coverage)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing

### Organization / Branch (`internal/modules/organization`)

- **Branch Usecase:** ✅ Table-driven (`branch_usecase_test.go` — full category coverage)
- **Branch Repository:** ❌ Missing
- **Branch Handler:** ❌ Missing

---

## Priority Matrix

| Priority | Package      | Gap      | Reason                                                               |
| -------- | ------------ | -------- | -------------------------------------------------------------------- |
| **P3**   | Repositories | No tests | Queue, Scanner, Settings repo queries warrant unit-style repo tests. |
