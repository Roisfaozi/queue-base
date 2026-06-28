package api

import (
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/access/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/model"
	roleEntity "github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPermissionE2E_RoleHierarchy(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "hier_admin"
		u.Email = "hier@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)

	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Supervisor"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Intern"})

	grantPayload := map[string]any{
		"role": "Intern", "path": "/api/v1/coffee", "method": "GET",
	}
	resp = client.POST("/api/v1/permissions/grant", grantPayload, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)

	supervisorUser := f.Create(func(u *userEntity.User) { u.Username = "supervisor"; u.Password = passHash })
	internUser := f.Create(func(u *userEntity.User) { u.Username = "intern"; u.Password = passHash })

	client.POST("/api/v1/permissions/assign-role", map[string]any{"user_id": supervisorUser.ID, "role": "Supervisor"}, setup.WithAuth(adminToken))
	client.POST("/api/v1/permissions/assign-role", map[string]any{"user_id": internUser.ID, "role": "Intern"}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": "supervisor", "password": "StrongPass123!"})
	var supLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&supLoginRes)
	require.NoError(t, err)
	supToken := supLoginRes.Data.AccessToken

	batchReq := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/api/v1/coffee", Action: "GET"},
		},
	}
	resp = client.POST("/api/v1/permissions/check-batch", batchReq, setup.WithAuth(supToken))
	require.Equal(t, 200, resp.StatusCode)
	var checkRes struct {
		Data model.BatchPermissionCheckResponse `json:"data"`
	}
	err = resp.JSON(&checkRes)
	require.NoError(t, err)

	checks := []struct {
		name     string
		expected bool
	}{
		{name: "BeforeInheritance", expected: false},
		{name: "AfterInheritance", expected: true},
	}

	assert.Equal(t, checks[0].expected, checkRes.Data.Results["/api/v1/coffee:GET"], "Supervisor should NOT have access yet")

	inheritPayload := map[string]any{
		"child_role":  "Supervisor",
		"parent_role": "Intern",
	}
	resp = client.POST("/api/v1/permissions/inheritance", inheritPayload, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	resp = client.POST("/api/v1/permissions/check-batch", batchReq, setup.WithAuth(supToken))
	require.Equal(t, 200, resp.StatusCode)
	err = resp.JSON(&checkRes)
	require.NoError(t, err)

	assert.Equal(t, checks[1].expected, checkRes.Data.Results["/api/v1/coffee:GET"], "Supervisor SHOULD have access after inheritance")
}

func TestPermissionE2E_BatchCheck(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) { u.Username = "batch_admin"; u.Password = passHash })
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	var adminRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&adminRes)
	require.NoError(t, err)
	adminToken := adminRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Editor"})

	client.POST("/api/v1/permissions/grant", map[string]any{"role": "Editor", "path": "/news", "method": "READ"}, setup.WithAuth(adminToken))
	client.POST("/api/v1/permissions/grant", map[string]any{"role": "Editor", "path": "/news", "method": "WRITE"}, setup.WithAuth(adminToken))

	user := f.Create(func(u *userEntity.User) { u.Username = "editor_user"; u.Password = passHash })
	client.POST("/api/v1/permissions/assign-role", map[string]any{"user_id": user.ID, "role": "Editor"}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": user.Username, "password": "StrongPass123!"})
	var userRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&userRes)
	require.NoError(t, err)
	userToken := userRes.Data.AccessToken

	roles, _ := server.Enforcer.GetRolesForUser(user.ID, "global")
	t.Logf("DEBUG: Roles for user %s: %v", user.ID, roles)

	tests := []struct {
		name     string
		item     model.PermissionCheckItem
		expected bool
	}{
		{name: "NewsRead", item: model.PermissionCheckItem{Resource: "/news", Action: "READ"}, expected: true},
		{name: "NewsWrite", item: model.PermissionCheckItem{Resource: "/news", Action: "WRITE"}, expected: true},
		{name: "NewsDelete", item: model.PermissionCheckItem{Resource: "/news", Action: "DELETE"}, expected: false},
		{name: "AdminGet", item: model.PermissionCheckItem{Resource: "/admin", Action: "GET"}, expected: false},
	}

	req := model.BatchPermissionCheckRequest{Items: make([]model.PermissionCheckItem, 0, len(tests))}
	for _, tt := range tests {
		req.Items = append(req.Items, tt.item)
	}

	resp = client.POST("/api/v1/permissions/check-batch", req, setup.WithAuth(userToken))
	require.Equal(t, 200, resp.StatusCode)

	var checkRes struct {
		Data model.BatchPermissionCheckResponse `json:"data"`
	}
	err = resp.JSON(&checkRes)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, checkRes.Data.Results[tt.item.Resource+":"+tt.item.Action])
		})
	}
}

