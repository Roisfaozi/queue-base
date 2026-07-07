# QMS Feature and E2E Guide

## Purpose

Dokumen ini menjelaskan fitur Queue Management System (QMS) pada repo ini dari sudut pandang runtime nyata:

- domain dan boundary multi-tenant
- logic flow end-to-end
- relasi antar master data dan queue lifecycle
- behavior settings inheritance
- skenario E2E fitur utama

Dokumen ini disusun dari live code, terutama route layer, controller, usecase, entity, dan test QMS yang aktif.

## Runtime Scope

Fitur QMS aktif di repo ini mencakup:

- Branch
- Service
- Counter
- Settings
- Queue
- Scanner

Frontend aktif yang sudah memakai surface QMS saat ini adalah `apps/web`.

## Core Concepts

### 1. Tenant-first boundary

Semua data QMS berada di bawah tenant atau organization aktif.

Implikasi runtime:

- semua route QMS berada di grup `tenantAuthorized`
- request wajib melewati auth/session/organization boundary
- query dan mutation harus tetap scoped ke tenant aktif

### 2. Branch adalah child dari tenant

Branch adalah boundary operasional harian. Queue tidak cukup hanya scoped ke tenant; queue juga scoped ke branch.

### 3. `queues` adalah master ticket row

Setiap ticket hanya punya satu row master di tabel `queues`.

Field penting:

- `id`
- `tenant_id`
- `branch_id`
- `queue_date`
- `ticket_no`
- `queue_no`
- `patient_id`
- `patient_name`
- `status`
- `current_journey_id`

### 4. `queue_journeys` menyimpan perjalanan operasional

Setiap perpindahan service/counter tidak membuat row master queue baru. Perubahan operasional terjadi di `queue_journeys`.

Contoh:

- queue baru dibuat ke service awal
- queue dipanggil
- queue dilayani
- queue di-forward ke service lain

Semua itu tetap 1 master queue, tetapi journey bertambah atau berubah.

### 5. `visit_journeys` adalah readable event history

`visit_journeys` dipakai sebagai jejak event yang mudah dibaca. Ia bukan sumber mutasi utama, tetapi projection/history yang mengikuti event runtime seperti:

- `registration`
- `call`
- `serve`
- `complete`
- `skip`
- `cancel`
- `forward`

## Master Data Flow

### Branch

Branch dipakai untuk membatasi operasi queue secara operasional. Queue register, queue list, queue stats, dan scanner flow semuanya butuh branch scope.

Status branch:

- `active`
- `inactive`

Default create:

- `status = active`

### Service

Service mewakili layanan tujuan dalam alur queue.

Field behavior penting:

- `is_pharmacy`
- `is_pharmacy_reception`

Status service:

- `active`
- `inactive`

Default create:

- `status = active`
- `is_pharmacy = false`
- `is_pharmacy_reception = false`

### Counter

Counter adalah loket/unit operasional di bawah branch.

Constraint operasional:

- counter selalu terkait ke `branch_id`

Status counter:

- `active`
- `inactive`

Default create:

- `status = active`

## Typed Configuration Inheritance Flow

Config QMS dipakai lewat typed tables per entity, bukan generic settings untuk core behavior.

### Inheritance order

Resolver berjalan dengan urutan:

1. Counter (`counter_queue_settings`)
2. Service (`service_queue_settings`)
3. Branch (`branch_queue_settings`)
4. Tenant (`tenant_queue_settings`)

Artinya bila key ditemukan di counter, nilai itu menang. Bila tidak ada, sistem turun ke service, lalu branch, lalu tenant. Kalau tenant juga kosong, pakai default runtime. API `GET /settings/effective` mengembalikan nilai efektif plus metadata asal (`*_source`, `*_inherited`).

### Value type

Value type yang valid:

- `string`
- `number`
- `boolean`
- `json`

### QMS typed keys yang dipakai queue flow

#### `queue_reset_time`

Dipakai untuk menentukan business date queue. Bila tidak ada, pakai default runtime `04:00`.

Behavior:

- bila request datang sebelum jam reset, `queue_date` dianggap hari sebelumnya
- bila request datang sesudah atau sama dengan jam reset, `queue_date` dianggap hari ini

#### `ticket_prefix`

Dipakai untuk prefix ticket.

Fallback:

- default runtime `A`

Contoh hasil ticket:

- `A001`
- `B014`

#### `numbering_strategy`

Saat ini runtime hanya mengizinkan strategi efektif `sequential`.

Fallback:

- default runtime `sequential`

## Queue Lifecycle Flow

## 1. Register queue

Input utama:

- tenant context aktif
- branch context aktif
- service tujuan awal
- patient identity

Runtime sequence:

1. validasi tenant dan branch context
2. validasi service lewat relasi `branch_services`
3. sanitasi `patient_name`
4. resolve reset time
5. hitung `queue_date`
6. resolve `ticket_prefix`
7. cek duplicate registration untuk branch + queue_date + patient
8. generate nomor antrean berikutnya
9. buat row `queues`
10. buat first row `queue_journeys` dengan `seq_no = 1` dan status `pending`
11. buat row `visit_journeys` dengan event `registration`

Output utama:

- queue master baru status `waiting`
- current journey menunjuk journey pertama

## 2. List queues

List queue selalu tenant-scoped dan branch-scoped.

Filter yang tersedia:

- `branch_id`
- `status`
- `queue_date`
- `service_id`

Catatan runtime:

- branch context wajib tersedia untuk hasil valid
- pada HTTP consumer normal, kirim `branch_id`

## 3. Transition queue

Action yang valid:

- `call`
- `serve`
- `complete`
- `skip`
- `cancel`

