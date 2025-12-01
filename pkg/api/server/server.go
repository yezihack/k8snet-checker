package server

import (
	"log"
	"net/http"
	"os"

	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/models"
	"github.com/yezihack/k8snet-checker/pkg/result"

	"github.com/gin-gonic/gin"
)

// APIServer 定义HTTP API服务器接口
type APIServer interface {
	Start(port string) error
	Stop() error
}

// apiServerImpl 是APIServer的实现
type apiServerImpl struct {
	router        *gin.Engine
	clientManager client.ClientManager
	resultManager result.TestResultManager
}

// NewAPIServer 创建一个新的APIServer实例
func NewAPIServer(clientManager client.ClientManager, resultManager result.TestResultManager) APIServer {
	// 根据LOG_LEVEL设置Gin模式
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	server := &apiServerImpl{
		router:        router,
		clientManager: clientManager,
		resultManager: resultManager,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册所有API路由
func (s *apiServerImpl) registerRoutes() {
	api := s.router.Group("/api/v1")

	// 客户端上报接口
	api.POST("/heartbeat", s.handleHeartbeat)
	api.POST("/test-results/hosts", s.handleHostTestResults)
	api.POST("/test-results/pods", s.handlePodTestResults)
	api.POST("/test-results/service", s.handleServiceTestResults)

	// 查询接口
	api.GET("/hosts", s.handleGetHosts)
	api.GET("/pods", s.handleGetPods)
	api.GET("/test-results/hosts", s.handleGetHostTestResults)
	api.GET("/test-results/pods", s.handleGetPodTestResults)
	api.GET("/test-results/service", s.handleGetServiceTestResults)
	api.GET("/clients/count", s.handleGetClientCount)
	api.GET("/results", s.handleGetAllResults)
	api.GET("/health", s.handleHealth)
}

// handleHeartbeat 处理心跳上报
// POST /api/v1/heartbeat
func (s *apiServerImpl) handleHeartbeat(c *gin.Context) {
	var nodeInfo models.NodeInfo

	// 解析请求体
	if err := c.ShouldBindJSON(&nodeInfo); err != nil {
		log.Printf("解析心跳请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	// 验证必需字段
	if nodeInfo.PodName == "" || nodeInfo.NodeIP == "" || nodeInfo.PodIP == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "缺少必需字段",
			Details: "PodName, NodeIP, PodIP不能为空",
		})
		return
	}

	// 处理心跳
	if err := s.clientManager.HandleHeartbeat(&nodeInfo); err != nil {
		log.Printf("处理心跳失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "处理心跳失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "心跳接收成功",
	})
}

// handleHostTestResults 处理宿主机测试结果上报
// POST /api/v1/test-results/hosts
func (s *apiServerImpl) handleHostTestResults(c *gin.Context) {
	var request struct {
		SourceIP string                      `json:"source_ip" binding:"required"`
		Results  []models.ConnectivityResult `json:"results" binding:"required"`
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("解析宿主机测试结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	// 保存测试结果
	if err := s.resultManager.SaveHostTestResults(request.SourceIP, request.Results); err != nil {
		log.Printf("保存宿主机测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "保存测试结果失败",
			Details: err.Error(),
		})
		return
	}

	log.Printf("宿主机测试结果保存成功: source_ip=%s, results_count=%d",
		request.SourceIP, len(request.Results))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "测试结果保存成功",
	})
}

// handlePodTestResults 处理Pod测试结果上报
// POST /api/v1/test-results/pods
func (s *apiServerImpl) handlePodTestResults(c *gin.Context) {
	var request struct {
		SourceIP string                      `json:"source_ip" binding:"required"`
		Results  []models.ConnectivityResult `json:"results" binding:"required"`
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("解析Pod测试结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	// 保存测试结果
	if err := s.resultManager.SavePodTestResults(request.SourceIP, request.Results); err != nil {
		log.Printf("保存Pod测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "保存测试结果失败",
			Details: err.Error(),
		})
		return
	}

	log.Printf("Pod测试结果保存成功: source_ip=%s, results_count=%d",
		request.SourceIP, len(request.Results))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "测试结果保存成功",
	})
}

// handleServiceTestResults 处理自定义服务测试结果上报
// POST /api/v1/test-results/service
func (s *apiServerImpl) handleServiceTestResults(c *gin.Context) {
	var request struct {
		SourceIP string                    `json:"source_ip" binding:"required"`
		Result   models.ConnectivityResult `json:"result" binding:"required"`
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("解析服务测试结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	// 保存测试结果
	if err := s.resultManager.SaveServiceTestResult(request.SourceIP, &request.Result); err != nil {
		log.Printf("保存服务测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "保存测试结果失败",
			Details: err.Error(),
		})
		return
	}

	log.Printf("服务测试结果保存成功: source_ip=%s, target=%s",
		request.SourceIP, request.Result.TargetIP)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "测试结果保存成功",
	})
}

