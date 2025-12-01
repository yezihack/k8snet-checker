package cache

import (
	"os"
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"
)

// TestNewCacheManager 测试创建CacheManager实例
func TestNewCacheManager(t *testing.T) {
	cm := NewCacheManager()
	if cm == nil {
		t.Fatal("NewCacheManager返回nil")
	}
}

// TestNewCacheManagerWithEnvVar 测试使用环境变量创建CacheManager
func TestNewCacheManagerWithEnvVar(t *testing.T) {
	// 设置环境变量
	os.Setenv("CACHE_KEY_SECOND", "20")
	defer os.Unsetenv("CACHE_KEY_SECOND")

	cm := NewCacheManager()
	if cm == nil {
		t.Fatal("NewCacheManager返回nil")
	}

	// 验证缓存管理器可以正常工作
	impl, ok := cm.(*cacheManagerImpl)
	if !ok {
		t.Fatal("类型断言失败")
	}

	if impl.expiration != 20*time.Second {
		t.Errorf("期望过期时间为20秒，实际为%v", impl.expiration)
	}
}

// TestUpsertClient 测试插入和更新客户端记录
func TestUpsertClient(t *testing.T) {
	cm := NewCacheManager()

	nodeInfo := &models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	// 第一次插入
	version1, err := cm.UpsertClient("test-pod", nodeInfo)
	if err != nil {
		t.Fatalf("UpsertClient失败: %v", err)
	}

	if version1 != 1 {
		t.Errorf("期望版本号为1，实际为%d", version1)
	}

	// 第二次更新
	version2, err := cm.UpsertClient("test-pod", nodeInfo)
	if err != nil {
		t.Fatalf("UpsertClient失败: %v", err)
	}

	if version2 != 2 {
		t.Errorf("期望版本号为2，实际为%d", version2)
	}

	// 验证版本号单调递增
	if version2 <= version1 {
		t.Errorf("版本号应该单调递增，v1=%d, v2=%d", version1, version2)
	}
}

// TestGetClient 测试获取客户端记录
func TestGetClient(t *testing.T) {
	cm := NewCacheManager()

	nodeInfo := &models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	// 插入客户端记录
	_, err := cm.UpsertClient("test-pod", nodeInfo)
	if err != nil {
		t.Fatalf("UpsertClient失败: %v", err)
	}

	// 获取客户端记录
	record, err := cm.GetClient("test-pod")
	if err != nil {
		t.Fatalf("GetClient失败: %v", err)
	}

	if record.NodeInfo.PodName != "test-pod" {
		t.Errorf("期望PodName为test-pod，实际为%s", record.NodeInfo.PodName)
	}

	if record.NodeInfo.NodeIP != "192.168.1.1" {
		t.Errorf("期望NodeIP为192.168.1.1，实际为%s", record.NodeInfo.NodeIP)
	}
}

// TestGetClientNotFound 测试获取不存在的客户端记录
func TestGetClientNotFound(t *testing.T) {
	cm := NewCacheManager()

	_, err := cm.GetClient("non-existent-pod")
	if err == nil {
		t.Error("期望返回错误，但没有错误")
	}
}

// TestGetAllClients 测试获取所有客户端记录
func TestGetAllClients(t *testing.T) {
	cm := NewCacheManager()

	// 插入多个客户端记录
	for i := 1; i <= 3; i++ {
		nodeInfo := &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0.1",
			PodName:   "test-pod",
			Timestamp: time.Now(),
		}

		podName := "test-pod-" + string(rune('0'+i))
		_, err := cm.UpsertClient(podName, nodeInfo)
		if err != nil {
			t.Fatalf("UpsertClient失败: %v", err)
		}
	}

	// 获取所有客户端记录
	clients, err := cm.GetAllClients()
	if err != nil {
		t.Fatalf("GetAllClients失败: %v", err)
	}

	if len(clients) != 3 {
		t.Errorf("期望3个客户端记录，实际为%d", len(clients))
	}
}

// TestDeleteClient 测试删除客户端记录
func TestDeleteClient(t *testing.T) {
	cm := NewCacheManager()

	nodeInfo := &models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	// 插入客户端记录
	_, err := cm.UpsertClient("test-pod", nodeInfo)
	if err != nil {
		t.Fatalf("UpsertClient失败: %v", err)
	}

	// 删除客户端记录
	err = cm.DeleteClient("test-pod")
	if err != nil {
		t.Fatalf("DeleteClient失败: %v", err)
	}

	// 验证记录已被删除
	_, err = cm.GetClient("test-pod")
	if err == nil {
		t.Error("期望返回错误，但记录仍然存在")
	}
}

// TestGetCurrentVersion 测试获取当前版本号
func TestGetCurrentVersion(t *testing.T) {
	cm := NewCacheManager()

	// 初始版本号应该为0
	version, err := cm.GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion失败: %v", err)
	}

	if version != 0 {
		t.Errorf("期望初始版本号为0，实际为%d", version)
	}

	// 插入客户端后版本号应该递增
	nodeInfo := &models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	_, err = cm.UpsertClient("test-pod", nodeInfo)
	if err != nil {
		t.Fatalf("UpsertClient失败: %v", err)
	}

	version, err = cm.GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion失败: %v", err)
	}

	if version != 1 {
		t.Errorf("期望版本号为1，实际为%d", version)
	}
}

