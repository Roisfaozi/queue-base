# API Access Workflow

This guide details the authentication, authorization, and endpoint definitions within the application, implementing secure Role-Based Access Control (RBAC) with Casbin.

## 🔑 I. Authentication Flow

We use **JWT (JSON Web Tokens)** for stateless authentication, backed by a **Redis-based Session** for instant revocability.

1. **Login**: User sends credentials to `POST /api/v1/auth/login`.
2. **Access Token**: Server returns JWT + refresh token.
3. **Authenticated Requests**: Send `Authorization: Bearer <access_token>`.
4. **Refresh Token**: Call `POST /api/v1/auth/refresh`.
5. **Logout**: Call `POST /api/v1/auth/logout`.

Additional public auth flows:
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/verify-email`
- `GET /api/v1/auth/sso/:provider`
- `GET /api/v1/auth/sso/:provider/callback`

Authenticated utility auth flows:
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/resend-verification`
- `POST /api/v1/auth/ticket`

---

## 🛡️ II. Authorization (Casbin RBAC)

We use **Casbin** with a RESTful model `(Subject, Object, Action)`.

### Policy Structure

- **Subject**: `role:admin`, `role:user`, or specific `user_id`.
- **Object**: API Path (e.g., `/api/v1/users`).
- **Action**: HTTP Method (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`).

### Advanced Role Hierarchy

- Multiple inheritance supported.
- Safe cycle detection on inheritance-tree traversal.
- Transactional enforcer keeps database and Casbin policy in sync.

---

## 🚀 III. API Endpoints Definition

### 1. Global & Realtime Endpoints

| Method | Path               | Description        | Access        |
| :----- | :----------------- | :----------------- | :------------ |
| `GET`  | `/api/v1/docs/*any`| Swagger/OpenAPI UI | Public        |
| `GET`  | `/api/v1/health`   | Health Check       | Public        |
| `GET`  | `/api/v1/ws`       | WebSocket Endpoint | Authenticated |
| `GET`  | `/api/v1/events`   | SSE Stream         | Public        |

### 2. Authentication Module

| Method | Path                                   | Description               | Access          |
| :----- | :------------------------------------- | :------------------------ | :-------------- |
| `POST` | `/api/v1/auth/register`                | Register user             | Public          |
| `POST` | `/api/v1/auth/login`                   | User login                | Public          |
| `POST` | `/api/v1/auth/refresh`                 | Refresh token             | Public          |
| `POST` | `/api/v1/auth/forgot-password`         | Forgot password           | Public          |
| `POST` | `/api/v1/auth/reset-password`          | Reset password            | Public          |
| `POST` | `/api/v1/auth/verify-email`            | Verify email              | Public          |
| `GET`  | `/api/v1/auth/sso/:provider`           | Start SSO login           | Public          |
| `GET`  | `/api/v1/auth/sso/:provider/callback`  | SSO callback              | Public          |
| `POST` | `/api/v1/auth/logout`                  | User logout               | Authenticated   |
| `POST` | `/api/v1/auth/resend-verification`     | Resend verification email | Authenticated   |
| `GET`  | `/api/v1/auth/me`                      | Current auth context      | Authenticated   |
| `POST` | `/api/v1/auth/ticket`                  | Short-lived auth ticket   | Authenticated   |

### 3. User Module

| Method  | Path                      | Description              | Access |
| :------ | :------------------------ | :----------------------- | :----- |
| `POST`  | `/api/v1/users/register`  | Register account         | Public |
| `GET`   | `/api/v1/users/me`        | Get own profile          | Authenticated |
| `PUT`   | `/api/v1/users/me`        | Update own profile       | Authenticated |
| `PATCH` | `/api/v1/users/me/avatar` | Update own avatar        | Authenticated |
| `GET`   | `/api/v1/users`           | List users               | Authorized |
| `POST`  | `/api/v1/users/search`    | Dynamic user search      | Authorized |
| `GET`   | `/api/v1/users/:id`       | Get user by id           | Authorized |
| `PATCH` | `/api/v1/users/:id/status`| Update user status       | Authorized |
| `DELETE`| `/api/v1/users/:id`       | Delete user              | Authorized |

### 4. Roles, Permissions, Access Rights

| Method | Path | Description |
| :----- | :--- | :---------- |
| `POST` | `/api/v1/roles` | Create role |
| `GET` | `/api/v1/roles` | List roles |
| `GET` | `/api/v1/roles/:id` | Get role |
| `PUT` | `/api/v1/roles/:id` | Update role |
| `DELETE` | `/api/v1/roles/:id` | Delete role |
| `POST` | `/api/v1/permissions/assign-role` | Assign role |
| `DELETE` | `/api/v1/permissions/revoke-role` | Revoke role |
| `POST` | `/api/v1/permissions/grant` | Grant permission |
| `GET` | `/api/v1/permissions` | List permissions |
| `GET` | `/api/v1/permissions/:role` | Get permissions for role |
| `GET` | `/api/v1/permissions/roles/:role/users` | Get users for role |
| `PUT` | `/api/v1/permissions` | Update permission |
| `DELETE` | `/api/v1/permissions/revoke` | Revoke permission |
| `POST` | `/api/v1/permissions/inheritance` | Add role inheritance |
| `DELETE` | `/api/v1/permissions/inheritance` | Remove role inheritance |
| `GET` | `/api/v1/permissions/:role/parents` | Get parent roles |
| `GET` | `/api/v1/permissions/resources` | Resource aggregation |
| `GET` | `/api/v1/permissions/inheritance-tree` | Inheritance tree |
| `GET` | `/api/v1/permissions/roles/:role/access-rights` | Role access-rights |
| `POST` | `/api/v1/permissions/assign-access-right` | Assign access-right |
| `DELETE` | `/api/v1/permissions/revoke-access-right` | Revoke access-right |
| `POST` | `/api/v1/permissions/check-batch` | Batch permission check |
| `POST` | `/api/v1/access-rights` | Create access right |
| `GET` | `/api/v1/access-rights` | List access rights |
| `GET` | `/api/v1/access-rights/:id` | Get access right |
| `PUT` | `/api/v1/access-rights/:id` | Update access right |
| `DELETE` | `/api/v1/access-rights/:id` | Delete access right |

### 5. Organization, Members, Projects, Integration

See:
- `documentation/api/ORG_API.md`
- `documentation/api/INTEGRATIONS_API.md`

---

## 🚫 IV. Common Errors

| Code    | Meaning           | Cause                                                   |
| :------ | :---------------- | :------------------------------------------------------ |
| **401** | Unauthorized      | Token missing, invalid, expired, or session revoked.    |
| **403** | Forbidden         | Authenticated, but no Casbin policy allows this action. |
| **422** | Validation Error  | Request body failed validation.                         |
| **429** | Too Many Requests | Rate limit exceeded.                                    |
