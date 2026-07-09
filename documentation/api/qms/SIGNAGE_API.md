# Signage API Reference

Signage surfaces read-only queue display for public or kiosk screens. Authentication uses QMS Client Credentials bound to `signage` type.

## `GET /api/v1/signage/me`
Returns signage device binding and branding context.

### Request Headers
- `X-Client-ID`: Client identifier
- `X-API-Key`: Client secret

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
