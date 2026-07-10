# Queue API Reference

## `POST /api/v1/queues`

Register queue ticket.

### Request Fields
| Field | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|
| `branch_id` | string(uuid) | Yes | - | - | active branch scope |
| `service_id` | string(uuid) | Yes | - | - | first destination service |
| `patient_id` | string(uuid) | No | - | - | optional patient ID |
| `patient_name` | string | Yes | - | - | 2..255 chars, sanitized |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | queue ID |
| `data.tenant_id` | string | Yes | tenant |
| `data.branch_id` | string | Yes | branch |
| `data.queue_date` | string(YYYY-MM-DD) | Yes | operational business date |
| `data.ticket_no` | string | Yes | alphanumeric code + queue no |
| `data.queue_no` | integer | Yes | sequence number |
| `data.patient_name` | string | Yes | sanitized display name |
| `data.status` | string | Yes | queue current status |
| `data.current_journey_id` | string | Yes | current active journey ID |
| `data.created_at` | integer | Yes | unix ms |
| `data.updated_at` | integer | Yes | unix ms |

### Example Success Response
```json
{
  "data": {
    "id": "queue-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "branch-uuid",
    "queue_date": "2026-06-26",
    "ticket_no": "A001",
    "queue_no": 1,
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "journey-uuid",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

## `GET /api/v1/queues`

List queues by active tenant and branch.

### Query Parameters
| Field | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|
| `branch_id` | string(uuid) | Yes | - | - | branch scope |
| `status` | string | No | queue status | - | filter |
| `service_id` | string(uuid) | No | - | - | filter |
| `counter_id` | string(uuid) | No | - | - | filter |
| `queue_date` | string | No | - | today | YYYY-MM-DD |

### Response Fields
Standard pagination envelope. Items shape matches `POST` response.

## `GET /api/v1/queues/{id}`

Get single queue detail.

### Response Fields
Matches `POST` response plus preloaded relations (depending on includes).

## `POST /api/v1/queues/{id}/transition`

Mutate queue state (call, serve, complete, skip, cancel).

### Request Fields
| Field | Type | Required | Enum | Notes |
|---|---|---|---|---|
| `action` | string | Yes | `call`, `serve`, `complete`, `skip`, `cancel` | Transition |
| `counter_id` | string(uuid) | No | - | Counter identity context |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | queue ID |
| `data.status` | string | Yes | new status |
| `data.current_journey_id` | string | Yes | active journey |

## `POST /api/v1/queues/{id}/forward`

Forward a serving queue to another service.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `destination_service_id` | string(uuid) | Yes | target service |
| `destination_counter_id` | string(uuid) | No | target counter |

### Response Fields
Returns updated queue record with a new `current_journey_id`.
