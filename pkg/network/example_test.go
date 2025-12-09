package network

import (
	"fmt"

	"go.uber.org/zap"
)

// ExampleNetworkTester 展示如何使用 NetworkTester
func ExampleNetworkTester() {
	// 创建日志记录器
	logger, _ := zap.NewDevelopment()

	// 创建 NetworkTester 实例
	// 参数：源IP, 宿主机端口, Pod端口, 服务端口, 最大并发数, 日志记录器
	tester := NewNetworkTester("192.168.1.100", 22, 6100, 80, 10, logger)

	// 1. 执行 ping 测试
	success, latency, err := tester.PingTest("192.168.1.1", 3)
	if err != nil {
		fmt.Printf("Ping 测试错误: %v\n", err)
	} else if success {
		fmt.Printf("Ping 成功，延迟: %v\n", latency)
	} else {
		fmt.Println("Ping 失败")
	}

	// 2. 执行端口测试
	portOpen, err := tester.PortTest("192.168.1.1", 22, 0)
	if err != nil {
		fmt.Printf("端口测试错误: %v\n", err)
	} else if portOpen {
		fmt.Println("端口 22 开放")
	} else {
		fmt.Println("端口 22 关闭")
	}

	// 3. 测试宿主机连通性
	hostIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	hostResults, err := tester.TestHostConnectivity(hostIPs)
	if err != nil {
		fmt.Printf("宿主机测试错误: %v\n", err)
	} else {
		fmt.Printf("测试了 %d 个宿主机\n", len(hostResults))
	}

	// 4. 测试 Pod 连通性
	podIPs := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	podResults, err := tester.TestPodConnectivity(podIPs)
	if err != nil {
		fmt.Printf("Pod 测试错误: %v\n", err)
	} else {
		fmt.Printf("测试了 %d 个 Pod\n", len(podResults))
	}

	// 5. 测试自定义服务
	serviceResult, err := tester.TestServiceConnectivity("kubernetes.default.svc.cluster.local")
	if err != nil {
		fmt.Printf("服务测试错误: %v\n", err)
	} else {
		fmt.Printf("服务测试完成，目标 IP: %s\n", serviceResult.TargetIP)
	}
}
