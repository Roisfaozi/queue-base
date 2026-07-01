//go:build e2e
// +build e2e

package api

import (
	"testing"

	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserE2E(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer)
	}{
		{
			name:     "Positive_GetAllUsers",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)

				f := fixtures.NewUserFactory(server.DB)
				f.Create(func(u *userEntity.User) { u.Username = "user_list_1"; u.Email = "list1@test.com" })
				f.Create(func(u *userEntity.User) { u.Username = "user_list_2"; u.Email = "list2@test.com" })

				resp := server.Client.GET("/api/v1/users", setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data []struct {
						ID       string `json:"id"`
						Username string `json:"username"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(result.Data), 2)
			},
		},
		{
			name:     "Negative_GetAllUsers_Unauthorized",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				resp := server.Client.GET("/api/v1/users")
				assert.Equal(t, 401, resp.StatusCode)
			},
		},
		{
			name:     "Positive_GetUserByID",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)

				f := fixtures.NewUserFactory(server.DB)
				targetUser := f.Create(func(u *userEntity.User) {
					u.Username = "target_by_id"
					u.Email = "target_byid@test.com"
					u.Name = "Target User"
				})

				resp := server.Client.GET("/api/v1/users/"+targetUser.ID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data struct {
						ID       string `json:"id"`
						Username string `json:"username"`
						Name     string `json:"name"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.Equal(t, targetUser.ID, result.Data.ID)
				assert.Equal(t, "Target User", result.Data.Name)
			},
		},
		{
			name:     "Negative_GetUserByID_NotFound",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)
				resp := server.Client.GET("/api/v1/users/nonexistent-id-12345", setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
		{
			name:     "Positive_DeleteUser",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)

				f := fixtures.NewUserFactory(server.DB)
				userToDelete := f.Create(func(u *userEntity.User) {
					u.Username = "user_to_delete"
					u.Email = "delete@test.com"
				})

				resp := server.Client.DELETE("/api/v1/users/"+userToDelete.ID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				resp = server.Client.GET("/api/v1/users/"+userToDelete.ID, setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
		{
			name:     "Negative_DeleteUser_NotFound",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)
				resp := server.Client.DELETE("/api/v1/users/nonexistent-id", setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
		{
			name:     "Positive_BanUser",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)

				f := fixtures.NewUserFactory(server.DB)
				hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)
				targetUser := f.Create(func(u *userEntity.User) {
					u.Username = "status_user_ban"
					u.Email = "statusban@test.com"
					u.Password = string(hash)
					u.Status = userEntity.UserStatusActive
				})

				resp := server.Client.POST("/api/v1/auth/login", map[string]any{
					"username": targetUser.Username,
					"password": "UserPass123!",
				})
				require.Equal(t, 200, resp.StatusCode)
				var userLoginRes struct {
					Data struct {
						AccessToken string `json:"access_token"`
					} `json:"data"`
				}
				resp.JSON(&userLoginRes)
				userToken := userLoginRes.Data.AccessToken

				resp = server.Client.PATCH("/api/v1/users/"+targetUser.ID+"/status",
					map[string]any{"status": "banned"},
					setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				resp = server.Client.GET("/api/v1/users/me", setup.WithAuth(userToken))
				assert.Equal(t, 401, resp.StatusCode, "Banned user's token should be invalidated")
			},
		},
		{
			name:     "Positive_ReactivateUser",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)

				f := fixtures.NewUserFactory(server.DB)
				targetUser := f.Create(func(u *userEntity.User) {
					u.Username = "status_user_reactivate"
					u.Email = "statusreactivate@test.com"
					u.Status = userEntity.UserStatusBanned
				})

				resp := server.Client.PATCH("/api/v1/users/"+targetUser.ID+"/status",
					map[string]any{"status": "active"},
					setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name:     "Negative_UpdateStatus_InvalidStatus",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				adminToken := setup.CreateAdminAndLogin(t, server)
				f := fixtures.NewUserFactory(server.DB)
				targetUser := f.Create(func(u *userEntity.User) {
					u.Username = "status_user_invalid"
					u.Email = "statusinvalid@test.com"
				})

				resp := server.Client.PATCH("/api/v1/users/"+targetUser.ID+"/status",
					map[string]any{"status": "invalid_status"},
					setup.WithAuth(adminToken))
				assert.Equal(t, 422, resp.StatusCode)
			},
		},
		{
			name:     "Security_IDOR_AttackerCannotDeleteVictim",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)
				passHash := string(hash)

				f := fixtures.NewUserFactory(server.DB)
				attacker := f.Create(func(u *userEntity.User) {
					u.Username = "idor_attacker_del"
					u.Email = "attackerdel@test.com"
					u.Password = passHash
					u.Status = userEntity.UserStatusActive
				})

				victim := f.Create(func(u *userEntity.User) {
					u.Username = "idor_victim_del"
					u.Email = "victimdel@test.com"
					u.Password = passHash
					u.Status = userEntity.UserStatusActive
				})

				resp := server.Client.POST("/api/v1/auth/login", map[string]any{
					"username": attacker.Username,
					"password": "UserPass123!",
				})
				require.Equal(t, 200, resp.StatusCode)
				var loginRes struct {
					Data struct {
						AccessToken string `json:"access_token"`
					} `json:"data"`
				}
				resp.JSON(&loginRes)
				attackerToken := loginRes.Data.AccessToken

				resp = server.Client.DELETE("/api/v1/users/"+victim.ID,
					setup.WithAuth(attackerToken))
				assert.True(t, resp.StatusCode == 403 || resp.StatusCode == 401,
					"Non-admin user must NOT be able to delete another user's account (got %d)", resp.StatusCode)
			},
		},
		{
			name:     "Security_IDOR_AttackerCannotChangeVictimStatus",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)
				passHash := string(hash)

				f := fixtures.NewUserFactory(server.DB)
				attacker := f.Create(func(u *userEntity.User) {
					u.Username = "idor_attacker_status"
					u.Email = "attackerstatus@test.com"
					u.Password = passHash
					u.Status = userEntity.UserStatusActive
				})

				victim := f.Create(func(u *userEntity.User) {
					u.Username = "idor_victim_status"
					u.Email = "victimstatus@test.com"
					u.Password = passHash
					u.Status = userEntity.UserStatusActive
				})

				resp := server.Client.POST("/api/v1/auth/login", map[string]any{
					"username": attacker.Username,
					"password": "UserPass123!",
				})
				require.Equal(t, 200, resp.StatusCode)
				var loginRes struct {
					Data struct {
						AccessToken string `json:"access_token"`
					} `json:"data"`
				}
				resp.JSON(&loginRes)
				attackerToken := loginRes.Data.AccessToken

				resp = server.Client.PATCH("/api/v1/users/"+victim.ID+"/status",
					map[string]any{"status": "banned"},
					setup.WithAuth(attackerToken))
				assert.True(t, resp.StatusCode == 403 || resp.StatusCode == 401,
					"Non-admin user must NOT be able to change another user's status (got %d)", resp.StatusCode)
			},
		},
		{
			name:     "Positive_UserCanUpdateOwnProfile",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)
				passHash := string(hash)

				f := fixtures.NewUserFactory(server.DB)
				user := f.Create(func(u *userEntity.User) {
					u.Username = "self_update_user"
					u.Email = "selfupdate@test.com"
					u.Password = passHash
					u.Status = userEntity.UserStatusActive
				})

				resp := server.Client.POST("/api/v1/auth/login", map[string]any{
					"username": user.Username,
					"password": "UserPass123!",
				})
				require.Equal(t, 200, resp.StatusCode)
				var loginRes struct {
					Data struct {
						AccessToken string `json:"access_token"`
					} `json:"data"`
				}
				resp.JSON(&loginRes)
				userToken := loginRes.Data.AccessToken

				updatePayload := map[string]any{
					"name": "User Updated Own",
				}
				resp = server.Client.PUT("/api/v1/users/me",
					updatePayload,
					setup.WithAuth(userToken))
				assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 422,
					"Self-update via /me returns 200 (success) or 422 (validation), got %d", resp.StatusCode)
			},
		},
		{
			name:     "Security_AdminBanUserAccessDenied",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				f := fixtures.NewUserFactory(server.DB)

				hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				passHash := string(hash)

				adminUser := f.Create(func(u *userEntity.User) {
					u.Username = "admin_banner"
					u.Email = "admin@ban.com"
					u.Password = passHash
				})
				server.Enforcer.AddGroupingPolicy(adminUser.ID, "role:superadmin", "global")

				targetUser := f.Create(func(u *userEntity.User) {
					u.Username = "target_user"
					u.Email = "target@ban.com"
					u.Password = passHash
				})
				server.Enforcer.AddGroupingPolicy(targetUser.ID, "role:user", "global")

				loginPayload := map[string]any{"username": targetUser.Username, "password": "StrongPass123!"}
				resp := server.Client.POST("/api/v1/auth/login", loginPayload)
				require.Equal(t, 200, resp.StatusCode)

				var loginRes struct {
					Data struct {
						AccessToken string `json:"access_token"`
					}
				}
				resp.JSON(&loginRes)
				targetToken := loginRes.Data.AccessToken

				resp = server.Client.GET("/api/v1/users/me", setup.WithAuth(targetToken))
				assert.Equal(t, 200, resp.StatusCode)

				server.DB.Model(&userEntity.User{}).Where("id = ?", targetUser.ID).Update("status", userEntity.UserStatusBanned)

				resp = server.Client.GET("/api/v1/users/me", setup.WithAuth(targetToken))
				assert.Equal(t, 403, resp.StatusCode, "Banned user should be denied access")
			},
		},
		{
			name:     "Positive_DynamicSearch",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				f := fixtures.NewUserFactory(server.DB)
				hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				passHash := string(hash)

				admin := f.Create(func(u *userEntity.User) {
					u.Username = "search_admin"
					u.Email = "search@admin.com"
					u.Password = passHash
				})
				server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
				server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
				server.Enforcer.SavePolicy()

				resp := server.Client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
				var loginRes struct {
					Data struct {
						AccessToken string `json:"access_token"`
					}
				}
				resp.JSON(&loginRes)
				token := loginRes.Data.AccessToken

				f.Create(func(u *userEntity.User) {
					u.Name = "Alice Wonderland"
					u.Email = "alice@test.com"
					u.Username = "alice_w"
				})
				f.Create(func(u *userEntity.User) {
					u.Name = "Bob Builder"
					u.Email = "bob@test.com"
					u.Username = "bob_b"
				})
				f.Create(func(u *userEntity.User) {
					u.Name = "Charlie Chocolate"
					u.Email = "charlie@test.com"
					u.Username = "charlie_c"
				})

				searchPayload := map[string]any{
					"filter": map[string]any{
						"email": map[string]any{"type": "contains", "from": "alice"},
					},
				}
				resp = server.Client.POST("/api/v1/users/search", searchPayload, setup.WithAuth(token))
				require.Equal(t, 200, resp.StatusCode)

				var userSearchRes struct {
					Data []struct {
						Email string `json:"email"`
						Name  string `json:"name"`
					} `json:"data"`
					Paging struct {
						Total int64 `json:"total"`
					} `json:"paging"`
				}
				resp.JSON(&userSearchRes)

				if userSearchRes.Paging.Total == 0 {
					t.Logf("DEBUG: Search Result Empty. Payload: %+v. Response: %+v", searchPayload, userSearchRes)
				}

				assert.Equal(t, int64(1), userSearchRes.Paging.Total)
				if len(userSearchRes.Data) > 0 {
					assert.Equal(t, "alice@test.com", userSearchRes.Data[0].Email)
				}

				sortPayload := map[string]any{
					"sort": []map[string]any{
						{"colId": "name", "sort": "desc"},
					},
					"page_size": 5,
				}
				resp = server.Client.POST("/api/v1/users/search", sortPayload, setup.WithAuth(token))
				require.Equal(t, 200, resp.StatusCode)

				var sortedRes struct {
					Data []struct {
						Name string `json:"name"`
					} `json:"data"`
				}
				resp.JSON(&sortedRes)

				assert.NotEmpty(t, sortedRes.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setup.SetupTestServer(t)
			defer server.Cleanup()
			tt.run(t, server)
		})
	}
}
