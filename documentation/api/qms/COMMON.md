# QMS API Common Reference

## Base URL

Semua endpoint QMS ada di bawah prefix:

`/api/v1`

## Common Security Requirement

Semua endpoint QMS aktif berada di route group tenant-authorized.

Secara operasional request normal perlu:

- `Authorization: Bearer <access_token>`
- `X-Organization-ID: <tenant_uuid>`

Boundary runtime yang juga berlaku:

- active user session
- active user status
- auto API scope requirement
- Casbin authorization

Scanner endpoint punya tambahan header sendiri.

## Common Response Envelope

### Success

```json
{
  "data": {}
}
```

### Error

```json
{
  "message": "failed to create queue",
  "error": "bad request"
}
```

## Shared Enums

### Branch / Service / Counter status

- `active`
- `inactive`

### Setting scope_type

- `tenant`
- `branch`
- `service`
- `counter`

### Setting value_type

- `string`
- `number`
- `boolean`
- `json`

### Queue action

- `call`
- `serve`
- `complete`
- `skip`
- `cancel`

### Queue status

- `waiting`
- `calling`
- `serving`
- `skipped`
- `canceled`
- `completed`

### Scanner action

- `register`
- `forward`

## Common Error Notes

### `400 Bad Request`

Biasanya terjadi bila:

- body tidak valid
- field enum salah
- branch scope tidak ada
- transition action tidak valid untuk status saat ini
- header scanner kurang

### `401 Unauthorized`

Biasanya terjadi bila:

- bearer token tidak valid
- session tidak aktif
- scanner client/API key gagal diautentikasi

### `403 Forbidden`

Biasanya terjadi bila:

- user tidak lolos Casbin/scope
- relation validator menolak akses lintas tenant/branch/service/counter

### `404 Not Found`

Biasanya terjadi bila:

- entity target tidak ada dalam tenant scope
- current journey queue tidak ditemukan
- setting/key tidak ditemukan pada inheritance chain resolve
