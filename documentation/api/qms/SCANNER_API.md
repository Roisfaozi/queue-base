# Scanner API Reference

## `POST /api/v1/scanner/check-in`

Scanner entrypoint untuk action `register` atau `forward`.

### Extra headers

| Name | Required | Description |
|---|---|---|
| `X-Client-ID` | Yes | Scanner client identifier |
| `X-API-Key` | Yes | Scanner API key |

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `action` | body | string | Yes | `register`, `forward` | - | scanner operation |
| `branch_id` | body | string(uuid) | Yes | - | - | branch scope |
| `service_id` | body | string(uuid) | Conditional | - | - | required for `action=register` |
| `patient_id` | body | string(uuid) | Optional | - | - | patient UUID |
| `patient_name` | body | string | Conditional | - | - | required for `action=register` in practical flow |
| `queue_id` | body | string(uuid) | Conditional | - | - | required for `action=forward` |
| `destination_service_id` | body | string(uuid) | Conditional | - | - | required for `action=forward` |
| `destination_counter_id` | body | string(uuid) | Optional | - | - | optional destination counter |

### Example request body for register

```json
{
  "action": "register",
  "branch_id": "550e8400-e29b-41d4-a716-446655440100",
  "service_id": "550e8400-e29b-41d4-a716-446655440200",
  "patient_id": "550e8400-e29b-41d4-a716-446655440900",
  "patient_name": "John Doe"
}
```

### Example request body for forward

```json
{
  "action": "forward",
  "branch_id": "550e8400-e29b-41d4-a716-446655440100",
  "queue_id": "550e8400-e29b-41d4-a716-446655440500",
  "destination_service_id": "550e8400-e29b-41d4-a716-446655440201",
  "destination_counter_id": "550e8400-e29b-41d4-a716-446655440301"
}
```

### Example success response for register

```json
{
  "data": {
    "action": "register",
    "queue": {
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
}
```

### Example success response for forward

```json
{
  "data": {
    "action": "forward",
    "queue": {
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
}
```

### Curl register

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/scanner/check-in' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'X-Client-ID: scanner-device-01' \
  -H 'X-API-Key: <scanner_api_key>' \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "register",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "service_id": "550e8400-e29b-41d4-a716-446655440200",
    "patient_id": "550e8400-e29b-41d4-a716-446655440900",
    "patient_name": "John Doe"
  }'
```

### Curl forward

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/scanner/check-in' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'X-Client-ID: scanner-device-01' \
  -H 'X-API-Key: <scanner_api_key>' \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "forward",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "queue_id": "550e8400-e29b-41d4-a716-446655440500",
    "destination_service_id": "550e8400-e29b-41d4-a716-446655440201",
    "destination_counter_id": "550e8400-e29b-41d4-a716-446655440301"
  }'
```
