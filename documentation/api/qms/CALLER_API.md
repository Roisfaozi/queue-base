# Caller API Reference

The Caller API provides dedicated operations for device and web UI acting as Queue Operators. Callers represent a specific Counter within a specific Branch and Service.

Authentication uses QMS Client Credentials bound to `caller` type.

## `POST /api/v1/caller/login`
Validates QMS client credentials and returns a short-lived token (if web UI) or establishes session.

### Request Headers
- `X-Client-ID`: Client identifier
- `X-API-Key`: Client secret

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `username` | string | Yes | Caller username, min 3 chars |
| `password` | string | Yes | Caller password, min 8 chars |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.access_token` | string | Yes | JWT access token |
| `data.context.tenant_id` | string | Yes | Tenant UUID |
| `data.context.tenant_name` | string | No | Tenant display name |
| `data.context.branch_id` | string | Yes | Branch UUID |
| `data.context.branch_name` | string | No | Branch display name |
| `data.context.branch_service_id` | string | No | Bound branch-service UUID |
| `data.context.service_name` | string | No | Service display name |
| `data.context.counter_id` | string | No | Bound counter UUID |
| `data.context.counter_name` | string | No | Counter display name |
| `data.context.display_name` | string | No | Human-friendly counter label |
| `data.permissions` | array<string> | Yes | Granted permission codes |

### Success Response
```json
{
  "data": {
    "access_token": "jwt-token-string",
    "context": {
      "tenant_id": "tenant-uuid",
      "tenant_name": "Hospital Name",
      "branch_id": "branch-uuid",
      "branch_name": "Main Branch",
      "branch_service_id": "bs-uuid",
      "service_name": "General Checkup",
      "counter_id": "counter-uuid",
      "counter_name": "C1",
      "display_name": "Counter 1"
    },
    "permissions": [
      "queue:manage",
      "queue:view"
    ]
  }
}
```

### Example Request
```json
{
  "username": "caller-01",
  "password": "secret-password"
}
```

## `GET /api/v1/caller/me`
Retrieves currently authenticated Caller context. Requires `caller` client authorization.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/caller/me' \
  -H 'Authorization: Bearer <access_token>'
```

### Success Response
Same shape as Login response.
### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.access_token` | string | Yes | Existing JWT access token |
| `data.context.tenant_id` | string | Yes | Tenant UUID |
| `data.context.branch_id` | string | Yes | Branch UUID |
| `data.context.branch_name` | string | No | Branch display name |
| `data.context.branch_service_id` | string | No | Bound branch-service UUID |
| `data.context.service_name` | string | No | Service display name |
| `data.context.counter_id` | string | No | Bound counter UUID |
| `data.context.counter_name` | string | No | Counter display name |
| `data.permissions` | array<string> | Yes | Granted permission codes |

```json
{
  "data": {
    "access_token": "jwt-token-string",
    "context": {
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "branch_name": "Main Branch",
      "branch_service_id": "bs-uuid",
      "service_name": "General Checkup",
      "counter_id": "counter-uuid",
      "counter_name": "C1"
    },
    "permissions": []
  }
}
```

## `POST /api/v1/caller/queue-journeys/:journey_id/action`
Executes an operational state transition on a specific queue journey. Action validation and state machine constraints apply.

### Path Params
- `journey_id`: UUID of the target `queue_journey`. Target journey must belong to the caller's bound branch.

### Body
| Field | Type | Required | Enum | Notes |
|---|---|---|---|---|
| `action` | string | Yes | `call`, `serve`, `complete`, `skip`, `cancel` | Operation to perform |

### Request Example
```json
{
  "action": "call"
}
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.success` | bool | Yes | Action outcome |
| `data.track_no` | string | No | Queue track number |
| `data.queue_no` | int | No | Queue sequence number |
| `data.status` | string | Yes | Resulting journey status |
| `data.journey_id` | string | Yes | Queue journey UUID |

### Special Action Behaviors
- **`call`**: Transitions waiting to called. If already called, transitions act as `recall` internally without needing a separate action string.
- **`serve`**: Marks the start of actual service (called -> serving).
- **`complete`**: Terminates service successfully.
- **`skip` / `cancel`**: Terminal failure/absence states.

### Example Response
```json
{
  "data": {
    "success": true,
    "track_no": "A001",
    "queue_no": 1,
    "status": "serving",
    "journey_id": "journey-uuid"
  }
}
```

### Curl (Action)
```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/caller/queue-journeys/journey-uuid/action' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "call"
  }'
```
