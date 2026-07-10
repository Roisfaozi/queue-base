# Organization & Project API Reference

## Organizations

### Public
- `POST /api/v1/organizations/invitations/accept`
  - request: `token` (required), `password` (conditionally required), `name` (no)

#### Example Request
```json
{
  "token": "invite-token-uuid",
  "password": "Password123!",
  "name": "John Doe"
}
```

### Authenticated
- `POST /api/v1/organizations` — create organization
- `GET /api/v1/organizations/me` — list organizations current user belongs to

#### Example Request `POST /api/v1/organizations`
```json
{
  "name": "Acme Corp",
  "slug": "acme-corp"
}
```

#### Example Request `GET /api/v1/organizations/me`
```bash
curl 'http://127.0.0.1:8080/api/v1/organizations/me' \
  -H 'Authorization: Bearer <token>'
```

### Tenant Context
Requires `X-Organization-ID` or `X-Organization-Slug`.

- `GET /api/v1/tenant/profile`
- `PATCH /api/v1/tenant/profile`
- `GET /api/v1/organizations/:id`
- `GET /api/v1/organizations/slug/:slug`
- `PUT /api/v1/organizations/:id`
- `DELETE /api/v1/organizations/:id`

#### Example Request `PATCH /api/v1/tenant/profile`
```json
{
  "name": "Acme Corp Updated",
  "timezone": "Asia/Jakarta",
  "logo_asset_id": "asset-uuid"
}
```

### Member Management
- `POST /api/v1/organizations/:id/members/invite`
  - request: `email` (yes), `role_id` (yes)
- `GET /api/v1/organizations/:id/members`
- `PATCH /api/v1/organizations/:id/members/:userId`
  - request fields: `role_id` (no), `status` (no, valid values: active/suspended)
- `DELETE /api/v1/organizations/:id/members/:userId`
- `GET /api/v1/organizations/:id/presence`

#### Example Request `POST /api/v1/organizations/:id/members/invite`
```json
{
  "email": "member@example.com",
  "role_id": "role-uuid"
}
```

#### Example Request `PATCH /api/v1/organizations/:id/members/:userId`
```json
{
  "role_id": "role-uuid",
  "status": "active"
}
```

### Response Fields (Organization)
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | org UUID |
| `data.name` | string | Yes | org name |
| `data.slug` | string | Yes | unique slug |
| `data.status` | string | Yes | active/suspended/draft |
| `data.owner_id` | string | Yes | org owner |
| `data.timezone` | string | No | default Asia/Jakarta |
| `data.logo_asset_id` | string | No | logo asset |

### Response Example
```json
{
  "data": {
    "id": "org-uuid",
    "name": "Acme Corp",
    "slug": "acme-corp",
    "status": "active",
    "owner_id": "user-uuid",
    "timezone": "Asia/Jakarta",
    "logo_asset_id": "asset-uuid"
  }
}
```

## Projects

Tenant-scoped CRUD.

- `POST /api/v1/projects`
  - request: `name` (yes), `domain` (yes)
- `GET /api/v1/projects`
- `GET /api/v1/projects/:id`
- `PUT /api/v1/projects/:id`
- `DELETE /api/v1/projects/:id`

#### Example Request `POST /api/v1/projects`
```json
{
  "name": "Landing Page",
  "domain": "landing.acme.test"
}
```

#### Example Request `PUT /api/v1/projects/:id`
```json
{
  "name": "Landing Page Updated",
  "domain": "landing-v2.acme.test",
  "status": "active"
}
```

### Response Fields (Project)
| Field | Type | Required | Notes |
|---|---|---|---|
| `data.id` | string | Yes | project UUID |
| `data.organization_id` | string | Yes | parent org |
| `data.user_id` | string | Yes | creator |
| `data.name` | string | Yes | project name |
| `data.domain` | string | Yes | project domain |
| `data.status` | string | Yes | active/inactive |
| `data.created_at` | integer | Yes | unix ms |
| `data.updated_at` | integer | Yes | unix ms |
