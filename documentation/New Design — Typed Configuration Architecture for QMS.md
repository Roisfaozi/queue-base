# New Design — Typed Configuration Architecture for QMS

## 1. Keputusan Final dari diskusi 1 JULI 2026

Generic `settings` table dengan `scope_type`, `scope_id`, `key`, dan `value` **tidak dipakai lagi untuk core QMS configuration**.

Core profile dan configuration harus dipisah menjadi typed tables:

```text
tenants
tenant_queue_settings

branches
branch_queue_settings

services
branch_services
service_queue_settings

counters
counter_queue_settings
```

Untuk display/signage MVP:

```text
tenant logo disimpan di tenants.logo_asset_id
branch running_text disimpan di branches.running_text
branch logo disimpan di branches.logo_asset_id nullable
```

Display table khusus seperti `tenant_display_settings` dan `branch_display_settings` **belum wajib untuk MVP**. Table itu bisa dibuat nanti kalau signage sudah punya theme, layout, warna, audio, dan display rule yang kompleks.

---

## 2. Prinsip Utama

### 2.1 Profile Data Masuk ke Main Entity

Data yang menjelaskan identitas entitas harus masuk ke table utama.

Contoh tenant:

```text
name
legal_name
address
city
province
phone
logo_asset_id
timezone
status
```

Contoh branch:

```text
name
address
city
province
phone
logo_asset_id
running_text
timezone
status
```

Data seperti ini **bukan settings**.

---

### 2.2 Behavior Configuration Masuk ke Typed Settings Table

Data yang mengatur perilaku sistem masuk ke table configuration khusus.

Contoh:

```text
queue_reset_time
ticket_prefix
default_estimated_duration
allow_forward
allow_skip
allow_recall
allow_cancel
numbering_strategy
```

---

### 2.3 Generic Settings Tidak Dipakai untuk Core QMS

Generic settings boleh dipertahankan hanya untuk:

```text
experimental flags
temporary feature toggles
non-critical UI preferences
future dynamic config
```

Tapi untuk core QMS, pakai typed tables.

---

## 3. NEW Table Design

## 3.1 `tenants`

Tenant adalah company, hospital, clinic, atau organization utama.

```sql
tenants
- id
- code
- name
- legal_name
- address
- city
- province
- postal_code nullable
- phone
- email nullable
- logo_asset_id
- timezone
- status
- created_at
- updated_at
- deleted_at
```

### Required Fields

```text
code
name
address
city
province
phone
logo_asset_id
timezone
status
```

### Tenant Status

```text
draft
active
inactive
suspended
```

### Tenant Activation Rule

Tenant tidak boleh `active` kalau field berikut kosong:

```text
name
address
city
province
phone
logo_asset_id
timezone
```

### Logo Rule

Logo tenant wajib.

```text
tenants.logo_asset_id required
```

Logo tidak disimpan sebagai binary di table tenant. Simpan reference ke asset/media module.

---

## 3.2 `tenant_queue_settings`

Tenant queue settings adalah default queue behavior untuk seluruh tenant.

```sql
tenant_queue_settings
- id
- tenant_id
- queue_reset_time
- default_ticket_prefix
- default_estimated_duration
- allow_forward
- allow_skip
- allow_recall
- allow_cancel
- numbering_strategy
- created_at
- updated_at
```

### Recommended Defaults

```text
queue_reset_time = 04:00
default_ticket_prefix = A
default_estimated_duration = 5
allow_forward = true
allow_skip = true
allow_recall = true
allow_cancel = true
numbering_strategy = daily_branch_sequence
```

### Unique Constraint

```text
unique tenant_id
```

Satu tenant hanya punya satu tenant queue setting.

---

## 3.3 `branches`

Branch adalah lokasi operasional di bawah tenant.

```sql
branches
- id
- tenant_id
- code
- name
- address
- city
- province
- postal_code nullable
- phone
- email nullable
- logo_asset_id nullable
- running_text
- timezone
- status
- created_at
- updated_at
- deleted_at
```

