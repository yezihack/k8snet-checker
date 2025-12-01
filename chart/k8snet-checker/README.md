# K8s Network Checker Helm Chart

Kubernetes 集群网络连通性检测工具的 Helm Chart。

## 简介

K8s Network Checker 是一个分布式网络监控工具，用于检测 Kubernetes 集群的网络连通性。系统采用客户端-服务器架构：

- **服务器端**: 单个 Deployment，负责聚合数据和提供 API
- **客户端**: DaemonSet，在每个节点上运行，执行网络测试

## 功能特性

- 自动节点发现和心跳监控
- 宿主机层面网络连通性测试
- Pod 层面网络连通性测试
- 自定义服务可达性测试
- RESTful API 查询接口
- 定期生成网络健康报告

## 前置要求

- Kubernetes 1.20+
- Helm 3.0+

## 安装

### 添加 Helm 仓库（如果已发布）

```bash
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
```

### 从本地安装

```bash
# 克隆仓库
git clone https://github.com/yezihack/k8snet-checker.git
cd k8snet-checker/chart

# 安装 Chart
helm install k8snet-checker ./k8snet-checker -n toolbox
```

### 自定义安装

```bash
# 使用自定义 values 文件
helm install k8snet-checker ./k8snet-checker \
  -n toolbox \
  -f custom-values.yaml

# 或使用 --set 参数
helm install k8snet-checker ./k8snet-checker \
  -n toolbox \
  --set server.image.tag=v1.0.0 \
  --set client.image.tag=v1.0.0
```

## 配置

### 主要配置项

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `namespace` | 部署命名空间 | `toolbox` |
| `server.replicaCount` | 服务器副本数 | `1` |
| `server.image.repository` | 服务器镜像仓库 | `sgfoot/k8snet-checker-server` |
| `server.image.tag` | 服务器镜像标签 | `latest` |
| `server.service.port` | 服务器端口 | `8080` |
| `server.env.cacheKeySecond` | 缓存过期时间（秒） | `15` |
| `server.env.logLevel` | 日志级别 | `info` |
| `server.env.reportInterval` | 报告生成间隔（秒） | `300` |
| `client.image.repository` | 客户端镜像仓库 | `sgfoot/k8snet-checker-client` |
| `client.image.tag` | 客户端镜像标签 | `latest` |
| `client.env.heartbeatInterval` | 心跳间隔（秒） | `5` |
| `client.env.testPort` | 宿主机测试端口 | `22` |
| `client.env.customServiceName` | 自定义服务名称 | `""` |

### 资源配置

#### 服务器资源

```yaml
server:
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi
```

#### 客户端资源

```yaml
client:
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 50m
      memory: 64Mi
```

### 自定义服务测试

如果需要测试自定义服务的可达性：

```yaml
client:
  env:
    customServiceName: "kubernetes.default.svc.cluster.local"
```

## 使用示例

### 基本安装

```bash
helm install k8snet-checker ./k8snet-checker -n toolbox
```

### 生产环境配置

```yaml
# production-values.yaml
server:
  replicaCount: 1
  image:
    tag: "v1.0.0"
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi
  env:
    logLevel: "info"
    reportInterval: "300"

client:
  image:
    tag: "v1.0.0"
  resources:
    limits:
      cpu: 300m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi
  env:
    logLevel: "info"
    customServiceName: "kubernetes.default.svc.cluster.local"
```

```bash
helm install k8snet-checker ./k8snet-checker \
  -n toolbox \
  -f production-values.yaml
```

### 开发环境配置

```bash
helm install k8snet-checker ./k8snet-checker \
  -n toolbox \
  --set server.env.logLevel=debug \
  --set client.env.logLevel=debug \
  --set server.image.tag=dev \
  --set client.image.tag=dev
```

## 验证部署

### 检查 Pod 状态

```bash
kubectl get pods -n toolbox -l app.kubernetes.io/name=k8snet-checker
```

### 查看服务器日志

```bash
kubectl logs -n toolbox -l app.kubernetes.io/component=server -f
```

### 查看客户端日志

```bash
kubectl logs -n toolbox -l app.kubernetes.io/component=client -f
```

### 测试 API

```bash
# 端口转发
kubectl port-forward -n toolbox svc/k8snet-checker-server 8080:8080

# 健康检查
curl http://localhost:8080/api/v1/health

# 获取活跃客户端数量
curl http://localhost:8080/api/v1/clients/count

# 获取所有测试结果
curl http://localhost:8080/api/v1/results | jq .
```

## 升级

```bash
# 升级到新版本
helm upgrade k8snet-checker ./k8snet-checker \
  -n toolbox \
  --set server.image.tag=v1.1.0 \
  --set client.image.tag=v1.1.0

# 查看升级历史
helm history k8snet-checker -n toolbox

# 回滚到上一个版本
helm rollback k8snet-checker -n toolbox
```

## 卸载

```bash
helm uninstall k8snet-checker -n toolbox
```

## 故障排查

### 客户端无法连接服务器

检查服务器 Service 是否正常：

```bash
kubectl get svc -n toolbox k8snet-checker-server
```

检查客户端环境变量：

```bash
kubectl describe pod -n toolbox <client-pod-name> | grep SERVER_URL
```

### 网络测试失败

检查客户端权限：

```bash
kubectl get daemonset -n toolbox k8snet-checker-client -o yaml | grep -A 5 securityContext
```

确保客户端有 `NET_RAW` 和 `NET_ADMIN` 权限。

### 活跃客户端数量为 0

检查心跳间隔和缓存过期时间配置：

```bash
helm get values k8snet-checker -n toolbox
```

确保 `heartbeatInterval` < `cacheKeySecond`。

## 最佳实践

1. **资源限制**: 根据集群规模调整资源配置
2. **日志级别**: 生产环境使用 `info`，开发环境使用 `debug`
3. **心跳间隔**: 保持默认 5 秒，确保及时检测
4. **缓存过期**: 设置为心跳间隔的 3 倍
5. **监控告警**: 配置 Prometheus 监控
6. **定期检查**: 定期查看报告，及时发现问题

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 联系方式

- 作者: sgfoot
- 邮箱: freeit@126.com
- 项目: https://github.com/yezihack/k8snet-checker
