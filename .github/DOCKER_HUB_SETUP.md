# Docker Hub 配置指南

本文档说明如何配置 GitHub Actions 以自动构建和推送 Docker 镜像到 Docker Hub。

## 前置要求

1. 拥有 Docker Hub 账号
2. 拥有 GitHub 仓库的管理员权限

## 配置步骤

### 1. 创建 Docker Hub Access Token

1. 登录 [Docker Hub](https://hub.docker.com/)
2. 点击右上角头像 -> **Account Settings**
3. 选择 **Security** 标签
4. 点击 **New Access Token**
5. 输入 Token 描述（例如：`github-actions`）
6. 选择权限：**Read, Write, Delete**
7. 点击 **Generate**
8. **重要**：复制生成的 Token（只显示一次）

### 2. 配置 GitHub Secrets

1. 进入 GitHub 仓库
2. 点击 **Settings** -> **Secrets and variables** -> **Actions**
3. 点击 **New repository secret**
4. 添加以下两个 Secrets：

#### Secret 1: DOCKER_HUB_USERNAME
- **Name**: `DOCKER_HUB_USERNAME`
- **Value**: 你的 Docker Hub 用户名

#### Secret 2: DOCKER_HUB_TOKEN
- **Name**: `DOCKER_HUB_TOKEN`
- **Value**: 刚才复制的 Access Token

### 3. 验证配置

配置完成后，工作流会在以下情况自动触发：

- **推送到 main 或 develop 分支**：构建并推送镜像，标签为分支名
- **创建 tag（如 v1.0.0）**：构建并推送镜像，标签为版本号
- **Pull Request**：仅构建镜像，不推送
- **手动触发**：在 Actions 页面手动运行

## 镜像标签说明

工作流会自动生成以下标签：

### 分支推送
- `main` 分支 -> `latest` 和 `main`
- `develop` 分支 -> `develop`

### 版本标签
- `v1.2.3` -> `1.2.3`, `1.2`, `1`, `latest`
- `v2.0.0` -> `2.0.0`, `2.0`, `2`, `latest`

### Pull Request
- PR #123 -> `pr-123`（仅构建，不推送）

## 支持的架构

镜像支持以下架构：
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM64/aarch64)

Docker 会自动为你的平台拉取正确的架构。

## 本地测试多架构构建

如果想在本地测试多架构构建：

```bash
# 创建并使用 buildx builder
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap

# 构建 Server 镜像（多架构）
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-username/k8snet-checker-server:test \
  -f Dockerfile.server \
  --push \
  .

# 构建 Client 镜像（多架构）
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-username/k8snet-checker-client:test \
  -f Dockerfile.client \
  --push \
  .
```

**注意**：多架构构建必须使用 `--push`，不能使用 `--load`。

## 仅构建单架构（本地测试）

```bash
# 仅构建 amd64
docker buildx build \
  --platform linux/amd64 \
  -t k8snet-checker-server:local \
  -f Dockerfile.server \
  --load \
  .

# 仅构建 arm64
docker buildx build \
  --platform linux/arm64 \
  -t k8snet-checker-server:local \
  -f Dockerfile.server \
  --load \
  .
```

## 查看镜像架构

```bash
# 查看镜像支持的架构
docker buildx imagetools inspect your-username/k8snet-checker-server:latest
```

## 故障排查

### 问题 1: "unauthorized: authentication required"
- 检查 `DOCKER_HUB_USERNAME` 和 `DOCKER_HUB_TOKEN` 是否正确配置
- 确认 Token 有 Write 权限

### 问题 2: "denied: requested access to the resource is denied"
- 确认 Docker Hub 用户名拼写正确
- 确认仓库名称符合 Docker Hub 命名规范

### 问题 3: 构建超时
- 多架构构建需要更长时间，这是正常的
- GitHub Actions 免费版有 6 小时的超时限制

### 问题 4: ARM64 构建失败
- 确保使用了 `setup-qemu-action` 和 `setup-buildx-action`
- 检查 Dockerfile 中是否正确使用了 `TARGETARCH` 变量

## 工作流文件说明

工作流文件位于：`.github/workflows/docker-build.yml`

主要特性：
- ✅ 多架构支持（amd64 + arm64）
- ✅ 自动标签管理
- ✅ 构建缓存优化
- ✅ PR 安全检查（仅构建不推送）
- ✅ 手动触发支持

## 更新镜像名称

如果需要修改镜像名称，编辑 `.github/workflows/docker-build.yml`：

```yaml
env:
  DOCKER_HUB_USERNAME: ${{ secrets.DOCKER_HUB_USERNAME }}
  SERVER_IMAGE_NAME: your-custom-server-name  # 修改这里
  CLIENT_IMAGE_NAME: your-custom-client-name  # 修改这里
```

## 参考链接

- [Docker Hub Access Tokens](https://docs.docker.com/docker-hub/access-tokens/)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/)
- [Multi-platform images](https://docs.docker.com/build/building/multi-platform/)
