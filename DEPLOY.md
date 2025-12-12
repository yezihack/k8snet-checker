# 部署指南

本文档说明如何将 K8s Network Checker 部署到 Kubernetes 集群。

## 目录

- [前置要求](#前置要求)
- [快速部署](#快速部署)
- [分步部署](#分步部署)
- [配置说明](#配置说明)
- [验证部署](#验证部署)
- [更新部署](#更新部署)
- [卸载](#卸载)
- [故障排查](#故障排查)

## 前置要求

### 必需条件

- **Kubernetes 集群**: v1.20 或更高版本
- **kubectl**: 已配置并可访问集群
- **权限**: 需要在 `toolbox` 命名空间创建资源的权限
- **镜像**: 已构建的 Docker 镜像（参考 [BUILD.md](BUILD.md)）

### 验证集群状态

```bash
# 检查 kubectl 配置
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查命名空间
kubectl get namespace toolbox
```

## 快速部署

使用一键部署文件快速部署整个系统：

```bash
# 部署所有组件
kubectl apply -f deploy/all-in-one.yaml

# 等待 Pod 就绪
kubectl wait --for=condition=ready pod -l app=k8snet-checker-server -n toolbox --timeout=60s
kubectl wait --for=condition=ready pod -l app=k8snet-checker-client -n toolbox --timeout=60s

# 查看部署状态
kubectl get all -n toolbox -l component=server
kubectl get all -n toolbox -l component=client
```

## 分步部署

### 1. 部署服务器

#### 1.1 创建 Deployment

```bash
# 部署服务器 Deployment
kubectl apply -f deploy/server-deployment.yaml

# 查看 Deployment 状态
kubectl get deployment -n toolbox k8snet-checker-server

# 查看 Pod 状态
kubectl get pods -n toolbox -l app=k8snet-checker-server
```

#### 1.2 创建 Service

```bash
# 创建服务器 Service
kubectl apply -f deploy/server-service.yaml

# 查看 Service 状态
kubectl get svc -n toolbox k8snet-checker-server

# 查看 Service 详细信息
kubectl describe svc -n toolbox k8snet-checker-server
```

#### 1.3 验证服务器

```bash
# 查看服务器日志
kubectl logs -n toolbox -l app=k8snet-checker-server

# 端口转发测试
kubectl port-forward -n toolbox svc/k8snet-checker-server 8080:8080

# 在另一个终端测试 API
curl http://localhost:8080/api/v1/health
```

### 2. 部署客户端

#### 2.1 创建 DaemonSet

```bash
# 部署客户端 DaemonSet
kubectl apply -f deploy/client-daemonset.yaml

# 查看 DaemonSet 状态
kubectl get daemonset -n toolbox k8snet-checker-client

# 查看所有客户端 Pod
kubectl get pods -n toolbox -l app=k8snet-checker-client -o wide
```

#### 2.2 验证客户端

```bash
# 查看客户端日志
kubectl logs -n toolbox -l app=k8snet-checker-client --tail=50

# 查看特定节点的客户端日志
kubectl logs -n toolbox <client-pod-name>

# 检查客户端健康状态
kubectl exec -n toolbox <client-pod-name> -- wget -qO- http://localhost:6100/health
```

## 配置说明

### 服务器配置

编辑 `deploy/server-deployment.yaml` 修改服务器配置：

```yaml
env:
- name: CACHE_KEY_SECOND
  value: "15"              # 缓存过期时间（秒）
- name: LOG_LEVEL
  value: "info"            # 日志级别: debug/info/warn/error
- name: HTTP_PORT
  value: "8080"            # HTTP 服务端口
- name: REPORT_INTERVAL
  value: "300"             # 报告生成间隔（秒）
```

**配置建议**:
- `CACHE_KEY_SECOND`: 建议设置为心跳间隔的 3 倍，默认 15 秒
- `LOG_LEVEL`: 生产环境使用 `info`，调试时使用 `debug`
- `REPORT_INTERVAL`: 根据需要调整，默认 5 分钟（300 秒）

### 客户端配置

编辑 `deploy/client-daemonset.yaml` 修改客户端配置：

```yaml
env:
- name: SERVER_URL
  value: "http://k8snet-checker-server.toolbox.svc.cluster.local:8080"
- name: HEARTBEAT_INTERVAL
  value: "5"               # 心跳间隔（秒）
- name: TEST_PORT
  value: "22"              # 宿主机测试端口
- name: CUSTOM_SERVICE_NAME
  value: ""                # 自定义服务名称（可选）
- name: CUSTOM_SERVICE_PORT
  value: "80"              # 自定义服务端口（可选）
- name: CLIENT_PORT
  value: "6100"            # 客户端监听端口
- name: LOG_LEVEL
  value: "info"            # 日志级别
```

**配置建议**:
- `HEARTBEAT_INTERVAL`: 建议 5 秒，不要超过 `CACHE_KEY_SECOND` 的 1/3
- `TEST_PORT`: 根据实际环境调整，默认测试 SSH 端口 22
- `CUSTOM_SERVICE_NAME`: 如需测试特定服务，填写服务名称（如 `kubernetes.default.svc.cluster.local`）

### 资源配置

根据集群规模调整资源限制：

**服务器资源**（小规模集群 < 50 节点）:
```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

**服务器资源**（大规模集群 > 50 节点）:
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

**客户端资源**:
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "100m"
```

### 镜像配置

如果使用私有镜像仓库，需要修改镜像地址：

```yaml
# 修改镜像地址
image: <your-registry>/k8snet-checker-server:v1.0.0
imagePullPolicy: IfNotPresent

# 如果需要认证，添加 imagePullSecrets
imagePullSecrets:
- name: registry-secret
```

创建镜像拉取密钥：

```bash
kubectl create secret docker-registry registry-secret \
  --docker-server=<your-registry> \
  --docker-username=<username> \
  --docker-password=<password> \
  --docker-email=<email> \
  -n toolbox
```

## 验证部署

### 1. 检查 Pod 状态

```bash
# 检查所有组件状态
kubectl get all -n toolbox | grep network-checker

# 检查服务器 Pod
kubectl get pods -n toolbox -l app=network-checker-server

# 检查客户端 Pod（应该每个节点一个）
kubectl get pods -n toolbox -l app=network-checker-client -o wide

# 检查 Pod 详细信息
kubectl describe pod -n toolbox <pod-name>
```

### 2. 检查日志

```bash
# 查看服务器日志
kubectl logs -n toolbox -l app=network-checker-server -f

# 查看客户端日志
kubectl logs -n toolbox -l app=network-checker-client --tail=100

# 查看特定 Pod 日志
kubectl logs -n toolbox <pod-name> -f
```

### 3. 测试 API

```bash
# 端口转发
kubectl port-forward -n toolbox svc/k8snet-checker-server 8080:8080 &

# 测试健康检查
curl http://localhost:8080/api/v1/health

# 获取活跃客户端数量
curl http://localhost:8080/api/v1/clients/count

# 获取宿主机 IP 列表
curl http://localhost:8080/api/v1/hosts

# 获取 Pod IP 列表
curl http://localhost:8080/api/v1/pods

# 获取所有测试结果
curl http://localhost:8080/api/v1/results | jq .
```

### 4. 验证网络测试

等待几分钟让客户端完成初始测试，然后查询结果：

```bash
# 获取宿主机互探结果
curl http://localhost:8080/api/v1/test-results/hosts | jq .

# 获取 Pod 互探结果
curl http://localhost:8080/api/v1/test-results/pods | jq .

# 如果配置了自定义服务，查看服务探测结果
curl http://localhost:8080/api/v1/test-results/service | jq .
```

### 5. 查看报告

服务器会每 5 分钟在日志中输出网络连通性报告：

```bash
# 查看服务器日志中的报告
kubectl logs -n toolbox -l app=network-checker-server | grep -A 50 "Network Connectivity Report"
```

## 更新部署

### 更新镜像版本

```bash
# 更新服务器镜像
kubectl set image deployment/k8snet-checker-server \
  server=k8snet-checker-server:v1.1.0 \
  -n toolbox

# 更新客户端镜像
kubectl set image daemonset/k8snet-checker-client \
  client=k8snet-checker-client:v1.1.0 \
  -n toolbox

# 查看更新状态
kubectl rollout status deployment/k8snet-checker-server -n toolbox
kubectl rollout status daemonset/k8snet-checker-client -n toolbox
```

### 更新配置

```bash
# 修改配置文件后重新应用
kubectl apply -f deploy/server-deployment.yaml
kubectl apply -f deploy/client-daemonset.yaml

# 或者直接编辑
kubectl edit deployment k8snet-checker-server -n toolbox
kubectl edit daemonset k8snet-checker-client -n toolbox
```

### 回滚部署

```bash
# 查看历史版本
kubectl rollout history deployment/k8snet-checker-server -n toolbox

# 回滚到上一个版本
kubectl rollout undo deployment/k8snet-checker-server -n toolbox

# 回滚到指定版本
kubectl rollout undo deployment/k8snet-checker-server --to-revision=2 -n toolbox
```

## 卸载

### 完全卸载

```bash
# 删除所有组件
kubectl delete -f deploy/all-in-one.yaml

# 或者分别删除
kubectl delete -f deploy/client-daemonset.yaml
kubectl delete -f deploy/server-service.yaml
kubectl delete -f deploy/server-deployment.yaml

# 验证删除
kubectl get all -n toolbox | grep k8snet-checker
```

### 仅删除客户端

```bash
# 删除客户端 DaemonSet
kubectl delete daemonset k8snet-checker-client -n toolbox

# 验证删除
kubectl get pods -n toolbox -l app=k8snet-checker-client
```

### 仅删除服务器

```bash
# 删除服务器 Deployment 和 Service
kubectl delete deployment k8snet-checker-server -n toolbox
kubectl delete service k8snet-checker-server -n toolbox

# 验证删除
kubectl get all -n toolbox -l app=k8snet-checker-server
```

## 故障排查

### 问题 1: 服务器 Pod 无法启动

**症状**: Pod 处于 `CrashLoopBackOff` 或 `Error` 状态

**排查步骤**:

```bash
# 查看 Pod 状态
kubectl get pods -n toolbox -l app=k8snet-checker-server

# 查看 Pod 详细信息
kubectl describe pod -n toolbox <server-pod-name>

# 查看日志
kubectl logs -n toolbox <server-pod-name>

# 查看上一次运行的日志
kubectl logs -n toolbox <server-pod-name> --previous
```

**常见原因**:
- 端口 8080 已被占用
- 资源不足（内存或 CPU）
- 镜像拉取失败

**解决方案**:
```bash
# 检查端口占用
kubectl get svc -n toolbox | grep 8080

# 增加资源限制
kubectl edit deployment k8snet-checker-server -n toolbox

# 检查镜像
kubectl describe pod -n toolbox <server-pod-name> | grep -A 5 "Events"
```

### 问题 2: 客户端 Pod 无法启动

**症状**: 客户端 Pod 无法启动或频繁重启

**排查步骤**:

```bash
# 查看所有客户端 Pod
kubectl get pods -n toolbox -l app=k8snet-checker-client -o wide

# 查看失败的 Pod
kubectl describe pod -n toolbox <client-pod-name>

# 查看日志
kubectl logs -n toolbox <client-pod-name>

# 检查环境变量
kubectl exec -n toolbox <client-pod-name> -- env | grep -E "NODE_IP|POD_IP|SERVER_URL"

# 测试服务器连接
kubectl exec -n toolbox <client-pod-name> -- wget -qO- http://k8snet-checker-server.toolbox.svc.cluster.local:8080/api/v1/health

# 检查安全上下文
kubectl get daemonset -n toolbox k8snet-checker-client -o yaml | grep -A 10 securityContext
```

### 问题 3: 客户端数量不正确

**症状**: 活跃客户端数量与节点数量不匹配

**排查步骤**:

```bash
# 查看节点数量
kubectl get nodes | wc -l

# 查看客户端 Pod 数量
kubectl get pods -n toolbox -l app=k8snet-checker-client | wc -l

# 查看 DaemonSet 状态
kubectl get daemonset -n toolbox k8snet-checker-client
```

**常见原因**:
- 某些节点有污点，客户端无法调度
- 客户端心跳失败
- 缓存过期时间配置不当

**解决方案**:

```bash
# 检查节点污点
kubectl describe nodes | grep -A 5 Taints

# 检查 DaemonSet 的 tolerations
kubectl get daemonset k8snet-checker-client -n toolbox -o yaml | grep -A 5 tolerations

# 调整缓存过期时间
kubectl set env deployment/k8snet-checker-server CACHE_KEY_SECOND=30 -n toolbox
```

### 问题 4: 网络测试失败

**症状**: 测试结果显示大量 `unreachable` 或 `closed`

**排查步骤**:

```bash
# 查看测试结果
curl http://localhost:8080/api/v1/test-results/hosts | jq .
curl http://localhost:8080/api/v1/test-results/pods | jq .

# 手动测试网络连通性
kubectl exec -n toolbox <client-pod-name> -- ping -c 3 <target-ip>
kubectl exec -n toolbox <client-pod-name> -- nc -zv <target-ip> 22
```

**常见原因**:
- 防火墙规则阻止 ICMP 或 TCP 连接
- 目标端口未开放
- 客户端缺少 `NET_RAW` 权限

**解决方案**:

```bash
# 检查安全上下文
kubectl get daemonset k8snet-checker-client -n toolbox -o yaml | grep -A 10 securityContext

# 确保有 NET_RAW 和 NET_ADMIN 权限
# 已在 client-daemonset.yaml 中配置

# 检查防火墙规则（在节点上执行）
iptables -L -n | grep -E "ICMP|22"
```

### 问题 5: 无法访问 API

**症状**: 端口转发后无法访问 API

**排查步骤**:

```bash
# 检查 Service
kubectl get svc -n toolbox k8snet-checker-server

# 检查端点
kubectl get endpoints -n toolbox k8snet-checker-server

# 测试 Service 内部访问
kubectl run -it --rm debug --image=alpine --restart=Never -n toolbox -- sh
# 在 Pod 内执行
wget -qO- http://k8snet-checker-server.toolbox.svc.cluster.local:8080/api/v1/health
```

**解决方案**:

```bash
# 重新创建 Service
kubectl delete svc k8snet-checker-server -n toolbox
kubectl apply -f deploy/server-service.yaml

# 检查 Service 选择器
kubectl get svc k8snet-checker-server -n toolbox -o yaml | grep -A 3 selector
kubectl get pods -n toolbox -l app=network-checker-server --show-labels
```

## 监控和告警

### 使用 Prometheus 监控

如果集群中部署了 Prometheus，可以添加 ServiceMonitor：

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: network-checker-server
  namespace: toolbox
spec:
  selector:
    matchLabels:
      app: network-checker-server
  endpoints:
  - port: http
    interval: 30s
```

### 日志聚合

使用 Fluentd 或 Filebeat 收集日志：

```yaml
# 添加日志标签
metadata:
  labels:
    logging: "enabled"
```

## 最佳实践

1. **资源限制**: 根据集群规模合理设置资源请求和限制
2. **日志级别**: 生产环境使用 `info` 级别，避免过多日志
3. **心跳间隔**: 保持默认 5 秒，确保及时检测客户端状态
4. **缓存过期**: 设置为心跳间隔的 3 倍，避免误判
5. **监控告警**: 配置 Prometheus 监控和告警规则
6. **定期检查**: 定期查看报告，及时发现网络问题
7. **版本管理**: 使用明确的版本标签，便于回滚
8. **备份配置**: 保存部署配置文件，便于恢复

## 相关文档

- [构建指南](BUILD.md)
- [测试指南](TESTING.md)
- [README](README.md)
