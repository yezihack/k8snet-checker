package main

import (
	"log"

	"github.com/yezihack/k8snet-checker/pkg/app"
)

func main() {
	// 创建并运行服务器应用
	serverApp, err := app.NewServerApp()
	if err != nil {
		log.Fatalf("创建应用失败: %v", err)
	}
	defer serverApp.Close()

	// 运行应用
	if err := serverApp.Run(); err != nil {
		log.Fatalf("运行应用失败: %v", err)
	}
}
