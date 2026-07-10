# Queue API Reference

## `POST /api/v1/queues`

Register queue ticket.

### Example Request
```json
{
  "branch_id": "branch-uuid",
  "service_id": "service-uuid",
  "patient_id": "patient-uuid",
  "patient_name": "John Doe"
}
```

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
| `data.patient_id` | string | No | linked patient UUID |
| `data.current_journey_id` | string | No | current active journey ID |
| `data.queue_left` | integer | No | waiting queues before this ticket |
| `data.estimate_time_minutes` | integer | No | estimated wait time |
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
    "patient_id": "patient-uuid",
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "journey-uuid",
    "queue_left": 0,
    "estimate_time_minutes": 5,
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

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/queues?branch_id=branch-uuid&status=waiting&service_id=service-uuid&queue_date=2026-06-26' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | queue ID |
| `data[].tenant_id` | string | Yes | tenant |
| `data[].branch_id` | string | Yes | branch |
| `data[].queue_date` | string(YYYY-MM-DD) | Yes | operational business date |
| `data[].ticket_no` | string | Yes | alphanumeric code + queue no |
| `data[].queue_no` | integer | Yes | sequence number |
| `data[].patient_id` | string | No | linked patient UUID |
| `data[].patient_name` | string | No | sanitized display name |
| `data[].status` | string | Yes | queue current status |
| `data[].current_journey_id` | string | No | current active journey ID |
| `data[].queue_left` | integer | No | waiting queues before this ticket |
| `data[].estimate_time_minutes` | integer | No | estimated wait time |
| `data[].created_at` | integer | Yes | unix ms |
| `data[].updated_at` | integer | Yes | unix ms |

### Example Success Response
```json
{
  "data": [
    {
      "id": "queue-uuid",
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "queue_date": "2026-06-26",
      "ticket_no": "A001",
      "queue_no": 1,
      "patient_id": "patient-uuid",
      "patient_name": "John Doe",
      "status": "waiting",
      "current_journey_id": "journey-uuid",
      "queue_left": 0,
      "estimate_time_minutes": 5,
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

## `GET /api/v1/queues/{id}`

Get single queue detail.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/queues/queue-uuid' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | queue ID |
| `data.tenant_id` | string | Yes | tenant |
| `data.branch_id` | string | Yes | branch |
| `data.queue_date` | string(YYYY-MM-DD) | Yes | operational business date |
| `data.ticket_no` | string | Yes | alphanumeric code + queue no |
| `data.queue_no` | integer | Yes | sequence number |
| `data.patient_id` | string | No | linked patient UUID |
| `data.patient_name` | string | No | sanitized display name |
| `data.status` | string | Yes | queue current status |
| `data.current_journey_id` | string | No | current active journey ID |
| `data.queue_left` | integer | No | waiting queues before this ticket |
| `data.estimate_time_minutes` | integer | No | estimated wait time |
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
    "patient_id": "patient-uuid",
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "journey-uuid",
    "queue_left": 0,
    "estimate_time_minutes": 5,
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

## `POST /api/v1/queues/{id}/transition`

Mutate queue state (call, serve, complete, skip, cancel).

### Request Fields
| Field | Type | Required | Enum | Notes |
|---|---|---|---|---|
| `action` | string | Yes | `call`, `serve`, `complete`, `skip`, `cancel` | Transition |
| `counter_id` | string(uuid) | No | - | Counter identity context |

### Example Request
```json
{
  "action": "serve",
  "counter_id": "counter-uuid"
}
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | queue ID |
| `data.tenant_id` | string | Yes | tenant |
| `data.branch_id` | string | Yes | branch |
| `data.queue_date` | string(YYYY-MM-DD) | Yes | operational business date |
| `data.ticket_no` | string | Yes | alphanumeric code + queue no |
| `data.queue_no` | integer | Yes | sequence number |
| `data.patient_id` | string | No | linked patient UUID |
| `data.patient_name` | string | No | sanitized display name |
| `data.status` | string | Yes | queue status after transition |
| `data.current_journey_id` | string | No | active journey ID after transition |
| `data.queue_left` | integer | No | waiting queues before this ticket |
| `data.estimate_time_minutes` | integer | No | estimated wait time |
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
    "patient_id": "patient-uuid",
    "patient_name": "John Doe",
    "status": "serving",
    "current_journey_id": "journey-uuid",
    "queue_left": 0,
    "estimate_time_minutes": 0,
    "created_at": 1761800000000,
    "updated_at": 1761800500000
  }
}
```

## `POST /api/v1/queues/{id}/forward`

Forward a serving queue to another service.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `destination_service_id` | string(uuid) | Yes | target service |
| `destination_counter_id` | string(uuid) | No | target counter |

### Example Request
```json
{
  "destination_service_id": "service-uuid-2",
  "destination_counter_id": "counter-uuid-2"
}
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | queue ID |
| `data.tenant_id` | string | Yes | tenant |
| `data.branch_id` | string | Yes | branch |
| `data.queue_date` | string(YYYY-MM-DD) | Yes | operational business date |
| `data.ticket_no` | string | Yes | alphanumeric code + queue no |
| `data.queue_no` | integer | Yes | sequence number |
| `data.patient_id` | string | No | linked patient UUID |
| `data.patient_name` | string | No | sanitized display name |
| `data.status` | string | Yes | queue status after forward |
| `data.current_journey_id` | string | No | new active journey UUID |
| `data.queue_left` | integer | No | waiting queues before this ticket |
| `data.estimate_time_minutes` | integer | No | estimated wait time |
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
    "patient_id": "patient-uuid",
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "next-journey-uuid",
    "queue_left": 2,
    "estimate_time_minutes": 10,
    "created_at": 1761800000000,
    "updated_at": 1761800700000
  }
}
```

## `GET /api/v1/queues/{id}/visit-journeys`

List visit journey history for one queue.

### Path Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Queue ID |

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/queues/queue-uuid/visit-journeys' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Visit journey UUID |
| `data[].queue_id` | string | Yes | Queue UUID |
| `data[].tenant_id` | string | Yes | Tenant UUID |
| `data[].branch_id` | string | Yes | Branch UUID |
| `data[].event_type` | string | Yes | Event type |
| `data[].payload` | string | No | JSON payload string |
| `data[].created_at` | int64 | Yes | Event timestamp |

### Example Success Response
```json
{
  "data": [
    {
      "id": "visit-journey-uuid",
      "queue_id": "queue-uuid",
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "event_type": "registered",
      "payload": "{\"ticket_no\":\"A001\"}",
      "created_at": 1761800000000
    }
  ]
}
```

## `GET /api/v1/branches/{id}/queue-stats`

Get queue statistics for a branch.

### Path Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/queue-stats' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.total_queues_today` | int64 | Yes | Total queues today |
| `data.total_active_journeys` | int64 | Yes | Active journeys count |
| `data.total_completed_visits` | int64 | Yes | Completed visit count |
| `data.waiting_by_service` | object | Yes | Map of `service_id` to waiting count |

