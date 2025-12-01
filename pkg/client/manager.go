package client

import (
	"fmt"
	"log"

	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/models"
)

// ClientManager 定义客户端管理接口
type ClientManager interface {
	// HandleHeartbeat 处理客户端心跳，更新缓存并递增版本号
	HandleHeartbeat(info *models.NodeInfo) error

	// GetActiveClientCount 获取活跃客户端数量
	GetActiveClientCount() (int, error)

	// GetAllHostIPs 获取所有宿主机IP列表
	GetAllHostIPs() ([]string, error)

	// GetAllPodIPs 获取所有Pod IP列表
	GetAllPodIPs() ([]string, error)
}

// clientManagerImpl 是ClientManager的实现
type clientManagerImpl struct {
	cacheManager cache.CacheManager
}

// NewClientManager 创建一个新的ClientManager实例
func NewClientManager(cacheManager cache.CacheManager) ClientManager {
	return &clientManagerImpl{
		cacheManager: cacheManager,
	}
}

// HandleHeartbeat 处理客户端心跳
// 接收NodeInfo，更新缓存，递增版本号
func (cm *clientManagerImpl) HandleHeartbeat(info *models.NodeInfo) error {
	if info == nil {
		return fmt.Errorf("NodeInfo不能为空")
	}

	// 验证必需字段
	if info.PodName == "" {
		return fmt.Errorf("PodName不能为空")
	}
	if info.NodeIP == "" {
		return fmt.Errorf("NodeIP不能为空")
	}
	if info.PodIP == "" {
		return fmt.Errorf("PodIP不能为空")
	}

	// 调用CacheManager更新客户端记录并递增版本号
	version, err := cm.cacheManager.UpsertClient(info.PodName, info)
	if err != nil {
		log.Printf("处理心跳失败: pod=%s, error=%v", info.PodName, err)
		return fmt.Errorf("更新客户端记录失败: %w", err)
	}

	log.Printf("心跳处理成功: pod=%s, node_ip=%s, pod_ip=%s, version=%d",
		info.PodName, info.NodeIP, info.PodIP, version)

	return nil
}

// GetActiveClientCount 获取活跃客户端数量
// 活跃客户端统计逻辑：
// 1. 获取当前全局版本号 currentVersion
// 2. 遍历所有客户端记录
// 3. 统计版本号等于 currentVersion 的客户端数量
// 4. 统计版本号与 currentVersion 相差小于3的客户端数量
// 5. 如果版本号匹配的客户端数量超过总数的一半，则这些客户端被认为是活跃的
// 6. 返回活跃客户端总数
func (cm *clientManagerImpl) GetActiveClientCount() (int, error) {
	// 获取当前全局版本号
	currentVersion, err := cm.cacheManager.GetCurrentVersion()
	if err != nil {
		return 0, fmt.Errorf("获取当前版本号失败: %w", err)
	}

	// 获取所有客户端记录
	allClients, err := cm.cacheManager.GetAllClients()
	if err != nil {
		return 0, fmt.Errorf("获取所有客户端失败: %w", err)
	}

	// 如果没有客户端，返回0
	if len(allClients) == 0 {
		return 0, nil
	}

	// 统计版本号匹配的客户端数量
	exactMatchCount := 0
	nearMatchCount := 0

	for _, record := range allClients {
		// 统计版本号等于 currentVersion 的客户端
		if record.Version == currentVersion {
			exactMatchCount++
		}

		// 统计版本号与 currentVersion 相差小于3的客户端
		versionDiff := currentVersion - record.Version
		if versionDiff < 0 {
			versionDiff = -versionDiff
		}
		if versionDiff < 3 {
			nearMatchCount++
		}
	}

	totalClients := len(allClients)

	// 如果版本号匹配的客户端数量超过总数的一半，则这些客户端被认为是活跃的
	if exactMatchCount > totalClients/2 {
		log.Printf("活跃客户端统计: 精确匹配=%d, 总数=%d, 活跃数=%d",
			exactMatchCount, totalClients, exactMatchCount)
		return exactMatchCount, nil
	}

	// 否则返回近似匹配的客户端数量
	log.Printf("活跃客户端统计: 精确匹配=%d, 近似匹配=%d, 总数=%d, 活跃数=%d",
		exactMatchCount, nearMatchCount, totalClients, nearMatchCount)
	return nearMatchCount, nil
}

// GetAllHostIPs 获取所有宿主机IP列表
func (cm *clientManagerImpl) GetAllHostIPs() ([]string, error) {
	// 获取所有客户端记录
	allClients, err := cm.cacheManager.GetAllClients()
	if err != nil {
		return nil, fmt.Errorf("获取所有客户端失败: %w", err)
	}

	// 使用map去重
	hostIPMap := make(map[string]bool)
	for _, record := range allClients {
		if record.NodeInfo.NodeIP != "" {
			hostIPMap[record.NodeInfo.NodeIP] = true
		}
	}

	// 转换为切片
	hostIPs := make([]string, 0, len(hostIPMap))
	for ip := range hostIPMap {
		hostIPs = append(hostIPs, ip)
	}

	log.Printf("获取宿主机IP列表: 数量=%d", len(hostIPs))
	return hostIPs, nil
}

// GetAllPodIPs 获取所有Pod IP列表
func (cm *clientManagerImpl) GetAllPodIPs() ([]string, error) {
	// 获取所有客户端记录
	allClients, err := cm.cacheManager.GetAllClients()
	if err != nil {
		return nil, fmt.Errorf("获取所有客户端失败: %w", err)
	}

	// 使用map去重
	podIPMap := make(map[string]bool)
	for _, record := range allClients {
		if record.NodeInfo.PodIP != "" {
			podIPMap[record.NodeInfo.PodIP] = true
		}
	}

	// 转换为切片
	podIPs := make([]string, 0, len(podIPMap))
	for ip := range podIPMap {
		podIPs = append(podIPs, ip)
	}

	log.Printf("获取Pod IP列表: 数量=%d", len(podIPs))
	return podIPs, nil
}
