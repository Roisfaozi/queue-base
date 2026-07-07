# QMS Manual Test Flow

## Purpose

Dokumen ini adalah runbook pengujian manual untuk QMS rebuild.

Fokus dokumen ini:

- langkah manual end-to-end
- precondition env dan data
- skenario positif, negatif, edge, dan security
- expected result per langkah

Dokumen ini ditujukan untuk QA, UAT, dan developer yang ingin memverifikasi flow QMS tanpa langsung membaca source code.

## Scope

Flow manual yang dicakup:

- branch
- service
- counter
- settings
- queue register
- queue list dan detail
- queue transition
- queue forward
- visit history
- queue stats
- scanner check-in

## Read First

- `documentation/QMS_FEATURE_AND_E2E_GUIDE.md`
- `documentation/api/qms/README.md`
- `documentation/api/qms/QUEUE_API.md`
- `documentation/api/qms/SCANNER_API.md`

## Preconditions

Sebelum test manual dimulai, pastikan:

- backend berjalan
- database dan redis aktif
- user test bisa login
- user test punya tenant/organization aktif
- endpoint QMS sudah terdaftar dan permission Casbin sudah diberikan ke role/user test

## Base Request Template

Gunakan template ini untuk semua request manual via curl atau Postman.

### Common headers

```http
Authorization: Bearer <ACCESS_TOKEN>
X-Organization-ID: <TENANT_ID>
Content-Type: application/json
```

### Curl template

```bash
curl -X <METHOD> 'http://127.0.0.1:8080/api/v1/<path>' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{ ... }'
```

## Required Test Actor

Minimal siapkan aktor ini:

### 1. QMS Admin

Dipakai untuk:

- create branch
- create service
- create counter
- create/update/delete setting
- melihat stats, queue, dan history

Permission minimal:

- `branch:view`, `branch:manage`
- `service:view`, `service:manage`
- `counter:view`, `counter:manage`
- `settings:view`, `settings:manage`
- `queue:view`, `queue:manage`

### 2. QMS Operator / Caller

Dipakai untuk:

- list queue
- transition queue
- forward queue
- lihat history dan stats bila memang diizinkan

Permission minimal:

- `queue:view`
- `queue:manage`

### 3. Scanner Client

Dipakai untuk:

- scanner register
- scanner forward

Requirement tambahan:

- `X-Client-ID`
- `X-API-Key`

## Required Test Data

Gunakan data contoh ini agar konsisten:

- Tenant ID: `<TENANT_ID>`
- Branch A: `Rawat Jalan`
- Service A: `General Clinic`
- Service B: `Pharmacy`
- Counter A: `Loket 1`
- Counter B: `Meja Farmasi 1`

## Global Expected Boundary

Semua request QMS normal harus:

- pakai bearer token valid
- pakai `X-Organization-ID`
- tetap scoped ke tenant aktif

Semua request scanner harus tambahan:

- `X-Client-ID`
- `X-API-Key`

## Manual Test Sequence

## Phase 1 — Login and Tenant Context

### Step 1. Login

Action:

- login via aplikasi atau API

Expected:

- login sukses
- access token tersedia
- user punya tenant aktif

### Step 2. Set active organization

Action:

- pastikan tenant/organization aktif sudah benar di frontend atau header

Expected:

- request berikutnya memakai `X-Organization-ID` yang benar

## Phase 2 — Master Data Setup

## Scenario A — Create branch

Action:

1. buka screen branch atau panggil `POST /api/v1/branches`
2. isi `code` dan `name`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/branches' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{
    "code": "RJ",
    "name": "Rawat Jalan"
  }'
```

Expected:

- branch baru terbentuk
- status default `active`
- branch muncul di list branch

Negative:

- kirim `code` kosong -> harus gagal validasi
- kirim token tanpa permission `branch:manage` -> harus ditolak

Security:

- user tenant lain tidak boleh membaca/memutasi branch ini

## Scenario B — Create service

Action:

1. buka screen service atau panggil `POST /api/v1/services`
2. isi `code`, `name`, optional flag pharmacy

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/services' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{
    "code": "GEN",
    "name": "General Clinic",
    "is_pharmacy": false,
    "is_pharmacy_reception": false
  }'
```

Expected:

- service baru terbentuk
- status default `active`

Negative:

- `name` terlalu pendek -> gagal validasi

## Scenario C — Create counter

Action:

1. buka screen counter atau panggil `POST /api/v1/counters`
2. pilih branch
3. isi `code` dan `name`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/counters' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{
    "branch_id": "<BRANCH_ID>",
    "code": "C1",
    "name": "Loket 1"
  }'
