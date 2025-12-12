# K8s Network Checker - Server

[![Docker Image Version](https://img.shields.io/docker/v/sgfoot/k8snet-checker-server?sort=semver)](https://hub.docker.com/r/sgfoot/k8snet-checker-server)
[![Docker Image Size](https://img.shields.io/docker/image-size/sgfoot/k8snet-checker-server/latest)](https://hub.docker.com/r/sgfoot/k8snet-checker-server)
[![Docker Pulls](https://img.shields.io/docker/pulls/sgfoot/k8snet-checker-server)](https://hub.docker.com/r/sgfoot/k8snet-checker-server)

K8s Network Checker 的服务器端组件，用于聚合网络测试数据并提供查询 API。

## 快速开始

### 在 Kubernetes 中部署

```bash
# 使用 Helm（推荐）
helm install k8snet-checker \
  https://github.com/yezihack/k8snet-checker/releases/download/v0.1.0/k8snet-checker-0.1.0.tgz \
  -n kube-system --create-namespace

# 或使用 kubectl
kubectl apply -f https://raw.githubusercontent.com/yezihack/k8snet-checker/main/deploy/all-in-one.yaml
```

### 单独运行（测试）

```bash
docker run -d \
  --name k8snet-checker-server \
  -p 8080:8080 \
  -e LOG_LEVEL=info \
  -e CACHE_KEY_SECOND=15 \
  sgfoot/k8snet-checker-server:latest
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `CACHE_KEY_SECOND` | 缓存过期时间（秒） | `15` |
| `LOG_LEVEL` | 日志级别 (debug/info/warn/error) | `info` |
| `HTTP_PORT` | HTTP 服务端口 | `8080` |
| `REPORT_INTERVAL` | 报告生成间隔（秒） | `300` |

## API 端点

### 客户端上报

- `POST /api/v1/heartbeat` - 心跳上报
- `POST /api/v1/test-results/hosts` - 宿主机测试结果
- `POST /api/v1/test-results/pods` - Pod 测试结果
- `POST /api/v1/test-results/service` - 服务测试结果

### 查询接口

- `GET /api/v1/health` - 健康检查
- `GET /api/v1/hosts` - 获取宿主机 IP 列表
- `GET /api/v1/pods` - 获取 Pod IP 列表
- `GET /api/v1/test-results/hosts` - 获取宿主机测试结果
- `GET /api/v1/test-results/pods` - 获取 Pod 测试结果
- `GET /api/v1/test-results/service` - 获取服务测试结果
- `GET /api/v1/clients/count` - 获取活跃客户端数量
- `GET /api/v1/results` - 获取所有测试结果

## 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

## 支持的架构

- `linux/amd64`
- `linux/arm64`

## 版本说明

- `latest` - 最新稳定版本
- `x.y.z` - 特定版本号（如 `0.1.0`）
- `develop` - 开发版本

## 资源要求

**最小配置**：
- CPU: 50m
- Memory: 64Mi

**推荐配置**：
- CPU: 100m
- Memory: 128Mi

**生产环境**（大规模集群）：
- CPU: 200m
- Memory: 256Mi

## 端口

- `8080` - HTTP API 端口

## 数据存储

使用内存缓存（go-cache），无需外部数据库。数据在 Pod 重启后会丢失。

## 日志

日志输出到标准输出（stdout），可通过 `kubectl logs` 查看：

```bash
kubectl logs -n kube-system -l app=k8snet-checker-server -f
```

## 配置示例

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8snet-checker-server
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8snet-checker-server
  template:
    metadata:
      labels:
        app: k8snet-checker-server
    spec:
      containers:
      - name: server
        image: sgfoot/k8snet-checker-server:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: CACHE_KEY_SECOND
          value: "15"
        - name: REPORT_INTERVAL
          value: "300"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Docker Compose

```yaml
version: '3.8'
services:
  server:
    image: sgfoot/k8snet-checker-server:latest
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
      - CACHE_KEY_SECOND=15
      - REPORT_INTERVAL=300
    restart: unless-stopped
```

## 故障排查

### 查看日志

```bash
# Kubernetes
kubectl logs -n kube-system -l app=k8snet-checker-server

# Docker
docker logs k8snet-checker-server
```

### 测试连接

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 获取活跃客户端数量
curl http://localhost:8080/api/v1/clients/count
```

### 常见问题

**Q: 活跃客户端数量为 0？**

A: 检查客户端是否正常运行，确保 `HEARTBEAT_INTERVAL` < `CACHE_KEY_SECOND`

**Q: API 返回空数据？**

A: 等待客户端完成首次测试（通常需要 1-2 分钟）

**Q: 内存占用过高？**

A: 减小 `CACHE_KEY_SECOND` 值，或增加内存限制

## 文档

- [项目主页](https://github.com/yezihack/k8snet-checker)
- [完整文档](https://github.com/yezihack/k8snet-checker/blob/main/README.md)
- [部署指南](https://github.com/yezihack/k8snet-checker/blob/main/DEPLOY.md)
- [API 文档](https://github.com/yezihack/k8snet-checker/blob/main/docs/API.md)

## 许可证

MIT License

## 联系方式

- GitHub: https://github.com/yezihack/k8snet-checker
- Issues: https://github.com/yezihack/k8snet-checker/issues

## 相关镜像

- **客户端**: [sgfoot/k8snet-checker-client](https://hub.docker.com/r/sgfoot/k8snet-checker-client)
