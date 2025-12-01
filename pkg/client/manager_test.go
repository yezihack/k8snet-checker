package client

import (
	"testing"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/models"
)

// TestHandleHeartbeat 测试心跳处理功能
func TestHandleHeartbeat(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	// 测试正常心跳处理
	nodeInfo := &models.NodeInfo{
		Namespace: "default",
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		PodName:   "test-pod-1",
		Timestamp: time.Now(),
	}

	err := clientManager.HandleHeartbeat(nodeInfo)
	if err != nil {
		t.Errorf("HandleHeartbeat失败: %v", err)
	}

	// 验证客户端记录已保存
	record, err := cacheManager.GetClient("test-pod-1")
	if err != nil {
		t.Errorf("获取客户端记录失败: %v", err)
	}
	if record.NodeInfo.NodeIP != "192.168.1.1" {
		t.Errorf("NodeIP不匹配: 期望=%s, 实际=%s", "192.168.1.1", record.NodeInfo.NodeIP)
	}
	if record.Version != 1 {
		t.Errorf("版本号不匹配: 期望=%d, 实际=%d", 1, record.Version)
	}
}

// TestHandleHeartbeatNilInfo 测试空NodeInfo的错误处理
func TestHandleHeartbeatNilInfo(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	err := clientManager.HandleHeartbeat(nil)
	if err == nil {
		t.Error("期望返回错误，但没有错误")
	}
}

// TestHandleHeartbeatMissingFields 测试缺少必需字段的错误处理
func TestHandleHeartbeatMissingFields(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	// 测试缺少PodName
	nodeInfo := &models.NodeInfo{
		NodeIP:    "192.168.1.1",
		PodIP:     "10.0.0.1",
		Timestamp: time.Now(),
	}
	err := clientManager.HandleHeartbeat(nodeInfo)
	if err == nil {
		t.Error("期望返回错误（缺少PodName），但没有错误")
	}

	// 测试缺少NodeIP
	nodeInfo = &models.NodeInfo{
		PodName:   "test-pod",
		PodIP:     "10.0.0.1",
		Timestamp: time.Now(),
	}
	err = clientManager.HandleHeartbeat(nodeInfo)
	if err == nil {
		t.Error("期望返回错误（缺少NodeIP），但没有错误")
	}

	// 测试缺少PodIP
	nodeInfo = &models.NodeInfo{
		PodName:   "test-pod",
		NodeIP:    "192.168.1.1",
		Timestamp: time.Now(),
	}
	err = clientManager.HandleHeartbeat(nodeInfo)
	if err == nil {
		t.Error("期望返回错误（缺少PodIP），但没有错误")
	}
}

// TestGetActiveClientCount 测试活跃客户端统计
func TestGetActiveClientCount(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	// 添加多个客户端
	// 版本号会递增：1, 2, 3, 4, 5
	// 当前版本号是5
	// 版本号等于5的客户端：1个
	// 版本号与5相差小于3的客户端：3, 4, 5，共3个
	// 由于精确匹配数（1）没有超过总数的一半（5/2=2.5），返回近似匹配数3
	for i := 1; i <= 5; i++ {
		nodeInfo := &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0." + string(rune('0'+i)),
			PodName:   "test-pod-" + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		err := clientManager.HandleHeartbeat(nodeInfo)
		if err != nil {
			t.Errorf("HandleHeartbeat失败: %v", err)
		}
	}

	// 获取活跃客户端数量
	count, err := clientManager.GetActiveClientCount()
	if err != nil {
		t.Errorf("GetActiveClientCount失败: %v", err)
	}

	// 根据活跃客户端统计逻辑，应该返回3（版本号相差小于3的客户端）
	if count != 3 {
		t.Errorf("活跃客户端数量不匹配: 期望=%d, 实际=%d", 3, count)
	}
}

