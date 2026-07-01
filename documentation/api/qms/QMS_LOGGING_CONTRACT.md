# QMS Logging Contract

## Purpose

Dokumen ini menutup Phase 11 slice 11.2 pada level kontrak logging lebih dulu, tanpa memaksa injection logger baru ke setiap usecase.

Tujuan:

- definisikan field log yang aman
- larang secret leakage
- tentukan kapan queue/scanner failure wajib terlihat di log

## Core Rules

### Must log

Untuk failure penting, log minimal harus bisa menjawab:

- action apa yang gagal
- tenant/organization context aktif apa
- branch context apa
- queue id mana bila tersedia
- service/counter mana bila tersedia
- error category apa

### Must not log

QMS runtime log tidak boleh berisi:

- `X-API-Key`
- raw scanner secret
- full patient payload yang tidak perlu
- auth token, cookie, atau session material

## Recommended Queue Failure Fields

Untuk queue register/forward/transition failure, field yang disarankan:

- `module=qms_queue`
- `operation=register|forward|transition`
- `tenant_id`
- `branch_id`
- `queue_id` bila ada
- `service_id` atau `destination_service_id` bila ada
- `counter_id` atau `destination_counter_id` bila ada
- `status=bad_request|not_found|forbidden|failed`
- `error`

## Recommended Scanner Failure Fields

Untuk scanner check-in failure, field yang disarankan:

- `module=qms_scanner`
- `action=register|forward`
- `tenant_id`
- `branch_id`
- `client_id`
- `queue_id` bila ada
- `destination_service_id` bila ada
- `destination_counter_id` bila ada
- `status=bad_request|unauthorized|forbidden|failed`
- `error`

## Event-to-log Guidance

### Scanner unauthorized

Log should include:

- `action`
- `tenant_id`
- `branch_id`
- `client_id`
- `status=unauthorized`

Log must not include:

- API key value

### Scanner forbidden

Log should include:

- `action`
- `tenant_id`
- `branch_id`
- `client_id`
- related `service_id` or `counter_id`
- `status=forbidden`

### Queue not found

Log should include:

- `operation`
- `tenant_id`
- `branch_id`
- `queue_id`
- `status=not_found`

### Queue failed write

Log should include:

- `operation`
- `tenant_id`
- `branch_id`
- `queue_id` if present
- relation ids if present
- `status=failed`

## Implementation Note

Saat ini Phase 11 logging contract boleh dipenuhi dulu sebagai dokumen bila runtime logger dependency belum mau diubah.

Jika nanti runtime logger ditambahkan ke usecase:

- gunakan low-noise structured logging
- pertahankan audit non-blocking policy
- jangan duplicate every success path into noisy logs

## Verification Note

Saat implementasi runtime logging dilakukan, minimal verifikasi:

- queue/scanner test tetap pass
- no API key appears in logs by test or manual inspection
- failure path logs include enough triage context
