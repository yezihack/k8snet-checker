package report

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientManager 是ClientManager的mock实现
type MockClientManager struct {
	mock.Mock
}

func (m *MockClientManager) HandleHeartbeat(info *models.NodeInfo) error {
	args := m.Called(info)
	return args.Error(0)
}

func (m *MockClientManager) GetActiveClientCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockClientManager) GetAllHostIPs() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockClientManager) GetAllPodIPs() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// MockTestResultManager 是TestResultManager的mock实现
type MockTestResultManager struct {
	mock.Mock
}

func (m *MockTestResultManager) SaveHostTestResults(sourceIP string, results []models.ConnectivityResult) error {
	args := m.Called(sourceIP, results)
	return args.Error(0)
}

func (m *MockTestResultManager) SavePodTestResults(sourceIP string, results []models.ConnectivityResult) error {
	args := m.Called(sourceIP, results)
	return args.Error(0)
}

func (m *MockTestResultManager) SaveServiceTestResult(sourceIP string, result *models.ConnectivityResult) error {
	args := m.Called(sourceIP, result)
	return args.Error(0)
}

func (m *MockTestResultManager) GetHostTestResults() (models.HostTestResults, error) {
	args := m.Called()
	return args.Get(0).(models.HostTestResults), args.Error(1)
}

func (m *MockTestResultManager) GetPodTestResults() (models.PodTestResults, error) {
	args := m.Called()
	return args.Get(0).(models.PodTestResults), args.Error(1)
}

func (m *MockTestResultManager) GetServiceTestResults() (models.ServiceTestResults, error) {
	args := m.Called()
	return args.Get(0).(models.ServiceTestResults), args.Error(1)
}

// TestNewReportGenerator 测试创建ReportGenerator
func TestNewReportGenerator(t *testing.T) {
	mockClientManager := new(MockClientManager)
	mockResultManager := new(MockTestResultManager)

	generator := NewReportGenerator(mockClientManager, mockResultManager)

	assert.NotNil(t, generator)
}

// TestGenerateReport 测试生成报告
func TestGenerateReport(t *testing.T) {
	mockClientManager := new(MockClientManager)
	mockResultManager := new(MockTestResultManager)

	// 设置mock返回值
	mockClientManager.On("GetActiveClientCount").Return(3, nil)
	mockClientManager.On("GetAllHostIPs").Return([]string{"192.168.1.1", "192.168.1.2"}, nil)
	mockClientManager.On("GetAllPodIPs").Return([]string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}, nil)

	hostTestResults := models.HostTestResults{
		"192.168.1.1": {
			"192.168.1.2": models.TestStatus{Ping: "reachable", PortStatus: "open"},
		},
	}
	mockResultManager.On("GetHostTestResults").Return(hostTestResults, nil)

	podTestResults := models.PodTestResults{
		"10.0.0.1": {
			"10.0.0.2": models.TestStatus{Ping: "reachable", PortStatus: "open"},
			"10.0.0.3": models.TestStatus{Ping: "unreachable", PortStatus: "closed"},
		},
	}
	mockResultManager.On("GetPodTestResults").Return(podTestResults, nil)

	serviceTestResults := models.ServiceTestResults{
		"10.0.0.1": &models.ConnectivityResult{
			SourceIP:   "10.0.0.1",
			TargetIP:   "my-service",
			PingStatus: "reachable",
		},
	}
	mockResultManager.On("GetServiceTestResults").Return(serviceTestResults, nil)

	generator := NewReportGenerator(mockClientManager, mockResultManager)

	// 生成报告
	report, err := generator.GenerateReport()

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 3, report.ActiveClientCount)
	assert.Equal(t, 2, len(report.HostIPs))
	assert.Equal(t, 3, len(report.PodIPs))
	assert.Equal(t, 1, report.HostTestSummary.TotalTests)
	assert.Equal(t, 1, report.HostTestSummary.SuccessfulTests)
	assert.Equal(t, 2, report.PodTestSummary.TotalTests)
	assert.Equal(t, 1, report.PodTestSummary.SuccessfulTests)
	assert.Equal(t, 1, report.PodTestSummary.FailedTests)

	mockClientManager.AssertExpectations(t)
	mockResultManager.AssertExpectations(t)
}