// TestGetActiveClientCountEmpty 测试空客户端列表
func TestGetActiveClientCountEmpty(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	count, err := clientManager.GetActiveClientCount()
	if err != nil {
		t.Errorf("GetActiveClientCount失败: %v", err)
	}
	if count != 0 {
		t.Errorf("活跃客户端数量应为0: 实际=%d", count)
	}
}

// TestGetAllHostIPs 测试获取所有宿主机IP
func TestGetAllHostIPs(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	// 添加多个客户端，有些在同一宿主机上
	clients := []struct {
		podName string
		nodeIP  string
		podIP   string
	}{
		{"pod-1", "192.168.1.1", "10.0.0.1"},
		{"pod-2", "192.168.1.1", "10.0.0.2"},
		{"pod-3", "192.168.1.2", "10.0.0.3"},
		{"pod-4", "192.168.1.3", "10.0.0.4"},
	}

	for _, c := range clients {
		nodeInfo := &models.NodeInfo{
			Namespace: "default",
			NodeIP:    c.nodeIP,
			PodIP:     c.podIP,
			PodName:   c.podName,
			Timestamp: time.Now(),
		}
		err := clientManager.HandleHeartbeat(nodeInfo)
		if err != nil {
			t.Errorf("HandleHeartbeat失败: %v", err)
		}
	}

	// 获取宿主机IP列表
	hostIPs, err := clientManager.GetAllHostIPs()
	if err != nil {
		t.Errorf("GetAllHostIPs失败: %v", err)
	}

	// 应该有3个不同的宿主机IP
	if len(hostIPs) != 3 {
		t.Errorf("宿主机IP数量不匹配: 期望=%d, 实际=%d", 3, len(hostIPs))
	}

	// 验证IP去重
	ipMap := make(map[string]bool)
	for _, ip := range hostIPs {
		if ipMap[ip] {
			t.Errorf("宿主机IP重复: %s", ip)
		}
		ipMap[ip] = true
	}
}

// TestGetAllPodIPs 测试获取所有Pod IP
func TestGetAllPodIPs(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	// 添加多个客户端
	for i := 1; i <= 4; i++ {
		nodeInfo := &models.NodeInfo{
			Namespace: "default",
			NodeIP:    "192.168.1.1",
			PodIP:     "10.0.0." + string(rune('0'+i)),
			PodName:   "test-pod-" + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		err := clientManager.HandleHeartbeat(nodeInfo)
		if err != nil {
			t.Errorf("HandleHeartbeat失败: %v", err)
		}
	}

	// 获取Pod IP列表
	podIPs, err := clientManager.GetAllPodIPs()
	if err != nil {
		t.Errorf("GetAllPodIPs失败: %v", err)
	}

	// 应该有4个Pod IP
	if len(podIPs) != 4 {
		t.Errorf("Pod IP数量不匹配: 期望=%d, 实际=%d", 4, len(podIPs))
	}

	// 验证IP去重
	ipMap := make(map[string]bool)
	for _, ip := range podIPs {
		if ipMap[ip] {
			t.Errorf("Pod IP重复: %s", ip)
		}
		ipMap[ip] = true
	}
}

// TestGetAllHostIPsEmpty 测试空客户端列表时获取宿主机IP
func TestGetAllHostIPsEmpty(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	hostIPs, err := clientManager.GetAllHostIPs()
	if err != nil {
		t.Errorf("GetAllHostIPs失败: %v", err)
	}
	if len(hostIPs) != 0 {
		t.Errorf("宿主机IP列表应为空: 实际长度=%d", len(hostIPs))
	}
}

// TestGetAllPodIPsEmpty 测试空客户端列表时获取Pod IP
func TestGetAllPodIPsEmpty(t *testing.T) {
	cacheManager := cache.NewCacheManager()
	clientManager := NewClientManager(cacheManager)

	podIPs, err := clientManager.GetAllPodIPs()
	if err != nil {
		t.Errorf("GetAllPodIPs失败: %v", err)
	}
	if len(podIPs) != 0 {
		t.Errorf("Pod IP列表应为空: 实际长度=%d", len(podIPs))
	}
}
