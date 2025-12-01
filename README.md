# K8s Network Checker

K8s Network Checker 是一个分布式网络监控工具，用于 Kubernetes 集群的网络连通性检测。系统采用客户端-服务器架构，通过在每个节点上运行客户端 Pod 来收集和测试网络连通性信息，服务器端负责聚合数据、管理客户端状态，并提供查询接口以生成网络连通性报告。

## 功能特性

- **自动节点发现**: 客户端自动上报节点信息，服务器维护集群拓扑视图
- **心跳监控**: 定期心跳机制，实时跟踪客户端状态
- **多层次网络测试**:
  - 宿主机层面的网络连通性测试（ping + SSH 端口检测）
  - Pod 层面的网络连通性测试（ping + 健康检查端口检测）
  - 自定义服务可达性测试（DNS 解析 + 连通性验证）
- **版本化客户端管理**: 基于版本号的活跃客户端统计和生命周期管理
- **RESTful API**: 提供完整的查询接口，支持获取测试结果和生成报告
- **定期报告生成**: 自动生成网络健康报告并输出到控制台
- **自动过期清理**: 离线客户端自动从缓存中清理

## 系统架构

```
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
│          │    心跳上报      │                 │             │
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
│                    │  管理员查询    │                       │
│                    └───────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## 技术栈

- **语言**: Go 1.24.0
- **Web 框架**: Gin
- **缓存**: go-cache (内存缓存)
- **容器化**: Docker
- **编排**: Kubernetes
- **测试**: Go testing, testify

## 快速开始

### 前置要求

- Kubernetes 集群 (v1.20+)
- kubectl 已配置并可访问集群
- Docker (用于构建镜像)
- Go 1.24.0+ (用于本地开发)

### 构建镜像

详细的构建说明请参考 [BUILD.md](BUILD.md)

```bash
# 构建服务器镜像
docker build -t sgfoot/k8snet-checker-server:latest -f Dockerfile.server .

# 构建客户端镜像
docker build -t sgfoot/k8snet-checker-client:latest -f Dockerfile.client .
```

### 部署到 Kubernetes

详细的部署说明请参考 [DEPLOY.md](DEPLOY.md)

```bash
# 部署服务器
kubectl apply -f deploy/server-deployment.yaml
kubectl apply -f deploy/server-service.yaml

# 部署客户端
kubectl apply -f deploy/client-daemonset.yaml

# 或者使用一键部署
kubectl apply -f deploy/all-in-one.yaml
```

### 验证部署

详细的测试说明请参考 [TESTING.md](TESTING.md)

```bash
# 检查 Pod 状态
kubectl get pods -n kube-system -l app=k8snet-checker-server
kubectl get pods -n kube-system -l app=k8snet-checker-client

# 查看服务器日志
kubectl logs -n kube-system -l app=k8snet-checker-server -f

