# QMS Observability Readiness Checklist

## Purpose

Playbook ini menutup Phase 11 slice 11.5: checklist readiness operasional untuk observability QMS.

## Scope

- metrics endpoint visibility
- queue/scanner alert preparedness
- audit visibility
- manual triage commands
- stuck journey / stale queue detection

## Actors

- backend engineer on-call
- operator dashboard reviewer
- release engineer

## Preconditions

- backend running
- metrics endpoint reachable
- QMS routes enabled
- access to dashboard or Prometheus/Grafana UI bila ada

## Checklist

### Metrics health

- `app_qms_queue_operations_total` ter-export
- `app_qms_scanner_check_ins_total` ter-export
- metrics scrape success di environment target
- no high-cardinality labels added accidentally

### Queue readiness

- queue register increment muncul saat register success
- queue forward increment muncul saat forward success
- transition increment muncul saat call/serve/complete/skip/cancel
- bad request / not found / forbidden terhitung benar

### Scanner readiness

- scanner register increment muncul saat valid login + header
- scanner forward increment muncul saat valid forward
- unauthorized / forbidden spike terlihat jelas
- API key tidak muncul di log/response/audit

### Audit readiness

- `QUEUE_REGISTER`, `QUEUE_FORWARD`, `QUEUE_CALL`, `SCANNER_REGISTER`, `SCANNER_FORWARD` tercatat pada jalur sukses yang sudah diuji
- audit failure tidak memblokir business success bila policy tetap sama
- audit payload minimal dan tidak bocor secret

### Dashboard readiness

- branch queue stats dapat dibaca dari dashboard
- queue list caller view dapat difilter by branch/service/counter/status
- visit history detail dapat dibuka tanpa payload parsing error

## Triage Commands

### Queue usecase verification

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -count=1 -run 'TestRegisterQueue|TestForwardQueue|TestTransitionQueue|TestGetQueueStats|TestGetVisitJourneys' ./internal/modules/queue/usecase/
```

### Scanner usecase verification

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go test -count=1 -run 'TestScannerCheckIn|TestCheckIn' ./internal/modules/scanner/usecase/
```

### Metrics code verification

```bash
PATH=/home/user/sdk/go/bin:$PATH GOCACHE=/tmp/gocache go vet ./pkg/telemetry/ ./internal/modules/queue/usecase/ ./internal/modules/scanner/usecase/
```

## Incident Questions

- spike mana yang naik?
- action mana yang dominan?
- branch mana yang terpengaruh?
- apakah spike mulai setelah deploy?
- apakah backend atau client yang berubah?

## Cleanup / Follow-up

- simpan baseline metrics setelah beberapa hari produksi
- sesuaikan threshold alert setelah baseline nyata tersedia
- tambah runbook insiden untuk stuck queue dan scanner outage bila sudah ada data produksi
