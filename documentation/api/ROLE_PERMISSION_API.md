# Role & Permission API Reference

Covers RBAC role CRUD, permission grants, access-rights, and inheritance.

## Roles

### `POST /api/v1/roles`
Create role.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | Yes | role slug like `role:editor` |

### `GET /api/v1/roles`
List roles.

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | role UUID |
| `data[].name` | string | Yes | role slug |
| `data[].description` | string | No | description if present |

### Permissions

#### `POST /api/v1/permissions/grant`

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `role` | string | Yes | role slug |
| `path` | string | Yes | protected resource path |
| `method` | string | Yes | action / HTTP method |
| `domain` | string | No | default/global domain fallback |

#### `POST /api/v1/permissions/assign-role`

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `user_id` | string | Yes | target user |
| `role` | string | Yes | assigned role |
| `domain` | string | No | tenant/domain |

#### `POST /api/v1/permissions/check-batch`

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `items` | array | Yes | batch items |
| `items[].resource` | string | Yes | resource path |
| `items[].action` | string | Yes | action |
| `items[].domain` | string | No | domain |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.results` | object | Yes | map of `resource:action` to boolean |

### Response Example
```json
{
  "data": {
    "results": {
      "/api/v1/users:GET": true,
      "/api/v1/projects:POST": false
    }
  }
}
```

#### Other Permission Routes
- `DELETE /api/v1/permissions/revoke`
- `DELETE /api/v1/permissions/revoke-role`
- `GET /api/v1/permissions`
- `GET /api/v1/permissions/:role`
- `GET /api/v1/permissions/roles/:role/users`
- `PUT /api/v1/permissions`
- `POST /api/v1/permissions/inheritance`
- `DELETE /api/v1/permissions/inheritance`
- `GET /api/v1/permissions/:role/parents`
- `GET /api/v1/permissions/resources`
- `GET /api/v1/permissions/inheritance-tree`
- `GET /api/v1/permissions/roles/:role/access-rights`
- `POST /api/v1/permissions/assign-access-right`
- `DELETE /api/v1/permissions/revoke-access-right`

## Access Rights

### `POST /api/v1/access-rights`
Create access right.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | Yes | logical access-right name |
| `resource` | string | Yes | target resource |
| `action` | string | Yes | target action |

### Response Fields
Depends on controller output; minimally includes created access-right identifier and metadata.
