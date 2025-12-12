package report

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/models"
	"github.com/yezihack/k8snet-checker/pkg/result"
)

// ReportGenerator 定义报告生成器接口
type ReportGenerator interface {
	// Start 启动报告生成器，定期生成报告
	Start(ctx context.Context, interval time.Duration) error

	// Stop 停止报告生成器
	Stop() error

	// GenerateReport 生成网络连通性报告
	GenerateReport() (*models.NetworkReport, error)
}

// reportGeneratorImpl 是ReportGenerator的实现
type reportGeneratorImpl struct {
	clientManager client.ClientManager
	resultManager result.TestResultManager
	stopChan      chan struct{}
	running       bool
}

// NewReportGenerator 创建一个新的ReportGenerator实例
func NewReportGenerator(clientManager client.ClientManager, resultManager result.TestResultManager) ReportGenerator {
	return &reportGeneratorImpl{
		clientManager: clientManager,
		resultManager: resultManager,
		stopChan:      make(chan struct{}),
		running:       false,
	}
}

// Start 启动报告生成器
// 启动独立goroutine定期生成报告
func (rg *reportGeneratorImpl) Start(ctx context.Context, interval time.Duration) error {
	if rg.running {
		return fmt.Errorf("报告生成器已经在运行")
	}

	rg.running = true

	// 启动独立goroutine
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("报告生成器已启动，间隔: %v", interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("报告生成器收到context取消信号，正在停止...")
				rg.running = false
				return
			case <-rg.stopChan:
				log.Println("报告生成器收到停止信号，正在停止...")
				rg.running = false
				return
			case <-ticker.C:
				// 生成并输出报告
				report, err := rg.GenerateReport()
				if err != nil {
					log.Printf("生成报告失败: %v", err)
					continue
				}

				// 格式化输出到控制台
				rg.printReport(report)
			}
		}
	}()

	return nil
}

// Stop 停止报告生成器
func (rg *reportGeneratorImpl) Stop() error {
	if !rg.running {
		return fmt.Errorf("报告生成器未运行")
	}

	close(rg.stopChan)
	return nil
}

// GenerateReport 生成网络连通性报告
// 聚合所有数据生成NetworkReport
func (rg *reportGeneratorImpl) GenerateReport() (*models.NetworkReport, error) {
	report := &models.NetworkReport{
		Timestamp: time.Now(),
	}

	// 获取活跃客户端数量
	activeCount, err := rg.clientManager.GetActiveClientCount()
	if err != nil {
		log.Printf("获取活跃客户端数量失败: %v", err)
		activeCount = 0
	}
	report.ActiveClientCount = activeCount

	// 获取所有宿主机IP列表
	hostIPs, err := rg.clientManager.GetAllHostIPs()
	if err != nil {
		log.Printf("获取宿主机IP列表失败: %v", err)
		hostIPs = []string{}
	}
	report.HostIPs = hostIPs

	// 获取所有Pod IP列表
	podIPs, err := rg.clientManager.GetAllPodIPs()
	if err != nil {
		log.Printf("获取Pod IP列表失败: %v", err)
		podIPs = []string{}
	}
	report.PodIPs = podIPs

	// 获取宿主机测试结果并生成统计
	hostTestResults, err := rg.resultManager.GetHostTestResults()
	if err != nil {
		log.Printf("获取宿主机测试结果失败: %v", err)
		hostTestResults = make(models.HostTestResults)
	}
	report.HostTestSummary = rg.calculateTestSummary(hostTestResults)

	// 获取Pod测试结果并生成统计
	podTestResults, err := rg.resultManager.GetPodTestResults()
	if err != nil {
		log.Printf("获取Pod测试结果失败: %v", err)
		podTestResults = make(models.PodTestResults)
	}
	report.PodTestSummary = rg.calculateTestSummary(podTestResults)

	// 获取自定义服务测试结果并生成统计
	serviceTestResults, err := rg.resultManager.GetServiceTestResults()
	if err != nil {
		log.Printf("获取自定义服务测试结果失败: %v", err)
		serviceTestResults = make(models.ServiceTestResults)
	}
	report.ServiceTestSummary = rg.calculateServiceTestSummary(serviceTestResults)

	return report, nil
}

// calculateTestSummary 计算测试统计信息
// 适用于HostTestResults和PodTestResults
func (rg *reportGeneratorImpl) calculateTestSummary(results map[string]map[string]models.TestStatus) models.TestSummary {
	summary := models.TestSummary{
		TotalTests:        0,
		SuccessfulTests:   0,
		FailedTests:       0,
		SuccessRate:       0.0,
		TotalTestDuration: 0,
		AvgTestDuration:   0,
	}

	// 遍历所有测试结果
	for _, targets := range results {
		for _, status := range targets {
			summary.TotalTests++
			summary.TotalTestDuration += status.TestDuration

			// 判断测试是否成功：ping可达且端口开放
			if status.Ping == "reachable" && status.PortStatus == "open" {
				summary.SuccessfulTests++
			} else {
				summary.FailedTests++
			}
		}
	}

	// 计算成功率
	if summary.TotalTests > 0 {
		summary.SuccessRate = float64(summary.SuccessfulTests) / float64(summary.TotalTests) * 100
		summary.AvgTestDuration = summary.TotalTestDuration / time.Duration(summary.TotalTests)
	}

	return summary
}

