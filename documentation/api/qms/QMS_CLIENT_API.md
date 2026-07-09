# QMS Client API Reference

QMS Clients represent credential-bound devices or applications such as caller, signage, scanner, or kiosk.

## `POST /api/v1/qms-clients`
Creates a QMS client.

### Body
| Field | Type | Required | Notes |
|---|---|---|---|
| `branch_id` | string(uuid) | Yes | Branch binding |
| `branch_service_id` | string(uuid) | No | Service binding for caller/scanner |
| `counter_id` | string(uuid) | No | Counter binding for caller |
| `client_type` | string | Yes | `caller`, `signage`, `scanner`, `kiosk` |
| `name` | string | Yes | Human-readable client name |

### Example Success Response
```json
{
  "data": {
    "id": "client-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "branch-uuid",
    "client_type": "caller",
    "name": "Loket 1 Admission",
    "branch_service_id": "bs-uuid",
    "counter_id": "counter-uuid",
    "is_active": true,
    "created_at": 1761800000000
  }
}
```

## `POST /api/v1/qms-clients/credentials`
Creates new credentials for a QMS client.

### Body
| Field | Type | Required | Notes |
|---|---|---|---|
| `client_id` | string(uuid) | Yes | Existing QMS client ID |
| `api_key` | string | Yes | Plain secret used to derive hash |
| `expires_at` | integer(ms) | No | Credential expiration |

### Example Success Response
```json
{
  "data": {
    "id": "credential-uuid",
    "client_id": "client-uuid",
    "expires_at": 1761890000000,
    "created_at": 1761800000000
  }
}
```

### Security Notes
- raw secret is returned once by backend flow if configured
- only hash is persisted
- rotate credentials by creating a new credential and deactivating old client if needed

## `GET /api/v1/qms-clients`
Lists QMS clients in active tenant scope.

### Example Response
```json
{
  "data": [
    {
      "id": "client-uuid",
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "client_type": "caller",
      "name": "Loket 1 Admission",
      "branch_service_id": "bs-uuid",
      "counter_id": "counter-uuid",
      "is_active": true,
      "created_at": 1761800000000
    }
  ]
}
```

## `GET /api/v1/qms-clients/:id`
Gets QMS client by ID.

### Example Response
Same shape as list item.

## `PATCH /api/v1/qms-clients/:id`
Updates QMS client metadata and binding.

### Body
```json
{
  "name": "Loket 1 Frontdesk",
  "branch_service_id": "bs-uuid",
  "counter_id": "counter-uuid",
  "is_active": true
}
```

### Example Response
```json
{
  "data": {
    "id": "client-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "branch-uuid",
    "client_type": "caller",
    "name": "Loket 1 Frontdesk",
    "branch_service_id": "bs-uuid",
    "counter_id": "counter-uuid",
    "is_active": true,
    "created_at": 1761800000000
  }
}
```

## `DELETE /api/v1/qms-clients/:id`
Deletes or deactivates QMS client depending on backend policy.

### Response
- `204 No Content`
