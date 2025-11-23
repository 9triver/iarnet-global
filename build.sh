#!/bin/bash

# Iarnet Global 镜像构建脚本
# 使用方法: ./build.sh [tag_name]

set -e

# 默认镜像标签
DEFAULT_TAG="iarnet-global:latest"
IMAGE_TAG="${1:-$DEFAULT_TAG}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}开始构建 Iarnet Global 镜像（前后端一体化）...${NC}"

# 检查是否在正确的目录
if [ ! -f "Dockerfile" ]; then
    echo -e "${RED}错误: 在当前目录找不到 Dockerfile${NC}"
    echo "请确保在 iarnet-global 项目根目录下运行此脚本"
    exit 1
fi

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}当前构建目录: $SCRIPT_DIR${NC}"
echo -e "${YELLOW}构建镜像标签: $IMAGE_TAG${NC}"

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}错误: Docker daemon 未运行${NC}"
    exit 1
fi

# 构建 Docker 镜像
echo -e "${YELLOW}开始 Docker 构建...${NC}"
docker build -t "$IMAGE_TAG" .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Docker 镜像构建成功!${NC}"
    echo -e "${GREEN}镜像标签: $IMAGE_TAG${NC}"
    
    # 显示镜像信息
    echo -e "${YELLOW}镜像信息:${NC}"
    docker images "$IMAGE_TAG"
    
    echo -e "${YELLOW}运行示例:${NC}"
    echo "docker run -d --name iarnet-global \\"
    echo "  -v \$(pwd)/data:/app/data \\"
    echo "  -p 3000:3000 \\"
    echo "  -p 8080:8080 \\"
    echo "  -p 50010:50010 \\"
    echo "  $IMAGE_TAG"
    echo ""
    echo -e "${YELLOW}注意:${NC}"
    echo -e "${YELLOW}  - 前端访问: http://localhost:3000${NC}"
    echo -e "${YELLOW}  - 后端 API: http://localhost:8080${NC}"
    echo -e "${YELLOW}  - RPC 服务: localhost:50010${NC}"
else
    echo -e "${RED}❌ Docker 镜像构建失败!${NC}"
    exit 1
fi