```

Expected:

- counter baru terbentuk
- `branch_id` tersimpan benar
- status default `active`

Negative:

- `branch_id` kosong / invalid -> gagal validasi

Security:

- counter tidak boleh dibuat ke branch lintas tenant

## Phase 3 — Typed Effective Configuration

## Scenario D — Check branch reset time

Action:

1. pastikan branch sudah punya typed config `queue_reset_time` lewat seed/admin API yang menulis `branch_queue_settings`
2. panggil effective config untuk branch itu

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/settings/effective?branch_id=<BRANCH_ID>' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- response mengembalikan `queue_reset_time`
- response mengembalikan `queue_reset_time_source`
- response mengembalikan `queue_reset_time_inherited`

Negative:

- `branch_id` lintas tenant -> gagal auth/tenant validation
- `branch_id` invalid -> gagal validation/not found

## Scenario E — Check service ticket prefix

Action:

1. pastikan branch sudah mengaktifkan service lewat `branch_services`
2. pastikan service punya typed config `ticket_prefix`
3. panggil effective config dengan `branch_id` dan `service_id`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/settings/effective?branch_id=<BRANCH_ID>&service_id=<SERVICE_ID>' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- response mengembalikan `ticket_prefix`
- response mengembalikan `ticket_prefix_source = service` bila service override aktif
- queue register ke service itu memakai prefix efektif

## Scenario F — Verify inheritance order

Action:

1. siapkan tenant config `queue_reset_time = 04:00`
2. siapkan branch config `queue_reset_time = 05:00`
3. resolve effective config dengan `branch_id`

Example resolve request:

```bash
curl 'http://127.0.0.1:8080/api/v1/settings/effective?branch_id=<BRANCH_ID>' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- hasil resolve mengambil nilai branch `05:00`, bukan tenant `04:00`
- `queue_reset_time_source = branch`
- `queue_reset_time_inherited = true`

Edge:

- bila counter/service/branch tidak ada typed config, sistem fallback sampai tenant lalu default runtime
- generic `/settings/resolve` hanya compatibility non-core, bukan sumber queue core

## Phase 4 — Queue Registration

## Scenario G — Register queue from dashboard/API

Action:

1. pilih branch aktif
2. panggil `POST /api/v1/queues`
3. isi `branch_id`, `service_id`, `patient_name`, optional `patient_id`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{
    "branch_id": "<BRANCH_ID>",
    "service_id": "<SERVICE_A_ID>",
    "patient_id": "<PATIENT_ID>",
    "patient_name": "John Doe"
  }'
```

Expected:

- queue baru terbentuk
- `status = waiting`
- `queue_no` bertambah
- `ticket_no` terbentuk sesuai prefix
- `current_journey_id` tidak kosong

Verify tambahan:

- queue muncul di `GET /api/v1/queues?branch_id=...`
- history awal ada event `registration`

Negative:

- `patient_name` kosong -> gagal validasi
- `service_id` invalid -> gagal validasi atau ditolak relasi

Edge:

- register sebelum reset time -> `queue_date` harus hari sebelumnya

Security:

- branch tenant lain tidak boleh dipakai untuk register

## Scenario H — Duplicate registration guard

Action:

1. register pasien yang sama lagi pada branch dan business date yang sama

Expected:

- request kedua gagal `409 conflict` atau error konflik yang setara

## Phase 5 — Queue List and Detail

## Scenario I — List queue by branch

Action:

1. panggil `GET /api/v1/queues?branch_id=<branch_id>`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/queues?branch_id=<BRANCH_ID>&status=waiting' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- queue branch tersebut tampil
- queue branch lain tidak ikut tampil

Negative:

- tanpa `branch_id` pada flow consumer normal -> hasil tidak valid untuk QMS flow dan harus dianggap setup salah

## Scenario J — Get queue detail

Action:

1. panggil `GET /api/v1/queues/{id}`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- detail queue kembali sesuai ID dan tenant aktif

Security:

- queue tenant lain harus tidak bisa diambil

## Phase 6 — Queue Transition

## Scenario K — Call queue

Precondition:

- queue status `waiting` atau `skipped`

Action:

1. panggil `POST /api/v1/queues/{id}/transition`
2. body `{"action":"call"}`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/transition' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{"action":"call"}'
```

Expected:

- queue status jadi `calling`
- history punya event `call`

## Scenario L — Serve queue

Precondition:

- queue status `calling`

Action:

1. transition `serve`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/transition' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{"action":"serve"}'
```

Expected:

- queue status jadi `serving`
- history punya event `serve`

## Scenario M — Complete queue

Precondition:

- queue status `serving`

Action:

1. transition `complete`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/transition' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{"action":"complete"}'
```

Expected:

- queue status jadi `completed`
- history punya event `complete`

## Scenario N — Skip queue

Precondition:

- queue status `waiting` atau `calling`

Action:

1. transition `skip`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/transition' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{"action":"skip"}'
```

Expected:

- queue status jadi `skipped`
- history punya event `skip`

## Scenario O — Cancel queue

Action:

1. transition `cancel`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/transition' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{"action":"cancel"}'
```

Expected:

- queue status jadi `canceled`
- history punya event `cancel`

Negative:

- `serve` langsung dari `waiting` -> harus gagal
- `complete` dari `calling` -> harus gagal
- `cancel` untuk queue yang sudah `completed` -> harus gagal

## Phase 7 — Queue Forward

## Scenario P — Forward queue to next service

Precondition:

- queue punya current journey aktif
- destination service valid di tenant dan branch yang benar

Action:

