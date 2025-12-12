package collector

import (
	"fmt"
	"os"
	"time"

	"github.com/yezihack/k8snet-checker/pkg/models"
)

// InfoCollector 定义了收集节点信息的接口
type InfoCollector interface {
	CollectNodeInfo() (*models.NodeInfo, error)
}

// EnvInfoCollector 从环境变量收集节点信息的实现
type EnvInfoCollector struct{}

// NewEnvInfoCollector 创建一个新的环境变量信息收集器
func NewEnvInfoCollector() *EnvInfoCollector {
	return &EnvInfoCollector{}
}

// CollectNodeInfo 从环境变量读取节点信息并创建NodeInfo结构
// 需要的环境变量: NODE_IP, POD_IP, POD_NAME, NAMESPACE
func (c *EnvInfoCollector) CollectNodeInfo() (*models.NodeInfo, error) {
	// 读取必需的环境变量
	nodeIP := os.Getenv("NODE_IP")
	podIP := os.Getenv("POD_IP")
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("NAMESPACE")

	// 验证所有必需的环境变量是否存在
	missingVars := []string{}
	if nodeIP == "" {
		missingVars = append(missingVars, "NODE_IP")
	}
	if podIP == "" {
		missingVars = append(missingVars, "POD_IP")
	}
	if podName == "" {
		missingVars = append(missingVars, "POD_NAME")
	}
	if namespace == "" {
		missingVars = append(missingVars, "NAMESPACE")
	}

	// 如果有缺失的环境变量，返回错误
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("缺少必需的环境变量: %v", missingVars)
	}

	// 创建并返回NodeInfo结构
	nodeInfo := &models.NodeInfo{
		Namespace: namespace,
		NodeIP:    nodeIP,
		PodIP:     podIP,
		PodName:   podName,
		Timestamp: time.Now(),
	}

	return nodeInfo, nil
}
