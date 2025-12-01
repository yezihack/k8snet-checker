package main

import (
	"log"

	"github.com/yezihack/k8snet-checker/pkg/app"
)

func main() {
	// 创建并运行客户端应用
	clientApp, err := app.NewClientApp()
	if err != nil {
		log.Fatalf("创建应用失败: %v", err)
	}
	defer clientApp.Close()

	// 运行应用
	if err := clientApp.Run(); err != nil {
		log.Fatalf("运行应用失败: %v", err)
	}
}
