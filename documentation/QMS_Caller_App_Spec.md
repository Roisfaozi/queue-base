# QMS Caller Application Spec

This document defines the future standalone **Caller Application** for QMS. The current `apps/web/dashboard/caller` page is only an administrative helper surface. The real Caller app is a dedicated operational surface used by staff at counters.

## 1. Purpose

The Caller Application is used by front-desk operators or service counters to:

- log in as an operator on a specific machine-bound device
- view waiting patients for the assigned counter/service
- call, serve, complete, skip, or cancel queue journeys
- forward queues to follow-up services (doctor → cashier → pharmacy)
- stay synchronized via real-time queue updates

It is not an admin dashboard. It is an operational execution tool.

---

## 2. User Model

Caller app uses **hybrid auth**:

1. **Machine identity** (device level)
   - `X-Client-ID`
   - `X-API-Key`
   - maps to `qms_clients`
   - scoped to tenant + branch + optional branch_service + optional counter

2. **Human identity** (operator level)
   - JWT / session from `POST /api/v1/caller/login`
   - operator must have active assignment in `operator_counter_assignments`

A valid caller session requires both layers:
- machine credential valid
- user credential valid
- user assignment valid for the same counter (when `counter_id` bound)

---

## 3. Core Responsibilities

### 3.1 Login and Context Resolution

The app must:
- accept `client_id` and `api_key`
- accept `username` and `password`
- call `POST /api/v1/caller/login`
- retrieve caller context (`tenant`, `branch`, `branch_service`, `counter`, `display_name`)
- show operator and counter context clearly on screen

### 3.2 Queue Journey Control

The app must support the action endpoint:
- `POST /api/v1/caller/queue-journeys/{journey_id}/action`

Allowed actions:
- `call`
- `serve`
- `complete`
- `skip`
- `cancel`

The app must disable invalid transitions in UI, but backend remains source of truth.

### 3.3 Queue List View

The app must show:
- waiting patients for relevant branch-service / counter
- current calling / serving journey
- last action result
- active journey refresh without manual reload

### 3.4 Forward Flow

If product wants forwarding from caller app, it should reuse queue forward APIs and relation validation. This is optional for MVP unless explicitly demanded.

---

## 4. Runtime Context

### 4.1 Scope Sources

Context is derived from:
- `qms_clients`
- `operator_counter_assignments`
- `queue_journeys`
- `branch_services`
- `counters`

### 4.2 Hard Rules

- tenant is first-class boundary
- branch is child of tenant
- operator assignment enforced only when `user_id` exists
- caller cannot act on journey outside same tenant/branch
- if `counter_id` is bound on client, action must match same counter
- inactive client must fail auth

---

## 5. Realtime Requirements

### 5.1 Channel

Caller app should subscribe to:

```text
queue:{tenant_id}:{branch_id}
```

### 5.2 Event Format

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

### 5.3 Caller App Behavior on Event

On any `queue_update` event:
- refresh active list
- refresh currently selected journey if visible
- reconcile state optimistically only when safe

For MVP, simplest safe behavior is **re-fetch**.

---

## 6. UI / UX Requirements

### 6.1 Minimal Surface

Must not expose admin features. Screen should focus on:
- current operator
- current counter
- queue list
- action buttons
- connection status
- logout / reset device session

### 6.2 States

The app must explicitly support:
- loading
- login failed
- no assignment
- no waiting queues
- queue calling
- queue serving
- disconnected realtime
- backend offline

### 6.3 Connectivity Indicator

Show device status:
- machine auth valid/invalid
- operator session active/inactive
- realtime connected/disconnected

---

## 7. Deployment Shape

### 7.1 Recommended Form

Prefer dedicated lightweight app under `apps/client` or separate deploy target. Reasons:
- smaller bundle
- fewer admin dependencies
- easier kiosk/fullscreen mode
- clearer access control boundary

### 7.2 Not Recommended

Do not keep long-term as dashboard-only route under `apps/web/dashboard/caller`. That surface is acceptable only as temporary admin helper.

---

## 8. Security Requirements

- never persist `api_key` in localStorage in plaintext if avoidable
- if persisted, device must be trusted/kiosk-only and document the risk
- JWT/session expiry must force re-login
- operator identity must be visible to reduce shared-terminal mistakes
- no cross-tenant or cross-counter action leakage
- no action allowed when client is inactive or credential expired

---

## 9. Observability

Caller app should emit frontend logs or structured telemetry for:
- login success/failure
- action success/failure
- websocket disconnect/reconnect
- stale context / branch mismatch

Backend already emits audit for caller actions; frontend logging is only for diagnosis.

---

## 10. MVP Build Order

1. Dedicated caller route/app shell
2. Device + operator login flow
3. Queue list + journey action buttons
4. Realtime queue subscription
5. Connection status indicator
6. Optional forwarding UI

---

## 11. Current State vs Target

### Current State
- `apps/web/src/app/[locale]/dashboard/caller/_components/caller-content.tsx`
- thin admin helper
- manual input of credentials
- queue fetch based on selected branch/service
- not standalone

### Target
- dedicated operator surface
- machine-bound and assignment-aware
- realtime by default
- minimal, kiosk-friendly, distraction-free UI
