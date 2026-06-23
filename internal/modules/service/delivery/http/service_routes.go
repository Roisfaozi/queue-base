package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterServiceRoutes(router *gin.RouterGroup, controller *ServiceController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	serviceGroup := router.Group("/services")
	{
		serviceGroup.POST("", apiKeyMiddleware.RequireScopes("service:manage"), controller.Create)
		serviceGroup.GET("", apiKeyMiddleware.RequireScopes("service:view", "service:manage"), controller.GetAll)
		serviceGroup.GET("/:id", apiKeyMiddleware.RequireScopes("service:view", "service:manage"), controller.GetByID)
		serviceGroup.PUT("/:id", apiKeyMiddleware.RequireScopes("service:manage"), controller.Update)
		serviceGroup.DELETE("/:id", apiKeyMiddleware.RequireScopes("service:manage"), controller.Delete)
	}
}
