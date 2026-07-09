# Role & Permission API Reference

Covers RBAC role CRUD, permission grants, access-rights, and inheritance.

## Roles

### `POST /api/v1/roles`
Create role.

**Request:**
```json
{ "name": "role:editor" }
```

### `GET /api/v1/roles`
List roles.

### `GET /api/v1/roles/:id`
Get role by ID.

### `PUT /api/v1/roles/:id`
Update role.

### `DELETE /api/v1/roles/:id`
Delete role.

## Permissions

### `POST /api/v1/permissions/grant`
Grant permission to role.

**Request:**
```json
{ "role": "role:editor", "path": "/api/v1/users", "method": "GET", "domain": "global" }
```

### `POST /api/v1/permissions/assign-role`
Assign role to user.

**Request:**
```json
{ "user_id": "user-uuid", "role": "role:admin", "domain": "acme-corp" }
```

### `DELETE /api/v1/permissions/revoke`
Revoke permission.

### `POST /api/v1/permissions/check-batch`
Batch check permissions.

**Response:**
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

### `GET /api/v1/permissions/inheritance-tree`
Returns role inheritance tree.

### `POST /api/v1/permissions/inheritance`
Add parent role inheritance.

### `DELETE /api/v1/permissions/inheritance`
Remove parent role inheritance.

### `GET /api/v1/permissions/resources`
Aggregated resources view.

## Access Rights

### `POST /api/v1/access-rights`
Create access right.

### `GET /api/v1/access-rights`
List access rights.

### `GET /api/v1/access-rights/:id`
Get access right.

### `PUT /api/v1/access-rights/:id`
Update access right.

### `DELETE /api/v1/access-rights/:id`
Delete access right.
