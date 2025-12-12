package result

import (
	"fmt"

	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/models"
)

// TestResultManager 定义测试结果管理接口
type TestResultManager interface {
	SaveHostTestResults(sourceIP string, results []models.ConnectivityResult) error
	SavePodTestResults(sourceIP string, results []models.ConnectivityResult) error
	SaveServiceTestResult(sourceIP string, result *models.ConnectivityResult) error
	GetHostTestResults() (models.HostTestResults, error)
	GetPodTestResults() (models.PodTestResults, error)
	GetServiceTestResults() (models.ServiceTestResults, error)
}

// testResultManagerImpl 是TestResultManager的实现
type testResultManagerImpl struct {
	cacheManager cache.CacheManager
}

// NewTestResultManager 创建一个新的TestResultManager实例
func NewTestResultManager(cacheManager cache.CacheManager) TestResultManager {
	return &testResultManagerImpl{
		cacheManager: cacheManager,
	}
}

// SaveHostTestResults 保存宿主机测试结果
// 将ConnectivityResult列表转换为TestStatus格式并存储
func (m *testResultManagerImpl) SaveHostTestResults(sourceIP string, results []models.ConnectivityResult) error {
	if sourceIP == "" {
		return fmt.Errorf("源IP不能为空")
	}

	// 转换ConnectivityResult为TestStatus格式
	testStatusMap := make(map[string]models.TestStatus)
	for _, result := range results {
		if result.TargetIP == "" {
			continue // 跳过无效的目标IP
		}

		// 提取端口状态（取第一个端口的状态）
		portStatus := "unknown"
		for _, status := range result.PortStatus {
			portStatus = status
			break // 只取第一个端口状态
		}

		testStatusMap[result.TargetIP] = models.TestStatus{
			Ping:         result.PingStatus,
			PortStatus:   portStatus,
			TestDuration: result.TestDuration,
		}
	}

	// 保存到缓存
	return m.cacheManager.SaveHostTestResults(sourceIP, testStatusMap)
}

// SavePodTestResults 保存Pod测试结果
// 将ConnectivityResult列表转换为TestStatus格式并存储
func (m *testResultManagerImpl) SavePodTestResults(sourceIP string, results []models.ConnectivityResult) error {
	if sourceIP == "" {
		return fmt.Errorf("源IP不能为空")
	}

	// 转换ConnectivityResult为TestStatus格式
	testStatusMap := make(map[string]models.TestStatus)
	for _, result := range results {
		if result.TargetIP == "" {
			continue // 跳过无效的目标IP
		}

		// 提取端口状态（取第一个端口的状态）
		portStatus := "unknown"
		for _, status := range result.PortStatus {
			portStatus = status
			break // 只取第一个端口状态
		}

		testStatusMap[result.TargetIP] = models.TestStatus{
			Ping:         result.PingStatus,
			PortStatus:   portStatus,
			TestDuration: result.TestDuration,
		}
	}

	// 保存到缓存
	return m.cacheManager.SavePodTestResults(sourceIP, testStatusMap)
}

// SaveServiceTestResult 保存自定义服务测试结果
func (m *testResultManagerImpl) SaveServiceTestResult(sourceIP string, result *models.ConnectivityResult) error {
	if sourceIP == "" {
		return fmt.Errorf("源IP不能为空")
	}

	if result == nil {
		return fmt.Errorf("测试结果不能为空")
	}

	// 直接保存ConnectivityResult
	return m.cacheManager.SaveServiceTestResults(sourceIP, result)
}

// GetHostTestResults 获取所有宿主机测试结果
func (m *testResultManagerImpl) GetHostTestResults() (models.HostTestResults, error) {
	return m.cacheManager.GetHostTestResults()
}

// GetPodTestResults 获取所有Pod测试结果
func (m *testResultManagerImpl) GetPodTestResults() (models.PodTestResults, error) {
	return m.cacheManager.GetPodTestResults()
}

// GetServiceTestResults 获取所有自定义服务测试结果
func (m *testResultManagerImpl) GetServiceTestResults() (models.ServiceTestResults, error) {
	return m.cacheManager.GetServiceTestResults()
}
