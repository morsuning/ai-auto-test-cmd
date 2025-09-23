#!/bin/bash

# 构建脚本：用于交叉编译 API 自动化测试命令行工具 (atc)
# 支持的平台：Windows x86_64, macOS ARM64, Linux ARM64, Linux x86_64

# 设置输出目录
OUTPUT_DIR="./bin"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 获取版本号
if [ -z "$1" ]; then
    echo "请输入版本号 (例如: v1.0.0):"
    read -r VERSION
    if [ -z "$VERSION" ]; then
        echo "❌ 版本号不能为空"
        exit 1
    fi
else
    VERSION="$1"
fi

# 验证版本号格式 (可选的v前缀 + 语义化版本号)
if ! echo "$VERSION" | grep -qE '^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$'; then
    echo "❌ 版本号格式不正确，请使用语义化版本号格式 (例如: v1.0.0 或 1.0.0)"
    exit 1
fi

# 确保版本号以v开头
if [[ ! "$VERSION" =~ ^v ]]; then
    VERSION="v$VERSION"
fi

# 获取构建信息
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建ldflags
LDFLAGS="-s -w -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT"

echo "开始构建 API 自动化测试命令行工具 (atc) $VERSION"
echo "目标平台: Windows amd64, macOS arm64, Linux arm64, Linux amd64"
echo "构建时间: $BUILD_TIME"
echo "Git提交: $GIT_COMMIT"
echo ""

# 构建 Windows amd64 版本
echo "正在构建 Windows amd64 版本..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/atc_windows_amd64_${VERSION}.exe" -ldflags="$LDFLAGS" .
if [ $? -eq 0 ]; then
    echo "✅ Windows amd64 版本构建成功: $OUTPUT_DIR/atc_windows_amd64_${VERSION}.exe"
else
    echo "❌ Windows amd64 版本构建失败"
fi
echo ""

# 构建 macOS ARM 版本
echo "正在构建 macOS ARM 版本..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$OUTPUT_DIR/atc_darwin_arm64_${VERSION}" -ldflags="$LDFLAGS" .
if [ $? -eq 0 ]; then
    echo "✅ macOS ARM 版本构建成功: $OUTPUT_DIR/atc_darwin_arm64_${VERSION}"
else
    echo "❌ macOS ARM 版本构建失败"
fi
echo ""

# 构建 Linux ARM 版本
echo "正在构建 Linux ARM 版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$OUTPUT_DIR/atc_linux_arm64_${VERSION}" -ldflags="$LDFLAGS" .
if [ $? -eq 0 ]; then
    echo "✅ Linux ARM 版本构建成功: $OUTPUT_DIR/atc_linux_arm64_${VERSION}"
else
    echo "❌ Linux ARM 版本构建失败"
fi
echo ""

# 构建 Linux amd64 版本
echo "正在构建 Linux amd64 版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/atc_linux_amd64_${VERSION}" -ldflags="$LDFLAGS" .
if [ $? -eq 0 ]; then
    echo "✅ Linux amd64 版本构建成功: $OUTPUT_DIR/atc_linux_amd64_${VERSION}"
else
    echo "❌ Linux amd64 版本构建失败"
fi
echo ""

echo "构建完成！所有版本已保存到 $OUTPUT_DIR 目录"
echo "Windows amd64: $OUTPUT_DIR/atc_windows_amd64_${VERSION}.exe"
echo "macOS ARM: $OUTPUT_DIR/atc_darwin_arm64_${VERSION}"
echo "Linux ARM: $OUTPUT_DIR/atc_linux_arm64_${VERSION}"
echo "Linux amd64: $OUTPUT_DIR/atc_linux_amd64_${VERSION}"
