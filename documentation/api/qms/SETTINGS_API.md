# Settings API Reference

## `POST /api/v1/settings`

Create setting.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `scope_type` | body | string | Yes | `tenant`, `branch`, `service`, `counter` | - | inheritance scope |
| `scope_id` | body | string(uuid) | Yes | - | - | tenant/branch/service/counter ID |
| `key` | body | string | Yes | - | - | config key |
| `value` | body | string | Yes | - | - | raw stored value |
| `value_type` | body | string | Optional | `string`, `number`, `boolean`, `json` | `string` | value semantic type |

### Example request body

```json
{
  "scope_type": "branch",
  "scope_id": "550e8400-e29b-41d4-a716-446655440100",
  "key": "queue_reset_time",
  "value": "04:00",
  "value_type": "string"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440400",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "scope_type": "branch",
    "scope_id": "550e8400-e29b-41d4-a716-446655440100",
    "key": "queue_reset_time",
    "value": "04:00",
    "value_type": "string",
    "is_active": true,
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/settings' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "scope_type": "branch",
    "scope_id": "550e8400-e29b-41d4-a716-446655440100",
    "key": "queue_reset_time",
    "value": "04:00",
    "value_type": "string"
  }'
```

## `GET /api/v1/settings/resolve`

Resolve setting by inheritance order: counter -> service -> branch -> tenant.

### Query

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `key` | query | string | Yes | - | - | target setting key |
| `branch_id` | query | string(uuid) | Optional | - | - | branch scope candidate |
| `service_id` | query | string(uuid) | Optional | - | - | service scope candidate |
| `counter_id` | query | string(uuid) | Optional | - | - | counter scope candidate |

### Example request

`GET /api/v1/settings/resolve?key=queue_reset_time&branch_id=550e8400-e29b-41d4-a716-446655440100`

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440400",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "scope_type": "branch",
    "scope_id": "550e8400-e29b-41d4-a716-446655440100",
    "key": "queue_reset_time",
    "value": "04:00",
    "value_type": "string",
    "is_active": true,
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/settings/resolve?key=queue_reset_time&branch_id=550e8400-e29b-41d4-a716-446655440100' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/settings/{id}`

Get one setting by ID.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Setting ID |

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440400",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "scope_type": "branch",
    "scope_id": "550e8400-e29b-41d4-a716-446655440100",
    "key": "queue_reset_time",
    "value": "04:00",
    "value_type": "string",
    "is_active": true,
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/settings/550e8400-e29b-41d4-a716-446655440400' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `PUT /api/v1/settings/{id}`

Update setting value or active flag.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `value` | body | string | Optional | - | - | new value |
| `is_active` | body | boolean | Optional | - | - | active flag |

### Example request body

```json
{
  "value": "05:00",
  "is_active": true
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440400",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "scope_type": "branch",
    "scope_id": "550e8400-e29b-41d4-a716-446655440100",
    "key": "queue_reset_time",
    "value": "05:00",
    "value_type": "string",
    "is_active": true,
    "created_at": 1761800000000,
    "updated_at": 1761800500000
  }
}
```

### Curl

```bash
curl -X PUT 'http://127.0.0.1:8080/api/v1/settings/550e8400-e29b-41d4-a716-446655440400' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "value": "05:00",
    "is_active": true
  }'
```

## `DELETE /api/v1/settings/{id}`

Delete setting.

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/settings/550e8400-e29b-41d4-a716-446655440400' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
