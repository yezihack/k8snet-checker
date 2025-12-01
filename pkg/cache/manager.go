package cache

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"

	gocache "github.com/patrickmn/go-cache"
)

const (
	// 缓存键常量
	checkPodPrefix        = "check-pod:"
	checkVersionKey       = "check-version"
	hostTestResultsKey    = "host-test-results"
	podTestResultsKey     = "pod-test-results"
	serviceTestResultsKey = "service-test-results"

	// 默认配置
	defaultCacheExpiration = 15 * time.Second
	defaultCleanupInterval = 30 * time.Second
)

// CacheManager 定义缓存管理接口
type CacheManager interface {
	// 客户端管理
	UpsertClient(podName string, info *models.NodeInfo) (int64, error)
	GetClient(podName string) (*models.ClientRecord, error)
	GetAllClients() (map[string]*models.ClientRecord, error)
	DeleteClient(podName string) error

	// 版本管理
	GetCurrentVersion() (int64, error)
	UpdateVersion(version int64) error

	// 测试结果管理
	SaveHostTestResults(sourceIP string, results map[string]models.TestStatus) error
	GetHostTestResults() (models.HostTestResults, error)
	SavePodTestResults(sourceIP string, results map[string]models.TestStatus) error
	GetPodTestResults() (models.PodTestResults, error)
	SaveServiceTestResults(sourceIP string, result *models.ConnectivityResult) error
	GetServiceTestResults() (models.ServiceTestResults, error)
}

// cacheManagerImpl 是CacheManager的实现
type cacheManagerImpl struct {
	cache      *gocache.Cache
	mu         sync.RWMutex // 保护版本号更新的互斥锁
	expiration time.Duration
}

// NewCacheManager 创建一个新的CacheManager实例
func NewCacheManager() CacheManager {
	// 从环境变量读取缓存过期时间
	expiration := defaultCacheExpiration
	if cacheKeySecond := os.Getenv("CACHE_KEY_SECOND"); cacheKeySecond != "" {
		if seconds, err := strconv.Atoi(cacheKeySecond); err == nil && seconds > 0 {
			expiration = time.Duration(seconds) * time.Second
		}
	}

	return &cacheManagerImpl{
		cache:      gocache.New(expiration, defaultCleanupInterval),
		expiration: expiration,
	}
}

// UpsertClient 插入或更新客户端记录，并递增版本号
func (cm *cacheManagerImpl) UpsertClient(podName string, info *models.NodeInfo) (int64, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 获取当前版本号
	currentVersion, err := cm.getCurrentVersionLocked()
	if err != nil {
		return 0, fmt.Errorf("获取当前版本号失败: %w", err)
	}

	// 递增版本号
	newVersion := currentVersion + 1

	// 创建客户端记录
	record := &models.ClientRecord{
		NodeInfo:      *info,
		Version:       newVersion,
		LastHeartbeat: time.Now(),
	}

	// 存储客户端记录
	key := checkPodPrefix + podName
	cm.cache.Set(key, record, cm.expiration)

	// 更新全局版本号
	versionInfo := &models.VersionInfo{
		CurrentVersion: newVersion,
		UpdatedAt:      time.Now(),
	}
	cm.cache.Set(checkVersionKey, versionInfo, gocache.NoExpiration)

	return newVersion, nil
}

// GetClient 获取指定客户端的记录
func (cm *cacheManagerImpl) GetClient(podName string) (*models.ClientRecord, error) {
	key := checkPodPrefix + podName
	value, found := cm.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("客户端记录不存在: %s", podName)
	}

	record, ok := value.(*models.ClientRecord)
	if !ok {
		return nil, fmt.Errorf("客户端记录类型错误")
	}

	return record, nil
}

// GetAllClients 获取所有客户端记录
func (cm *cacheManagerImpl) GetAllClients() (map[string]*models.ClientRecord, error) {
	result := make(map[string]*models.ClientRecord)

	// 遍历缓存中所有项
	items := cm.cache.Items()
	for key, item := range items {
		// 只处理客户端记录
		if len(key) > len(checkPodPrefix) && key[:len(checkPodPrefix)] == checkPodPrefix {
			if record, ok := item.Object.(*models.ClientRecord); ok {
				podName := key[len(checkPodPrefix):]
				result[podName] = record
			}
		}
	}

	return result, nil
}

