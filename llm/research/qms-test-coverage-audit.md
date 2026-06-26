# QMS Test Coverage Audit

Based on the QMS architecture domain rules (Queue Master, Scanner, Settings Inheritance, Counter, Service, Organization), the following is the current test coverage map and gap analysis.

## Module-by-Module Coverage

### Queue (`internal/modules/queue`)
- **Usecase:** ✅ Table-driven (`queue_usecase_test.go` — 9 methods, full category coverage)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing
- **Integration:** ✅ Table-driven (`qms_queue_integration_test.go` — 4 phases)

### Scanner (`internal/modules/scanner`)
- **Usecase:** ✅ Table-driven (`scanner_usecase_test.go` — `TestScannerCheckIn` with 14 cases)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing
- **Integration:** ❌ Missing
- **E2E:** ❌ Missing (No test for API-key-scoped check-in flow)

### Settings (`internal/modules/settings`)
- **Usecase:** ✅ Table-driven (`settings_usecase_test.go` — 16 inheritance-chain cases)
- **Repository:** ❌ Missing
- **Handler:** ❌ Missing

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

| Priority | Package | Gap | Reason |
|---|---|---|---|
| **P0** | Scanner Handler | No HTTP handler tests | Scanner is the external API surface for ticket check-in. |
| **P0** | Scanner Integration | No integration test | Scanner check-in hits real DB, Casbin, API keys — high risk. |
| **P1** | Scanner E2E | No E2E test | API-key-scoped check-in flow not exercised end-to-end. |
| **P1** | Queue Handler | No HTTP handler tests | Queue register/forward/transition/cancel endpoints are user-facing. |
| **P2** | Counter Usecase | Old-style tests | Needs table-driven rewrite + category coverage. |
| **P2** | Service Usecase | Old-style tests | Needs table-driven rewrite + category coverage. |
| **P2** | Branch Usecase | Old-style tests | Needs table-driven rewrite + category coverage. |
| **P3** | Repositories | No tests | Queue, Scanner, Settings repo queries warrant unit-style repo tests. |
