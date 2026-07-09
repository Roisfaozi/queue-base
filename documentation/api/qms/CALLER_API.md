# Caller API Reference

The Caller API provides dedicated operations for device and web UI acting as Queue Operators. Callers represent a specific Counter within a specific Branch and Service.

Authentication uses QMS Client Credentials bound to `caller` type.

## `POST /api/v1/caller/login`
Validates QMS client credentials and returns a short-lived token (if web UI) or establishes session.

### Request Headers
- `X-Client-ID`: Client identifier
- `X-API-Key`: Client secret

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

## `GET /api/v1/caller/me`
Retrieves currently authenticated Caller context. Requires `caller` client authorization.

### Success Response
Same shape as Login response.
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
