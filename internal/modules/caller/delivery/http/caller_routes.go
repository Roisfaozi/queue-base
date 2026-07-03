package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterCallerRoutes(router *gin.RouterGroup, controller *CallerController, qmsClientMiddleware *middleware.QMSClientMiddleware) {
	group := router.Group("/caller")
	group.Use(qmsClientMiddleware.Authenticate())
	group.Use(qmsClientMiddleware.RequireClientType("caller"))
	{
		group.POST("/queue-journeys/:journey_id/action", controller.Action)
	}
}
