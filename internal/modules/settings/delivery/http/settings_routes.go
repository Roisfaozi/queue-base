package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(router *gin.RouterGroup, controller *SettingsController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	settingsGroup := router.Group("/settings")
	{
		settingsGroup.GET("/effective", apiKeyMiddleware.RequireScopes("settings:view", "settings:manage", "queue-config:view", "queue-config:manage"), controller.EffectiveQueueConfig)
	}

	// QMS domain compliant routes
	qmsConfigGroup := router.Group("/queue-config")
	{
		qmsConfigGroup.GET("/effective", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveQueueConfig)
	}

	branchGroup := router.Group("/branches/:id")
	{
		branchGroup.GET("/effective-config", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveBranchConfig)
		branchGroup.DELETE("/queue-settings/:field", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.ResetBranchQueueSetting)
		branchGroup.GET("/services/:service_id/effective-config", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveBranchServiceConfig)
		branchGroup.DELETE("/services/:branch_service_id/queue-settings/:field", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.ResetBranchServiceQueueSetting)
		branchGroup.GET("/counters/:counter_id/effective-config", apiKeyMiddleware.RequireScopes("queue-config:view", "queue-config:manage"), controller.EffectiveCounterConfig)
		branchGroup.DELETE("/counters/:counter_id/queue-settings/:field", apiKeyMiddleware.RequireScopes("queue-config:manage"), controller.ResetCounterQueueSetting)
	}
}
