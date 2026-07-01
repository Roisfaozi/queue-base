//go:build integration
// +build integration

package scenarios

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestScenario_ExceptionHandling_PanicRecovery(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "PanicRecovery",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)

				gin.SetMode(gin.TestMode)
				router := gin.New()
				router.Use(middleware.RecoveryMiddleware(env.Logger))

				router.GET("/panic-trigger", func(c *gin.Context) {
					panic("intentional crash for testing")
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/panic-trigger", nil)

				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected 500 Internal Server Error after panic")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
