package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/Roisfaozi/queue-base/internal/modules/access"
	"github.com/Roisfaozi/queue-base/internal/modules/api_key"
	"github.com/Roisfaozi/queue-base/internal/modules/audit"
	"github.com/Roisfaozi/queue-base/internal/modules/auth"
	"github.com/Roisfaozi/queue-base/internal/modules/counter"
	"github.com/Roisfaozi/queue-base/internal/modules/organization"
	"github.com/Roisfaozi/queue-base/internal/modules/permission"
	"github.com/Roisfaozi/queue-base/internal/modules/project"
	"github.com/Roisfaozi/queue-base/internal/modules/queue"
	"github.com/Roisfaozi/queue-base/internal/modules/role"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner"
	"github.com/Roisfaozi/queue-base/internal/modules/service"
	"github.com/Roisfaozi/queue-base/internal/modules/settings"
	"github.com/Roisfaozi/queue-base/internal/modules/stats"
	"github.com/Roisfaozi/queue-base/internal/modules/user"
	"github.com/Roisfaozi/queue-base/internal/modules/webhook"
	"github.com/Roisfaozi/queue-base/pkg/sse"
	"github.com/Roisfaozi/queue-base/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tus/tusd/v2/pkg/handler"
	"gorm.io/gorm"
)

func createTestRouter(cfg RouterConfig) *gin.Engine {
	return SetupRouter(
		cfg,
		&auth.AuthModule{},
		&user.UserModule{},
		&permission.PermissionModule{},
		&access.AccessModule{},
		&role.RoleModule{},
		&organization.OrganizationModule{},
		&organization.BranchModule{},
		&audit.AuditModule{},
		&stats.StatsModule{},
		&project.ProjectModule{},
		&service.ServiceModule{},
		&counter.CounterModule{},
		&settings.SettingsModule{},
		&queue.QueueModule{},
		&scanner.ScannerModule{},
		&api_key.ApiKeyModule{},
		&webhook.WebhookModule{},
		&middleware.AuthMiddleware{},
		&middleware.APIKeyMiddleware{},
		func(c *gin.Context) { c.Next() },
		&middleware.TenantMiddleware{},
		&ws.WebSocketController{},
		sse.NewManager(),
		&gorm.DB{},
		&redis.Client{},
		&handler.Handler{},
		logrus.New(),
	)
}

// ============================================================================
// Table Driven Tests
// ============================================================================

func TestTrustedProxies(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "TrustedProxies_Tests",
			category: "edge",
			run: func(t *testing.T) {
				gin.SetMode(gin.TestMode)

				t.Run("Should not trust X-Forwarded-For by default", func(t *testing.T) {
					cfg := RouterConfig{
						AllowedOrigins: []string{"*"},
					}

					router := createTestRouter(cfg)
					router.GET("/test-ip", func(c *gin.Context) {
						c.String(200, c.ClientIP())
					})

					req, _ := http.NewRequest("GET", "/test-ip", nil)

					req.Header.Set("X-Forwarded-For", "1.2.3.4")
					req.RemoteAddr = "10.0.0.1:12345"

					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					assert.Equal(t, "10.0.0.1", w.Body.String())
				})

				t.Run("Should trust configured proxy", func(t *testing.T) {
					cfg := RouterConfig{
						AllowedOrigins: []string{"*"},
						TrustedProxies: []string{"10.0.0.1"},
					}

					router := createTestRouter(cfg)
					router.GET("/test-ip-trusted", func(c *gin.Context) {
						c.String(200, c.ClientIP())
					})

					req, _ := http.NewRequest("GET", "/test-ip-trusted", nil)

					req.Header.Set("X-Forwarded-For", "1.2.3.4")
					req.RemoteAddr = "10.0.0.1:12345"

					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					assert.Equal(t, "1.2.3.4", w.Body.String())
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
