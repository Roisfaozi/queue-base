# Table-Driven Test Feasibility — Role, Access, Permission, Queue Registration Endpoint

## Tujuan

Menilai apakah **scope yang diminta user** bisa diubah ke table-driven test.

Scope yang dinilai hanya ini:

- `role`
- `access`
- `permission`
- flow pendaftaran endpoint baru, dengan seam utama `POST /api/v1/queues`

Dokumen ini memakai dua referensi sebagai pola, tanpa mengubahnya:

- `llm/plans/roadmap/table-driven-test-migration.md`
- `llm/conventions/table-driven-test-template.go.tmpl`

## Verdict Final

**Ya. Semua scope di atas bisa diubah ke table-driven test.**

Tidak ada bagian dari scope ini yang perlu dinyatakan “tidak bisa”.

Yang perlu dibedakan hanya:

- file mana yang jadi target utama
- level mana yang paling cocok: usecase atau controller
- cara memecah tabel agar tetap terbaca

## Batas Scope

Dokumen ini **tidak** sedang menilai:

- seluruh E2E naratif panjang lintas banyak modul
- seluruh flow lifecycle aplikasi di luar `role`, `access`, `permission`, dan pendaftaran endpoint queue
- seluruh repo secara umum

Jadi semua kesimpulan di bawah ini sengaja sempit: hanya untuk 4 area yang diminta.

## Precedent Lokal

Repo ini sudah punya precedent table-driven yang kuat di:

- `internal/modules/queue/usecase/queue_usecase_test.go`

File ini sudah menunjukkan pola yang diinginkan:

- `tests := []struct { ... }`
- `category string`
- `t.Run(tt.name, ...)`
- kombinasi `positive`, `negative`, `edge`, `vulnerability`

Artinya migrasi untuk `role`, `access`, `permission`, dan `queue registration endpoint` bukan eksperimen baru.

## Ringkasan Singkat per Area

| Area | Bisa table-driven? | Level paling cocok | Catatan utama |
|---|---|---|---|
| `role` | Ya | usecase + controller | branch kecil, error diskrit |
| `access` | Ya | usecase + controller | CRUD/validation sangat repetitif |
| `permission` | Ya | usecase + controller | wajib satu method satu tabel |
| `queue registration endpoint` | Ya | controller utama, usecase sudah ada precedent | register endpoint cocok untuk case matrix |

## 1. Role

### Kesimpulan

**Bisa diubah ke table-driven test.**

### Target file utama

- `internal/modules/role/usecase/role_usecase_test.go`
- `internal/modules/role/delivery/http/role_controller_test.go`

### Kenapa bisa

- dependency sederhana: repository, transaction manager, permission cleanup
- branch mudah dipetakan: success, conflict, not found, forbidden, internal error
- input domain kecil dan stabil

### Bentuk yang disarankan

#### Usecase

Satu tabel per method:

- `TestRoleUseCase_Create`
- `TestRoleUseCase_Update`
- `TestRoleUseCase_Delete`
- `TestRoleUseCase_GetAll`

#### Controller

Satu tabel per handler:

- create
- update
- delete
- dynamic search
- error mapping

### Kategori test yang cocok

- positive: create role baru, update description, delete role biasa
- negative: duplicate role, invalid payload, role tidak ditemukan
- edge: dynamic filter kosong, description kosong tapi valid
- vulnerability: delete `role:superadmin`

## 2. Access

### Kesimpulan

**Bisa diubah ke table-driven test.**

### Target file utama

- `internal/modules/access/test/access_usecase_test.go`
- `internal/modules/access/test/access_controller_test.go`

### Kenapa bisa

- operasi kecil dan berulang
- banyak skenario hanya beda body, validation, atau expected error
- relasi access-right ↔ endpoint cocok jadi matriks kasus

### Bentuk yang disarankan

#### Usecase

Satu tabel per method:

- `CreateAccessRight`
- `CreateEndpoint`
- `LinkEndpointToAccessRight`
- `UnlinkEndpointFromAccessRight`
- `DeleteAccessRight`
- `DeleteEndpoint`

#### Controller

Satu tabel per handler:

- create access-right
- create endpoint
- link/unlink
- dynamic search

### Kategori test yang cocok

- positive: create access-right, create endpoint, link success
- negative: malformed body, validation fail, invalid ID, delete non-existent
- edge: search filter kosong, description kosong
- vulnerability: access-right scoped context tidak bocor lintas tenant

## 3. Permission

### Kesimpulan

**Bisa diubah ke table-driven test.**

### Target file utama

- `internal/modules/permission/test/permission_usecase_test.go`
- `internal/modules/permission/test/access_right_assignment_test.go`
- `internal/modules/permission/test/permission_controller_test.go`

