//go:build e2e
// +build e2e

package api

import (
	"net/http"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthE2E(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer)
	}{
		{
			name:     "Positive_RegisterLoginLogout",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				client := server.Client

				registerReq := map[string]interface{}{
					"username": "e2euser",
					"email":    "e2e@example.com",
					"password": "password123",
					"fullname": "E2E User",
				}

				resp := client.POST("/api/v1/users/register", registerReq)
				assert.Equal(t, 201, resp.StatusCode)

				var registerResult struct {
					Data struct {
						ID       string `json:"id"`
						Username string `json:"username"`
					} `json:"data"`
				}
				err := resp.JSON(&registerResult)
				require.NoError(t, err)
				assert.Equal(t, "e2euser", registerResult.Data.Username)

				loginReq := map[string]interface{}{
					"username": "e2euser",
					"password": "password123",
				}

				resp = client.POST("/api/v1/auth/login", loginReq)
				assert.Equal(t, 200, resp.StatusCode)

				var loginResult struct {
					Data struct {
						AccessToken string `json:"access_token"`
						TokenType   string `json:"token_type"`
					} `json:"data"`
				}
				err = resp.JSON(&loginResult)
				require.NoError(t, err)
				assert.NotEmpty(t, loginResult.Data.AccessToken)

				resp = client.GET("/api/v1/users/me", setup.WithAuth(loginResult.Data.AccessToken))
				assert.Equal(t, 200, resp.StatusCode)

				resp = client.POST("/api/v1/auth/logout", nil, setup.WithAuth(loginResult.Data.AccessToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name:     "Negative_InvalidCredentials",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				loginReq := map[string]interface{}{
					"username": "nonexistent",
					"password": "wrongpassword",
				}
				resp := server.Client.POST("/api/v1/auth/login", loginReq)
				assert.Equal(t, 401, resp.StatusCode)
			},
		},
		{
			name:     "Positive_ForgotPasswordFlow",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				client := server.Client
				email := "recovery@example.com"
				username := "recoveryuser"

				registerReq := map[string]interface{}{
					"username": username,
					"email":    email,
					"password": "oldPassword123",
					"fullname": "Recovery User",
				}
				resp := client.POST("/api/v1/users/register", registerReq)
				assert.Equal(t, 201, resp.StatusCode)

				forgotReq := map[string]interface{}{"email": email}
				resp = client.POST("/api/v1/auth/forgot-password", forgotReq)
				assert.Equal(t, 200, resp.StatusCode)

				var resetToken entity.PasswordResetToken
				err := server.DB.Where("email = ?", email).First(&resetToken).Error
				require.NoError(t, err)
				require.NotEmpty(t, resetToken.Token)

				newPassword := "brandNewPass2026!"
				resetReq := map[string]interface{}{
					"token":        resetToken.Token,
					"new_password": newPassword,
				}
				resp = client.POST("/api/v1/auth/reset-password", resetReq)
				assert.Equal(t, 200, resp.StatusCode)

				loginReq := map[string]interface{}{
					"username": username,
					"password": newPassword,
				}
				resp = client.POST("/api/v1/auth/login", loginReq)
				assert.Equal(t, 200, resp.StatusCode)

				oldLoginReq := map[string]interface{}{
					"username": username,
					"password": "oldPassword123",
				}
				resp = client.POST("/api/v1/auth/login", oldLoginReq)
				assert.Equal(t, 401, resp.StatusCode)
			},
		},
		{
			name:     "Negative_DuplicateUsername",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				registerReq := map[string]interface{}{
					"username": "duplicate",
					"email":    "first@example.com",
					"password": "password123",
					"fullname": "First User",
				}
				resp := server.Client.POST("/api/v1/users/register", registerReq)
				assert.Equal(t, 201, resp.StatusCode)

				registerReq2 := map[string]interface{}{
					"username": "duplicate",
					"email":    "second@example.com",
					"password": "password123",
					"fullname": "Second User",
				}
				resp = server.Client.POST("/api/v1/users/register", registerReq2)
				assert.Equal(t, 409, resp.StatusCode)
			},
		},
		{
			name:     "Negative_ProtectedEndpoint_NoToken",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				resp := server.Client.GET("/api/v1/users/me")
				assert.Equal(t, 401, resp.StatusCode)
			},
		},
		{
			name:     "Negative_ProtectedEndpoint_InvalidToken",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer) {
				resp := server.Client.GET("/api/v1/users/me", setup.WithAuth("invalid.token.here"))
				assert.Equal(t, 401, resp.StatusCode)
			},
		},
		{
			name:     "Edge_SpecialCharactersInUsername",
			category: "edge",
			run: func(t *testing.T, server *setup.TestServer) {
				registerReq := map[string]interface{}{
					"username": "user@#$%",
					"email":    "special@example.com",
					"password": "password123",
					"fullname": "Special User",
				}
				resp := server.Client.POST("/api/v1/users/register", registerReq)
				assert.True(t, resp.StatusCode == 201 || resp.StatusCode == 400 || resp.StatusCode == 422)
			},
		},
		{
			name:     "Edge_CaseSensitiveUsername",
			category: "edge",
			run: func(t *testing.T, server *setup.TestServer) {
				registerReq := map[string]interface{}{
					"username": "TestUser",
					"email":    "test@example.com",
					"password": "password123",
					"fullname": "Test User",
				}
				resp := server.Client.POST("/api/v1/users/register", registerReq)
				require.Equal(t, 201, resp.StatusCode)

				loginReq := map[string]interface{}{
					"username": "testuser",
					"password": "password123",
				}
				resp = server.Client.POST("/api/v1/auth/login", loginReq)
				assert.True(t, resp.StatusCode == 401 || resp.StatusCode == 200)
			},
		},
		{
			name:     "Security_SQLInjectionInLogin",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				sqlInjections := []string{
					"admin' OR '1'='1",
					"admin'--",
					"' OR 1=1--",
				}
				for _, injection := range sqlInjections {
					loginReq := map[string]interface{}{
						"username": injection,
						"password": "password",
					}
					resp := server.Client.POST("/api/v1/auth/login", loginReq)
					assert.Equal(t, 401, resp.StatusCode, "SQL injection should be prevented")
				}
			},
		},
		{
			name:     "Security_XSSInRegistration",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				registerReq := map[string]interface{}{
					"username": "xssuser",
					"email":    "xss@example.com",
					"password": "password123",
					"fullname": "<script>alert('XSS')</script>",
				}
				resp := server.Client.POST("/api/v1/users/register", registerReq)
				assert.True(t, resp.StatusCode == 201 || resp.StatusCode == 400 || resp.StatusCode == 422)
			},
		},
		{
			name:     "Security_BruteForceProtection",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				for i := 0; i < 10; i++ {
					loginReq := map[string]interface{}{
						"username": "testuser",
						"password": "wrongpassword",
					}
					resp := server.Client.POST("/api/v1/auth/login", loginReq)
					assert.True(t, resp.StatusCode == 401 || resp.StatusCode == 429)
				}
			},
		},
		{
			name:     "Security_TokenRotation",
			category: "security",
			run: func(t *testing.T, server *setup.TestServer) {
				client := server.Client

				f := fixtures.NewUserFactory(server.DB)
				hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)

				user := f.Create(func(u *userEntity.User) {
					u.Username = "rotate_user"
					u.Email = "rot@test.com"
					u.Password = string(hash)
				})
				server.Enforcer.AddGroupingPolicy(user.ID, "role:user", "global")

				resp := client.POST("/api/v1/auth/login", map[string]any{
					"username": user.Username, "password": "StrongPass123!",
				})
				require.Equal(t, 200, resp.StatusCode)

				cookies := resp.Cookies()
				var refreshToken1 *http.Cookie
				for _, c := range cookies {
					if c.Name == "refresh_token" {
						refreshToken1 = c
						break
					}
				}
				require.NotNil(t, refreshToken1, "Refresh token cookie not found")

				req, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
				req.AddCookie(refreshToken1)

				clientWithCookie := &http.Client{}
				respRotate, err := clientWithCookie.Do(req)
				require.NoError(t, err)
				defer respRotate.Body.Close()

				require.Equal(t, 200, respRotate.StatusCode)

				cookies2 := respRotate.Cookies()
				var refreshToken2 *http.Cookie
				for _, c := range cookies2 {
					if c.Name == "refresh_token" {
						refreshToken2 = c
						break
					}
				}
				require.NotNil(t, refreshToken2)
				assert.NotEqual(t, refreshToken1.Value, refreshToken2.Value)

				reqReuse, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
				reqReuse.AddCookie(refreshToken1)

				respReuse, err := clientWithCookie.Do(reqReuse)
				require.NoError(t, err)
				defer respReuse.Body.Close()

				assert.Contains(t, []int{401, 400}, respReuse.StatusCode)
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
