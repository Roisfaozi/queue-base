# Queue Config API Reference

Typed hierarchical configuration for queue management.

Scope resolution: `Counter` -> `Branch Service` -> `Branch` -> `Tenant` -> `System Default`.

## `GET /api/v1/queue-config/effective`

Get effective tenant-level queue configuration.

### Query Parameters
| Field | Type | Required | Notes |
|---|---|---|---|
| `branch_id` | string(uuid) | No | Filter effective config for a branch scope |
| `service_id` | string(uuid) | No | Further narrow to service scope |
| `counter_id` | string(uuid) | No | Further narrow to counter scope |

### Example Success Response
```json
{
  "data": {
    "tenant_id": "org-uuid",
    "tenant": { "tenant_id": "org-uuid" },
    "branch": {},
    "queue": {
      "queue_reset_time": {
        "key": "queue_reset_time",
        "value": "04:00",
        "source": "tenant",
        "inherited": false,
        "can_override": false,
        "can_reset": false
      },
      "ticket_prefix": {
        "key": "ticket_prefix",
        "value": "A",
        "source": "tenant",
        "inherited": false,
        "can_override": false,
        "can_reset": false
      },
      "numbering_strategy": {
        "key": "numbering_strategy",
        "value": "daily_branch_sequence",
        "source": "tenant",
        "inherited": false,
        "can_override": false,
        "can_reset": false
      }
    },
    "queue_reset_time": "04:00",
    "queue_reset_time_source": "tenant",
    "queue_reset_time_inherited": false,
    "ticket_prefix": "A",
    "ticket_prefix_source": "tenant",
    "ticket_prefix_inherited": false,
    "numbering_strategy": "daily_branch_sequence",
    "numbering_strategy_source": "tenant",
    "numbering_strategy_inherited": false
  }
}
```

### Curl
```bash
curl 'http://127.0.0.1:8080/api/v1/queue-config/effective?branch_id=550e8400-e29b-41d4-a716-446655440100' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000'
```

## `PATCH /api/v1/queue-config`

Update tenant-level queue configuration.

### Body (partial update)
```json
{
  "queue_reset_time": "04:00",
  "ticket_prefix": "A",
  "numbering_strategy": "daily_branch_sequence",
  "allow_forward": true,
  "allow_skip": true,
  "allow_recall": true,
  "allow_cancel": true,
  "default_estimated_duration": 5
}
```

### Response
Status: `204 No Content`

## Branch-Level Endpoints

### `GET /api/v1/branches/:branch_id/effective-config`
Same response shape as tenant-level, but scoped to branch. Depends on `branch_id` path param.

### `PATCH /api/v1/branches/:branch_id/queue-config`
Update branch-level override. Same body as tenant patch.

### `DELETE /api/v1/branches/:branch_id/queue-config/:field`
Reset a single field to inherit from tenant.

**Response:** `204 No Content`

Valid fields: `queue_reset_time`, `ticket_prefix`, `numbering_strategy`, `allow_forward`, `allow_skip`, `allow_recall`, `allow_cancel`, `auto_call_next`, `default_estimated_duration`.

## Branch-Service-Level Endpoints

### `GET /api/v1/branches/:branch_id/services/:service_id/effective-config`
Same shape, scoped to branch-service.

### `PATCH /api/v1/branches/:branch_id/services/:branch_service_id/queue-config`
Update branch-service-level override.

Additional branch-service field: `require_counter`, `allow_forward_from`, `allow_forward_to`.

### `DELETE /api/v1/branches/:branch_id/services/:branch_service_id/queue-config/:field`
Reset single field.

## Counter-Level Endpoints

### `GET /api/v1/branches/:branch_id/counters/:counter_id/effective-config`
Same shape, scoped to counter.

### `PATCH /api/v1/branches/:branch_id/counters/:counter_id/queue-config`
Update counter-level override.

### `DELETE /api/v1/branches/:branch_id/counters/:counter_id/queue-config/:field`
Reset counter-level field.

## Field Reference

| Field | Type | Scopes | Description |
|---|---|---|---|
| `queue_reset_time` | string(HH:MM) | tenant, branch, counter | Daily queue number reset time |
| `ticket_prefix` | string(1-10) | tenant, branch, counter | Ticket number prefix |
| `numbering_strategy` | string | tenant, branch, counter | `daily_branch_sequence`, `continuous`, etc |
| `default_estimated_duration` | int | tenant, branch, branch_service, counter | Default wait estimate in minutes |
| `allow_forward` | bool | tenant, branch, counter | Allow forwarding |
| `allow_skip` | bool | tenant, branch, branch_service, counter | Allow skipping |
| `allow_recall` | bool | tenant, branch, branch_service, counter | Allow recall |
| `allow_cancel` | bool | tenant, branch, branch_service, counter | Allow cancel |
| `auto_call_next` | bool | tenant, branch, branch_service, counter | Auto-call next on complete |
| `require_counter` | bool | branch_service | Require counter for destination |
| `allow_forward_from` | bool | branch_service | Allow forward from this service |
| `allow_forward_to` | bool | branch_service | Allow forward to this service |

## Scope Requirements
- `queue-config:view`: `GET` operations. Requires `X-Organization-ID` in header.
- `queue-config:manage`: `PATCH` and `DELETE` operations.
