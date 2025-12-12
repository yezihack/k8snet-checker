package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration 自定义 Duration 类型，用于人性化的 JSON 序列化
type Duration time.Duration

// MarshalJSON 实现 JSON 序列化，输出人性化的时间格式
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON 实现 JSON 反序列化
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return fmt.Errorf("invalid duration")
	}
}

// String 返回人性化的时间字符串
func (d Duration) String() string {
	duration := time.Duration(d)

	// 小于 1 微秒
	if duration < time.Microsecond {
		return fmt.Sprintf("%dns", duration.Nanoseconds())
	}

	// 小于 1 毫秒
	if duration < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(duration.Nanoseconds())/1000.0)
	}

	// 小于 1 秒
	if duration < time.Second {
		return fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1000000.0)
	}

	// 小于 1 分钟
	if duration < time.Minute {
		return fmt.Sprintf("%.2fs", duration.Seconds())
	}

	// 大于等于 1 分钟
	return duration.String()
}

// NodeInfo represents the information about a Kubernetes node and pod
type NodeInfo struct {
	Namespace string    `json:"namespace"`
	NodeIP    string    `json:"node_ip"`
	PodIP     string    `json:"pod_ip"`
	PodName   string    `json:"pod_name"`
	Timestamp time.Time `json:"timestamp"`
}

// ConnectivityResult represents the result of a network connectivity test
type ConnectivityResult struct {
	SourceIP     string         `json:"source_ip"`
	TargetIP     string         `json:"target_ip"`
	PingStatus   string         `json:"ping_status"`   // "reachable" or "unreachable"
	PortStatus   map[int]string `json:"port_status"`   // port -> "open" or "closed"
	Latency      Duration       `json:"latency"`       // ping 延迟
	TestDuration Duration       `json:"test_duration"` // 整个测试耗时
	Timestamp    time.Time      `json:"timestamp"`
}

// ClientRecord represents a client's registration record in the server cache
type ClientRecord struct {
	NodeInfo      NodeInfo  `json:"node_info"`
	Version       int64     `json:"version"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// VersionInfo stores the current global version number
type VersionInfo struct {
	CurrentVersion int64     `json:"current_version"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TestStatus represents the status of a connectivity test
type TestStatus struct {
	Ping         string   `json:"ping"`          // "reachable" or "unreachable"
	PortStatus   string   `json:"port_status"`   // "open" or "closed"
	TestDuration Duration `json:"test_duration"` // 测试耗时
}

// HostTestResults stores host-to-host connectivity test results
// Structure: map[sourceIP]map[targetIP]TestStatus
type HostTestResults map[string]map[string]TestStatus

// PodTestResults stores pod-to-pod connectivity test results
// Structure: map[sourcePodIP]map[targetPodIP]TestStatus
type PodTestResults map[string]map[string]TestStatus

// ServiceTestResults stores custom service connectivity test results
// Structure: map[sourceIP]ConnectivityResult
type ServiceTestResults map[string]*ConnectivityResult

// NetworkReport represents a comprehensive network connectivity report
type NetworkReport struct {
	Timestamp          time.Time          `json:"timestamp"`
	ActiveClientCount  int                `json:"active_client_count"`
	HostIPs            []string           `json:"host_ips"`
	PodIPs             []string           `json:"pod_ips"`
	HostTestSummary    TestSummary        `json:"host_test_summary"`
	PodTestSummary     TestSummary        `json:"pod_test_summary"`
	ServiceTestSummary ServiceTestSummary `json:"service_test_summary"`
}

// TestSummary provides statistics about connectivity tests
type TestSummary struct {
	TotalTests        int      `json:"total_tests"`
	SuccessfulTests   int      `json:"successful_tests"`
	FailedTests       int      `json:"failed_tests"`
	SuccessRate       float64  `json:"success_rate"`
	AvgTestDuration   Duration `json:"avg_test_duration"`   // 平均测试耗时
	TotalTestDuration Duration `json:"total_test_duration"` // 总测试耗时
}

// ServiceTestSummary provides statistics about custom service tests
type ServiceTestSummary struct {
	ServiceName     string  `json:"service_name"`
	TotalTests      int     `json:"total_tests"`
	SuccessfulTests int     `json:"successful_tests"`
	FailedTests     int     `json:"failed_tests"`
	SuccessRate     float64 `json:"success_rate"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
