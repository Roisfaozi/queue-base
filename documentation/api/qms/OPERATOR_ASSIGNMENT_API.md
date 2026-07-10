# Operator Assignment API Reference

Operator assignments bind users to counters for operational QMS caller workflows.

## `POST /api/v1/operator-counter-assignments`
Creates a counter assignment for a user.

### Body
| Field | Type | Required | Notes |
|---|---|---|---|
| `branch_id` | string(uuid) | Yes | Branch scope |
| `user_id` | string(uuid) | Yes | Operator user |
| `counter_id` | string(uuid) | Yes | Counter to operate |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | Assignment UUID |
| `data.tenant_id` | string | No | Tenant UUID |
| `data.branch_id` | string | Yes | Branch UUID |
| `data.user_id` | string | Yes | User UUID |
| `data.counter_id` | string | Yes | Counter UUID |
| `data.assigned_at` | int64 | Yes | Assignment timestamp |
| `data.unassigned_at` | int64 | No | Unassignment timestamp |

### Example Request
```json
{
  "branch_id": "branch-uuid",
  "user_id": "user-uuid",
  "counter_id": "counter-uuid"
}
```

### Example Response
```json
{
  "data": {
    "id": "assignment-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "branch-uuid",
    "user_id": "user-uuid",
    "counter_id": "counter-uuid",
    "assigned_at": 1761800000000,
    "unassigned_at": 1761890000000
  }
}
```

## `GET /api/v1/operator-counter-assignments`
Lists current assignments in active tenant scope.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/operator-counter-assignments' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | Assignment UUID |
| `data[].tenant_id` | string | No | Tenant UUID |
| `data[].branch_id` | string | Yes | Branch UUID |
| `data[].user_id` | string | Yes | User UUID |
| `data[].counter_id` | string | Yes | Counter UUID |
| `data[].assigned_at` | int64 | Yes | Assignment timestamp |
| `data[].unassigned_at` | int64 | No | Unassignment timestamp |

### Example Response
```json
{
  "data": [
    {
      "id": "assignment-uuid",
      "tenant_id": "tenant-uuid",
      "branch_id": "branch-uuid",
      "user_id": "user-uuid",
      "counter_id": "counter-uuid",
      "assigned_at": 1761800000000,
      "unassigned_at": null
    }
  ]
}
```

## `DELETE /api/v1/operator-counter-assignments/:id`
Unassigns or deletes an operator assignment.

### Path Params
| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string(uuid) | Yes | Assignment ID |

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/operator-counter-assignments/assignment-uuid' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### Success Response
### Response Fields
None (204 No Content).

## Security Notes
- Requires `operator_assignment:manage` scope.
- All writes are tenant-scoped.
- Runtime caller behavior must not trust frontend-only assignment checks.
