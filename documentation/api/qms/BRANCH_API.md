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
| `address` | body | string | No | - | - | branch address |
| `city` | body | string | No | - | - | branch city |
| `province` | body | string | No | - | - | branch province |
| `phone` | body | string | No | - | - | branch phone |
| `running_text` | body | string | No | - | - | public running text |
| `timezone` | body | string | No | - | - | branch timezone |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Branch UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.code` | string | Yes | Branch code |
| `data.name` | string | Yes | Branch name |
| `data.address` | string | No | Branch address |
| `data.city` | string | No | Branch city |
| `data.province` | string | No | Branch province |
| `data.postal_code` | string | No | Branch postal code |
| `data.phone` | string | No | Branch phone |
| `data.email` | string | No | Branch email |
| `data.logo_asset_id` | string | No | Logo asset UUID |
| `data.running_text` | string | No | Public running text |
| `data.timezone` | string | No | Branch timezone |
| `data.status` | string | Yes | Branch status |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example request body

```json
{
  "code": "RJ",
  "name": "Rawat Jalan",
  "address": "Jl. Sudirman No. 10",
  "city": "Jakarta",
  "province": "DKI Jakarta",
  "phone": "+62-21-5550001",
  "running_text": "Selamat datang di Rawat Jalan",
  "timezone": "Asia/Jakarta"
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
    "address": "Jl. Sudirman No. 10",
    "city": "Jakarta",
    "province": "DKI Jakarta",
    "postal_code": "10220",
    "phone": "+62-21-5550001",
    "email": "rawatjalan@example.com",
    "logo_asset_id": "asset-uuid",
    "running_text": "Selamat datang di Rawat Jalan",
    "timezone": "Asia/Jakarta",
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

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Branch UUID |
| `data[].tenant_id` | string | Yes | Tenant UUID |
| `data[].code` | string | Yes | Branch code |
| `data[].name` | string | Yes | Branch name |
| `data[].address` | string | No | Branch address |
| `data[].city` | string | No | Branch city |
| `data[].province` | string | No | Branch province |
| `data[].postal_code` | string | No | Branch postal code |
| `data[].phone` | string | No | Branch phone |
| `data[].email` | string | No | Branch email |
| `data[].logo_asset_id` | string | No | Logo asset UUID |
| `data[].running_text` | string | No | Public running text |
| `data[].timezone` | string | No | Branch timezone |
| `data[].status` | string | Yes | Branch status |
| `data[].created_at` | int64 | Yes | Creation timestamp |
| `data[].updated_at` | int64 | Yes | Update timestamp |

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

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Branch UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.code` | string | Yes | Branch code |
| `data.name` | string | Yes | Branch name |
| `data.address` | string | No | Branch address |
| `data.city` | string | No | Branch city |
| `data.province` | string | No | Branch province |
| `data.postal_code` | string | No | Branch postal code |
| `data.phone` | string | No | Branch phone |
| `data.email` | string | No | Branch email |
| `data.logo_asset_id` | string | No | Logo asset UUID |
| `data.running_text` | string | No | Public running text |
| `data.timezone` | string | No | Branch timezone |
| `data.status` | string | Yes | Branch status |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

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
| `address` | body | string | Optional | - | - | branch address |
| `city` | body | string | Optional | - | - | branch city |
| `province` | body | string | Optional | - | - | branch province |
| `postal_code` | body | string | Optional | - | - | branch postal code |
| `phone` | body | string | Optional | - | - | branch phone |
| `email` | body | string | Optional | - | - | branch email |
| `logo_asset_id` | body | string | Optional | - | - | logo asset UUID |
| `running_text` | body | string | Optional | - | - | public running text |
| `timezone` | body | string | Optional | - | - | branch timezone |
| `status` | body | string | Optional | `active`, `inactive` | - | branch status |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Branch UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.code` | string | Yes | Branch code |
| `data.name` | string | Yes | Branch name |
| `data.address` | string | No | Branch address |
| `data.city` | string | No | Branch city |
| `data.province` | string | No | Branch province |
| `data.postal_code` | string | No | Branch postal code |
| `data.phone` | string | No | Branch phone |
| `data.email` | string | No | Branch email |
| `data.logo_asset_id` | string | No | Logo asset UUID |
| `data.running_text` | string | No | Public running text |
| `data.timezone` | string | No | Branch timezone |
| `data.status` | string | Yes | Branch status |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example request body

```json
{
  "code": "RJU",
  "name": "Rawat Jalan Utama",
  "address": "Jl. Sudirman No. 12",
  "city": "Jakarta Selatan",
  "province": "DKI Jakarta",
  "postal_code": "12190",
  "phone": "+62-21-5550002",
  "email": "rju@example.com",
  "logo_asset_id": "asset-uuid-v2",
  "running_text": "Mohon siapkan dokumen pendaftaran",
  "timezone": "Asia/Jakarta",
  "status": "active"
}
```

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440100",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "RJU",
    "name": "Rawat Jalan Utama",
    "address": "Jl. Sudirman No. 12",
    "city": "Jakarta Selatan",
    "province": "DKI Jakarta",
    "postal_code": "12190",
    "phone": "+62-21-5550002",
    "email": "rju@example.com",
    "logo_asset_id": "asset-uuid-v2",
    "running_text": "Mohon siapkan dokumen pendaftaran",
    "timezone": "Asia/Jakarta",
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

### Response Fields
None (204 No Content).

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/branches/550e8400-e29b-41d4-a716-446655440100' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
