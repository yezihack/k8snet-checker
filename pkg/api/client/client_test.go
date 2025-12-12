package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"

	"github.com/stretchr/testify/assert"
)

// TestSendHeartbeat 测试发送心跳功能
func TestSendHeartbeat(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/heartbeat", r.URL.Path)

		// 验证请求头
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 解析请求体
		var nodeInfo models.NodeInfo
		err := json.NewDecoder(r.Body).Decode(&nodeInfo)
		assert.NoError(t, err)

		// 验证请求数据
		assert.Equal(t, "test-namespace", nodeInfo.Namespace)
		assert.Equal(t, "192.168.1.1", nodeInfo.NodeIP)
		assert.Equal(t, "10.0.0.1", nodeInfo.PodIP)
		assert.Equal(t, "test-pod", nodeInfo.PodName)

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "心跳接收成功",
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 发送心跳
	nodeInfo := &models.NodeInfo{
		Namespace: "test-namespace",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	err := client.SendHeartbeat(nodeInfo)
	assert.NoError(t, err)
}

// TestGetHostIPs 测试获取宿主机IP列表
func TestGetHostIPs(t *testing.T) {
	expectedIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/hosts", r.URL.Path)

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"host_ips": expectedIPs,
			"count":    len(expectedIPs),
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 获取宿主机IP列表
	hostIPs, err := client.GetHostIPs()
	assert.NoError(t, err)
	assert.Equal(t, expectedIPs, hostIPs)
}

// TestGetPodIPs 测试获取Pod IP列表
func TestGetPodIPs(t *testing.T) {
	expectedIPs := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/pods", r.URL.Path)

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pod_ips": expectedIPs,
			"count":   len(expectedIPs),
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 获取Pod IP列表
	podIPs, err := client.GetPodIPs()
	assert.NoError(t, err)
	assert.Equal(t, expectedIPs, podIPs)
}

// TestReportHostTestResults 测试上报宿主机测试结果
func TestReportHostTestResults(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/test-results/hosts", r.URL.Path)

		// 解析请求体
		var request struct {
			SourceIP string                      `json:"source_ip"`
			Results  []models.ConnectivityResult `json:"results"`
		}
		err := json.NewDecoder(r.Body).Decode(&request)
		assert.NoError(t, err)

		// 验证请求数据
		assert.Equal(t, "10.0.0.1", request.SourceIP)
		assert.Equal(t, 2, len(request.Results))

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "测试结果保存成功",
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 上报测试结果
	results := []models.ConnectivityResult{
		{
			SourceIP:   "10.0.0.1",
			TargetIP:   "192.168.1.1",
			PingStatus: "reachable",
			PortStatus: map[int]string{22: "open"},
			Latency:    10 * time.Millisecond,
			Timestamp:  time.Now(),
		},
		{
			SourceIP:   "10.0.0.1",
			TargetIP:   "192.168.1.2",
			PingStatus: "unreachable",
			PortStatus: map[int]string{22: "closed"},
			Latency:    0,
			Timestamp:  time.Now(),
		},
	}

	err := client.ReportHostTestResults(results)
	assert.NoError(t, err)
}

// TestReportPodTestResults 测试上报Pod测试结果
func TestReportPodTestResults(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/test-results/pods", r.URL.Path)

		// 解析请求体
		var request struct {
			SourceIP string                      `json:"source_ip"`
			Results  []models.ConnectivityResult `json:"results"`
		}
		err := json.NewDecoder(r.Body).Decode(&request)
		assert.NoError(t, err)

		// 验证请求数据
		assert.Equal(t, "10.0.0.1", request.SourceIP)
		assert.Equal(t, 1, len(request.Results))

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "测试结果保存成功",
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 上报测试结果
	results := []models.ConnectivityResult{
		{
			SourceIP:   "10.0.0.1",
			TargetIP:   "10.0.0.2",
			PingStatus: "reachable",
			PortStatus: map[int]string{6100: "open"},
			Latency:    5 * time.Millisecond,
			Timestamp:  time.Now(),
		},
	}

	err := client.ReportPodTestResults(results)
	assert.NoError(t, err)
}

// TestReportServiceTestResults 测试上报服务测试结果
func TestReportServiceTestResults(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/test-results/service", r.URL.Path)

		// 解析请求体
		var request struct {
			SourceIP string                    `json:"source_ip"`
			Result   models.ConnectivityResult `json:"result"`
		}
		err := json.NewDecoder(r.Body).Decode(&request)
		assert.NoError(t, err)

		// 验证请求数据
		assert.Equal(t, "10.0.0.1", request.SourceIP)
		assert.Equal(t, "my-service.default.svc.cluster.local", request.Result.TargetIP)

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "测试结果保存成功",
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 上报测试结果
	result := &models.ConnectivityResult{
		SourceIP:   "10.0.0.1",
		TargetIP:   "my-service.default.svc.cluster.local",
		PingStatus: "reachable",
		PortStatus: map[int]string{80: "open"},
		Latency:    15 * time.Millisecond,
		Timestamp:  time.Now(),
	}

	err := client.ReportServiceTestResults(result)
	assert.NoError(t, err)
}

// TestRetryLogic 测试重试逻辑
func TestRetryLogic(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器，前2次失败，第3次成功
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount < 3 {
			// 前2次返回错误
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "临时错误",
			})
			return
		}

		// 第3次返回成功
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"host_ips": []string{"192.168.1.1"},
			"count":    1,
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 获取宿主机IP列表（应该在第3次尝试时成功）
	hostIPs, err := client.GetHostIPs()
	assert.NoError(t, err)
	assert.Equal(t, []string{"192.168.1.1"}, hostIPs)
	assert.Equal(t, 3, attemptCount)
}

// TestMaxRetriesExceeded 测试超过最大重试次数
func TestMaxRetriesExceeded(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器，始终返回错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "持续错误",
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 获取宿主机IP列表（应该失败）
	_, err := client.GetHostIPs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已重试5次")
	assert.Equal(t, 5, attemptCount)
}

// TestErrorResponse 测试错误响应处理
func TestErrorResponse(t *testing.T) {
	// 创建测试服务器，返回400错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "无效的请求数据",
			Details: "缺少必需字段",
		})
	}))
	defer server.Close()

	// 创建客户端
	client := NewAPIClient(server.URL, "10.0.0.1")

	// 发送心跳（应该失败）
	nodeInfo := &models.NodeInfo{
		Namespace: "test-namespace",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	err := client.SendHeartbeat(nodeInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_REQUEST")
	assert.Contains(t, err.Error(), "无效的请求数据")
}
