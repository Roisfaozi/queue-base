# Counter API Reference

## `POST /api/v1/counters`

Create counter under one branch.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `branch_id` | body | string(uuid) | Yes | - | - | target branch |
| `code` | body | string | Yes | - | - | 2..50 chars, uppercased |
| `name` | body | string | Yes | - | - | 3..255 chars |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Counter UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.branch_id` | string | Yes | Branch UUID |
| `data.code` | string | Yes | Counter code |
| `data.name` | string | Yes | Counter name |
| `data.status` | string | Yes | Counter status |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example request body

```json
{
  "branch_id": "550e8400-e29b-41d4-a716-446655440100",
  "code": "C1",
  "name": "Counter 1"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440300",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "code": "C1",
    "name": "Counter 1",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/counters' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "code": "C1",
    "name": "Counter 1"
  }'
```

## `GET /api/v1/counters`

List counters.

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Counter UUID |
| `data[].tenant_id` | string | Yes | Tenant UUID |
| `data[].branch_id` | string | Yes | Branch UUID |
| `data[].code` | string | Yes | Counter code |
| `data[].name` | string | Yes | Counter name |
| `data[].status` | string | Yes | Counter status |
| `data[].created_at` | int64 | Yes | Creation timestamp |
| `data[].updated_at` | int64 | Yes | Update timestamp |

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440300",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "branch_id": "550e8400-e29b-41d4-a716-446655440100",
      "code": "C1",
      "name": "Counter 1",
      "status": "active",
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/counters' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/counters/{id}`

Get counter by ID.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Counter ID |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Counter UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.branch_id` | string | Yes | Branch UUID |
| `data.code` | string | Yes | Counter code |
| `data.name` | string | Yes | Counter name |
| `data.status` | string | Yes | Counter status |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440300",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "code": "C1",
    "name": "Counter 1",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/counters/550e8400-e29b-41d4-a716-446655440300' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `PUT /api/v1/counters/{id}`

Update counter.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `code` | body | string | Optional | - | - |
| `name` | body | string | Optional | - | - |
| `status` | body | string | Optional | `active`, `inactive` | - |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Counter UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.branch_id` | string | Yes | Branch UUID |
| `data.code` | string | Yes | Counter code |
| `data.name` | string | Yes | Counter name |
| `data.status` | string | Yes | Counter status |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example request body

```json
{
  "name": "Counter Utama",
  "status": "active"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440300",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440100",
    "code": "C1",
    "name": "Counter Utama",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800500000
  }
}
```

### Curl

```bash
curl -X PUT 'http://127.0.0.1:8080/api/v1/counters/550e8400-e29b-41d4-a716-446655440300' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Counter Utama",
    "status": "active"
  }'
```

## `DELETE /api/v1/counters/{id}`

Delete counter.

### Response Fields
None (204 No Content).

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/counters/550e8400-e29b-41d4-a716-446655440300' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
