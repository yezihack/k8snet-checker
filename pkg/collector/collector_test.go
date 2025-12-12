package collector

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvInfoCollector_CollectNodeInfo_Success(t *testing.T) {
	// 设置环境变量
	os.Setenv("NODE_IP", "192.168.1.100")
	os.Setenv("POD_IP", "10.244.0.5")
	os.Setenv("POD_NAME", "test-pod-abc")
	os.Setenv("NAMESPACE", "default")
	defer func() {
		os.Unsetenv("NODE_IP")
		os.Unsetenv("POD_IP")
		os.Unsetenv("POD_NAME")
		os.Unsetenv("NAMESPACE")
	}()

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证没有错误
	assert.NoError(t, err)
	assert.NotNil(t, nodeInfo)

	// 验证所有字段都正确设置
	assert.Equal(t, "192.168.1.100", nodeInfo.NodeIP)
	assert.Equal(t, "10.244.0.5", nodeInfo.PodIP)
	assert.Equal(t, "test-pod-abc", nodeInfo.PodName)
	assert.Equal(t, "default", nodeInfo.Namespace)

	// 验证时间戳在合理范围内（最近1秒内）
	assert.WithinDuration(t, time.Now(), nodeInfo.Timestamp, time.Second)
}

func TestEnvInfoCollector_CollectNodeInfo_MissingNodeIP(t *testing.T) {
	// 只设置部分环境变量，缺少NODE_IP
	os.Setenv("POD_IP", "10.244.0.5")
	os.Setenv("POD_NAME", "test-pod-abc")
	os.Setenv("NAMESPACE", "default")
	defer func() {
		os.Unsetenv("POD_IP")
		os.Unsetenv("POD_NAME")
		os.Unsetenv("NAMESPACE")
	}()

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证返回错误
	assert.Error(t, err)
	assert.Nil(t, nodeInfo)
	assert.Contains(t, err.Error(), "NODE_IP")
}

func TestEnvInfoCollector_CollectNodeInfo_MissingPodIP(t *testing.T) {
	// 只设置部分环境变量，缺少POD_IP
	os.Setenv("NODE_IP", "192.168.1.100")
	os.Setenv("POD_NAME", "test-pod-abc")
	os.Setenv("NAMESPACE", "default")
	defer func() {
		os.Unsetenv("NODE_IP")
		os.Unsetenv("POD_NAME")
		os.Unsetenv("NAMESPACE")
	}()

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证返回错误
	assert.Error(t, err)
	assert.Nil(t, nodeInfo)
	assert.Contains(t, err.Error(), "POD_IP")
}

func TestEnvInfoCollector_CollectNodeInfo_MissingPodName(t *testing.T) {
	// 只设置部分环境变量，缺少POD_NAME
	os.Setenv("NODE_IP", "192.168.1.100")
	os.Setenv("POD_IP", "10.244.0.5")
	os.Setenv("NAMESPACE", "default")
	defer func() {
		os.Unsetenv("NODE_IP")
		os.Unsetenv("POD_IP")
		os.Unsetenv("NAMESPACE")
	}()

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证返回错误
	assert.Error(t, err)
	assert.Nil(t, nodeInfo)
	assert.Contains(t, err.Error(), "POD_NAME")
}

func TestEnvInfoCollector_CollectNodeInfo_MissingNamespace(t *testing.T) {
	// 只设置部分环境变量，缺少NAMESPACE
	os.Setenv("NODE_IP", "192.168.1.100")
	os.Setenv("POD_IP", "10.244.0.5")
	os.Setenv("POD_NAME", "test-pod-abc")
	defer func() {
		os.Unsetenv("NODE_IP")
		os.Unsetenv("POD_IP")
		os.Unsetenv("POD_NAME")
	}()

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证返回错误
	assert.Error(t, err)
	assert.Nil(t, nodeInfo)
	assert.Contains(t, err.Error(), "NAMESPACE")
}

func TestEnvInfoCollector_CollectNodeInfo_AllMissing(t *testing.T) {
	// 确保所有环境变量都未设置
	os.Unsetenv("NODE_IP")
	os.Unsetenv("POD_IP")
	os.Unsetenv("POD_NAME")
	os.Unsetenv("NAMESPACE")

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证返回错误
	assert.Error(t, err)
	assert.Nil(t, nodeInfo)
	// 验证错误消息包含所有缺失的变量
	assert.Contains(t, err.Error(), "NODE_IP")
	assert.Contains(t, err.Error(), "POD_IP")
	assert.Contains(t, err.Error(), "POD_NAME")
	assert.Contains(t, err.Error(), "NAMESPACE")
}

func TestEnvInfoCollector_CollectNodeInfo_MultipleMissing(t *testing.T) {
	// 只设置一个环境变量
	os.Setenv("NODE_IP", "192.168.1.100")
	defer os.Unsetenv("NODE_IP")

	collector := NewEnvInfoCollector()
	nodeInfo, err := collector.CollectNodeInfo()

	// 验证返回错误
	assert.Error(t, err)
	assert.Nil(t, nodeInfo)
	// 验证错误消息包含所有缺失的变量
	assert.Contains(t, err.Error(), "POD_IP")
	assert.Contains(t, err.Error(), "POD_NAME")
	assert.Contains(t, err.Error(), "NAMESPACE")
}