func TestPermissionE2E_RevokeRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "revoke_admin"
		u.Email = "revoke@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "RevokeTestRole"})
	user := f.Create(func(u *userEntity.User) { u.Username = "revoke_user"; u.Password = passHash })

	resp = client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    "RevokeTestRole",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	roles, _ := server.Enforcer.GetRolesForUser(user.ID, "global")
	assert.Contains(t, roles, "RevokeTestRole")

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_RevokeRole",
			category: "positive",
			run: func(t *testing.T) {
				resp := client.DELETE("/api/v1/permissions/revoke-role", map[string]any{
					"user_id": user.ID,
					"role":    "RevokeTestRole",
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				rolesAfter, _ := server.Enforcer.GetRolesForUser(user.ID, "global")
				assert.NotContains(t, rolesAfter, "RevokeTestRole")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestPermissionE2E_RemoveInheritance(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "remove_inherit_admin"
		u.Email = "remove_inherit@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "ParentRole"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "ChildRole"})

	client.POST("/api/v1/permissions/grant", map[string]any{
		"role": "ChildRole", "path": "/api/v1/inherited", "method": "GET",
	}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/permissions/inheritance", map[string]any{
		"child_role":  "ParentRole",
		"parent_role": "ChildRole",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	checks := []struct {
		name     string
		expected bool
	}{
		{name: "BeforeRemoveInheritance", expected: true},
		{name: "AfterRemoveInheritance", expected: false},
	}

	ok, _ := server.Enforcer.Enforce("ParentRole", "global", "/api/v1/inherited", "GET")
	assert.Equal(t, checks[0].expected, ok, "Parent should have access via inheritance")

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_RemoveInheritance",
			category: "positive",
			run: func(t *testing.T) {
				resp := client.DELETE("/api/v1/permissions/inheritance", map[string]any{
					"child_role":  "ParentRole",
					"parent_role": "ChildRole",
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				ok, _ := server.Enforcer.Enforce("ParentRole", "global", "/api/v1/inherited", "GET")
				assert.Equal(t, checks[1].expected, ok, "Parent should NOT have access after inheritance removed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestPermissionE2E_Security_PostRoleDeletion_PermissionsRevoked(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "post_del_admin"
		u.Email = "postdel@admin.com"
		u.Password = passHash
	})
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var adminLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&adminLoginRes)
	require.NoError(t, err)
	adminToken := adminLoginRes.Data.AccessToken

	roleID := uuid.New().String()
	server.DB.Create(&roleEntity.Role{ID: roleID, Name: "EphemeralRole"})
	_, err = server.Enforcer.AddPolicy("EphemeralRole", "global", "/api/v1/secret", "GET")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	targetUser := f.Create(func(u *userEntity.User) { u.Username = "ephemeral_user"; u.Password = passHash })
	client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": targetUser.ID,
		"role":    "EphemeralRole",
	}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": targetUser.Username, "password": "StrongPass123!"})
	var targetLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&targetLoginRes)
	require.NoError(t, err)
	targetToken := targetLoginRes.Data.AccessToken

	roles, _ := server.Enforcer.GetRolesForUser(targetUser.ID, "global")
	assert.Contains(t, roles, "EphemeralRole", "User should have role before deletion")

	resp = client.DELETE("/api/v1/roles/"+roleID, setup.WithAuth(adminToken))
	assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 204, "Role deletion should succeed")

	// Explicitly reload policy in the test enforcer to reflect DB changes
	err = server.Enforcer.LoadPolicy()
	require.NoError(t, err)

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "PoliciesRemovedAfterRoleDeletion",
			category: "positive",
			run: func(t *testing.T) {
				policies, _ := server.Enforcer.GetFilteredPolicy(0, "EphemeralRole")
				assert.Empty(t, policies, "All Casbin policies for deleted role must be removed")
			},
		},
		{
			name:     "RoleAssignmentRemovedAfterRoleDeletion",
			category: "positive",
			run: func(t *testing.T) {
				rolesAfter, _ := server.Enforcer.GetRolesForUser(targetUser.ID, "global")
				assert.NotContains(t, rolesAfter, "EphemeralRole", "User's role assignment must be removed after role deletion")
			},
		},
		{
			name:     "AccessRevokedAfterRoleDeletion",
			category: "positive",
			run: func(t *testing.T) {
				_ = targetToken
				hasAccess, _ := server.Enforcer.Enforce(targetUser.ID, "global", "/api/v1/secret", "GET")
				assert.False(t, hasAccess, "User must NOT have access after role deletion")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestPermissionE2E_AccessRightAssignment(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "ar_admin"
		u.Email = "ar@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	ep1 := entity.Endpoint{Path: "/api/test", Method: "GET"}
	ep2 := entity.Endpoint{Path: "/api/test", Method: "POST"}
	server.DB.Create(&ep1)
	server.DB.Create(&ep2)

	ar := entity.AccessRight{
		Name:      "Test Access Right",
		Endpoints: []entity.Endpoint{ep1, ep2},
	}
	server.DB.Create(&ar)

	roleName := "TestRole"

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "GetRoleAccessRights_Unassigned",
			category: "positive",
			run: func(t *testing.T) {
				resp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
				require.Equal(t, 200, resp.StatusCode)

				var res struct {
					Data []model.RoleAccessRightStatus `json:"data"`
				}
				err = resp.JSON(&res)
				require.NoError(t, err)

				found := false
				for _, s := range res.Data {
					if s.ID == ar.ID {
						found = true
						assert.False(t, s.Assigned)
						assert.False(t, s.Partial)
					}
				}
				assert.True(t, found, "Access right should be in the list")
			},
		},
		{
			name:     "AssignAccessRight",
			category: "positive",
			run: func(t *testing.T) {
				payload := map[string]any{
					"role":            roleName,
					"access_right_id": ar.ID,
				}
				resp := client.POST("/api/v1/permissions/assign-access-right", payload, setup.WithAuth(adminToken))
				require.Equal(t, 200, resp.StatusCode)

				ok, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "GET")
				assert.True(t, ok)
				ok2, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "POST")
				assert.True(t, ok2)
			},
		},
		{
			name:     "GetRoleAccessRights_Assigned",
			category: "positive",
			run: func(t *testing.T) {
				resp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
				require.Equal(t, 200, resp.StatusCode)

				var res struct {
					Data []model.RoleAccessRightStatus `json:"data"`
				}
				err = resp.JSON(&res)
				require.NoError(t, err)

				for _, s := range res.Data {
					if s.ID == ar.ID {
						assert.True(t, s.Assigned)
					}
				}
			},
		},
		{
			name:     "RevokeAccessRight",
			category: "positive",
			run: func(t *testing.T) {
				payload := map[string]any{
					"role":            roleName,
					"access_right_id": ar.ID,
				}
				resp := client.DELETE("/api/v1/permissions/revoke-access-right", payload, setup.WithAuth(adminToken))
				require.Equal(t, 200, resp.StatusCode)

				ok, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "GET")
				assert.False(t, ok)
				ok2, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "POST")
				assert.False(t, ok2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestAccessRightsFlowE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("AdminPass123!"), bcrypt.DefaultCost)
	admin := f.Create(func(u *userEntity.User) {
		u.Username = "flow_admin"
		u.Password = string(hash)
	})
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "AdminPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	resp = client.POST("/api/v1/access-rights", map[string]any{
		"name":        "User Management",
		"description": "Operations related to user accounts",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var arRes struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	err = resp.JSON(&arRes)
	require.NoError(t, err)
	arID := arRes.Data.ID

	resp = client.POST("/api/v1/endpoints", map[string]any{
		"path":   "/api/v1/users",
		"method": "GET",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var epRes struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	err = resp.JSON(&epRes)
	require.NoError(t, err)
	epID := epRes.Data.ID

	resp = client.POST("/api/v1/access-rights/link", map[string]any{
		"access_right_id": arID,
		"endpoint_id":     epID,
	}, setup.WithAuth(adminToken))
	assert.Equal(t, 200, resp.StatusCode)

	roleID := uuid.New().String()
	server.DB.Create(&roleEntity.Role{ID: roleID, Name: "UserManager"})

	resp = client.POST("/api/v1/permissions/grant", map[string]any{
		"role":   "UserManager",
		"path":   "/api/v1/users",
		"method": "GET",
	}, setup.WithAuth(adminToken))
	assert.Equal(t, 201, resp.StatusCode)

	user := f.Create(func(u *userEntity.User) { u.Username = "manager_user"; u.Password = string(hash) })
	client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    "UserManager",
	}, setup.WithAuth(adminToken))

	tests := []struct {
		name     string
		expected bool
	}{
		{name: "UserGetsLinkedEndpointAccess", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := server.Enforcer.Enforce(user.ID, "global", "/api/v1/users", "GET")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, ok, "User should have access to linked endpoint")
		})
	}
}