### Kenapa bisa

- branch permission sebenarnya matriks kecil yang berulang
- kombinasi user/role/domain/enforcer result cocok dimodelkan dalam tabel
- controller permission juga pola klasik: bind → validate → call usecase → map error

### Aturan penting

Untuk `permission`, bentuk yang benar adalah:

- **satu method = satu tabel**
- **jangan satu file = satu tabel raksasa**

Ini bukan karena “tidak bisa”, tapi karena kalau semua method digabung, test jadi sulit dibaca dan expectation mock Casbin cepat berantakan.

### Bentuk yang disarankan

#### Usecase

Satu tabel per method:

- `AssignRoleToUser`
- `RevokeRoleFromUser`
- `GrantPermissionToRole`
- `RevokePermissionFromRole`
- `BatchCheckPermission`
- `AddParentRole`
- `RemoveParentRole`

#### Access-right bridge

Satu tabel per method:

- `GetRoleAccessRights`
- `AssignAccessRight`
- `RevokeAccessRight`

#### Controller

Satu tabel per handler:

- assign-role
- revoke-role
- grant
- revoke
- batch-check
- assign-access-right
- revoke-access-right

### Kategori test yang cocok

- positive: assign role sukses, grant sukses, batch-check map benar
- negative: user tidak ada, role tidak ada, invalid payload, policy tidak ditemukan
- edge: domain kosong fallback ke `global`, items batch kosong
- vulnerability: org context override domain request, self-inheritance ditolak

## 4. Flow Pendaftaran Endpoint Baru (`POST /api/v1/queues`)

### Kesimpulan

**Bisa diubah ke table-driven test.**

### Target file utama

- `internal/modules/queue/delivery/http/queue_controller_test.go`

### Precedent terdekat

- `internal/modules/queue/usecase/queue_usecase_test.go`

### Kenapa bisa

Controller `Register` punya branch kecil dan stabil:

- invalid JSON
- validation fail
- usecase error
- success
- branch context terset dari body

Semua branch ini sangat cocok jadi tabel.

### Bentuk yang disarankan

Satu tabel untuk `TestQueueController_Register`, dengan field seperti:

- `name`
- `category`
- `body`
- `tenantID`
- `useCaseErr`
- `wantCode`
- `wantBranchID`
- `wantUseCaseCalled`

### Kategori test yang cocok

- positive: body valid, response `201`, branch context terset
- negative: malformed JSON, required field kosong, usecase `ErrBadRequest`
- edge: `patient_id` kosong tapi `patient_name` valid
- vulnerability: tenant context kosong atau branch invalid diteruskan dan ditolak di seam bawah

## File Target Tegas

Berikut file yang termasuk scope dan **bisa** dijadikan target table-driven:

### Role

- `internal/modules/role/usecase/role_usecase_test.go`
- `internal/modules/role/delivery/http/role_controller_test.go`

### Access

- `internal/modules/access/test/access_usecase_test.go`
- `internal/modules/access/test/access_controller_test.go`

### Permission

- `internal/modules/permission/test/permission_usecase_test.go`
- `internal/modules/permission/test/access_right_assignment_test.go`
- `internal/modules/permission/test/permission_controller_test.go`

### Queue registration endpoint

- `internal/modules/queue/delivery/http/queue_controller_test.go`

## Rule Implementasi

Saat migrasi nanti, pakai rule ini:

1. satu fungsi test per method/handler
2. `tests := []struct { ... }`
3. wajib ada `category string`
4. wajib `t.Run(tt.name, ...)`
5. setup mock/stub diturunkan dari `tt`
6. jangan hardcode skenario per `t.Run` di luar tabel
7. untuk `permission`, pecah per method agar mock Casbin tetap terkendali

## Contoh Bentuk Minimal

```go
func TestAssignRoleToUser(t *testing.T) {
	tests := []struct {
		name     string
		category string
		userID   string
		role     string
		domain   string
		stub     struct {
			userErr        error
			roleErr        error
			removeErr      error
			addGroupingErr error
		}
		wantErr error
	}{
		// cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup mock dari tt.stub
			// execute
			// assert
		})
	}
}
```

## Penutup

Untuk scope yang diminta user:

- `role`
- `access`
- `permission`
- flow pendaftaran endpoint baru

**semuanya bisa diubah ke table-driven test.**

Perbedaan hanya pada strategi pemecahan:

- `role` dan `access`: paling lurus
- `permission`: tetap bisa, tapi pecah per method
- `queue registration endpoint`: sangat lurus di controller, dan sudah punya precedent lokal kuat di usecase
