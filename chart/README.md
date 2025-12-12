# K8s Network Checker Helm Repository

这是 K8s Network Checker 的 Helm Chart 仓库。

## 快速开始

### 安装

```bash
# 从本地安装
helm install k8snet-checker ./k8snet-checker -n kube-system

# 查看状态
kubectl get pods -n kube-system -l app.kubernetes.io/name=k8snet-checker
```

### 卸载

```bash
helm uninstall k8snet-checker -n kube-system
```

## 文档

详细文档请查看 [k8snet-checker/README.md](k8snet-checker/README.md)

## Chart 列表

- [k8snet-checker](k8snet-checker/) - Kubernetes 集群网络连通性检测工具

## 版本

- Chart Version: 1.0.0
- App Version: 1.0.0

## 维护者

- sgfoot <freeit@126.com>
