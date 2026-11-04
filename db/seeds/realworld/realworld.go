package realworld

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedRealWorldQMSData populates realistic hospital/clinic queue management data.
func SeedRealWorldQMSData(db *gorm.DB, defaultOrgID string, superadminID string) error {
	now := time.Now().UnixMilli()
	todayStr := time.Now().Format("2006-01-02")

	log.Println("Seeding Real-World QMS data (RS Sentra Medika)...")

	// 1. Update/Enrich Default Organization
	orgUpdates := map[string]interface{}{
		"code":        "RSSM",
		"name":        "RS Sentra Medika",
		"legal_name":  "PT Sentra Medika Sehat",
		"address":     "Jl. MH Thamrin No. 45, Menteng",
		"city":        "Jakarta Pusat",
		"province":    "DKI Jakarta",
		"postal_code": "10350",
		"phone":       "021-3901234",
		"email":       "kontak@sentramedika.co.id",
		"timezone":    "Asia/Jakarta",
		"status":      "active",
		"updated_at":  now,
	}
	db.Table("organizations").Where("id = ?", defaultOrgID).Updates(orgUpdates)

	// 1b. Seed Second Organization (for Multi-Tenant Demonstration)
	tenant2Slug := "sehat-bersama"
	var tenant2ID string
	_ = db.Table("organizations").Where("slug = ?", tenant2Slug).Select("id").Row().Scan(&tenant2ID)
	if tenant2ID == "" {
		tenant2ID = uuid.NewString()
		if err := db.Table("organizations").Create(map[string]interface{}{
			"id":          tenant2ID,
			"code":        "KPSB",
			"name":        "Klinik Pratama Sehat Bersama",
			"legal_name":  "PT Sehat Bersama Medika",
			"slug":        tenant2Slug,
			"owner_id":    superadminID,
			"address":     "Jl. Dago No. 112",
			"city":        "Bandung",
			"province":    "Jawa Barat",
			"postal_code": "40132",
			"phone":       "022-2505678",
			"email":       "info@sehatbersama.com",
			"timezone":    "Asia/Jakarta",
			"status":      "active",
			"created_at":  now,
			"updated_at":  now,
		}).Error; err != nil {
			log.Printf("Warning: failed to seed tenant 2: %v", err)
		} else {
			log.Println("Seeded second organization: Klinik Pratama Sehat Bersama (Bandung)")
		}
	}

	// 2. Seed Real-World Staff Users
	staffUsers := []struct {
		Username string
		FullName string
		Email    string
		RoleName string
	}{
		{"dr.budi", "dr. Budi Santoso, Sp.PD", "budi.santoso@sentramedika.co.id", "role:doctor"},
		{"dr.anita", "drg. Anita Wijaya", "anita.wijaya@sentramedika.co.id", "role:doctor"},
		{"operator.pendaftaran", "Lina Marlina", "lina.marlina@sentramedika.co.id", "role:operator"},
		{"kasir.bayu", "Bayu Pratama", "bayu.pratama@sentramedika.co.id", "role:cashier"},
		{"apoteker.siti", "apt. Siti Rahma, S.Farm", "siti.rahma@sentramedika.co.id", "role:pharmacist"},
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password0!"), bcrypt.DefaultCost)
	userIDs := make(map[string]string)

	for _, staff := range staffUsers {
		var uID string
		_ = db.Table("users").Where("username = ?", staff.Username).Select("id").Row().Scan(&uID)
		if uID == "" {
			uID = uuid.NewString()
			if err := db.Table("users").Create(map[string]interface{}{
				"id":         uID,
				"username":   staff.Username,
				"name":       staff.FullName,
				"email":      staff.Email,
				"password":   string(hashedPassword),
				"status":     "active",
				"created_at": now,
				"updated_at": now,
			}).Error; err != nil {
				log.Printf("Failed to create staff user %s: %v", staff.Username, err)
				continue
			}
		}
		userIDs[staff.Username] = uID

		// Get role ID
		var rID string
		_ = db.Table("roles").Where("name = ?", staff.RoleName).Select("id").Row().Scan(&rID)
		if rID != "" {
			var memberCount int64
			db.Table("organization_members").Where("organization_id = ? AND user_id = ?", defaultOrgID, uID).Count(&memberCount)
			if memberCount == 0 {
				db.Table("organization_members").Create(map[string]interface{}{
					"id":              uuid.NewString(),
					"organization_id": defaultOrgID,
					"user_id":         uID,
					"role_id":         rID,
					"status":          "active",
					"joined_at":       now,
				})
			}
			ensurePolicyRow(db, "g", uID, staff.RoleName, defaultOrgID, "", "")
			ensurePolicyRow(db, "g", uID, staff.RoleName, "global", "", "")
		}
	}

	// 3. Seed Branches
	branches := []struct {
		Code        string
		Name        string
		Address     string
		City        string
		Province    string
		PostalCode  string
		Phone       string
		Email       string
		RunningText string
	}{
		{
			Code:        "JKT-PST",
			Name:        "Cabang Jakarta Pusat - Thamrin",
			Address:     "Jl. MH Thamrin No. 45",
			City:        "Jakarta Pusat",
			Province:    "DKI Jakarta",
			PostalCode:  "10350",
			Phone:       "021-3901234",
			Email:       "thamrin@sentramedika.co.id",
			RunningText: "Selamat datang di RS Sentra Medika Thamrin. Budayakan tertib mengantre untuk kenyamanan bersama. Pasien prioritas lansia & disabilitas silakan menuju Loket 1.",
		},
		{
			Code:        "JKT-SEL",
			Name:        "Cabang Jakarta Selatan - Simatupang",
			Address:     "Jl. TB Simatupang No. 88",
			City:        "Jakarta Selatan",
			Province:    "DKI Jakarta",
			PostalCode:  "12520",
			Phone:       "021-7801234",
			Email:       "simatupang@sentramedika.co.id",
			RunningText: "Klinik Sentra Medika Simatupang melayani Poli Umum, Poli Gigi, dan Farmasi setiap hari pukul 08:00 - 20:00 WIB.",
		},
	}

	branchIDs := make(map[string]string)
	for _, b := range branches {
		var bID string
		_ = db.Table("branches").Where("tenant_id = ? AND code = ?", defaultOrgID, b.Code).Select("id").Row().Scan(&bID)
		if bID == "" {
			bID = uuid.NewString()
			if err := db.Table("branches").Create(map[string]interface{}{
				"id":           bID,
				"tenant_id":    defaultOrgID,
				"code":         b.Code,
				"name":         b.Name,
				"address":      b.Address,
				"city":         b.City,
				"province":     b.Province,
				"postal_code":  b.PostalCode,
				"phone":        b.Phone,
				"email":        b.Email,
				"running_text": b.RunningText,
				"timezone":     "Asia/Jakarta",
				"status":       "active",
				"created_at":   now,
				"updated_at":   now,
			}).Error; err != nil {
				log.Printf("Failed to create branch %s: %v", b.Code, err)
				continue
			}
		} else {
			db.Table("branches").Where("id = ?", bID).Updates(map[string]interface{}{
				"running_text": b.RunningText,
				"status":       "active",
				"updated_at":   now,
			})
		}
		branchIDs[b.Code] = bID
	}

	mainBranchID := branchIDs["JKT-PST"]
	if mainBranchID == "" {
		return fmt.Errorf("main branch JKT-PST was not created")
	}

	// 4. Seed Services
	servicesData := []struct {
		Code                     string
		Name                     string
		Type                     string
		IsPharmacy               bool
		IsPharmacyReception      bool
		DefaultEstimatedDuration int
		NarrativeInstruction     string
		AudioID                  string
	}{
		{
			Code:                     "REG",
			Name:                     "Pendaftaran & Informasi",
			Type:                     "general",
			IsPharmacy:               false,
			IsPharmacyReception:      true,
			DefaultEstimatedDuration: 3,
			NarrativeInstruction:     "Silakan menuju loket pendaftaran dan siapkan kartu identitas atau kartu BPJS Anda.",
			AudioID:                  "/uploads/services_audio_bell.mp3",
		},
		{
			Code:                     "POLI-UMUM",
			Name:                     "Poli Umum & Penyakit Dalam",
			Type:                     "general",
			IsPharmacy:               false,
			IsPharmacyReception:      false,
			DefaultEstimatedDuration: 10,
			NarrativeInstruction:     "Silakan masuk ke ruang periksa poli umum.",
			AudioID:                  "/uploads/services_audio_bell.mp3",
		},
		{
			Code:                     "POLI-GIGI",
			Name:                     "Poli Gigi & Mulut",
			Type:                     "general",
			IsPharmacy:               false,
			IsPharmacyReception:      false,
			DefaultEstimatedDuration: 15,
			NarrativeInstruction:     "Silakan masuk ke ruang dokter gigi.",
			AudioID:                  "/uploads/services_audio_bell.mp3",
		},
		{
			Code:                     "KASIR",
			Name:                     "Kasir & Pembayaran",
			Type:                     "general",
			IsPharmacy:               false,
			IsPharmacyReception:      false,
			DefaultEstimatedDuration: 4,
			NarrativeInstruction:     "Silakan menuju loket kasir untuk menyelesaikan administrasi dan pembayaran.",
			AudioID:                  "/uploads/services_audio_bell.mp3",
		},
		{
			Code:                     "FARMASI",
			Name:                     "Farmasi & Penyerahan Obat",
			Type:                     "general",
			IsPharmacy:               true,
			IsPharmacyReception:      false,
			DefaultEstimatedDuration: 8,
			NarrativeInstruction:     "Nomor antrean Anda telah siap. Silakan menuju loket farmasi untuk penyerahan obat.",
			AudioID:                  "/uploads/services_audio_bell.mp3",
		},
	}

	serviceIDs := make(map[string]string)
	for _, s := range servicesData {
		var sID string
		_ = db.Table("services").Where("tenant_id = ? AND code = ?", defaultOrgID, s.Code).Select("id").Row().Scan(&sID)
		if sID == "" {
			sID = uuid.NewString()
			if err := db.Table("services").Create(map[string]interface{}{
				"id":                         sID,
				"tenant_id":                  defaultOrgID,
				"code":                       s.Code,
				"name":                       s.Name,
				"type":                       s.Type,
				"is_pharmacy":                s.IsPharmacy,
				"is_pharmacy_reception":      s.IsPharmacyReception,
				"default_estimated_duration": s.DefaultEstimatedDuration,
				"narrative_instruction_id":   s.NarrativeInstruction,
				"audio_id":                   s.AudioID,
				"status":                     "active",
				"created_at":                 now,
				"updated_at":                 now,
			}).Error; err != nil {
				log.Printf("Failed to create service %s: %v", s.Code, err)
				continue
			}
		} else {
			db.Table("services").Where("id = ?", sID).Updates(map[string]interface{}{
				"is_pharmacy":                s.IsPharmacy,
				"is_pharmacy_reception":      s.IsPharmacyReception,
				"default_estimated_duration": s.DefaultEstimatedDuration,
				"narrative_instruction_id":   s.NarrativeInstruction,
				"audio_id":                   s.AudioID,
				"status":                     "active",
				"updated_at":                 now,
			})
		}
		serviceIDs[s.Code] = sID
	}

	// 5. Seed Branch Services
	branchServiceIDs := make(map[string]string)
	sortOrder := 1
	for _, s := range servicesData {
		sID := serviceIDs[s.Code]
		if sID == "" {
			continue
		}
		var bsID string
		_ = db.Table("branch_services").Where("tenant_id = ? AND branch_id = ? AND service_id = ?", defaultOrgID, mainBranchID, sID).Select("id").Row().Scan(&bsID)
		if bsID == "" {
			bsID = uuid.NewString()
			if err := db.Table("branch_services").Create(map[string]interface{}{
				"id":         bsID,
				"tenant_id":  defaultOrgID,
				"branch_id":  mainBranchID,
				"service_id": sID,
				"is_active":  true,
				"sort_order": sortOrder,
				"created_at": now,
				"updated_at": now,
			}).Error; err != nil {
				log.Printf("Failed to link branch service %s: %v", s.Code, err)
				continue
			}
		}
		branchServiceIDs[s.Code] = bsID
		sortOrder++
	}

	// 6. Seed Counters
	countersData := []struct {
		Code             string
		Name             string
		DisplayName      string
		BranchServiceKey string
	}{
		{"LOKET-01", "Loket 1", "Loket 1 - BPJS & Prioritas", "REG"},
		{"LOKET-02", "Loket 2", "Loket 2 - Pendaftaran Umum", "REG"},
		{"DOKTER-01", "Ruang Periksa 1", "Ruang Periksa 1 - dr. Budi (Poli Umum)", "POLI-UMUM"},
		{"DOKTER-02", "Ruang Periksa 2", "Ruang Periksa 2 - drg. Anita (Poli Gigi)", "POLI-GIGI"},
		{"KASIR-01", "Loket 3", "Loket 3 - Kasir Pembayaran", "KASIR"},
		{"FARMASI-01", "Loket 4", "Loket 4 - Penyerahan Obat Farmasi", "FARMASI"},
	}

	counterIDs := make(map[string]string)
	for _, c := range countersData {
		bsID := branchServiceIDs[c.BranchServiceKey]
		var cID string
		_ = db.Table("counters").Where("tenant_id = ? AND branch_id = ? AND code = ?", defaultOrgID, mainBranchID, c.Code).Select("id").Row().Scan(&cID)
		if cID == "" {
			cID = uuid.NewString()
			insertData := map[string]interface{}{
				"id":           cID,
				"tenant_id":    defaultOrgID,
				"branch_id":    mainBranchID,
				"code":         c.Code,
				"name":         c.Name,
				"display_name": c.DisplayName,
				"status":       "active",
				"created_at":   now,
				"updated_at":   now,
			}
			if bsID != "" {
				insertData["branch_service_id"] = bsID
			}
			if err := db.Table("counters").Create(insertData).Error; err != nil {
				log.Printf("Failed to create counter %s: %v", c.Code, err)
				continue
			}
		} else {
			db.Table("counters").Where("id = ?", cID).Updates(map[string]interface{}{
				"display_name":      c.DisplayName,
				"branch_service_id": bsID,
				"status":            "active",
				"updated_at":        now,
			})
		}
		counterIDs[c.Code] = cID
	}

	// 7. Seed Typed Queue Settings Hierarchy
	// 7a. Tenant Queue Settings
	var tenantSettingCount int64
	db.Table("tenant_queue_settings").Where("tenant_id = ?", defaultOrgID).Count(&tenantSettingCount)
	if tenantSettingCount == 0 {
		db.Table("tenant_queue_settings").Create(map[string]interface{}{
			"id":                         uuid.NewString(),
			"tenant_id":                  defaultOrgID,
			"queue_reset_time":           "00:00",
			"default_ticket_prefix":      "A",
			"default_estimated_duration": 5,
			"allow_forward":              true,
			"allow_skip":                 true,
			"allow_recall":               true,
			"allow_cancel":               true,
			"numbering_strategy":         "daily_branch_sequence",
			"created_at":                 now,
			"updated_at":                 now,
		})
	}

	// 7b. Branch Queue Settings
	var branchSettingCount int64
	db.Table("branch_queue_settings").Where("tenant_id = ? AND branch_id = ?", defaultOrgID, mainBranchID).Count(&branchSettingCount)
	if branchSettingCount == 0 {
		autoCallNext := false
		allowForward := true
		allowSkip := true
		allowRecall := true
		allowCancel := true
		prefixA := "A"
		resetTime := "00:00"
		duration5 := 5
		strategy := "daily_branch_sequence"

		db.Table("branch_queue_settings").Create(map[string]interface{}{
			"id":                         uuid.NewString(),
			"tenant_id":                  defaultOrgID,
			"branch_id":                  mainBranchID,
			"queue_reset_time":           resetTime,
			"ticket_prefix":              prefixA,
			"default_estimated_duration": duration5,
			"allow_forward":              allowForward,
			"allow_skip":                 allowSkip,
			"allow_recall":               allowRecall,
			"allow_cancel":               allowCancel,
			"auto_call_next":             autoCallNext,
			"numbering_strategy":         strategy,
			"created_at":                 now,
			"updated_at":                 now,
		})
	}

	// 7c. Branch Service Queue Settings (Prefix and Duration Overrides)
	serviceSettings := []struct {
		ServiceKey string
		Prefix     string
		Duration   int
	}{
		{"REG", "A", 3},
		{"POLI-UMUM", "B", 10},
		{"POLI-GIGI", "C", 15},
		{"KASIR", "D", 4},
		{"FARMASI", "F", 8},
	}

	for _, ss := range serviceSettings {
		bsID := branchServiceIDs[ss.ServiceKey]
		if bsID == "" {
			continue
		}
		var count int64
		db.Table("branch_service_queue_settings").Where("tenant_id = ? AND branch_id = ? AND branch_service_id = ?", defaultOrgID, mainBranchID, bsID).Count(&count)
		if count == 0 {
			requireCounter := true
			allowForwardFrom := true
			allowForwardTo := true
			allowSkip := true
			allowRecall := true
			allowCancel := true
			autoCallNext := false

			db.Table("branch_service_queue_settings").Create(map[string]interface{}{
				"id":                         uuid.NewString(),
				"tenant_id":                  defaultOrgID,
				"branch_id":                  mainBranchID,
				"branch_service_id":          bsID,
				"default_estimated_duration": ss.Duration,
				"require_counter":            requireCounter,
				"allow_forward_from":         allowForwardFrom,
				"allow_forward_to":           allowForwardTo,
				"allow_skip":                 allowSkip,
				"allow_recall":               allowRecall,
				"allow_cancel":               allowCancel,
				"auto_call_next":             autoCallNext,
				"created_at":                 now,
				"updated_at":                 now,
			})
		}
	}

	// 8. Seed Operator Counter Assignments
	operatorMap := []struct {
		Username   string
		CounterKey string
	}{
		{"operator.pendaftaran", "LOKET-01"},
		{"dr.budi", "DOKTER-01"},
		{"dr.anita", "DOKTER-02"},
		{"kasir.bayu", "KASIR-01"},
		{"apoteker.siti", "FARMASI-01"},
	}

	for _, op := range operatorMap {
		uID := userIDs[op.Username]
		cID := counterIDs[op.CounterKey]
		if uID == "" || cID == "" {
			continue
		}
		var count int64
		db.Table("operator_counter_assignments").Where("tenant_id = ? AND branch_id = ? AND user_id = ? AND counter_id = ? AND unassigned_at IS NULL", defaultOrgID, mainBranchID, uID, cID).Count(&count)
		if count == 0 {
			db.Table("operator_counter_assignments").Create(map[string]interface{}{
				"id":          uuid.NewString(),
				"tenant_id":   defaultOrgID,
				"branch_id":   mainBranchID,
				"user_id":     uID,
				"counter_id":  cID,
				"assigned_at": now - 3600000, // 1 hour ago
			})
		}
	}

	// 9. Seed QMS Hardware Clients & Credentials
	qmsClients := []struct {
		Name        string
		ClientType  string
		CounterKey  string
		Secret      string
		Description string
	}{
		{
			Name:        "TV Signage Ruang Tunggu Utama Lobi",
			ClientType:  "signage",
			Secret:      "SignageLobiSecret123!",
			Description: "Layar TV Samsung 55 inch di atas ruang tunggu lobi pendaftaran",
		},
		{
			Name:        "TV Signage Depan Farmasi",
			ClientType:  "signage",
			Secret:      "SignageFarmasiSecret123!",
			Description: "Layar TV display khusus status panggilan obat apotek",
		},
		{
			Name:        "Kios Cetak Tiket Mandiri Lobi",
			ClientType:  "scanner",
			Secret:      "KioskSecret123!",
			Description: "Mesin Kiosk self-service touchscreen dengan thermal printer",
		},
		{
			Name:        "Caller Tablet Loket 1",
			ClientType:  "caller",
			CounterKey:  "LOKET-01",
			Secret:      "CallerSecret123!",
			Description: "Tablet Android pemanggil antrean di Loket 1 Pendaftaran",
		},
		{
			Name:        "Caller PC Dokter Ruang 1",
			ClientType:  "caller",
			CounterKey:  "DOKTER-01",
			Secret:      "CallerDokter123!",
			Description: "Aplikasi caller web di PC meja dokter Budi",
		},
	}

	for _, qc := range qmsClients {
		var clientID string
		_ = db.Table("qms_clients").Where("tenant_id = ? AND branch_id = ? AND name = ?", defaultOrgID, mainBranchID, qc.Name).Select("id").Row().Scan(&clientID)
		if clientID == "" {
			clientID = uuid.NewString()
			clientData := map[string]interface{}{
				"id":          clientID,
				"tenant_id":   defaultOrgID,
				"branch_id":   mainBranchID,
				"client_type": qc.ClientType,
				"name":        qc.Name,
				"description": qc.Description,
				"is_active":   true,
				"created_at":  now,
				"updated_at":  now,
			}
			if qc.CounterKey != "" && counterIDs[qc.CounterKey] != "" {
				clientData["counter_id"] = counterIDs[qc.CounterKey]
			}
			if err := db.Table("qms_clients").Create(clientData).Error; err != nil {
				log.Printf("Failed to create QMS client %s: %v", qc.Name, err)
				continue
			}

			// Seed credential
			credHash, _ := bcrypt.GenerateFromPassword([]byte(qc.Secret), bcrypt.DefaultCost)
			db.Table("qms_client_credentials").Create(map[string]interface{}{
				"id":                 uuid.NewString(),
				"tenant_id":          defaultOrgID,
				"client_id":          clientID,
				"client_secret_hash": string(credHash),
				"created_at":         now,
				"updated_at":         now,
			})
		}
	}

	// 10. Seed Realistic Active & Historical Queues for Today
	// 10a. Queue Counters for today
	queuePrefixCounters := map[string]int{
		"A": 5, // Pendaftaran (5 tickets generated today)
		"B": 2, // Poli Umum
		"C": 1, // Poli Gigi
		"D": 2, // Kasir
		"F": 2, // Farmasi
	}
	for prefix, lastVal := range queuePrefixCounters {
		var cnt int64
		db.Table("queue_counters").Where("tenant_id = ? AND branch_id = ? AND queue_date = ? AND prefix = ?", defaultOrgID, mainBranchID, todayStr, prefix).Count(&cnt)
		if cnt == 0 {
			db.Table("queue_counters").Create(map[string]interface{}{
				"tenant_id":  defaultOrgID,
				"branch_id":  mainBranchID,
				"queue_date": todayStr,
				"prefix":     prefix,
				"last_value": lastVal,
				"created_at": now,
				"updated_at": now,
			})
		}
	}

	// 10b. Queues Master Data
	type JourneySpec struct {
		ServiceKey string
		CounterKey string
		Status     string
	}

	queueScenarios := []struct {
		TicketNo    string
		QueueNo     int
		PatientName string
		Status      string // waiting, calling, serving, skipped, canceled, completed
		Journeys    []JourneySpec
	}{
		{
			TicketNo:    "A001",
			QueueNo:     1,
			PatientName: "Bambang Sutrisno",
			Status:      "completed",
			Journeys: []JourneySpec{
				{"REG", "LOKET-01", "completed"},
				{"POLI-UMUM", "DOKTER-01", "completed"},
				{"KASIR", "KASIR-01", "completed"},
				{"FARMASI", "FARMASI-01", "completed"},
			},
		},
		{
			TicketNo:    "A002",
			QueueNo:     2,
			PatientName: "Hendra Wijaya",
			Status:      "skipped",
			Journeys: []JourneySpec{
				{"REG", "LOKET-02", "skipped"},
			},
		},
		{
			TicketNo:    "A003",
			QueueNo:     3,
			PatientName: "Reza Rahadian",
			Status:      "calling",
			Journeys: []JourneySpec{
				{"REG", "LOKET-01", "calling"},
			},
		},
		{
			TicketNo:    "B002",
			QueueNo:     4,
			PatientName: "Siti Aminah",
			Status:      "serving",
			Journeys: []JourneySpec{
				{"REG", "LOKET-01", "completed"},
				{"POLI-UMUM", "DOKTER-01", "serving"},
			},
		},
		{
			TicketNo:    "F001",
			QueueNo:     5,
			PatientName: "Dewi Lestari",
			Status:      "waiting",
			Journeys: []JourneySpec{
				{"REG", "LOKET-01", "completed"},
				{"POLI-UMUM", "DOKTER-01", "completed"},
				{"KASIR", "KASIR-01", "completed"},
				{"FARMASI", "", "pending"},
			},
		},
		{
			TicketNo:    "A004",
			QueueNo:     6,
			PatientName: "Nurul Hidayah",
			Status:      "waiting",
			Journeys: []JourneySpec{
				{"REG", "", "pending"},
			},
		},
		{
			TicketNo:    "A005",
			QueueNo:     7,
			PatientName: "Agus Prasetyo",
			Status:      "waiting",
			Journeys: []JourneySpec{
				{"REG", "", "pending"},
			},
		},
	}

	for _, qs := range queueScenarios {
		var qID string
		_ = db.Table("queues").Where("tenant_id = ? AND branch_id = ? AND queue_date = ? AND ticket_no = ?", defaultOrgID, mainBranchID, todayStr, qs.TicketNo).Select("id").Row().Scan(&qID)
		if qID == "" {
			qID = uuid.NewString()
			patientID := uuid.NewString()
			queueCreatedAt := now - int64(100-qs.QueueNo)*60000

			if err := db.Table("queues").Create(map[string]interface{}{
				"id":           qID,
				"tenant_id":    defaultOrgID,
				"branch_id":    mainBranchID,
				"queue_date":   todayStr,
				"ticket_no":    qs.TicketNo,
				"queue_no":     qs.QueueNo,
				"patient_id":   patientID,
				"patient_name": qs.PatientName,
				"status":       qs.Status,
				"created_at":   queueCreatedAt,
				"updated_at":   now,
			}).Error; err != nil {
				log.Printf("Failed to seed queue %s: %v", qs.TicketNo, err)
				continue
			}

			// Seed journeys
			var lastJourneyID string
			for seqIdx, j := range qs.Journeys {
				sID := serviceIDs[j.ServiceKey]
				cID := counterIDs[j.CounterKey]
				jID := uuid.NewString()
				lastJourneyID = jID

				journeyData := map[string]interface{}{
					"id":         jID,
					"queue_id":   qID,
					"tenant_id":  defaultOrgID,
					"branch_id":  mainBranchID,
					"service_id": sID,
					"seq_no":     seqIdx + 1,
					"status":     j.Status,
					"created_at": queueCreatedAt + int64(seqIdx)*180000,
					"updated_at": queueCreatedAt + int64(seqIdx+1)*180000,
				}
				if cID != "" {
					journeyData["counter_id"] = cID
				}
				db.Table("queue_journeys").Create(journeyData)

				// Seed visit journeys event audit
				payloadJSON, _ := json.Marshal(map[string]interface{}{
					"ticket_no":   qs.TicketNo,
					"service_id":  sID,
					"service_key": j.ServiceKey,
					"status":      j.Status,
				})
				eventType := "registration"
				switch j.Status {
				case "completed":
					eventType = "complete"
				case "serving":
					eventType = "serve"
				case "calling":
					eventType = "call"
				case "skipped":
					eventType = "skip"
				}

				db.Table("visit_journeys").Create(map[string]interface{}{
					"id":         uuid.NewString(),
					"queue_id":   qID,
					"tenant_id":  defaultOrgID,
					"branch_id":  mainBranchID,
					"event_type": eventType,
					"payload":    string(payloadJSON),
					"created_at": queueCreatedAt + int64(seqIdx)*180000,
				})
			}

			if lastJourneyID != "" {
				db.Table("queues").Where("id = ?", qID).Update("current_journey_id", lastJourneyID)
			}
		}
	}

	// 11. Seed API Keys
	rawKey := "sk_live_sentramedika_his_2026_demo_key"
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	var apiKeyCount int64
	db.Table("api_keys").Where("key_hash = ?", keyHash).Count(&apiKeyCount)
	if apiKeyCount == 0 {
		var uID string
		_ = db.Table("users").Where("username = ?", "adminuser").Select("id").Row().Scan(&uID)
		if uID == "" {
			uID = superadminID
		}
		db.Table("api_keys").Create(map[string]interface{}{
			"id":              uuid.NewString(),
			"name":            "SIMRS Hospital Information System Connector",
			"key_hash":        keyHash,
			"organization_id": defaultOrgID,
			"user_id":         uID,
			"scopes":          "service:view,service:manage,queue:view,queue:manage,branch:view,counter:view",
			"is_active":       true,
		})
		log.Println("Seeded API Key: SIMRS Hospital Connector")
	}

	// 12. Seed Webhooks & Delivery Log
	var webhookID string
	_ = db.Table("webhooks").Where("organization_id = ? AND name = ?", defaultOrgID, "WhatsApp Patient Notification Gateway").Select("id").Row().Scan(&webhookID)
	if webhookID == "" {
		webhookID = uuid.NewString()
		eventsJSON, _ := json.Marshal([]string{"queue.registered", "queue.called", "queue.completed"})
		if err := db.Table("webhooks").Create(map[string]interface{}{
			"id":              webhookID,
			"name":            "WhatsApp Patient Notification Gateway",
			"organization_id": defaultOrgID,
			"url":             "https://webhook.site/mock-wa-gateway/antrean",
			"events":          string(eventsJSON),
			"secret":          "whsec_sentra_medika_demo_secret_9988",
			"is_active":       true,
			"created_at":      now,
			"updated_at":      now,
		}).Error; err == nil {
			// Seed sample log
			samplePayload, _ := json.Marshal(map[string]interface{}{
				"event":     "queue.called",
				"ticket_no": "A003",
				"patient":   "Reza Rahadian",
				"counter":   "Loket 1 - BPJS & Prioritas",
				"called_at": now,
			})
			db.Table("webhook_logs").Create(map[string]interface{}{
				"id":                   uuid.NewString(),
				"webhook_id":           webhookID,
				"event_type":           "queue.called",
				"payload":              string(samplePayload),
				"response_status_code": 200,
				"response_body":        `{"status":"success","message":"WhatsApp notification queued"}`,
				"execution_time":       132,
				"retry_count":          0,
				"created_at":           now - 60000,
			})
			log.Println("Seeded Webhook & Webhook Logs: WhatsApp Gateway")
		}
	}

	log.Println("Real-World QMS data seeded successfully!")
	return nil
}

func ensurePolicyRow(db *gorm.DB, ptype, v0, v1, v2, v3, v4 string) {
	var count int64
	query := db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", ptype, v0, v1, v2)
	if v3 != "" {
		query = query.Where("v3 = ?", v3)
	}
	if v4 != "" {
		query = query.Where("v4 = ?", v4)
	}
	query.Count(&count)

	if count == 0 {
		db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": ptype, "v0": v0, "v1": v1, "v2": v2, "v3": v3, "v4": v4,
		})
	}
}
