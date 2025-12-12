package server

import (
	"log"
	"net/http"

	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/models"
	"github.com/yezihack/k8snet-checker/pkg/result"

	"github.com/gin-gonic/gin"
)

// Handler API处理器
type Handler struct {
	clientManager client.ClientManager
	resultManager result.TestResultManager
}

// NewHandler 创建处理器实例
func NewHandler(clientManager client.ClientManager, resultManager result.TestResultManager) *Handler {
	return &Handler{
		clientManager: clientManager,
		resultManager: resultManager,
	}
}

// HandleHeartbeat 处理心跳上报
// POST /api/v1/heartbeat
func (h *Handler) HandleHeartbeat(c *gin.Context) {
	var nodeInfo models.NodeInfo

	if err := c.ShouldBindJSON(&nodeInfo); err != nil {
		log.Printf("解析心跳请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	if nodeInfo.PodName == "" || nodeInfo.NodeIP == "" || nodeInfo.PodIP == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "缺少必需字段",
			Details: "PodName, NodeIP, PodIP不能为空",
		})
		return
	}

	if err := h.clientManager.HandleHeartbeat(&nodeInfo); err != nil {
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

// HandleHostTestResults 处理宿主机测试结果上报
// POST /api/v1/test-results/hosts
func (h *Handler) HandleHostTestResults(c *gin.Context) {
	var request struct {
		SourceIP string                      `json:"source_ip" binding:"required"`
		Results  []models.ConnectivityResult `json:"results" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("解析宿主机测试结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	if err := h.resultManager.SaveHostTestResults(request.SourceIP, request.Results); err != nil {
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

// HandlePodTestResults 处理Pod测试结果上报
// POST /api/v1/test-results/pods
func (h *Handler) HandlePodTestResults(c *gin.Context) {
	var request struct {
		SourceIP string                      `json:"source_ip" binding:"required"`
		Results  []models.ConnectivityResult `json:"results" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("解析Pod测试结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	if err := h.resultManager.SavePodTestResults(request.SourceIP, request.Results); err != nil {
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

// HandleServiceTestResults 处理自定义服务测试结果上报
// POST /api/v1/test-results/service
func (h *Handler) HandleServiceTestResults(c *gin.Context) {
	var request struct {
		SourceIP string                    `json:"source_ip" binding:"required"`
		Result   models.ConnectivityResult `json:"result" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("解析服务测试结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: err.Error(),
		})
		return
	}

	if err := h.resultManager.SaveServiceTestResult(request.SourceIP, &request.Result); err != nil {
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

// HandleGetHosts 获取所有宿主机IP列表
// GET /api/v1/hosts
func (h *Handler) HandleGetHosts(c *gin.Context) {
	hostIPs, err := h.clientManager.GetAllHostIPs()
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

// HandleGetPods 获取所有Pod IP列表
// GET /api/v1/pods
func (h *Handler) HandleGetPods(c *gin.Context) {
	podIPs, err := h.clientManager.GetAllPodIPs()
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

// HandleGetHostTestResults 获取宿主机互探结果
// GET /api/v1/test-results/hosts
func (h *Handler) HandleGetHostTestResults(c *gin.Context) {
	results, err := h.resultManager.GetHostTestResults()
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

// HandleGetPodTestResults 获取Pod互探结果
// GET /api/v1/test-results/pods
func (h *Handler) HandleGetPodTestResults(c *gin.Context) {
	results, err := h.resultManager.GetPodTestResults()
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

// HandleGetServiceTestResults 获取自定义服务探测结果
// GET /api/v1/test-results/service
func (h *Handler) HandleGetServiceTestResults(c *gin.Context) {
	results, err := h.resultManager.GetServiceTestResults()
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

// HandleGetClientCount 获取活跃客户端数量
// GET /api/v1/clients/count
func (h *Handler) HandleGetClientCount(c *gin.Context) {
	count, err := h.clientManager.GetActiveClientCount()
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

// HandleGetAllResults 获取所有测试结果汇总
// GET /api/v1/results
func (h *Handler) HandleGetAllResults(c *gin.Context) {
	hostIPs, err := h.clientManager.GetAllHostIPs()
	if err != nil {
		log.Printf("获取宿主机IP列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取宿主机IP列表失败",
			Details: err.Error(),
		})
		return
	}

	podIPs, err := h.clientManager.GetAllPodIPs()
	if err != nil {
		log.Printf("获取Pod IP列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取Pod IP列表失败",
			Details: err.Error(),
		})
		return
	}

	hostResults, err := h.resultManager.GetHostTestResults()
	if err != nil {
		log.Printf("获取宿主机测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取宿主机测试结果失败",
			Details: err.Error(),
		})
		return
	}

	podResults, err := h.resultManager.GetPodTestResults()
	if err != nil {
		log.Printf("获取Pod测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取Pod测试结果失败",
			Details: err.Error(),
		})
		return
	}

	serviceResults, err := h.resultManager.GetServiceTestResults()
	if err != nil {
		log.Printf("获取服务测试结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "CACHE_ERROR",
			Message: "获取服务测试结果失败",
			Details: err.Error(),
		})
		return
	}

	activeCount, err := h.clientManager.GetActiveClientCount()
	if err != nil {
		log.Printf("获取活跃客户端数量失败: %v", err)
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

// HandleHealth 健康检查端点
// GET /api/v1/health
func (h *Handler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
