package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(router *gin.RouterGroup, controller *SettingsController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	settingsGroup := router.Group("/settings")
	{
		settingsGroup.GET("/effective", apiKeyMiddleware.RequireScopes("settings:view", "settings:manage"), controller.EffectiveQueueConfig)
	}
}