// TestUpdateVersion 测试更新版本号
func TestUpdateVersion(t *testing.T) {
	cm := NewCacheManager()

	// 更新版本号
	err := cm.UpdateVersion(100)
	if err != nil {
		t.Fatalf("UpdateVersion失败: %v", err)
	}

	// 验证版本号已更新
	version, err := cm.GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion失败: %v", err)
	}

	if version != 100 {
		t.Errorf("期望版本号为100，实际为%d", version)
	}
}

// TestSaveAndGetHostTestResults 测试保存和获取宿主机测试结果
func TestSaveAndGetHostTestResults(t *testing.T) {
	cm := NewCacheManager()

	results := map[string]models.TestStatus{
		"192.168.1.2": {
			Ping:       "reachable",
			PortStatus: "open",
		},
		"192.168.1.3": {
			Ping:       "unreachable",
			PortStatus: "closed",
		},
	}

	// 保存测试结果
	err := cm.SaveHostTestResults("192.168.1.1", results)
	if err != nil {
		t.Fatalf("SaveHostTestResults失败: %v", err)
	}

	// 获取测试结果
	allResults, err := cm.GetHostTestResults()
	if err != nil {
		t.Fatalf("GetHostTestResults失败: %v", err)
	}

	if len(allResults) != 1 {
		t.Errorf("期望1个源IP的测试结果，实际为%d", len(allResults))
	}

	sourceResults, ok := allResults["192.168.1.1"]
	if !ok {
		t.Fatal("未找到源IP 192.168.1.1的测试结果")
	}

	if len(sourceResults) != 2 {
		t.Errorf("期望2个目标IP的测试结果，实际为%d", len(sourceResults))
	}
}

// TestSaveAndGetPodTestResults 测试保存和获取Pod测试结果
func TestSaveAndGetPodTestResults(t *testing.T) {
	cm := NewCacheManager()

	results := map[string]models.TestStatus{
		"10.0.0.2": {
			Ping:       "reachable",
			PortStatus: "open",
		},
	}

	// 保存测试结果
	err := cm.SavePodTestResults("10.0.0.1", results)
	if err != nil {
		t.Fatalf("SavePodTestResults失败: %v", err)
	}

	// 获取测试结果
	allResults, err := cm.GetPodTestResults()
	if err != nil {
		t.Fatalf("GetPodTestResults失败: %v", err)
	}

	if len(allResults) != 1 {
		t.Errorf("期望1个源IP的测试结果，实际为%d", len(allResults))
	}
}

// TestSaveAndGetServiceTestResults 测试保存和获取服务测试结果
func TestSaveAndGetServiceTestResults(t *testing.T) {
	cm := NewCacheManager()

	result := &models.ConnectivityResult{
		SourceIP:   "10.0.0.1",
		TargetIP:   "kubernetes.default.svc.cluster.local",
		PingStatus: "reachable",
		PortStatus: map[int]string{443: "open"},
		Latency:    10 * time.Millisecond,
		Timestamp:  time.Now(),
	}

	// 保存测试结果
	err := cm.SaveServiceTestResults("10.0.0.1", result)
	if err != nil {
		t.Fatalf("SaveServiceTestResults失败: %v", err)
	}

	// 获取测试结果
	allResults, err := cm.GetServiceTestResults()
	if err != nil {
		t.Fatalf("GetServiceTestResults失败: %v", err)
	}

	if len(allResults) != 1 {
		t.Errorf("期望1个源IP的测试结果，实际为%d", len(allResults))
	}

	savedResult, ok := allResults["10.0.0.1"]
	if !ok {
		t.Fatal("未找到源IP 10.0.0.1的测试结果")
	}

	if savedResult.SourceIP != "10.0.0.1" {
		t.Errorf("期望SourceIP为10.0.0.1，实际为%s", savedResult.SourceIP)
	}
}

// TestCacheExpiration 测试缓存过期
func TestCacheExpiration(t *testing.T) {
	// 设置较短的过期时间用于测试
	os.Setenv("CACHE_KEY_SECOND", "1")
	defer os.Unsetenv("CACHE_KEY_SECOND")

	cm := NewCacheManager()

	nodeInfo := &models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod",
		Timestamp: time.Now(),
	}

	// 插入客户端记录
	_, err := cm.UpsertClient("test-pod", nodeInfo)
	if err != nil {
		t.Fatalf("UpsertClient失败: %v", err)
	}

	// 立即获取应该成功
	_, err = cm.GetClient("test-pod")
	if err != nil {
		t.Fatalf("GetClient失败: %v", err)
	}

	// 等待缓存过期
	time.Sleep(2 * time.Second)

	// 再次获取应该失败
	_, err = cm.GetClient("test-pod")
	if err == nil {
		t.Error("期望缓存已过期，但记录仍然存在")
	}
}
