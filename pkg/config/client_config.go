package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	ServerURL         string
	HeartbeatInterval time.Duration
	TestPort          int
	CustomServiceName string
	ServicePort       int
	ClientPort        int
	LogLevel          string
}

// LoadClientConfig 从环境变量加载客户端配置
func LoadClientConfig() *ClientConfig {
	return &ClientConfig{
		ServerURL:         getEnv("SERVER_URL", "http://k8snet-checker-server.kube-system.svc.cluster.local:8080"),
		HeartbeatInterval: getDurationEnv("HEARTBEAT_INTERVAL", 5) * time.Second,
		TestPort:          getIntEnv("TEST_PORT", 22),
		CustomServiceName: getEnv("CUSTOM_SERVICE_NAME", ""),
		ServicePort:       getIntEnv("CUSTOM_SERVICE_PORT", 80),
		ClientPort:        getIntEnv("CLIENT_PORT", 6100),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv 获取整数类型的环境变量
func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("警告: 无法解析环境变量 %s='%s'，使用默认值 %d: %v",
			key, valueStr, defaultValue, err)
		return defaultValue
	}

	return value
}

// getDurationEnv 获取时间间隔类型的环境变量（秒）
func getDurationEnv(key string, defaultValue int) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return time.Duration(defaultValue)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("警告: 无法解析环境变量 %s='%s'，使用默认值 %d秒: %v",
			key, valueStr, defaultValue, err)
		return time.Duration(defaultValue)
	}

	if value <= 0 {
		log.Printf("警告: 环境变量 %s 值无效 (%d)，使用默认值 %d秒",
			key, value, defaultValue)
		return time.Duration(defaultValue)
	}

	return time.Duration(value)
}
