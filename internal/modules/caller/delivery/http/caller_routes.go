package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterCallerRoutes(router *gin.RouterGroup, controller *CallerController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	group := router.Group("/caller")
	{
		group.POST("/queue-journeys/:journey_id/action", controller.Action)
	}
}
