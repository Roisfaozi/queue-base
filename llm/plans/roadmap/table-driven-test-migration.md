# Table-Driven Test Migration Plan — Integration & E2E

## 1. Why Table-Driven?

| Problem | Solution |
|---|---|
| 30+ test functions di `queue_usecase_test.go`, masing-masing 15-20 baris boilerplate setup | Satu struct slice, satu `t.Run()` loop |
| Edge case sering terlewat karena developer harus menulis fungsi baru setiap kali | Tambah 1 entry ke slice tabel = 1 test case baru, 5 detik kerja |
| Duplikasi `stubQueueRepo{...}` dan `NewQueueUseCase(...)` identik di semua fungsi | Setup dalam tabel, shared sebelum loop |
| Sulit membaca semua kasus untuk 1 metode sekaligus | Tabel jadi dokumentasi langsung dari behavior metode |
| QMS TDD Rule: positive, negative, edge, vulnerability wajib | Tabel dengan kolom `category` auto-audit coverage |

## 2. Go Table-Driven Pattern — Standar Proyek Ini

```go
func TestRegisterQueue(t *testing.T) {
    tests := []struct {
        name     string
        req      *model.RegisterQueueRequest
        stub     *stubQueueRepo
        settings map[string]string
        tenantID string
        branchID string
        wantErr  error
        wantRes  func(t *testing.T, res *entity.Queue)
    }{
        {
            name:     "Positive_CreatesQueue",
            req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John Doe"},
            tenantID: "t-1", branchID: "b-1",
            wantRes: func(t *testing.T, res *entity.Queue) {
                assert.Equal(t, "t-1", res.TenantID)
                assert.Equal(t, entity.QueueStatusWaiting, res.Status)
            },
        },
        {
            name:     "Negative_NoTenant",
            req:      &model.RegisterQueueRequest{ServiceID: "s-1"},
            branchID: "b-1",
            wantErr:  exception.ErrBadRequest,
        },
        {
            name:     "Negative_Duplicate",
            req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John"},
            stub:     &stubQueueRepo{exists: true},
            tenantID: "t-1", branchID: "b-1",
            wantErr: exception.ErrConflict,
        },
        {
            name:     "Edge_InvalidNumberingFallsBack",
            req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John"},
            settings: map[string]string{"numbering_strategy": "random"},
            tenantID: "t-1", branchID: "b-1",
            wantRes: func(t *testing.T, res *entity.Queue) {
                assert.Equal(t, 1, res.QueueNo)
                assert.Equal(t, "A001", res.TicketNo)
            },
        },
        {
            name:     "Vulnerability_SQLInjection",
            req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John'; DROP TABLE queues;--"},
            tenantID: "t-1", branchID: "b-1",
            wantRes: func(t *testing.T, res *entity.Queue) {
                assert.Equal(t, "John&#39;; DROP TABLE queues;--", res.PatientName)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := tt.stub
            if repo == nil {
                repo = &stubQueueRepo{}
            }
            uc := NewQueueUseCase(repo, &stubSettingsResolver{values: tt.settings}, nil)
            ctx := context.Background()
            if tt.tenantID != "" {
                ctx = database.SetOrganizationContext(ctx, tt.tenantID)
            }
            if tt.branchID != "" {
                ctx = database.SetBranchContext(ctx, tt.branchID)
            }

            res, err := uc.RegisterQueue(ctx, tt.req)

            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                return
            }
            assert.NoError(t, err)
            if tt.wantRes != nil {
                tt.wantRes(t, res)
            }
        })
    }
}
```

## 3. Feasibility Matrix

### 3.1 Unit Tests — ⭐⭐⭐⭐⭐ Sangat Cocok

| File | # Fungsi Sekarang | Target # Tabel | Kategori Tertutup |
|---|---|---|---|
| `internal/modules/queue/usecase/queue_usecase_test.go` | 30+ fungsi | 4 tabel (RegisterQueue, ForwardQueue, TransitionQueue, ListQueues) | Positive, Negative, Edge, Vulnerability |
| `internal/modules/queue/repository/queue_repository_test.go` | ~10 fungsi | 3 tabel (CreateRegistration, ListQueues, ListActiveJourneys) | Positive, Negative, Edge |
| `internal/modules/scanner/usecase/scanner_usecase_test.go` | ~8 fungsi | 2 tabel (CheckIn, Validate) | Positive, Negative, Vulnerability |
| `internal/modules/settings/usecase/settings_usecase_test.go` | ~6 fungsi | 2 tabel (CreateSetting, Resolve) | Positive, Negative, Edge |

### 3.2 Integration Tests — ⭐⭐⭐⭐ Cocok dengan Phase Split

**Pola yang bisa:**

```go
func TestQMSQueueIntegration(t *testing.T) {
    // Phase 1: Shared setup (sequential)
    env := setup.SetupIntegrationEnvironment(t)
    defer env.Cleanup()
    // create branch, service, counter, settings...

    // Phase 2: RegisterQueue — table-driven
    registerTests := []struct { ... }
    for _, tt := range registerTests {
        t.Run("Register/"+tt.name, func(t *testing.T) {
            // each case creates its own fresh queue
        })
    }

    // Phase 3: ForwardQueue — uses last queue from setup as baseline
    forwardTests := []struct { ... }
    for _, tt := range forwardTests {
        t.Run("Forward/"+tt.name, func(t *testing.T) {
            // register fresh queue, then forward
        })
    }
}
```

