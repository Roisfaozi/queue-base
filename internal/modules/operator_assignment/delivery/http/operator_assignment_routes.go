package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *Controller, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	group := router.Group("/operator-counter-assignments")
	{
		group.POST("", apiKeyMiddleware.RequireScopes("operator_assignment:manage"), controller.Create)
		group.GET("", apiKeyMiddleware.RequireScopes("operator_assignment:manage"), controller.GetAll)
		group.DELETE("/:id", apiKeyMiddleware.RequireScopes("operator_assignment:manage"), controller.Delete)
	}
}
