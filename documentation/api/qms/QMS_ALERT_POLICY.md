# QMS Alert Policy

## Purpose

Dokumen ini menutup Phase 11 slice 11.3: policy alert operasional untuk QMS queue, scanner, audit, dan observability.

Tujuan:

- deteksi spike error lebih cepat
- jaga tenant/branch isolation tetap aman
- hindari secret leakage dari scanner
- jaga queue flow tetap stabil saat loket ramai

## Metric Sources

Metrics yang dipakai dari implementasi runtime saat ini:

- `app_qms_queue_operations_total{operation,status}`
- `app_qms_scanner_check_ins_total{action,status}`

Metrics pendukung dari sistem lain yang relevan:

- dashboard stats endpoint QMS
- audit log list
- websocket / realtime metrics bila dipakai untuk dashboard operasional

## Alert Rules

### 1. Scanner API key failure spike

Signal:

- `app_qms_scanner_check_ins_total{status="unauthorized"}` naik cepat

Suggested threshold:

- `>= 5` per 5 menit per tenant/cluster

Why it matters:

- bisa berarti scanner secret expired, salah deploy, atau abuse

First triage:

- cek validitas API key scanner
- cek apakah branch/org context berubah
- cek recent deploy scanner client

### 2. Scanner forbidden spike

Signal:

- `app_qms_scanner_check_ins_total{status="forbidden"}` naik

Suggested threshold:

- `>= 3` per 5 menit

Why it matters:

- branch mismatch, relation validator drift, atau device salah konteks

First triage:

- cek `branch_id` request
- cek service/counter relation
- cek org context active user/session

### 3. Queue forward failure spike

Signal:

- `app_qms_queue_operations_total{operation="forward",status="failed"}` naik

Suggested threshold:

- `>= 5` per 10 menit

Why it matters:

- forward path paling sensitif karena mengubah journey state dan history projection

First triage:

- cek current journey tersedia
- cek destination service/counter relation
- cek DB write / transaction error

### 4. Queue transition bad request spike

Signal:

- `app_qms_queue_operations_total{operation="transition",status="bad_request"}` naik

Suggested threshold:

- `>= 10` per 10 menit

Why it matters:

- caller flow salah state atau UI masih kirim action tidak valid

First triage:

- cek UI action mapping
- cek state machine matrix
- cek apakah user salah pilih queue status

### 5. Duplicate registration spike

Signal:

- `app_qms_queue_operations_total{operation="register",status="duplicate"}` naik

Suggested threshold:

- `>= 5` per 10 menit

Why it matters:

- bisa berarti double submit, retry bug, atau patient registration duplicate guard terlalu longgar

First triage:

- cek client double click
- cek network retry/refresh
- cek registration dedupe rule

### 6. Queue not found spike on forward/transition

Signal:

- `app_qms_queue_operations_total{status="not_found"}` naik

Suggested threshold:

- `>= 5` per 10 menit

Why it matters:

- foreign tenant access, stale queue id, atau branch context salah

First triage:

- cek queue id
- cek org/branch context
- cek resolver branch lookup dari queue id

## Triage Playbook

### Scanner incident

1. cek request header `X-Client-ID` dan `X-API-Key`
2. cek request body `branch_id`
3. cek relation validator untuk service/counter
4. cek audit log apakah event tercatat tanpa secret
5. cek apakah spike terjadi setelah deploy client baru

### Queue incident

1. cek queue state saat error
2. cek current journey id
3. cek destination relation
4. cek DB transaction and repo write path
5. cek stats endpoint apakah backlog naik abnormal

### Audit incident

1. cek audit log persisted
2. cek outbox/worker bila event delivery expected
3. cek apakah audit failure non-blocking behavior masih benar

## Suggested Dashboard Views

- queue operations by status
- scanner check-ins by action/status
- top branches by forward failure
- top branches by unauthorized scanner attempts
- active journeys vs completed visits trend

## Notes

- Jangan tambahkan label tenant_id atau branch_id ke Prometheus counter langsung kalau bisa dihindari; cardinality akan naik terlalu cepat.
- Audit log tetap sumber detail forensik, metrics hanya sinyal agregat.
- Alert threshold di atas baseline awal; sesuaikan setelah produksi punya data riil.