**Yang TIDAK bisa table-driven:**
- Lifecycle penuh (create → register → forward → transition) karena tiap step butuh state step sebelumnya.
- Solusi: lifecycle tetap 1 fungsi sequential, tapi setiap guard point (transition rules, scanner validation) bisa table-driven di dalamnya.

### 3.3 E2E Tests — ⭐⭐⭐ Sebagian Cocok

**Yang BISA table-driven:**

| Skenario | Cara |
|---|---|
| Scanner action validation | Setelah queue terdaftar, test berbagai (action, scope, expected_status) via tabel |
| Transition edge cases | Setelah queue dalam state X, test berbagai action via tabel |
| API key scope guard | Test berbagai (scope, endpoint, expected_status) via tabel |
| WebSocket event isolation | Test berbagai (user, org, expected_event) via tabel |

**Yang TIDAK bisa:**
- Full lifecycle E2E (`TestQMSQueueE2E_LifecycleAndScannerGuard`) — 90% sequential state dependency.
- Solusi: split jadi 2 fungsi:
  1. `TestQMSQueueE2E_Lifecycle` (sequential, keep as-is)
  2. `TestQMSQueueE2E_ScannerGuard` (table-driven, 10+ action combo cases)
  3. `TestQMSQueueE2E_APIKeyScopeGuard` (table-driven, 8+ scope cases)

## 4. Kolom Wajib Tabel (QMS TDD Rule)

Setiap tabel tes harus punya kolom:

```go
strukturTabel := []struct {
    name     string          // Nama test case (contoh: "Positive_Success", "Negative_NoTenant")
    category string          // Wajib: "positive" | "negative" | "edge" | "vulnerability"
    // ... input fields ...
    wantErr  error           // Error yang diharapkan (nil = sukses)
    wantRes  func(t, res)    // Assertion closure (opsional, untuk sukses)
}
```

## 5. Migration Roadmap

### Phase 1: Unit Tests (Prioritas Tertinggi, Dampak Terbesar)

| # | File | Estimasi | Dependensi |
|---|---|---|---|
| 1.1 | `queue_usecase_test.go` — RegisterQueue | 2 jam | Tidak ada |
| 1.2 | `queue_usecase_test.go` — ForwardQueue | 1 jam | Tidak ada |
| 1.3 | `queue_usecase_test.go` — TransitionQueue | 1 jam | Tidak ada |
| 1.4 | `queue_usecase_test.go` — ListQueues + ListActiveJourneys | 1 jam | Tidak ada |
| 1.5 | `scanner_usecase_test.go` | 2 jam | Tidak ada |
| 1.6 | `settings_usecase_test.go` | 1 jam | Tidak ada |

**Verifikasi:** `make test-unit` atau `go test ./internal/modules/queue/... -v -count=1`

### Phase 2: Integration Tests

| # | File | Estimasi | Dependensi |
|---|---|---|---|
| 2.1 | `qms_queue_integration_test.go` — RegisterQueue sub-tabel | 1 jam | Phase 1 selesai |
| 2.2 | `qms_queue_integration_test.go` — ForwardQueue + Transition sub-tabel | 1.5 jam | 2.1 |
| 2.3 | Stats, Audit integration (yang sudah multi-fungsi) | 1 jam | Tidak ada |

**Verifikasi:** Docker harus menyala. `make test-integration`

### Phase 3: E2E Tests

| # | File | Estimasi | Dependensi |
|---|---|---|---|
| 3.1 | `qms_queue_e2e_test.go` — Scanner guard table | 2 jam | Phase 2 selesai |
| 3.2 | `qms_queue_e2e_test.go` — API key scope table | 1.5 jam | 3.1 |
| 3.3 | `websocket_e2e_test.go` — Event isolation table | 2 jam | Tidak ada |

**Verifikasi:** Docker harus menyala. `make test-e2e`

## 6. Acceptance Criteria

- [ ] Setiap metode usecase di modul QMS punya **satu** fungsi `Test<Method>` dengan tabel `tests := []struct{...}`
- [ ] Setiap entri tabel punya kolom `category` (`positive`|`negative`|`edge`|`vulnerability`)
- [ ] Existing assertions 100% terwakili di entri tabel (tidak ada coverage loss)
- [ ] `t.Run(tt.name, ...)` digunakan untuk setiap entri (laporan verbose per kasus)
- [ ] `make test-unit` lulus tanpa perubahan hasil
- [ ] `make test-integration` (butuh Docker) lulus untuk file yang disentuh
- [ ] `make test-e2e` (butuh Docker) lulus untuk file yang disentuh

## 7. File Target

```
internal/modules/queue/usecase/queue_usecase_test.go     ← Prioritas #1
internal/modules/queue/repository/queue_repository_test.go
internal/modules/scanner/usecase/scanner_usecase_test.go
internal/modules/settings/usecase/settings_usecase_test.go
tests/integration/modules/qms_queue_integration_test.go
tests/e2e/api/qms_queue_e2e_test.go
tests/e2e/realtime/websocket_e2e_test.go                 ← Subtabel untuk event isolation & presence
```

## 8. Template File

Lihat `llm/conventions/table-driven-test-template.go` untuk template siap pakai.
