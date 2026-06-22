# 1. Diagram Arsitektur Rewrite

```mermaid
flowchart TD
    A[HTTP Request] --> B[Echo Router]

    B --> C[Middleware Layer]
    C --> C1[JWT Auth]
    C --> C2[Casbin RBAC]
    C --> C3[Scanner Auth]
    C --> C4[Request Logger]
    C --> C5[Error Mapper]

    C --> D[Presentation / Handler Layer]

    D --> E[Business Orchestration Layer]

    E --> F1[QueueService]
    E --> F2[ForwardService]
    E --> F3[PharmacyFlowValidator]
    E --> F4[ScannerCheckInService]
    E --> F5[QueueStateMachine]
    E --> F6[Notification Dispatcher]

    F1 --> G[Repository Layer]
    F2 --> G
    F3 --> G
    F4 --> G
    F5 --> G

    G --> H[(MySQL / GORM)]

    F6 --> I[Outbox / Post Commit Job]
    I --> J[Bithealth / Notification API]

    F4 --> K[Medifrans Integration]
    K --> L[Appointment / Patient / Visit Journey APIs]
```

**Makna rewrite:**
Handler tidak lagi menyimpan logic forward, queue, pharmacy, atau scanner secara langsung. Handler hanya menerima request, validasi input awal, lalu memanggil service yang tepat.

---

# 2. Diagram Struktur Module Baru

```mermaid
flowchart LR
    A[features/customers] --> A1[business/queue_service.go]
    A --> A2[business/forward_service.go]
    A --> A3[business/pharmacy_validator.go]
    A --> A4[business/queue_state_machine.go]
    A --> A5[business/errors.go]

    A --> B1[data/customer_repository.go]
    A --> B2[data/queue_repository.go]
    A --> B3[data/queue_counter_repository.go]

    C[integration/medifrans] --> C1[service/scanner_checkin_service.go]
    C --> C2[repository/scannerClient_repository.go]
    C --> C3[routes/medifrans_middleware.go]
    C --> C4[handler/scanner_handler.go]

    D[shared/domain] --> D1[business_day.go]
    D --> D2[response.go]
    D --> D3[domain_errors.go]
```

---

# 3. Diagram Flow Register Queue Baru

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant QueueService
    participant Repo
    participant DB

    Client->>Handler: POST /customer
    Handler->>QueueService: RegisterQueue(request)

    QueueService->>QueueService: Validate request
    QueueService->>QueueService: Calculate business day 04:00 Asia/Jakarta
    QueueService->>Repo: Find active duplicate queue
    Repo->>DB: SELECT active queue by patient/appointment/menu/day

    alt Duplicate active queue exists
        DB-->>Repo: Existing queue
        Repo-->>QueueService: Existing queue
        QueueService-->>Handler: Return existing queue idempotently
        Handler-->>Client: 200 OK existing queue
    else No duplicate
        QueueService->>DB: BEGIN TRANSACTION
        QueueService->>Repo: Lock queue counter
        Repo->>DB: SELECT counter FOR UPDATE
        QueueService->>Repo: Increment counter
        Repo->>DB: UPDATE queue_counter
        QueueService->>Repo: Create customer queue
        Repo->>DB: INSERT customers
        QueueService->>DB: COMMIT
        QueueService-->>Handler: New queue response
        Handler-->>Client: 201 Created
    end
```

**Tujuan:** mencegah nomor antrian dobel ketika banyak request masuk bersamaan.

---

# 4. Diagram Flow Forward Queue Baru

```mermaid
flowchart TD
    A[Request Forward Queue] --> B[Validate Request]
    B --> C[BEGIN TRANSACTION]

    C --> D[Lock Source Customer]
    D --> E{Source Exists?}

    E -- No --> E1[Rollback + 404]
    E -- Yes --> F{Source Status Valid?}

    F -- No --> F1[Rollback + 409 Invalid Transition]
    F -- Yes --> G[Load Source Menu + Target Menu]

    G --> H[Validate Target Branch/Menu/Device]
    H --> I[Validate Pharmacy Rule]

    I --> J{Allowed?}
    J -- No --> J1[Rollback + 400 Invalid Pharmacy Flow]
    J -- Yes --> K[Calculate Queue Business Day]

    K --> L[Check Existing Destination Queue]
    L --> M{Destination Exists?}

    M -- Yes --> N[Mark Source Done/Continued If Safe]
    N --> O[COMMIT]
    O --> P[Return Existing Destination Queue]

    M -- No --> Q[Generate Queue Number With Lock]
    Q --> R[Create New Destination Queue]
    R --> S[Mark Source Queue Done + Continued]
    S --> T[COMMIT]
    T --> U[Send Notification After Commit]
    U --> V[Return New Queue Response]
