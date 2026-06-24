# QMS Frontend Dashboard Handoff

## Scope

Dokumen ini handoff workflow frontend AI untuk QMS rebuild, fokus slice dashboard dulu.

Target sekarang:
- `apps/web` dashboard area
- `apps/client` dashboard/admin area

Non-target untuk dokumen ini:
- project domain cleanup atau removal note
- queue forwarding detail flow
- queue caller flow
- scanner check-in flow

Catatan domain:
- forwarding dan queue flow sebenarnya fokus di caller
- dashboard hanya jadi observability, control plane, dan admin surface
- `projects` bukan fokus QMS dashboard handoff ini dan tidak perlu dimasukkan ke note AI untuk dihapus nanti

## Source Of Truth

Urutan keputusan:
1. live code
2. route wiring
3. API contract
4. UI state/consumer
5. docs/test playbook

Backend boundary untuk dashboard:
- `internal/router/router.go`
- `internal/modules/stats`
- `internal/modules/organization`
- `internal/modules/user`
- `internal/modules/role`
- `internal/modules/permission`
- `internal/modules/access`
- `internal/modules/audit`
- `internal/modules/webhook`
- `internal/modules/api_key`
- `internal/modules/settings`
- `internal/modules/queue`
- `internal/modules/counter`
- `internal/modules/service`

Backend exclusion for QMS note:
- `internal/modules/project` bukan fitur yang perlu dibawa dalam implementasi QMS aktif
- jangan masukkan backend `project` sebagai dependency wajib untuk dashboard QMS
- jangan buat note AI yang mengarahkan pengembangan fitur `project` untuk QMS kecuali user minta eksplisit

