# 网络测试模块

## 概述

网络测试模块提供了完整的网络连通性测试功能，包括 ping 测试、端口测试和并发测试能力。

## 功能特性

### 1. Ping 测试
- 支持 Windows 和 Linux 系统
- 可配置 ping 次数（默认 3 次）
- 自动计算平均延迟
- 10 秒超时保护

### 2. 端口测试
- TCP 连接测试
- 可配置超时时间（默认 5 秒）
- 支持任意端口测试

### 3. 宿主机连通性测试
- 批量测试多个宿主机 IP
- 同时执行 ping 和端口测试（默认端口 22）
- 并发测试，提高效率

### 4. Pod 连通性测试
- 批量测试多个 Pod IP
- 同时执行 ping 和端口测试（默认端口 6100）
- 并发测试，提高效率

### 5. 自定义服务测试
- DNS 解析验证
- 连通性测试
- 可配置服务端口测试

### 6. 并发控制
- 使用 semaphore 限制并发数
- 默认最大 10 个并发 goroutine
- 避免网络拥塞

### 7. 超时处理
- Ping 测试：10 秒超时
- 端口测试：5 秒超时
- DNS 解析：5 秒超时

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/yezihack/k8snet-checker/pkg/network"
    "go.uber.org/zap"
)

func main() {
    // 创建日志记录器
    logger, _ := zap.NewDevelopment()
    
    // 创建 NetworkTester
    // 参数：源IP, 宿主机端口, Pod端口, 服务端口, 最大并发数, 日志记录器
    tester := network.NewNetworkTester("192.168.1.100", 22, 6100, 80, 10, logger)
    
    // 测试宿主机连通性
    hostIPs := []string{"192.168.1.1", "192.168.1.2"}
    results, err := tester.TestHostConnectivity(hostIPs)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    
    // 打印结果
    for _, result := range results {
        fmt.Printf("目标: %s, Ping: %s, 端口: %s\n",
            result.TargetIP,
            result.PingStatus,
            result.PortStatus[22],
        )
    }
}
```

## 接口定义

```go
type NetworkTester interface {
    // PingTest 执行 ping 测试
    PingTest(targetIP string, count int) (bool, time.Duration, error)
    
    // PortTest 执行端口测试
    PortTest(targetIP string, port int, timeout time.Duration) (bool, error)
    
    // TestHostConnectivity 测试宿主机连通性
    TestHostConnectivity(hostIPs []string) ([]models.ConnectivityResult, error)
    
    // TestPodConnectivity 测试 Pod 连通性
    TestPodConnectivity(podIPs []string) ([]models.ConnectivityResult, error)
    
    // TestServiceConnectivity 测试自定义服务
    TestServiceConnectivity(serviceName string) (*models.ConnectivityResult, error)
}
```

## 配置参数

- **sourceIP**: 源 IP 地址
- **hostPort**: 宿主机测试端口（默认 22）
- **podPort**: Pod 测试端口（默认 6100）
- **servicePort**: 自定义服务测试端口（默认 80）
- **maxWorkers**: 最大并发数（默认 10）

## 测试覆盖

模块包含完整的单元测试：
- `TestNewNetworkTester`: 测试实例创建
- `TestPingTest`: 测试 ping 功能
- `TestPortTest`: 测试端口检测
- `TestTestHostConnectivity`: 测试宿主机连通性
- `TestTestPodConnectivity`: 测试 Pod 连通性
- `TestTestServiceConnectivity`: 测试服务连通性
- `TestConcurrentTesting`: 测试并发功能

## 日志输出

模块使用 zap 日志库，提供以下级别的日志：
- **DEBUG**: 详细的测试过程信息
- **INFO**: 测试开始和完成信息
- **WARN**: DNS 解析失败等警告

## 错误处理

- Ping 失败不返回错误，只返回失败状态
- 端口不可达不返回错误，只返回失败状态
- DNS 解析失败不返回错误，只返回失败状态
- 只有参数错误（如空服务名称）才返回错误

## 性能考虑

- 使用并发测试提高效率
- 限制并发数避免资源耗尽
- 跳过源 IP 避免自测
- 合理的超时时间避免长时间等待
