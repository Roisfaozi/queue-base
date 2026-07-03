package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterQMSClientRoutes(router *gin.RouterGroup, controller *QMSClientController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	group := router.Group("/qms-clients")
	{
		group.POST("", apiKeyMiddleware.RequireScopes("qms_client:manage"), controller.Create)
		group.POST("/credentials", apiKeyMiddleware.RequireScopes("qms_client:manage"), controller.CreateCredential)
	}
}
