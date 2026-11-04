package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Roisfaozi/queue-base/db/seeds/realworld"
	"github.com/Roisfaozi/queue-base/internal/config"
	roleEntity "github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.User,
		cfg.Mysql.Password,
		cfg.Mysql.Host,
		cfg.Mysql.Port,
		cfg.Mysql.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected. Starting Tiered Authorization Seeder...")

	// 1. Seed Roles
	seedRoles(db)

	// 2. Seed Superadmin User
	superadminID := seedSuperAdmin(db)

	// 3. Authenticate as Superadmin to get the token
	token, err := authenticateSuperadmin(cfg)
	if err != nil {
		log.Printf("Notice: Could not authenticate via API (%v). Seeding directly to database...", err)
		seedAccessRightsAndPoliciesDirectly(db, getAccessMap(), getRoleToRights())
		seedOrganizationsAndUsersDirectly(db)
		var defaultOrgID string
		_ = db.Table("organizations").Where("slug = ?", "default-org").Select("id").Row().Scan(&defaultOrgID)
		if defaultOrgID != "" {
			_ = realworld.SeedRealWorldQMSData(db, defaultOrgID, superadminID)
		}
		log.Println("Direct database seeding completed successfully.")
		return
	}

	// 4. Seed Access Rights, Endpoints, and Tiered Policies via API
	seedAccessRightsAndPoliciesViaAPI(cfg, token)

	// 5. Seed Default Organization and Users via API
	seedOrganizationsAndUsers(cfg, token)

	// 6. Seed Real-World QMS Data
	var defaultOrgID string
	_ = db.Table("organizations").Where("slug = ?", "default-org").Select("id").Row().Scan(&defaultOrgID)
	if defaultOrgID != "" {
		_ = realworld.SeedRealWorldQMSData(db, defaultOrgID, superadminID)
	}

	log.Println("Seeding process completed successfully.")
}

func seedRoles(db *gorm.DB) {
	roles := []roleEntity.Role{
		{Name: "role:superadmin", Description: "Full Access", OrganizationID: nil},
		{Name: "role:admin", Description: "Org Administrator", OrganizationID: nil},
		{Name: "role:doctor", Description: "Dokter Pemeriksa", OrganizationID: nil},
		{Name: "role:operator", Description: "Petugas Loket / Pendaftaran", OrganizationID: nil},
		{Name: "role:cashier", Description: "Petugas Kasir & Pembayaran", OrganizationID: nil},
		{Name: "role:pharmacist", Description: "Apoteker & Penyerahan Obat", OrganizationID: nil},
		{Name: "role:user", Description: "Org User", OrganizationID: nil},
	}

	for _, r := range roles {
		var count int64
		db.Model(&roleEntity.Role{}).Where("name = ?", r.Name).Count(&count)
		if count == 0 {
			r.ID = uuid.NewString()
			r.CreatedAt = time.Now().UnixMilli()
			r.UpdatedAt = time.Now().UnixMilli()
			db.Create(&r)
			log.Printf("Role '%s' created.", r.Name)
		} else {
			// Update existing role description just in case
			db.Model(&roleEntity.Role{}).Where("name = ?", r.Name).Update("description", r.Description)
		}
	}
}