```

---

# 5. Diagram Transaction Boundary untuk Forward

```mermaid
flowchart LR
    subgraph TX[Database Transaction]
        A[Lock Source Customer]
        B[Validate Source Status]
        C[Validate Target Menu]
        D[Validate Pharmacy Rule]
        E[Check Duplicate Destination]
        F[Generate Queue Number]
        G[Create Destination Queue]
        H[Update Source Queue Done/Continued]
    end

    I[Send Notification] -. after commit .-> J[Bithealth Notification API]

    TX --> K[Commit]
    K --> I

    TX --> L[Rollback on Error]
```

**Prinsip penting:** notification tidak boleh dikirim sebelum `COMMIT`, supaya tidak ada kasus user dapat notifikasi tapi data queue gagal tersimpan.

---

# 6. Diagram Scanner Check-In Rewrite

```mermaid
sequenceDiagram
    participant Scanner
    participant ScannerMiddleware
    participant ScannerService
    participant ForwardService
    participant QueueService
    participant Repo
    participant DB

    Scanner->>ScannerMiddleware: Request with x-client-id + x-api-key
    ScannerMiddleware->>Repo: Validate client_id + api_key
    Repo->>DB: SELECT client by client_id
    DB-->>Repo: Client + branch_id

    alt Invalid Credential
        ScannerMiddleware-->>Scanner: 401 Unauthorized
    else Valid Credential
        ScannerMiddleware->>ScannerService: Continue request
        ScannerService->>Repo: Validate branch-station-menu relation
        Repo->>DB: Check client branch + menu relation

        ScannerService->>Repo: Find appointment/patient active queue
        Repo->>DB: SELECT active queue

        alt Already checked-in at target station
            ScannerService-->>Scanner: Return existing queue
        else Needs forward
            ScannerService->>ForwardService: Forward to target station/menu
            ForwardService->>DB: Atomic forward transaction
            ForwardService-->>ScannerService: New queue
            ScannerService-->>Scanner: Queue response
        else New registration
            ScannerService->>QueueService: Register new queue
            QueueService->>DB: Atomic queue creation
            QueueService-->>ScannerService: New queue
            ScannerService-->>Scanner: Queue response
        end
    end
```

**Perubahan utama:** scanner tidak membuat logic forward sendiri. Scanner hanya menentukan kasusnya, lalu memanggil `ForwardService` atau `QueueService`.

---

# 7. Diagram Pharmacy Validation Baru

```mermaid
flowchart TD
    A[Validate Pharmacy Flow] --> B[Load Source Menu]
    B --> C[Load Target Menu]

    C --> D{Target Is Pharmacy?}

    D -- No --> D1[Allow Forward]
    D -- Yes --> E{Source Is Pharmacy?}

    E -- Yes --> E1[Allow Pharmacy to Pharmacy / Pharmacy to Any]
    E -- No --> F{Source Code == PENERIMAAN_RESEP?}

    F -- Yes --> F1[Allow]
    F -- No --> G[Check Customer History]

    G --> H{History Contains PENERIMAAN_RESEP?}
    H -- Yes --> H1[Allow]
    H -- No --> H2[Reject: Must Pass Penerimaan Resep]
