# New Design Document — QMS MVP Operational Flow, Typed Configuration, Caller, Signage, Credential Binding

## 1. Purpose

Dokumen ini mendefinisikan desain New untuk rebuild QMS berbasis multi-tenant dan multi-branch.

Fokus utama:

1. Tenant sebagai boundary utama.
2. Branch berada di bawah tenant.
3. Semua data QMS scoped by `tenant_id`.
4. Semua data operasional branch scoped by `tenant_id + branch_id`.
5. Core QMS menggunakan typed configuration.
6. Generic settings dihapus sepenuhnya.
7. Queue menggunakan normalized design.
8. Forwarding tidak membuat queue baru.
9. Forwarding membuat `queue_journeys`.
10. Satu queue hanya memiliki satu `ticket_no` dan satu `queue_no`.
11. `visit_journeys` menjadi internal readable timeline.
12. Estimate time tetap dihitung untuk operasional.
13. Timestamp lama seperti `State1Timestamp` sampai `State5Timestamp` tidak digunakan.
14. Timestamp operasional diganti dengan typed journey timestamps.
15. UI dibagi menjadi Dashboard, Caller, dan Signage.
16. Caller dan Signage langsung resolve tenant, branch, service, dan counter dari credential/context.
17. Caller API menggunakan satu endpoint action.
18. Recall bukan action terpisah; repeated call dianggap recall internal.
19. Audit log dan structured error log wajib untuk QMS.

---

# 2. MVP Non-Goals

Fitur berikut tidak masuk scope MVP:

```text
submenu
device as core domain
arrive / check-in
Medifrans integration
Bithealth integration
third-party appointment integration
payment
insurance
notification provider
advanced display layout engine
advanced doctor schedule
advanced analytics
```

Catatan:

```text
Konsep device lama diganti menjadi qms_clients untuk caller/signage/scanner/kiosk credential binding.
```

---

# 3. New MVP Operational Flow

Flow MVP utama:

```text
Tenant setup
→ Branch setup
→ Service setup
→ Branch service setup
→ Counter setup
→ QMS client setup
→ Operator assignment
→ Caller login
→ Create queue
→ Call queue
→ Start service
→ Forward queue
→ Complete queue
→ Visit journey timeline
→ Signage display
→ Basic dashboard
```

---

# 4. Core Architecture Principles

## 4.1 Tenant First

Tenant adalah boundary utama.

Semua business data wajib memiliki:

```text
tenant_id
```

Rule:

```text
Every request must resolve tenant context first.
Every repository query must filter by tenant_id.
Every mutation must write tenant_id.
Every relationship must validate tenant ownership.
```

Tidak boleh ada query QMS tanpa tenant scope.

---

## 4.2 Branch Under Tenant

Branch selalu berada di bawah tenant.

```text
Tenant
  └── Branch
        ├── Branch Services
        ├── Counters
        ├── Queues
        ├── Queue Journeys
        ├── Visit Journeys
        ├── QMS Clients
        └── Branch Configuration
```

Branch valid hanya jika:

```text
branches.tenant_id = request.tenant_id
```

Query untuk branch-owned resource wajib menggunakan:

```text
WHERE tenant_id = ? AND branch_id = ?
```

---

## 4.3 One Queue Record Per Ticket Per Visit

Queue parent harus tetap satu record.

```text
1 queue = 1 ticket
1 queue = 1 queue number
1 queue = 1 visit
```

Forwarding tidak boleh membuat queue baru.

---

## 4.4 Forwarding Uses Queue Journey

Bad design:

```text
Forwarding creates another queue/customer record.
```

New design:

```text
Forwarding keeps the same queue record.
Forwarding creates a new queue_journey.
```

Forwarding harus menjaga:

```text
same queue_id
same ticket_no
same queue_no
```

---

## 4.5 No Generic Settings Anywhere

Generic `settings` table dihapus sepenuhnya dari desain dan implementasi QMS.

Tidak ada:

```text
scope_type
scope_id
key
value
generic fallback
```

Semua konfigurasi wajib menggunakan typed tables.

---

## 4.6 Profile Data Is Not Settings

Data identitas entitas harus berada di main table.

Tenant profile masuk ke:

```text
tenants
```

Branch profile masuk ke:

```text
branches
```

Service identity masuk ke:

```text
services
```

Counter identity masuk ke:

```text
counters
```

---

## 4.7 Behavior Configuration Uses Typed Tables

Configuration yang mengatur perilaku sistem masuk ke typed config tables:

```text
tenant_queue_settings
branch_queue_settings
branch_service_queue_settings
counter_queue_settings
```

---

# 5. New Core Table List

```text
tenants
tenant_queue_settings

branches
branch_queue_settings

services
branch_services
branch_service_queue_settings

counters
counter_queue_settings

qms_clients
qms_client_credentials
operator_counter_assignments

queues
queue_journeys
visit_journeys
queue_counters

audit_logs
```

Platform/auth starter boleh tetap memiliki:

```text
users
roles
permissions
role_permissions
user_branch_access
sessions
assets
files
```

---

# 6. Relationship Summary

```text
tenants
  ├── tenant_queue_settings
  ├── branches
  │     ├── branch_queue_settings
  │     ├── branch_services
  │     │     ├── branch_service_queue_settings
  │     │     └── counters
  │     │           └── counter_queue_settings
  │     ├── qms_clients
  │     │     └── qms_client_credentials
  │     ├── operator_counter_assignments
  │     ├── queues
  │     │     ├── queue_journeys
  │     │     └── visit_journeys
  │     └── queue_counters
  └── services
```

---

# 7. Tenant Design

## 7.1 `tenants`

```text
tenants
- id
- code
- name
- legal_name nullable
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
```

Tidak menggunakan `deleted_at`.

Jika tenant tidak boleh dipakai, gunakan `status`.

## 7.2 Required Fields

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

## 7.3 Tenant Status

