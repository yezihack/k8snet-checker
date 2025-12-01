package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"
)

const (
	// 发送心跳到服务器
	SEND_HEART_BEAT_URI = "/api/v1/heartbeat"
	// 获取所有宿主机IP列表
	GET_HOST_IPS_URI = "/api/v1/hosts"
	// 获取所有pod IP列表
	GET_POD_IPS_URI = "/api/v1/pods"
	// 上报测试结果
	REPORT_TEST_RESULT_URI = "/api/v1/test-results/hosts"
	// 上报自定义服务测试结果
	REPORT_POD_TEST_RESULTS_URI = "/api/v1/test-results/pods"
	// 上报自定义服务测试结果
	REPORT_SERVICE_TEST_RESULTS_URI = "/api/v1/test-results/service"
)

// APIClient defines the interface for client-side API interactions with the server
type APIClient interface {
	// SendHeartbeat sends node information to the server as a heartbeat
	SendHeartbeat(info *models.NodeInfo) error

	// GetHostIPs retrieves the list of all host IPs from the server
	GetHostIPs() ([]string, error)

	// GetPodIPs retrieves the list of all pod IPs from the server
	GetPodIPs() ([]string, error)

	// ReportHostTestResults sends host connectivity test results to the server
	ReportHostTestResults(results []models.ConnectivityResult) error

	// ReportPodTestResults sends pod connectivity test results to the server
	ReportPodTestResults(results []models.ConnectivityResult) error

	// ReportServiceTestResults sends custom service test results to the server
	ReportServiceTestResults(result *models.ConnectivityResult) error
}

// apiClientImpl 是APIClient的实现
type apiClientImpl struct {
	serverURL  string
	httpClient *http.Client
	sourceIP   string // 用于上报测试结果时标识源IP
}

// NewAPIClient 创建一个新的APIClient实例
func NewAPIClient(serverURL string, sourceIP string) APIClient {
	return &apiClientImpl{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		sourceIP: sourceIP,
	}
}

// SendHeartbeat 发送心跳到服务器
func (c *apiClientImpl) SendHeartbeat(info *models.NodeInfo) error {
	url := fmt.Sprintf("%s"+SEND_HEART_BEAT_URI, c.serverURL)

	// 序列化请求体
	body, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化心跳数据失败: %w", err)
	}

	// 发送POST请求，带重试逻辑
	err = c.doRequestWithRetry("POST", url, body, nil)
	if err != nil {
		return fmt.Errorf("发送心跳失败: %w", err)
	}

	log.Printf("心跳发送成功: pod=%s, node_ip=%s, pod_ip=%s",
		info.PodName, info.NodeIP, info.PodIP)
	return nil
}

// GetHostIPs 从服务器获取所有宿主机IP列表
func (c *apiClientImpl) GetHostIPs() ([]string, error) {
	url := fmt.Sprintf("%s"+GET_HOST_IPS_URI, c.serverURL)

	var response struct {
		HostIPs []string `json:"host_ips"`
		Count   int      `json:"count"`
	}

	// 发送GET请求，带重试逻辑
	err := c.doRequestWithRetry("GET", url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("获取宿主机IP列表失败: %w", err)
	}

	log.Printf("获取宿主机IP列表成功: count=%d", response.Count)
	return response.HostIPs, nil
}

// GetPodIPs 从服务器获取所有Pod IP列表
func (c *apiClientImpl) GetPodIPs() ([]string, error) {
	url := fmt.Sprintf("%s"+GET_POD_IPS_URI, c.serverURL)

	var response struct {
		PodIPs []string `json:"pod_ips"`
		Count  int      `json:"count"`
	}

	// 发送GET请求，带重试逻辑
	err := c.doRequestWithRetry("GET", url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("获取Pod IP列表失败: %w", err)
	}

	log.Printf("获取Pod IP列表成功: count=%d", response.Count)
	return response.PodIPs, nil
}

