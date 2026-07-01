# QMS Final Verification Playbook

## Purpose

Playbook ini menutup Fase 8 QMS rebuild: handoff verifikasi yang bisa diulang oleh agent, reviewer, atau CI tanpa membaca raw log panjang.

Scope playbook:

- queue unit/usecase/repository hardening
- scanner unit/usecase/controller hardening
- QMS integration tests
- QMS E2E tests
- race and vet checks for changed queue/scanner packages

## Preconditions

Required tools:

- Go from `/home/user/sdk/go/bin/go`
- writable Go cache, recommended `GOCACHE=/tmp/gocache`
- Docker for integration and E2E tests that use Testcontainers

Recommended prefix:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache
```

Do not treat Snap Go as reliable on this machine.

## Actor And Context

QMS API tests create their own fixture state through integration/E2E setup.

Runtime assumptions:

- request tenant context comes from authenticated organization context
- queue operations require branch context
- scanner operations require `X-Client-ID` and `X-API-Key`
- forwarding uses `queue_journeys`, not a second `queues` row

## Phase 8 Verification Commands

### 1. Queue and scanner unit package suite

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -count=1 ./internal/modules/queue/... ./internal/modules/scanner/...
```

Expected result:

- all queue package tests pass
- all scanner package tests pass
- no cross-tenant or branch-scope regression appears in usecase/controller tests

### 2. Queue and scanner race suite

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -race -count=1 -timeout=60s ./internal/modules/queue/... ./internal/modules/scanner/...
```

Expected result:

- package tests pass under race detector
- no race report emitted

### 3. Queue and scanner vet

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go vet ./internal/modules/queue/... ./internal/modules/scanner/...
```

Expected result:

- vet exits successfully
- no package-level vet finding for QMS queue/scanner code

### 4. Focused QMS integration tests

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v -tags=integration -count=1 -timeout=20m ./tests/integration/modules -run 'TestQMSQueueIntegration|TestQMSScannerIntegration'
```

Expected result with Docker available:

- `TestQMSQueueIntegration` runs and passes
- `TestQMSScannerIntegration` runs and passes
- Testcontainers starts MySQL and Redis successfully

Expected result without Docker:

- tests skip with explicit Docker/Testcontainers setup message
- do not claim integration passed when skip occurs

### 5. Focused QMS E2E tests

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v -tags=e2e -count=1 -timeout=20m ./tests/e2e/api -run 'TestQMSQueueE2E'
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v -tags=e2e -count=1 -timeout=20m ./tests/e2e/modules -run 'TestScannerE2E_APIKeyCheckInFlow'
```

Expected result with Docker available:

- queue lifecycle E2E test runs and passes
- scanner API-key check-in flow runs and passes
- no QMS route emits unexpected auth, tenant, branch, or relation error

Expected result without Docker:

- tests skip with explicit Testcontainers setup message
- report Docker as environment blocker, not code success

## Optional Full E2E Sweep

Use this only when full-suite confidence is needed and Docker is available:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/e2e/... -tags=e2e -p 1 -timeout=15m
```

Expected result:

- full E2E package set passes
- QMS route registration appears in Gin route logs
- QMS scanner test includes register and forward success paths

## Evidence To Capture

Capture concise evidence, not raw logs:

- command run
- pass/fail/skip status
- Docker/Testcontainers availability
- exact failing test name if any
- exact error text for auth, tenant, branch, or relation failures

Do not commit raw E2E logs such as `e2e-test.txt` unless explicitly requested.

## Failure Triage

### Unit or repository failure

Check first:

- `internal/modules/queue/usecase/queue_usecase.go`
- `internal/modules/queue/repository/queue_repository.go`
- `internal/modules/scanner/usecase/scanner_usecase.go`

Likely causes:

- missing tenant or branch context
- queue lookup not tenant-scoped
- active journey fixture missing `queue_id`
- scanner payload missing required action fields

### Controller failure

Check first:

- `internal/modules/queue/delivery/http/queue_controller.go`
- `internal/modules/scanner/delivery/http/scanner_controller.go`

Likely causes:

- branch context not resolved before usecase call
- route param missing or malformed
- scanner header validation drift

### Integration or E2E failure

Check first:

- `tests/integration/setup/test_container.go`
- `tests/integration/modules/qms_queue_integration_test.go`
- `tests/integration/modules/qms_scanner_integration_test.go`
- `tests/e2e/api/qms_queue_e2e_test.go`
- `tests/e2e/modules/qms_scanner_e2e_test.go`

Likely causes:

- Docker unavailable
- migrations drift from entity shape
- seeded auth/tenant context drift
- route or middleware protection changed

## Completion Criteria

Fase 8 selesai bila:

- package unit tests pass for queue/scanner
- race detector passes for queue/scanner
- vet passes for queue/scanner
- QMS integration and E2E commands have explicit pass or explicit Docker skip evidence
- handoff report separates verified pass from skipped verification

