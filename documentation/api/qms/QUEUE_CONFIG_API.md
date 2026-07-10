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

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.tenant_id` | string | Yes | Tenant UUID |
| `data.branch_id` | string | No | Effective branch UUID when branch scope supplied |
| `data.service_id` | string | No | Effective service UUID when service scope supplied |
| `data.counter_id` | string | No | Effective counter UUID when counter scope supplied |
| `data.tenant.tenant_id` | string | Yes | Tenant UUID wrapper |
| `data.branch.branch_id` | string | No | Branch UUID in branch context |
| `data.branch.effective_logo_asset_id` | string | No | Resolved branch logo asset UUID |
| `data.queue.<field>.key` | string | Yes | Config field key |
| `data.queue.<field>.value` | string | Yes | Resolved value as string |
| `data.queue.<field>.source` | string | No | Source scope, e.g. `tenant`, `branch`, `counter` |
| `data.queue.<field>.inherited` | bool | No | Whether value is inherited |
| `data.queue.<field>.can_override` | bool | Yes | Whether caller may override field |
| `data.queue.<field>.can_reset` | bool | Yes | Whether caller may reset field |
| `data.queue_reset_time` | string | Yes | Flattened reset time |
| `data.queue_reset_time_source` | string | No | Source scope for flattened reset time |
| `data.queue_reset_time_inherited` | bool | No | Inheritance flag for flattened reset time |
| `data.ticket_prefix` | string | Yes | Flattened ticket prefix |
| `data.ticket_prefix_source` | string | No | Source scope for flattened prefix |
| `data.ticket_prefix_inherited` | bool | No | Inheritance flag for flattened prefix |
| `data.numbering_strategy` | string | Yes | Flattened numbering strategy |
| `data.numbering_strategy_source` | string | No | Source scope for flattened strategy |
| `data.numbering_strategy_inherited` | bool | No | Inheritance flag for flattened strategy |
| `data.default_estimated_duration` | string | No | Flattened default duration |
| `data.default_estimated_duration_source` | string | No | Source scope for flattened duration |
| `data.default_estimated_duration_inherited` | bool | No | Inheritance flag for flattened duration |
| `data.allow_forward` | bool | No | Forwarding toggle |
| `data.allow_skip` | bool | No | Skip toggle |
| `data.allow_recall` | bool | No | Recall toggle |
| `data.allow_cancel` | bool | No | Cancel toggle |
| `data.auto_call_next` | bool | No | Auto-call toggle |
| `data.max_service_duration` | int | No | Maximum service duration in minutes |
| `data.min_service_duration` | int | No | Minimum service duration in minutes |
| `data.require_counter` | bool | No | Branch-service destination counter requirement |
| `data.allow_forward_from` | bool | No | Branch-service source forwarding toggle |
| `data.allow_forward_to` | bool | No | Branch-service destination forwarding toggle |
| `data.audio_id` | string | No | ID audio asset UUID |
| `data.audio_en` | string | No | EN audio asset UUID |
| `data.narrative_instruction_id` | string | No | ID narrative instruction asset UUID |
| `data.narrative_instruction_en` | string | No | EN narrative instruction asset UUID |
| `data.effective_until` | string | No | Expiration or validity marker if configured |

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
| Field | Type | Required | Notes |
|---|---|---|---|
| `queue_reset_time` | string(HH:MM) | No | Daily reset time |
| `ticket_prefix` | string | No | Ticket prefix |
| `numbering_strategy` | string | No | Queue numbering strategy |
| `allow_forward` | bool | No | Enable forwarding |
| `allow_skip` | bool | No | Enable skip action |
| `allow_recall` | bool | No | Enable recall action |
| `allow_cancel` | bool | No | Enable cancel action |
| `auto_call_next` | bool | No | Auto-call next queue |
| `default_estimated_duration` | int | No | Estimated duration in minutes |
| `max_service_duration` | int | No | Maximum service duration in minutes |
| `min_service_duration` | int | No | Minimum service duration in minutes |
| `require_counter` | bool | No | Branch-service only |
| `allow_forward_from` | bool | No | Branch-service only |
| `allow_forward_to` | bool | No | Branch-service only |
| `audio_id` | string | No | ID audio asset UUID |
| `audio_en` | string | No | EN audio asset UUID |
| `narrative_instruction_id` | string | No | ID instruction asset UUID |
| `narrative_instruction_en` | string | No | EN instruction asset UUID |

### Example Request Body
```json
{
  "queue_reset_time": "04:00",
  "ticket_prefix": "A",
  "numbering_strategy": "daily_branch_sequence",
  "allow_forward": true,
  "allow_skip": true,
  "allow_recall": true,
  "allow_cancel": true,
  "auto_call_next": false,
  "default_estimated_duration": 5,
  "max_service_duration": 30,
  "min_service_duration": 1,
  "audio_id": "audio-id-uuid",
  "audio_en": "audio-en-uuid",
  "narrative_instruction_id": "instruction-id-uuid",
  "narrative_instruction_en": "instruction-en-uuid"
}
```

### Response
Status: `204 No Content`

### Example Request
```bash
curl -X PATCH 'http://127.0.0.1:8080/api/v1/queue-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000' \
  -H 'Content-Type: application/json' \
  -d '{"ticket_prefix":"A"}'
```

## Branch-Level Endpoints

### `GET /api/v1/branches/:branch_id/effective-config`
Same response shape as tenant-level, but scoped to branch. Depends on `branch_id` path param.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/effective-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### `PATCH /api/v1/branches/:branch_id/queue-config`
Update branch-level override. Same body as tenant patch.

### Example Request
```bash
curl -X PATCH 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/queue-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid' \
  -H 'Content-Type: application/json' \
  -d '{"ticket_prefix":"B"}'
```

### `DELETE /api/v1/branches/:branch_id/queue-config/:field`
Reset a single field to inherit from tenant.

**Response:** `204 No Content`

Valid fields: `queue_reset_time`, `ticket_prefix`, `numbering_strategy`, `allow_forward`, `allow_skip`, `allow_recall`, `allow_cancel`, `auto_call_next`, `default_estimated_duration`.

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/queue-config/ticket_prefix' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

## Branch-Service-Level Endpoints

### `GET /api/v1/branches/:branch_id/services/:service_id/effective-config`
Same shape, scoped to branch-service.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/services/service-uuid/effective-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### `PATCH /api/v1/branches/:branch_id/services/:branch_service_id/queue-config`
Update branch-service-level override.

Additional branch-service field: `require_counter`, `allow_forward_from`, `allow_forward_to`.

### Example Request
```bash
curl -X PATCH 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/services/bs-uuid/queue-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid' \
  -H 'Content-Type: application/json' \
  -d '{"require_counter":true}'
```

### `DELETE /api/v1/branches/:branch_id/services/:branch_service_id/queue-config/:field`
Reset single field.

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/services/bs-uuid/queue-config/require_counter' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

## Counter-Level Endpoints

### `GET /api/v1/branches/:branch_id/counters/:counter_id/effective-config`
Same shape, scoped to counter.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/counters/counter-uuid/effective-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

### `PATCH /api/v1/branches/:branch_id/counters/:counter_id/queue-config`
Update counter-level override.

### Example Request
```bash
curl -X PATCH 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/counters/counter-uuid/queue-config' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid' \
  -H 'Content-Type: application/json' \
  -d '{"allow_forward":false}'
```

### `DELETE /api/v1/branches/:branch_id/counters/:counter_id/queue-config/:field`
Reset counter-level field.

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/branches/branch-uuid/counters/counter-uuid/queue-config/allow_forward' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: tenant-uuid'
```

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
