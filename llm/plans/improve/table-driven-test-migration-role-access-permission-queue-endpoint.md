# Plan — Table-Driven Migration for Role, Access, Permission, and Queue Registration Endpoint Tests

## Objective

Ubah test existing untuk scope berikut menjadi table-driven test **tanpa menambah skenario baru dan tanpa kehilangan skenario yang sudah ada**:

- `role`
- `access`
- `permission`
- flow pendaftaran endpoint baru pada seam `POST /api/v1/queues`

Layer yang masuk audit dan migrasi:

- unit test
- integration test
- E2E test

## Non-Goals

- tidak menambah skenario baru
- tidak memperluas scope ke modul lain
- tidak mengubah behavior runtime
- tidak mengubah dokumen referensi `llm/plans/roadmap/table-driven-test-migration.md`
- tidak mengubah template `llm/conventions/table-driven-test-template.go.tmpl`

## Core Rule

Migrasi harus menjaga **coverage parity**:

- setiap skenario existing tetap ada
- nama case boleh berubah ke format tabel
- grouping boleh berubah
- setup boleh dipusatkan ke `tests := []struct{...}`
- **tidak boleh ada skenario existing yang hilang**

## Evidence Inventory

### Unit Test Targets

#### Role

- `internal/modules/role/usecase/role_usecase_test.go`
  - `TestRoleUseCase_Create`
    - `Success`
    - `Conflict`
    - `DBErrorFind`
    - `DBErrorCreate`
  - `TestRoleUseCase_Update`
    - `Success`
    - `NotFound`
    - `DBErrorFind`
    - `DBErrorUpdate`
  - `TestRoleUseCase_GetAll`
    - `Success`
    - `DBError`
  - `TestRoleUseCase_Delete`
    - `Success`
    - `ForbiddenSuperadmin`
    - `NotFound`
    - `DBErrorFind`
    - `DBErrorDelete`
    - `DBErrorPermissionCleanup`
  - `TestRoleUseCase_GetAllRolesDynamic`
    - `Success`
    - `DBError`

- `internal/modules/role/delivery/http/role_controller_test.go`
  - create variants
  - get all variants
  - delete success/not found/forbidden
  - dynamic search success/error/binding error
  - XSS sanitization variants
  - update success/binding/usecase/XSS
  - `HandleError_Variants`

#### Access

- `internal/modules/access/test/access_usecase_test.go`
  - `TestCreateAccessRight`
    - `Success - Create Valid Access Right`
    - `Error - Repository Create Fails`
  - `TestGetAllAccessRights`
    - `Success - Has Data`
    - `Success - No Data`
    - `Error - Repository Fails`
    - `Success - Sanitize Inputs`
  - `TestCreateEndpoint`
    - `Success - Create Valid Endpoint`
    - `Error - Repository Create Fails`
  - `TestLinkEndpointToAccessRight`
    - `Success - Link Valid IDs`
    - `Error - Repository Fails`
  - `TestDeleteAccessRight`
    - `Success - Delete Access Right`
    - `Error - Not Found`
  - `TestDeleteEndpoint`
    - success + error branch

- `internal/modules/access/test/access_controller_test.go`
  - create access-right handler variants
  - list/search variants
  - create endpoint variants
  - link/unlink variants
  - delete variants

#### Permission

- `internal/modules/permission/test/permission_usecase_test.go`
  - `AssignRoleToUser`
    - success
    - user not found
    - user repo error
    - role not found
    - role repo error
    - enforcer remove error
    - enforcer add error
    - empty input
  - `GrantPermissionToRole`
    - success
    - role not found
    - role repo error
    - enforcer error
    - empty input
  - `RevokePermissionFromRole`
    - success
    - role not found
    - role repo error
    - remove policy not found
    - enforcer error
    - empty input
  - additional methods in same file follow same one-test-per-scenario shape and must be grouped per method, not dropped

- `internal/modules/permission/test/access_right_assignment_test.go`
  - role access-right status cases
  - assign access-right cases
  - revoke access-right cases

- `internal/modules/permission/test/permission_controller_test.go`
  - assign/revoke/grant/list/get/update/batch-check/access-right handler variants

#### Queue Registration Endpoint

