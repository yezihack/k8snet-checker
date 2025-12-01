package heartbeat

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"
)

// mockInfoCollector 是InfoCollector的mock实现
type mockInfoCollector struct {
	nodeInfo *models.NodeInfo
	err      error
}

func (m *mockInfoCollector) CollectNodeInfo() (*models.NodeInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nodeInfo, nil
}

// mockAPIClient 是APIClient的mock实现
type mockAPIClient struct {
	heartbeatCalls int
	heartbeatErr   error
	lastNodeInfo   *models.NodeInfo
}

func (m *mockAPIClient) SendHeartbeat(info *models.NodeInfo) error {
	m.heartbeatCalls++
	m.lastNodeInfo = info
	return m.heartbeatErr
}

func (m *mockAPIClient) GetHostIPs() ([]string, error) {
	return nil, nil
}

func (m *mockAPIClient) GetPodIPs() ([]string, error) {
	return nil, nil
}

func (m *mockAPIClient) ReportHostTestResults(results []models.ConnectivityResult) error {
	return nil
}

func (m *mockAPIClient) ReportPodTestResults(results []models.ConnectivityResult) error {
	return nil
}

func (m *mockAPIClient) ReportServiceTestResults(result *models.ConnectivityResult) error {
	return nil
}

// TestNewHeartbeatReporter 测试创建HeartbeatReporter
func TestNewHeartbeatReporter(t *testing.T) {
	collector := &mockInfoCollector{
		nodeInfo: &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0.1",
			PodName:   "test-pod",
		},
	}
	apiClient := &mockAPIClient{}

	reporter := NewHeartbeatReporter(collector, apiClient)

	if reporter == nil {
		t.Fatal("NewHeartbeatReporter返回nil")
	}
}

// TestHeartbeatReporter_Start_Success 测试成功启动心跳上报
func TestHeartbeatReporter_Start_Success(t *testing.T) {
	collector := &mockInfoCollector{
		nodeInfo: &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0.1",
			PodName:   "test-pod",
			Timestamp: time.Now(),
		},
	}
	apiClient := &mockAPIClient{}

	reporter := NewHeartbeatReporter(collector, apiClient)

	ctx := context.Background()
	interval := 100 * time.Millisecond

	err := reporter.Start(ctx, interval)
	if err != nil {
		t.Fatalf("Start失败: %v", err)
	}

	// 等待一段时间，确保至少发送了几次心跳
	time.Sleep(350 * time.Millisecond)

	// 停止心跳上报
	err = reporter.Stop()
	if err != nil {
		t.Fatalf("Stop失败: %v", err)
	}

	// 验证心跳被发送了多次（至少2次：立即发送1次 + 定时器触发至少2次）
	if apiClient.heartbeatCalls < 2 {
		t.Errorf("期望至少发送2次心跳，实际发送了%d次", apiClient.heartbeatCalls)
	}

	// 验证最后一次心跳的内容
	if apiClient.lastNodeInfo == nil {
		t.Fatal("没有记录到心跳信息")
	}
	if apiClient.lastNodeInfo.PodName != "test-pod" {
		t.Errorf("期望PodName为'test-pod'，实际为'%s'", apiClient.lastNodeInfo.PodName)
	}
}

