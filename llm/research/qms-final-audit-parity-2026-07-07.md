# QMS Final Audit Parity — 2026-07-07

Source of truth checked against live code:
- `documentation/New Design Document — QMS MVP Operatio.md`
- `internal/router/router.go`
- `internal/modules/*/delivery/http/*_routes.go`
- `internal/modules/*/usecase/*.go`
- `internal/modules/settings/delivery/http/settings_controller.go`
- `internal/modules/settings/delivery/http/settings_controller_test.go`
- `tests/integration/**`
- `tests/e2e/**`

## 1. Route Alias vs Target Design

### 1.1 Exact parity achieved

| Design route | Live status | Evidence |
|---|---|---|
| `GET /api/v1/tenant/profile` | done | `internal/modules/organization/delivery/http/organization_routes.go:30` |
| `PATCH /api/v1/tenant/profile` | done | `internal/modules/organization/delivery/http/organization_routes.go:31` |
| `GET /api/v1/branches/{branch_id}/profile` | done | `internal/modules/organization/delivery/http/branch_routes.go:13` |
| `PATCH /api/v1/branches/{branch_id}/profile` | done | `internal/modules/organization/delivery/http/branch_routes.go:15` |
| `POST/GET/PATCH/DELETE /api/v1/branches/{branch_id}/counters` | done | `internal/modules/counter/delivery/http/counter_routes.go:18-25` |
| `GET /api/v1/branches/{branch_id}/effective-config` | done | `internal/modules/settings/delivery/http/settings_routes.go:16` |
| `GET /api/v1/branches/{branch_id}/services/{service_id}/effective-config` | done | `internal/modules/settings/delivery/http/settings_routes.go:18` |
| `GET /api/v1/branches/{branch_id}/counters/{counter_id}/effective-config` | done | `internal/modules/settings/delivery/http/settings_routes.go:20` |
| `DELETE /api/v1/branches/{branch_id}/queue-settings/{field}` | done | `internal/modules/settings/delivery/http/settings_routes.go:17` |
| `DELETE /api/v1/branches/{branch_id}/services/{branch_service_id}/queue-settings/{field}` | done | `internal/modules/settings/delivery/http/settings_routes.go:19` |
| `DELETE /api/v1/branches/{branch_id}/counters/{counter_id}/queue-settings/{field}` | done | `internal/modules/settings/delivery/http/settings_routes.go:21` |
| `POST/GET/PATCH /api/v1/qms-clients` | done | `internal/modules/qms_client/delivery/http/qms_client_routes.go:9-15` |
| `POST/GET/PATCH /api/v1/operator-counter-assignments` | done | `internal/modules/operator_assignment/delivery/http/operator_assignment_routes.go:9-14` |
| `POST /api/v1/caller/login` | done | `internal/modules/caller/delivery/http/caller_routes.go` |
| `GET /api/v1/caller/me` | done | `internal/modules/caller/delivery/http/caller_routes.go` |
| `POST /api/v1/caller/queue-journeys/{journey_id}/action` | done | `internal/modules/caller/delivery/http/caller_routes.go` |
| `GET /api/v1/signage/me` | done | `internal/modules/signage/delivery/http/signage_routes.go` |
| `GET /api/v1/signage/current-calls` | done | `internal/modules/signage/delivery/http/signage_routes.go` |
| `GET /api/v1/signage/queues` | done | `internal/modules/signage/delivery/http/signage_routes.go` |

### 1.2 Intentional MVP deviations

| Design route | Live status | Decision |
|---|---|---|
| `GET/PATCH /api/v1/tenant/queue-settings` | deferred | MVP reads through `GET /api/v1/settings/effective`; generic typed write path not exposed as dedicated route. |
| `GET/PATCH /api/v1/branches/{branch_id}/queue-settings` | deferred | Same as above; reset alias exists, effective alias exists, but direct CRUD route skipped. |
| `POST /api/v1/qms-clients/{client_id}/rotate-key` | alternative | Live runtime exposes `POST /api/v1/qms-clients/credentials` creation path instead of design verb-specific rotate alias. |

Conclusion: route parity is complete for runtime-critical MVP paths; remaining differences are explicit MVP scope choices, not missing backend wiring.

## 2. Typed Settings Write Audit Parity

### 2.1 Verified audited write paths

| Write path | Audit status | Evidence |
|---|---|---|
| Queue register | done | `internal/modules/queue/usecase/queue_usecase.go:301` |
| Queue forward | done | `internal/modules/queue/usecase/queue_usecase.go:455` |
| Queue transition | done | `internal/modules/queue/usecase/queue_usecase.go` (`QUEUE_CALL` / transition audit helper) |
| Scanner register/forward | done | `internal/modules/scanner/usecase/scanner_usecase.go:131-160` |
| Service create/update/delete | done | `internal/modules/service/usecase/service_usecase.go:77`, `:156`, `:168` |
| Branch-service create/update/delete | done | `internal/modules/service/usecase/branch_service_usecase.go:69`, `:114`, `:126` |
| Counter write paths | done | `internal/modules/counter/usecase/counter_usecase.go` |
| Organization / branch profile writes | done | `internal/modules/organization/usecase/organization_usecase.go`, `internal/modules/organization/usecase/branch_usecase.go` |
| QMS client update/deactivate | done | `internal/modules/qms_client/usecase/qms_client_admin_usecase.go` |
| Operator assignment create/unassign | done | `internal/modules/operator_assignment/usecase/operator_assignment_usecase.go:65`, `:125` |
| Typed settings reset-to-inherit | done | `internal/modules/settings/delivery/http/settings_controller.go:94` |