- `internal/modules/queue/delivery/http/queue_controller_test.go`
  - `TestQueueController_Register`
  - `TestQueueController_GetAllSetsBranchContextFromQuery`
  - `TestQueueController_Transition`
  - `TestQueueController_Forward`
  - `TestQueueController_GetByID`
  - `TestQueueController_GetQueueStats`
  - `TestQueueController_GetAll`
  - `TestQueueController_GetAll_WithFilters`
  - `TestQueueController_GetJourneysByBranchAndService`
  - `TestQueueController_GetJourneysByBranchAndCounter`
  - `TestQueueController_GetVisitJourneys`

### Integration Test Targets

#### Role

- `tests/integration/modules/role_integration_test.go`
  - existing integration cases must be grouped into table form without adding new scenarios

#### Access

- `tests/integration/modules/access_integration_test.go`
  - create access-right success
  - delete not found
  - create endpoint + link
  - delete endpoint success
  - dynamic search
  - SQL injection prevention
  - duplicate link negative
  - unlink success
  - unlink non-existent negative

#### Permission

- `tests/integration/modules/permission_integration_test.go`
  - assign role
  - grant permission
  - revoke permission
  - update permission
  - get all permissions
  - get permissions for role
  - full lifecycle
  - grant non-existent role
  - assign role to non-existent user
  - revoke non-existent permission

- `tests/integration/scenarios/permission_batch_test.go`
  - permission batch scenario

- `tests/integration/scenarios/role_hierarchy_test.go`
  - role hierarchy scenario

### E2E Test Targets

#### Role

- `tests/e2e/api/role_e2e_test.go`
  - create: success, duplicate, empty name
  - delete: success, non-existent
  - get all: success
  - update: success, non-existent
  - dynamic search: success

#### Access

- `tests/e2e/api/access_e2e_test.go`
  - access-right CRUD flow
  - endpoint CRUD flow
  - link endpoint to access-right: success, invalid IDs
  - dynamic search: access-rights and endpoints

#### Permission

- `tests/e2e/api/permission_e2e_test.go`
  - role hierarchy
  - batch check
  - revoke role
  - remove inheritance
  - security after role deletion
  - access-right assignment flow
  - access-right flow E2E
  - dynamic RBAC security

#### Queue Registration Endpoint / Queue Flow

- `tests/e2e/api/qms_queue_e2e_test.go`
  - lifecycle and scanner guard flow remains in-scope only for endpoint-registration-related cases already present in this file

## Coverage Parity Mapping

### Rule A — One Existing Test Function May Become One Table

Contoh:

- `TestRoleUseCase_Create` yang sekarang punya 4 `t.Run` harus jadi satu tabel 4 entries
- `TestCreateAccessRight` yang sekarang punya 2 `t.Run` harus jadi satu tabel 2 entries

### Rule B — Many Single-Scenario Functions May Become One Table by Method/Handler

Contoh:

- `TestRoleHandler_Create_Success`
- `TestRoleHandler_Create_BindingError`
- `TestRoleHandler_Create_ValidationError`
- `TestRoleHandler_Create_UseCaseError`

Semua ini boleh digabung menjadi satu `TestRoleHandler_Create` berbasis tabel, **asal keempat skenario tetap ada**.

### Rule C — Scenario Names Must Be Preserved Semantically

Nama case boleh berubah ke bentuk seperti:

- `Positive_CreateSuccess`
- `Negative_BindingError`
- `Negative_ValidationError`
- `Negative_UseCaseError`

Tapi semantic case tidak boleh hilang.

## Migration Grouping Strategy

### Phase 1 — Unit Tests

#### Slice 1.1 — Role Usecase

Owner files:

- `internal/modules/role/usecase/role_usecase_test.go`

Action:

- ubah per method jadi `tests := []struct{...}`
- gabungkan semua `t.Run` existing ke tabel
- jangan tambah case baru

Verification:

- narrow check: target package role usecase tests

#### Slice 1.2 — Role Controller

Owner files:

- `internal/modules/role/delivery/http/role_controller_test.go`

Action:

- gabungkan single-scenario handler tests per handler menjadi tabel
- pertahankan XSS/sanitization/error mapping cases yang existing

Verification:

- narrow check: target package role controller tests

#### Slice 1.3 — Access Usecase

Owner files:

- `internal/modules/access/test/access_usecase_test.go`

Action:

- ubah semua method usecase tests ke tabel per method
- pertahankan sanitize/no-data/error cases

#### Slice 1.4 — Access Controller

Owner files:

- `internal/modules/access/test/access_controller_test.go`

Action:

- gabungkan variants per handler ke tabel
- pertahankan malformed JSON, validation, usecase error, invalid IDs

#### Slice 1.5 — Permission Usecase

Owner files:

