package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterQueueConfigRoutes(router *gin.RouterGroup, controller *QueueConfigController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	qmsConfigGroup := router.Group("/queue-config")
	{
		qmsConfigGroup.GET("/effective", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveQueueConfig)
		qmsConfigGroup.PATCH("", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.UpdateTenantQueueSetting)
	}

	branchGroup := router.Group("/branches/:id")
	{
		branchGroup.GET("/effective-config", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveBranchConfig)
		branchGroup.PATCH("/queue-config", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.UpdateBranchQueueSetting)
		branchGroup.DELETE("/queue-config/:field", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.ResetBranchQueueSetting)
		branchGroup.GET("/services/:service_id/effective-config", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveBranchServiceConfig)
		branchGroup.PATCH("/services/:branch_service_id/queue-config", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.UpdateBranchServiceQueueSetting)
		branchGroup.DELETE("/services/:branch_service_id/queue-config/:field", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.ResetBranchServiceQueueSetting)
		branchGroup.GET("/counters/:counter_id/effective-config", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveCounterConfig)
		branchGroup.PATCH("/counters/:counter_id/queue-config", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.UpdateCounterQueueSetting)
		branchGroup.DELETE("/counters/:counter_id/queue-config/:field", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.ResetCounterQueueSetting)
	}
}
