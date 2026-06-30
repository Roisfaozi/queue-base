# QMS Slice 7 Handoff - Caller Flow Before Final Hardening

## Status

- status: historical handoff reference
- slice 7 state: completed
- follow-up state: targeted phase 9 integration and E2E checks already passed on 2026-06-26 Asia/Jakarta
- keep this doc as caller-flow intent and coverage checklist reference, not as open-work tracker

## Purpose

Dokumen ini handoff khusus untuk mengerjakan slice 7 QMS sebelum masuk fase 9 final hardening, integration, E2E, dan security audit akhir.

Slice 7 pada konteks repo ini adalah penyelesaian flow caller operasional di atas fondasi queue, journey, settings, dan scanner yang sudah lebih dulu dibangun.

## Position In Plan

Urutan praktis QMS saat ini:
1. tenant and branch foundation
2. queue master
3. queue journeys
4. visit journeys
5. settings inheritance
6. scanner orchestration
7. caller flow and operational route behavior
8. dashboard/admin support surface
9. final hardening + integration + E2E + security audit

Fokus dokumen ini hanya slice 7.

## Core Rule

- caller flow adalah pusat operasional queue nyata
- dashboard bukan pemilik flow queue
- forwarding dan transition harus tetap memakai `queue_journeys`
- `queues` tetap satu master row per ticket per hari
- `visit_journeys` tetap readable history/projection, bukan sumber mutasi utama

## In Scope

- route dan usecase operasional queue untuk caller
- list active journeys by service
- list active journeys by counter
- queue transition behavior
- queue forward behavior
- visit history read behavior bila diperlukan untuk caller context
- branch-scoped queue stats bila relevan untuk screen caller
- TDD coverage untuk positive, negative, edge, security

## Out Of Scope

- final integration suite penuh
- final E2E suite penuh
- final OWASP/security signoff
- frontend dashboard execution
- `project` feature
- generic admin cleanup yang tidak terkait caller

## Runtime Files To Read First

- `AGENTS.md`
- `internal/router/router.go`
- `internal/modules/queue/delivery/http/queue_routes.go`
- `internal/modules/queue/delivery/http/queue_controller.go`
- `internal/modules/queue/usecase/*`
- `internal/modules/queue/repository/*`
- `internal/modules/scanner/*`
- `internal/modules/settings/*`
- `llm/plans/roadmap/queue-management-rebuild.md`
- `llm/test-playbooks/qms-integration-e2e-parallel-plan.md`

## Current API Surface Relevant To Caller

Queue routes:
- `POST /queues`
- `GET /queues`
- `GET /queues/{id}`
- `GET /queues/{id}/visit-journeys`
- `POST /queues/{id}/forward`
- `POST /queues/{id}/transition`

Branch routes:
- `GET /branches/{id}/queue-stats`
- `GET /branches/{id}/services/{service_id}/queue-journeys`
- `GET /branches/{id}/counters/{counter_id}/queue-journeys`

## Caller Flow Goals

Slice 7 dianggap selesai bila caller flow berikut konsisten:
- caller bisa melihat antrean aktif per service
- caller bisa melihat antrean aktif per counter
- caller bisa memindahkan status antrean via transition
- caller bisa forward antrean ke service/counter berikutnya tanpa membuat master queue baru
- caller tidak bisa membaca atau memutasi data lintas tenant/branch
- caller behavior tetap konsisten terhadap reset date dan setting inheritance yang sudah ada

## Behavior Checklist

### 1. Active list by service

Protect:
- positive: branch + service valid returns active journeys
- negative: invalid branch/service rejected
- edge: empty journey list returns empty success response
- security: cross-tenant service/branch access rejected

### 2. Active list by counter

Protect:
- positive: branch + counter valid returns active journeys
- negative: invalid counter rejected
- edge: no active counter journeys returns empty success response
- security: cross-tenant counter access rejected

### 3. Transition queue

Protect:
- positive: valid transition updates current active journey status
- negative: invalid transition action/status rejected
- edge: repeated transition or already-finished journey handled deterministically
- security: queue from foreign tenant/branch rejected

### 4. Forward queue

Protect:
- positive: current journey closes and next journey opens with same queue master
- negative: invalid destination service/counter rejected
- edge: forward from queue without active journey handled safely
- security: forward into foreign tenant/branch/service/counter rejected

### 5. Visit journey read

Protect:
- positive: queue history visible for same tenant scope
- negative: invalid queue id rejected or not found
- edge: queue with no history beyond first journey still returns deterministic payload
- security: foreign tenant queue history rejected

## TDD Rule For Slice 7

Untuk setiap caller-flow change:
- write failing test first bila feasible
- implement smallest code to pass
- refactor setelah behavior terlindungi
- jangan nyatakan selesai tanpa positive, negative, edge, security case

## Recommended Work Order

1. audit live route and usecase behavior
2. identify missing caller invariants
3. write or extend tests first
4. patch usecase/repository/controller minimally
5. verify narrow package tests
6. verify caller-related route behavior
7. stop before phase 9 broad-suite finalization

## Verification Order

1. unit/controller test untuk queue route behavior
2. usecase test untuk forward/transition invariants
3. repository or integration-adjacent test hanya bila boundary DB benar-benar penting
4. no broad final-suite claim yet

## Historical Note

Saat dokumen ini dibuat, fase 9 belum ditutup.

Status terbaru sekarang:
- targeted integration pass for `TestQMSQueueIntegration_LifecycleAndSettingsGuard`
- targeted E2E pass for `TestQMSQueueE2E_LifecycleAndScannerGuard`
- caller-flow branch scope pada web queue consumer sudah diselaraskan
- final production signoff repo-wide tetap harus dibedakan dari QMS-scope signoff

## Expected Output From Agent

Saat mengerjakan slice 7, agent harus melaporkan:
- route/usecase mana yang disentuh
- invariant caller mana yang diperkuat
- test mana yang ditambah atau diperbaiki
- verification mana yang benar-benar dijalankan
- apa yang sengaja ditunda untuk fase 9
