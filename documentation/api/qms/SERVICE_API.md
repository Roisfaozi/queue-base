# Service API Reference

## `POST /api/v1/services`

Create service.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `code` | body | string | Yes | - | - | 2..50 chars, uppercased |
| `name` | body | string | Yes | - | - | 3..255 chars |
| `is_pharmacy` | body | boolean | Optional | - | `false` | pharmacy flag |
| `is_pharmacy_reception` | body | boolean | Optional | - | `false` | pharmacy reception flag |

### Example request body

```json
{
  "code": "GEN",
  "name": "General Clinic",
  "is_pharmacy": false,
  "is_pharmacy_reception": false
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440200",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "GEN",
    "name": "General Clinic",
    "status": "active",
    "is_pharmacy": false,
    "is_pharmacy_reception": false,
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl -X POST 'http://127.0.0.1:8080/api/v1/services' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "code": "GEN",
    "name": "General Clinic",
    "is_pharmacy": false,
    "is_pharmacy_reception": false
  }'
```

## `GET /api/v1/services`

List services in active tenant.

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440200",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "GEN",
      "name": "General Clinic",
      "status": "active",
      "is_pharmacy": false,
      "is_pharmacy_reception": false,
      "created_at": 1761800000000,
      "updated_at": 1761800000000
    }
  ]
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/services' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `GET /api/v1/services/{id}`

Get service by ID.

### Path params

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Service ID |

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440200",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "GEN",
    "name": "General Clinic",
    "status": "active",
    "is_pharmacy": false,
    "is_pharmacy_reception": false,
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

### Curl

```bash
curl 'http://127.0.0.1:8080/api/v1/services/550e8400-e29b-41d4-a716-446655440200' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `PUT /api/v1/services/{id}`

Update service.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `code` | body | string | Optional | - | - |
| `name` | body | string | Optional | - | - |
| `status` | body | string | Optional | `active`, `inactive` | - |
| `is_pharmacy` | body | boolean | Optional | - | - |
| `is_pharmacy_reception` | body | boolean | Optional | - | - |

### Example request body

```json
{
  "status": "inactive",
  "is_pharmacy": true
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440200",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "GEN",
    "name": "General Clinic",
    "status": "inactive",
    "is_pharmacy": true,
    "is_pharmacy_reception": false,
    "created_at": 1761800000000,
    "updated_at": 1761800500000
  }
}
```

### Curl

```bash
curl -X PUT 'http://127.0.0.1:8080/api/v1/services/550e8400-e29b-41d4-a716-446655440200' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "inactive",
    "is_pharmacy": true
  }'
```

## `DELETE /api/v1/services/{id}`

Delete service.

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/services/550e8400-e29b-41d4-a716-446655440200' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
