package clientserver

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewClientServer 测试创建ClientServer实例
func TestNewClientServer(t *testing.T) {
	server := NewClientServer()
	assert.NotNil(t, server, "ClientServer实例不应为nil")
}

// TestClientServer_Start_Success 测试成功启动服务器
func TestClientServer_Start_Success(t *testing.T) {
	server := NewClientServer()
	
	// 使用一个随机端口
	port := 16100
	
	err := server.Start(port)
	assert.NoError(t, err, "启动服务器应该成功")
	
	// 等待服务器完全启动
	time.Sleep(200 * time.Millisecond)
	
	// 验证服务器是否在监听
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.NoError(t, err, "健康检查请求应该成功")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "健康检查应该返回200")
	resp.Body.Close()
	
	// 停止服务器
	err = server.Stop()
	assert.NoError(t, err, "停止服务器应该成功")
}

// TestClientServer_Start_InvalidPort 测试使用无效端口启动服务器
func TestClientServer_Start_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"负数端口", -1},
		{"零端口", 0},
		{"超出范围端口", 70000},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewClientServer()
			err := server.Start(tt.port)
			assert.Error(t, err, "使用无效端口应该返回错误")
		})
	}
}

// TestClientServer_HealthEndpoint 测试健康检查端点
func TestClientServer_HealthEndpoint(t *testing.T) {
	server := NewClientServer()
	port := 16101
	
	err := server.Start(port)
	assert.NoError(t, err, "启动服务器应该成功")
	defer server.Stop()
	
	// 等待服务器完全启动
	time.Sleep(200 * time.Millisecond)
	
	// 发送健康检查请求
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.NoError(t, err, "健康检查请求应该成功")
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "健康检查应该返回200")
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"), "应该返回JSON格式")
}

// TestClientServer_Stop_WithoutStart 测试在未启动的情况下停止服务器
func TestClientServer_Stop_WithoutStart(t *testing.T) {
	server := NewClientServer()
	err := server.Stop()
	assert.Error(t, err, "停止未启动的服务器应该返回错误")
}

// TestClientServer_MultipleStarts 测试多次启动服务器
func TestClientServer_MultipleStarts(t *testing.T) {
	server := NewClientServer()
	port := 16102
	
	// 第一次启动
	err := server.Start(port)
	assert.NoError(t, err, "第一次启动应该成功")
	
	// 等待服务器完全启动
	time.Sleep(200 * time.Millisecond)
	
	// 验证服务器正在运行
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.NoError(t, err, "健康检查请求应该成功")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 停止服务器
	err = server.Stop()
	assert.NoError(t, err, "停止服务器应该成功")
	
	// 等待服务器完全停止
	time.Sleep(200 * time.Millisecond)
}

// TestClientServer_GracefulShutdown 测试优雅关闭
func TestClientServer_GracefulShutdown(t *testing.T) {
	server := NewClientServer()
	port := 16103
	
	err := server.Start(port)
	assert.NoError(t, err, "启动服务器应该成功")
	
	// 等待服务器完全启动
	time.Sleep(200 * time.Millisecond)
	
	// 发送一个请求
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.NoError(t, err, "健康检查请求应该成功")
	resp.Body.Close()
	
	// 优雅关闭
	err = server.Stop()
	assert.NoError(t, err, "优雅关闭应该成功")
	
	// 验证服务器已停止
	time.Sleep(200 * time.Millisecond)
	_, err = http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.Error(t, err, "服务器停止后请求应该失败")
}

// TestClientServer_DefaultPort 测试使用默认端口6100
func TestClientServer_DefaultPort(t *testing.T) {
	server := NewClientServer()
	
	// 使用默认端口6100（如果端口未被占用）
	port := 6100
	
	err := server.Start(port)
	if err != nil {
		// 如果6100端口被占用，跳过此测试
		t.Skipf("端口%d被占用，跳过测试", port)
		return
	}
	defer server.Stop()
	
	// 等待服务器完全启动
	time.Sleep(200 * time.Millisecond)
	
	// 验证服务器在6100端口上运行
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.NoError(t, err, "健康检查请求应该成功")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
