# API Keys, Audit, Webhook, Realtime & Stats API Reference

## API Keys

Tenant-scoped API-key CRUD.

- `POST /api/v1/api-keys`
  - request: (depends on controller)
  - response: returns raw secret once only
- `GET /api/v1/api-keys`
  - response: `data[]` — list of api key metadata (no secret)
- `DELETE /api/v1/api-keys/:id`
  - response: `204 No Content`

### Security Notes
- API key creation returns the raw secret once.
- Persisted records store secret hash only.
- Requires active organization context.

## Audit Logs

### `POST /api/v1/audit-logs/search`
Dynamic query-builder search. Uses filter/sort/page/limit pattern.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `filter` | object | No | dynamic filter payload |
| `sort` | array | No | sort field + direction |
| `page` | integer | No | pagination |
| `limit` | integer | No | results per page |

### Response Fields
Standard paginated envelope with `data[]`.

## Webhooks

Tenant-scoped outbound webhook management.

- `POST /api/v1/webhooks`
  - request: endpoint URL, events
- `GET /api/v1/webhooks`
  - response: `data[]` with webhook configs
- `GET /api/v1/webhooks/:id`
- `PUT /api/v1/webhooks/:id`
- `DELETE /api/v1/webhooks/:id`
- `GET /api/v1/webhooks/:id/logs`

### Delivery Details
- Signed with HMAC-SHA256 (`X-Webhook-Signature` header)
- Contains `X-Webhook-ID` and `X-Webhook-Event` headers
- Each attempt is logged in `webhook_logs`
- Requires `webhook:manage` scope

## Realtime

- `GET /api/v1/events` — Server-Sent Events
- `GET /api/v1/ws` — WebSocket entrypoint

Use `POST /api/v1/auth/ticket` when a short-lived ticket is required for WebSocket upgrade flows.

## Stats

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.total_users` | integer | Yes | user count |
| `data.total_roles` | integer | Yes | role count |
| `data.total_audit_logs` | integer | Yes | audit log count |
| `data.total_org_members` | integer | Yes | member count |

### Response Example
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
