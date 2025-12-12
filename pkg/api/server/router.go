package server

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有API路由
func RegisterRoutes(router *gin.Engine, handler *Handler) {
	api := router.Group("/api/v1")

	// 客户端上报接口
	api.POST("/heartbeat", handler.HandleHeartbeat)
	api.POST("/test-results/hosts", handler.HandleHostTestResults)
	api.POST("/test-results/pods", handler.HandlePodTestResults)
	api.POST("/test-results/service", handler.HandleServiceTestResults)

	// 查询接口
	api.GET("/hosts", handler.HandleGetHosts)
	api.GET("/pods", handler.HandleGetPods)
	api.GET("/test-results/hosts", handler.HandleGetHostTestResults)
	api.GET("/test-results/pods", handler.HandleGetPodTestResults)
	api.GET("/test-results/service", handler.HandleGetServiceTestResults)
	api.GET("/clients/count", handler.HandleGetClientCount)
	api.GET("/results", handler.HandleGetAllResults)
	api.GET("/health", handler.HandleHealth)
}
