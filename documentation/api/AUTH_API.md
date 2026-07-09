# Authentication API Reference

Handles user registration, login, token refresh, and SSO.

## Public Routes

### `POST /api/v1/auth/register`
Creates a new user and assigns default role in the global domain.

**Request Body:**
```json
{
  "name": "John Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "Password123!"
}
```

**Response (201 Created):**
Standard envelope with empty data or user info.

### `POST /api/v1/auth/login`
Authenticates a user and establishes a Redis-backed session.

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "Password123!"
}
```

**Response (200 OK):**
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

**Request Body:**
```json
{
  "refresh_token": "refresh-token-uuid"
}
```

**Response (200 OK):**
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
- `POST /api/v1/auth/forgot-password` (`{"email": ""}`)
- `POST /api/v1/auth/reset-password` (`{"token": "", "new_password": ""}`)
- `POST /api/v1/auth/verify-email` (`{"token": ""}`)

## Authenticated Routes

Requires `Authorization: Bearer <token>`.

### `GET /api/v1/auth/me`
Gets current session context.

### `POST /api/v1/auth/logout`
Destroys the current Redis session.

### `POST /api/v1/auth/ticket`
Exchanges session for a short-lived ticket (useful for WebSocket upgrades).
