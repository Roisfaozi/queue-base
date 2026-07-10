# Authentication API Reference

Handles user registration, login, token refresh, password reset, email verification, and SSO.

## Public Routes

### `POST /api/v1/auth/register`
Creates a new user and assigns default role in the global domain.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | Yes | 3..100 chars |
| `username` | string | Yes | 3..50 chars |
| `email` | string(email) | Yes | max 100 |
| `password` | string | Yes | 8..72 chars |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data` | object/null | No | handler may return empty payload or created user summary |

### Example Request
```json
{
  "name": "John Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "Password123!"
}
```

### `POST /api/v1/auth/login`
Authenticates a user and establishes a Redis-backed session.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `username` | string | Yes | login identifier |
| `password` | string | Yes | plain password |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.access_token` | string | Yes | bearer token |
| `data.token_type` | string | Yes | usually `Bearer` |
| `data.expires_in` | integer | Yes | seconds |
| `data.refresh_token` | string | Yes | refresh token |
| `data.expires_at` | string(datetime) | Yes | token expiry |
| `data.user.id` | string | Yes | user UUID |
| `data.user.name` | string | Yes | display name |
| `data.user.email` | string | Yes | email |
| `data.user.username` | string | Yes | username |
| `data.user.role` | string | Yes | effective role string |
| `data.user.avatar_url` | string | No | avatar URL |

### Response (200 OK)
```json
{
  "data": {
    "access_token": "jwt-token",
    "token_type": "Bearer",
    "expires_in": 3600,
    "refresh_token": "refresh-token-uuid",
    "expires_at": "2026-07-09T10:00:00Z",
    "user": {
      "id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "username": "johndoe",
      "role": "role:user",
      "avatar_url": ""
    }
  }
}
```

### `POST /api/v1/auth/refresh`
Issues a new access token using a valid refresh token.

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `refresh_token` | string | Yes | previously issued refresh token |

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.access_token` | string | Yes | new bearer token |
| `data.token_type` | string | Yes | usually `Bearer` |
| `data.expires_in` | integer | Yes | seconds |

### Response (200 OK)
```json
{
  "data": {
    "access_token": "new-jwt-token",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

### Password & Verification
- `POST /api/v1/auth/forgot-password`
  - request: `email` required
- `POST /api/v1/auth/reset-password`
  - request: `token`, `new_password` required
- `POST /api/v1/auth/verify-email`
  - request: `token` required
- `GET /api/v1/auth/sso/:provider`
- `GET /api/v1/auth/sso/:provider/callback`

## Authenticated Routes
Requires `Authorization: Bearer <token>`.

### `GET /api/v1/auth/me`
Gets current session context.

### Response Fields
Same shape as login response.

### `POST /api/v1/auth/logout`
Destroys the current Redis session.

### Response
- `204 No Content`

### `POST /api/v1/auth/ticket`
Exchanges session for a short-lived ticket.

### Response Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.ticket` | string | Yes | short-lived auth ticket |
| `data.expires_at` | string/datetime | No | expiry if returned by handler |
