package scheduler

import (
	"context"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/api/client"
	"github.com/yezihack/k8snet-checker/pkg/network"
	"go.uber.org/zap"
)

// TestScheduler 测试任务调度器
type TestScheduler struct {
	apiClient         client.APIClient
	networkTester     network.NetworkTester
	customServiceName string
	logger            *zap.Logger
	interval          time.Duration
}

// NewTestScheduler 创建测试调度器
func NewTestScheduler(
	apiClient client.APIClient,
	networkTester network.NetworkTester,
	customServiceName string,
	logger *zap.Logger,
) *TestScheduler {
	return &TestScheduler{
		apiClient:         apiClient,
		networkTester:     networkTester,
		customServiceName: customServiceName,
		logger:            logger,
		interval:          60 * time.Second,
	}
}

// Start 启动测试调度器
func (s *TestScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 立即执行第一次测试
	s.runTests()

	// 定期执行测试
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("测试调度器收到停止信号")
			return
		case <-ticker.C:
			s.runTests()
		}
	}
}

// runTests 执行所有网络连通性测试
func (s *TestScheduler) runTests() {
	s.logger.Info("开始执行网络连通性测试")

	// 测试宿主机连通性
	s.testHostConnectivity()

	// 测试Pod连通性
	s.testPodConnectivity()

	// 测试自定义服务连通性
	if s.customServiceName != "" {
		s.testServiceConnectivity()
	} else {
		s.logger.Debug("跳过自定义服务探测（未配置CUSTOM_SERVICE_NAME）")
	}

	s.logger.Info("网络连通性测试完成")
}

// testHostConnectivity 测试宿主机连通性
func (s *TestScheduler) testHostConnectivity() {
	s.logger.Info("开始宿主机连通性测试")

	hostIPs, err := s.apiClient.GetHostIPs()
	if err != nil {
		s.logger.Error("获取宿主机IP列表失败", zap.Error(err))
		return
	}

	if len(hostIPs) == 0 {
		s.logger.Info("没有宿主机IP需要测试")
		return
	}

	s.logger.Info("获取到宿主机IP列表", zap.Int("count", len(hostIPs)))

	results, err := s.networkTester.TestHostConnectivity(hostIPs)
	if err != nil {
		s.logger.Error("宿主机连通性测试失败", zap.Error(err))
		return
	}

	s.logger.Info("宿主机连通性测试完成", zap.Int("results_count", len(results)))

	if len(results) > 0 {
		if err := s.apiClient.ReportHostTestResults(results); err != nil {
			s.logger.Error("上报宿主机测试结果失败", zap.Error(err))
			return
		}
		s.logger.Info("宿主机测试结果上报成功")
	}
}

// testPodConnectivity 测试Pod连通性
func (s *TestScheduler) testPodConnectivity() {
	s.logger.Info("开始Pod连通性测试")

	podIPs, err := s.apiClient.GetPodIPs()
	if err != nil {
		s.logger.Error("获取Pod IP列表失败", zap.Error(err))
		return
	}

	if len(podIPs) == 0 {
		s.logger.Info("没有Pod IP需要测试")
		return
	}

	s.logger.Info("获取到Pod IP列表", zap.Int("count", len(podIPs)))

	results, err := s.networkTester.TestPodConnectivity(podIPs)
	if err != nil {
		s.logger.Error("Pod连通性测试失败", zap.Error(err))
		return
	}

	s.logger.Info("Pod连通性测试完成", zap.Int("results_count", len(results)))

	if len(results) > 0 {
		if err := s.apiClient.ReportPodTestResults(results); err != nil {
			s.logger.Error("上报Pod测试结果失败", zap.Error(err))
			return
		}
		s.logger.Info("Pod测试结果上报成功")
	}
}

// testServiceConnectivity 测试自定义服务连通性
func (s *TestScheduler) testServiceConnectivity() {
	s.logger.Info("开始自定义服务连通性测试", zap.String("service_name", s.customServiceName))

	result, err := s.networkTester.TestServiceConnectivity(s.customServiceName)
	if err != nil {
		s.logger.Error("自定义服务连通性测试失败", zap.Error(err))
		return
	}

	s.logger.Info("自定义服务连通性测试完成",
		zap.String("service_name", s.customServiceName),
		zap.String("target_ip", result.TargetIP),
		zap.String("ping_status", result.PingStatus),
	)

	if err := s.apiClient.ReportServiceTestResults(result); err != nil {
		s.logger.Error("上报自定义服务测试结果失败", zap.Error(err))
		return
	}

	s.logger.Info("自定义服务测试结果上报成功")
}