```text
draft
active
inactive
suspended
```

## 7.4 Tenant Activation Rule

Tenant tidak boleh `active` jika field berikut belum lengkap:

```text
name
address
city
province
phone
logo_asset_id
timezone
```

Logo tenant wajib.

Logo disimpan sebagai asset reference:

```text
logo_asset_id
```

bukan binary file langsung di table tenant.

---

# 8. Tenant Queue Settings

## 8.1 `tenant_queue_settings`

```text
tenant_queue_settings
- id
- tenant_id
- queue_reset_time
- default_ticket_prefix
- default_estimated_duration
- allow_forward
- allow_recall
- auto_call_next
- numbering_strategy
- created_at
- updated_at
```

## 8.2 Defaults

```text
queue_reset_time = 04:00
default_ticket_prefix = A
default_estimated_duration = 5
allow_forward = true
allow_recall = false
auto_call_next = false
numbering_strategy = daily_branch_sequence
```

## 8.3 Unique Constraint

```text
unique tenant_id
```

## 8.4 Removed Config

Tidak ada:

```text
allow_skip
allow_cancel
audio_enabled
display_enabled
```

Alasan:

```text
skip = RBAC permission + journey state machine
cancel = RBAC permission + journey state machine
audio call = system behavior
display = system behavior
```

---

# 9. Branch Design

## 9.1 `branches`

```text
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
```

Tidak menggunakan `deleted_at`.

Jika branch tidak boleh dipakai, gunakan `status`.

## 9.2 Required Fields

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

## 9.3 Branch Logo Rule

Branch logo optional.

```text
effective_logo_asset_id =
  branch.logo_asset_id if exists
  else tenant.logo_asset_id
```

## 9.4 Running Text

`running_text` wajib.

Contoh:

```text
Selamat Datang di Rumah Sakit RS KITA SEMUA
```

Default saat branch dibuat boleh:

```text
Selamat Datang di {branch.name}
```

Tetapi tetap harus tersimpan di `branches.running_text`.

## 9.5 Branch Activation Rule

Branch tidak boleh `active` jika field berikut belum lengkap:

```text
name
address
city
province
phone
running_text
timezone
```

Logo branch tidak wajib selama tenant memiliki logo.

---

# 10. Branch Queue Settings

## 10.1 `branch_queue_settings`

```text
branch_queue_settings
- id
- tenant_id
- branch_id
- queue_reset_time nullable
- ticket_prefix nullable
- default_estimated_duration nullable
- allow_forward nullable
- allow_recall nullable
- auto_call_next nullable
- numbering_strategy nullable
- created_at
- updated_at
```

## 10.2 Nullable Means Inherit

Jika field bernilai `null`, gunakan value dari `tenant_queue_settings`.

Contoh:

```text
tenant_queue_settings.queue_reset_time = 04:00
branch_queue_settings.queue_reset_time = null

effective queue_reset_time = 04:00
source = tenant
```

Jika branch override:

```text
tenant_queue_settings.queue_reset_time = 04:00
branch_queue_settings.queue_reset_time = 05:00

effective queue_reset_time = 05:00
source = branch
```

## 10.3 Unique Constraint

```text
unique tenant_id + branch_id
```

---

# 11. Service Design

## 11.1 `services`

Service adalah template layanan milik tenant.

Contoh:

```text
Registration
Doctor
Pharmacy
Cashier
Laboratory
```

New table:

```text
services
- id
- tenant_id
- code
- name
- type
- is_pharmacy
- is_pharmacy_reception
- estimated_duration
- min_service_duration
- max_service_duration
- audio_id nullable
- audio_en nullable
- narrative_instruction_id nullable
- narrative_instruction_en nullable
- status
- created_at
- updated_at
```

Tidak menggunakan:

```text
deleted_at
default_estimated_duration
```

Gunakan:

```text
estimated_duration
```

agar tidak duplikatif dengan config lain.

## 11.2 Required Fields

```text
tenant_id
code
name
type
estimated_duration
min_service_duration
max_service_duration
status
```

## 11.3 Duration Rule

```text
min_service_duration <= estimated_duration <= max_service_duration
```

Tidak boleh:

```text
min_service_duration > estimated_duration
estimated_duration > max_service_duration
```

## 11.4 Service Type

```text
registration
doctor
pharmacy
cashier
laboratory
general
```

## 11.5 Pharmacy Flags

```text
is_pharmacy
is_pharmacy_reception
```

`is_pharmacy = true` untuk service farmasi.

`is_pharmacy_reception = true` untuk service yang menjadi pintu masuk valid ke alur farmasi, misalnya:

```text
Penerimaan Resep
```

## 11.6 Audio and Narrative

`audio_id` dan `audio_en` adalah asset reference untuk audio pemanggilan service.

`narrative_instruction_id` dan `narrative_instruction_en` adalah teks instruksi untuk signage/caller.

Contoh:

```text
Silakan menuju loket pendaftaran.
Please proceed to the registration counter.
```

Semua field audio/narrative optional.

## 11.7 Audio Fallback Rule

```text
if requested language = en and audio_en exists:
    use audio_en
else if audio_id exists:
    use audio_id
else:
    use text-to-speech or no service audio
```

---

# 12. Branch Service Design

## 12.1 `branch_services`