// TestHeartbeatReporter_SendHeartbeatError 测试心跳发送失败的情况
func TestHeartbeatReporter_SendHeartbeatError(t *testing.T) {
	collector := &mockInfoCollector{
		nodeInfo: &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0.1",
			PodName:   "test-pod",
			Timestamp: time.Now(),
		},
	}
	apiClient := &mockAPIClient{
		heartbeatErr: errors.New("网络错误"),
	}

	reporter := NewHeartbeatReporter(collector, apiClient)

	ctx := context.Background()
	interval := 100 * time.Millisecond

	err := reporter.Start(ctx, interval)
	if err != nil {
		t.Fatalf("Start失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(250 * time.Millisecond)

	// 停止心跳上报
	err = reporter.Stop()
	if err != nil {
		t.Fatalf("Stop失败: %v", err)
	}

	// 验证即使发送失败，也会继续尝试
	if apiClient.heartbeatCalls < 2 {
		t.Errorf("期望至少尝试发送2次心跳，实际尝试了%d次", apiClient.heartbeatCalls)
	}
}

// TestHeartbeatReporter_CollectInfoError 测试收集信息失败的情况
func TestHeartbeatReporter_CollectInfoError(t *testing.T) {
	collector := &mockInfoCollector{
		err: errors.New("环境变量缺失"),
	}
	apiClient := &mockAPIClient{}

	reporter := NewHeartbeatReporter(collector, apiClient)

	ctx := context.Background()
	interval := 100 * time.Millisecond

	err := reporter.Start(ctx, interval)
	if err != nil {
		t.Fatalf("Start失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(250 * time.Millisecond)

	// 停止心跳上报
	err = reporter.Stop()
	if err != nil {
		t.Fatalf("Stop失败: %v", err)
	}

	// 验证没有发送心跳（因为收集信息失败）
	if apiClient.heartbeatCalls != 0 {
		t.Errorf("期望没有发送心跳，实际发送了%d次", apiClient.heartbeatCalls)
	}
}

// TestHeartbeatReporter_ContextCancellation 测试通过context取消心跳上报
func TestHeartbeatReporter_ContextCancellation(t *testing.T) {
	collector := &mockInfoCollector{
		nodeInfo: &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0.1",
			PodName:   "test-pod",
			Timestamp: time.Now(),
		},
	}
	apiClient := &mockAPIClient{}

	reporter := NewHeartbeatReporter(collector, apiClient)

	ctx, cancel := context.WithCancel(context.Background())
	interval := 100 * time.Millisecond

	err := reporter.Start(ctx, interval)
	if err != nil {
		t.Fatalf("Start失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(150 * time.Millisecond)

	// 取消context
	cancel()

	// 等待goroutine停止
	time.Sleep(100 * time.Millisecond)

	// 记录当前的心跳次数
	callsBefore := apiClient.heartbeatCalls

	// 再等待一段时间，确保不再发送心跳
	time.Sleep(200 * time.Millisecond)

	// 验证心跳已停止
	if apiClient.heartbeatCalls != callsBefore {
		t.Errorf("context取消后仍在发送心跳")
	}
}

// TestGetHeartbeatIntervalFromEnv_Default 测试默认心跳间隔
func TestGetHeartbeatIntervalFromEnv_Default(t *testing.T) {
	// 清除环境变量
	os.Unsetenv("HEARTBEAT_INTERVAL")

	interval := GetHeartbeatIntervalFromEnv()

	expected := 5 * time.Second
	if interval != expected {
		t.Errorf("期望默认间隔为%v，实际为%v", expected, interval)
	}
}

// TestGetHeartbeatIntervalFromEnv_Custom 测试自定义心跳间隔
func TestGetHeartbeatIntervalFromEnv_Custom(t *testing.T) {
	// 设置环境变量
	os.Setenv("HEARTBEAT_INTERVAL", "10")
	defer os.Unsetenv("HEARTBEAT_INTERVAL")

	interval := GetHeartbeatIntervalFromEnv()

	expected := 10 * time.Second
	if interval != expected {
		t.Errorf("期望间隔为%v，实际为%v", expected, interval)
	}
}

// TestGetHeartbeatIntervalFromEnv_Invalid 测试无效的心跳间隔配置
func TestGetHeartbeatIntervalFromEnv_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{"非数字", "abc"},
		{"负数", "-5"},
		{"零", "0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("HEARTBEAT_INTERVAL", tc.value)
			defer os.Unsetenv("HEARTBEAT_INTERVAL")

			interval := GetHeartbeatIntervalFromEnv()

			expected := 5 * time.Second
			if interval != expected {
				t.Errorf("无效配置'%s'应返回默认值%v，实际为%v",
					tc.value, expected, interval)
			}
		})
	}
}

// TestHeartbeatReporter_MultipleStops 测试多次调用Stop
func TestHeartbeatReporter_MultipleStops(t *testing.T) {
	collector := &mockInfoCollector{
		nodeInfo: &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0.1",
			PodName:   "test-pod",
			Timestamp: time.Now(),
		},
	}
	apiClient := &mockAPIClient{}

	reporter := NewHeartbeatReporter(collector, apiClient)

	ctx := context.Background()
	interval := 100 * time.Millisecond

	err := reporter.Start(ctx, interval)
	if err != nil {
		t.Fatalf("Start失败: %v", err)
	}

	// 第一次停止
	err = reporter.Stop()
	if err != nil {
		t.Fatalf("第一次Stop失败: %v", err)
	}

	// 第二次停止（应该不会panic）
	err = reporter.Stop()
	if err != nil {
		t.Fatalf("第二次Stop失败: %v", err)
	}
}
