# Queue API Reference

## `POST /api/v1/queues`

Register queue ticket.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `branch_id` | body | string(uuid) | Yes | - | - | active branch scope |
| `service_id` | body | string(uuid) | Yes | - | - | first destination service |
| `patient_id` | body | string(uuid) | Optional | - | - | optional patient ID |
| `patient_name` | body | string | Yes | - | - | 2..255 chars, sanitized |

### Runtime defaults

- `queue_reset_time` fallback `reset_time` fallback `04:00`
- `ticket_prefix` fallback `prefix` fallback `A`
- numbering strategy effective default `sequential`

### Example request body

```json
{
  "branch_id": "550e8400-e29b-41d4-a716-446655440100",
  "service_id": "550e8400-e29b-41d4-a716-446655440200",
  "patient_id": "550e8400-e29b-41d4-a716-446655440900",
  "patient_name": "John Doe"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440500",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "queue_date": "2026-06-26",
    "ticket_no": "A001",
    "queue_no": 1,
    "patient_id": "550e8400-e29b-41d4-a716-446655440900",
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "550e8400-e29b-41d4-a716-446655440501",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "service_id": "550e8400-e29b-41d4-a716-446655440200",
    "patient_id": "550e8400-e29b-41d4-a716-446655440900",
    "patient_name": "John Doe"
  }'
```

## `GET /api/v1/queues`

List queues by active tenant and branch.

### Query

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `branch_id` | query | string(uuid) | Operationally Yes | - | - | current live consumer sends this; branch context must exist |
| `status` | query | string | Optional | queue status values | - | queue status filter |
| `queue_date` | query | string(date) | Optional | - | current business date in usage | exact queue date |
| `service_id` | query | string(uuid) | Optional | - | - | filter by active journey service |

### Example request

`GET /api/v1/queues?branch_id=550e8400-e29b-41d4-a716-446655440100&status=waiting`

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440500",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "branch_id": "550e8400-e29b-41d4-a716-446655440100",
      "queue_date": "2026-06-26",
      "ticket_no": "A001",
      "queue_no": 1,
      "patient_name": "John Doe",
      "status": "waiting",
      "current_journey_id": "550e8400-e29b-41d4-a716-446655440501",
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/queues?branch_id=550e8400-e29b-41d4-a716-446655440100&status=waiting' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/queues/{id}`

Get queue by ID.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Queue ID |

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440500",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "queue_date": "2026-06-26",
    "ticket_no": "A001",
    "queue_no": 1,
    "patient_id": "550e8400-e29b-41d4-a716-446655440900",
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "550e8400-e29b-41d4-a716-446655440501",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/queues/550e8400-e29b-41d4-a716-446655440500' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `POST /api/v1/queues/{id}/transition`

Transition queue state.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Queue ID |

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `action` | body | string | Yes | `call`, `serve`, `complete`, `skip`, `cancel` | - | queue state action |

### Example request body

```json
{
  "action": "call"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440500",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "queue_date": "2026-06-26",
    "ticket_no": "A001",
    "queue_no": 1,
    "patient_name": "John Doe",
    "status": "calling",
    "current_journey_id": "550e8400-e29b-41d4-a716-446655440501",
    "created_at": 1761800000000,
    "updated_at": 1761800100000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/550e8400-e29b-41d4-a716-446655440500/transition' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "call"
  }'
```

## `POST /api/v1/queues/{id}/forward`

Forward queue to next service/counter without creating new queue master row.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Queue ID |

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `destination_service_id` | body | string(uuid) | Yes | - | - | next service |
| `destination_counter_id` | body | string(uuid) | Optional | - | - | next counter |

### Example request body

```json
{
  "destination_service_id": "550e8400-e29b-41d4-a716-446655440201",
  "destination_counter_id": "550e8400-e29b-41d4-a716-446655440301"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440500",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "queue_date": "2026-06-26",
    "ticket_no": "A001",
    "queue_no": 1,
    "patient_name": "John Doe",
    "status": "waiting",
    "current_journey_id": "550e8400-e29b-41d4-a716-446655440599",
    "created_at": 1761800000000,
    "updated_at": 1761800200000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/queues/550e8400-e29b-41d4-a716-446655440500/forward' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "destination_service_id": "550e8400-e29b-41d4-a716-446655440201",
    "destination_counter_id": "550e8400-e29b-41d4-a716-446655440301"
  }'
```

## `GET /api/v1/queues/{id}/visit-journeys`

Get readable visit history for one queue.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Queue ID |

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440700",
      "queue_id": "550e8400-e29b-41d4-a716-446655440500",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "event_type": "registration",
      "payload": "{\"service_id\":\"550e8400-e29b-41d4-a716-446655440200\"}",
      "created_at": 1761800000000
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440701",
      "queue_id": "550e8400-e29b-41d4-a716-446655440500",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "event_type": "call",
      "created_at": 1761800100000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/queues/550e8400-e29b-41d4-a716-446655440500/visit-journeys' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/branches/{id}/queue-stats`

Get queue stats for one branch and active business date.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |

### Example success response

```json
{
  "data": {
    "total_queues_today": 14,
    "total_active_journeys": 5,
    "total_completed_visits": 9,
    "waiting_by_service": {
      "550e8400-e29b-41d4-a716-446655440200": 3,
      "550e8400-e29b-41d4-a716-446655440201": 2
    }
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100/queue-stats' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/branches/{id}/services/{service_id}/queue-journeys`

List active queue journeys for one branch + service.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |
| `service_id` | string(uuid) | Yes | Service ID |

### Query

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `queue_date` | query | string(date) | Optional | - | current business date in usage | exact queue date |
| `status` | query | string | Optional | journey statuses | - | journey status filter |

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440501",
      "queue_id": "550e8400-e29b-41d4-a716-446655440500",
      "service_id": "550e8400-e29b-41d4-a716-446655440200",
      "seq_no": 1,
      "status": "pending",
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100/services/550e8400-e29b-41d4-a716-446655440200/queue-journeys?status=pending' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/branches/{id}/counters/{counter_id}/queue-journeys`

List active queue journeys for one branch + counter.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |
| `counter_id` | string(uuid) | Yes | Counter ID |

### Query

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `queue_date` | query | string(date) | Optional | - | current business date in usage | exact queue date |
| `status` | query | string | Optional | journey statuses | - | journey status filter |

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440599",
      "queue_id": "550e8400-e29b-41d4-a716-446655440500",
      "service_id": "550e8400-e29b-41d4-a716-446655440201",
      "counter_id": "550e8400-e29b-41d4-a716-446655440301",
      "seq_no": 2,
      "status": "pending",
      "created_at": 1761800200000,
      "updated_at": 1761800200000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100/counters/550e8400-e29b-41d4-a716-446655440301/queue-journeys?status=pending' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