```text
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

## 12.2 Required Fields

```text
tenant_id
branch_id
service_id
is_active
sort_order
```

## 12.3 Unique Constraint

```text
unique tenant_id + branch_id + service_id
```

## 12.4 Purpose

`services` adalah template tenant-level.

`branch_services` menentukan service mana yang aktif di branch.

Queue creation, forwarding, counter, caller, dan signage harus memakai:

```text
branch_service_id
```

bukan `service_id` langsung.

## 12.5 Custom Name

`custom_name` dipakai untuk nama service khusus branch.

Contoh:

```text
services.name = Registration
branch_services.custom_name = Pendaftaran
```

Jika `custom_name` kosong, tampilkan `services.name`.

---

# 13. Branch Service Queue Settings

## 13.1 `branch_service_queue_settings`

```text
branch_service_queue_settings
- id
- tenant_id
- branch_id
- branch_service_id
- estimated_duration nullable
- min_service_duration nullable
- max_service_duration nullable
- require_counter nullable
- allow_forward_from nullable
- allow_forward_to nullable
- allow_recall nullable
- auto_call_next nullable
- audio_id nullable
- audio_en nullable
- narrative_instruction_id nullable
- narrative_instruction_en nullable
- created_at
- updated_at
```

## 13.2 Nullable Means Inherit

Jika field null, gunakan value dari `services`, `branch_queue_settings`, atau `tenant_queue_settings` sesuai field.

Contoh duration:

```text
services.estimated_duration = 15
branch_service_queue_settings.estimated_duration = null

effective estimated_duration = 15
source = service
```

Jika override:

```text
services.estimated_duration = 15
branch_service_queue_settings.estimated_duration = 20

effective estimated_duration = 20
source = branch_service
```

## 13.3 Unique Constraint

```text
unique tenant_id + branch_id + branch_service_id
```

---

# 14. Counter Design

## 14.1 `counters`

```text
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
```

Tidak menggunakan `deleted_at`.

Jika counter tidak boleh dipakai, gunakan:

```text
status = inactive
```

## 14.2 Required Fields

```text
tenant_id
branch_id
branch_service_id
code
name
display_name
status
```

## 14.3 Why Counter Uses `branch_service_id`

Counter harus menunjuk ke `branch_service_id`, bukan langsung `service_id`.

Alasannya:

```text
branch_service_id memastikan service tersebut aktif di branch tersebut.
```

Dengan ini, counter tidak bisa menunjuk service yang belum diaktifkan di branch.

## 14.4 Counter Validation

Counter valid jika:

```text
counter.tenant_id = request.tenant_id
counter.branch_id = request.branch_id
counter.branch_service_id belongs to same tenant + branch
counter.status = active
```

---

# 15. Counter Queue Settings

## 15.1 `counter_queue_settings`

```text
counter_queue_settings
- id
- tenant_id
- branch_id
- counter_id
- auto_call_next
- allow_recall
- created_at
- updated_at
```

## 15.2 Defaults

```text
auto_call_next = false
allow_recall = false
```

## 15.3 Unique Constraint

```text
unique tenant_id + branch_id + counter_id
```

## 15.4 Removed Fields

Tidak ada:

```text
audio_enabled
display_enabled
allow_skip
allow_cancel
```

Alasan:

```text
audio/display = system behavior
skip/cancel = RBAC permission + state machine
```

---

# 16. Queue Parent Design

## 16.1 `queues`

```text
queues
- id
- tenant_id
- branch_id
- queue_date
- ticket_no
- queue_no
- patient_ref nullable
- patient_name
- patient_phone nullable
- source
- priority
- status
- current_journey_id
- created_by nullable
- created_at
- updated_at
```

## 16.2 Queue Status

Queue parent hanya menyimpan lifecycle visit:

```text
waiting
in_progress
completed
cancelled
```

Operational status seperti:

```text
called
serving
skipped
forwarded
```

masuk ke `queue_journeys`.

## 16.3 Unique Constraints

```text
unique tenant_id + branch_id + queue_date + ticket_no
unique tenant_id + branch_id + queue_date + queue_no
```

## 16.4 Queue Rule

```text
1 queue = 1 ticket
1 queue = 1 queue number
1 queue = 1 visit
```

Forwarding tidak pernah membuat queue baru.

---

# 17. Queue Journey Design

## 17.1 `queue_journeys`

```text
queue_journeys
- id
- tenant_id
- branch_id
- queue_id
- branch_service_id
- counter_id nullable
- sequence_no
- status
- source_journey_id nullable
- forward_reason nullable

- called_at nullable
- last_called_at nullable
- call_count
- started_at nullable
- completed_at nullable
- skipped_at nullable
- cancelled_at nullable
- forwarded_at nullable