```

**Yang diperbaiki:**
Jangan lagi pakai `regexp (?i)resep` sebagai sumber kebenaran. Di code sekarang, menu apa pun yang mengandung kata “resep” bisa dianggap valid. Lebih aman pakai `menu.Code == "PENERIMAAN_RESEP"` atau field `is_pharmacy_reception`.

---

# 8. Diagram Queue State Machine

```mermaid
stateDiagram-v2
    [*] --> not_called

    not_called --> called
    not_called --> missed
    not_called --> cancel

    called --> start
    called --> missed
    called --> done

    start --> done
    start --> cancel

    missed --> called
    missed --> cancel

    done --> [*]
    cancel --> [*]
```

**Tujuan:** semua update status harus lewat state machine ini, bukan bebas update status langsung dari handler atau repository.

---

# 9. Diagram Queue Number Counter

```mermaid
sequenceDiagram
    participant QueueService
    participant DB

    QueueService->>DB: BEGIN
    QueueService->>DB: SELECT queue_counter FOR UPDATE
    DB-->>QueueService: current_number

    QueueService->>DB: UPDATE queue_counter SET current_number = current_number + 1
    QueueService->>DB: INSERT customers(queue = next_number)

    alt Success
        QueueService->>DB: COMMIT
    else Error
        QueueService->>DB: ROLLBACK
    end
```

**Recommended table:**

```sql
CREATE TABLE queue_counters (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    branch_id BIGINT NOT NULL,
    menu_id BIGINT NOT NULL,
    business_date DATE NOT NULL,
    current_number INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uq_queue_counter_day (branch_id, menu_id, business_date)
);
```

---

# 10. Diagram Full Rewrite Flow

```mermaid
flowchart TD
    A[Start Rewrite] --> B[Phase 1: Safe Fixes]
    B --> B1[Business Day Helper]
    B --> B2[Queue State Machine]
    B --> B3[Domain Errors]
    B --> B4[Fix Queue Estimate]
    B --> B5[Fix Pharmacy Validator]

    B --> C[Phase 2: Transactional Forward]
    C --> C1[ForwardService]
    C --> C2[Duplicate Destination Check]
    C --> C3[Atomic Source Done + Destination Create]
    C --> C4[Post Commit Notification]

    C --> D[Phase 3: Queue Counter]
    D --> D1[queue_counters Table]
    D --> D2[SELECT FOR UPDATE]
    D --> D3[Unique Index]
    D --> D4[Concurrent Queue Test]

    D --> E[Phase 4: Scanner Cleanup]
    E --> E1[Validate Scanner Credential]
    E --> E2[Validate Branch Station]
    E --> E3[Use ForwardService]
    E --> E4[Idempotent Retry]

    E --> F[Phase 5: Testing]
    F --> F1[Forward Tests]
    F --> F2[Queue Tests]
    F --> F3[Pharmacy Tests]
    F --> F4[Scanner Tests]
    F --> F5[Race Condition Tests]
```

---

# 11. Diagram Before vs After

```mermaid
flowchart LR
    subgraph BEFORE[Before]
        A1[Handler / Business Mixed Logic]
        A2[Forward Logic Duplicated]
        A3[Queue Number Risk Race]
        A4[Pharmacy Regex]
        A5[Estimate Return 1]
        A6[Scanner Header Presence Only]
    end

    subgraph AFTER[After Rewrite]
        B1[QueueService]
        B2[ForwardService]
        B3[Queue Counter Transaction]
        B4[PharmacyFlowValidator with Code/Flag]
        B5[Real Estimate Calculation]
        B6[Scanner Credential + Station Validation]
    end

    A1 --> B1
    A2 --> B2
    A3 --> B3
    A4 --> B4
    A5 --> B5
    A6 --> B6
```

---

# 12. Diagram Data Ownership

```mermaid
flowchart TD
    A[Handler] -->|DTO only| B[Business Service]
    B -->|Domain Core| C[Repository Interface]
    C -->|GORM Model| D[(MySQL)]

    B --> E[Domain Validator]
    B --> F[State Machine]
    B --> G[Business Day Helper]

    D -. no business logic .-> C
    A -. no DB query .-> B
```

**Rule rewrite:**
Repository hanya query database. Business rule seperti forward, pharmacy, status transition, dan queue estimate harus pindah ke service/domain layer.
