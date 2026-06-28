# User & Auth TDT Migration Plan

## Scope
The goal is to migrate all remaining flat tests in the **User** and **Auth** (IAM) domains to the Table-Driven Testing (TDT) pattern. This includes:
1. **Unit Tests** (Repositories, Usecases, Controllers)
2. **Integration Tests** (Docker/Database-backed module tests)
3. **E2E Tests** (API boundaries and flow tests)

## Current State Analysis

### 1. Unit Tests
*   `internal/modules/user/test/user_usecase_test.go` (42 flat tests)
*   `internal/modules/user/repository/user_repository_test.go` (29 flat tests)
*   `internal/modules/user/delivery/http/user_controller_test.go` (14 flat tests)
*   `internal/modules/auth/test/auth_usecase_test.go` (95 flat tests)
*   `internal/modules/auth/test/auth_controller_test.go` (26 flat tests)
*   `internal/modules/auth/repository/token_repository_test.go` (25 flat tests)
*   Minor auth edge case files (`auth_timing_repro_test.go`, `repro_test.go`)

### 2. Integration Tests
*   `tests/integration/modules/user_integration_test.go` (23 flat tests)
*   `tests/integration/modules/auth_integration_test.go` (4 flat tests)

### 3. E2E Tests
*   `tests/e2e/api/user_e2e_test.go` (7 flat/mixed tests)
*   `tests/e2e/api/auth_e2e_test.go` (12 flat/mixed tests)

## Migration Strategy

### Phase 1: Repositories (Unit)
Convert `user_repository_test.go` and `token_repository_test.go`.
*   Pattern: `TestUserRepository(t *testing.T)` > `t.Run("Create", ...)` > `tests := []struct{}`
*   Challenge: Mocking `redis` for `token_repository` requires careful `setup` closures in the struct to inject exact redis-mock expectations per test case.

### Phase 2: Usecases (Unit)
Convert `user_usecase_test.go` and `auth_usecase_test.go`.
*   Pattern: `TestUserUseCase(t *testing.T)` > `t.Run("Register", ...)` > `tests := []struct{}`
*   Challenge: The 95 tests in `auth_usecase_test.go` must be carefully grouped by method (`Login`, `Refresh`, `Logout`, `ResetPassword`). Each test struct must define comprehensive mock behaviors (Repo, Casbin, Token, Hash).

### Phase 3: Controllers (Unit)
Convert `user_controller_test.go` and `auth_controller_test.go`.
*   Pattern: `TestUserController(t *testing.T)` > `t.Run("Create", ...)`
*   Challenge: HTTP request construction and mock assertions must be encapsulated in the test table struct.

### Phase 4: Integration Tests
Convert `user_integration_test.go` and `auth_integration_test.go`.
*   Pattern: Single setup block per test group (spinning up DB/Redis), followed by table cases. 
*   Challenge: State bleeding. In integration tests, the database persists between table rows. To fix this, each struct case MUST include a `cleanup` function or a transaction wrapper that rolls back, OR the `setup` function for the case must seed isolated, unique data (e.g. random UUIDs for usernames).

### Phase 5: E2E Tests
Convert `user_e2e_test.go` and `auth_e2e_test.go`.
*   Pattern: `t.Run` for endpoints with table-driven assertions on `resp.StatusCode` and response bodies.
*   Challenge: Same as integration — database state persists. API calls must use unique identifiers per table row or rely on isolated test fixtures.

## Required TDT Struct Format

For every converted test, we will enforce the standard QMS TDT struct:

```go
tests := []struct {
    name     string
    category string // "positive", "negative", "edge", "vulnerability"
    reqBody  interface{} // (Controllers)
    setup    func()      // (Mocks or DB Seeding)
    cleanup  func()      // (For Integration/E2E DB state reset)
    wantCode int         // (Controllers)
    wantErr  error       // (Usecases/Repos)
    assert   func(t *testing.T, ...)
}
```

## Execution Order

1.  Auth Repository -> User Repository
2.  Auth Usecase (massive file, needs extreme care)
3.  User Usecase
4.  Auth & User Controllers
5.  Integration Tests (Auth -> User)
6.  E2E Tests (Auth -> User)

By migrating layer by layer (bottom-up: repo -> usecase -> controller -> e2e), we guarantee that domain logic testing stays intact before modifying the API surface tests.