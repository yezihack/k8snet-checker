# K8s Network Checker - Client

[![GitHub](https://img.shields.io/badge/GitHub-k8snet--checker-blue?logo=github)](https://github.com/yezihack/k8snet-checker)
[![Docker Image Version](https://img.shields.io/docker/v/sgfoot/k8snet-checker-client?sort=semver)](https://hub.docker.com/r/sgfoot/k8snet-checker-client)
[![Docker Image Size](https://img.shields.io/docker/image-size/sgfoot/k8snet-checker-client/latest)](https://hub.docker.com/r/sgfoot/k8snet-checker-client)
[![Docker Pulls](https://img.shields.io/docker/pulls/sgfoot/k8snet-checker-client)](https://hub.docker.com/r/sgfoot/k8snet-checker-client)
[![License](https://img.shields.io/github/license/yezihack/k8snet-checker)](https://github.com/yezihack/k8snet-checker/blob/main/LICENSE)

K8s Network Checker 的客户端组件，通过 DaemonSet 在每个节点上运行，执行网络连通性测试并上报结果。

## 快速开始

### 在 Kubernetes 中部署

```bash
# 使用 Helm（推荐）
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system

# 或使用 kubectl
kubectl apply -f https://raw.githubusercontent.com/yezihack/k8snet-checker/main/deploy/all-in-one.yaml
```

## 环境变量

### 必需变量（Kubernetes 自动注入）

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `NODE_IP` | 宿主机 IP | `192.168.1.100` |
| `POD_IP` | Pod IP | `10.244.0.10` |
| `POD_NAME` | Pod 名称 | `k8snet-checker-client-abc123` |
| `NAMESPACE` | 命名空间 | `kube-system` |

### 可选变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_URL` | 服务器 URL | `http://k8snet-checker-server.kube-system.svc.cluster.local:8080` |
| `HEARTBEAT_INTERVAL` | 心跳间隔（秒） | `5` |
| `TEST_PORT` | 宿主机测试端口 | `22` |
| `CUSTOM_SERVICE_NAME` | 自定义服务名称 | `""` |
| `CUSTOM_SERVICE_PORT` | 自定义服务端口 | `80` |
| `CLIENT_PORT` | 客户端监听端口 | `6100` |
| `LOG_LEVEL` | 日志级别 | `info` |

## 功能特性

### 网络测试

1. **宿主机层面测试**
   - Ping 测试节点之间的连通性
   - 测试 SSH 端口（默认 22）可达性
   - 记录延迟和测试耗时

2. **Pod 层面测试**
   - Ping 测试 Pod 之间的连通性
   - 测试健康检查端口（6100）可达性
   - 记录延迟和测试耗时

3. **自定义服务测试**
   - DNS 解析验证
   - 服务端口可达性测试
   - 支持自定义端口配置

### 自动化功能

- 自动发现集群中的所有节点和 Pod
- 定期执行网络测试（可配置间隔）
- 实时上报测试结果到服务器
- 心跳机制保持与服务器的连接

## 权限要求

客户端需要以下权限：

- `NET_RAW` - 执行 ping 测试
- `NET_ADMIN` - 网络管理权限
- `hostNetwork: true` - 访问宿主机网络

## 健康检查

```bash
curl http://localhost:6100/health
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

- CPU: 25m
- Memory: 32Mi

**推荐配置**：

- CPU: 50m
- Memory: 64Mi

**生产环境**：

- CPU: 100m
- Memory: 128Mi

## 端口

- `6100` - 健康检查端口

## 配置示例

### Kubernetes DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: k8snet-checker-client
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: k8snet-checker-client
  template:
    metadata:
      labels:
        app: k8snet-checker-client
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: client
        image: sgfoot/k8snet-checker-client:latest
        ports:
        - containerPort: 6100
          name: health
        env:
        - name: NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: SERVER_URL
          value: "http://k8snet-checker-server.kube-system.svc.cluster.local:8080"
        - name: HEARTBEAT_INTERVAL
          value: "5"
        - name: TEST_PORT
          value: "22"
        - name: CLIENT_PORT
          value: "6100"
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        securityContext:
          capabilities:
            add:
            - NET_RAW
            - NET_ADMIN
        livenessProbe:
          httpGet:
            path: /health
            port: 6100
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 6100
          initialDelaySeconds: 5
          periodSeconds: 10
      tolerations:
      - operator: Exists
```

## 测试自定义服务

如果需要测试特定服务的可达性：

```yaml
env:
- name: CUSTOM_SERVICE_NAME
  value: "kubernetes.default.svc.cluster.local"
- name: CUSTOM_SERVICE_PORT
  value: "443"
```

## 故障排查

### 查看日志

```bash
# 查看所有客户端日志
kubectl logs -n kube-system -l app=k8snet-checker-client --tail=100

# 查看特定节点的客户端日志
kubectl logs -n kube-system <client-pod-name> -f
```

### 测试连接

```bash
# 健康检查
kubectl exec -n kube-system <client-pod-name> -- wget -qO- http://localhost:6100/health

# 测试服务器连接
kubectl exec -n kube-system <client-pod-name> -- wget -qO- http://k8snet-checker-server.kube-system.svc.cluster.local:8080/api/v1/health
```

### 常见问题

**Q: 客户端无法启动？**

A: 检查是否有 `NET_RAW` 和 `NET_ADMIN` 权限，确保 `hostNetwork: true`

**Q: 无法连接到服务器？**

A: 检查 `SERVER_URL` 配置是否正确，确保服务器 Service 存在

**Q: Ping 测试失败？**

A: 检查防火墙规则，确保 ICMP 协议未被阻止

**Q: 端口测试失败？**

A: 检查目标端口是否开放，确认防火墙规则

**Q: 客户端数量不正确？**

A: 检查 DaemonSet 的 tolerations 配置，确保可以调度到所有节点

## 网络测试流程

1. **启动阶段**
   - 收集节点和 Pod 信息
   - 向服务器发送心跳注册
   - 获取集群中其他节点和 Pod 列表

2. **测试阶段**
   - 执行宿主机互探测试
   - 执行 Pod 互探测试
   - 执行自定义服务测试（如果配置）
   - 记录测试结果和耗时

3. **上报阶段**
   - 将测试结果上报到服务器
   - 定期发送心跳保持连接
   - 更新客户端状态

## 性能优化

- 使用并发测试提高效率（默认 10 个并发）
- 自动跳过源 IP 避免自测
- 合理的超时设置避免长时间等待
- 内存占用低，适合大规模部署

## 安全考虑

- 仅在集群内部通信，不暴露外部端口
- 使用 Kubernetes Service 进行服务发现
- 支持网络策略限制
- 最小权限原则

## 文档

- [项目主页](https://github.com/yezihack/k8snet-checker)
- [完整文档](https://github.com/yezihack/k8snet-checker/blob/main/README.md)
- [部署指南](https://github.com/yezihack/k8snet-checker/blob/main/DEPLOY.md)
- [网络测试模块](https://github.com/yezihack/k8snet-checker/blob/main/pkg/network/README.md)

## 许可证

MIT License

## 联系方式

- GitHub: <https://github.com/yezihack/k8snet-checker>
- Issues: <https://github.com/yezihack/k8snet-checker/issues>

## 相关镜像

- **服务器端**: [sgfoot/k8snet-checker-server](https://hub.docker.com/r/sgfoot/k8snet-checker-server)
