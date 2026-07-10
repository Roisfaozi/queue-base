# User API Reference

Manages user profiles and identities.

## `GET /api/v1/users/me`
Gets the authenticated user's profile.

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
Same as `GET /me`.

## `PATCH /api/v1/users/me/avatar`
Sets avatar URL.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `avatar_url` | string | Yes | avatar URL |

## `GET /api/v1/users`
Lists users.

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

## `GET /api/v1/users/:id`
Retrieves any user by ID.

## `PATCH /api/v1/users/:id/status`
Changes a user's status.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `status` | string | Yes | `active`, `suspended`, `banned` |

## `DELETE /api/v1/users/:id`
Soft-deletes a user account.

### Response
- `204 No Content`
