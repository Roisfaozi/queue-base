# User API Reference

Manages user profiles and identities.

## `GET /api/v1/users/me`
Gets the authenticated user's profile.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/users/me' \
  -H 'Authorization: Bearer <access_token>'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | user UUID |
| `data.name` | string | Yes | full name |
| `data.username` | string | Yes | username |
| `data.email` | string | Yes | email |
| `data.avatar_url` | string | No | avatar URL |
| `data.status` | string | No | active/suspended/banned |
| `data.created_at` | integer | No | unix ms |
| `data.updated_at` | integer | No | unix ms |

### Response Example
```json
{
  "data": {
    "id": "user-uuid",
    "name": "John Doe",
    "username": "johndoe",
    "email": "john@example.com",
    "avatar_url": "https://example.com/avatar.jpg",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```

## `PUT /api/v1/users/me`
Updates profile information.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | No | new display name |
| `username` | string | Yes | new username |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | user UUID |
| `data.name` | string | Yes | full name |
| `data.username` | string | Yes | username |
| `data.email` | string | Yes | email |
| `data.avatar_url` | string | No | avatar URL |
| `data.status` | string | No | active/suspended/banned |
| `data.created_at` | integer | No | unix ms |
| `data.updated_at` | integer | No | unix ms |

### Example Request
```json
{
  "name": "John Doe",
  "username": "johndoe"
}
```

## `PATCH /api/v1/users/me/avatar`
Sets avatar URL.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `avatar_url` | string | Yes | avatar URL |

### Example Request
```json
{
  "avatar_url": "https://example.com/avatar.jpg"
}
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | user UUID |
| `data.name` | string | Yes | full name |
| `data.username` | string | Yes | username |
| `data.email` | string | Yes | email |
| `data.avatar_url` | string | No | avatar URL |
| `data.status` | string | No | active/suspended/banned |
| `data.created_at` | integer | No | unix ms |
| `data.updated_at` | integer | No | unix ms |

## `GET /api/v1/users`
Lists users.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/users' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: <organization_id>'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | user UUID |
| `data[].name` | string | Yes | full name |
| `data[].username` | string | Yes | username |
| `data[].email` | string | Yes | email |
| `data[].avatar_url` | string | No | avatar URL |
| `data[].status` | string | No | user status |
| `data[].created_at` | integer | No | unix ms |
| `data[].updated_at` | integer | No | unix ms |

## `POST /api/v1/users/search`
Dynamic query-builder search for users. Request/response are paginated and depend on filters.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `page` | integer | No | default 1 |
| `limit` | integer | No | 1..100 |
| `username` | string | No | filter |
| `email` | string | No | filter |

### Example Request
```json
{
  "page": 1,
  "limit": 20,
  "username": "johndoe",
  "email": "john@example.com"
}
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data[].id` | string | Yes | user UUID |
| `data[].name` | string | Yes | full name |
| `data[].username` | string | Yes | username |
| `data[].email` | string | Yes | email |
| `data[].avatar_url` | string | No | avatar URL |
| `data[].status` | string | No | user status |
| `data[].created_at` | integer | No | unix ms |
| `data[].updated_at` | integer | No | unix ms |

## `GET /api/v1/users/:id`
Retrieves any user by ID.

### Example Request
```bash
curl 'http://127.0.0.1:8080/api/v1/users/user-uuid' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: <organization_id>'
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | user UUID |
| `data.name` | string | Yes | full name |
| `data.username` | string | Yes | username |
| `data.email` | string | Yes | email |
| `data.avatar_url` | string | No | avatar URL |
| `data.status` | string | No | active/suspended/banned |
| `data.created_at` | integer | No | unix ms |
| `data.updated_at` | integer | No | unix ms |

## `PATCH /api/v1/users/:id/status`
Changes a user's status.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `status` | string | Yes | `active`, `suspended`, `banned` |

### Example Request
```json
{
  "status": "suspended"
}
```

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | user UUID |
| `data.name` | string | Yes | full name |
| `data.username` | string | Yes | username |
| `data.email` | string | Yes | email |
| `data.avatar_url` | string | No | avatar URL |
| `data.status` | string | Yes | updated status |
| `data.created_at` | integer | No | unix ms |
| `data.updated_at` | integer | No | unix ms |

## `DELETE /api/v1/users/:id`
Soft-deletes a user account.

### Response
- `204 No Content`

### Example Request
```bash
curl -X DELETE 'http://127.0.0.1:8080/api/v1/users/user-uuid' \
  -H 'Authorization: Bearer <access_token>' \
  -H 'X-Organization-ID: <organization_id>'
```
