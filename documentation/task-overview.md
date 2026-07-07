# QMS Typed Configuration MVP — Task Overview & Status

This document summarizes the high-level implementation state of the queue management system rebuild based on the architecture and design specs.

## Current State (As of 2026-07-07)

Backend core is **almost complete** and operational against the MVP design. The operational flows, tenant/branch modeling, typed configuration resolution, and dedicated client boundaries (caller/signage) are live in code and covered by tests.

### What is Live and Completed

1. **Typed Configuration Schema**: `tenant_queue_settings`, `branch_queue_settings`, `branch_service_queue_settings`, and `counter_queue_settings` replaced the legacy JSON config.
2. **Settings Resolution**: The `queueSettingsResolver` resolves fallback chains (`counter` -> `branch_service` -> `branch` -> `tenant` -> `default`) transparently for both API reads and internal business rules.
3. **Queue Journey Core**: The queue statemachine (register -> forward -> call -> recall -> serve -> complete -> skip) runs correctly with event auditing and visit journey tracking.
4. **Caller Domain**: Dedicated `caller` endpoints exist for machine+human hybrid auth, fetching allowed scope, and transacting against queues (call/serve/complete) with counter-binding validation.
5. **Signage Domain**: Dedicated `signage` endpoints exist for fetching current bound calls, audio/narrative payloads, and branding fallback rules (branch logo falling back to tenant logo).
6. **Device & Operator Guard**: `qms_clients`, `qms_client_credentials`, and `operator_counter_assignments` tables are active, with admin CRUD and strong middleware protection preventing scope leakage.
7. **Security & Audit**: All setup writes, state transitions, and config reset events emit structured logs and `tryAudit` events natively.

### What is Explicitly Deferred or Skipped

1. **Realtime Websocket Apps**: The specs (`documentation/QMS_Frontend_App_Specs.md`) exist, but the standalone consumer apps for caller and signage are intentionally delayed.
2. **UI Setup Wizard**: A multi-step setup wizard flow is deferred in favor of standard CRUD pages on the dashboard.
3. **Nested Config Response**: The `effective-config` API still returns flat compat-mode JSON to satisfy current frontend needs, rather than the heavily nested (`tenant: {}, branch: {}`) format proposed in some design docs.
4. **Data Backfill Migration**: Data structures are prepared but scripts to backfill production legacy data are not run yet (dev-phase only).
5. **Native Docker E2E Execution**: Tests are written and compile (skip-path verified locally), but full native database execution relies on the CI/QA environment providing Docker.

### Next Directions

The codebase is ready for integration QA, frontend consumption of the new Caller/Signage APIs, or focused deployment of the admin CRUD UI on `apps/web`.

lihat `QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md`
kemudian `New Design — Typed Configuration Architecture for QMS.md`