### Required Fields

```text
tenant_id
code
name
address
city
province
phone
running_text
timezone
status
```

### Optional Fields

```text
logo_asset_id
email
postal_code
```

### Branch Activation Rule

Branch tidak boleh `active` kalau field berikut kosong:

```text
name
address
city
province
phone
running_text
timezone
```

Logo branch **optional**.

Jika branch logo kosong, display/signage menggunakan tenant logo.

```text
effective_logo_asset_id =
  if branches.logo_asset_id is not null:
      branches.logo_asset_id
  else:
      tenants.logo_asset_id
```

### Running Text Rule

`branches.running_text` wajib karena akan dipakai untuk display/signage.

Contoh:

```text
Selamat Datang di Rumah Sakit RS KITA SEMUA
```

Saat branch dibuat, default running text boleh otomatis:

```text
Selamat Datang di {branch.name}
```

Tapi tetap harus tersimpan di `branches.running_text`.

---

## 3.4 `branch_queue_settings`

Branch queue settings adalah override dari tenant queue settings.

```sql
branch_queue_settings
- id
- tenant_id
- branch_id
- queue_reset_time nullable
- ticket_prefix nullable
- default_estimated_duration nullable
- allow_forward nullable
- allow_skip nullable
- allow_recall nullable
- allow_cancel nullable
- numbering_strategy nullable
- created_at
- updated_at
```

### Nullable Means Inherit

Semua field nullable berarti:

```text
gunakan tenant_queue_settings
```

Contoh:

```text
tenant_queue_settings.queue_reset_time = 04:00
branch_queue_settings.queue_reset_time = null

effective = 04:00
```

Kalau branch override:

```text
tenant_queue_settings.queue_reset_time = 04:00
branch_queue_settings.queue_reset_time = 05:00

effective = 05:00
```

### Unique Constraint

```text
unique tenant_id + branch_id
```

Satu branch hanya punya satu branch queue settings row.

---

## 3.5 `services`

Service adalah template layanan milik tenant.

Contoh:

```text
Registration
Doctor
Pharmacy
Cashier
Laboratory
```

```sql
services
- id
- tenant_id
- code
- name
- type
- is_pharmacy
- is_pharmacy_reception
- default_estimated_duration
- status
- created_at
- updated_at
- deleted_at
```

### Required Fields

```text
tenant_id
code
name
type
default_estimated_duration
status
```

### Service Type

```text
registration
doctor
pharmacy
cashier
laboratory
general
```

### Pharmacy Flags

```text
is_pharmacy
is_pharmacy_reception
```

Rule:

```text
is_pharmacy_reception = true
```

hanya untuk service yang menjadi pintu masuk valid menuju pharmacy flow, misalnya:

```text
Penerimaan Resep
```

---

## 3.6 `branch_services`

Branch services menentukan service mana yang aktif di sebuah branch.

```sql
branch_services
- id
- tenant_id
- branch_id
- service_id
- custom_name nullable
- is_active
- sort_order
- created_at
- updated_at
```

### Required Fields

```text
tenant_id
branch_id
service_id
is_active
sort_order
```

### Unique Constraint

```text
unique tenant_id + branch_id + service_id
```

### Why This Table Exists

Service adalah template tenant-level.

Branch service adalah aktivasi service di branch tertentu.

Contoh:

```text
Tenant services:
- Registration
- Doctor
- Pharmacy

Branch Jakarta enabled services:
- Registration
- Doctor
- Pharmacy

Branch Bandung enabled services:
- Registration
- Pharmacy
```

Queue creation dan forwarding hanya boleh menggunakan service yang aktif di `branch_services`.

---

## 3.7 `service_queue_settings`

Service queue settings mengatur behavior khusus service.

```sql
service_queue_settings
- id
- tenant_id
- service_id
- default_estimated_duration nullable
- require_counter nullable
- allow_forward_from nullable
- allow_forward_to nullable
- allow_skip nullable
- allow_recall nullable
- allow_cancel nullable
- created_at
- updated_at
```