// handleGetHosts 获取所有宿主机IP列表
// GET /api/v1/hosts
func (s *apiServerImpl) handleGetHosts(c *gin.Context) {
	hostIPs, err := s.clientManager.GetAllHostIPs()
	if err != nil {
		log.Printf("获取宿主机IP列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取宿主机IP列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"host_ips": hostIPs,
		"count":    len(hostIPs),
	})
}

// handleGetPods 获取所有Pod IP列表
// GET /api/v1/pods
func (s *apiServerImpl) handleGetPods(c *gin.Context) {
	podIPs, err := s.clientManager.GetAllPodIPs()
	if err != nil {
		log.Printf("获取Pod IP列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取Pod IP列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pod_ips": podIPs,
		"count":   len(podIPs),
	})
}

// handleGetHostTestResults 获取宿主机互探结果
// GET /api/v1/test-results/hosts
func (s *apiServerImpl) handleGetHostTestResults(c *gin.Context) {
	results, err := s.resultManager.GetHostTestResults()
	if err != nil {
		log.Printf("获取宿主机测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取测试结果失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// handleGetPodTestResults 获取Pod互探结果
// GET /api/v1/test-results/pods
func (s *apiServerImpl) handleGetPodTestResults(c *gin.Context) {
	results, err := s.resultManager.GetPodTestResults()
	if err != nil {
		log.Printf("获取Pod测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取测试结果失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// handleGetServiceTestResults 获取自定义服务探测结果
// GET /api/v1/test-results/service
func (s *apiServerImpl) handleGetServiceTestResults(c *gin.Context) {
	results, err := s.resultManager.GetServiceTestResults()
	if err != nil {
		log.Printf("获取服务测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取测试结果失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// handleGetClientCount 获取活跃客户端数量
// GET /api/v1/clients/count
func (s *apiServerImpl) handleGetClientCount(c *gin.Context) {
	count, err := s.clientManager.GetActiveClientCount()
	if err != nil {
		log.Printf("获取活跃客户端数量失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取活跃客户端数量失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_client_count": count,
	})
}

// handleGetAllResults 获取所有测试结果汇总
// GET /api/v1/results
func (s *apiServerImpl) handleGetAllResults(c *gin.Context) {
	// 获取宿主机IP列表
	hostIPs, err := s.clientManager.GetAllHostIPs()
	if err != nil {
		log.Printf("获取宿主机IP列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取宿主机IP列表失败",
			Details: err.Error(),
		})
		return
	}

	// 获取Pod IP列表
	podIPs, err := s.clientManager.GetAllPodIPs()
	if err != nil {
		log.Printf("获取Pod IP列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取Pod IP列表失败",
			Details: err.Error(),
		})
		return
	}

	// 获取宿主机测试结果
	hostResults, err := s.resultManager.GetHostTestResults()
	if err != nil {
		log.Printf("获取宿主机测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取宿主机测试结果失败",
			Details: err.Error(),
		})
		return
	}

	// 获取Pod测试结果
	podResults, err := s.resultManager.GetPodTestResults()
	if err != nil {
		log.Printf("获取Pod测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取Pod测试结果失败",
			Details: err.Error(),
		})
		return
	}

	// 获取服务测试结果
	serviceResults, err := s.resultManager.GetServiceTestResults()
	if err != nil {
		log.Printf("获取服务测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取服务测试结果失败",
			Details: err.Error(),
		})
		return
	}

	// 获取活跃客户端数量
	activeCount, err := s.clientManager.GetActiveClientCount()
	if err != nil {
		log.Printf("获取活跃客户端数量失败: %v", err)
		// 不返回错误，继续返回其他数据
		activeCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"active_client_count":  activeCount,
		"host_ips":             hostIPs,
		"pod_ips":              podIPs,
		"host_test_results":    hostResults,
		"pod_test_results":     podResults,
		"service_test_results": serviceResults,
	})
}

// handleHealth 健康检查端点
// GET /api/v1/health
func (s *apiServerImpl) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Start 启动HTTP服务器
func (s *apiServerImpl) Start(port string) error {
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP服务器启动在端口: %s", port)
	return s.router.Run(":" + port)
}

// Stop 停止HTTP服务器
func (s *apiServerImpl) Stop() error {
	// Gin框架没有内置的优雅关闭方法
	// 这里返回nil，实际的优雅关闭会在main.go中通过context处理
	return nil
}
