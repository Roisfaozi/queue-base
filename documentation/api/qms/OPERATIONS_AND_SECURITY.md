# QMS Operations And Security Contract

## Purpose

Dokumen ini menutup Fase 9 QMS rebuild: kontrak operasional backend untuk API, audit, security boundary, dan edge-case production readiness.

Scope resmi Fase 9:

- API contract final untuk queue dan scanner
- audit event matrix
- tenant/branch isolation checklist
- scanner API-key behavior
- operational edge cases yang wajib dipertahankan

## Runtime Boundaries

### Tenant boundary

Semua QMS operation berjalan di tenant aktif.

Required behavior:

- request normal membawa `Authorization: Bearer <access_token>`
- request normal membawa `X-Organization-ID`
- repository read/write harus filter `tenant_id`
- entity dari tenant lain harus terlihat sebagai not found atau forbidden sesuai layer pemilik
- tidak ada fallback ke tenant kosong

### Branch boundary

Branch adalah child dari tenant dan boundary operasional harian.

Required behavior:

- queue register wajib punya branch context
- queue list/stat/journey read wajib branch-scoped
- queue detail/forward/transition/visit-history harus resolve branch dari queue tenant-scoped sebelum operasi branch-scoped
- scanner payload `branch_id` harus valid untuk tenant aktif

### Queue master invariant

`queues` adalah master ticket row.

Required behavior:

- register membuat satu row `queues`
- forward tidak membuat row `queues` baru
- forward membuat journey baru di `queue_journeys`
- `current_journey_id` menunjuk journey aktif terakhir
- `visit_journeys` menyimpan readable event history

## Queue API Contract

See detailed fields in `documentation/api/qms/QUEUE_API.md`.

### Register queue

Endpoint:

- `POST /api/v1/queues`

Required behavior:

- validates active tenant and branch context
- validates destination service in tenant scope
- calculates business date using settings fallback
- generates ticket number from scoped counter
- creates initial `queue_journeys` row
- creates readable registration history in `visit_journeys`

Failure behavior:

- missing tenant or branch context fails closed
- invalid service fails request
- duplicate registration rules must not cross tenant boundary

### Transition queue

Endpoint:

- `POST /api/v1/queues/{id}/transition`

Allowed actions:

- `call`
- `serve`
- `complete`
- `skip`
- `cancel`

Required behavior:

- queue lookup is tenant-scoped
- branch context is resolved from the queue before usecase transition
- invalid state transition is rejected
- terminal states cannot be re-opened accidentally
- transition appends or updates readable visit history

Failure behavior:

- queue from foreign tenant is rejected
- missing current active journey is rejected as not found
- empty or unknown action is rejected as bad request

### Forward queue

Endpoint:

- `POST /api/v1/queues/{id}/forward`

Required behavior:

- queue lookup is tenant-scoped
- branch context is resolved from the queue before forwarding
- current active journey is closed deterministically
- next journey is created with incremented sequence
- queue master row remains the same
- destination service is tenant-scoped
- destination counter, when present, is branch/tenant-scoped

Failure behavior:

- foreign tenant queue is rejected
- missing active journey is rejected as not found
- invalid destination relation is rejected
- required destination service is enforced

### Queue reads

Endpoints:

- `GET /api/v1/queues`
- `GET /api/v1/queues/{id}`
- `GET /api/v1/queues/{id}/visit-journeys`
- `GET /api/v1/branches/{id}/queue-stats`
- `GET /api/v1/branches/{id}/services/{service_id}/queue-journeys`
- `GET /api/v1/branches/{id}/counters/{counter_id}/queue-journeys`

Required behavior:

- tenant filter is mandatory
- branch filter is mandatory for branch-scoped list/stat/journey surfaces
- service/counter filters must preserve `queue_date` and `status` when supplied
- empty result is success with empty list or zero stats, not accidental error
- malformed query behavior must match Gin binding behavior documented by tests

## Scanner API Contract

See detailed fields in `documentation/api/qms/SCANNER_API.md`.

Endpoint:

- `POST /api/v1/scanner/check-in`

Required headers:

- `X-Client-ID`
- `X-API-Key`

Required common body field:

- `branch_id`

Supported actions:

- `register`
- `forward`

### Scanner register

Required behavior:

- authenticates scanner client/API key
- validates `branch_id` in tenant scope
- requires `service_id`
- validates service relation and workflow settings
- delegates queue creation to queue usecase
- never logs or audits API key value

### Scanner forward

Required behavior:

- authenticates scanner client/API key
- validates `branch_id` in tenant scope
- requires `queue_id`
- requires `destination_service_id`
- treats `destination_counter_id` as optional unless workflow setting requires counter for destination service
- delegates forwarding to queue usecase
- never logs or audits API key value

Failure behavior:

- missing scanner headers returns auth or bad request failure by controller contract
- invalid client/API key is rejected
- foreign branch/service/counter relation is rejected
- patient data and API key must not leak to audit metadata

## Audit Event Matrix

Audit behavior is operational evidence, not source of truth for mutation.

