package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/api/server"
	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/report"
	"github.com/yezihack/k8snet-checker/pkg/result"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("K8s Network Checker Server 启动中...")

	// 读取环境变量配置
	config := loadConfig()

	// 设置日志级别
	setupLogging(config.LogLevel)

	log.Printf("配置信息: CACHE_KEY_SECOND=%d, LOG_LEVEL=%s, HTTP_PORT=%s, REPORT_INTERVAL=%d秒",
		config.CacheKeySecond, config.LogLevel, config.HTTPPort, int(config.ReportInterval.Seconds()))

	// 初始化缓存管理器
	log.Println("初始化缓存管理器...")
	cacheManager := cache.NewCacheManager()

	// 初始化客户端管理器
	log.Println("初始化客户端管理器...")
	clientManager := client.NewClientManager(cacheManager)

	// 初始化测试结果管理器
	log.Println("初始化测试结果管理器...")
	resultManager := result.NewTestResultManager(cacheManager)

	// 初始化报告生成器
	log.Println("初始化报告生成器...")
	reportGenerator := report.NewReportGenerator(clientManager, resultManager)

	// 初始化HTTP服务器
	log.Println("初始化HTTP服务器...")
	apiServer := server.NewAPIServer(clientManager, resultManager)

	// 创建context用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动报告生成器goroutine
	log.Printf("启动报告生成器，间隔: %v", config.ReportInterval)
	if err := reportGenerator.Start(ctx, config.ReportInterval); err != nil {
		log.Fatalf("启动报告生成器失败: %v", err)
	}

	// 在独立goroutine中启动HTTP服务器
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP服务器启动在端口: %s", config.HTTPPort)
		if err := apiServer.Start(config.HTTPPort); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP服务器启动失败: %w", err)
		}
	}()

	// 等待中断信号以优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待关闭信号或服务器错误
	select {
	case err := <-serverErrors:
		log.Fatalf("服务器错误: %v", err)
	case sig := <-quit:
		log.Printf("收到信号: %v，开始优雅关闭...", sig)
	}

	// 优雅关闭
	log.Println("正在关闭服务器...")

	// 取消context，停止报告生成器
	cancel()

	// 给服务器一些时间来完成正在处理的请求
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// 等待关闭完成或超时
	<-shutdownCtx.Done()

	log.Println("服务器已优雅关闭")
}

// Config 存储服务器配置
type Config struct {
	CacheKeySecond int           // 缓存过期时间（秒）
	LogLevel       string        // 日志级别
	HTTPPort       string        // HTTP服务端口
	ReportInterval time.Duration // 报告生成间隔
}

// loadConfig 从环境变量加载配置
func loadConfig() Config {
	config := Config{
		CacheKeySecond: 15,                // 默认15秒
		LogLevel:       "info",            // 默认info级别
		HTTPPort:       "8080",            // 默认8080端口
		ReportInterval: 300 * time.Second, // 默认300秒（5分钟）
	}

	// 读取CACHE_KEY_SECOND
	if cacheKeySecond := os.Getenv("CACHE_KEY_SECOND"); cacheKeySecond != "" {
		if val, err := strconv.Atoi(cacheKeySecond); err == nil && val > 0 {
			config.CacheKeySecond = val
		} else {
			log.Printf("警告: CACHE_KEY_SECOND值无效(%s)，使用默认值15", cacheKeySecond)
		}
	}

	// 读取LOG_LEVEL
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// 读取HTTP_PORT
	if httpPort := os.Getenv("HTTP_PORT"); httpPort != "" {
		config.HTTPPort = httpPort
	}

	// 读取REPORT_INTERVAL
	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		if val, err := strconv.Atoi(reportInterval); err == nil && val > 0 {
			config.ReportInterval = time.Duration(val) * time.Second
		} else {
			log.Printf("警告: REPORT_INTERVAL值无效(%s)，使用默认值300秒", reportInterval)
		}
	}

	return config
}

// setupLogging 设置日志级别
func setupLogging(logLevel string) {
	// Go标准库的log包不支持日志级别过滤
	// 这里只是记录日志级别，实际的日志级别控制可以通过第三方库实现
	// 例如: logrus, zap等
	switch logLevel {
	case "debug":
		log.Println("日志级别设置为: DEBUG")
	case "info":
		log.Println("日志级别设置为: INFO")
	case "warn":
		log.Println("日志级别设置为: WARN")
	case "error":
		log.Println("日志级别设置为: ERROR")
	default:
		log.Printf("未知的日志级别: %s，使用默认级别INFO", logLevel)
	}
}
