# Service API Reference

## `POST /api/v1/services`

Create service.

### Body

| Field | In | Type | Required | Enum | Default | Notes |
|---|---|---|---|---|---|---|
| `code` | body | string | Yes | - | - | 2..50 chars, uppercased |
| `name` | body | string | Yes | - | - | 3..255 chars |
| `type` | body | string | No | - | - | service type label |
| `default_estimated_duration` | body | int | No | - | - | wait estimate in minutes |
| `audio_id` | body | string | No | - | - | audio asset UUID |
| `audio_en` | body | string | No | - | - | EN audio asset UUID |
| `narrative_instruction_id` | body | string | No | - | - | ID instruction asset UUID |
| `narrative_instruction_en` | body | string | No | - | - | EN instruction asset UUID |
| `is_pharmacy` | body | boolean | Optional | - | `false` | pharmacy flag |
| `is_pharmacy_reception` | body | boolean | Optional | - | `false` | pharmacy reception flag |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Service UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.code` | string | Yes | Service code |
| `data.name` | string | Yes | Service name |
| `data.type` | string | No | Service type |
| `data.default_estimated_duration` | int | Yes | Wait estimate minutes |
| `data.audio_id` | string | No | Audio asset UUID |
| `data.audio_en` | string | No | EN audio asset UUID |
| `data.narrative_instruction_id` | string | No | ID instruction asset UUID |
| `data.narrative_instruction_en` | string | No | EN instruction asset UUID |
| `data.status` | string | Yes | Service status |
| `data.is_pharmacy` | bool | Yes | Pharmacy service flag |
| `data.is_pharmacy_reception` | bool | Yes | Pharmacy reception flag |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example request body

```json
{
  "code": "GEN",
  "name": "General Clinic",
  "type": "outpatient",
  "default_estimated_duration": 5,
  "audio_id": "audio-id-uuid",
  "audio_en": "audio-en-uuid",
  "narrative_instruction_id": "instruction-id-uuid",
  "narrative_instruction_en": "instruction-en-uuid",
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
    "type": "outpatient",
    "default_estimated_duration": 5,
    "audio_id": "audio-id-uuid",
    "audio_en": "audio-en-uuid",
    "narrative_instruction_id": "instruction-id-uuid",
    "narrative_instruction_en": "instruction-en-uuid",
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
    "type": "outpatient",
    "default_estimated_duration": 5,
    "audio_id": "audio-id-uuid",
    "audio_en": "audio-en-uuid",
    "narrative_instruction_id": "instruction-id-uuid",
    "narrative_instruction_en": "instruction-en-uuid",
    "is_pharmacy": false,
    "is_pharmacy_reception": false
  }'
```

## `GET /api/v1/services`

List services in active tenant.

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Service UUID |
| `data[].tenant_id` | string | Yes | Tenant UUID |
| `data[].code` | string | Yes | Service code |
| `data[].name` | string | Yes | Service name |
| `data[].type` | string | No | Service type |
| `data[].default_estimated_duration` | int | Yes | Wait estimate minutes |
| `data[].audio_id` | string | No | Audio asset UUID |
| `data[].audio_en` | string | No | EN audio asset UUID |
| `data[].narrative_instruction_id` | string | No | ID instruction asset UUID |
| `data[].narrative_instruction_en` | string | No | EN instruction asset UUID |
| `data[].status` | string | Yes | Service status |
| `data[].is_pharmacy` | bool | Yes | Pharmacy service flag |
| `data[].is_pharmacy_reception` | bool | Yes | Pharmacy reception flag |
| `data[].created_at` | int64 | Yes | Creation timestamp |
| `data[].updated_at` | int64 | Yes | Update timestamp |

### Example success response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440200",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "GEN",
      "name": "General Clinic",
      "type": "outpatient",
      "default_estimated_duration": 5,
      "audio_id": "audio-id-uuid",
      "audio_en": "audio-en-uuid",
      "narrative_instruction_id": "instruction-id-uuid",
      "narrative_instruction_en": "instruction-en-uuid",
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

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Service UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.code` | string | Yes | Service code |
| `data.name` | string | Yes | Service name |
| `data.type` | string | No | Service type |
| `data.default_estimated_duration` | int | Yes | Wait estimate minutes |
| `data.audio_id` | string | No | Audio asset UUID |
| `data.audio_en` | string | No | EN audio asset UUID |
| `data.narrative_instruction_id` | string | No | ID instruction asset UUID |
| `data.narrative_instruction_en` | string | No | EN instruction asset UUID |
| `data.status` | string | Yes | Service status |
| `data.is_pharmacy` | bool | Yes | Pharmacy service flag |
| `data.is_pharmacy_reception` | bool | Yes | Pharmacy reception flag |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example success response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440200",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "GEN",
    "name": "General Clinic",
    "type": "outpatient",
    "default_estimated_duration": 5,
    "audio_id": "audio-id-uuid",
    "audio_en": "audio-en-uuid",
    "narrative_instruction_id": "instruction-id-uuid",
    "narrative_instruction_en": "instruction-en-uuid",
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
| `type` | body | string | Optional | - | - | service type label |
| `default_estimated_duration` | body | int | Optional | - | - | wait estimate in minutes |
| `audio_id` | body | string | Optional | - | - | audio asset UUID |
| `audio_en` | body | string | Optional | - | - | EN audio asset UUID |
| `narrative_instruction_id` | body | string | Optional | - | - | ID instruction asset UUID |
| `narrative_instruction_en` | body | string | Optional | - | - | EN instruction asset UUID |
| `status` | body | string | Optional | `active`, `inactive` | - |
| `is_pharmacy` | body | boolean | Optional | - | - |
| `is_pharmacy_reception` | body | boolean | Optional | - | - |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Service UUID |
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.code` | string | Yes | Service code |
| `data.name` | string | Yes | Service name |
| `data.type` | string | No | Service type |
| `data.default_estimated_duration` | int | Yes | Wait estimate minutes |
| `data.audio_id` | string | No | Audio asset UUID |
| `data.audio_en` | string | No | EN audio asset UUID |
| `data.narrative_instruction_id` | string | No | ID instruction asset UUID |
| `data.narrative_instruction_en` | string | No | EN instruction asset UUID |
| `data.status` | string | Yes | Service status |
| `data.is_pharmacy` | bool | Yes | Pharmacy service flag |
| `data.is_pharmacy_reception` | bool | Yes | Pharmacy reception flag |
| `data.created_at` | int64 | Yes | Creation timestamp |
| `data.updated_at` | int64 | Yes | Update timestamp |

### Example request body

```json
{
  "code": "GEN",
  "name": "General Clinic",
  "type": "outpatient",
  "default_estimated_duration": 10,
  "audio_id": "audio-id-uuid",
  "audio_en": "audio-en-uuid",
  "narrative_instruction_id": "instruction-id-uuid",
  "narrative_instruction_en": "instruction-en-uuid",
  "status": "inactive",
  "is_pharmacy": true,
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
    "type": "outpatient",
    "default_estimated_duration": 10,
    "audio_id": "audio-id-uuid",
    "audio_en": "audio-en-uuid",
    "narrative_instruction_id": "instruction-id-uuid",
    "narrative_instruction_en": "instruction-en-uuid",
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

### Response Fields
None (204 No Content).

### Example success response

Status: `204 No Content`

Response body: empty

### Curl

```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/services/550e8400-e29b-41d4-a716-446655440200' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```
