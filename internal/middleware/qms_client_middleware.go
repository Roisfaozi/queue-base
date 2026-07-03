package middleware

import (
	"errors"

	qmsClientUsecase "github.com/Roisfaozi/queue-base/internal/modules/qms_client/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	qmsClientIDContextKey   = "qms_client_id"
	qmsClientTypeContextKey = "qms_client_type"
)

type QMSClientMiddleware struct {
	Authenticator qmsClientUsecase.QMSClientAuthenticator
	Log           *logrus.Logger
}

func NewQMSClientMiddleware(auth qmsClientUsecase.QMSClientAuthenticator, log *logrus.Logger) *QMSClientMiddleware {
	return &QMSClientMiddleware{Authenticator: auth, Log: log}
}

func (m *QMSClientMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetHeader("X-Client-ID")
		apiKey := c.GetHeader("X-API-Key")

		if clientID == "" || apiKey == "" {
			c.Next()
			return
		}

		identity, err := m.Authenticator.Authenticate(c.Request.Context(), clientID, apiKey)
		if err != nil {
			m.Log.WithError(err).Warn("QMS Client authentication failed")
			response.Unauthorized(c, err, "unauthorized client")
			c.Abort()
			return
		}

		c.Set(qmsClientIDContextKey, identity.ClientID)
		c.Set(qmsClientTypeContextKey, identity.ClientType)
		c.Set(authMethodContextKey, "qms_client")

		// Force context boundaries
		if identity.TenantID != "" {
			c.Set("organization_id", identity.TenantID)
			ctx := database.SetOrganizationContext(c.Request.Context(), identity.TenantID)
			ctx = database.SetBranchContext(ctx, identity.BranchID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

func (m *QMSClientMiddleware) RequireClientType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		method, ok := c.Get(authMethodContextKey)
		if !ok || method != "qms_client" {
			response.Forbidden(c, errors.New("must use QMS client credential"), "forbidden")
			c.Abort()
			return
		}

		clientType, ok := c.Get(qmsClientTypeContextKey)
		if !ok {
			response.Forbidden(c, errors.New("client type not resolved"), "forbidden")
			c.Abort()
			return
		}

		strType := clientType.(string)
		for _, allowed := range allowedTypes {
			if strType == allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, errors.New("client type not allowed"), "forbidden")
		c.Abort()
	}
}

func GetQMSClientIDFromContext(c *gin.Context) string {
	id, _ := c.Get(qmsClientIDContextKey)
	if idStr, ok := id.(string); ok {
		return idStr
	}
	return ""
}
