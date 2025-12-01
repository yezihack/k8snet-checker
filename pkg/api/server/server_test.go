package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/models"
	"github.com/yezihack/k8snet-checker/pkg/result"

	"github.com/stretchr/testify/assert"
)

// setupTestServer 创建测试用的服务器实例
func setupTestServer() APIServer {
	cacheManager := cache.NewCacheManager()
	clientManager := client.NewClientManager(cacheManager)
	resultManager := result.NewTestResultManager(cacheManager)

	return NewAPIServer(clientManager, resultManager)
}

// TestHealthEndpoint 测试健康检查端点
func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

// TestHeartbeatEndpoint 测试心跳端点
func TestHeartbeatEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	nodeInfo := models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(nodeInfo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

// TestHeartbeatEndpointInvalidData 测试心跳端点的无效数据处理
func TestHeartbeatEndpointInvalidData(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 缺少必需字段
	nodeInfo := models.NodeInfo{
		Namespace: "default",
		// 缺少NodeIP, PodIP, PodName
	}

	body, _ := json.Marshal(nodeInfo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", response.Code)
}

// TestGetHostsEndpoint 测试获取宿主机IP列表端点
func TestGetHostsEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 先发送心跳注册客户端
	nodeInfo := models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(nodeInfo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	// 获取宿主机IP列表
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/hosts", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["host_ips"])
	assert.Equal(t, float64(1), response["count"])
}

// TestGetPodsEndpoint 测试获取Pod IP列表端点
func TestGetPodsEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 先发送心跳注册客户端
	nodeInfo := models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(nodeInfo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	// 获取Pod IP列表
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/pods", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["pod_ips"])
	assert.Equal(t, float64(1), response["count"])
}

// TestHostTestResultsEndpoint 测试宿主机测试结果端点
func TestHostTestResultsEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 上报测试结果
	request := struct {
		SourceIP string                      `json:"source_ip"`
		Results  []models.ConnectivityResult `json:"results"`
	}{
		SourceIP: "192.168.1.1",
		Results: []models.ConnectivityResult{
			{
				SourceIP:   "192.168.1.1",
				TargetIP:   "192.168.1.2",
				PingStatus: "reachable",
				PortStatus: map[int]string{22: "open"},
				Latency:    10 * time.Millisecond,
				Timestamp:  time.Now(),
			},
		},
	}

	body, _ := json.Marshal(request)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-results/hosts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 获取测试结果
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/test-results/hosts", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["results"])
}

// TestPodTestResultsEndpoint 测试Pod测试结果端点
func TestPodTestResultsEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 上报测试结果
	request := struct {
		SourceIP string                      `json:"source_ip"`
		Results  []models.ConnectivityResult `json:"results"`
	}{
		SourceIP: "10.0.0.1",
		Results: []models.ConnectivityResult{
			{
				SourceIP:   "10.0.0.1",
				TargetIP:   "10.0.0.2",
				PingStatus: "reachable",
				PortStatus: map[int]string{6100: "open"},
				Latency:    5 * time.Millisecond,
				Timestamp:  time.Now(),
			},
		},
	}

	body, _ := json.Marshal(request)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-results/pods", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 获取测试结果
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/test-results/pods", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["results"])
}

// TestServiceTestResultsEndpoint 测试服务测试结果端点
func TestServiceTestResultsEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 上报测试结果
	request := struct {
		SourceIP string                    `json:"source_ip"`
		Result   models.ConnectivityResult `json:"result"`
	}{
		SourceIP: "10.0.0.1",
		Result: models.ConnectivityResult{
			SourceIP:   "10.0.0.1",
			TargetIP:   "kubernetes.default.svc.cluster.local",
			PingStatus: "reachable",
			PortStatus: map[int]string{443: "open"},
			Latency:    3 * time.Millisecond,
			Timestamp:  time.Now(),
		},
	}

	body, _ := json.Marshal(request)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-results/service", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 获取测试结果
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/test-results/service", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["results"])
}

// TestGetClientCountEndpoint 测试获取活跃客户端数量端点
func TestGetClientCountEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 先发送心跳注册客户端
	nodeInfo := models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(nodeInfo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	// 获取活跃客户端数量
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/clients/count", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["active_client_count"])
}

// TestGetAllResultsEndpoint 测试获取所有结果汇总端点
func TestGetAllResultsEndpoint(t *testing.T) {
	server := setupTestServer()
	apiServer := server.(*apiServerImpl)

	// 先发送心跳注册客户端
	nodeInfo := models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(nodeInfo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	apiServer.router.ServeHTTP(w, req)

	// 获取所有结果
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/results", nil)
	apiServer.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["host_ips"])
	assert.NotNil(t, response["pod_ips"])
	assert.NotNil(t, response["host_test_results"])
	assert.NotNil(t, response["pod_test_results"])
	assert.NotNil(t, response["service_test_results"])
	assert.NotNil(t, response["active_client_count"])
}