### Nullable Means Inherit

Jika field kosong, gunakan fallback dari branch atau tenant.

Contoh estimated duration:

```text
tenant default_estimated_duration = 5
branch default_estimated_duration = null
service default_estimated_duration = 8

effective = 8
```

### Unique Constraint

```text
unique tenant_id + service_id
```

---

## 3.8 `counters`

Counter adalah loket, station, room, atau titik layanan.

```sql
counters
- id
- tenant_id
- branch_id
- branch_service_id
- code
- name
- display_name
- status
- created_at
- updated_at
- deleted_at
```

### Important Decision

Counter sebaiknya terhubung ke `branch_service_id`, bukan hanya `service_id`.

Alasannya:

```text
branch_service_id sudah memastikan service tersebut aktif di branch itu.
```

Dengan ini, counter tidak bisa tidak sengaja menunjuk service yang belum aktif di branch.

### Required Fields

```text
tenant_id
branch_id
branch_service_id
code
name
display_name
status
```

### Counter Validation

Counter valid jika:

```text
counter.tenant_id = request.tenant_id
counter.branch_id = request.branch_id
counter.branch_service_id belongs to same tenant + branch
```

---

## 3.9 `counter_queue_settings`

Counter queue settings mengatur behavior spesifik counter.

```sql
counter_queue_settings
- id
- tenant_id
- branch_id
- counter_id
- audio_enabled nullable
- display_enabled nullable
- auto_call_next nullable
- allow_recall nullable
- allow_skip nullable
- created_at
- updated_at
```

### Nullable Means Inherit

Jika kosong, gunakan service/branch/tenant default.

### Unique Constraint

```text
unique tenant_id + branch_id + counter_id
```

---

# 4. NEW Relationship Design

```text
tenants
  ├── tenant_queue_settings
  ├── branches
  │     ├── branch_queue_settings
  │     ├── branch_services
  │     │     ├── counters
  │     │     │     └── counter_queue_settings
  │     │     └── service_queue_settings through service
  │     ├── queues
  │     ├── queue_journeys
  │     └── visit_journeys
  └── services
        └── service_queue_settings
```

---

# 5. Effective Configuration Resolution

Walaupun generic `settings` dihapus dari core QMS, effective config tetap ada.

## 5.1 Queue Config Resolution Order

```text
tenant_queue_settings
  ↓ overridden by
branch_queue_settings
  ↓ overridden by
service_queue_settings
  ↓ overridden by
counter_queue_settings
```

### Example: Queue Reset Time

Queue reset time hanya relevan tenant/branch.

```text
tenant_queue_settings.queue_reset_time = 04:00
branch_queue_settings.queue_reset_time = 05:00

effective queue_reset_time = 05:00
source = branch
```

### Example: Ticket Prefix

```text
tenant_queue_settings.default_ticket_prefix = A
branch_queue_settings.ticket_prefix = JS

effective ticket_prefix = JS
source = branch
```

### Example: Estimated Duration

```text
tenant_queue_settings.default_estimated_duration = 5
branch_queue_settings.default_estimated_duration = null
service_queue_settings.default_estimated_duration = 8

effective estimated_duration = 8
source = service
```

### Example: Auto Call Next

```text
counter_queue_settings.auto_call_next = true

effective auto_call_next = true
source = counter
```

---

## 5.2 Effective Config Response

Effective config API harus tetap UX-friendly.

