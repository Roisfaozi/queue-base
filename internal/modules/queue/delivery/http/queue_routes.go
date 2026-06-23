package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterQueueRoutes(router *gin.RouterGroup, controller *QueueController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	queueGroup := router.Group("/queues")
	{
		queueGroup.POST("", apiKeyMiddleware.RequireScopes("queue:manage"), controller.Register)
		queueGroup.GET("", apiKeyMiddleware.RequireScopes("queue:view", "queue:manage"), controller.GetAll)
		queueGroup.GET("/:id", apiKeyMiddleware.RequireScopes("queue:view", "queue:manage"), controller.GetByID)
		queueGroup.POST("/:id/forward", apiKeyMiddleware.RequireScopes("queue:manage"), controller.Forward)
		queueGroup.POST("/:id/transition", apiKeyMiddleware.RequireScopes("queue:manage"), controller.Transition)
	}

	branchGroup := router.Group("/branches")
	{
		branchGroup.GET("/:branch_id/services/:service_id/queue-journeys", apiKeyMiddleware.RequireScopes("queue:view", "queue:manage"), controller.GetJourneysByBranchAndService)
		branchGroup.GET("/:branch_id/counters/:counter_id/queue-journeys", apiKeyMiddleware.RequireScopes("queue:view", "queue:manage"), controller.GetJourneysByBranchAndCounter)
	}
}