Frontend boundary:
- `apps/web/src/app/[locale]/dashboard`
- `apps/web/src/components/dashboard`
- `apps/web/src/components/layout`
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/web/src/proxy.ts`
- `apps/client/app/features/*`
- `apps/client/app/routes/api-proxy.ts`

## Dashboard Surface Map

### `apps/web`

Observed routes:
- `/dashboard`
- `/dashboard/projects`
- `/dashboard/users`
- `/dashboard/organization/members`
- `/dashboard/organization/settings`
- `/dashboard/roles`
- `/dashboard/access`
- `/dashboard/access-rights`
- `/dashboard/audit`
- `/dashboard/settings`
- `/dashboard/profile`

Shared shell pieces:
- `apps/web/src/app/[locale]/dashboard/layout-client.tsx`
- `apps/web/src/app/[locale]/dashboard/_components/dashboard-context.tsx`
- `apps/web/src/app/[locale]/dashboard/_components/dashboard-shell-context.tsx`
- `apps/web/src/components/layout/sidebar.tsx`
- `apps/web/src/components/layout/dashboard/header.tsx`
- `apps/web/src/components/layout/dashboard/breadcrumbs.tsx`

High-signal consumers:
- `apps/web/src/app/[locale]/dashboard/_components/dashboard-stats.tsx`
- `apps/web/src/app/[locale]/dashboard/_components/recent-activity.tsx`
- `apps/web/src/components/dashboard/organization-switcher.tsx`
- `apps/web/src/components/dashboard/user-nav.tsx`
- `apps/web/src/components/dashboard/ai-chat/ai-chat-widget.tsx`

### `apps/client`

Observed dashboard/admin surfaces:
- `app/features/organizations/*`
- `app/features/projects/*`
- `app/features/users/*`
- route wiring in `app/routes.ts`
- backend pass-through in `app/routes/api-proxy.ts`

Main intent:
- keep dashboard/admin CRUD aligned with backend contracts
- avoid frontend-only authorization assumptions
- let proxy layer preserve auth/session semantics

## Dashboard AI Workflow

### Phase 1 - Inventory

Before editing UI:
- identify owner app
- list actual route tree
- list backend endpoints the screen consumes
- inspect proxy route if browser fetch goes through app proxy
- inspect session/org boundary on the request path

### Phase 2 - Contract

For each dashboard screen define:
- route path
- API endpoints
- response shape
- loading/error/empty states
- tenant or org context
- which actions mutate state
- which caches or revalidation keys change

### Phase 3 - Slice Order

Implement in this order:
1. route shell and navigation
2. data fetch hook or server loader
3. table/card/forms state
4. mutation flow
5. cache invalidation or revalidation
6. route guard and auth handling
7. E2E coverage

### Phase 4 - Protection

For every dashboard change protect these cases:
- positive case
- negative case
- edge case
- security or boundary case

Security examples:
- user from different tenant cannot see foreign org data
- session missing redirects or blocks
- API proxy cannot bypass backend auth
- stale org context does not leak previous organization data

## Dashboard E2E Matrix

### `apps/web`

#### `/dashboard`

Positive:
- authenticated user with org context opens dashboard and sees stats cards

Negative:
- unauthenticated user is redirected by proxy or layout guard

Edge:
- empty stats still render skeleton or zero-state without crash

Security:
- user with wrong org context cannot read dashboard stats from another tenant

#### `/dashboard/projects`

Status:
- existing starter/dashboard surface
- bukan prioritas QMS dashboard handoff
- jangan dijadikan removal target hanya karena tidak dipakai di slice QMS awal
- jangan dijadikan backend dependency target untuk implementasi QMS aktif

Positive:
- project list loads and create/edit/delete work with active organization

Negative:
- invalid payload shows validation error

Edge:
- empty project list renders zero state

Security:
- user without scope cannot mutate projects

#### `/dashboard/users`

Positive:
- users table loads and status/role actions work for authorized actor

Negative:
- invalid search or filter rejected by backend contract

Edge:
- pagination or empty result state renders cleanly

Security:
- cross-tenant user read blocked

#### `/dashboard/organization/*`

Positive:
- member list and settings reflect current org

Negative:
- unknown org or missing membership errors cleanly

Edge:
- org switch updates visible data without stale flash

Security:
- member admin action blocked outside tenant scope

#### `/dashboard/roles`, `/dashboard/access`, `/dashboard/access-rights`

Positive:
- matrix, role CRUD, and access-right assignment render from backend data

Negative:
- invalid permission payload fails with validation message

Edge:
- inherited permission tree with empty children still renders

Security:
- unauthorized role mutation blocked

#### `/dashboard/audit`

Positive:
- audit log table loads, search and export work

Negative:
- invalid date/filter rejected

Edge:
- export async shows pending state

Security:
- audit scope cannot read foreign tenant logs

#### `/dashboard/settings`

Positive:
- settings resolve and save with correct inheritance display

Negative:
- invalid setting key/value rejected

Edge:
- branch/service/counter fallback chain shows effective value

Security:
- tenant cannot edit setting outside scope

### `apps/client`

#### Organization dashboard pages

Positive:
- organization list page loads and CRUD mutation works through proxy

Negative:
- invalid organization create/update rejected

Edge:
- no org selected still shows deterministic empty state

Security:
- org admin action blocked without correct scope

#### Project dashboard pages

Status:
- existing client admin surface
- out of primary QMS dashboard scope
- jangan dimasukkan ke future removal note tanpa keputusan eksplisit user
- jangan diperlakukan sebagai backend feature wajib untuk QMS

Positive:
- project list and create/edit/delete flows work

Negative:
- duplicate or invalid project input fails

Edge:
- pagination/search state persists across reload

Security:
- project belongs to active org only

#### User dashboard pages

Positive:
- user list and role assignment flow works

Negative:
- invalid status or role change rejected

Edge:
- no users found state renders cleanly

Security:
- role assignment blocked by missing permission

## Implementation Checklist For Each Dashboard Slice

- confirm backend endpoint exists and is documented
- confirm proxy route preserves auth/cookies
- confirm active org context source
- confirm loading, empty, error, mutation, and success states
- confirm one positive, one negative, one edge, one security test
- confirm browser path after reload still stable
- confirm no frontend-only auth assumption

## Verification Order

1. narrowest component/unit test
2. route/proxy consumer test
3. local app build or typecheck
4. browser E2E
5. broader integration if backend mutation crosses tenant boundary

## Next Slice

After dashboard:
- users
- organization management
- roles and access control
- audit and settings
- queue caller flow
- queue forwarding flow

## Explicit Exclusion

Untuk handoff AI QMS ini:
- jangan buat note yang menyarankan penghapusan `project` surface hanya karena belum dipakai oleh queue flow awal
- anggap `project` sebagai starter surface existing sampai user minta refactor atau removal eksplisit
- backend `project` module juga tidak perlu dimasukkan ke scope implementasi fitur QMS aktif
- bila agent membuat roadmap backend QMS, `project` harus dianggap non-required kecuali ada permintaan eksplisit