```json
{
  "tenant": {
    "id": "tenant_001",
    "name": "RS Kita Semua",
    "logo_asset_id": "asset_tenant_logo"
  },
  "branch": {
    "id": "branch_001",
    "name": "RS Kita Semua - Jakarta Selatan",
    "logo_asset_id": null,
    "effective_logo_asset_id": "asset_tenant_logo",
    "running_text": "Selamat Datang di Rumah Sakit RS KITA SEMUA"
  },
  "queue": {
    "queue_reset_time": {
      "value": "05:00",
      "source": "branch",
      "is_overridden": true,
      "can_override": true,
      "can_reset": true
    },
    "ticket_prefix": {
      "value": "JS",
      "source": "branch",
      "is_overridden": true,
      "can_override": true,
      "can_reset": true
    },
    "default_estimated_duration": {
      "value": 5,
      "source": "tenant",
      "is_overridden": false,
      "can_override": true,
      "can_reset": false
    },
    "allow_forward": {
      "value": true,
      "source": "tenant",
      "is_overridden": false,
      "can_override": true,
      "can_reset": false
    }
  }
}
```

---

# 6. API Design

## 6.1 Tenant Profile

```text
GET    /api/v1/tenant/profile
PATCH  /api/v1/tenant/profile
```

### PATCH Request

```json
{
  "name": "RS Kita Semua",
  "legal_name": "PT Rumah Sakit Kita Semua",
  "address": "Jl. Merdeka No. 1",
  "city": "Jakarta Selatan",
  "province": "DKI Jakarta",
  "postal_code": "12190",
  "phone": "021-123456",
  "email": "admin@rskitasemua.id",
  "logo_asset_id": "asset_logo_001",
  "timezone": "Asia/Jakarta"
}
```

---

## 6.2 Tenant Queue Settings

```text
GET    /api/v1/tenant/queue-settings
PATCH  /api/v1/tenant/queue-settings
```

### PATCH Request

```json
{
  "queue_reset_time": "04:00",
  "default_ticket_prefix": "A",
  "default_estimated_duration": 5,
  "allow_forward": true,
  "allow_skip": true,
  "allow_recall": true,
  "allow_cancel": true,
  "numbering_strategy": "daily_branch_sequence"
}
```

---

## 6.3 Branch Profile

```text
GET    /api/v1/branches/{branch_id}/profile
PATCH  /api/v1/branches/{branch_id}/profile
```

### PATCH Request

```json
{
  "name": "RS Kita Semua - Jakarta Selatan",
  "address": "Jl. Sehat No. 10",
  "city": "Jakarta Selatan",
  "province": "DKI Jakarta",
  "postal_code": "12240",
  "phone": "021-999999",
  "email": "jaksel@rskitasemua.id",
  "logo_asset_id": null,
  "running_text": "Selamat Datang di Rumah Sakit RS KITA SEMUA",
  "timezone": "Asia/Jakarta"
}
```

---

## 6.4 Branch Queue Settings

```text
GET    /api/v1/branches/{branch_id}/queue-settings
PATCH  /api/v1/branches/{branch_id}/queue-settings
DELETE /api/v1/branches/{branch_id}/queue-settings/{field}
```

### PATCH Request

```json
{
  "queue_reset_time": "05:00",
  "ticket_prefix": "JS",
  "default_estimated_duration": null,
  "allow_forward": true,
  "allow_skip": null,
  "allow_recall": null,
  "allow_cancel": null
}
```

### DELETE Meaning

`DELETE` berarti reset override field ke inherited value.

Contoh:

```text
DELETE /api/v1/branches/{branch_id}/queue-settings/ticket_prefix
```

Maka:

```text
branch_queue_settings.ticket_prefix = null
```

---

## 6.5 Service

```text
POST   /api/v1/services
GET    /api/v1/services
GET    /api/v1/services/{service_id}
PATCH  /api/v1/services/{service_id}
DELETE /api/v1/services/{service_id}
```

### Create Service Request

```json
{
  "code": "REG",
  "name": "Registration",
  "type": "registration",
  "is_pharmacy": false,
  "is_pharmacy_reception": false,
  "default_estimated_duration": 5,
  "status": "active"
}
```

---

## 6.6 Branch Service

```text
GET    /api/v1/branches/{branch_id}/services
POST   /api/v1/branches/{branch_id}/services/{service_id}/enable
POST   /api/v1/branches/{branch_id}/services/{service_id}/disable
PATCH  /api/v1/branches/{branch_id}/services/{service_id}
```

