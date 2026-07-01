# QMS API Reference Index

Dokumen API QMS dipecah per domain agar lebih mudah dibaca dan dirawat.

## Read First

- `documentation/QMS_FEATURE_AND_E2E_GUIDE.md` — logic flow, lifecycle, boundary, dan skenario E2E fitur QMS
- `documentation/api/qms/COMMON.md` — base URL, auth, response envelope, enum umum, dan error notes

## Per-Domain API Reference

- `documentation/api/qms/BRANCH_API.md` — CRUD branch
- `documentation/api/qms/SERVICE_API.md` — CRUD service
- `documentation/api/qms/COUNTER_API.md` — CRUD counter
- `documentation/api/qms/SETTINGS_API.md` — create, resolve, get, update, delete setting
- `documentation/api/qms/QUEUE_API.md` — register, list, detail, transition, forward, stats, active journeys, visit history
- `documentation/api/qms/SCANNER_API.md` — scanner check-in register/forward
- `documentation/api/qms/OPERATIONS_AND_SECURITY.md` — Phase 9 backend operations, audit, tenant/branch isolation, and scanner API-key contract
- `documentation/api/qms/QMS_ALERT_POLICY.md` — Phase 11 observability alert policy for queue/scanner metrics and triage

## Suggested Reading Order

1. `documentation/api/qms/COMMON.md`
2. domain master data yang relevan (`BRANCH_API.md`, `SERVICE_API.md`, `COUNTER_API.md`)
3. `documentation/api/qms/SETTINGS_API.md`
4. `documentation/api/qms/QUEUE_API.md`
5. `documentation/api/qms/SCANNER_API.md`
6. `documentation/api/qms/OPERATIONS_AND_SECURITY.md`
7. `documentation/api/qms/QMS_ALERT_POLICY.md`
