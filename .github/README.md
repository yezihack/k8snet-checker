# GitHub Actions 工作流

本目录包含 K8s Network Checker 项目的 CI/CD 工作流配置。

## 工作流列表

### docker-build.yml

自动构建和推送多架构 Docker 镜像到 Docker Hub。

**触发条件:**

- 推送到 `main` 或 `develop` 分支
- 创建 Pull Request 到 `main` 分支
- 创建版本标签（如 `v1.0.0`）
- 手动触发

**构建架构:**

- `linux/amd64`
- `linux/arm64`

**生成的镜像:**

- `<username>/k8snet-checker-server`
- `<username>/k8snet-checker-client`

**标签策略:**

| 触发事件 | 生成的标签 |
|---------|-----------|
| 推送到 `main` | `latest`, `main` |
| 推送到 `develop` | `develop` |
| 创建 `v1.2.3` | `1.2.3` `latest` |
| PR #123 | `pr-123` (仅构建) |

## 配置说明

详细的配置步骤请参考：[DOCKER_HUB_SETUP.md](DOCKER_HUB_SETUP.md)

### 必需的 Secrets

- `DOCKER_HUB_USERNAME`: Docker Hub 用户名
- `DOCKER_HUB_TOKEN`: Docker Hub Access Token

## 本地测试

### 使用 act 工具本地运行工作流

```bash
# 安装 act
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# 运行工作流
act push -s DOCKER_HUB_USERNAME=your-username -s DOCKER_HUB_TOKEN=your-token
```

### 使用构建脚本

```bash
# Linux/macOS
./scripts/build-multiarch.sh -u your-username -t test -p

# Windows
.\scripts\build-multiarch.ps1 -Username your-username -Tag test -Push
```

## 工作流优化

### 构建缓存

工作流使用 GitHub Actions 缓存来加速构建：

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

### 并行构建

Server 和 Client 镜像在同一个 job 中顺序构建，但可以通过修改工作流实现并行构建：

```yaml
jobs:
  build-server:
    # Server 构建配置
  
  build-client:
    # Client 构建配置
```

## 故障排查

### 构建失败

1. 检查 Secrets 是否正确配置
2. 查看 Actions 日志获取详细错误信息
3. 验证 Dockerfile 语法是否正确

### 推送失败

1. 确认 Docker Hub Token 有 Write 权限
2. 检查镜像名称是否符合规范
3. 验证网络连接是否正常

### 多架构构建问题

1. 确保使用了 `setup-qemu-action` 和 `setup-buildx-action`
2. 检查 Dockerfile 中的 `TARGETARCH` 变量使用是否正确
3. 验证基础镜像是否支持目标架构

## 扩展工作流

### 添加代码检查

```yaml
- name: Run tests
  run: go test ./...

- name: Run linter
  run: golangci-lint run
```

### 添加安全扫描

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ steps.meta-server.outputs.tags }}
    format: 'sarif'
    output: 'trivy-results.sarif'
```

### 添加通知

```yaml
- name: Send notification
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## 参考资源

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Docker Buildx Action](https://github.com/docker/build-push-action)
- [Docker Metadata Action](https://github.com/docker/metadata-action)
- [QEMU Action](https://github.com/docker/setup-qemu-action)
