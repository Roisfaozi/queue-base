# Branch API Reference

## `POST /api/v1/branches`

Create branch under active tenant.

### Headers

| Name | Required | Description |
|---|---|---|
| `Authorization` | Yes | Bearer access token |
| `X-Organization-ID` | Yes | Tenant/organization UUID |

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `code` | body | string | Yes | - | - | 2..50 chars, sanitized and uppercased |
| `name` | body | string | Yes | - | - | 3..255 chars, sanitized |

### Example request body

```json
{
  "code": "RJ",
  "name": "Rawat Jalan"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440100",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "RJ",
    "name": "Rawat Jalan",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/branches' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "code": "RJ",
    "name": "Rawat Jalan"
  }'
```

## `GET /api/v1/branches`

List all branches in active tenant.

### Query

Tidak ada query parameter.

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440100",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "RJ",
      "name": "Rawat Jalan",
      "status": "active",
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/branches' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/branches/{id}`

Get one branch by ID.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440100",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "RJ",
    "name": "Rawat Jalan",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `PUT /api/v1/branches/{id}`

Update branch.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `code` | body | string | Optional | - | - | 2..50 chars |
| `name` | body | string | Optional | - | - | 3..255 chars |
| `status` | body | string | Optional | `active`, `inactive` | - | branch status |

### Example request body

```json
{
  "name": "Rawat Jalan Utama",
  "status": "active"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440100",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "RJ",
    "name": "Rawat Jalan Utama",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800500000
  }
}
```

### Curl

```bash
curl -X PUT 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Rawat Jalan Utama",
    "status": "active"
  }'
```

## `DELETE /api/v1/branches/{id}`

Delete branch in active tenant scope.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Branch ID |

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