- created_by nullable
- created_at
- updated_at
```

## 17.2 Journey Status

```text
waiting
called
serving
completed
skipped
cancelled
forwarded
```

## 17.3 Active Journey Status

```text
waiting
called
serving
skipped
```

## 17.4 Terminal Journey Status

```text
completed
cancelled
forwarded
```

## 17.5 One Active Journey Rule

Satu queue hanya boleh memiliki satu active journey.

Rule:

```text
one queue_id can only have one journey with active status
```

Jika database mendukung partial unique index, enforce di DB.

Jika tidak, enforce dengan transaction lock:

```text
lock queue row
lock active journey row
validate no other active journey exists
```

---

# 18. Replacing Old State Timestamps

Old QMS memiliki timestamp generic seperti:

```text
State1Timestamp
State2Timestamp
State3Timestamp
State4Timestamp
State5Timestamp
```

New rebuild tidak memakai field generic tersebut.

Diganti dengan typed timestamps:

```text
called_at
last_called_at
started_at
completed_at
skipped_at
cancelled_at
forwarded_at
```

Dan untuk event history lengkap, gunakan:

```text
visit_journeys
```

## 18.1 Mapping Concept

```text
StartTimestamp   → queue_journeys.started_at
DoneTimestamp    → queue_journeys.completed_at
State timestamps → typed journey timestamps + visit_journeys
```

---

# 19. Queue Action State Machine

Caller menggunakan satu endpoint action:

```text
POST /api/v1/caller/queue-journeys/{journey_id}/action
```

Supported actions:

```text
call
start
forward
complete
skip
cancel
```

Tidak ada action:

```text
recall
```

Recall dilakukan dengan `action = call` pada journey yang statusnya sudah `called`.

---

## 19.1 Call

Request:

```json
{
  "action": "call"
}
```

Transition:

```text
waiting → called
```

Update:

```text
queue_journeys.status = called
queue_journeys.called_at = now
queue_journeys.last_called_at = now
queue_journeys.call_count = 1
visit_journeys event = queue_called
audit log = qms.queue.called
```

---

## 19.2 Internal Recall by Repeated Call

Request:

```json
{
  "action": "call"
}
```

If current status:

```text
called
```

then backend treats it as recall internally.

Transition:

```text
called → called
```

Update:

```text
called_at remains unchanged
last_called_at = now
call_count = call_count + 1
visit_journeys event = queue_recalled
audit log = qms.queue.recalled
```

Guard:

```text
operator has queue.call permission
effective allow_recall = true
journey status = called
```

New policy:

```text
queue.call covers first call and repeated call.
allow_recall controls whether repeated call is allowed.
```

---

## 19.3 Call Skipped Journey

Request:

```json
{
  "action": "call"
}
```

Allowed transition:

```text
skipped → called
```

Update:

```text
status = called
last_called_at = now
call_count = call_count + 1
visit_journeys event = queue_called
audit log = qms.queue.called
```

Jika `called_at` sebelumnya kosong, isi `called_at = now`.

Jika sebelumnya sudah pernah dipanggil, jangan ubah `called_at`.

---

## 19.4 Start

Request:

```json
{
  "action": "start"
}
```

Transition:

```text
called → serving
```

Update:

```text
queue_journeys.status = serving
queue_journeys.started_at = now
queues.status = in_progress
visit_journeys event = service_started
audit log = qms.queue.started
```

Guard:

```text
journey is current_journey_id
journey status is called
queue is not completed/cancelled
counter belongs to tenant + branch + branch_service
operator/client is bound to this counter
operator has queue.start permission
```

---

## 19.5 Forward

Request:

```json
{
  "action": "forward",
  "target_branch_service_id": "branch_service_doctor",
  "target_counter_id": null,
  "reason": "Need doctor consultation"
}
```

Current journey transition:

```text
serving → forwarded
```

Create next journey:

```text
status = waiting
branch_service_id = target_branch_service_id
counter_id = target_counter_id
source_journey_id = current journey id
```

Update:

```text
current journey status = forwarded
current journey forwarded_at = now
queues.current_journey_id = next_journey_id
queues.status = in_progress
visit_journeys event = queue_forwarded
audit log = qms.queue.forwarded
```

Forwarding must not:

```text
create new queue
generate new ticket_no
generate new queue_no
forward to inactive branch_service
forward across tenant
forward to invalid counter
create two active journeys
```

---

## 19.6 Complete

Request:

```json
{
  "action": "complete"
}
```

Transition:

```text
serving → completed
```

Update:

```text
queue_journeys.status = completed
queue_journeys.completed_at = now
queues.status = completed
visit_journeys event = service_completed
visit_journeys event = queue_completed
audit log = qms.queue.completed
```

For MVP, operator chooses one of:

```text
Forward
or
Complete Visit
```

Do not create ambiguous state such as completed journey then forward.

---

## 19.7 Skip

Request:

```json
{
  "action": "skip",
  "reason": "Patient not present"
}
```

Allowed transitions:

```text
waiting → skipped
called → skipped
```

Update:

```text
queue_journeys.status = skipped
queue_journeys.skipped_at = now
visit_journeys event = journey_skipped
audit log = qms.queue.skipped
```

Skipped journey remains callable:

```text
skipped → called
```

via `action = call`.

---

## 19.8 Cancel

Request:

```json
{
  "action": "cancel",
  "reason": "Patient cancelled visit"
}
```

Allowed transitions:

```text
waiting → cancelled
called → cancelled
serving → cancelled
skipped → cancelled
```

Update:

```text
queue_journeys.status = cancelled
queue_journeys.cancelled_at = now
queues.status = cancelled
visit_journeys event = queue_cancelled
audit log = qms.queue.cancelled
```

---

# 20. New State Transition Summary

```text
waiting --call--> called
called --call--> called/recalled internally
skipped --call--> called

called --start--> serving

serving --forward--> forwarded + next journey waiting
serving --complete--> completed + queue completed

waiting --skip--> skipped
called --skip--> skipped

waiting --cancel--> cancelled + queue cancelled
called --cancel--> cancelled + queue cancelled
serving --cancel--> cancelled + queue cancelled
skipped --cancel--> cancelled + queue cancelled
```

---

# 21. Estimate Time and Queue Left

Estimate time tetap penting untuk caller, signage, dan dashboard.

Namun `EstimateTime` dan `QueueLeft` tidak disimpan permanen di queue table.

Keduanya dihitung sebagai read model.

## 21.1 Queue Left Calculation

Untuk journey yang sedang waiting:

```text
queue_left =
count waiting queue_journeys before current queue
within same tenant
within same branch
within same branch_service
within same queue_date
with queue_no lower than current queue_no
```

Query concept:

```text
queue_journeys
JOIN queues ON queues.id = queue_journeys.queue_id

