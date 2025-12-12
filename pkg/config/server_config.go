package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	CacheKeySecond int           // 缓存过期时间（秒）
	LogLevel       string        // 日志级别
	HTTPPort       string        // HTTP服务端口
	ReportInterval time.Duration // 报告生成间隔
}

// LoadServerConfig 从环境变量加载服务器配置
func LoadServerConfig() *ServerConfig {
	config := &ServerConfig{
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
