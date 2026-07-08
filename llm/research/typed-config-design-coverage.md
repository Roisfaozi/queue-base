# QMS Typed Configuration Design Coverage

This document maps the typed configuration design against the current implementation.

## Schema Status
| Component | Table | Status | Notes |
| :--- | :--- | :--- | :--- |
| Tenant Settings | `tenant_queue_settings` | ✅ Done | Replaced default JSON scope. |
| Branch Settings | `branch_queue_settings` | ✅ Done | Replaced branch JSON scope. |
| Branch Service Settings | `branch_service_queue_settings` | ✅ Done | Replaced service JSON scope. Added `require_counter` flag. |
| Counter Settings | `counter_queue_settings` | ✅ Done | Replaced counter JSON scope. Limited fields (e.g. `auto_call_next`, `allow_recall`). |
| Profile Fields | `organizations` / `branches` / `services` | ✅ Done | Logo, timezone, and duration fields live in master entities. |
| Clients & Devices | `qms_clients` / `qms_client_credentials` | ✅ Done | Scoped client credentials for dedicated device UX. |
| Operator Assignment | `operator_counter_assignments` | ✅ Done | Scoped operator binding to branch/counter/user. |

## Runtime Backend Coverage
| Flow | Status | Notes |
| :--- | :--- | :--- |
| Settings Fallback (Resolver) | ✅ Done | Helper handles fallback from Counter -> Branch Service -> Branch -> Tenant correctly. |
| Queue Transitions | ✅ Done | New `queue_journeys` approach implemented for tracking service lifecycle. |
| Visit Journeys | ✅ Done | Internal timeline built for observability. |
| Security Context (Tenant/Branch) | ✅ Done | Strong context separation applied everywhere in `queue_config` and `queue`. |
| API Contracts | ✅ Done | Dedicated routes for config patching. `effective-config` API preserved for backward-compatibility structure. |
| Operator Setup/Auth | ✅ Done | Admin CRUD via operator assignment routes + Operator enforcement in Caller. |

## Frontend Coverage
| Component | Status | Notes |
| :--- | :--- | :--- |
| QMS Setup Wizard | ✅ Done | Added to `apps/web` with steps spanning Tenant -> Branch -> Services -> Branch Services -> Counters -> QMS Clients -> Operator Assignment. |
| Config Override UI | ✅ Done | `queue-config-dialog.tsx` allows typed patch and reset for nested hierarchy config. |
| Dashboard Queue Live | ✅ Done | Built using QMS client components. |
| Operator Assignments UI | ✅ Done | Added standalone `/dashboard/operator-assignments` CRUD page. |
| Caller/Signage Display | 🚧 Deferred | Dedicated React/Next implementations explicitly delayed in favor of specs. |

## Outstanding / Deprecated Features
1. **Realtime Websocket Apps**: Caller/Signage UI deferred.
2. **Generic Settings `service_queue_settings`**: Deprecated in favor of `branch_service_queue_settings`.
3. **Arrive / Check-In Flow**: Not part of MVP.
