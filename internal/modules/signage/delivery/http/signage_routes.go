package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterSignageRoutes(router *gin.RouterGroup, controller *SignageController, qmsClientMiddleware *middleware.QMSClientMiddleware) {
	group := router.Group("/signage")
	group.Use(qmsClientMiddleware.Authenticate())
	group.Use(qmsClientMiddleware.RequireClientType("signage"))
	{
		group.GET("/me", controller.Me)
		group.GET("/current-calls", controller.CurrentCalls)
		group.GET("/queues", controller.Queues)
	}
}
