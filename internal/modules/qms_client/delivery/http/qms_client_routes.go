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
		group.GET("", apiKeyMiddleware.RequireScopes("qms_client:manage"), controller.GetAll)
		group.GET("/:id", apiKeyMiddleware.RequireScopes("qms_client:manage"), controller.GetByID)
		group.PATCH("/:id", apiKeyMiddleware.RequireScopes("qms_client:manage"), controller.Update)
		group.DELETE("/:id", apiKeyMiddleware.RequireScopes("qms_client:manage"), controller.Delete)
	}
}