### Enable Service Request

```json
{
  "custom_name": "Pendaftaran",
  "sort_order": 1,
  "is_active": true
}
```

---

## 6.7 Service Queue Settings

```text
GET    /api/v1/services/{service_id}/queue-settings
PATCH  /api/v1/services/{service_id}/queue-settings
DELETE /api/v1/services/{service_id}/queue-settings/{field}
```

---

## 6.8 Counter

```text
POST   /api/v1/branches/{branch_id}/counters
GET    /api/v1/branches/{branch_id}/counters
GET    /api/v1/branches/{branch_id}/counters/{counter_id}
PATCH  /api/v1/branches/{branch_id}/counters/{counter_id}
DELETE /api/v1/branches/{branch_id}/counters/{counter_id}
```

### Create Counter Request

```json
{
  "branch_service_id": "branch_service_001",
  "code": "REG-01",
  "name": "Registration Counter 1",
  "display_name": "Loket Pendaftaran 1",
  "status": "active"
}
```

---

## 6.9 Counter Queue Settings

```text
GET    /api/v1/branches/{branch_id}/counters/{counter_id}/queue-settings
PATCH  /api/v1/branches/{branch_id}/counters/{counter_id}/queue-settings
DELETE /api/v1/branches/{branch_id}/counters/{counter_id}/queue-settings/{field}
```

---

## 6.10 Effective Config

```text
GET /api/v1/branches/{branch_id}/effective-config
GET /api/v1/branches/{branch_id}/services/{service_id}/effective-config
GET /api/v1/branches/{branch_id}/counters/{counter_id}/effective-config
```

---

# 7. Setup Wizard UX

## Step 1 — Tenant Profile

Required:

```text
Tenant name
Legal name optional
Address
City
Province
Phone
Logo
Timezone
```

Tidak bisa lanjut ke aktivasi tenant jika belum lengkap.

---

## Step 2 — Tenant Queue Default

Required:

```text
Queue reset time
Default ticket prefix
Default estimated duration
Allow forward
Allow skip
Allow recall
Allow cancel
```

---

## Step 3 — Branch Profile

Required:

```text
Branch name
Address
City
Province
Phone
Running text
Timezone
```

Optional:

```text
Branch logo
```

Jika branch logo kosong, UI tampilkan:

```text
This branch will use the tenant logo.
```

---

## Step 4 — Branch Queue Override

UI menampilkan inherited value dari tenant.

Contoh:

```text
Queue reset time: 04:00 inherited from tenant
Ticket prefix: A inherited from tenant
Estimated duration: 5 minutes inherited from tenant
```

Admin boleh override:

```text
Queue reset time: 05:00
Ticket prefix: JS
```

---

## Step 5 — Service Setup

Admin membuat template service:

```text
Registration
Doctor
Pharmacy
Cashier
Laboratory
```

---

## Step 6 — Enable Service for Branch

Admin memilih service yang aktif di branch:

```text
Registration enabled
Doctor enabled
Pharmacy enabled
Cashier disabled
```

---

## Step 7 — Counter Setup

Admin membuat counter berdasarkan branch service:

```text
Registration → Loket Pendaftaran 1
Doctor → Ruang Dokter 1
Pharmacy → Loket Farmasi 1
```

---

# 8. Validation Rules

## 8.1 Tenant Validation

Tenant profile update harus validasi:

```text
name required
address required
city required
province required
phone required
logo_asset_id required
timezone required
```

Tenant activation harus menolak jika profile incomplete.

---

## 8.2 Branch Validation

Branch profile update harus validasi:

```text
tenant_id required
name required
address required
city required
province required
phone required
running_text required
timezone required
```

Branch activation harus menolak jika profile incomplete.

Branch logo optional.

---

## 8.3 Branch Logo Fallback

Display/signage harus resolve logo dengan rule:

```text
branch.logo_asset_id if exists
else tenant.logo_asset_id
```

Jika tenant logo tidak ada, tenant seharusnya tidak aktif.

---