func seedSuperAdmin(db *gorm.DB) string {
	adminUsername := "superadmin"
	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("SUPERADMIN_PASSWORD environment variable is missing in .env")
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	var user userEntity.User
	if err := db.Where("username = ?", adminUsername).First(&user).Error; err != nil {
		now := time.Now().UnixMilli()
		userID := uuid.NewString()

		// Use Map to avoid "Unknown column 'avatar_url'" errors if DB schema is not fully up to date
		userData := map[string]interface{}{
			"id":         userID,
			"username":   adminUsername,
			"email":      "superadmin@example.com",
			"password":   string(hashedPwd),
			"name":       "Super Admin",
			"created_at": now,
			"updated_at": now,
		}

		if err := db.Table("users").Create(userData).Error; err != nil {
			log.Fatalf("Failed to create superadmin: %v", err)
		}
		user.ID = userID
		log.Printf("Superadmin user '%s' created.", adminUsername)
	} else {
		// ALWAYS reset superadmin password to ensure login works with current .env
		db.Table("users").Where("id = ?", user.ID).Update("password", string(hashedPwd))
		log.Printf("Superadmin user '%s' password reset.", adminUsername)
	}

	// Policy: superadmin USER has superadmin ROLE
	ensurePolicy(db, "g", user.ID, "role:superadmin", "global", "", "")
	// Policy: superadmin ROLE has all permission
	ensurePolicy(db, "p", "role:superadmin", "global", "*", "*", "")

	return user.ID
}

func authenticateSuperadmin(cfg *config.AppConfig) (string, error) {
	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		return "", fmt.Errorf("SUPERADMIN_PASSWORD environment variable is missing in .env")
	}

	apiBaseURL := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)
	loginURL := fmt.Sprintf("%s/auth/login", apiBaseURL)

	payload := map[string]string{
		"username": "superadmin",
		"password": adminPassword,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("login failed with status %d: %v", resp.StatusCode, errResp)
	}

	var loginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode login response: %v", err)
	}

	return loginResp.Data.AccessToken, nil
}

func doJSONRequest(method, url string, payload interface{}, token string, expectedStatus int) (map[string]interface{}, error) {
	var bodyReader *bytes.Buffer
	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(payloadBytes)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != expectedStatus {
		return result, fmt.Errorf("API request %s %s failed with status %d: %v", method, url, resp.StatusCode, result)
	}

	return result, nil
}

