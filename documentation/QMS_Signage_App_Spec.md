# QMS Signage Application Spec

This document defines the future standalone **Signage Application** for QMS. The current `apps/web/dashboard/signage` page is only an administrative helper surface. The real Signage app is a public-facing display shown to customers.

## 0. Runtime Alignment

Live backend endpoints already exist for the signage data path:
- `GET /api/v1/signage/me`
- `GET /api/v1/signage/current-calls`
- `GET /api/v1/signage/queues`
- machine-only auth via `X-Client-ID` + `X-API-Key`
- running_text and logo fallback logic in `signage_usecase.go`

This spec stays as target-state for the dedicated signage app, not as proof the standalone UI lives in `apps/web` yet.

## 1. Purpose

The Signage Application is used to display queue information in waiting areas.

It must:
- authenticate as a machine/device only
- display tenant/branch branding
- show current called queues
- show waiting queues
- react to realtime queue updates
- optionally play audio / announcements

It is not an admin dashboard and should not require human login in normal operation.

---

## 2. Authentication Model

Signage uses **machine-only auth**:
- `X-Client-ID`
- `X-API-Key`

These resolve a `qms_client` with:
- tenant_id
- branch_id
- optional branch_service_id
- optional counter_id
- client_type = signage

No user session is required.

---

## 3. Core Responsibilities

### 3.1 Branding and Display Metadata

The app must call:
- `GET /api/v1/signage/me`

And use it to display:
- branch name
- running text
- logo asset id
- optional service scope
- optional counter scope
- optional audio metadata (`audio_id`, `audio_en`)

Fallback behavior:
- branch `running_text` → tenant `running_text`
- branch `logo_asset_id` → tenant `logo_asset_id`

### 3.2 Current Calls

The app must call:
- `GET /api/v1/signage/current-calls`

This drives the large “currently being called” section.

### 3.3 Waiting Queues

The app must call:
- `GET /api/v1/signage/queues`

This drives the side list / lower-third / ticker list of waiting queues.

---

## 4. Realtime Requirements

### 4.1 Channel

Subscribe to:

```text
queue:{tenant_id}:{branch_id}
```

### 4.2 Event Format

```json
{
  "channel": "queue:{tenant}:{branch}",
  "type": "queue_update",
  "event": "QUEUE_CALL",
  "data": {
    "id": "queue-id",
    "branch_id": "branch-id",
    "ticket_no": "A001",
    "queue_no": 1,
    "status": "calling"
  }
}
```

### 4.3 Signage Reaction Rules

On `queue_update`:
- re-fetch `current-calls`
- re-fetch `queues`
- if event is `QUEUE_CALL`, optionally trigger audio/visual highlight

For MVP: use **re-fetch** instead of local optimistic patching.

---

## 5. Audio / Announcement Behavior

If service metadata exists:
- `audio_id`
- `audio_en`
- later optional narrative instruction fields

Then signage app may:
- play pre-recorded bell or intro tone
- display highlighted queue number animation
- optionally call browser TTS or pre-generated audio composition

This is optional for MVP but should be accommodated in layout and event flow.

---

## 6. UI / UX Requirements

### 6.1 Main Layout Zones

Recommended zones:
- branding header (logo + branch name)
- hero current call area
- waiting list panel
- running text ticker
- connectivity status hidden/admin-only overlay

### 6.2 Display Constraints

The app must be resilient for:
- smart TV browser
- kiosk browser
- Raspberry Pi / low-power PC
- fullscreen mode

### 6.3 Required States

- loading metadata
- unauthorized / expired credential
- no current calls
- no waiting queues
- realtime disconnected
- backend offline
- fallback branding active

---

## 7. Runtime Scope Rules

- signage cannot access other branch data
- signage cannot access other tenant data
- if branch_service is bound, display may be narrowed to that service
- if counter is bound, display may be narrowed to that counter
- inactive client must fail immediately
- expired credential must fail immediately

---

## 8. Deployment Shape

### 8.1 Recommended Form

Prefer dedicated signage route/app under `apps/client` or separate deploy target with fullscreen-optimized rendering.

Reasons:
- lighter bundle
- better kiosk compatibility
- easier offline splash / reconnect loop
- no admin dashboard chrome

### 8.2 Not Recommended

Do not keep long-term as `apps/web/dashboard/signage`. That route is only a thin operator/admin test harness.

---

## 9. Security Requirements

- device credentials should be provisioned during setup, not manually entered every day
- avoid exposing admin tokens or user sessions
- do not allow arbitrary branch selection in production signage app
- no privileged admin actions from signage surface

---

## 10. Observability

Frontend logging should capture:
- signage metadata fetch success/failure
- current calls fetch success/failure
- queue list fetch success/failure
- websocket connect/disconnect/reconnect
- audio playback failure

Backend logging should capture client binding and credential failures.

---

## 11. MVP Build Order

1. Dedicated signage shell
2. Device credential bootstrap
3. Metadata (`/signage/me`) render
4. Current calls + queue list render
5. Realtime queue subscription
6. Optional audio / animation enhancements

---

## 12. Current State vs Target

### Current State
- `apps/web/src/app/[locale]/dashboard/signage/_components/signage-content.tsx`
- admin helper only
- manual credential input
- fetch buttons for me/current-calls/queues
- no standalone deployment assumptions

### Target
- fullscreen public display
- machine-auth only
- auto-refresh and realtime by default
- branch-scoped branding and queue visibility