### 2.2 Real parity result

- `SETTING_RESET` is audited.
- Write-heavy setup domains are audited.
- Read/list endpoints are intentionally not audited.
- Missing piece remains only where dedicated typed settings create/update routes do not exist yet; no hidden unaudited live write path was found in `internal/modules/settings`.

Conclusion: typed settings write-audit parity is complete for all live write paths. Remaining audit gaps are route-surface gaps, not missed side effects.

## 3. Exact Coverage Map — Unit / Integration / E2E

### 3.1 Typed settings / effective config

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Resolver fallback chain | yes | indirect | indirect | `internal/modules/settings/queue_settings_resolver_test.go:28` |
| Effective config response | yes | indirect | yes | `internal/modules/settings/delivery/http/settings_controller_test.go:79`, `tests/e2e/modules/qms_client_admin_config_e2e_test.go:18` |
| Effective config alias paths | yes | no dedicated | indirect | `internal/modules/settings/delivery/http/settings_controller_test.go:159` |
| Reset-to-inherit audit/write | yes | indirect | indirect | `internal/modules/settings/delivery/http/settings_controller_test.go:197` |

### 3.2 Queue lifecycle / operational rules

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Queue register / forward / transition | yes | yes | yes | `internal/modules/queue/usecase/queue_usecase_test.go`, `tests/integration/modules/qms_queue_integration_test.go:68`, `tests/e2e/api/qms_queue_e2e_test.go:45` |
| Journey lifecycle call/recall/serve/complete | yes | yes | partial | `internal/modules/queue/usecase/queue_usecase_test.go`, `tests/integration/modules/qms_journey_lifecycle_integration_test.go:27` |
| Scanner guard / branch-service validation | yes | yes | yes | `internal/modules/scanner/usecase/scanner_usecase_test.go`, `tests/integration/modules/qms_scanner_integration_test.go:116`, `tests/e2e/api/qms_queue_e2e_test.go:45` |
| Queue estimate / queue-left rules | yes | indirect | indirect | `internal/modules/queue/usecase/queue_usecase_test.go` |

### 3.3 Caller

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Caller login | yes | indirect | indirect | `internal/modules/caller/usecase/caller_usecase_test.go:179` |
| Caller me | yes | indirect | indirect | `internal/modules/caller/usecase/caller_usecase_test.go:260` |
| Caller execute action boundary | yes | yes | yes | `internal/modules/caller/usecase/caller_usecase_test.go:295`, `tests/integration/qms_caller_integration_test.go:28`, `tests/e2e/modules/qms_caller_signage_e2e_test.go:18` |

### 3.4 Signage

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Signage me branding/service payload | yes | yes | indirect | `internal/modules/signage/usecase/signage_usecase_test.go:127`, `tests/integration/qms_signage_integration_test.go:24` |
| Signage current calls binding | yes | yes | yes | `internal/modules/signage/usecase/signage_usecase_test.go:201`, `tests/integration/modules/qms_caller_signage_integration_test.go:29`, `tests/e2e/modules/qms_caller_signage_e2e_test.go:18` |
| Signage queues feed | yes | yes | indirect | `internal/modules/signage/usecase/signage_usecase_test.go:308`, `tests/integration/qms_signage_integration_test.go:24` |
| Disconnect / reconnect behavior | no live consumer backend test | no | placeholder | `tests/e2e/modules/qms_signage_disconnect_e2e_test.go:21` |

### 3.5 QMS client admin / credentials

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Create client | yes | yes | yes | `internal/modules/qms_client/usecase/qms_client_admin_usecase_test.go:45`, `tests/integration/modules/qms_client_integration_test.go:65`, `tests/e2e/modules/qms_client_admin_config_e2e_test.go:18` |
| Create credential / expiry auth | yes | indirect | indirect | `internal/modules/qms_client/usecase/qms_client_admin_usecase_test.go:108`, `internal/modules/qms_client/usecase/qms_client_authenticator_test.go:43` |
| Update / deactivate | yes | yes | yes | `internal/modules/qms_client/usecase/qms_client_admin_usecase_test.go:173`, `tests/integration/modules/qms_client_integration_test.go:65`, `tests/e2e/modules/qms_client_admin_config_e2e_test.go:18` |

### 3.6 Operator assignment

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Create/list/unassign + tenant guard | yes | yes | indirect | `internal/modules/operator_assignment/usecase/operator_assignment_usecase_test.go:38`, `tests/integration/modules/qms_operator_assignment_integration_test.go:82` |
| Caller enforcement against assignment | yes | indirect | indirect | `internal/modules/caller/usecase/caller_usecase_test.go`, `tests/integration/qms_caller_integration_test.go:28` |

### 3.7 Audit visibility

| Feature | Unit | Integration | E2E | Evidence |
|---|---|---|---|---|
| Audit create visibility on QMS writes | indirect | yes | yes | `tests/integration/modules/qms_audit_integration_test.go:27`, `tests/e2e/api/qms_audit_e2e_test.go:17` |

## 4. Final Result

### Closed
- Route alias parity for MVP runtime surface.
- Typed settings write-audit parity for all live write paths.
- Exact feature-to-test map now documented.

### Still intentionally open
- Standalone realtime consumer apps for caller/signage.
- Wizard UX.
- Dedicated tenant/branch queue-settings write endpoints.
- Docker-native execution remains environment-dependent even though tests exist.
