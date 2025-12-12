# K8s Network Checker

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24.0+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Hub-2496ED?logo=docker)](https://hub.docker.com/u/sgfoot)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.20+-326CE5?logo=kubernetes)](https://kubernetes.io/)

K8s Network Checker is a distributed network monitoring tool for Kubernetes clusters. It uses a client-server architecture where client Pods run on each node to collect and test network connectivity, while the server aggregates data, manages client states, and provides query APIs to generate comprehensive network connectivity reports.

[中文文档](README.md) | English

## Features

- **Automatic Node Discovery**: Clients automatically report node information, server maintains cluster topology view
- **Heartbeat Monitoring**: Periodic heartbeat mechanism for real-time client status tracking
- **Multi-Level Network Testing**:
  - Host-level network connectivity testing (ping + SSH port detection)
  - Pod-level network connectivity testing (ping + health check port detection)
  - Custom service reachability testing (DNS resolution + connectivity verification)
- **Versioned Client Management**: Active client statistics and lifecycle management based on version numbers
- **RESTful API**: Complete query interface supporting test results retrieval and report generation
- **Periodic Report Generation**: Automatic network health report generation and console output
- **Automatic Expiration Cleanup**: Offline clients automatically cleaned from cache

## System Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Node 1     │  │   Node 2     │  │   Node N     │     │
│  │              │  │              │  │              │     │
│  │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │     │
│  │  │ Client │  │  │  │ Client │  │  │  │ Client │  │     │
│  │  │  Pod   │  │  │  │  Pod   │  │  │  │  Pod   │  │     │
│  │  │ :6100  │  │  │  │ :6100  │  │  │  │ :6100  │  │     │
│  │  └────┬───┘  │  │  └────┬───┘  │  │  └────┬───┘  │     │
│  └───────┼──────┘  └───────┼──────┘  └───────┼──────┘     │
│          │                 │                 │             │
│          │   Heartbeat     │                 │             │
│          └─────────────────┼─────────────────┘             │
│                            │                               │
│                    ┌───────▼────────┐                      │
│                    │  Server Pod    │                      │
│                    │  (Deployment)  │                      │
│                    │     :8080      │                      │
│                    │                │                      │
│                    │  ┌──────────┐  │                      │
│                    │  │ go-cache │  │                      │
│                    │  │  Memory  │  │                      │
│                    │  └──────────┘  │                      │
│                    └────────────────┘                      │
│                            │                               │
│                            │ REST API                      │
│                            ▼                               │
│                    ┌───────────────┐                       │
│                    │ Admin Query   │                       │
│                    └───────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Gin
- **Cache**: go-cache (in-memory cache)
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Testing**: Go testing, testify

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured and accessible to cluster
- Docker (for building images)
- Go 1.24.0+ (for local development)

### Build Images

For detailed build instructions, see [BUILD.md](docs/BUILD.md)

```bash
# Build server image
docker build -t sgfoot/k8snet-checker-server:latest -f Dockerfile.server .

# Build client image
docker build -t sgfoot/k8snet-checker-client:latest -f Dockerfile.client .
```

### Deploy to Kubernetes

For detailed deployment instructions, see [DEPLOY.md](DEPLOY.md)

#### Method 1: Using Helm (Recommended)

```bash
# Add Helm repository
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update k8snet-checker
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system

# Or install from local
git clone https://github.com/yezihack/k8snet-checker.git
cd k8snet-checker
helm install k8snet-checker ./chart/k8snet-checker -n kube-system
```

#### Method 2: Using kubectl

```bash
# use all-in-one deployment
kubectl apply -f deploy/all-in-one.yaml
```

### Verify Deployment

For detailed testing instructions, see [TESTING.md](TESTING.md)

```bash
# Check Pod status
kubectl get pods -n kube-system -l app=k8snet-checker-server
kubectl get pods -n kube-system -l app=k8snet-checker-client

# View server logs
kubectl logs -n kube-system -l app=k8snet-checker-server -f

# Test API
kubectl port-forward -n kube-system svc/k8snet-checker-server 8080:8080
curl http://localhost:8080/api/v1/health
```

## Environment Variables

### Server Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `CACHE_KEY_SECOND` | Cache expiration time (seconds) | 15 | No |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | info | No |
| `HTTP_PORT` | HTTP service port | 8080 | No |
| `REPORT_INTERVAL` | Report generation interval (seconds) | 300 | No |

### Client Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NODE_IP` | Host IP (K8s auto-injected) | - | Yes |
| `POD_IP` | Pod IP (K8s auto-injected) | - | Yes |
| `POD_NAME` | Pod name (K8s auto-injected) | - | Yes |
| `NAMESPACE` | Namespace (K8s auto-injected) | - | Yes |
| `SERVER_URL` | Server URL | - | Yes |
| `HEARTBEAT_INTERVAL` | Heartbeat interval (seconds) | 5 | No |
| `TEST_PORT` | Host test port | 22 | No |
| `CUSTOM_SERVICE_NAME` | Custom service name | "" | No |
| `CUSTOM_SERVICE_PORT` | Custom service port | 80 | No |
| `CLIENT_PORT` | Client listening port | 6100 | No |
| `LOG_LEVEL` | Log level | info | No |

## API Endpoints

### Client Reporting Endpoints

- `POST /api/v1/heartbeat` - Receive heartbeat and node information
- `POST /api/v1/test-results/hosts` - Receive host test results
- `POST /api/v1/test-results/pods` - Receive Pod test results
- `POST /api/v1/test-results/service` - Receive custom service test results

### Query Endpoints

- `GET /api/v1/hosts` - Get all host IP list
- `GET /api/v1/pods` - Get all Pod IP list
- `GET /api/v1/test-results/hosts` - Get host connectivity test results
- `GET /api/v1/test-results/pods` - Get Pod connectivity test results
- `GET /api/v1/test-results/service` - Get custom service test results
- `GET /api/v1/clients/count` - Get active client count
- `GET /api/v1/results` - Get all test results summary
- `GET /api/v1/health` - Health check

### API Response Examples

#### Get Host Test Results

```bash
curl http://localhost:8080/api/v1/test-results/hosts
```

Response:

```json
{
  "results": {
    "10.42.0.26": {
      "192.168.9.35": {
        "ping": "reachable",
        "port_status": "open",
        "test_duration": "2.01s"
      }
    }
  }
}
```

#### Get Network Report

```bash
curl http://localhost:8080/api/v1/results
```

Response:

```json
{
  "timestamp": "2025-12-12T21:00:00Z",
  "active_client_count": 3,
  "host_ips": ["192.168.1.10", "192.168.1.11", "192.168.1.12"],
  "pod_ips": ["10.42.0.1", "10.42.0.2", "10.42.0.3"],
  "host_test_summary": {
    "total_tests": 6,
    "successful_tests": 6,
    "failed_tests": 0,
    "success_rate": 100.0,
    "avg_test_duration": "1.85s",
    "total_test_duration": "11.10s"
  },
  "pod_test_summary": {
    "total_tests": 6,
    "successful_tests": 5,
    "failed_tests": 1,
    "success_rate": 83.33,
    "avg_test_duration": "1.92s",
    "total_test_duration": "11.52s"
  }
}
```

## Project Structure

```text
.
├── cmd/                    # Application entry points
│   ├── server/            # Server main program
│   └── client/            # Client main program
├── pkg/                   # Core libraries
│   ├── models/           # Data models
│   ├── cache/            # Cache management
│   ├── network/          # Network testing
│   ├── api/              # API implementation
│   │   ├── client/       # Client API
│   │   └── server/       # Server API
│   ├── heartbeat/        # Heartbeat reporting
│   ├── collector/        # Information collection
│   ├── clientserver/     # Client HTTP service
│   ├── client/           # Client management
│   ├── result/           # Result management
│   └── report/           # Report generation
├── deploy/               # Kubernetes deployment files
├── chart/                # Helm chart
├── Dockerfile.server     # Server image
├── Dockerfile.client     # Client image
└── docs/                 # Documentation
```

## Development Guide

### Local Development

```bash
# Clone repository
git clone https://github.com/yezihack/k8snet-checker.git
cd k8snet-checker

# Install dependencies
go mod download

# Run tests
go test ./...

# Build binaries
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/cache
go test ./pkg/network
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check diagnostics
go vet ./...
```

## Troubleshooting

### 1. Client Pods Fail to Start

**Issue**: Client Pods are in CrashLoopBackOff state

**Troubleshooting Steps**:

```bash
# View Pod logs
kubectl logs -n kube-system <client-pod-name>

# Check if environment variables are correctly injected
kubectl describe pod -n kube-system <client-pod-name>
```

**Common Causes**:

- Environment variable `SERVER_URL` is misconfigured
- Server Service not created correctly
- Network policies blocking client access to server

### 2. Network Tests Failing

**Issue**: Test results show many connectivity failures

**Troubleshooting Steps**:

```bash
# Check if client has NET_RAW capability
kubectl get daemonset -n kube-system k8snet-checker-client -o yaml | grep -A 5 securityContext

# Manually test network connectivity
kubectl exec -n kube-system <client-pod-name> -- ping -c 3 <target-ip>
```

**Common Causes**:

- Client missing `NET_RAW` and `NET_ADMIN` capabilities
- Firewall rules blocking ICMP or TCP connections
- Target ports not open

### 3. Server Not Receiving Heartbeats

**Issue**: Active client count is 0

**Troubleshooting Steps**:

```bash
# Check server logs
kubectl logs -n kube-system -l app=k8snet-checker-server

# Check if Service is working
kubectl get svc -n kube-system k8snet-checker-server

# Test server API
kubectl port-forward -n kube-system svc/k8snet-checker-server 8080:8080
curl http://localhost:8080/api/v1/health
```

**Common Causes**:

- Service misconfigured
- Server Pod not running properly
- Client's `SERVER_URL` configuration incorrect

### 4. Cache Data Loss

**Issue**: Client information frequently lost

**Troubleshooting Steps**:

```bash
# Check cache expiration time configuration
kubectl get deployment -n kube-system k8snet-checker-server -o yaml | grep CACHE_KEY_SECOND

# Check heartbeat interval configuration
kubectl get daemonset -n kube-system k8snet-checker-client -o yaml | grep HEARTBEAT_INTERVAL
```

**Solution**:

- Ensure `HEARTBEAT_INTERVAL` < `CACHE_KEY_SECOND`
- Recommended: `HEARTBEAT_INTERVAL` = 5 seconds, `CACHE_KEY_SECOND` = 15 seconds

## Configuration Examples

### Custom Service Testing

To test connectivity to a custom service (e.g., internal API):

```yaml
# In client-daemonset.yaml
env:
  - name: CUSTOM_SERVICE_NAME
    value: "my-api-service.default.svc.cluster.local"
  - name: CUSTOM_SERVICE_PORT
    value: "8080"
```

### Adjust Test Intervals

```yaml
# Server configuration
env:
  - name: REPORT_INTERVAL
    value: "600"  # Generate report every 10 minutes

# Client configuration
env:
  - name: HEARTBEAT_INTERVAL
    value: "10"  # Send heartbeat every 10 seconds
```

## Performance Considerations

- **Memory Usage**: Server memory usage depends on cluster size. Approximately 50MB base + 1MB per 100 clients
- **Network Overhead**: Each client sends heartbeat every 5 seconds (configurable)
- **Test Frequency**: Network tests run after each heartbeat
- **Concurrent Testing**: Client uses worker pool (default 10 workers) to limit concurrent tests

## Security Considerations

- **Capabilities**: Client requires `NET_RAW` and `NET_ADMIN` for ping tests
- **Network Policies**: Ensure clients can reach server on port 8080
- **RBAC**: No special RBAC permissions required
- **Data Storage**: All data stored in memory, no persistent storage

## Roadmap

- [ ] Add Prometheus metrics export
- [ ] Support for custom test protocols (HTTP, gRPC)
- [ ] Web UI for visualization
- [ ] Historical data persistence
- [ ] Alert notifications
- [ ] Multi-cluster support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Gin Web Framework](https://github.com/gin-gonic/gin)
- Caching powered by [go-cache](https://github.com/patrickmn/go-cache)
- Logging with [Zap](https://github.com/uber-go/zap)

## Contact

- **Author**: sgfoot
- **Email**: freeit@126.com
- **GitHub**: [https://github.com/yezihack/k8snet-checker](https://github.com/yezihack/k8snet-checker)
- **Issues**: [https://github.com/yezihack/k8snet-checker/issues](https://github.com/yezihack/k8snet-checker/issues)

## Star History

If you find this project helpful, please consider giving it a star ⭐️

---

**Made with ❤️ for the Kubernetes community**