func getAccessMap() map[string][]map[string]string {
	return map[string][]map[string]string{
		"dashboard:view": {
			{"Path": "/api/v1/stats/summary", "Method": "GET"},
			{"Path": "/api/v1/stats/activity", "Method": "GET"},
			{"Path": "/api/v1/stats/insights", "Method": "GET"},
		},
		"user:view": {
			{"Path": "/api/v1/users", "Method": "GET"},
			{"Path": "/api/v1/users/", "Method": "GET"},
			{"Path": "/api/v1/users/me", "Method": "GET"},
			{"Path": "/api/v1/users/search", "Method": "POST"},
			{"Path": "/api/v1/users/:id", "Method": "GET"},
		},
		"user:manage": {
			{"Path": "/api/v1/users/me", "Method": "PUT"},
			{"Path": "/api/v1/users/me/avatar", "Method": "PATCH"},
			{"Path": "/api/v1/users/:id/status", "Method": "PATCH"},
			{"Path": "/api/v1/users/:id", "Method": "DELETE"},
		},
		"org:view": {
			{"Path": "/api/v1/tenant/profile", "Method": "GET"},
			{"Path": "/api/v1/organizations/:id", "Method": "GET"},
			{"Path": "/api/v1/organizations/me", "Method": "GET"},
			{"Path": "/api/v1/organizations/slug/:slug", "Method": "GET"},
		},
		"org:manage": {
			{"Path": "/api/v1/tenant/profile", "Method": "PATCH"},
			{"Path": "/api/v1/organizations", "Method": "POST"},
			{"Path": "/api/v1/organizations/:id", "Method": "PUT"},
			{"Path": "/api/v1/organizations/:id", "Method": "DELETE"},
			{"Path": "/api/v1/organizations/:id/restore", "Method": "POST"},
			{"Path": "/api/v1/organizations/:id/hard", "Method": "DELETE"},
		},
		"member:manage": {
			{"Path": "/api/v1/organizations/:id/members/invite", "Method": "POST"},
			{"Path": "/api/v1/organizations/:id/members", "Method": "GET"},
			{"Path": "/api/v1/organizations/:id/members/:userId", "Method": "PATCH"},
			{"Path": "/api/v1/organizations/:id/members/:userId", "Method": "DELETE"},
		},
		"presence:view": {
			{"Path": "/api/v1/organizations/:id/presence", "Method": "GET"},
		},
		"role:view": {
			{"Path": "/api/v1/roles", "Method": "GET"},
			{"Path": "/api/v1/roles/search", "Method": "POST"},
		},
		"role:manage": {
			{"Path": "/api/v1/roles", "Method": "POST"},
			{"Path": "/api/v1/roles/:id", "Method": "PUT"},
			{"Path": "/api/v1/roles/:id", "Method": "DELETE"},
		},
		"permission:view": {
			{"Path": "/api/v1/permissions", "Method": "GET"},
			{"Path": "/api/v1/permissions/:role", "Method": "GET"},
			{"Path": "/api/v1/permissions/roles/:role/users", "Method": "GET"},
			{"Path": "/api/v1/permissions/:role/parents", "Method": "GET"},
			{"Path": "/api/v1/permissions/resources", "Method": "GET"},
			{"Path": "/api/v1/permissions/inheritance-tree", "Method": "GET"},
		},
		"permission:manage": {
			{"Path": "/api/v1/permissions/assign-role", "Method": "POST"},
			{"Path": "/api/v1/permissions/revoke-role", "Method": "DELETE"},
			{"Path": "/api/v1/permissions/grant", "Method": "POST"},
			{"Path": "/api/v1/permissions", "Method": "PUT"},
			{"Path": "/api/v1/permissions/revoke", "Method": "DELETE"},
			{"Path": "/api/v1/permissions/inheritance", "Method": "POST"},
			{"Path": "/api/v1/permissions/inheritance", "Method": "DELETE"},
			// Bulk Access Right assignment (new)
			{"Path": "/api/v1/permissions/assign-access-right", "Method": "POST"},
			{"Path": "/api/v1/permissions/revoke-access-right", "Method": "DELETE"},
			{"Path": "/api/v1/permissions/roles/:role/access-rights", "Method": "GET"},
		},
		"access:view": {
			{"Path": "/api/v1/access-rights", "Method": "GET"},
			{"Path": "/api/v1/access-rights/search", "Method": "POST"},
			{"Path": "/api/v1/endpoints/search", "Method": "POST"},
		},
		"access:manage": {
			{"Path": "/api/v1/access-rights", "Method": "POST"},
			{"Path": "/api/v1/access-rights/:id", "Method": "DELETE"},
			{"Path": "/api/v1/access-rights/link", "Method": "POST"},
			{"Path": "/api/v1/access-rights/unlink", "Method": "POST"},
			{"Path": "/api/v1/endpoints", "Method": "POST"},
			{"Path": "/api/v1/endpoints/:id", "Method": "DELETE"},
		},
		"audit:view": {
			{"Path": "/api/v1/audit-logs/search", "Method": "POST"},
			{"Path": "/api/v1/audit-logs/export", "Method": "GET"},
			{"Path": "/api/v1/audit-logs/export-async", "Method": "POST"},
		},
		"webhook:manage": {
			{"Path": "/api/v1/webhooks", "Method": "POST"},
			{"Path": "/api/v1/webhooks", "Method": "GET"},
			{"Path": "/api/v1/webhooks/:id", "Method": "GET"},
			{"Path": "/api/v1/webhooks/:id", "Method": "PUT"},
			{"Path": "/api/v1/webhooks/:id", "Method": "DELETE"},
			{"Path": "/api/v1/webhooks/:id/logs", "Method": "GET"},
		},
		"branch:view": {
			{"Path": "/api/v1/branches", "Method": "GET"},
			{"Path": "/api/v1/branches/:id", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/profile", "Method": "GET"},
		},
		"branch:manage": {
			{"Path": "/api/v1/branches", "Method": "POST"},
			{"Path": "/api/v1/branches/:id", "Method": "PUT"},
			{"Path": "/api/v1/branches/:id/profile", "Method": "PATCH"},
			{"Path": "/api/v1/branches/:id", "Method": "DELETE"},
		},
		"service:view": {
			{"Path": "/api/v1/services", "Method": "GET"},
			{"Path": "/api/v1/services/:id", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/services", "Method": "GET"},
		},
		"service:manage": {
			{"Path": "/api/v1/services", "Method": "POST"},
			{"Path": "/api/v1/services/:id", "Method": "PUT"},
			{"Path": "/api/v1/services/audio", "Method": "POST"},
			{"Path": "/api/v1/services/uploads", "Method": "POST"},
			{"Path": "/api/v1/services/upload", "Method": "POST"},
			{"Path": "/api/v1/services/:id/enable", "Method": "POST"},
			{"Path": "/api/v1/services/:id/disable", "Method": "POST"},
			{"Path": "/api/v1/services/:id", "Method": "DELETE"},
			{"Path": "/api/v1/branches/:id/services", "Method": "POST"},
			{"Path": "/api/v1/branches/:id/services/:branch_service_id", "Method": "PUT"},
			{"Path": "/api/v1/branches/:id/services/:branch_service_id", "Method": "DELETE"},
		},
		"counter:view": {
			{"Path": "/api/v1/counters", "Method": "GET"},
			{"Path": "/api/v1/counters/:id", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/counters", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id", "Method": "GET"},
		},
		"counter:manage": {
			{"Path": "/api/v1/counters", "Method": "POST"},
			{"Path": "/api/v1/counters/:id", "Method": "PUT"},
			{"Path": "/api/v1/counters/:id", "Method": "DELETE"},
			{"Path": "/api/v1/branches/:id/counters", "Method": "POST"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id", "Method": "PUT"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id", "Method": "PATCH"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id", "Method": "DELETE"},
		},
		"queue-config:view": {
			{"Path": "/api/v1/queue-config/effective", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/effective-config", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/services/:service_id/effective-config", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id/effective-config", "Method": "GET"},
		},
		"queue-config:manage": {
			{"Path": "/api/v1/queue-config", "Method": "PATCH"},
			{"Path": "/api/v1/branches/:id/queue-config", "Method": "PATCH"},
			{"Path": "/api/v1/branches/:id/queue-config/:field", "Method": "DELETE"},
			{"Path": "/api/v1/branches/:id/services/:branch_service_id/queue-config", "Method": "PATCH"},
			{"Path": "/api/v1/branches/:id/services/:branch_service_id/queue-config/:field", "Method": "DELETE"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id/queue-config", "Method": "PATCH"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id/queue-config/:field", "Method": "DELETE"},
		},
		"queue:view": {
			{"Path": "/api/v1/queues", "Method": "GET"},
			{"Path": "/api/v1/queues/:id", "Method": "GET"},
			{"Path": "/api/v1/queues/:id/visit-journeys", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/queue-stats", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/services/:service_id/queue-journeys", "Method": "GET"},
			{"Path": "/api/v1/branches/:id/counters/:counter_id/queue-journeys", "Method": "GET"},
		},
		"queue:manage": {
			{"Path": "/api/v1/queues", "Method": "POST"},
			{"Path": "/api/v1/queues/:id/forward", "Method": "POST"},
			{"Path": "/api/v1/queues/:id/transition", "Method": "POST"},
		},
		"operator_assignment:manage": {
			{"Path": "/api/v1/operator-counter-assignments", "Method": "POST"},
			{"Path": "/api/v1/operator-counter-assignments", "Method": "GET"},
			{"Path": "/api/v1/operator-counter-assignments/:id", "Method": "DELETE"},
		},
		"qms_client:manage": {
			{"Path": "/api/v1/qms-clients", "Method": "POST"},
			{"Path": "/api/v1/qms-clients/credentials", "Method": "POST"},
			{"Path": "/api/v1/qms-clients", "Method": "GET"},
			{"Path": "/api/v1/qms-clients/:id", "Method": "GET"},
			{"Path": "/api/v1/qms-clients/:id", "Method": "PATCH"},
			{"Path": "/api/v1/qms-clients/:id", "Method": "DELETE"},
		},
		"api_key:manage": {
			{"Path": "/api/v1/api-keys", "Method": "POST"},
			{"Path": "/api/v1/api-keys", "Method": "GET"},
			{"Path": "/api/v1/api-keys/:id", "Method": "DELETE"},
		},
		"caller:call": {
			{"Path": "/api/v1/caller/login", "Method": "POST"},
			{"Path": "/api/v1/caller/me", "Method": "GET"},
			{"Path": "/api/v1/caller/queue-journeys/:journey_id/action", "Method": "POST"},
		},
		"signage:view": {
			{"Path": "/api/v1/signage/me", "Method": "GET"},
			{"Path": "/api/v1/signage/current-calls", "Method": "GET"},
			{"Path": "/api/v1/signage/queues", "Method": "GET"},
		},
		"scanner:operate": {
			{"Path": "/api/v1/scanner/check-in", "Method": "POST"},
		},
	}
}

func getRoleToRights() map[string][]string {
	return map[string][]string{
		"role:admin": {
			"dashboard:view",
			"user:view", "user:manage",
			"role:view", "role:manage",
			"org:view", "org:manage",
			"member:manage", "presence:view",
			"audit:view",
			"permission:view", "permission:manage",
			"access:view", "access:manage",
			"webhook:manage",
			"api_key:manage",
			"branch:view", "branch:manage",
			"service:view", "service:manage",
			"counter:view", "counter:manage",
			"queue-config:view", "queue-config:manage",
			"queue:view", "queue:manage",
			"operator_assignment:manage",
			"qms_client:manage",
			"caller:call",
			"signage:view",
			"scanner:operate",
		},
		"role:doctor": {
			"dashboard:view",
			"org:view",
			"branch:view",
			"service:view",
			"counter:view",
			"queue:view",
			"queue:manage",
			"caller:call",
		},
		"role:operator": {
			"dashboard:view",
			"org:view",
			"branch:view",
			"service:view",
			"counter:view",
			"queue:view",
			"queue:manage",
			"caller:call",
			"scanner:operate",
		},
		"role:cashier": {
			"dashboard:view",
			"org:view",
			"branch:view",
			"service:view",
			"counter:view",
			"queue:view",
			"queue:manage",
			"caller:call",
		},
		"role:pharmacist": {
			"dashboard:view",
			"org:view",
			"branch:view",
			"service:view",
			"counter:view",
			"queue:view",
			"queue:manage",
			"caller:call",
		},
		"role:user": {
			"dashboard:view", "org:view", "presence:view",
			"queue:view", "queue:manage",
		},
	}
}

func seedAccessRightsAndPoliciesViaAPI(cfg *config.AppConfig, token string) {
	apiBaseURL := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)

	accessMap := getAccessMap()
	roleToRights := getRoleToRights()

	// 1. Seed Endpoints and AccessRights into DB
	for arName, eps := range accessMap {

		// Search for access right
		searchARPayload := map[string]interface{}{
			"filter": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "equals",
					"from": arName,
				},
			},
		}

		resp, err := doJSONRequest("POST", fmt.Sprintf("%s/access-rights/search", apiBaseURL), searchARPayload, token, http.StatusOK)

		var arID string
		if err == nil && resp["data"] != nil {
			if data, ok := resp["data"].([]interface{}); ok && len(data) > 0 {
				if item, ok := data[0].(map[string]interface{}); ok {
					if id, ok := item["id"].(string); ok {
						arID = id
					}
				}
			}
		} else if err != nil {
			log.Printf("Warning: search access right failed: %v", err)
		}

		if arID == "" {
			// Create it
			createARPayload := map[string]interface{}{
				"name": arName,
			}
			resp, err := doJSONRequest("POST", fmt.Sprintf("%s/access-rights", apiBaseURL), createARPayload, token, http.StatusCreated)
			if err != nil {
				log.Printf("Failed to create access right %s: %v", arName, err)
				continue
			}
			arID = ""
			if resp["data"] != nil {
				if item, ok := resp["data"].(map[string]interface{}); ok {
					if id, ok := item["id"].(string); ok {
						arID = id
					}
				}
			}
		}

		for _, ep := range eps {
			// Search for endpoint
			searchEpPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"path": map[string]interface{}{
						"type": "equals",
						"from": ep["Path"],
					},
					"method": map[string]interface{}{
						"type": "equals",
						"from": ep["Method"],
					},
				},
			}

			resp, err := doJSONRequest("POST", fmt.Sprintf("%s/endpoints/search", apiBaseURL), searchEpPayload, token, http.StatusOK)

			var epID string
			if err == nil && resp["data"] != nil {
				if data, ok := resp["data"].([]interface{}); ok && len(data) > 0 {
					if item, ok := data[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							epID = id
						}
					}
				}
			}

			if epID == "" {
				// Create endpoint
				createEpPayload := map[string]interface{}{
					"path":   ep["Path"],
					"method": ep["Method"],
				}
				resp, err := doJSONRequest("POST", fmt.Sprintf("%s/endpoints", apiBaseURL), createEpPayload, token, http.StatusCreated)
				if err != nil {
					log.Printf("Failed to create endpoint %s %s: %v", ep["Method"], ep["Path"], err)
					continue
				}
				epID = ""
				if resp["data"] != nil {
					if item, ok := resp["data"].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							epID = id
						}
					}
				}
			}

			// Link in DB
			linkPayload := map[string]interface{}{
				"access_right_id": arID,
				"endpoint_id":     epID,
			}
			_, _ = doJSONRequest("POST", fmt.Sprintf("%s/access-rights/link", apiBaseURL), linkPayload, token, http.StatusOK)
			// Ignore error on link if it's already linked
		}
	}

	// 3. Seed Casbin Inheritance: g, role, accessRight, global
	for roleName, rights := range roleToRights {
		for _, arName := range rights {
			// Need to fetch AR ID again, or rely assigning access right
			// find AR ID
			searchARPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "equals",
						"from": arName,
					},
				},
			}
			resp, err := doJSONRequest("POST", fmt.Sprintf("%s/access-rights/search", apiBaseURL), searchARPayload, token, http.StatusOK)
			var arID string
			if err == nil && resp["data"] != nil {
				if data, ok := resp["data"].([]interface{}); ok && len(data) > 0 {
					if item, ok := data[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							arID = id
						}
					}
				}
			}

			if arID != "" {
				assignPayload := map[string]interface{}{
					"role":            roleName,
					"access_right_id": arID,
					"domain":          "global",
				}
				_, err := doJSONRequest("POST", fmt.Sprintf("%s/permissions/assign-access-right", apiBaseURL), assignPayload, token, http.StatusOK)
				if err != nil {
					log.Printf("Warning: failed to assign %s to %s: %v", arName, roleName, err)
				}
			}
		}
	}
}

