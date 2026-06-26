# QMS API Reference

## Purpose

Dokumen ini adalah referensi API QMS lengkap berdasarkan live route dan model runtime saat ini.

Isi dokumen:

- daftar endpoint
- auth dan header requirement
- path, query, dan body attributes
- required / optional
- enum values
- default values jika ada
- contoh request
- contoh response
- contoh `curl`

## Base URL

Semua endpoint QMS ada di bawah prefix:

`/api/v1`

## Common Security Requirement

Semua endpoint QMS aktif berada di route group tenant-authorized.

Secara operasional request normal perlu:

- `Authorization: Bearer <access_token>`
- `X-Organization-ID: <tenant_uuid>`

Boundary runtime yang juga berlaku:

- active user session
- active user status
- auto API scope requirement
- Casbin authorization

Scanner endpoint punya tambahan header sendiri.

## Common Response Envelope

### Success

```json
{
  "data": {}
}
```

### Error

```json
{
  "message": "failed to create queue",
  "error": "bad request"
}
```

## Shared Enums

### Branch / Service / Counter status

- `active`
- `inactive`

### Setting scope_type

- `tenant`
- `branch`
- `service`
- `counter`

### Setting value_type

- `string`
- `number`
- `boolean`
- `json`

### Queue action

- `call`
- `serve`
- `complete`
- `skip`
- `cancel`

### Queue status

- `waiting`
- `calling`
- `serving`
- `skipped`
- `canceled`
- `completed`

### Scanner action

- `register`
- `forward`

---

## Branch APIs

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

---

## Service APIs

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

---

## Counter APIs

## `POST /api/v1/counters`

Create counter under one branch.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `branch_id` | body | string(uuid) | Yes | - | - | target branch |
| `code` | body | string | Yes | - | - | 2..50 chars, uppercased |
| `name` | body | string | Yes | - | - | 3..255 chars |

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

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/counters/550e8400-e29b-41d4-a716-446655440300' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

---

## Settings APIs

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

---

## Queue APIs

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

---

## Scanner API

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

---

## Common Error Notes

### `400 Bad Request`

Biasanya terjadi bila:

- body tidak valid
- field enum salah
- branch scope tidak ada
- transition action tidak valid untuk status saat ini
- header scanner kurang

### `401 Unauthorized`

Biasanya terjadi bila:

- bearer token tidak valid
- session tidak aktif
- scanner client/API key gagal diautentikasi

### `403 Forbidden`

Biasanya terjadi bila:

- user tidak lolos Casbin/scope
- relation validator menolak akses lintas tenant/branch/service/counter

### `404 Not Found`

Biasanya terjadi bila:

- entity target tidak ada dalam tenant scope
- current journey queue tidak ditemukan
- setting/key tidak ditemukan pada inheritance chain resolve