WHERE queue_journeys.tenant_id = ?
AND queue_journeys.branch_id = ?
AND queue_journeys.branch_service_id = ?
AND queues.queue_date = ?
AND queue_journeys.status = 'waiting'
AND queues.queue_no < current_queue.queue_no
```

## 21.2 Estimate Time Calculation

```text
estimate_time = queue_left × effective_estimated_duration
```

Effective estimated duration:

```text
branch_service_queue_settings.estimated_duration
fallback services.estimated_duration
```

Example:

```text
queue_left = 4
effective_estimated_duration = 5 minutes
estimate_time = 20 minutes
```

## 21.3 Response Example

```json
{
  "queue_id": "queue_001",
  "ticket_no": "A001",
  "queue_no": 1,
  "queue_left": 4,
  "estimate_time_seconds": 1200,
  "estimate_time_minutes": 20,
  "effective_estimated_duration_minutes": 5
}
```

## 21.4 Serving Journey

Jika journey sudah `serving`:

```text
queue_left = 0
estimate_time = 0
service_elapsed_time = now - started_at
```

---

# 22. Visit Journey Design

## 22.1 `visit_journeys`

```text
visit_journeys
- id
- tenant_id
- branch_id
- queue_id
- queue_journey_id nullable
- event_type
- title
- description
- metadata_json nullable
- actor_type nullable
- actor_id nullable
- occurred_at
- created_at
```

## 22.2 Event Types

```text
queue_created
journey_created
queue_called
queue_recalled
service_started
service_completed
queue_forwarded
journey_skipped
queue_cancelled
queue_completed
```

## 22.3 Purpose

Visit journey dipakai untuk:

```text
queue detail timeline
operational history
debugging queue flow
readable patient journey
```

Audit log tetap terpisah.

---

# 23. Queue Counter Design

## 23.1 `queue_counters`

```text
queue_counters
- id
- tenant_id
- branch_id
- queue_date
- prefix
- current_number
- created_at
- updated_at
```

## 23.2 Unique Constraint

```text
unique tenant_id + branch_id + queue_date + prefix
```

## 23.3 Queue Number Generation

Queue number hanya dibuat saat `queues` dibuat.

Forwarding tidak pernah generate queue number baru.

## 23.4 Atomic Create Queue

```text
BEGIN
  resolve effective queue config
  calculate queue_date
  lock/create queue_counter
  increment queue_counter
  create queues
  create first queue_journey
  update queues.current_journey_id
  create visit_journey queue_created
  create audit log qms.queue.created
COMMIT
```

---

# 24. Forward Transaction

Forwarding harus berjalan dalam transaction.

```text
BEGIN
  lock queue row
  lock active journey row
  validate current journey
  validate target branch_service active
  validate target counter if provided
  mark current journey forwarded
  create next queue_journey
  update queues.current_journey_id
  insert visit_journey queue_forwarded
  insert audit log qms.queue.forwarded
COMMIT
```

---

# 25. UI Architecture

QMS memiliki 3 UI utama:

```text
Dashboard / Manage UI
Caller UI
Signage UI
```

---

# 26. Dashboard / Manage UI

Dashboard dipakai oleh admin/manager.

Fungsi:

```text
manage tenant
manage branch
manage services
enable service for branch
manage counters
manage qms clients
manage operator assignments
view queue dashboard
view visit journey
view audit log
```

Dashboard boleh memiliki branch/service/counter selector karena admin bisa punya akses luas.

Auth dashboard:

```text
user login
JWT/session
RBAC permission
tenant context
branch access if needed
```

---

# 27. Caller UI

Caller dipakai operator counter.

Prinsip utama:

```text
operator tidak memilih tenant
operator tidak memilih branch
operator tidak memilih service
operator tidak memilih counter
```

Caller harus langsung resolve context:

```text
tenant
branch
branch_service
counter
operator identity
permissions
```

---

# 28. Operator Counter Assignment

## 28.1 `operator_counter_assignments`

```text
operator_counter_assignments
- id
- tenant_id
- branch_id
- branch_service_id
- counter_id
- user_id
- is_active
- valid_from nullable
- valid_until nullable
- created_at
- updated_at
```

## 28.2 Purpose

Table ini menentukan operator mana yang boleh login sebagai caller untuk counter tertentu.

## 28.3 Assignment Rule

Operator caller valid jika:

```text
user belongs to tenant
user has access to branch
assignment is active
counter belongs to branch
counter belongs to branch_service
branch_service is active
```

---

# 29. QMS Client Design

`qms_clients` menggantikan konsep device lama untuk MVP.

Dipakai untuk:

```text
caller
signage
scanner future
kiosk future
```

---

## 29.1 `qms_clients`

```text
qms_clients
- id
- tenant_id
- branch_id
- branch_service_id nullable
- counter_id nullable
- client_id
- client_type
- name
- status
- allowed_actions_json nullable
- last_used_at nullable
- created_at
- updated_at
```

## 29.2 Client Type

```text
caller
signage
scanner
kiosk
```

Untuk MVP:

```text
caller
signage
```

## 29.3 Caller Client Rule

Untuk `client_type = caller`:

```text
branch_service_id required
counter_id required
```

## 29.4 Signage Client Rule

Untuk `client_type = signage`:

```text
branch_id required
branch_service_id required if signage is service-specific
counter_id nullable
```

Signage bisa menampilkan:

```text
one service with multiple counters
or one specific counter
```

---

# 30. QMS Client Credentials

## 30.1 `qms_client_credentials`

```text
qms_client_credentials
- id
- qms_client_id
- api_key_hash
- key_prefix
- status
- expires_at nullable
- rotated_at nullable
- created_at
- updated_at
```

## 30.2 Security Rule

Raw API key tidak pernah disimpan.

Simpan hanya:

```text
api_key_hash
key_prefix
```

Raw API key hanya ditampilkan sekali saat dibuat.

Header:

```http
X-Client-ID: caller-reg-01
X-API-Key: secret
```

---

# 31. Caller Authentication Design

Caller menggunakan hybrid authentication.

## 31.1 Step 1 — Client Credential Resolves Counter Context

Caller app mengirim:

```http
X-Client-ID
X-API-Key
```

Backend resolve:

```text
qms_client.client_type = caller
qms_client.status = active
tenant_id
branch_id
branch_service_id
counter_id
```

## 31.2 Step 2 — Human Operator Login

Operator login menggunakan credential user.

Backend validasi:

```text
user belongs to tenant
user has branch access
user has active operator_counter_assignment
assignment counter matches qms_client.counter_id
user has caller permissions
```

## 31.3 Caller Session Context

Response caller login:

```json
{
  "access_token": "jwt",
  "context": {
    "tenant_id": "tenant_001",
    "tenant_name": "RS Kita Semua",
    "branch_id": "branch_001",
    "branch_name": "Rawat Jalan",
    "branch_service_id": "branch_service_registration",
    "service_name": "Pendaftaran",
    "counter_id": "counter_reg_01",
    "counter_name": "Registration Counter 1",
    "display_name": "Loket Pendaftaran 1"
  },
  "permissions": [
    "queue.read",
    "queue.call",
    "queue.start",
    "queue.forward",
    "queue.complete",
    "queue.skip",
    "queue.cancel"
  ]
}
```

---

# 32. Caller API

Caller API tidak membutuhkan branch/counter selector dari UI.

Backend menggunakan session context.

## 32.1 Caller Endpoints

```text
POST /api/v1/caller/login
GET  /api/v1/caller/me
GET  /api/v1/caller/queue-journeys
GET  /api/v1/caller/current