func TestPermissionE2E_EndpointRegistrationToRoleAccess_Monolith(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client
	type flowState struct {
		adminToken          string
		nonAdminToken       string
		endpointPath        string
		accessRightID       string
		roleName            string
		roleID              string
		userID              string
		nonAdminID          string
		foreignUserID       string
		userHasRoleAccess   bool
		roleHasDirectAccess bool
	}

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("EndpointFlow123!"), bcrypt.DefaultCost)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "endpoint_flow_admin"
		u.Password = string(hash)
	})
	nonAdmin := f.Create(func(u *userEntity.User) {
		u.Username = "endpoint_flow_non_admin"
		u.Password = string(hash)
	})
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginResp := client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "EndpointFlow123!",
	})
	require.Equal(t, 200, loginResp.StatusCode)
	var loginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, loginResp.JSON(&loginData))
	adminToken := loginData.Data.AccessToken

	nonAdminLoginResp := client.POST("/api/v1/auth/login", map[string]any{
		"username": nonAdmin.Username,
		"password": "EndpointFlow123!",
	})
	require.Equal(t, 200, nonAdminLoginResp.StatusCode)
	var nonAdminLoginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, nonAdminLoginResp.JSON(&nonAdminLoginData))
	nonAdminToken := nonAdminLoginData.Data.AccessToken
	state := &flowState{adminToken: adminToken, nonAdminToken: nonAdminToken, nonAdminID: nonAdmin.ID}

	endpointPath := "/api/v1/monolith-endpoint-" + uuid.New().String()[:8]
	accessRightName := "MonolithAccess_" + uuid.New().String()[:8]
	roleName := "MonolithRole_" + uuid.New().String()[:8]
	foreignUser := f.Create(func(u *userEntity.User) {
		u.Username = "endpoint_flow_foreign_user"
		u.Password = string(hash)
	})
	state.endpointPath = endpointPath
	state.roleName = roleName
	state.foreignUserID = foreignUser.ID

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, state *flowState)
	}{
		{
			name:     "NonAdminCannotRegisterEndpoint",
			category: "security",
			run: func(t *testing.T, state *flowState) {
				resp := client.POST("/api/v1/endpoints", map[string]any{
					"path":   "/api/v1/non-admin-blocked-" + uuid.New().String()[:8],
					"method": "GET",
				}, setup.WithAuth(state.nonAdminToken))
				assert.Equal(t, 403, resp.StatusCode)
			},
		},
	}

	createAccessRightResp := client.POST("/api/v1/access-rights", map[string]any{
		"name":        accessRightName,
		"description": "Monolith endpoint flow",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, createAccessRightResp.StatusCode)
	var accessRightData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, createAccessRightResp.JSON(&accessRightData))
	state.accessRightID = accessRightData.Data.ID

	createEndpointResp := client.POST("/api/v1/endpoints", map[string]any{
		"path":   endpointPath,
		"method": "GET",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, createEndpointResp.StatusCode)
	var endpointData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, createEndpointResp.JSON(&endpointData))
	state.roleID = uuid.New().String()
	tests = append(tests, struct {
		name     string
		category string
		run      func(t *testing.T, state *flowState)
	}{
		name:     "InvalidLinkPayloadRejected",
		category: "negative",
		run: func(t *testing.T, state *flowState) {
			resp := client.POST("/api/v1/access-rights/link", map[string]any{
				"access_right_id": "",
				"endpoint_id":     "",
			}, setup.WithAuth(state.adminToken))
			assert.Equal(t, 422, resp.StatusCode)
		},
	})

	linkResp := client.POST("/api/v1/access-rights/link", map[string]any{
		"access_right_id": accessRightData.Data.ID,
		"endpoint_id":     endpointData.Data.ID,
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, linkResp.StatusCode)

	require.NoError(t, server.DB.Create(&roleEntity.Role{ID: state.roleID, Name: roleName}).Error)

	roleAccessBeforeResp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
	require.Equal(t, 200, roleAccessBeforeResp.StatusCode)
	var roleAccessBefore struct {
		Data []model.RoleAccessRightStatus `json:"data"`
	}
	require.NoError(t, roleAccessBeforeResp.JSON(&roleAccessBefore))
	foundUnassigned := false
	for _, status := range roleAccessBefore.Data {
		if status.ID == accessRightData.Data.ID {
			foundUnassigned = true
			assert.False(t, status.Assigned)
		}
	}
	tests = append(tests,
		struct {
			name     string
			category string
			run      func(t *testing.T, state *flowState)
		}{
			name:     "RoleAccessRightInitiallyUnassigned",
			category: "edge",
			run: func(t *testing.T, state *flowState) {
				assert.True(t, foundUnassigned)
			},
		},
		struct {
			name     string
			category string
			run      func(t *testing.T, state *flowState)
		}{
			name:     "UserWithoutRoleCannotAccessRegisteredEndpoint",
			category: "security",
			run: func(t *testing.T, state *flowState) {
				hasAccess, err := server.Enforcer.Enforce(state.nonAdminID, "global", state.endpointPath, "GET")
				require.NoError(t, err)
				assert.False(t, hasAccess)
			},
		},
	)

	assignAccessRightResp := client.POST("/api/v1/permissions/assign-access-right", map[string]any{
		"role":            roleName,
		"access_right_id": accessRightData.Data.ID,
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, assignAccessRightResp.StatusCode)
	state.roleHasDirectAccess, err = server.Enforcer.Enforce(roleName, "global", endpointPath, "GET")
	require.NoError(t, err)
	tests = append(tests, struct {
		name     string
		category string
		run      func(t *testing.T, state *flowState)
	}{
		name:     "RoleReceivesEndpointAccessAfterAssignment",
		category: "positive",
		run: func(t *testing.T, state *flowState) {
			assert.True(t, state.roleHasDirectAccess)
		},
	})

	user := f.Create(func(u *userEntity.User) {
		u.Username = "endpoint_flow_user"
		u.Password = string(hash)
	})

	assignRoleResp := client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    roleName,
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, assignRoleResp.StatusCode)
	state.userID = user.ID
	state.userHasRoleAccess, err = server.Enforcer.Enforce(user.ID, "global", endpointPath, "GET")
	require.NoError(t, err)
	tests = append(tests,
		struct {
			name     string
			category string
			run      func(t *testing.T, state *flowState)
		}{
			name:     "AssignedUserReceivesEndpointAccess",
			category: "positive",
			run: func(t *testing.T, state *flowState) {
				assert.True(t, state.userHasRoleAccess)
			},
		},
		struct {
			name     string
			category string
			run      func(t *testing.T, state *flowState)
		}{
			name:     "UnassignedOtherUserDoesNotInheritAccess",
			category: "vulnerability",
			run: func(t *testing.T, state *flowState) {
				hasAccess, err := server.Enforcer.Enforce(state.foreignUserID, "global", state.endpointPath, "GET")
				require.NoError(t, err)
				assert.False(t, hasAccess)
			},
		},
	)

	revokeAccessRightResp := client.DELETE("/api/v1/permissions/revoke-access-right", map[string]any{
		"role":            roleName,
		"access_right_id": accessRightData.Data.ID,
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, revokeAccessRightResp.StatusCode)
	tests = append(tests,
		struct {
			name     string
			category string
			run      func(t *testing.T, state *flowState)
		}{
			name:     "RevokeRemovesRoleAccess",
			category: "edge",
			run: func(t *testing.T, state *flowState) {
				roleHasAccessAfterRevoke, err := server.Enforcer.Enforce(state.roleName, "global", state.endpointPath, "GET")
				require.NoError(t, err)
				assert.False(t, roleHasAccessAfterRevoke)
			},
		},
		struct {
			name     string
			category string
			run      func(t *testing.T, state *flowState)
		}{
			name:     "RevokeRemovesUserAccessImmediately",
			category: "vulnerability",
			run: func(t *testing.T, state *flowState) {
				userHasAccessAfterRevoke, err := server.Enforcer.Enforce(state.userID, "global", state.endpointPath, "GET")
				require.NoError(t, err)
				assert.False(t, userHasAccessAfterRevoke)
			},
		},
	)

	for _, tt := range tests {
		t.Run(tt.category+"_"+tt.name, func(t *testing.T) {
			tt.run(t, state)
		})
	}
}

func TestSecurityE2E_DynamicRBAC(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	user := f.Create(func(u *userEntity.User) {
		u.Username = "rbac_user"
		u.Email = "rbac@test.com"
		u.Password = passHash
	})

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "super_admin"
		u.Email = "super@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "/api/v1/permissions/grant", "POST")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "/api/v1/permissions/assign-role", "POST")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "/api/v1/roles", "GET")
	require.NoError(t, err)

	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginPayload := map[string]any{"username": user.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	userToken := loginRes.Data.AccessToken

	adminLoginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	respAdmin := client.POST("/api/v1/auth/login", adminLoginPayload)
	var adminLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = respAdmin.JSON(&adminLoginRes)
	require.NoError(t, err)
	adminToken := adminLoginRes.Data.AccessToken

	tests := []struct {
		name        string
		beforeGrant bool
		expected    int
	}{
		{name: "BeforeGrant", beforeGrant: true, expected: 403},
		{name: "AfterGrant", beforeGrant: false, expected: 200},
	}

	resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))
	assert.Equal(t, tests[0].expected, resp.StatusCode)

	roleName := "DynamicViewer"

	server.DB.Create(&roleEntity.Role{Name: roleName})

	grantPayload := map[string]any{
		"role":   roleName,
		"path":   "/api/v1/roles",
		"method": "GET",
	}
	resp = client.POST("/api/v1/permissions/grant", grantPayload, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)

	assignPayload := map[string]any{

		"user_id": user.ID,

		"role": roleName,
	}

	resp = client.POST("/api/v1/permissions/assign-role", assignPayload, setup.WithAuth(adminToken))

	require.Equal(t, 200, resp.StatusCode)

	userRoles, _ := server.Enforcer.GetRolesForUser(user.ID, "global")

	t.Logf("DEBUG: Roles for user %s: %v", user.ID, userRoles)

	hasPermission, _ := server.Enforcer.HasPolicy(roleName, "global", "/api/v1/roles", "GET")

	t.Logf("DEBUG: Role %s has permission GET /api/v1/roles: %v", roleName, hasPermission)

	err = server.Enforcer.LoadPolicy()
	require.NoError(t, err)

	resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))
	assert.Equal(t, tests[1].expected, resp.StatusCode)

}
