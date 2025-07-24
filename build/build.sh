#!/bin/bash

# 构建脚本：用于交叉编译 API 自动化测试命令行工具 (atc)
# 支持的平台：Windows x86, macOS ARM, Linux ARM, Linux x86

# 设置版本号
VERSION="1.2.2"

# 设置输出目录
OUTPUT_DIR="./bin"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

echo "开始构建 API 自动化测试命令行工具 (atc) v$VERSION"
echo "目标平台: Windows amd64, macOS arm64, Linux arm64, Linux amd64"
echo ""

# 构建 Windows amd64 版本
echo "正在构建 Windows amd64 版本..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/atc_windows_amd64.exe" -ldflags="-s -w" .
if [ $? -eq 0 ]; then
    echo "✅ Windows amd64 版本构建成功: $OUTPUT_DIR/atc_windows_amd64.exe"
else
    echo "❌ Windows amd64 版本构建失败"
fi
echo ""

# 构建 macOS ARM 版本
echo "正在构建 macOS ARM 版本..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$OUTPUT_DIR/atc_darwin_arm64" -ldflags="-s -w" .
if [ $? -eq 0 ]; then
    echo "✅ macOS ARM 版本构建成功: $OUTPUT_DIR/atc_darwin_arm64"
else
    echo "❌ macOS ARM 版本构建失败"
fi
echo ""

# 构建 Linux ARM 版本
echo "正在构建 Linux ARM 版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$OUTPUT_DIR/atc_linux_arm64" -ldflags="-s -w" .
if [ $? -eq 0 ]; then
    echo "✅ Linux ARM 版本构建成功: $OUTPUT_DIR/atc_linux_arm64"
else
    echo "❌ Linux ARM 版本构建失败"
fi
echo ""

# 构建 Linux amd64 版本
echo "正在构建 Linux amd64 版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/atc_linux_amd64" -ldflags="-s -w" .
if [ $? -eq 0 ]; then
    echo "✅ Linux amd64 版本构建成功: $OUTPUT_DIR/atc_linux_amd64"
else
    echo "❌ Linux amd64 版本构建失败"
fi
echo ""
