package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterScannerRoutes(router *gin.RouterGroup, controller *ScannerController) {
	scannerGroup := router.Group("/scanner")
	{
		scannerGroup.POST("/check-in", controller.CheckIn)
	}
}
