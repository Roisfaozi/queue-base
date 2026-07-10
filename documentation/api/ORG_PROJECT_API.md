# Organization & Project API Reference

## Organizations

### Public
- `POST /api/v1/organizations/invitations/accept`
  - request: `token` (required), `password` (conditionally required), `name` (no)

### Authenticated
- `POST /api/v1/organizations` — create organization
- `GET /api/v1/organizations/me` — list organizations current user belongs to

### Tenant Context
Requires `X-Organization-ID` or `X-Organization-Slug`.

- `GET /api/v1/tenant/profile`
- `PATCH /api/v1/tenant/profile`
- `GET /api/v1/organizations/:id`
- `GET /api/v1/organizations/slug/:slug`
- `PUT /api/v1/organizations/:id`
- `DELETE /api/v1/organizations/:id`

### Member Management
- `POST /api/v1/organizations/:id/members/invite`
  - request: `email` (yes), `role_id` (yes)
- `GET /api/v1/organizations/:id/members`
- `PATCH /api/v1/organizations/:id/members/:userId`
  - request fields: `role_id` (no), `status` (no, valid values: active/suspended)
- `DELETE /api/v1/organizations/:id/members/:userId`
- `GET /api/v1/organizations/:id/presence`

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