| Event | Trigger | Required Metadata | Must Not Include |
|---|---|---|---|
| `QUEUE_REGISTER` | queue register success | queue id, tenant id, branch id, service id, ticket/queue number | raw patient sensitive fields beyond intended audit policy |
| `QUEUE_FORWARD` | queue forward success | queue id, source/destination journey context, destination service/counter | API key, full scanner secret |
| `QUEUE_CALL` | transition `call` success | queue id, branch id, journey id, counter/service when available | unrelated patient payload |
| `SCANNER_REGISTER` | scanner register success | action, client id, branch id, resulting queue id | `X-API-Key`, raw secret |
| `SCANNER_FORWARD` | scanner forward success | action, client id, queue id, destination service/counter | `X-API-Key`, raw secret, unnecessary patient data |

Required behavior:

- audit write failure is non-blocking for queue/scanner business success where tests enforce this behavior
- audit metadata must stay minimal
- scanner API key must never be persisted in audit metadata

## Security Checklist

### Tenant isolation

- Cross-tenant queue read returns not found or forbidden.
- Cross-tenant transition is rejected.
- Cross-tenant forward is rejected.
- Cross-tenant visit history read is rejected.
- Cross-tenant service/counter relation is rejected.

### Branch isolation

- Queue operations cannot use a branch from another tenant.
- Counter must belong to target branch when counter is supplied.
- Branch-scoped journey list must not return journeys from another branch.
- Queue stats must not aggregate another branch.

### Scanner protection

- `X-Client-ID` is required.
- `X-API-Key` is required.
- invalid API key fails closed.
- missing `branch_id` fails closed.
- action is case-normalized only where covered by tests.
- API key never appears in response, audit, or normal logs.

### State machine protection

- `waiting -> calling` allowed.
- `calling -> serving` allowed.
- serving completion is allowed by tested lifecycle.
- skip/cancel paths are constrained by current state rules.
- completed, canceled, and skipped terminal paths cannot be mutated incorrectly.

## Operational Edge Cases

### Business date

Queue date must use configured reset time.

Fallback chain:

- `queue_reset_time`
- `reset_time`
- `04:00`

### Ticket prefix

Ticket prefix must use scoped settings.

Fallback chain:

- `ticket_prefix`
- `prefix`
- `A`

### Numbering strategy

Effective runtime strategy is sequential unless future implementation expands it.

Fallback chain:

- `numbering_strategy`
- `numbering`
- `sequential`

### Empty reads

Empty queue list, empty journey list, empty visit history, and zero stats are valid success responses when tenant/branch scope is valid.

### Missing active journey

Forward and transition must reject queues without active current journey. This protects corruption cases where the master queue exists but journey projection is incomplete.

## Verification References

Use `llm/test-playbooks/qms-final-verification.md` for repeatable commands.

Minimum proof before changing this contract:

- queue/scanner unit suite passes
- queue/scanner race suite passes when stateful code changes
- queue/scanner vet passes
- focused QMS integration/E2E pass or skip is reported with Docker blocker
- security-sensitive changes include cross-tenant or scanner-secret regression tests

## Audit Visibility Matrix Enrichment

Dokumen ini melengkapi matrix audit di Phase 11 slice 11.4.

### Event to actor mapping

| Event | Typical Actor | Entity | Required Core Fields |
|---|---|---|---|
| `QUEUE_REGISTER` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `ticket_no` |
| `QUEUE_FORWARD` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `from_journey_id`, `to_service_id` |
| `QUEUE_CALL` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `journey_id`, `status` |
| `QUEUE_SERVE` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `journey_id`, `status` |
| `QUEUE_COMPLETE` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `journey_id`, `status` |
| `QUEUE_SKIP` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `journey_id`, `status` |
| `QUEUE_CANCEL` | authenticated user or `system` | `queue` | `organization_id`, `user_id`, `entity_id`, `branch_id`, `journey_id`, `status` |
| `SCANNER_REGISTER` | scanner flow under authenticated org context | `scanner` | `organization_id`, `user_id`, `entity_id`, `branch_id` |
| `SCANNER_FORWARD` | scanner flow under authenticated org context | `scanner` | `organization_id`, `user_id`, `entity_id`, `branch_id` |

### Sensitive-field deny list

Audit payloads untuk QMS tidak boleh menyimpan:

- `X-API-Key`
- scanner raw secret
- full patient secret payload yang tidak perlu untuk audit tujuan operasional
- tenant-switch internal cookies atau session material

### Audit to metrics relationship

Metrics adalah sinyal agregat. Audit adalah detail forensik.

- metrics jawab: spike error apa naik?
- audit jawab: queue atau scanner event mana yang terpengaruh?
- audit tidak boleh dipaksa jadi metrics replacement

### Failure policy

Untuk queue/scanner flows yang sudah diuji:

- audit failure tetap non-blocking terhadap business success
- jika audit write gagal, incident harus terlihat dari application logs atau worker/audit visibility lain
- jangan ubah policy ini tanpa test regresi eksplisit
