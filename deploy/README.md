# Kubernetes 部署文件

本目录包含 K8s Network Checker 的 Kubernetes 部署配置文件。

## 文件说明

- `server-deployment.yaml` - Server 组件的 Deployment 配置
- `server-service.yaml` - Server 组件的 Service 配置
- `client-daemonset.yaml` - Client 组件的 DaemonSet 配置
- `all-in-one.yaml` - 包含所有资源的统一部署文件

## 快速部署

```bash
kubectl apply -f deploy/all-in-one.yaml
```

## 验证部署

### 检查 Pod 状态

```bash
# 检查 Server Pod
kubectl get pods -n kube-system -l app=k8snet-checker-server

# 检查 Client Pod（应该在每个节点上都有一个）
kubectl get pods -n kube-system -l app=k8snet-checker-client -o wide
```

### 检查 Service

```bash
kubectl get svc -n kube-system k8snet-checker-server
```

### 查看日志

```bash
# 查看 Server 日志
kubectl logs -n kube-system -l app=k8snet-checker-server -f

# 查看 Client 日志（指定某个 Pod）
kubectl logs -n kube-system <client-pod-name> -f
```

## 环境变量配置

### Server 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| CACHE_KEY_SECOND | 15 | 缓存过期时间（秒） |
| LOG_LEVEL | info | 日志级别（debug/info/warn/error） |
| HTTP_PORT | 8080 | HTTP 服务端口 |
| REPORT_INTERVAL | 300 | 报告生成间隔（秒） |

### Client 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| NODE_IP | - | 宿主机 IP（自动注入） |
| POD_IP | - | Pod IP（自动注入） |
| POD_NAME | - | Pod 名称（自动注入） |
| NAMESPACE | - | 命名空间（自动注入） |
| SERVER_URL | <http://k8snet-checker-server.kube-system.svc.cluster.local:8080> | Server 地址 |
| HEARTBEAT_INTERVAL | 5 | 心跳间隔（秒） |
| TEST_PORT | 22 | 宿主机测试端口 |
| CUSTOM_SERVICE_NAME | "" | 自定义服务名称 |
| CLIENT_PORT | 6100 | 客户端监听端口 |
| LOG_LEVEL | info | 日志级别 |

## 资源配置

### Server 资源限制

- Requests: CPU 100m, Memory 128Mi
- Limits: CPU 200m, Memory 256Mi

### Client 资源限制

- Requests: CPU 50m, Memory 64Mi
- Limits: CPU 100m, Memory 128Mi

## 健康检查

### Server

- Liveness Probe: `GET /api/v1/health`
- Readiness Probe: `GET /api/v1/health`

### Client

- Liveness Probe: `GET /health`
- Readiness Probe: `GET /health`

## 网络配置

### Server

- 使用 ClusterIP Service
- 端口：8080

### Client

- 使用 hostNetwork: true（需要访问宿主机网络）
- 端口：6100
- 需要 NET_RAW 和 NET_ADMIN 权限（用于 ping 测试）

## 容忍度配置

Client DaemonSet 配置了以下容忍度，确保可以在所有节点上运行：

- NoSchedule: 允许在有 NoSchedule 污点的节点上调度
- NoExecute: 允许在有 NoExecute 污点的节点上运行

## 卸载

```bash
# 删除所有资源
kubectl delete -f deploy/all-in-one.yaml
```

## 自定义配置

如需修改配置，可以编辑相应的 YAML 文件：

1. 修改环境变量值
2. 调整资源限制
3. 修改镜像地址
4. 调整副本数（仅 Server）

修改后重新应用：

```bash
kubectl apply -f deploy/all-in-one.yaml
```

## 故障排查

### Server 无法启动

```bash
# 查看 Pod 事件
kubectl describe pod -n kube-system -l app=network-checker-server

# 查看日志
kubectl logs -n kube-system -l app=network-checker-server
```

### Client 无法连接 Server

```bash
# 检查 Service 是否正常
kubectl get svc -n kube-system k8snet-checker-server

# 测试 DNS 解析
kubectl exec -n kube-system <client-pod-name> -- nslookup k8snet-checker-server.kube-system.svc.cluster.local

# 测试连通性
kubectl exec -n kube-system <client-pod-name> -- curl http://k8snet-checker-server.kube-system.svc.cluster.local:8080/api/v1/health
```

### Client 无法执行 ping 测试

确保 Client Pod 有足够的权限：

- securityContext 中添加了 NET_RAW 和 NET_ADMIN 权限
- 使用 hostNetwork: true

## 注意事项

1. 本系统部署在 `kube-system` 命名空间
2. Client 使用 hostNetwork，可能与其他使用相同端口的应用冲突
3. Client 需要特殊权限（NET_RAW, NET_ADMIN）来执行网络测试
4. 确保集群节点间网络连通性良好
5. 建议在测试环境先验证后再部署到生产环境