func ensurePolicy(db *gorm.DB, ptype, v0, v1, v2, v3, v4 string) {
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

func seedOrganizationsAndUsers(cfg *config.AppConfig, token string) {
	apiBaseURL := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)

	// 1. Create Default Organization
	orgPayload := map[string]interface{}{
		"name": "Default Organization",
		"slug": "default-org",
	}

	resp, err := doJSONRequest("POST", fmt.Sprintf("%s/organizations", apiBaseURL), orgPayload, token, http.StatusCreated)
	var orgID string
	if err != nil {
		log.Printf("Failed to create default organization (or it already exists): %v", err)
		// Try to fetch it
		resp, err = doJSONRequest("GET", fmt.Sprintf("%s/organizations/slug/default-org", apiBaseURL), nil, token, http.StatusOK)
		if err == nil && resp["data"] != nil {
			if item, ok := resp["data"].(map[string]interface{}); ok {
				if id, ok := item["id"].(string); ok {
					orgID = id
				}
			}
		}
	} else if resp["data"] != nil {
		if item, ok := resp["data"].(map[string]interface{}); ok {
			if id, ok := item["id"].(string); ok {
				orgID = id
			}
		}
	}

	if orgID == "" {
		log.Println("Could not determine default organization ID. Skipping user seeding.")
		return
	}
	log.Printf("Default Organization seeded/found with ID: %s", orgID)

	// 2. Assign Superadmin as member/owner of default org (often handled by creation, but let's ensure)
	// We might need to invite or just rely on the API auto-assigning the creator.
	// The access control model uses organizations/:id/members.
	// But first let's create the other users.

	usersToSeed := []map[string]string{
		{"username": "adminuser", "fullname": "Admin User", "email": "admin@example.com", "role": "role:admin"},
		{"username": "regularuser", "fullname": "Regular User", "email": "user@example.com", "role": "role:user"},
	}

	for _, u := range usersToSeed {
		userPayload := map[string]interface{}{
			"username": u["username"],
			"password": "Password0!",
			"fullname": u["fullname"],
			"email":    u["email"],
		}

		resp, err := doJSONRequest("POST", fmt.Sprintf("%s/users", apiBaseURL), userPayload, token, http.StatusCreated)
		var userID string
		if err != nil {
			log.Printf("User %s might already exist. Attempting to fetch...", u["username"])
			// Search user by username
			searchPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"username": map[string]interface{}{
						"type": "equals",
						"from": u["username"],
					},
				},
			}
			searchResp, searchErr := doJSONRequest("POST", fmt.Sprintf("%s/users/search", apiBaseURL), searchPayload, token, http.StatusOK)
			if searchErr == nil && searchResp["data"] != nil {
				if dataList, ok := searchResp["data"].([]interface{}); ok && len(dataList) > 0 {
					if item, ok := dataList[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							userID = id
						}
					}
				}
			}
		} else if resp["data"] != nil {
			if item, ok := resp["data"].(map[string]interface{}); ok {
				if id, ok := item["id"].(string); ok {
					userID = id
				}
			}
		}

		if userID != "" {
			log.Printf("User %s seeded/found with ID: %s", u["username"], userID)

			// Get Role ID
			roleSearchPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "equals",
						"from": u["role"],
					},
				},
			}
			roleResp, roleErr := doJSONRequest("POST", fmt.Sprintf("%s/roles/search", apiBaseURL), roleSearchPayload, token, http.StatusOK)
			var roleID string
			if roleErr == nil && roleResp["data"] != nil {
				if list, ok := roleResp["data"].([]interface{}); ok && len(list) > 0 {
					if item, ok := list[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							roleID = id
						}
					}
				}
			}

			if roleID != "" {
				invitePayload := map[string]interface{}{
					"email":   u["email"],
					"user_id": userID,
					"role_id": roleID,
				}
				_, err = doJSONRequest("POST", fmt.Sprintf("%s/organizations/%s/members/invite", apiBaseURL, orgID), invitePayload, token, http.StatusOK)
				if err != nil {
					log.Printf("Warning: Could not invite user %s to org %s: %v", u["username"], orgID, err)
				} else {
					log.Printf("User %s added to organization.", u["username"])
				}
			} else {
				log.Printf("Could not find role %s for user %s", u["role"], u["username"])
			}
		}
	}
}