## 8.4 Branch Service Validation

Queue creation dan forward hanya boleh menggunakan service jika:

```text
service belongs to tenant
branch_service exists
branch_service belongs to tenant + branch
branch_service.is_active = true
```

---

## 8.5 Counter Validation

Counter valid jika:

```text
counter belongs to tenant
counter belongs to branch
counter.branch_service_id belongs to same tenant + branch
counter status = active
```

---

# 9. Queue Config Usage

## 9.1 Queue Date

Queue date memakai effective queue reset time.

Resolution:

```text
branch_queue_settings.queue_reset_time
fallback tenant_queue_settings.queue_reset_time
```

Rule:

```text
if local branch time < queue_reset_time:
    queue_date = previous date
else:
    queue_date = current date
```

---

## 9.2 Ticket Prefix

Ticket prefix memakai:

```text
branch_queue_settings.ticket_prefix
fallback tenant_queue_settings.default_ticket_prefix
```

---

## 9.3 Estimated Duration

Estimated duration memakai urutan:

```text
counter_queue_settings if available
fallback service_queue_settings
fallback branch_queue_settings
fallback tenant_queue_settings
fallback services.default_estimated_duration
```

Untuk MVP, cukup:

```text
service_queue_settings.default_estimated_duration
fallback branch_queue_settings.default_estimated_duration
fallback tenant_queue_settings.default_estimated_duration
fallback services.default_estimated_duration
```

---

# 10. Audit Log Design

Profile dan configuration sekarang adalah domain penting, jadi harus diaudit.

## Tenant Audit Events

```text
qms.tenant.profile.updated
qms.tenant.logo.updated
qms.tenant.queue_settings.updated
qms.tenant.activated
qms.tenant.deactivated
```

## Branch Audit Events

```text
qms.branch.profile.updated
qms.branch.logo.updated
qms.branch.running_text.updated
qms.branch.queue_settings.updated
qms.branch.activated
qms.branch.deactivated
```

## Service Audit Events

```text
qms.service.created
qms.service.updated
qms.service.enabled_for_branch
qms.service.disabled_for_branch
qms.service.queue_settings.updated
```

## Counter Audit Events

```text
qms.counter.created
qms.counter.updated
qms.counter.queue_settings.updated
qms.counter.activated
qms.counter.deactivated
```

## Audit Metadata

```json
{
  "tenant_id": "tenant_001",
  "branch_id": "branch_001",
  "actor_id": "user_001",
  "resource_type": "branch",
  "resource_id": "branch_001",
  "action": "qms.branch.running_text.updated",
  "changed_fields": ["running_text"],
  "before": {
    "running_text": "Old running text"
  },
  "after": {
    "running_text": "Selamat Datang di Rumah Sakit RS KITA SEMUA"
  }
}
```

Untuk logo:

```json
{
  "old_logo_asset_id": "asset_old",
  "new_logo_asset_id": "asset_new"
}
```

Jangan simpan file binary di audit log.

---

# 11. Error Logging Design

Semua error path harus structured.

## Required Log Context

```text
module
action
tenant_id
branch_id if available
user_id / actor_id
request_id / trace_id
resource_id
error
```

## Required Error Logs

```text
tenant profile update failed
tenant activation failed because required profile incomplete
tenant logo missing
tenant queue settings update failed

branch profile update failed
branch activation failed because required profile incomplete
branch running_text missing
branch queue settings update failed
branch effective config resolve failed

service create/update failed
service enable for branch failed
invalid service for branch

counter create/update failed
invalid counter branch relation
invalid counter branch service relation

effective config resolution failed
```

## Sensitive Data Rule

Do not log:

```text
raw API keys
JWT tokens
passwords
full request body
sensitive patient data
raw uploaded file data
```

---

# 12. Migration Strategy

## Step 1 — Add New Columns

Add to tenants:

```text
address
city
province
postal_code nullable
phone
email nullable
logo_asset_id
timezone
status
```

Add to branches:

