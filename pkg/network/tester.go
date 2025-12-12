package network

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"

	"go.uber.org/zap"
)

// NetworkTester defines the interface for performing network connectivity tests
type NetworkTester interface {
	// PingTest performs a ping test to the target IP
	// count: number of ping attempts
	// Returns: success status, average latency, error
	PingTest(targetIP string, count int) (bool, time.Duration, error)

	// PortTest performs a TCP port connectivity test
	// Returns: true if port is open, false otherwise
	PortTest(targetIP string, port int, timeout time.Duration) (bool, error)

	// TestHostConnectivity tests connectivity to all host IPs
	// Tests both ping and port 22 (or configured port)
	TestHostConnectivity(hostIPs []string) ([]models.ConnectivityResult, error)

	// TestPodConnectivity tests connectivity to all pod IPs
	// Tests both ping and port 6100
	TestPodConnectivity(podIPs []string) ([]models.ConnectivityResult, error)

	// TestServiceConnectivity tests connectivity to a custom service
	// Performs DNS resolution and connectivity test
	TestServiceConnectivity(serviceName string) (*models.ConnectivityResult, error)
}

// networkTester 是 NetworkTester 接口的实现
type networkTester struct {
	sourceIP    string      // 源 IP 地址
	hostPort    int         // 宿主机测试端口（默认 22）
	podPort     int         // Pod 测试端口（默认 6100）
	servicePort int         // 自定义服务测试端口（默认 80）
	maxWorkers  int         // 最大并发 goroutine 数量
	logger      *zap.Logger // 日志记录器
}

// NewNetworkTester 创建一个新的 NetworkTester 实例
func NewNetworkTester(sourceIP string, hostPort, podPort, servicePort, maxWorkers int, logger *zap.Logger) NetworkTester {
	if maxWorkers <= 0 {
		maxWorkers = 10 // 默认最大并发数为 10
	}
	if hostPort <= 0 {
		hostPort = 22 // 默认宿主机端口为 22
	}
	if podPort <= 0 {
		podPort = 6100 // 默认 Pod 端口为 6100
	}
	if servicePort <= 0 {
		servicePort = 80 // 默认服务端口为 80
	}

	return &networkTester{
		sourceIP:    sourceIP,
		hostPort:    hostPort,
		podPort:     podPort,
		servicePort: servicePort,
		maxWorkers:  maxWorkers,
		logger:      logger,
	}
}

// PingTest 执行 ping 测试
func (nt *networkTester) PingTest(targetIP string, count int) (bool, time.Duration, error) {
	if count <= 0 {
		count = 3 // 默认 ping 3 次
	}

	// 根据操作系统选择 ping 命令参数
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), "-w", "5000", targetIP)
	} else {
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-W", "5", targetIP)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	if err != nil {
		nt.logger.Debug("ping 测试失败",
			zap.String("target_ip", targetIP),
			zap.Error(err),
			zap.String("output", string(output)),
		)
		return false, 0, nil // ping 失败不返回错误，只返回失败状态
	}

	// 计算平均延迟
	avgLatency := latency / time.Duration(count)

	nt.logger.Debug("ping 测试成功",
		zap.String("target_ip", targetIP),
		zap.Duration("latency", avgLatency),
	)

	return true, avgLatency, nil
}

// PortTest 执行 TCP 端口连接测试
func (nt *networkTester) PortTest(targetIP string, port int, timeout time.Duration) (bool, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second // 默认超时 5 秒
	}

	address := net.JoinHostPort(targetIP, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		nt.logger.Debug("端口测试失败",
			zap.String("target_ip", targetIP),
			zap.Int("port", port),
			zap.Error(err),
		)
		return false, nil // 端口不可达不返回错误，只返回失败状态
	}

	conn.Close()

	nt.logger.Debug("端口测试成功",
		zap.String("target_ip", targetIP),
		zap.Int("port", port),
	)

	return true, nil
}

// TestHostConnectivity 测试所有宿主机 IP 的连通性
func (nt *networkTester) TestHostConnectivity(hostIPs []string) ([]models.ConnectivityResult, error) {
	return nt.testConnectivity(hostIPs, nt.hostPort, "宿主机")
}

// TestPodConnectivity 测试所有 Pod IP 的连通性
func (nt *networkTester) TestPodConnectivity(podIPs []string) ([]models.ConnectivityResult, error) {
	return nt.testConnectivity(podIPs, nt.podPort, "Pod")
}