POST /api/v1/caller/queue-journeys/{journey_id}/action
```

## 32.2 Supported Actions

```text
call
start
forward
complete
skip
cancel
```

Tidak ada:

```text
recall
```

## 32.3 Recall Rule

```text
Recall is not a separate action.
Calling an already called journey is treated as recall internally.
Repeated call requires effective allow_recall = true.
```

## 32.4 Backend Context Filter

Backend always filters by:

```text
session.tenant_id
session.branch_id
session.branch_service_id
session.counter_id
```

Caller tidak boleh mengirim atau override:

```text
tenant_id
branch_id
branch_service_id
counter_id
```

di request body.

---

# 33. Signage Authentication Design

Signage tidak membutuhkan human login.

Signage menggunakan:

```http
X-Client-ID
X-API-Key
```

Backend resolve:

```text
tenant
branch
branch_service
counter optional
display config
```

---

# 34. Signage API

```text
GET /api/v1/signage/me
GET /api/v1/signage/current-calls
GET /api/v1/signage/queues
```

## 34.1 `/signage/me` Response

```json
{
  "client_type": "signage",
  "tenant": {
    "id": "tenant_001",
    "name": "RS Kita Semua",
    "logo_asset_id": "asset_logo"
  },
  "branch": {
    "id": "branch_001",
    "name": "Rawat Jalan",
    "running_text": "Selamat Datang di Rumah Sakit RS KITA SEMUA",
    "effective_logo_asset_id": "asset_logo"
  },
  "service": {
    "branch_service_id": "branch_service_registration",
    "name": "Pendaftaran"
  },
  "counters": [
    {
      "id": "counter_reg_01",
      "display_name": "Loket 1"
    },
    {
      "id": "counter_reg_02",
      "display_name": "Loket 2"
    }
  ]
}
```

## 34.2 `/signage/current-calls` Response

```json
{
  "current_calls": [
    {
      "ticket_no": "A001",
      "counter_display_name": "Loket 1",
      "service_name": "Pendaftaran",
      "called_at": "2026-07-01T10:00:00+07:00",
      "last_called_at": "2026-07-01T10:01:00+07:00",
      "call_count": 2,
      "audio": {
        "service_audio_id": "asset_audio_pendaftaran",
        "narrative_instruction_id": "Silakan menuju loket pendaftaran."
      }
    }
  ],
  "waiting": [
    {
      "ticket_no": "A002",
      "queue_left": 0,
      "estimate_time_minutes": 0
    },
    {
      "ticket_no": "A003",
      "queue_left": 1,
      "estimate_time_minutes": 5
    }
  ]
}
```

---

# 35. Credential Security Rules

Wajib:

```text
X-Client-ID required
X-API-Key required
API key stored hashed
API key never logged
API key never returned
client_type must match endpoint
client status must be active
client tenant/branch relation must be valid
counter must belong to branch_service
branch_service must be active
```

Do not log:

```text
raw API key
JWT token
password
full request body
sensitive patient data
```

---

# 36. Effective Configuration Resolution

New hierarchy:

```text
System Default
  ↓
tenant_queue_settings
  ↓
branch_queue_settings
  ↓
services
  ↓
branch_service_queue_settings
  ↓
counter_queue_settings
```

## 36.1 Queue Reset Time

```text
tenant_queue_settings.queue_reset_time
↓
branch_queue_settings.queue_reset_time
```

## 36.2 Ticket Prefix

```text
tenant_queue_settings.default_ticket_prefix
↓
branch_queue_settings.ticket_prefix
```

## 36.3 Estimated Duration

```text
services.estimated_duration
↓
branch_service_queue_settings.estimated_duration
```

## 36.4 Audio

```text
services.audio_id / audio_en
↓
branch_service_queue_settings.audio_id / audio_en
```

## 36.5 Narrative

```text
services.narrative_instruction_id / narrative_instruction_en
↓
branch_service_queue_settings.narrative_instruction_id / narrative_instruction_en
```

## 36.6 Auto Call Next

```text
tenant_queue_settings.auto_call_next
↓
branch_queue_settings.auto_call_next
↓
branch_service_queue_settings.auto_call_next
↓
counter_queue_settings.auto_call_next
```

## 36.7 Allow Recall

```text
tenant_queue_settings.allow_recall
↓
branch_queue_settings.allow_recall
↓
branch_service_queue_settings.allow_recall
↓
counter_queue_settings.allow_recall
```

---

# 37. Dashboard MVP

Dashboard minimal menampilkan:

```text
Total queues today
Waiting journeys
Called journeys
Serving journeys
Completed queues
Skipped journeys
Cancelled queues
Average service duration
Average waiting estimate
```

Breakdown:

```text
by branch
by branch_service
by counter
by queue_date
```

Data source:

```text
queues = visit count
queue_journeys = operational status
visit_journeys = timeline/history
```

Stats wajib scoped by:

```text
tenant_id
branch_id
queue_date
```

`queue_date` dihitung dari effective `queue_reset_time`.

---

# 38. Audit Log Events

## 38.1 Setup Events

```text
qms.tenant.profile.updated
qms.tenant.queue_settings.updated