// DeleteClient 删除指定客户端的记录
func (cm *cacheManagerImpl) DeleteClient(podName string) error {
	key := checkPodPrefix + podName
	cm.cache.Delete(key)
	return nil
}

// GetCurrentVersion 获取当前全局版本号
func (cm *cacheManagerImpl) GetCurrentVersion() (int64, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.getCurrentVersionLocked()
}

// getCurrentVersionLocked 获取当前版本号（内部使用，调用者需持有锁）
func (cm *cacheManagerImpl) getCurrentVersionLocked() (int64, error) {
	value, found := cm.cache.Get(checkVersionKey)
	if !found {
		// 如果版本号不存在，返回0
		return 0, nil
	}

	versionInfo, ok := value.(*models.VersionInfo)
	if !ok {
		return 0, fmt.Errorf("版本信息类型错误")
	}

	return versionInfo.CurrentVersion, nil
}

// UpdateVersion 更新全局版本号
func (cm *cacheManagerImpl) UpdateVersion(version int64) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	versionInfo := &models.VersionInfo{
		CurrentVersion: version,
		UpdatedAt:      time.Now(),
	}
	cm.cache.Set(checkVersionKey, versionInfo, gocache.NoExpiration)

	return nil
}

// SaveHostTestResults 保存宿主机测试结果
func (cm *cacheManagerImpl) SaveHostTestResults(sourceIP string, results map[string]models.TestStatus) error {
	// 获取现有的测试结果
	allResults, err := cm.GetHostTestResults()
	if err != nil {
		// 如果获取失败，创建新的结果集
		allResults = make(models.HostTestResults)
	}

	// 更新源IP的测试结果
	allResults[sourceIP] = results

	// 保存回缓存
	cm.cache.Set(hostTestResultsKey, allResults, gocache.NoExpiration)

	return nil
}

// GetHostTestResults 获取所有宿主机测试结果
func (cm *cacheManagerImpl) GetHostTestResults() (models.HostTestResults, error) {
	value, found := cm.cache.Get(hostTestResultsKey)
	if !found {
		return make(models.HostTestResults), nil
	}

	results, ok := value.(models.HostTestResults)
	if !ok {
		return nil, fmt.Errorf("宿主机测试结果类型错误")
	}

	return results, nil
}

// SavePodTestResults 保存Pod测试结果
func (cm *cacheManagerImpl) SavePodTestResults(sourceIP string, results map[string]models.TestStatus) error {
	// 获取现有的测试结果
	allResults, err := cm.GetPodTestResults()
	if err != nil {
		// 如果获取失败，创建新的结果集
		allResults = make(models.PodTestResults)
	}

	// 更新源IP的测试结果
	allResults[sourceIP] = results

	// 保存回缓存
	cm.cache.Set(podTestResultsKey, allResults, gocache.NoExpiration)

	return nil
}

// GetPodTestResults 获取所有Pod测试结果
func (cm *cacheManagerImpl) GetPodTestResults() (models.PodTestResults, error) {
	value, found := cm.cache.Get(podTestResultsKey)
	if !found {
		return make(models.PodTestResults), nil
	}

	results, ok := value.(models.PodTestResults)
	if !ok {
		return nil, fmt.Errorf("Pod测试结果类型错误")
	}

	return results, nil
}

// SaveServiceTestResults 保存自定义服务测试结果
func (cm *cacheManagerImpl) SaveServiceTestResults(sourceIP string, result *models.ConnectivityResult) error {
	// 获取现有的测试结果
	allResults, err := cm.GetServiceTestResults()
	if err != nil {
		// 如果获取失败，创建新的结果集
		allResults = make(models.ServiceTestResults)
	}

	// 更新源IP的测试结果
	allResults[sourceIP] = result

	// 保存回缓存
	cm.cache.Set(serviceTestResultsKey, allResults, gocache.NoExpiration)

	return nil
}

// GetServiceTestResults 获取所有自定义服务测试结果
func (cm *cacheManagerImpl) GetServiceTestResults() (models.ServiceTestResults, error) {
	value, found := cm.cache.Get(serviceTestResultsKey)
	if !found {
		return make(models.ServiceTestResults), nil
	}

	results, ok := value.(models.ServiceTestResults)
	if !ok {
		return nil, fmt.Errorf("服务测试结果类型错误")
	}

	return results, nil
}
