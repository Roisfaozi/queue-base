package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterQueueRoutes(router *gin.RouterGroup, controller *QueueController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	queueGroup := router.Group("/queues")
	{
		queueGroup.POST("", apiKeyMiddleware.RequireScopes("queue:manage"), controller.Register)
		queueGroup.POST("/:id/forward", apiKeyMiddleware.RequireScopes("queue:manage"), controller.Forward)
		queueGroup.POST("/:id/transition", apiKeyMiddleware.RequireScopes("queue:manage"), controller.Transition)
	}
}
