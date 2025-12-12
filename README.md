# K8s Network Checker Helm Repository

这是 K8s Network Checker 的 Helm Chart 仓库。

## 使用方法

### 添加仓库

```bash
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
```

### 安装

```bash
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

### 查看可用版本

```bash
helm search repo k8snet-checker -l
```

## 文档

- [项目主页](https://github.com/yezihack/k8snet-checker)
- [Chart 文档](https://github.com/yezihack/k8snet-checker/tree/main/chart/k8snet-checker)

