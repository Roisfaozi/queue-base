# QMS Integration and E2E Parallel Plan

## Purpose

Reusable test plan for QMS rebuild work. Use when integration and e2e verification should run in parallel across independent slices.

This plan separates queue-domain behavior into parallel-safe test groups so different slices can be implemented and verified without stepping on the same fixtures.

## Scope

Current QMS targets:

- tenant and branch foundation
- queue master registration
- queue_journeys forward history
- visit_journeys readable history
- scanner check-in orchestration
- settings inheritance
- service and counter CRUD boundaries

## Preconditions

- Docker / Testcontainers available
- MySQL and Redis can start locally
- `go test -tags=integration` and `go test -tags=e2e` allowed
- user-installed Go used if Snap Go fails

## Parallel Slice Map

Run these slices independently when possible:

### Slice A — Tenant / Branch / Service / Counter CRUD

Owns:

- branch create / list / get / update / delete
- service create / list / get / update / delete
- counter create / list / get / update / delete
- settings resolve on tenant/branch/service/counter scope

Positive checks:

- create record under correct tenant succeeds
- list returns only same-tenant rows
- update persists sanitized fields

Negative checks:

- missing tenant context fails
- missing required body/query fields fail
- invalid UUID scope or branch ID fails

Edge checks:

- delete returns no content
- resolve falls back tenant -> branch -> service -> counter

Security checks:

- cross-tenant CRUD lookup rejects access
- tenant A cannot resolve tenant B override

Suggested commands:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/organization/... ./internal/modules/service/... ./internal/modules/counter/... ./internal/modules/settings/...`

### Slice B — Queue Master / Journey / Transition

Owns:

- queue registration
- queue get/list
- queue transition call / serve / complete / skip / cancel
- `queue_journeys` creation and forward history
- `visit_journeys` read path

Positive checks:

- register queue creates one master row
- forward appends journey and visit record
- transition call -> serve -> complete works

Negative checks:

- invalid transition action fails
- empty queue ID fails
- missing tenant/branch context fails

Edge checks:

- skip from waiting or calling allowed
- cancel blocked after terminal states
- forward from same service still keeps master row

Security checks:

- cross-tenant queue lookup rejects access
- cross-tenant transition rejects access
- `visit_journeys` only returns same-tenant queue history

Suggested commands:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/queue/...`

### Slice C — Scanner Orchestration

Owns:

- scanner register flow
- scanner forward flow
- credential validation
- workflow validation via settings resolver

Positive checks:

- register path sends request to queue usecase
- forward path sends destination service / counter correctly

Negative checks:

- missing headers fail
- invalid action fails
- invalid credential fails

Edge checks:

- whitespace or uppercase action normalizes safely
- nil request fails fast

Security checks:

- workflow forbidden rejects at relation validator
- cross-tenant scanner request rejects

Suggested commands:

- `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test ./internal/modules/scanner/...`

## E2E Flow Map

Run these only after integration slices are green.

### E2E 1 — QMS Tenant Flow

Actor: tenant admin / authenticated API client

Start path:

- create or seed tenant context

Expected outcome:

- tenant-scoped CRUD works
- cross-tenant access rejected

### E2E 2 — QMS Queue Lifecycle

Actor: operational API client

Start path:

- register queue
- forward queue
- inspect journeys and visit history

Expected outcome:

- master queue stays single row
- journeys append correctly

- visit history readable only in tenant scope

### E2E 3 — QMS Scanner Flow

Actor: scanner client

Start path:

- send `/scanner/check-in` with API-key headers

Expected outcome:

- register and forward flows succeed with valid headers
- bad headers or bad workflow return failure

### E2E 4 — QMS Settings Inheritance

Actor: tenant admin / settings client

Start path:

- create tenant, branch, service, counter overrides

Expected outcome:

- resolve follows inheritance order correctly
- invalid scope access rejected

## Parallel Execution Order

Recommended order:

1. Slice A and Slice B in parallel
2. Slice C in parallel with late Slice B cleanup
3. E2E 1-4 after integration pass

Reason:

- Slice A owns CRUD and scope foundation
- Slice B owns queue lifecycle, which depends on Slice A domain validity
- Slice C depends on Slice A and Slice B data boundaries but is test-isolated

## Acceptance Criteria

Consider QMS test plan ready when:

- every slice has positive, negative, edge, and security coverage
- focused Go tests pass for each slice
- integration suite passes with Docker available
- e2e suite passes with Docker available
- no cross-tenant leak is observable in route or usecase boundaries

## Final Audit Pass

When all integration and e2e slices are green, run one final audit pass before declaring QMS rebuild ready.

### Security Gap Report

Produce a short security-gap report that lists:

- confirmed safe paths
- unresolved or unverified paths
- Docker or environment blockers
- remaining OWASP-style risks by slice

Required categories:

- injection
- broken access control
- authentication/session integrity
- tenant isolation
- insecure direct object reference
- configuration / scope enforcement
- logging and sensitive data exposure

### OWASP-Style Boundary Audit

Check each QMS flow against:

- positive path survives validation and auth
- negative path fails closed
- edge path preserves business rules
- security path blocks cross-tenant or malformed access

Boundary checklist:

- branch under tenant only
- counter under branch only
- queue lookup scoped to tenant and queue ID
- scanner credentials bound to tenant and branch
- settings inheritance cannot escape tenant scope
- forward creates journey history, not another queue master row

### Final Pass Command Order

1. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/integration/... -tags=integration -p 1 -timeout=15m`
2. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -v ./tests/e2e/... -tags=e2e -p 1 -timeout=20m`
3. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test`
4. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-integration`
5. `PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache make test-e2e`

### Final Go / No-Go

Go only when:

- integration and e2e pass on local Docker-backed run
- QMS slice tests cover positive, negative, edge, and security cases
- remaining security gap report is either empty or explicitly accepted by user
- generated docs are current with route and contract changes

## Notes

- Do not rely on broad suite alone for security proof.
- Keep slice fixtures disjoint so tests can be run in parallel later.
- If Docker is unavailable, report exact blocker instead of claiming pass.
