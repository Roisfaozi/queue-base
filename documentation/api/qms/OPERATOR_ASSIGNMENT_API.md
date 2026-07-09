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

### Success Response
- `204 No Content`

## Security Notes
- Requires `operator_assignment:manage` scope.
- All writes are tenant-scoped.
- Runtime caller behavior must not trust frontend-only assignment checks.