// ReportHostTestResults 上报宿主机测试结果到服务器
func (c *apiClientImpl) ReportHostTestResults(results []models.ConnectivityResult) error {
	url := fmt.Sprintf("%s"+REPORT_TEST_RESULT_URI, c.serverURL)

	// 构造请求体
	request := struct {
		SourceIP string                      `json:"source_ip"`
		Results  []models.ConnectivityResult `json:"results"`
	}{
		SourceIP: c.sourceIP,
		Results:  results,
	}

	// 序列化请求体
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化宿主机测试结果失败: %w", err)
	}

	// 发送POST请求，带重试逻辑
	err = c.doRequestWithRetry("POST", url, body, nil)
	if err != nil {
		return fmt.Errorf("上报宿主机测试结果失败: %w", err)
	}

	log.Printf("宿主机测试结果上报成功: source_ip=%s, results_count=%d",
		c.sourceIP, len(results))
	return nil
}

// ReportPodTestResults 上报Pod测试结果到服务器
func (c *apiClientImpl) ReportPodTestResults(results []models.ConnectivityResult) error {
	url := fmt.Sprintf("%s"+REPORT_POD_TEST_RESULTS_URI, c.serverURL)

	// 构造请求体
	request := struct {
		SourceIP string                      `json:"source_ip"`
		Results  []models.ConnectivityResult `json:"results"`
	}{
		SourceIP: c.sourceIP,
		Results:  results,
	}

	// 序列化请求体
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化Pod测试结果失败: %w", err)
	}

	// 发送POST请求，带重试逻辑
	err = c.doRequestWithRetry("POST", url, body, nil)
	if err != nil {
		return fmt.Errorf("上报Pod测试结果失败: %w", err)
	}

	log.Printf("Pod测试结果上报成功: source_ip=%s, results_count=%d",
		c.sourceIP, len(results))
	return nil
}

// ReportServiceTestResults 上报自定义服务测试结果到服务器
func (c *apiClientImpl) ReportServiceTestResults(result *models.ConnectivityResult) error {
	url := fmt.Sprintf("%s"+REPORT_SERVICE_TEST_RESULTS_URI, c.serverURL)

	// 构造请求体
	request := struct {
		SourceIP string                    `json:"source_ip"`
		Result   models.ConnectivityResult `json:"result"`
	}{
		SourceIP: c.sourceIP,
		Result:   *result,
	}

	// 序列化请求体
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化服务测试结果失败: %w", err)
	}

	// 发送POST请求，带重试逻辑
	err = c.doRequestWithRetry("POST", url, body, nil)
	if err != nil {
		return fmt.Errorf("上报服务测试结果失败: %w", err)
	}

	log.Printf("服务测试结果上报成功: source_ip=%s, target=%s",
		c.sourceIP, result.TargetIP)
	return nil
}

// doRequestWithRetry 执行HTTP请求，带指数退避重试逻辑（最多5次）
func (c *apiClientImpl) doRequestWithRetry(method, url string, body []byte, response interface{}) error {
	maxRetries := 5
	baseDelay := 1 * time.Second

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 如果不是第一次尝试，等待一段时间（指数退避）
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1)) // 1s, 2s, 4s, 8s, 16s
			log.Printf("请求失败，%v后重试 (尝试 %d/%d): %v", delay, attempt+1, maxRetries, lastErr)
			time.Sleep(delay)
		}

		// 执行HTTP请求
		err := c.doRequest(method, url, body, response)
		if err == nil {
			// 请求成功
			return nil
		}

		lastErr = err
		log.Printf("请求失败 (尝试 %d/%d): %v", attempt+1, maxRetries, err)
	}

	// 所有重试都失败
	return fmt.Errorf("请求失败，已重试%d次: %w", maxRetries, lastErr)
}

// doRequest 执行单次HTTP请求
func (c *apiClientImpl) doRequest(method, url string, body []byte, response interface{}) error {
	var req *http.Request
	var err error

	// 创建HTTP请求
	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 尝试解析错误响应
		var errorResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			return fmt.Errorf("服务器返回错误 (状态码=%d): %s - %s",
				resp.StatusCode, errorResp.Code, errorResp.Message)
		}
		return fmt.Errorf("服务器返回错误 (状态码=%d): %s",
			resp.StatusCode, string(respBody))
	}

	// 如果需要解析响应体
	if response != nil {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("解析响应体失败: %w", err)
		}
	}

	return nil
}