// calculateServiceTestSummary 计算自定义服务测试统计信息
func (rg *reportGeneratorImpl) calculateServiceTestSummary(results models.ServiceTestResults) models.ServiceTestSummary {
	summary := models.ServiceTestSummary{
		ServiceName:     "",
		TotalTests:      0,
		SuccessfulTests: 0,
		FailedTests:     0,
		SuccessRate:     0.0,
	}

	// 遍历所有服务测试结果
	for _, result := range results {
		summary.TotalTests++

		// 从第一个结果中获取服务名称（假设所有测试针对同一服务）
		if summary.ServiceName == "" && result.TargetIP != "" {
			summary.ServiceName = result.TargetIP // 使用TargetIP作为服务名称
		}

		// 判断测试是否成功：ping可达
		if result.PingStatus == "reachable" {
			summary.SuccessfulTests++
		} else {
			summary.FailedTests++
		}
	}

	// 计算成功率
	if summary.TotalTests > 0 {
		summary.SuccessRate = float64(summary.SuccessfulTests) / float64(summary.TotalTests) * 100
	}

	return summary
}

// printReport 格式化输出报告到控制台
func (rg *reportGeneratorImpl) printReport(report *models.NetworkReport) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("网络连通性报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("生成时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 80))

	// 活跃客户端信息
	fmt.Printf("活跃客户端数量: %d\n", report.ActiveClientCount)
	fmt.Println()

	// 宿主机IP列表
	fmt.Printf("宿主机IP列表 (共%d个):\n", len(report.HostIPs))
	if len(report.HostIPs) > 0 {
		for i, ip := range report.HostIPs {
			fmt.Printf("  %d. %s\n", i+1, ip)
		}
	} else {
		fmt.Println("  无")
	}
	fmt.Println()

	// Pod IP列表
	fmt.Printf("Pod IP列表 (共%d个):\n", len(report.PodIPs))
	if len(report.PodIPs) > 0 {
		for i, ip := range report.PodIPs {
			fmt.Printf("  %d. %s\n", i+1, ip)
		}
	} else {
		fmt.Println("  无")
	}
	fmt.Println()

	// 宿主机测试统计
	fmt.Println("宿主机连通性测试统计:")
	fmt.Printf("  总测试数: %d\n", report.HostTestSummary.TotalTests)
	fmt.Printf("  成功: %d\n", report.HostTestSummary.SuccessfulTests)
	fmt.Printf("  失败: %d\n", report.HostTestSummary.FailedTests)
	fmt.Printf("  成功率: %.2f%%\n", report.HostTestSummary.SuccessRate)
	if report.HostTestSummary.TotalTests > 0 {
		fmt.Printf("  平均耗时: %v\n", report.HostTestSummary.AvgTestDuration)
		fmt.Printf("  总耗时: %v\n", report.HostTestSummary.TotalTestDuration)
	}
	fmt.Println()

	// Pod测试统计
	fmt.Println("Pod连通性测试统计:")
	fmt.Printf("  总测试数: %d\n", report.PodTestSummary.TotalTests)
	fmt.Printf("  成功: %d\n", report.PodTestSummary.SuccessfulTests)
	fmt.Printf("  失败: %d\n", report.PodTestSummary.FailedTests)
	fmt.Printf("  成功率: %.2f%%\n", report.PodTestSummary.SuccessRate)
	if report.PodTestSummary.TotalTests > 0 {
		fmt.Printf("  平均耗时: %v\n", report.PodTestSummary.AvgTestDuration)
		fmt.Printf("  总耗时: %v\n", report.PodTestSummary.TotalTestDuration)
	}
	fmt.Println()

	// 自定义服务测试统计
	if report.ServiceTestSummary.TotalTests > 0 {
		fmt.Println("自定义服务连通性测试统计:")
		fmt.Printf("  服务名称: %s\n", report.ServiceTestSummary.ServiceName)
		fmt.Printf("  总测试数: %d\n", report.ServiceTestSummary.TotalTests)
		fmt.Printf("  成功: %d\n", report.ServiceTestSummary.SuccessfulTests)
		fmt.Printf("  失败: %d\n", report.ServiceTestSummary.FailedTests)
		fmt.Printf("  成功率: %.2f%%\n", report.ServiceTestSummary.SuccessRate)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
}

// GetReportIntervalFromEnv 从环境变量获取报告生成间隔
// 默认300秒（5分钟）
func GetReportIntervalFromEnv() time.Duration {
	intervalStr := os.Getenv("REPORT_INTERVAL")
	if intervalStr == "" {
		return 300 * time.Second // 默认5分钟
	}

	intervalSec, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("解析REPORT_INTERVAL失败，使用默认值300秒: %v", err)
		return 300 * time.Second
	}

	if intervalSec <= 0 {
		log.Printf("REPORT_INTERVAL值无效(%d)，使用默认值300秒", intervalSec)
		return 300 * time.Second
	}

	return time.Duration(intervalSec) * time.Second
}
