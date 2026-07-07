package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
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

	branchCounterGroup := router.Group("/branches/:id/counters")
	{
		branchCounterGroup.POST("", apiKeyMiddleware.RequireScopes("counter:manage"), controller.CreateUnderBranch)
		branchCounterGroup.GET("", apiKeyMiddleware.RequireScopes("counter:view", "counter:manage"), controller.GetAllUnderBranch)
		branchCounterGroup.GET("/:counter_id", apiKeyMiddleware.RequireScopes("counter:view", "counter:manage"), controller.GetByIDUnderBranch)
		branchCounterGroup.PUT("/:counter_id", apiKeyMiddleware.RequireScopes("counter:manage"), controller.UpdateUnderBranch)
		branchCounterGroup.PATCH("/:counter_id", apiKeyMiddleware.RequireScopes("counter:manage"), controller.UpdateUnderBranch)
		branchCounterGroup.DELETE("/:counter_id", apiKeyMiddleware.RequireScopes("counter:manage"), controller.DeleteUnderBranch)
	}
}
