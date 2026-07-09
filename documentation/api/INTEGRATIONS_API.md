# API Keys, Audit, Webhook, Realtime & Stats API Reference

## API Keys

Tenant-scoped API-key CRUD.

- `POST /api/v1/api-keys`
- `GET /api/v1/api-keys`
- `DELETE /api/v1/api-keys/:id`

### Security Notes
- API key creation returns the raw secret once.
- Persisted records store secret hash only.
- Requires active organization context.

## Audit Logs

### `POST /api/v1/audit-logs/search`
Dynamic query-builder search.

**Request:**
```json
{
  "filter": {
    "action": { "type": "equals", "from": "LOGIN" }
  },
  "sort": [{ "field": "created_at", "direction": "desc" }],
  "page": 1,
  "limit": 20
}
```

### `GET /api/v1/audit-logs/export`
Synchronous export.

### `GET /api/v1/audit-logs/export-async`
Asynchronous export.

## Webhooks

Tenant-scoped outbound webhook management.

- `POST /api/v1/webhooks`
- `GET /api/v1/webhooks`
- `GET /api/v1/webhooks/:id`
- `PUT /api/v1/webhooks/:id`
- `DELETE /api/v1/webhooks/:id`
- `GET /api/v1/webhooks/:id/logs`

See `documentation/guides/WEBHOOKS.md` for signing and event payload details.

## Realtime

- `GET /api/v1/events` — Server-Sent Events
- `GET /api/v1/ws` — WebSocket entrypoint

Use `POST /api/v1/auth/ticket` when a short-lived ticket is required for WebSocket upgrade flows.

## Stats

Dashboard stats endpoints expose summaries, activity points, and system insight metrics.

Response example:
```json
{
  "data": {
    "total_users": 100,
    "total_roles": 8,
    "total_audit_logs": 1200,
    "total_org_members": 20
  }
}
```
