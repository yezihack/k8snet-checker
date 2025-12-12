package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/api/client"
	"github.com/yezihack/k8snet-checker/pkg/clientserver"
	"github.com/yezihack/k8snet-checker/pkg/collector"
	"github.com/yezihack/k8snet-checker/pkg/config"
	"github.com/yezihack/k8snet-checker/pkg/heartbeat"
	"github.com/yezihack/k8snet-checker/pkg/logger"
	"github.com/yezihack/k8snet-checker/pkg/network"
	"github.com/yezihack/k8snet-checker/pkg/scheduler"

	"go.uber.org/zap"
)

// ClientApp 客户端应用程序
type ClientApp struct {
	ctx               context.Context
	cancel            context.CancelFunc
	heartbeatReporter heartbeat.HeartbeatReporter
	clientServer      clientserver.ClientServer
	testScheduler     *scheduler.TestScheduler
	config            *config.ClientConfig
	logger            *zap.Logger
}

// NewClientApp 创建客户端应用程序实例
func NewClientApp() (*ClientApp, error) {
	// 加载配置
	cfg := config.LoadClientConfig()

	// 初始化日志
	log := logger.NewLogger(cfg.LogLevel)

	log.Info("K8s Network Checker Client 启动中...")

	// 初始化组件
	app, err := initializeComponents(cfg, log)
	if err != nil {
		log.Error("初始化应用失败", zap.Error(err))
		return nil, err
	}

	return app, nil
}

// initializeComponents 初始化所有组件
func initializeComponents(cfg *config.ClientConfig, log *zap.Logger) (*ClientApp, error) {
	log.Info("配置加载成功",
		zap.String("server_url", cfg.ServerURL),
		zap.Duration("heartbeat_interval", cfg.HeartbeatInterval),
		zap.Int("test_port", cfg.TestPort),
		zap.Int("service_port", cfg.ServicePort),
		zap.Int("client_port", cfg.ClientPort),
		zap.String("custom_service_name", cfg.CustomServiceName),
	)

	// 初始化信息收集器
	infoCollector := collector.NewEnvInfoCollector()

	// 收集节点信息
	nodeInfo, err := infoCollector.CollectNodeInfo()
	if err != nil {
		return nil, err
	}

	log.Info("节点信息收集成功",
		zap.String("pod_name", nodeInfo.PodName),
		zap.String("node_ip", nodeInfo.NodeIP),
		zap.String("pod_ip", nodeInfo.PodIP),
		zap.String("namespace", nodeInfo.Namespace),
	)

	// 初始化API客户端
	apiClient := client.NewAPIClient(cfg.ServerURL, nodeInfo.PodIP)

	// 初始化网络测试器
	networkTester := network.NewNetworkTester(
		nodeInfo.PodIP,
		cfg.TestPort,
		6100,            // Pod端口固定为6100
		cfg.ServicePort, // 自定义服务端口
		10,              // 最大并发数为10
		log,
	)

	// 初始化心跳上报器
	heartbeatReporter := heartbeat.NewHeartbeatReporter(infoCollector, apiClient)

	// 初始化客户端HTTP服务器
	clientServer := clientserver.NewClientServer()

	// 初始化测试调度器
	testScheduler := scheduler.NewTestScheduler(apiClient, networkTester, cfg.CustomServiceName, log)

	// 创建主上下文
	ctx, cancel := context.WithCancel(context.Background())

	return &ClientApp{
		ctx:               ctx,
		cancel:            cancel,
		heartbeatReporter: heartbeatReporter,
		clientServer:      clientServer,
		testScheduler:     testScheduler,
		config:            cfg,
		logger:            log,
	}, nil
}

// Run 运行应用程序
func (a *ClientApp) Run() error {
	// 启动所有服务
	if err := a.start(); err != nil {
		a.logger.Error("启动服务失败", zap.Error(err))
		return err
	}

	// 等待退出信号
	a.waitForShutdown()

	// 优雅关闭
	a.shutdown()

	a.logger.Info("K8s Network Checker Client 已停止")
	return nil
}

// start 启动所有服务
func (a *ClientApp) start() error {
	// 启动心跳上报
	if err := a.heartbeatReporter.Start(a.ctx, a.config.HeartbeatInterval); err != nil {
		return err
	}

	// 启动客户端HTTP服务器
	if err := a.clientServer.Start(a.config.ClientPort); err != nil {
		return err
	}

	// 启动测试任务调度器
	go a.testScheduler.Start(a.ctx)

	return nil
}

// shutdown 优雅关闭
func (a *ClientApp) shutdown() {
	a.logger.Info("收到退出信号，开始优雅关闭...")

	// 停止心跳上报
	if err := a.heartbeatReporter.Stop(); err != nil {
		a.logger.Error("停止心跳上报失败", zap.Error(err))
	}

	// 停止客户端HTTP服务器
	if err := a.clientServer.Stop(); err != nil {
		a.logger.Error("停止客户端HTTP服务器失败", zap.Error(err))
	}

	// 取消上下文，停止所有goroutine
	a.cancel()

	// 等待一小段时间，确保所有goroutine完成
	time.Sleep(1 * time.Second)
}

// waitForShutdown 等待退出信号
func (a *ClientApp) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// Close 关闭应用程序资源
func (a *ClientApp) Close() error {
	if a.logger != nil {
		return a.logger.Sync()
	}
	return nil
}
