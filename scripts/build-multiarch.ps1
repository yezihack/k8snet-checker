# PowerShell 脚本：构建多架构 Docker 镜像
# 作者：K8s Network Checker Team
# 日期：2025-11-28

param(
    [Parameter(Mandatory=$true, HelpMessage="Docker Hub 用户名")]
    [string]$Username,
    
    [Parameter(HelpMessage="镜像标签")]
    [string]$Tag = "latest",
    
    [Parameter(HelpMessage="推送到 Docker Hub")]
    [switch]$Push,
    
    [Parameter(HelpMessage="仅构建 Server 镜像")]
    [switch]$ServerOnly,
    
    [Parameter(HelpMessage="仅构建 Client 镜像")]
    [switch]$ClientOnly,
    
    [Parameter(HelpMessage="指定架构")]
    [string]$Arch = "linux/amd64,linux/arm64",
    
    [Parameter(HelpMessage="显示帮助信息")]
    [switch]$Help
)

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $Message" -ForegroundColor $Color
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput -Message "[INFO] $Message" -Color Green
}

function Write-Error-Custom {
    param([string]$Message)
    Write-ColorOutput -Message "[ERROR] $Message" -Color Red
}

function Write-Warn {
    param([string]$Message)
    Write-ColorOutput -Message "[WARN] $Message" -Color Yellow
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput -Message "[SUCCESS] $Message" -Color Green
}

# 显示帮助信息
function Show-Help {
    Write-Host @"
使用方法: .\build-multiarch.ps1 [参数]

参数:
    -Username       Docker Hub 用户名（必需）
    -Tag            镜像标签（默认：latest）
    -Push           推送到 Docker Hub
    -ServerOnly     仅构建 Server 镜像
    -ClientOnly     仅构建 Client 镜像
    -Arch           指定架构（默认：linux/amd64,linux/arm64）
    -Help           显示帮助信息

示例:
    # 构建并推送到 Docker Hub
    .\build-multiarch.ps1 -Username myusername -Tag v1.0.0 -Push

    # 仅构建 Server 镜像（不推送）
    .\build-multiarch.ps1 -Username myusername -ServerOnly

    # 构建单架构镜像
    .\build-multiarch.ps1 -Username myusername -Arch linux/amd64

    # 构建并推送特定标签
    .\build-multiarch.ps1 -Username myusername -Tag develop -Push
"@
    exit 0
}

# 检查命令是否存在
function Test-Command {
    param([string]$Command)
    
    $exists = $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
    if (-not $exists) {
        Write-Error-Custom "未找到必需的命令: $Command"
        exit 1
    }
}

# 检查 Docker Buildx
function Test-Buildx {
    try {
        docker buildx version | Out-Null
    } catch {
        Write-Error-Custom "Docker Buildx 未安装或未启用"
        Write-Info "请运行: docker buildx create --use"
        exit 1
    }
}

# 创建 Buildx builder
function Initialize-Builder {
    $builderName = "multiarch-builder"
    
    $builderExists = docker buildx inspect $builderName 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Info "创建 Buildx builder: $builderName"
        docker buildx create --name $builderName --use
        docker buildx inspect --bootstrap
    } else {
        Write-Info "使用现有 Buildx builder: $builderName"
        docker buildx use $builderName
    }
}

# 构建镜像
function Build-Image {
    param(
        [string]$Dockerfile,
        [string]$ImageName,
        [string]$Tag,
        [string]$Platforms,
        [bool]$ShouldPush
    )
    
    $fullImage = "${ImageName}:${Tag}"
    
    Write-Info "开始构建镜像: $fullImage"
    Write-Info "架构: $Platforms"
    Write-Info "Dockerfile: $Dockerfile"
    
    $buildArgs = @(
        "buildx", "build",
        "--platform", $Platforms,
        "-t", $fullImage,
        "-f", $Dockerfile
    )
    
    if ($ShouldPush) {
        $buildArgs += "--push"
        Write-Info "构建完成后将推送到 Docker Hub"
    } else {
        Write-Warn "多架构构建不会加载到本地 Docker"
        Write-Warn "如需本地测试，请使用单架构构建: -Arch linux/amd64"
    }
    
    $buildArgs += "."
    
    & docker $buildArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "镜像构建成功: $fullImage"
        if ($ShouldPush) {
            Write-Success "镜像已推送到 Docker Hub"
        }
    } else {
        Write-Error-Custom "镜像构建失败: $fullImage"
        exit 1
    }
}

# 主函数
function Main {
    # 显示帮助
    if ($Help) {
        Show-Help
    }
    
    # 检查依赖
    Test-Command "docker"
    Test-Buildx
    
    # 设置 builder
    Initialize-Builder
    
    # 确定要构建的镜像
    $buildServer = -not $ClientOnly
    $buildClient = -not $ServerOnly
    
    # 镜像名称
    $serverImage = "$Username/k8snet-checker-server"
    $clientImage = "$Username/k8snet-checker-client"
    
    Write-Info "=========================================="
    Write-Info "Docker 多架构镜像构建"
    Write-Info "=========================================="
    Write-Info "用户名: $Username"
    Write-Info "标签: $Tag"
    Write-Info "架构: $Arch"
    Write-Info "推送: $Push"
    Write-Info "=========================================="
    
    # 构建 Server 镜像
    if ($buildServer) {
        Build-Image -Dockerfile "Dockerfile.server" `
                    -ImageName $serverImage `
                    -Tag $Tag `
                    -Platforms $Arch `
                    -ShouldPush $Push
    }
    
    # 构建 Client 镜像
    if ($buildClient) {
        Build-Image -Dockerfile "Dockerfile.client" `
                    -ImageName $clientImage `
                    -Tag $Tag `
                    -Platforms $Arch `
                    -ShouldPush $Push
    }
    
    Write-Success "=========================================="
    Write-Success "所有镜像构建完成！"
    Write-Success "=========================================="
    
    if ($Push) {
        Write-Info "查看镜像信息:"
        if ($buildServer) {
            Write-Info "  docker buildx imagetools inspect ${serverImage}:${Tag}"
        }
        if ($buildClient) {
            Write-Info "  docker buildx imagetools inspect ${clientImage}:${Tag}"
        }
    } else {
        Write-Info "镜像未推送到 Docker Hub"
        Write-Info "如需推送，请添加 -Push 参数"
    }
}

# 执行主函数
Main