func seedAccessRightsAndPoliciesDirectly(db *gorm.DB, accessMap map[string][]map[string]string, roleToRights map[string][]string) {
	now := time.Now().UnixMilli()
	for arName, eps := range accessMap {
		var arID string
		row := db.Table("access_rights").Where("name = ?", arName).Select("id").Row()
		_ = row.Scan(&arID)
		if arID == "" {
			arID = uuid.NewString()
			if err := db.Table("access_rights").Create(map[string]interface{}{
				"id":         arID,
				"name":       arName,
				"created_at": now,
				"updated_at": now,
			}).Error; err != nil {
				log.Printf("Failed to create access right %s directly: %v", arName, err)
				continue
			}
		}

		for _, ep := range eps {
			var epID string
			row := db.Table("endpoints").Where("path = ? AND method = ?", ep["Path"], ep["Method"]).Select("id").Row()
			_ = row.Scan(&epID)
			if epID == "" {
				epID = uuid.NewString()
				if err := db.Table("endpoints").Create(map[string]interface{}{
					"id":         epID,
					"path":       ep["Path"],
					"method":     ep["Method"],
					"created_at": now,
					"updated_at": now,
					"deleted_at": 0,
				}).Error; err != nil {
					log.Printf("Failed to create endpoint %s %s directly: %v", ep["Method"], ep["Path"], err)
					continue
				}
			}

			db.Exec("INSERT IGNORE INTO access_right_endpoints (access_right_id, endpoint_id) VALUES (?, ?)", arID, epID)
		}
	}

	for roleName, rights := range roleToRights {
		for _, arName := range rights {
			eps, exists := accessMap[arName]
			if !exists {
				continue
			}
			for _, ep := range eps {
				ensurePolicy(db, "p", roleName, "global", ep["Path"], ep["Method"], "")
			}
		}
	}
}

