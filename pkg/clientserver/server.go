package clientserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ClientServer 定义客户端HTTP服务器接口
type ClientServer interface {
	// Start 启动HTTP服务器
	Start(port int) error
	
	// Stop 停止HTTP服务器
	Stop() error
}

// clientServerImpl 是ClientServer的实现
type clientServerImpl struct {
	server *http.Server
	port   int
}

// NewClientServer 创建一个新的ClientServer实例
func NewClientServer() ClientServer {
	return &clientServerImpl{}
}

// Start 启动HTTP服务器
// 在指定端口上启动HTTP服务器，实现健康检查端点
func (cs *clientServerImpl) Start(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("无效的端口号: %d", port)
	}
	
	cs.port = port
	
	// 设置Gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	
	// 创建Gin路由器
	router := gin.New()
	
	// 添加恢复中间件，防止panic导致服务器崩溃
	router.Use(gin.Recovery())
	
	// 添加日志中间件
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		
		c.Next()
		
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		log.Printf("客户端HTTP请求: method=%s, path=%s, status=%d, latency=%v",
			c.Request.Method, path, statusCode, latency)
	})
	
	// 注册健康检查端点
	router.GET("/health", cs.healthHandler)
	
	// 创建HTTP服务器
	cs.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	
	// 在独立的goroutine中启动服务器
	go func() {
		log.Printf("客户端HTTP服务器启动: 端口=%d", port)
		if err := cs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("客户端HTTP服务器启动失败: %v", err)
		}
	}()
	
	// 等待一小段时间，确保服务器启动
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// Stop 停止HTTP服务器
// 优雅地关闭HTTP服务器，等待现有请求完成
func (cs *clientServerImpl) Stop() error {
	if cs.server == nil {
		return fmt.Errorf("服务器未启动")
	}
	
	log.Printf("正在停止客户端HTTP服务器...")
	
	// 创建5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 优雅关闭服务器
	if err := cs.server.Shutdown(ctx); err != nil {
		log.Printf("客户端HTTP服务器关闭失败: %v", err)
		return fmt.Errorf("服务器关闭失败: %w", err)
	}
	
	log.Printf("客户端HTTP服务器已停止")
	return nil
}

// healthHandler 处理健康检查请求
// 返回200 OK表示服务器正常运行
func (cs *clientServerImpl) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
