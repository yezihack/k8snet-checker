package server

import (
	"log"
	"os"

	"github.com/yezihack/k8snet-checker/pkg/client"
	"github.com/yezihack/k8snet-checker/pkg/result"

	"github.com/gin-gonic/gin"
)

// APIServer 定义HTTP API服务器接口
type APIServer interface {
	Start(port string) error
	Stop() error
}

// apiServerImpl 是APIServer的实现
type apiServerImpl struct {
	router  *gin.Engine
	handler *Handler
}

// NewAPIServer 创建一个新的APIServer实例
func NewAPIServer(clientManager client.ClientManager, resultManager result.TestResultManager) APIServer {
	// 根据LOG_LEVEL设置Gin模式
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	handler := NewHandler(clientManager, resultManager)

	// 注册路由
	RegisterRoutes(router, handler)

	return &apiServerImpl{
		router:  router,
		handler: handler,
	}
}

// Start 启动HTTP服务器
func (s *apiServerImpl) Start(port string) error {
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP服务器启动在端口: %s", port)
	return s.router.Run(":" + port)
}

// Stop 停止HTTP服务器
func (s *apiServerImpl) Stop() error {
	// Gin框架没有内置的优雅关闭方法
	// 这里返回nil，实际的优雅关闭会在main.go中通过context处理
	return nil
}
