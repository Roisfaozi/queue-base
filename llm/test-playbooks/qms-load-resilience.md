# QMS Load And Resilience Playbook

## Purpose

Playbook ini menutup Phase 12 slice load/resilience untuk QMS: membuktikan queue numbering, registration, forwarding, scanner check-in, dan observability tetap aman saat traffic naik atau dependency bermasalah.

## Scope

- concurrent queue numbering
- concurrent registration pressure
- forwarding transaction rollback
- scanner check-in burst behavior
- metrics and audit signal during failures

## Preconditions

Required tools:

- Go from `/home/user/sdk/go/bin/go`
- writable Go cache, recommended `GOCACHE=/tmp/gocache`
- Docker for integration/E2E and load-style request lifecycle tests
- backend running for external load tools if used

Recommended prefix:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache
```

Do not use Snap Go for this repo on this machine.

## Slice 12.1 — Concurrent Numbering

Run repository concurrency proof:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -count=1 -run 'TestQueueRepository/NextQueueNumberConcurrent' ./internal/modules/queue/repository
```

Expected result:

- all workers receive unique sequential queue numbers
- no duplicate number appears for same tenant, branch, date, and prefix
- no skipped number appears in controlled package-local pressure test

Risk if failed:

- counter transaction is not safe under concurrent registration
- queue ticket uniqueness may rely only on database unique key after user-visible failure

## Slice 12.2 — Registration Burst

Run queue usecase and repository package tests together:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -count=1 ./internal/modules/queue/usecase ./internal/modules/queue/repository
```

Expected result:

- registration path uses repository-owned number allocation
- duplicate patient guard remains tenant and branch scoped
- no queue journey is created without its master queue

## Slice 12.3 — Transaction Rollback

Run rollback proof:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -count=1 -run 'TestQueueRepository/CreateRegistrationWithNumber/Negative_RollsBackCounterAndChildrenOnQueueInsertFailure|TestQueueRepository/UpdateQueueState/Negative_RollsBackOnJourneyError|TestQueueRepository/CreateForwarding/Negative_' ./internal/modules/queue/repository
```

Expected result:

- failed registration rolls back counter row and child rows
- failed state update rolls back queue mutation and visit insert
- failed forwarding does not create next journey or visit history

Risk if failed:

- operations can leave orphaned journeys or visit history
- counter can advance without queue row and create user-visible ticket gaps

## Slice 12.4 — Scanner Burst Smoke

When Docker-backed E2E is available, run scanner lifecycle test:

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v -tags=e2e -count=1 -timeout=20m ./tests/e2e/modules -run 'TestScannerE2E_APIKeyCheckInFlow'
```

Expected result:

- scanner API key auth still gates the flow
- register and forward actions complete through HTTP stack
- no scanner secret is accepted from body payload only

## Slice 12.5 — Operational Load Checklist

For staging or local Docker environment:

1. start backend with metrics enabled
2. generate concurrent queue registrations for one tenant/branch/service
3. generate mixed actions: call, serve, forward, complete, cancel
4. generate scanner register and forward calls with valid and invalid API keys
5. scrape metrics endpoint before, during, and after burst
6. inspect audit rows for expected QMS actions
7. verify queue list, stats, and visit history remain consistent

Expected signals:

- `app_qms_queue_operations_total{operation,status}` increments by operation and status
- `app_qms_scanner_check_ins_total{action,status}` increments by action and status
- invalid scanner/API requests do not leak secrets in response, logs, metrics, or audit
- dashboard stats match persisted queue and journey state after load stops

## Failure Triage

### Duplicate or missing queue numbers

Check first:

- `internal/modules/queue/repository/queue_repository.go`
- `db/migrations/000027_create_queues_and_journeys.up.sql`

Likely causes:

- counter transaction does not serialize by tenant, branch, date, and prefix
- database-specific locking behavior differs from SQLite package test
- registration moved number generation outside repository transaction

### Orphaned journey or visit rows

Check first:

- `CreateRegistrationWithNumber`
- `CreateForwarding`
- `UpdateQueueState`

Likely causes:

- write order changed without rollback proof
- child rows created outside transaction
- error path ignores affected row count

### Scanner failure spike

Check first:

- `internal/modules/scanner/usecase/scanner_usecase.go`
- scanner API key middleware/header validation
- queue relation validator

Likely causes:

- invalid branch/service/counter relation
- API key header drift
- tenant context missing in E2E fixture

## Completion Criteria

Phase 12 load/resilience is acceptable when:

- concurrent numbering unit proof passes
- rollback repository proof passes
- queue/scanner focused packages pass
- Docker-backed E2E either passes or is reported as Docker/Testcontainers skip
- metrics and audit signals are checked in staging before production release
