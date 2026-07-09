# User API Reference

Manages user profiles and identities.

## Authenticated Routes (Own Profile)

Requires `Authorization: Bearer <token>`.

### `GET /api/v1/users/me`
Gets the authenticated user's profile.

**Response:**
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

### `PUT /api/v1/users/me`
Updates profile information.

**Request Body:**
```json
{
  "name": "John Doe Updated",
  "username": "johndoe2"
}
```

### `PATCH /api/v1/users/me/avatar`
Sets avatar URL (usually after a TUS upload).

## Authorized Routes (Admin)

Requires `role:admin` or higher via Casbin.

### `GET /api/v1/users`
Lists users with optional pagination.

### `POST /api/v1/users/search`
Dynamic query-builder search for users.

### `GET /api/v1/users/:id`
Retrieves any user by ID.

### `PATCH /api/v1/users/:id/status`
Changes a user's status.

**Request Body:**
```json
{
  "status": "suspended"
}
```

### `DELETE /api/v1/users/:id`
Soft-deletes a user account.