// testConnectivity 是通用的连通性测试方法，支持并发测试
func (nt *networkTester) testConnectivity(targetIPs []string, port int, testType string) ([]models.ConnectivityResult, error) {
	if len(targetIPs) == 0 {
		nt.logger.Info("没有目标 IP 需要测试", zap.String("test_type", testType))
		return []models.ConnectivityResult{}, nil
	}

	nt.logger.Info("开始连通性测试",
		zap.String("test_type", testType),
		zap.Int("target_count", len(targetIPs)),
		zap.Int("port", port),
	)

	// 创建结果切片和互斥锁
	results := make([]models.ConnectivityResult, 0, len(targetIPs))
	var resultsMutex sync.Mutex

	// 创建工作池，限制并发数
	semaphore := make(chan struct{}, nt.maxWorkers)
	var wg sync.WaitGroup

	// 并发测试每个目标 IP
	for _, targetIP := range targetIPs {
		// 跳过自己
		if targetIP == nt.sourceIP {
			continue
		}

		wg.Add(1)
		go func(target string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行测试
			result := nt.testSingleTarget(target, port)

			// 将结果添加到切片
			resultsMutex.Lock()
			results = append(results, result)
			resultsMutex.Unlock()
		}(targetIP)
	}

	// 等待所有测试完成
	wg.Wait()

	nt.logger.Info("连通性测试完成",
		zap.String("test_type", testType),
		zap.Int("tested_count", len(results)),
	)

	return results, nil
}

// testSingleTarget 测试单个目标的连通性
func (nt *networkTester) testSingleTarget(targetIP string, port int) models.ConnectivityResult {
	startTime := time.Now()

	result := models.ConnectivityResult{
		SourceIP:   nt.sourceIP,
		TargetIP:   targetIP,
		PortStatus: make(map[int]string),
		Timestamp:  startTime,
	}

	// 执行 ping 测试
	pingSuccess, latency, _ := nt.PingTest(targetIP, 3)
	if pingSuccess {
		result.PingStatus = "reachable"
		result.Latency = models.Duration(latency)
	} else {
		result.PingStatus = "unreachable"
		result.Latency = 0
	}

	// 执行端口测试
	portOpen, _ := nt.PortTest(targetIP, port, 5*time.Second)
	if portOpen {
		result.PortStatus[port] = "open"
	} else {
		result.PortStatus[port] = "closed"
	}

	// 记录测试耗时
	result.TestDuration = models.Duration(time.Since(startTime))

	nt.logger.Debug("单个目标测试完成",
		zap.String("target_ip", targetIP),
		zap.String("ping_status", result.PingStatus),
		zap.String("port_status", result.PortStatus[port]),
		zap.String("test_duration", result.TestDuration.String()),
	)

	return result
}

// TestServiceConnectivity 测试自定义服务的连通性
func (nt *networkTester) TestServiceConnectivity(serviceName string) (*models.ConnectivityResult, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("服务名称不能为空")
	}

	startTime := time.Now()
	nt.logger.Info("开始自定义服务测试", zap.String("service_name", serviceName))

	result := &models.ConnectivityResult{
		SourceIP:   nt.sourceIP,
		PortStatus: make(map[int]string),
		Timestamp:  startTime,
	}

	// 执行 DNS 解析
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupHost(ctx, serviceName)
	if err != nil {
		nt.logger.Warn("DNS 解析失败",
			zap.String("service_name", serviceName),
			zap.Error(err),
		)
		result.TargetIP = serviceName
		result.PingStatus = "unreachable"
		result.TestDuration = models.Duration(time.Since(startTime))
		return result, nil
	}

	if len(ips) == 0 {
		nt.logger.Warn("DNS 解析未返回 IP",
			zap.String("service_name", serviceName),
		)
		result.TargetIP = serviceName
		result.PingStatus = "unreachable"
		result.TestDuration = models.Duration(time.Since(startTime))
		return result, nil
	}

	// 使用第一个解析的 IP 地址
	targetIP := ips[0]
	result.TargetIP = targetIP

	nt.logger.Info("DNS 解析成功",
		zap.String("service_name", serviceName),
		zap.String("resolved_ip", targetIP),
		zap.Strings("all_ips", ips),
	)

	// 执行 ping 测试
	pingSuccess, latency, _ := nt.PingTest(targetIP, 3)
	if pingSuccess {
		result.PingStatus = "reachable"
		result.Latency = models.Duration(latency)
	} else {
		result.PingStatus = "unreachable"
		result.Latency = 0
	}

	// 测试配置的服务端口
	portOpen, _ := nt.PortTest(targetIP, nt.servicePort, 5*time.Second)
	if portOpen {
		result.PortStatus[nt.servicePort] = "open"
	} else {
		result.PortStatus[nt.servicePort] = "closed"
	}

	// 记录测试耗时
	result.TestDuration = models.Duration(time.Since(startTime))

	nt.logger.Info("自定义服务测试完成",
		zap.String("service_name", serviceName),
		zap.String("target_ip", targetIP),
		zap.String("ping_status", result.PingStatus),
		zap.Duration("test_duration", time.Duration(result.TestDuration)),
	)

	return result, nil
}