1. panggil `POST /api/v1/queues/{id}/forward`
2. isi `destination_service_id`
3. optional isi `destination_counter_id`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/forward' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'Content-Type: application/json' \
  -d '{
    "destination_service_id": "<SERVICE_B_ID>",
    "destination_counter_id": "<COUNTER_B_ID>"
  }'
```

Expected:

- queue master tetap sama
- `current_journey_id` berubah ke journey baru
- journey lama tertutup sebagai `forwarded`
- journey baru status `pending`
- history punya event `forward`

Verify:

- `GET /api/v1/branches/{id}/services/{service_id}/queue-journeys` menampilkan journey baru

Negative:

- destination service invalid -> gagal
- queue ID invalid -> not found

Security:

- tidak boleh forward ke destination lintas tenant/branch yang tidak valid

## Phase 8 — Visit History and Operational Views

## Scenario Q — Visit journey history

Action:

1. panggil `GET /api/v1/queues/{id}/visit-journeys`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/queues/<QUEUE_ID>/visit-journeys' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- event history tampil berurutan
- minimal ada `registration`
- bila queue sudah dioperasikan, event tambahan ikut tampil

## Scenario R — Active journeys by service

Action:

1. panggil `GET /api/v1/branches/{id}/services/{service_id}/queue-journeys`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/<BRANCH_ID>/services/<SERVICE_A_ID>/queue-journeys?status=pending' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- hanya journey aktif untuk branch + service itu yang muncul

Security:

- akses lintas tenant harus ditolak

## Scenario S — Active journeys by counter

Action:

1. panggil `GET /api/v1/branches/{id}/counters/{counter_id}/queue-journeys`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/<BRANCH_ID>/counters/<COUNTER_A_ID>/queue-journeys?status=pending' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- hanya journey aktif untuk branch + counter itu yang muncul

## Scenario T — Queue stats

Action:

1. panggil `GET /api/v1/branches/{id}/queue-stats`

Example request:

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/<BRANCH_ID>/queue-stats' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>'
```

Expected:

- total queue hari ini benar
- active journeys benar
- completed visits benar
- waiting by service benar

Edge:

- branch tanpa queue tetap memberi hasil sukses dengan angka nol

## Phase 9 — Scanner Flow

## Scenario U — Scanner register

Precondition:

- scanner client ID dan API key valid

Action:

1. panggil `POST /api/v1/scanner/check-in`
2. kirim header `X-Client-ID` dan `X-API-Key`
3. body action `register`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/scanner/check-in' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'X-Client-ID: scanner-device-01' \
  -H 'X-API-Key: <SCANNER_API_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "register",
    "branch_id": "<BRANCH_ID>",
    "service_id": "<SERVICE_A_ID>",
    "patient_id": "<PATIENT_ID>",
    "patient_name": "John Doe"
  }'
```

Expected:

- queue baru terbentuk seperti flow register biasa
- response `data.action = register`

Negative:

- tanpa `X-Client-ID` -> gagal
- tanpa `X-API-Key` -> gagal
- action invalid -> gagal

## Scenario V — Scanner forward

Precondition:

- queue target valid
- destination service valid

Action:

1. panggil scanner check-in action `forward`

Example request:

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/scanner/check-in' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'X-Organization-ID: <TENANT_ID>' \
  -H 'X-Client-ID: scanner-device-01' \
  -H 'X-API-Key: <SCANNER_API_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "forward",
    "branch_id": "<BRANCH_ID>",
    "queue_id": "<QUEUE_ID>",
    "destination_service_id": "<SERVICE_B_ID>",
    "destination_counter_id": "<COUNTER_B_ID>"
  }'
```

Expected:

- queue berhasil forward
- response `data.action = forward`

Security:

- scanner branch A tidak boleh mutasi queue branch asing

## Negative Matrix Summary

Minimal jalankan negatif ini:

- create branch tanpa `code`
- create counter tanpa `branch_id`
- create setting enum invalid
- register queue tanpa `branch_id`
- register queue tanpa `patient_name`
- transition action tidak valid
- transition status sequence salah
- forward queue ke service invalid
- scanner tanpa header
- scanner action invalid

## Security Matrix Summary

Minimal jalankan boundary ini:

- tenant A tidak bisa baca branch/service/counter/queue tenant B
- tenant A tidak bisa forward queue tenant B
- tenant A tidak bisa baca visit history tenant B
- scanner branch A tidak bisa memproses queue branch lain bila validator aktif menolak
- user tanpa permission QMS tidak boleh memutasi endpoint manage

## Exit Criteria

Manual test dianggap lulus bila:

- semua master data bisa dibuat dan dibaca ulang
- settings inheritance terbukti benar
- queue register berhasil
- queue transition berhasil sesuai state machine
- queue forward tidak membuat queue master baru
- history dan stats bisa dibaca
- scanner register dan scanner forward berjalan
- negative dan security checks utama gagal dengan benar

## Notes

- untuk pengujian API manual, pakai referensi field dari `documentation/api/qms/*.md`
- untuk behavior bisnis, pakai `documentation/QMS_FEATURE_AND_E2E_GUIDE.md`
- untuk validasi otomatis, cocokkan hasil manual dengan targeted integration dan E2E tests QMS yang sudah ada di repo
