# Settings API Reference

Legacy note: for core QMS runtime, **do not use generic `/settings` CRUD as primary contract**.

QMS core now uses typed queue configuration endpoints documented in:

- `documentation/api/qms/QUEUE_CONFIG_API.md`

## Current QMS Core Rule

Use these endpoints instead of generic `settings` CRUD:

- `GET /api/v1/queue-config/effective`
- `PATCH /api/v1/queue-config`
- `GET /api/v1/branches/:branch_id/effective-config`
- `PATCH /api/v1/branches/:branch_id/queue-config`
- `DELETE /api/v1/branches/:branch_id/queue-config/:field`
- `GET /api/v1/branches/:branch_id/services/:service_id/effective-config`
- `PATCH /api/v1/branches/:branch_id/services/:branch_service_id/queue-config`
- `DELETE /api/v1/branches/:branch_id/services/:branch_service_id/queue-config/:field`
- `GET /api/v1/branches/:branch_id/counters/:counter_id/effective-config`
- `PATCH /api/v1/branches/:branch_id/counters/:counter_id/queue-config`
- `DELETE /api/v1/branches/:branch_id/counters/:counter_id/queue-config/:field`

## Compatibility Note

Generic `/settings` routes may still exist elsewhere in the platform for non-QMS domains or backward compatibility, but they are no longer the source of truth for Queue Management MVP behavior.
