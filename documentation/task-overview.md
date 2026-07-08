# QMS Typed Configuration MVP — Task Overview & Status

This document summarizes the high-level implementation state of the queue management system rebuild based on the architecture and design specs.

## Current State (As of 2026-07-08)

Backend core and frontend dashboard setup are **complete** and operational against the MVP design. The operational flows, tenant/branch modeling, typed configuration resolution, dedicated client boundaries (caller/signage), and operator assignments are live in code and fully covered by tests.

### What is Live and Completed

1. **Typed Configuration Schema**: `tenant_queue_settings`, `branch_queue_settings`, `branch_service_queue_settings`, and `counter_queue_settings` replaced the legacy JSON config.
2. **Settings Resolution**: The `queueSettingsResolver` resolves fallback chains (`counter` -> `branch_service` -> `branch` -> `tenant` -> `default`) transparently for both API reads and internal business rules.
3. **Queue Journey Core**: The queue statemachine (register -> forward -> call -> recall -> serve -> complete -> skip) runs correctly with event auditing and visit journey tracking.
4. **Caller Domain**: Dedicated `caller` endpoints exist for machine+human hybrid auth, fetching allowed scope, and transacting against queues (call/serve/complete) with counter-binding validation.
5. **Signage Domain**: Dedicated `signage` endpoints exist for fetching current bound calls, audio/narrative payloads, and branding fallback rules.
6. **Device & Operator Guard**: `qms_clients`, `qms_client_credentials`, and `operator_counter_assignments` tables are active, with admin CRUD and strong middleware protection preventing scope leakage.
7. **Frontend Dashboard UI**: Setup wizard, nested Queue Config patching, QMS Client management, and Operator Assignment pages are live under `/dashboard` in `apps/web`.
8. **Security & Audit**: All setup writes, state transitions, and config reset events emit structured logs and `tryAudit` events natively. Queue Config CRUD patches feature active audit logging.

### What is Explicitly Deferred or Skipped

1. **Realtime Websocket Apps**: The specs (`documentation/QMS_Frontend_App_Specs.md`) exist, but the standalone consumer apps for caller and signage are intentionally delayed. Basic components exist but production polish is deferred.
2. **Nested Config Response**: The `effective-config` API still returns flat compat-mode JSON to satisfy current frontend needs, rather than the heavily nested (`tenant: {}, branch: {}`) format.
3. **Data Backfill Migration**: Data structures are prepared but scripts to backfill production legacy data are not run yet (dev-phase only).

### Next Directions

The codebase MVP gap has been successfully closed. Ready for manual test validation via standard QA cycles and eventually production rollout.