# 测试 API
kubectl port-forward -n kube-system svc/k8snet-checker-server 8080:8080
curl http://localhost:8080/api/v1/health
```

## 环境变量配置

### 服务器环境变量

| 变量名 | 说明 | 默认值 | 必需 |
|--------|------|--------|------|
| `CACHE_KEY_SECOND` | 缓存过期时间（秒） | 15 | 否 |
| `LOG_LEVEL` | 日志级别 (debug/info/warn/error) | info | 否 |
| `HTTP_PORT` | HTTP 服务端口 | 8080 | 否 |
| `REPORT_INTERVAL` | 报告生成间隔（秒） | 300 | 否 |

### 客户端环境变量

| 变量名 | 说明 | 默认值 | 必需 |
|--------|------|--------|------|
| `NODE_IP` | 宿主机 IP（K8s 自动注入） | - | 是 |
| `POD_IP` | Pod IP（K8s 自动注入） | - | 是 |
| `POD_NAME` | Pod 名称（K8s 自动注入） | - | 是 |
| `NAMESPACE` | 命名空间（K8s 自动注入） | - | 是 |
| `SERVER_URL` | 服务器 URL | - | 是 |
| `HEARTBEAT_INTERVAL` | 心跳间隔（秒） | 5 | 否 |
| `TEST_PORT` | 宿主机测试端口 | 22 | 否 |
| `CUSTOM_SERVICE_NAME` | 自定义服务名称 | "" | 否 |
| `CLIENT_PORT` | 客户端监听端口 | 6100 | 否 |
| `LOG_LEVEL` | 日志级别 | info | 否 |

## API 接口

### 客户端上报接口

- `POST /api/v1/heartbeat` - 接收心跳和节点信息
- `POST /api/v1/test-results/hosts` - 接收宿主机测试结果
- `POST /api/v1/test-results/pods` - 接收 Pod 测试结果
- `POST /api/v1/test-results/service` - 接收自定义服务测试结果

### 查询接口

- `GET /api/v1/hosts` - 获取所有宿主机 IP 列表
- `GET /api/v1/pods` - 获取所有 Pod IP 列表
- `GET /api/v1/test-results/hosts` - 获取宿主机互探结果
- `GET /api/v1/test-results/pods` - 获取 Pod 互探结果
- `GET /api/v1/test-results/service` - 获取自定义服务探测结果
- `GET /api/v1/clients/count` - 获取活跃客户端数量
- `GET /api/v1/results` - 获取所有测试结果汇总
- `GET /api/v1/health` - 健康检查

## 项目结构

```
.
├── cmd/                    # 应用入口
│   ├── server/            # 服务器主程序
│   └── client/            # 客户端主程序
├── pkg/                   # 核心库
│   ├── models/           # 数据模型
│   ├── cache/            # 缓存管理
│   ├── network/          # 网络测试
│   ├── api/              # API 实现
│   │   ├── client/       # 客户端 API
│   │   └── server/       # 服务器 API
│   ├── heartbeat/        # 心跳上报
│   ├── collector/        # 信息收集
│   ├── clientserver/     # 客户端 HTTP 服务
│   ├── client/           # 客户端管理
│   ├── result/           # 结果管理
│   └── report/           # 报告生成
├── deploy/               # Kubernetes 部署文件
├── Dockerfile.server     # 服务器镜像
├── Dockerfile.client     # 客户端镜像
└── docs/                 # 文档
```

## 开发指南

### 本地开发

```bash
# 克隆仓库
git clone <repository-url>
cd k8snet-checker

# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建二进制文件
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定包的测试
go test ./pkg/cache
go test ./pkg/network
```

## 常见问题

### 1. 客户端 Pod 无法启动

**问题**: 客户端 Pod 处于 CrashLoopBackOff 状态

**排查步骤**:
```bash
# 查看 Pod 日志
kubectl logs -n kube-system <client-pod-name>

# 检查环境变量是否正确注入
kubectl describe pod -n kube-system <client-pod-name>
```

**常见原因**:
- 环境变量 `SERVER_URL` 配置错误
- 服务器 Service 未正确创建
- 网络策略阻止了客户端访问服务器

### 2. 网络测试失败

**问题**: 测试结果显示大量连通性失败

**排查步骤**:
```bash
# 检查客户端是否有 NET_RAW 权限
kubectl get daemonset -n kube-system k8snet-checker-client -o yaml | grep -A 5 securityContext

# 手动测试网络连通性
kubectl exec -n kube-system <client-pod-name> -- ping -c 3 <target-ip>
```

**常见原因**:
- 客户端缺少 `NET_RAW` 和 `NET_ADMIN` 权限
- 防火墙规则阻止了 ICMP 或 TCP 连接
- 目标端口未开放

### 3. 服务器无法接收心跳

**问题**: 活跃客户端数量为 0

**排查步骤**:
```bash
# 检查服务器日志
kubectl logs -n kube-system -l app=k8snet-checker-server

# 检查 Service 是否正常
kubectl get svc -n kube-system k8snet-checker-server

# 测试服务器 API
kubectl port-forward -n kube-system svc/k8snet-checker-server 8080:8080
curl http://localhost:8080/api/v1/health
```

**常见原因**:
- Service 配置错误
- 服务器 Pod 未正常运行
- 客户端配置的 `SERVER_URL` 不正确

### 4. 缓存数据丢失

**问题**: 客户端信息频繁丢失

**排查步骤**:
```bash
# 检查缓存过期时间配置
kubectl get deployment -n kube-system k8snet-checker-server -o yaml | grep CACHE_KEY_SECOND

# 检查心跳间隔配置
kubectl get daemonset -n kube-system k8snet-checker-client -o yaml | grep HEARTBEAT_INTERVAL
```

**解决方案**:
- 确保 `HEARTBEAT_INTERVAL` < `CACHE_KEY_SECOND`
- 建议 `HEARTBEAT_INTERVAL` 为 5 秒，`CACHE_KEY_SECOND` 为 15 秒

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 许可证

[MIT License](LICENSE)

## 联系方式

如有问题或建议，请提交 Issue 或联系维护者。