func seedOrganizationsAndUsersDirectly(db *gorm.DB) {
	now := time.Now().UnixMilli()
	var orgID string
	row := db.Table("organizations").Where("slug = ?", "default-org").Select("id").Row()
	_ = row.Scan(&orgID)
	if orgID == "" {
		orgID = uuid.NewString()
		if err := db.Table("organizations").Create(map[string]interface{}{
			"id":         orgID,
			"name":       "Default Organization",
			"slug":       "default-org",
			"owner_id":   "system",
			"status":     "active",
			"created_at": now,
			"updated_at": now,
		}).Error; err != nil {
			log.Printf("Failed to seed default organization directly: %v", err)
			return
		}
	}

	usersToSeed := []map[string]string{
		{"username": "adminuser", "fullname": "Admin User", "email": "admin@example.com", "role": "role:admin"},
		{"username": "regularuser", "fullname": "Regular User", "email": "user@example.com", "role": "role:user"},
	}

	for _, u := range usersToSeed {
		var userID string
		row := db.Table("users").Where("username = ?", u["username"]).Select("id").Row()
		_ = row.Scan(&userID)
		if userID == "" {
			userID = uuid.NewString()
			hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("Password0!"), bcrypt.DefaultCost)
			if err := db.Table("users").Create(map[string]interface{}{
				"id":         userID,
				"username":   u["username"],
				"email":      u["email"],
				"password":   string(hashedPwd),
				"name":       u["fullname"],
				"status":     "active",
				"created_at": now,
				"updated_at": now,
			}).Error; err != nil {
				log.Printf("Failed to seed user %s directly: %v", u["username"], err)
				continue
			}
		}

		var roleID string
		row = db.Table("roles").Where("name = ?", u["role"]).Select("id").Row()
		_ = row.Scan(&roleID)
		if roleID != "" {
			var memberCount int64
			db.Table("organization_members").Where("organization_id = ? AND user_id = ?", orgID, userID).Count(&memberCount)
			if memberCount == 0 {
				db.Table("organization_members").Create(map[string]interface{}{
					"id":              uuid.NewString(),
					"organization_id": orgID,
					"user_id":         userID,
					"role_id":         roleID,
					"status":          "active",
					"joined_at":       now,
				})
			}
			ensurePolicy(db, "g", userID, u["role"], orgID, "", "")
			ensurePolicy(db, "g", userID, u["role"], "global", "", "")
		}
	}
}
