package network

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestNewNetworkTester 测试 NetworkTester 的创建
func TestNewNetworkTester(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name       string
		sourceIP   string
		hostPort   int
		podPort    int
		maxWorkers int
		wantHost   int
		wantPod    int
		wantWorker int
	}{
		{
			name:       "使用默认值",
			sourceIP:   "192.168.1.1",
			hostPort:   0,
			podPort:    0,
			maxWorkers: 0,
			wantHost:   22,
			wantPod:    6100,
			wantWorker: 10,
		},
		{
			name:       "使用自定义值",
			sourceIP:   "192.168.1.2",
			hostPort:   2222,
			podPort:    8080,
			maxWorkers: 5,
			wantHost:   2222,
			wantPod:    8080,
			wantWorker: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewNetworkTester(tt.sourceIP, tt.hostPort, tt.podPort, tt.maxWorkers, logger)
			assert.NotNil(t, tester)

			nt := tester.(*networkTester)
			assert.Equal(t, tt.sourceIP, nt.sourceIP)
			assert.Equal(t, tt.wantHost, nt.hostPort)
			assert.Equal(t, tt.wantPod, nt.podPort)
			assert.Equal(t, tt.wantWorker, nt.maxWorkers)
		})
	}
}

// TestPingTest 测试 ping 功能
func TestPingTest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tester := NewNetworkTester("127.0.0.1", 22, 6100, 10, logger)

	tests := []struct {
		name     string
		targetIP string
		count    int
		wantErr  bool
	}{
		{
			name:     "ping 本地回环地址",
			targetIP: "127.0.0.1",
			count:    3,
			wantErr:  false,
		},
		{
			name:     "ping 不存在的地址",
			targetIP: "192.0.2.1", // TEST-NET-1，保留用于文档和示例
			count:    1,
			wantErr:  false, // 不返回错误，只返回失败状态
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success, latency, err := tester.PingTest(tt.targetIP, tt.count)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.targetIP == "127.0.0.1" {
				assert.True(t, success, "本地回环地址应该可达")
				assert.Greater(t, latency, time.Duration(0), "延迟应该大于 0")
			}
		})
	}
}

// TestPortTest 测试端口连接功能
func TestPortTest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tester := NewNetworkTester("127.0.0.1", 22, 6100, 10, logger)

	tests := []struct {
		name     string
		targetIP string
		port     int
		timeout  time.Duration
		wantErr  bool
	}{
		{
			name:     "测试不存在的端口",
			targetIP: "127.0.0.1",
			port:     9999, // 假设这个端口没有服务
			timeout:  2 * time.Second,
			wantErr:  false, // 不返回错误，只返回失败状态
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, err := tester.PortTest(tt.targetIP, tt.port, tt.timeout)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 端口 9999 通常不会开放
			if tt.port == 9999 {
				assert.False(t, open, "端口 9999 应该是关闭的")
			}
		})
	}
}

// TestTestHostConnectivity 测试宿主机连通性测试
func TestTestHostConnectivity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tester := NewNetworkTester("127.0.0.1", 22, 6100, 10, logger)

	tests := []struct {
		name    string
		hostIPs []string
		wantErr bool
	}{
		{
			name:    "空 IP 列表",
			hostIPs: []string{},
			wantErr: false,
		},
		{
			name:    "单个 IP",
			hostIPs: []string{"127.0.0.1"},
			wantErr: false,
		},
		{
			name:    "多个 IP",
			hostIPs: []string{"127.0.0.1", "192.0.2.1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := tester.TestHostConnectivity(tt.hostIPs)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// 结果数量应该等于输入 IP 数量（排除源 IP）
				expectedCount := len(tt.hostIPs)
				for _, ip := range tt.hostIPs {
					if ip == "127.0.0.1" {
						expectedCount-- // 跳过源 IP
						break
					}
				}
				assert.Equal(t, expectedCount, len(results))
			}
		})
	}
}

// TestTestPodConnectivity 测试 Pod 连通性测试
func TestTestPodConnectivity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tester := NewNetworkTester("10.0.0.1", 22, 6100, 10, logger)

	tests := []struct {
		name    string
		podIPs  []string
		wantErr bool
	}{
		{
			name:    "空 IP 列表",
			podIPs:  []string{},
			wantErr: false,
		},
		{
			name:    "单个 Pod IP",
			podIPs:  []string{"10.0.0.2"},
			wantErr: false,
		},
		{
			name:    "多个 Pod IP",
			podIPs:  []string{"10.0.0.2", "10.0.0.3", "10.0.0.4"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := tester.TestPodConnectivity(tt.podIPs)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// 结果数量应该等于输入 IP 数量（排除源 IP）
				expectedCount := len(tt.podIPs)
				for _, ip := range tt.podIPs {
					if ip == "10.0.0.1" {
						expectedCount-- // 跳过源 IP
						break
					}
				}
				assert.Equal(t, expectedCount, len(results))
			}
		})
	}
}

// TestTestServiceConnectivity 测试自定义服务连通性
func TestTestServiceConnectivity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tester := NewNetworkTester("127.0.0.1", 22, 6100, 10, logger)

	tests := []struct {
		name        string
		serviceName string
		wantErr     bool
	}{
		{
			name:        "空服务名称",
			serviceName: "",
			wantErr:     true,
		},
		{
			name:        "有效的域名",
			serviceName: "localhost",
			wantErr:     false,
		},
		{
			name:        "不存在的域名",
			serviceName: "this-domain-does-not-exist-12345.com",
			wantErr:     false, // DNS 解析失败不返回错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tester.TestServiceConnectivity(tt.serviceName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "127.0.0.1", result.SourceIP)
			}
		})
	}
}

// TestConcurrentTesting 测试并发测试功能
func TestConcurrentTesting(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tester := NewNetworkTester("10.0.0.1", 22, 6100, 3, logger) // 限制 3 个并发

	// 创建少量测试目标以避免超时
	podIPs := []string{
		"10.0.0.2", "10.0.0.3", "10.0.0.4",
	}

	start := time.Now()
	results, err := tester.TestPodConnectivity(podIPs)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, len(podIPs), len(results))

	t.Logf("测试 %d 个目标耗时: %v", len(podIPs), duration)
}
