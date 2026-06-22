package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterCounterRoutes(router *gin.RouterGroup, controller *CounterController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	counterGroup := router.Group("/counters")
	{
		counterGroup.POST("", apiKeyMiddleware.RequireScopes("counter:manage"), controller.Create)
		counterGroup.GET("", apiKeyMiddleware.RequireScopes("counter:view", "counter:manage"), controller.GetAll)
		counterGroup.GET("/:id", apiKeyMiddleware.RequireScopes("counter:view", "counter:manage"), controller.GetByID)
		counterGroup.PUT("/:id", apiKeyMiddleware.RequireScopes("counter:manage"), controller.Update)
		counterGroup.DELETE("/:id", apiKeyMiddleware.RequireScopes("counter:manage"), controller.Delete)
	}
}
