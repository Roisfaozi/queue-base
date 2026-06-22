package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterBranchRoutes(router *gin.RouterGroup, controller *BranchController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	branchGroup := router.Group("/branches")
	{
		branchGroup.POST("", apiKeyMiddleware.RequireScopes("branch:manage"), controller.Create)
		branchGroup.GET("", apiKeyMiddleware.RequireScopes("branch:view", "branch:manage"), controller.GetAll)
		branchGroup.GET("/:id", apiKeyMiddleware.RequireScopes("branch:view", "branch:manage"), controller.GetByID)
		branchGroup.PUT("/:id", apiKeyMiddleware.RequireScopes("branch:manage"), controller.Update)
		branchGroup.DELETE("/:id", apiKeyMiddleware.RequireScopes("branch:manage"), controller.Delete)
	}
}