- `internal/modules/permission/test/permission_usecase_test.go`
- `internal/modules/permission/test/access_right_assignment_test.go`

Action:

- **satu method = satu tabel**
- jangan gabungkan seluruh file ke satu tabel besar
- pertahankan semua branches existing termasuk empty input, repo error, enforcer error, not found

#### Slice 1.6 — Permission Controller

Owner files:

- `internal/modules/permission/test/permission_controller_test.go`

Action:

- gabungkan handler variants per endpoint jadi tabel
- pertahankan domain/context-sensitive scenarios yang sudah ada

#### Slice 1.7 — Queue Registration Controller

Owner files:

- `internal/modules/queue/delivery/http/queue_controller_test.go`

Action:

- fokus pertama pada `Register`
- lanjutkan endpoint lain yang masih satu file dengan pola tabel per handler
- pertahankan branch-context assertions, filters, and route-param assertions

### Phase 2 — Integration Tests

#### Slice 2.1 — Role Integration

Owner files:

- `tests/integration/modules/role_integration_test.go`

Action:

- kelompokkan test per capability bila setup cukup mirip
- jangan paksa seluruh file jadi satu tabel bila flow lifecycle berbeda
- target tetap table-driven per method/flow yang setara

#### Slice 2.2 — Access Integration

Owner files:

- `tests/integration/modules/access_integration_test.go`

Action:

- group CRUD/link/unlink/search/security cases ke tabel per capability
- pertahankan duplicate-link dan SQL-injection cases

#### Slice 2.3 — Permission Integration

Owner files:

- `tests/integration/modules/permission_integration_test.go`
- `tests/integration/scenarios/permission_batch_test.go`
- `tests/integration/scenarios/role_hierarchy_test.go`

Action:

- method-based integration tests boleh jadi tabel
- scenario tests boleh tetap satu fungsi jika satu skenario penuh, tapi internal case matrix dapat dipindah ke tabel bila memang sudah berupa variasi data
- jangan tambah scenario baru

### Phase 3 — E2E Tests

#### Slice 3.1 — Role E2E

Owner files:

- `tests/e2e/api/role_e2e_test.go`

Action:

- group subtests existing menjadi tabel per endpoint group

#### Slice 3.2 — Access E2E

Owner files:

- `tests/e2e/api/access_e2e_test.go`

Action:

- group CRUD and search variants ke tabel

#### Slice 3.3 — Permission E2E

Owner files:

- `tests/e2e/api/permission_e2e_test.go`

Action:

- hanya table-drive bagian yang sudah berupa variasi request/assertion dalam flow yang sama
- flow E2E existing tetap dijaga, tidak dipecah jadi skenario baru

#### Slice 3.4 — Queue Registration E2E

Owner files:

- `tests/e2e/api/qms_queue_e2e_test.go`

Action:

- pertahankan flow existing
- table-drive hanya bagian endpoint-registration-related yang sekarang sebenarnya hanya variasi payload/result dalam flow sama, tanpa menambah step baru

## No-Miss Checklist

Sebelum menyatakan migrasi selesai, cek:

- [ ] setiap nama `func Test...` existing di scope sudah dipetakan ke target tabel baru atau tetap eksplisit bila masih satu-flow scenario
- [ ] setiap `t.Run(...)` existing sudah menjadi entry tabel atau tetap ada jika function itu memang sudah table-driven
- [ ] setiap negative/security/sanitization case existing tetap hidup
- [ ] tidak ada skenario existing yang hilang demi “merapikan” tabel
- [ ] tidak ada skenario baru yang ditambahkan

## Verification Plan

### Narrow verification first

- role package tests
- access package tests
- permission package tests
- queue controller package tests

### Then integration/E2E for touched files only

- integration files yang disentuh
- E2E files yang disentuh

### Environment note

Integration dan E2E di repo ini kemungkinan perlu Docker/Testcontainers. Kalau environment blokir, laporkan blocker exact dan tetap jalankan narrow unit coverage yang tersedia.

## Recommended Execution Order

1. role unit
2. access unit
3. permission unit
4. queue registration controller unit
5. role integration + E2E
6. access integration + E2E
7. permission integration + E2E
8. qms queue integration/E2E touched slice

## Final Recommendation

Migrasi ini layak dilakukan **secara coverage-preserving refactor**:

- pakai table-driven test
- jangan tambah skenario
- jangan buang skenario
- pecah `permission` per method
- pakai `queue_usecase_test.go` sebagai precedent lokal utama