qms.branch.profile.updated
qms.branch.running_text.updated
qms.branch.queue_settings.updated

qms.service.created
qms.service.updated
qms.service.inactivated
qms.service.deleted
qms.service.audio.updated
qms.service.narrative.updated
qms.service.duration.updated

qms.branch_service.enabled
qms.branch_service.disabled
qms.branch_service.updated
qms.branch_service.queue_settings.updated

qms.counter.created
qms.counter.updated
qms.counter.inactivated
qms.counter.deleted
qms.counter.queue_settings.updated

qms.qms_client.created
qms.qms_client.updated
qms.qms_client.credential.rotated

qms.operator_counter_assignment.created
qms.operator_counter_assignment.updated
qms.operator_counter_assignment.deactivated
```

## 38.2 Queue Events

```text
qms.queue.created
qms.queue.called
qms.queue.recalled
qms.queue.started
qms.queue.forwarded
qms.queue.completed
qms.queue.skipped
qms.queue.cancelled
```

## 38.3 Audit Metadata

```json
{
  "tenant_id": "tenant_001",
  "branch_id": "branch_001",
  "actor_id": "user_001",
  "actor_type": "operator",
  "resource_type": "queue_journey",
  "resource_id": "journey_001",
  "action": "qms.queue.called",
  "metadata": {
    "queue_id": "queue_001",
    "ticket_no": "A001",
    "counter_id": "counter_reg_01"
  }
}
```

Untuk logo/audio:

```json
{
  "old_asset_id": "asset_old",
  "new_asset_id": "asset_new"
}
```

Jangan simpan raw binary file di audit log.

---

# 39. Error Logging Requirements

All QMS error paths must log structured context.

## 39.1 Required Context

```text
module
action
tenant_id
branch_id
branch_service_id if available
counter_id if available
user_id if available
qms_client_id if available
request_id / trace_id
resource_id
safe error message
```

## 39.2 Required Error Logging Areas

```text
tenant profile update failed
tenant activation failed
branch profile update failed
branch running_text missing
service setup failed
service duration invalid
branch service setup failed
counter setup failed
qms client auth failed
operator assignment validation failed
caller login failed
signage credential auth failed
queue creation failed
queue counter lock failed
call action failed
repeated call / recall failed
start service failed
forward failed
complete failed
skip failed
cancel failed
estimate calculation failed
visit journey creation failed
audit log creation failed
```

## 39.3 Sensitive Data Rule

Do not log:

```text
raw API key
JWT token
password
full request body
sensitive patient data
raw uploaded file data
```

---

# 40. API Summary

## 40.1 Dashboard / Manage

```text
GET/PATCH /api/v1/tenant/profile
GET/PATCH /api/v1/tenant/queue-config

GET/PATCH /api/v1/branches/{branch_id}/profile
GET/PATCH /api/v1/branches/{branch_id}/queue-config

POST/GET/PATCH/DELETE /api/v1/services

GET/POST/PATCH /api/v1/branches/{branch_id}/services
GET/PATCH /api/v1/branches/{branch_id}/branch-services/{branch_service_id}/queue-config

POST/GET/PATCH/DELETE /api/v1/branches/{branch_id}/counters
GET/PATCH /api/v1/branches/{branch_id}/counters/{counter_id}/queue-config

POST/GET/PATCH /api/v1/qms-clients
POST /api/v1/qms-clients/{client_id}/rotate-key

POST/GET/PATCH /api/v1/operator-counter-assignments
```

## 40.2 Queue Operation

```text
POST /api/v1/branches/{branch_id}/queues
GET  /api/v1/branches/{branch_id}/queues
GET  /api/v1/branches/{branch_id}/queues/{queue_id}
GET  /api/v1/branches/{branch_id}/queues/{queue_id}/visit-journeys
```

## 40.3 Caller

```text
POST /api/v1/caller/login
GET  /api/v1/caller/me
GET  /api/v1/caller/queue-journeys
GET  /api/v1/caller/current

