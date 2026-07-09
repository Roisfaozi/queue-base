# Organization & Project API Reference

## Organizations

### Public
- `POST /api/v1/organizations/invitations/accept`

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
- `GET /api/v1/organizations/:id/members`
- `PATCH /api/v1/organizations/:id/members/:userId`
- `DELETE /api/v1/organizations/:id/members/:userId`
- `GET /api/v1/organizations/:id/presence`

### Organization Response Example
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
- `GET /api/v1/projects`
- `GET /api/v1/projects/:id`
- `PUT /api/v1/projects/:id`
- `DELETE /api/v1/projects/:id`

### Project Response Example
```json
{
  "data": {
    "id": "project-uuid",
    "organization_id": "org-uuid",
    "user_id": "user-uuid",
    "name": "Customer Portal",
    "domain": "portal.example.com",
    "status": "active",
    "created_at": 1761800000000,
    "updated_at": 1761800000000
  }
}
```
