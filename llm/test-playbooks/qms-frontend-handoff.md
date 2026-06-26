# QMS Frontend Handoff

## Status

- status: active reference
- current state: queue dashboard consumer already aligned with backend branch-scoped QMS contract
- verified on: 2026-06-26 Asia/Jakarta
- verified by: `pnpm --dir apps/web typecheck`, `pnpm --dir apps/client typecheck`, and repo pre-commit `lint` + `typecheck`

## Purpose

Dokumen ini handoff untuk agent frontend QMS rebuild.

Fokus: dashboard dan frontend admin surface.

## Scope

### In scope
- `apps/web`
- `apps/client`
- proxy route dan auth/session boundary di frontend
- dashboard shell, navigation, data loading, forms, tables, dialogs
- E2E and route-level verification untuk frontend surface

### Out of scope
- backend implementation detail
- migration
- queue core backend logic
- `project` backend feature
- `project` frontend as QMS dependency
- caller/backend queue forwarding logic
- scanner backend business logic

## Frontend Rules

- ikuti live route dan proxy, bukan asumsi docs lama
- auth guard frontend tidak boleh mengganti backend auth
- tenant/org context harus konsisten di loader, proxy, dan UI state
- loading, empty, error, mutation, and success states harus jelas
- jangan asumsikan route aman hanya karena UI menyembunyikan tombol

## Working Order

1. baca `AGENTS.md`
2. baca frontend map dan dashboard handoff yang sudah ada
3. cek route owner `apps/web` vs `apps/client`
4. cek endpoint yang dipakai screen
5. patch UI paling kecil yang perlu
6. sync proxy/type/hook jika contract berubah
7. verifikasi dengan test yang relevan

## Dashboard First

Prioritas dashboard UI:
- overview stats
- users
- organizations
- roles and access
- audit
- settings

Catatan:
- queue forwarding dan queue flow sebenarnya fokus di caller, bukan dashboard
- dashboard hanya observability / admin surface

## Test Matrix

Setiap perubahan frontend harus punya:
- positive case
- negative case
- edge case
- security/boundary case

Contoh:
- positive: dashboard load with valid session and tenant context
- negative: invalid payload shows validation error
- edge: empty state or no data still renders cleanly
- security: cross-tenant data is blocked by backend and not leaked by proxy/UI

## Verification Order

1. component/unit test
2. typecheck / build relevan
3. route or proxy verification
4. browser E2E untuk flow penting
5. broader integration only if needed

## Handoff Note

Jika agent frontend menemukan fitur `project`, anggap itu starter surface existing unless user explicitly asks to develop or remove it.

## Current QMS Frontend Truth

- active QMS dashboard consumer saat ini ada di `apps/web`
- `apps/client` belum punya consumer QMS aktif untuk queue/scanner lifecycle
- queue list dan queue register sekarang wajib membawa branch scope dari UI
- backend scanner route ada, tapi belum ada UI scanner aktif di frontend baru ini
