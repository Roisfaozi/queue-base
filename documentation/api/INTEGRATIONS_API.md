# API Keys, Audit, Webhook, Realtime & Stats API Reference

## API Keys

Tenant-scoped API-key CRUD.

### `POST /api/v1/api-keys`
Create API key.

### Example Request
```json
{
  "name": "Web Dashboard Key",
  "description": "Used by admin dashboard"
}
```

### `GET /api/v1/api-keys`
List API keys.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/api-keys' \
  -H 'Authorization: Bearer <token>' \
  -H 'X-Organization-ID: <organization_id>'
```

### `DELETE /api/v1/api-keys/:id`
Delete API key.

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/api-keys/api-key-uuid' \
  -H 'Authorization: Bearer <token>' \
  -H 'X-Organization-ID: <organization_id>'
```

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

### Example Request
```json
{
  "filter": {"event_type": "USER_LOGIN"},
  "sort": [{"field": "created_at", "direction": "desc"}],
  "page": 1,
  "limit": 20
}
```

### Response Fields
Standard paginated envelope with `data[]`.

## Webhooks

Tenant-scoped outbound webhook management.

### `POST /api/v1/webhooks`
Create webhook.

### Example Request
```json
{
  "url": "https://example.com/webhook",
  "events": ["user.created", "queue.registered"]
}
```

### `GET /api/v1/webhooks`
List webhooks.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/webhooks' \
  -H 'Authorization: Bearer <token>' \
  -H 'X-Organization-ID: <organization_id>'
```

### `GET /api/v1/webhooks/:id`
Get webhook.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/webhooks/webhook-uuid' \
  -H 'Authorization: Bearer <token>' \
  -H 'X-Organization-ID: <organization_id>'
```

### `PUT /api/v1/webhooks/:id`
Update webhook.

### Example Request
```json
{
  "url": "https://example.com/webhook-v2",
  "events": ["queue.registered"]
}
```

### `DELETE /api/v1/webhooks/:id`
Delete webhook.

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/webhooks/webhook-uuid' \
  -H 'Authorization: Bearer <token>' \
  -H 'X-Organization-ID: <organization_id>'
```

### `GET /api/v1/webhooks/:id/logs`
List webhook logs.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/webhooks/webhook-uuid/logs' \
  -H 'Authorization: Bearer <token>' \
  -H 'X-Organization-ID: <organization_id>'
```

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
