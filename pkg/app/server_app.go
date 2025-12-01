package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/api/server"
	"github.com/yezihack/k8snet-checker/pkg/cache"
	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/config"
	"github.com/yezihack/k8snet-checker/pkg/report"
	"github.com/yezihack/k8snet-checker/pkg/result"
)

// ServerApp 服务器应用程序
type ServerApp struct {
	ctx             context.Context
	cancel          context.CancelFunc
	apiServer       server.APIServer
	reportGenerator report.ReportGenerator
	config          *config.ServerConfig
}

// NewServerApp 创建服务器应用程序实例
func NewServerApp() (*ServerApp, error) {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("K8s Network Checker Server 启动中...")

	// 加载配置
	cfg := config.LoadServerConfig()

	// 设置日志级别
	setupLogging(cfg.LogLevel)

	log.Printf("配置信息: CACHE_KEY_SECOND=%d, LOG_LEVEL=%s, HTTP_PORT=%s, REPORT_INTERVAL=%d秒",
		cfg.CacheKeySecond, cfg.LogLevel, cfg.HTTPPort, int(cfg.ReportInterval.Seconds()))

	// 初始化组件
	app, err := initializeServerComponents(cfg)
	if err != nil {
		log.Printf("初始化应用失败: %v", err)
		return nil, err
	}

	return app, nil
}

// initializeServerComponents 初始化所有服务器组件
func initializeServerComponents(cfg *config.ServerConfig) (*ServerApp, error) {
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

	// 创建主上下文
	ctx, cancel := context.WithCancel(context.Background())

	return &ServerApp{
		ctx:             ctx,
		cancel:          cancel,
		apiServer:       apiServer,
		reportGenerator: reportGenerator,
		config:          cfg,
	}, nil
}

// Run 运行应用程序
func (a *ServerApp) Run() error {
	// 启动所有服务
	serverErrors := make(chan error, 1)
	if err := a.start(serverErrors); err != nil {
		return err
	}

	// 等待退出信号
	a.waitForShutdown(serverErrors)

	// 优雅关闭
	a.shutdown()

	log.Println("服务器已优雅关闭")
	return nil
}

// start 启动所有服务
func (a *ServerApp) start(serverErrors chan error) error {
	// 启动报告生成器
	log.Printf("启动报告生成器，间隔: %v", a.config.ReportInterval)
	if err := a.reportGenerator.Start(a.ctx, a.config.ReportInterval); err != nil {
		return fmt.Errorf("启动报告生成器失败: %w", err)
	}

	// 在独立goroutine中启动HTTP服务器
	go func() {
		log.Printf("HTTP服务器启动在端口: %s", a.config.HTTPPort)
		if err := a.apiServer.Start(a.config.HTTPPort); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP服务器启动失败: %w", err)
		}
	}()

	return nil
}

// shutdown 优雅关闭
func (a *ServerApp) shutdown() {
	log.Println("正在关闭服务器...")

	// 取消context，停止报告生成器
	a.cancel()

	// 给服务器一些时间来完成正在处理的请求
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// 等待关闭完成或超时
	<-shutdownCtx.Done()
}

// waitForShutdown 等待退出信号或服务器错误
func (a *ServerApp) waitForShutdown(serverErrors chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待关闭信号或服务器错误
	select {
	case err := <-serverErrors:
		log.Fatalf("服务器错误: %v", err)
	case sig := <-quit:
		log.Printf("收到信号: %v，开始优雅关闭...", sig)
	}
}

// Close 关闭应用程序资源
func (a *ServerApp) Close() error {
	// 服务器端暂无需要关闭的资源
	return nil
}

// setupLogging 设置日志级别
func setupLogging(logLevel string) {
	// Go标准库的log包不支持日志级别过滤
	// 这里只是记录日志级别，实际的日志级别控制可以通过第三方库实现
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