### Example Success Response
```json
{
  "data": {
    "total_queues_today": 42,
    "total_active_journeys": 8,
    "total_completed_visits": 34,
    "waiting_by_service": {
      "service-uuid": 5
    }
  }
}
```

## `GET /api/v1/branches/{id}/services/{service_id}/queue-journeys`

List queue journeys for one branch service.

### Path Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |
| `service_id` | string(uuid) | Yes | Service ID |

### Query Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `queue_date` | string(YYYY-MM-DD) | No | Business date filter |
| `status` | string | No | Journey status filter |
| `counter_id` | string(uuid) | No | Counter filter |

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/services/service-uuid/queue-journeys?queue_date=2026-06-26&status=waiting&counter_id=counter-uuid' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Queue journey UUID |
| `data[].queue_id` | string | Yes | Queue UUID |
| `data[].tenant_id` | string | Yes | Tenant UUID |
| `data[].branch_id` | string | Yes | Branch UUID |
| `data[].service_id` | string | Yes | Service UUID |
| `data[].counter_id` | string | No | Counter UUID |
| `data[].seq_no` | int | Yes | Journey sequence number |
| `data[].status` | string | Yes | Journey status |
| `data[].created_at` | int64 | Yes | Creation timestamp |
| `data[].updated_at` | int64 | Yes | Update timestamp |

### Example Success Response
```json
{
  "data": [
    {
      "id": "journey-uuid",
      "queue_id": "queue-uuid",
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "service_id": "service-uuid",
      "counter_id": "counter-uuid",
      "seq_no": 1,
      "status": "waiting",
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

## `GET /api/v1/branches/{id}/counters/{counter_id}/queue-journeys`

List queue journeys assigned to one counter.

### Path Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |
| `counter_id` | string(uuid) | Yes | Counter ID |

### Query Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `queue_date` | string(YYYY-MM-DD) | No | Business date filter |
| `status` | string | No | Journey status filter |
| `service_id` | string(uuid) | No | Service filter |

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/counters/counter-uuid/queue-journeys?queue_date=2026-06-26&status=waiting&service_id=service-uuid' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Queue journey UUID |
| `data[].queue_id` | string | Yes | Queue UUID |
| `data[].tenant_id` | string | Yes | Tenant UUID |
| `data[].branch_id` | string | Yes | Branch UUID |
| `data[].service_id` | string | Yes | Service UUID |
| `data[].counter_id` | string | No | Counter UUID |
| `data[].seq_no` | int | Yes | Journey sequence number |
| `data[].status` | string | Yes | Journey status |
| `data[].created_at` | int64 | Yes | Creation timestamp |
| `data[].updated_at` | int64 | Yes | Update timestamp |

### Example Success Response
```json
{
  "data": [
    {
      "id": "journey-uuid",
      "queue_id": "queue-uuid",
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "service_id": "service-uuid",
      "counter_id": "counter-uuid",
      "seq_no": 1,
      "status": "waiting",
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```
