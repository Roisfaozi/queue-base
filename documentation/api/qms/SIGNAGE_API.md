# Signage API Reference

Signage surfaces read-only queue display for public or kiosk screens. Authentication uses QMS Client Credentials bound to `signage` type.

## `GET /api/v1/signage/me`
Returns signage device binding and branding context.

### Request Headers
- `X-Client-ID`: Client identifier
- `X-API-Key`: Client secret

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.client_id` | string | Yes | QMS client UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.branch_id` | string | Yes | Branch UUID |
| `data.branch_service_id` | string | No | Bound branch-service UUID |
| `data.counter_id` | string | No | Bound counter UUID |
| `data.client_type` | string | Yes | Always `signage` for signage client |
| `data.name` | string | Yes | Client display name |
| `data.running_text` | string | No | Marquee or info text |
| `data.logo_asset_id` | string | No | Asset UUID for logo |
| `data.branch_name` | string | No | Branch display name |
| `data.service_name` | string | No | Service display name |
| `data.counter_display_name` | string | No | Counter label |
| `data.audio_id` | string | No | Audio asset UUID for ID voice |
| `data.audio_en` | string | No | Audio asset UUID for EN voice |
| `data.narrative_instruction_id` | string | No | Narrative instruction asset UUID |
| `data.narrative_instruction_en` | string | No | EN narrative instruction asset UUID |

### Success Response
```json
{
  "data": {
    "client_id": "client-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "branch-uuid",
    "branch_service_id": "bs-uuid",
    "counter_id": "counter-uuid",
    "client_type": "signage",
    "name": "Waiting Area Display",
    "running_text": "Welcome to our clinic.",
    "logo_asset_id": "asset-uuid",
    "branch_name": "Main Branch",
    "service_name": "General Checkup",
    "counter_display_name": "Loket 1",
    "audio_id": "audio-uuid",
    "audio_en": "audio-uuid-en",
    "narrative_instruction_id": "audio-instruction-uuid",
    "narrative_instruction_en": "audio-instruction-uuid-en"
  }
}
```

## `GET /api/v1/signage/current-calls`
Returns current live calls bound to signage branch.

### Notes
- Filtered by authenticated signage client binding (by branch).
- Should be safe for auto-refresh / polling.
- Array order typically depends on the most recently called items.

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].queue_id` | string | Yes | Queue UUID |
| `data[].ticket_no` | string | Yes | Ticket label shown on display |
| `data[].counter_id` | string | Yes | Counter UUID |
| `data[].counter_display_name` | string | No | Counter label |
| `data[].service_id` | string | Yes | Service UUID |
| `data[].service_type` | string | No | Service type label |
| `data[].audio_id` | string | No | Audio asset UUID for ID voice |
| `data[].audio_en` | string | No | Audio asset UUID for EN voice |
| `data[].narrative_instruction_id` | string | No | Narrative instruction asset UUID |
| `data[].narrative_instruction_en` | string | No | EN narrative instruction asset UUID |

### Success Response
```json
{
  "data": [
    {
      "queue_id": "queue-uuid",
      "ticket_no": "A001",
      "counter_id": "counter-uuid",
      "counter_display_name": "Counter 1",
      "service_id": "service-uuid",
      "service_type": "regular",
      "audio_id": "audio-uuid",
      "audio_en": "audio-en-uuid",
      "narrative_instruction_id": "inst-uuid",
      "narrative_instruction_en": "inst-en-uuid"
    }
  ]
}
```

## `GET /api/v1/signage/queues`
Returns queue list view for signage screens.

### Query Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `status` | string | No | Filter by status (e.g., `waiting`, `called`) |
| `limit` | int | No | Pagination |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Queue UUID |
| `data[].ticket_no` | string | Yes | Queue ticket label |
| `data[].status` | string | Yes | Queue status |
| `data[].patient_name` | string | No | Display name if public policy allows |
| `data[].queue_no` | int | No | Numeric queue order |

### Success Response
Standard paginated queue array.

```json
{
  "data": [
    {
      "id": "queue-uuid",
      "ticket_no": "A001",
      "status": "waiting",
      "patient_name": "John Doe",
      "queue_no": 1
    }
  ]
}
```