Status queue yang dipakai runtime:

- `waiting`
- `calling`
- `serving`
- `skipped`
- `canceled`
- `completed`

Rule transisi:

- `call`: hanya valid dari `waiting` atau `skipped`
- `serve`: hanya valid dari `calling`
- `complete`: hanya valid dari `serving`
- `skip`: valid dari `waiting` atau `calling`
- `cancel`: tidak valid bila queue sudah `completed` atau `canceled`

Setiap transition juga:

- mengubah status journey aktif
- menulis event baru ke `visit_journeys`

## 4. Forward queue

Forward tidak membuat queue master baru.

Runtime sequence:

1. ambil queue master
2. ambil current journey aktif
3. validasi destination service/counter terhadap tenant dan branch queue
4. ubah current journey jadi `forwarded`
5. buat next journey baru status `pending`
6. update `current_journey_id` pada queue master
7. tulis event `forward` di `visit_journeys`

Output:

- queue master tetap sama
- service/counter aktif bergeser ke journey baru

## 5. Queue journeys by service/counter

API ini dipakai untuk caller/operational board.

Tujuan:

- lihat antrean aktif per service
- lihat antrean aktif per counter

Boundary penting:

- branch diambil dari path param
- service/counter diambil dari path param
- tenant tetap diambil dari auth + org context

## 6. Visit history

Visit history adalah pembacaan event history queue dalam tenant yang sama.

Tujuan:

- audit operasional ringan
- tracking langkah queue
- debugging status flow

## 7. Queue stats

Stats queue diambil per branch dan business date aktif.

Field output:

- total queues today
- total active journeys
- total completed visits
- waiting count per service

Business date stats mengikuti resolver `queue_reset_time`, bukan selalu calendar date biasa.

## Scanner Flow

Scanner adalah entrypoint alternatif untuk operasi queue lewat client device.

### Additional headers

Selain auth tenant route biasa, scanner juga butuh:

- `X-Client-ID`
- `X-API-Key`

### Scanner actions

- `register`
- `forward`

### Scanner register flow

1. validasi auth tenant route
2. validasi header scanner
3. validasi body scanner
4. autentikasi scanner client ke tenant + branch
5. validasi relation service/destination bila perlu
6. panggil `RegisterQueue`

Catatan:

- generic `settings` tetap ada untuk compatibility non-core
- queue core baca typed config lewat effective config endpoint, bukan list settings generic

### Scanner forward flow

1. validasi auth tenant route
2. validasi header scanner
3. validasi body scanner
4. autentikasi scanner client ke tenant + branch
5. validasi destination service/counter
6. panggil `ForwardQueue`

## Security and Boundary Rules

QMS saat ini dilindungi oleh beberapa boundary sekaligus:

- bearer access token valid
- active user session
- tenant/organization context aktif
- API scope auto requirement
- Casbin authorization
- relation validation untuk tenant-branch-service-counter
- scanner header auth untuk route scanner

Contoh security behavior yang diproteksi:

- branch asing lintas tenant harus ditolak
- service/counter asing harus ditolak
- queue history tenant lain harus ditolak
- scanner tanpa header wajib gagal

## End-to-End Scenarios

## Scenario A — Setup master data

1. pilih tenant aktif
2. buat branch
3. buat service
4. buat counter
5. optional: buat setting tenant/branch/service/counter

Expected result:

- semua master data tersimpan di tenant aktif
- branch, service, counter dapat di-resolve kembali via ID dan list

## Scenario B — Register queue from dashboard

1. frontend pilih branch aktif
2. kirim `POST /api/v1/queues`
3. backend resolve business date dan prefix
4. backend buat queue + first journey + registration event

Expected result:

- queue status `waiting`
- `ticket_no` terbentuk sesuai prefix dan urutan
- `current_journey_id` tidak kosong

## Scenario C — Register queue from scanner

1. scanner kirim `POST /api/v1/scanner/check-in`
2. action `register`
3. header scanner tervalidasi
4. queue didaftarkan melalui queue usecase yang sama

Expected result:

- hasil queue register tetap konsisten dengan dashboard flow

## Scenario D — Caller transitions queue

1. caller ambil antrean aktif per service/counter
2. caller jalankan `call`
3. caller jalankan `serve`
4. caller jalankan `complete`

Expected result:

- queue status bergerak `waiting -> calling -> serving -> completed`
- visit history mencatat event `call`, `serve`, `complete`

## Scenario E — Forward queue

1. queue sudah punya current journey aktif
2. caller atau scanner jalankan forward ke service lain
3. journey lama menjadi `forwarded`
4. journey baru dibuat `pending`

Expected result:

- queue master tetap sama
- `current_journey_id` pindah ke journey baru
- visit history punya event `forward`

## Scenario F — Queue stats

1. kirim `GET /api/v1/branches/{id}/queue-stats`
2. backend hitung stats untuk business date aktif

Expected result:

- stats mengikuti branch dan reset time yang aktif

## Test Coverage Status

QMS saat ini sudah diverifikasi pada jalur utama berikut:

- package-level unit tests untuk queue, scanner, settings, database scope
- integration test `TestQMSQueueIntegration_LifecycleAndSettingsGuard`
- E2E test `TestQMSQueueE2E_LifecycleAndScannerGuard`

## Operational Notes

- untuk consumer HTTP biasa, kirim `branch_id` saat list/register queue
- status create branch/service/counter default ke `active`
- create setting default `value_type=string` dan `is_active=true`
- reset date default runtime adalah `04:00` Asia/Jakarta bila setting tidak ada
- prefix default runtime adalah `A` bila setting tidak ada