```text
address
city
province
postal_code nullable
phone
email nullable
logo_asset_id nullable
running_text
timezone
status
```

---

## Step 2 — Create Typed Settings Tables

Create:

```text
tenant_queue_settings
branch_queue_settings
service_queue_settings
counter_queue_settings
branch_services
```

---

## Step 3 — Backfill From Generic Settings

Mapping examples:

```text
settings scope=tenant key=queue_reset_time
→ tenant_queue_settings.queue_reset_time

settings scope=tenant key=ticket_prefix
→ tenant_queue_settings.default_ticket_prefix

settings scope=tenant key=default_estimated_duration
→ tenant_queue_settings.default_estimated_duration

settings scope=branch key=queue_reset_time
→ branch_queue_settings.queue_reset_time

settings scope=branch key=ticket_prefix
→ branch_queue_settings.ticket_prefix

settings scope=branch key=running_text
→ branches.running_text
```

---

## Step 4 — Backfill Defaults

If tenant queue settings missing:

```text
queue_reset_time = 04:00
default_ticket_prefix = A
default_estimated_duration = 5
allow_forward = true
allow_skip = true
allow_recall = true
allow_cancel = true
numbering_strategy = daily_branch_sequence
```

If branch running text missing:

```text
Selamat Datang di {branch.name}
```

But mark branch as `draft` until admin confirms.

---

## Step 5 — Deprecate Generic Settings

Do not immediately drop generic settings.

Recommended phase:

```text
Phase A:
new typed tables active
old settings read-only

Phase B:
old settings no longer read by QMS core

Phase C:
drop old settings only after tests and migration validation pass
```

---

# 13. Testing Requirements

## Tenant Tests

```text
tenant cannot activate without address
tenant cannot activate without city
tenant cannot activate without province
tenant cannot activate without phone
tenant cannot activate without logo_asset_id
tenant profile update writes audit log
tenant queue settings update writes audit log
```

## Branch Tests

```text
branch cannot activate without address
branch cannot activate without city
branch cannot activate without province
branch cannot activate without phone
branch cannot activate without running_text
branch can activate without logo if tenant logo exists
branch effective logo falls back to tenant logo
branch running_text update writes audit log
branch queue settings update writes audit log
```

## Effective Config Tests

```text
branch queue reset time inherits tenant value
branch queue reset time overrides tenant value
branch ticket prefix inherits tenant value
branch ticket prefix overrides tenant value
service estimated duration overrides branch/tenant
counter setting overrides service/branch/tenant
reset branch override returns tenant value
effective config response includes source metadata
```

## Relation Tests

```text
service not enabled for branch cannot be used for queue creation
service not enabled for branch cannot be used for forward
counter from different branch rejected
counter with invalid branch_service rejected
cross-tenant branch service rejected
```

---

# 14. NEW Architecture Decision

## Use Typed Domain Tables

Core QMS configuration must use typed tables:

```text
tenant_queue_settings
branch_queue_settings
service_queue_settings
counter_queue_settings
```

## Keep Profile Data on Main Entity

Tenant profile belongs to `tenants`.

Branch profile belongs to `branches`.

## Keep Running Text on Branch

For MVP:

```text
branches.running_text
```

Future signage customization can use:

```text
branch_display_settings
```

but do not create it now unless needed.

## Use Branch Services

Service activation per branch must use:

```text
branch_services
```

Counter should reference:

```text
branch_service_id
```

not just `service_id`.

## Keep Effective Config API

Even though generic settings is removed from core QMS, effective config API remains important for UX.

## Generic Settings Future Role

Generic settings can remain only for non-core, experimental, or future dynamic configuration.

It should not control core QMS behavior.

---

# 15. NEW Recommended Table List for MVP

```text
tenants
tenant_queue_settings

branches
branch_queue_settings

services
branch_services
service_queue_settings

counters
counter_queue_settings

queues
queue_journeys
visit_journeys
queue_counters

audit_logs
```

lanjut lihat diagram `QMS NEW Design Diagrams.md`