// TestCalculateTestSummary 测试计算测试统计
func TestCalculateTestSummary(t *testing.T) {
	mockClientManager := new(MockClientManager)
	mockResultManager := new(MockTestResultManager)

	generator := NewReportGenerator(mockClientManager, mockResultManager).(*reportGeneratorImpl)

	// 测试用例1: 全部成功
	results1 := map[string]map[string]models.TestStatus{
		"source1": {
			"target1": models.TestStatus{Ping: "reachable", PortStatus: "open"},
			"target2": models.TestStatus{Ping: "reachable", PortStatus: "open"},
		},
	}
	summary1 := generator.calculateTestSummary(results1)
	assert.Equal(t, 2, summary1.TotalTests)
	assert.Equal(t, 2, summary1.SuccessfulTests)
	assert.Equal(t, 0, summary1.FailedTests)
	assert.Equal(t, 100.0, summary1.SuccessRate)

	// 测试用例2: 部分失败
	results2 := map[string]map[string]models.TestStatus{
		"source1": {
			"target1": models.TestStatus{Ping: "reachable", PortStatus: "open"},
			"target2": models.TestStatus{Ping: "unreachable", PortStatus: "closed"},
		},
	}
	summary2 := generator.calculateTestSummary(results2)
	assert.Equal(t, 2, summary2.TotalTests)
	assert.Equal(t, 1, summary2.SuccessfulTests)
	assert.Equal(t, 1, summary2.FailedTests)
	assert.Equal(t, 50.0, summary2.SuccessRate)

	// 测试用例3: 空结果
	results3 := map[string]map[string]models.TestStatus{}
	summary3 := generator.calculateTestSummary(results3)
	assert.Equal(t, 0, summary3.TotalTests)
	assert.Equal(t, 0, summary3.SuccessfulTests)
	assert.Equal(t, 0, summary3.FailedTests)
	assert.Equal(t, 0.0, summary3.SuccessRate)
}

// TestCalculateServiceTestSummary 测试计算服务测试统计
func TestCalculateServiceTestSummary(t *testing.T) {
	mockClientManager := new(MockClientManager)
	mockResultManager := new(MockTestResultManager)

	generator := NewReportGenerator(mockClientManager, mockResultManager).(*reportGeneratorImpl)

	// 测试用例1: 全部成功
	results1 := models.ServiceTestResults{
		"10.0.0.1": &models.ConnectivityResult{
			SourceIP:   "10.0.0.1",
			TargetIP:   "my-service",
			PingStatus: "reachable",
		},
		"10.0.0.2": &models.ConnectivityResult{
			SourceIP:   "10.0.0.2",
			TargetIP:   "my-service",
			PingStatus: "reachable",
		},
	}
	summary1 := generator.calculateServiceTestSummary(results1)
	assert.Equal(t, 2, summary1.TotalTests)
	assert.Equal(t, 2, summary1.SuccessfulTests)
	assert.Equal(t, 0, summary1.FailedTests)
	assert.Equal(t, 100.0, summary1.SuccessRate)

	// 测试用例2: 部分失败
	results2 := models.ServiceTestResults{
		"10.0.0.1": &models.ConnectivityResult{
			SourceIP:   "10.0.0.1",
			TargetIP:   "my-service",
			PingStatus: "reachable",
		},
		"10.0.0.2": &models.ConnectivityResult{
			SourceIP:   "10.0.0.2",
			TargetIP:   "my-service",
			PingStatus: "unreachable",
		},
	}
	summary2 := generator.calculateServiceTestSummary(results2)
	assert.Equal(t, 2, summary2.TotalTests)
	assert.Equal(t, 1, summary2.SuccessfulTests)
	assert.Equal(t, 1, summary2.FailedTests)
	assert.Equal(t, 50.0, summary2.SuccessRate)
}

// TestStartAndStop 测试启动和停止报告生成器
func TestStartAndStop(t *testing.T) {
	mockClientManager := new(MockClientManager)
	mockResultManager := new(MockTestResultManager)

	// 设置mock返回值
	mockClientManager.On("GetActiveClientCount").Return(0, nil)
	mockClientManager.On("GetAllHostIPs").Return([]string{}, nil)
	mockClientManager.On("GetAllPodIPs").Return([]string{}, nil)
	mockResultManager.On("GetHostTestResults").Return(models.HostTestResults{}, nil)
	mockResultManager.On("GetPodTestResults").Return(models.PodTestResults{}, nil)
	mockResultManager.On("GetServiceTestResults").Return(models.ServiceTestResults{}, nil)

	generator := NewReportGenerator(mockClientManager, mockResultManager)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动报告生成器（使用很短的间隔进行测试）
	err := generator.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// 等待一段时间让goroutine运行
	time.Sleep(250 * time.Millisecond)

	// 停止报告生成器
	err = generator.Stop()
	assert.NoError(t, err)

	// 等待goroutine完全停止
	time.Sleep(100 * time.Millisecond)
}

// TestGetReportIntervalFromEnv 测试从环境变量获取报告间隔
func TestGetReportIntervalFromEnv(t *testing.T) {
	// 测试用例1: 未设置环境变量，使用默认值
	os.Unsetenv("REPORT_INTERVAL")
	interval1 := GetReportIntervalFromEnv()
	assert.Equal(t, 300*time.Second, interval1)

	// 测试用例2: 设置有效值
	os.Setenv("REPORT_INTERVAL", "60")
	interval2 := GetReportIntervalFromEnv()
	assert.Equal(t, 60*time.Second, interval2)

	// 测试用例3: 设置无效值（非数字）
	os.Setenv("REPORT_INTERVAL", "invalid")
	interval3 := GetReportIntervalFromEnv()
	assert.Equal(t, 300*time.Second, interval3)

	// 测试用例4: 设置无效值（负数）
	os.Setenv("REPORT_INTERVAL", "-10")
	interval4 := GetReportIntervalFromEnv()
	assert.Equal(t, 300*time.Second, interval4)

	// 清理环境变量
	os.Unsetenv("REPORT_INTERVAL")
}