POST /api/v1/caller/queue-journeys/{journey_id}/action
```

Supported action body:

```json
{
  "action": "call"
}
```

Forward action body:

```json
{
  "action": "forward",
  "target_branch_service_id": "branch_service_doctor",
  "target_counter_id": null,
  "reason": "Need doctor consultation"
}
```

Skip action body:

```json
{
  "action": "skip",
  "reason": "Patient not present"
}
```

Cancel action body:

```json
{
  "action": "cancel",
  "reason": "Patient cancelled visit"
}
```

## 40.4 Signage

```text
GET /api/v1/signage/me
GET /api/v1/signage/current-calls
GET /api/v1/signage/queues
```

---

# 41. Migration Strategy

## Step 1 — Remove Generic Settings

Remove generic settings usage completely.

```text
Drop table settings
Remove all code references
Remove fallback logic
Ensure all config resolved via typed tables
```

## Step 2 — Create Typed Config Tables

Create:

```text
tenant_queue_settings
branch_queue_settings
branch_service_queue_settings
counter_queue_settings
```

## Step 3 — Add Tenant Profile Fields

Add:

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

## Step 4 — Add Branch Profile Fields

Add:

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

## Step 5 — Update Services

Add:

```text
estimated_duration
min_service_duration
max_service_duration
audio_id nullable
audio_en nullable
narrative_instruction_id nullable
narrative_instruction_en nullable
```

Remove or stop using:

```text
default_estimated_duration
deleted_at
```

## Step 6 — Create Branch Service

Create:

```text
branch_services
```

Update:

```text
counters.branch_service_id
queue_journeys.branch_service_id
```

## Step 7 — Create Client and Assignment Tables

Create:

```text
qms_clients
qms_client_credentials
operator_counter_assignments
```

## Step 8 — Timestamp Refactor

Add to `queue_journeys`:

```text
last_called_at
call_count
forwarded_at
```

Ensure existing timestamp fields are typed:

```text
called_at
started_at
completed_at
skipped_at
cancelled_at
```

Do not create:

```text
state1_timestamp
state2_timestamp
state3_timestamp
state4_timestamp
state5_timestamp
```

---

# 42. Testing Requirements

## 42.1 Tenant Tests

```text
tenant cannot activate without address
tenant cannot activate without city
tenant cannot activate without province
tenant cannot activate without phone
tenant cannot activate without logo_asset_id
tenant queue settings default values are created
tenant profile update writes audit log
tenant queue settings update writes audit log
```

## 42.2 Branch Tests

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

## 42.3 Service Tests

```text
service requires estimated_duration
service requires min_service_duration
service requires max_service_duration
service rejects min greater than estimated
service rejects estimated greater than max
service audio is optional
service narrative is optional
service update writes audit log
service cannot hard delete if already used
service can be inactivated if already used
```

## 42.4 Branch Service Tests

```text
service can be enabled for branch
service cannot be enabled twice for same branch
inactive branch service cannot be used for queue creation
branch service override changes effective config
branch service reset returns service template value
```

## 42.5 Counter Tests

```text
counter requires branch_service_id
counter rejects branch_service from another branch
counter rejects branch_service from another tenant
counter_queue_settings defaults auto_call_next false
counter_queue_settings defaults allow_recall false
counter queue settings update writes audit log
```

## 42.6 Estimate Tests

```text
queue_left counts waiting journeys before current queue
queue_left filters by tenant
queue_left filters by branch
queue_left filters by branch_service
estimate_time = queue_left × effective duration
serving journey returns queue_left 0
```

## 42.7 Timestamp Tests

```text
call sets called_at, last_called_at, call_count 1
repeated call updates last_called_at and increments call_count
repeated call does not change called_at
start sets started_at
complete sets completed_at
forward sets forwarded_at
skip sets skipped_at
cancel sets cancelled_at
```

## 42.8 Caller Tests

```text
caller client credential resolves tenant/branch/branch_service/counter
caller rejects inactive client
caller rejects wrong client_type
operator login rejects user without assignment
operator login rejects assignment to different counter
caller queue list filters by session context
caller cannot act on journey from another counter/service/branch/tenant
caller action endpoint accepts call/start/forward/complete/skip/cancel
caller action endpoint rejects unsupported action
repeated call requires allow_recall true
```

## 42.9 Signage Tests

```text
signage client credential resolves tenant/branch/service
signage rejects caller client_type
signage feed filters by branch_service
signage feed includes current called queues
signage feed includes running_text and effective logo
```

## 42.10 Queue Flow Tests

```text
create queue creates one queue
create queue creates first queue_journey
call transitions waiting to called
repeated call acts as recall
start transitions called to serving
forward does not create new queue
forward creates new queue_journey
complete marks queue completed
skip transitions journey to skipped
cancel marks queue cancelled
visit_journeys contain all key events
audit_logs contain all key actions
```

---

# 43. New Architecture Decisions

## ADR-001 — No Submenu for MVP

Submenu is removed from MVP.

Service is the main operational unit.

## ADR-002 — No Device Domain for MVP

Old device concept is replaced by `qms_clients`.

## ADR-003 — No Arrive for MVP

Arrive/check-in is not part of MVP.

## ADR-004 — Estimate Time Is Calculated

`EstimateTime` and `QueueLeft` are calculated read models, not stored permanently.

## ADR-005 — No Generic State Timestamp

`State1Timestamp` to `State5Timestamp` are not used.

Use typed timestamps and visit events.

## ADR-006 — Caller Has Bound Context

Caller login must resolve tenant, branch, branch_service, and counter automatically.

No manual selector in Caller UI.

## ADR-007 — Signage Uses Client Credential

Signage uses `X-Client-ID` and `X-API-Key`.

No human login required.

## ADR-008 — API Key Is Hashed

Raw API key is never stored or logged.

## ADR-009 — Forwarding Uses Queue Journey

Forwarding never creates a queue.

It creates a new `queue_journey`.

## ADR-010 — Caller Uses Single Action Endpoint

Caller journey operations use:

```text
POST /api/v1/caller/queue-journeys/{journey_id}/action
```

instead of multiple action-specific endpoints.

## ADR-011 — Recall Is Repeated Call

Recall is not a separate action.

Repeated `call` on an already called journey is treated as recall internally.

---

# 44. New Summary

New MVP design:

```text
Dashboard manages setup and monitoring.
Caller operates one bound counter/service context.
Signage displays one bound branch/service/counter context.

Tenant and branch are typed profile entities.
Settings are typed tables only.
Generic settings is removed.

Service owns duration, audio, and narrative defaults.
Branch service activates service per branch and can override behavior.
Counter belongs to branch_service.
Counter settings only contain auto_call_next and allow_recall.

Queue parent remains one ticket/visit.
Queue journey tracks service movement.
Visit journey tracks readable event timeline.

EstimateTime is calculated from QueueLeft and effective service duration.
Old state timestamps are replaced by typed timestamps.

Caller and signage use credential binding so they immediately know tenant, branch, service, and counter.

Caller uses one action endpoint.
Recall is handled by repeated call.
```

New operational API shape:

```text
POST /api/v1/caller/queue-journeys/{journey_id}/action
```

Supported actions:

```text
call
start
forward
complete
skip
cancel
```

New behavior:

```text
waiting --call--> called
called --call--> called/recalled internally
skipped --call--> called

called --start--> serving

serving --forward--> forwarded + next journey waiting
serving --complete--> completed + queue completed

waiting --skip--> skipped
called --skip--> skipped

waiting --cancel--> cancelled + queue cancelled
called --cancel--> cancelled + queue cancelled
serving --cancel--> cancelled + queue cancelled
skipped --cancel--> cancelled + queue cancelled
```
