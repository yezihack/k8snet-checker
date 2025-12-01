package heartbeat

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/api/client"
	"github.com/yezihack/k8snet-checker/pkg/collector"
)

// HeartbeatReporter 定义了心跳上报的接口
type HeartbeatReporter interface {
	// Start 启动心跳上报goroutine
	Start(ctx context.Context, interval time.Duration) error

	// Stop 停止心跳上报
	Stop() error
}

// heartbeatReporterImpl 是HeartbeatReporter的实现
type heartbeatReporterImpl struct {
	collector  collector.InfoCollector
	apiClient  client.APIClient
	cancelFunc context.CancelFunc
	done       chan struct{}
}

// NewHeartbeatReporter 创建一个新的HeartbeatReporter实例
func NewHeartbeatReporter(collector collector.InfoCollector, apiClient client.APIClient) HeartbeatReporter {
	return &heartbeatReporterImpl{
		collector: collector,
		apiClient: apiClient,
		done:      make(chan struct{}),
	}
}

// Start 启动心跳上报goroutine
// interval: 心跳间隔时间
func (r *heartbeatReporterImpl) Start(ctx context.Context, interval time.Duration) error {
	// 创建可取消的context
	ctx, cancel := context.WithCancel(ctx)
	r.cancelFunc = cancel

	log.Printf("启动心跳上报goroutine，间隔: %v", interval)

	// 启动独立的goroutine进行心跳上报
	go r.heartbeatLoop(ctx, interval)

	return nil
}

// Stop 停止心跳上报
func (r *heartbeatReporterImpl) Stop() error {
	if r.cancelFunc != nil {
		log.Println("停止心跳上报goroutine")
		r.cancelFunc()

		// 等待goroutine完成
		<-r.done
		log.Println("心跳上报goroutine已停止")
	}
	return nil
}

// heartbeatLoop 心跳上报循环
func (r *heartbeatReporterImpl) heartbeatLoop(ctx context.Context, interval time.Duration) {
	defer close(r.done)

	// 创建定时器
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即发送第一次心跳
	r.sendHeartbeat()

	// 定期发送心跳
	for {
		select {
		case <-ctx.Done():
			// 收到停止信号
			log.Println("心跳上报goroutine收到停止信号")
			return

		case <-ticker.C:
			// 定时器触发，发送心跳
			r.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送单次心跳
func (r *heartbeatReporterImpl) sendHeartbeat() {
	// 收集节点信息
	nodeInfo, err := r.collector.CollectNodeInfo()
	if err != nil {
		log.Printf("错误: 收集节点信息失败: %v", err)
		return
	}

	// 发送心跳到服务器
	err = r.apiClient.SendHeartbeat(nodeInfo)
	if err != nil {
		log.Printf("错误: 发送心跳失败: %v", err)
		// 注意：这里不返回，下一个心跳周期会继续尝试
		return
	}

	log.Printf("心跳发送成功: pod=%s", nodeInfo.PodName)
}

// GetHeartbeatIntervalFromEnv 从环境变量获取心跳间隔配置
// 默认值为5秒
func GetHeartbeatIntervalFromEnv() time.Duration {
	intervalStr := os.Getenv("HEARTBEAT_INTERVAL")
	if intervalStr == "" {
		return 5 * time.Second
	}

	intervalSec, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("警告: 无法解析HEARTBEAT_INTERVAL环境变量 '%s'，使用默认值5秒: %v",
			intervalStr, err)
		return 5 * time.Second
	}

	if intervalSec <= 0 {
		log.Printf("警告: HEARTBEAT_INTERVAL值无效 (%d)，使用默认值5秒", intervalSec)
		return 5 * time.Second
	}

	return time.Duration(intervalSec) * time.Second
}

// NewHeartbeatReporterFromEnv 从环境变量创建HeartbeatReporter
// 这是一个便捷函数，用于在客户端主程序中快速创建实例
func NewHeartbeatReporterFromEnv(apiClient client.APIClient) (HeartbeatReporter, error) {
	// 创建信息收集器
	infoCollector := collector.NewEnvInfoCollector()

	// 验证能否收集到节点信息
	_, err := infoCollector.CollectNodeInfo()
	if err != nil {
		return nil, err
	}

	// 创建心跳上报器
	reporter := NewHeartbeatReporter(infoCollector, apiClient)

	return reporter, nil
}
