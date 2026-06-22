package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(router *gin.RouterGroup, controller *SettingsController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	settingsGroup := router.Group("/settings")
	{
		settingsGroup.POST("", apiKeyMiddleware.RequireScopes("settings:manage"), controller.Create)
		settingsGroup.GET("/resolve", apiKeyMiddleware.RequireScopes("settings:view", "settings:manage"), controller.Resolve)
		settingsGroup.GET("/:id", apiKeyMiddleware.RequireScopes("settings:view", "settings:manage"), controller.GetByID)
		settingsGroup.PUT("/:id", apiKeyMiddleware.RequireScopes("settings:manage"), controller.Update)
		settingsGroup.DELETE("/:id", apiKeyMiddleware.RequireScopes("settings:manage"), controller.Delete)
	}
}
