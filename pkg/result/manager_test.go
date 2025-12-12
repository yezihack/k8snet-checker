package result

import (
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestNewTestResultManager(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	assert.NotNil(t, manager, "TestResultManager不应为空")
}

func TestSaveHostTestResults(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	// 准备测试数据
	sourceIP := "192.168.1.1"
	results := []models.ConnectivityResult{
		{
			SourceIP:   sourceIP,
			TargetIP:   "192.168.1.2",
			PingStatus: "reachable",
			PortStatus: map[int]string{22: "open"},
			Latency:    10 * time.Millisecond,
			Timestamp:  time.Now(),
		},
		{
			SourceIP:   sourceIP,
			TargetIP:   "192.168.1.3",
			PingStatus: "unreachable",
			PortStatus: map[int]string{22: "closed"},
			Latency:    0,
			Timestamp:  time.Now(),
		},
	}

	// 保存测试结果
	err := manager.SaveHostTestResults(sourceIP, results)
	assert.NoError(t, err, "保存宿主机测试结果不应出错")

	// 验证保存的结果
	allResults, err := manager.GetHostTestResults()
	assert.NoError(t, err, "获取宿主机测试结果不应出错")
	assert.Contains(t, allResults, sourceIP, "结果应包含源IP")

	sourceResults := allResults[sourceIP]
	assert.Len(t, sourceResults, 2, "应有2个目标IP的测试结果")

	// 验证第一个结果
	assert.Contains(t, sourceResults, "192.168.1.2")
	assert.Equal(t, "reachable", sourceResults["192.168.1.2"].Ping)
	assert.Equal(t, "open", sourceResults["192.168.1.2"].PortStatus)

	// 验证第二个结果
	assert.Contains(t, sourceResults, "192.168.1.3")
	assert.Equal(t, "unreachable", sourceResults["192.168.1.3"].Ping)
	assert.Equal(t, "closed", sourceResults["192.168.1.3"].PortStatus)
}

func TestSaveHostTestResults_EmptySourceIP(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	results := []models.ConnectivityResult{}
	err := manager.SaveHostTestResults("", results)

	assert.Error(t, err, "空源IP应返回错误")
	assert.Contains(t, err.Error(), "源IP不能为空")
}

func TestSavePodTestResults(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	// 准备测试数据
	sourceIP := "10.244.1.1"
	results := []models.ConnectivityResult{
		{
			SourceIP:   sourceIP,
			TargetIP:   "10.244.1.2",
			PingStatus: "reachable",
			PortStatus: map[int]string{6100: "open"},
			Latency:    5 * time.Millisecond,
			Timestamp:  time.Now(),
		},
		{
			SourceIP:   sourceIP,
			TargetIP:   "10.244.1.3",
			PingStatus: "reachable",
			PortStatus: map[int]string{6100: "open"},
			Latency:    8 * time.Millisecond,
			Timestamp:  time.Now(),
		},
	}

	// 保存测试结果
	err := manager.SavePodTestResults(sourceIP, results)
	assert.NoError(t, err, "保存Pod测试结果不应出错")

	// 验证保存的结果
	allResults, err := manager.GetPodTestResults()
	assert.NoError(t, err, "获取Pod测试结果不应出错")
	assert.Contains(t, allResults, sourceIP, "结果应包含源IP")

	sourceResults := allResults[sourceIP]
	assert.Len(t, sourceResults, 2, "应有2个目标IP的测试结果")

	// 验证结果
	assert.Contains(t, sourceResults, "10.244.1.2")
	assert.Equal(t, "reachable", sourceResults["10.244.1.2"].Ping)
	assert.Equal(t, "open", sourceResults["10.244.1.2"].PortStatus)
}

func TestSavePodTestResults_EmptySourceIP(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	results := []models.ConnectivityResult{}
	err := manager.SavePodTestResults("", results)

	assert.Error(t, err, "空源IP应返回错误")
	assert.Contains(t, err.Error(), "源IP不能为空")
}

func TestSaveServiceTestResult(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	// 准备测试数据
	sourceIP := "10.244.1.1"
	result := &models.ConnectivityResult{
		SourceIP:   sourceIP,
		TargetIP:   "kubernetes.default.svc.cluster.local",
		PingStatus: "reachable",
		PortStatus: map[int]string{443: "open"},
		Latency:    15 * time.Millisecond,
		Timestamp:  time.Now(),
	}

	// 保存测试结果
	err := manager.SaveServiceTestResult(sourceIP, result)
	assert.NoError(t, err, "保存服务测试结果不应出错")

	// 验证保存的结果
	allResults, err := manager.GetServiceTestResults()
	assert.NoError(t, err, "获取服务测试结果不应出错")
	assert.Contains(t, allResults, sourceIP, "结果应包含源IP")

	savedResult := allResults[sourceIP]
	assert.Equal(t, result.TargetIP, savedResult.TargetIP)
	assert.Equal(t, result.PingStatus, savedResult.PingStatus)
	assert.Equal(t, "open", savedResult.PortStatus[443])
}

func TestSaveServiceTestResult_EmptySourceIP(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	result := &models.ConnectivityResult{
		TargetIP: "test.svc",
	}
	err := manager.SaveServiceTestResult("", result)

	assert.Error(t, err, "空源IP应返回错误")
	assert.Contains(t, err.Error(), "源IP不能为空")
}

func TestSaveServiceTestResult_NilResult(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	err := manager.SaveServiceTestResult("10.244.1.1", nil)

	assert.Error(t, err, "空结果应返回错误")
	assert.Contains(t, err.Error(), "测试结果不能为空")
}

func TestGetHostTestResults_Empty(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	results, err := manager.GetHostTestResults()

	assert.NoError(t, err, "获取空结果不应出错")
	assert.NotNil(t, results, "结果不应为nil")
	assert.Empty(t, results, "结果应为空")
}

func TestGetPodTestResults_Empty(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	results, err := manager.GetPodTestResults()

	assert.NoError(t, err, "获取空结果不应出错")
	assert.NotNil(t, results, "结果不应为nil")
	assert.Empty(t, results, "结果应为空")
}

func TestGetServiceTestResults_Empty(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	results, err := manager.GetServiceTestResults()

	assert.NoError(t, err, "获取空结果不应出错")
	assert.NotNil(t, results, "结果不应为nil")
	assert.Empty(t, results, "结果应为空")
}

func TestMultipleSourceIPs(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	// 保存第一个源IP的结果
	source1 := "192.168.1.1"
	results1 := []models.ConnectivityResult{
		{
			SourceIP:   source1,
			TargetIP:   "192.168.1.2",
			PingStatus: "reachable",
			PortStatus: map[int]string{22: "open"},
		},
	}
	err := manager.SaveHostTestResults(source1, results1)
	assert.NoError(t, err)

	// 保存第二个源IP的结果
	source2 := "192.168.1.3"
	results2 := []models.ConnectivityResult{
		{
			SourceIP:   source2,
			TargetIP:   "192.168.1.4",
			PingStatus: "unreachable",
			PortStatus: map[int]string{22: "closed"},
		},
	}
	err = manager.SaveHostTestResults(source2, results2)
	assert.NoError(t, err)

	// 验证两个源IP的结果都存在
	allResults, err := manager.GetHostTestResults()
	assert.NoError(t, err)
	assert.Len(t, allResults, 2, "应有2个源IP的结果")
	assert.Contains(t, allResults, source1)
	assert.Contains(t, allResults, source2)
}

func TestUpdateExistingResults(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	manager := NewTestResultManager(cacheManager)

	sourceIP := "192.168.1.1"

	// 第一次保存
	results1 := []models.ConnectivityResult{
		{
			SourceIP:   sourceIP,
			TargetIP:   "192.168.1.2",
			PingStatus: "reachable",
			PortStatus: map[int]string{22: "open"},
		},
	}
	err := manager.SaveHostTestResults(sourceIP, results1)
	assert.NoError(t, err)

	// 第二次保存（更新）
	results2 := []models.ConnectivityResult{
		{
			SourceIP:   sourceIP,
			TargetIP:   "192.168.1.3",
			PingStatus: "unreachable",
			PortStatus: map[int]string{22: "closed"},
		},
	}
	err = manager.SaveHostTestResults(sourceIP, results2)
	assert.NoError(t, err)

	// 验证结果被更新
	allResults, err := manager.GetHostTestResults()
	assert.NoError(t, err)
	assert.Len(t, allResults[sourceIP], 1, "应只有最新的1个目标IP")
	assert.Contains(t, allResults[sourceIP], "192.168.1.3")
	assert.NotContains(t, allResults[sourceIP], "192.168.1.2", "旧结果应被覆盖")
}
