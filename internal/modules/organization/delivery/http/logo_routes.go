package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterLogoRoutes registers direct (non-TUS) logo upload endpoints under tenant scope.
func RegisterLogoRoutes(router *gin.RouterGroup, controller *LogoController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	router.POST("/tenant/logo", apiKeyMiddleware.RequireScopes("org:manage"), controller.UploadOrganizationLogo)
	router.POST("/branches/logo", apiKeyMiddleware.RequireScopes("branch:manage"), controller.UploadBranchLogo)
}
