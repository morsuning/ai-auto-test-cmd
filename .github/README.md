# GitHub Actions 快速使用指南

## 📋 概述

本项目已配置完整的 GitHub Actions 自动化工作流，包括持续集成 (CI) 和自动化发布 (Release)。

## 🚀 快速开始

### 1. 推送代码触发 CI
当您向 `main` 或 `develop` 分支推送代码时，会自动触发 CI 流程：

```bash
git add .
git commit -m "feat: 新增功能"
git push origin main
```

CI 会自动执行：
- ✅ 运行所有测试
- ✅ 代码格式检查
- ✅ 静态分析
- ✅ 多平台构建验证

### 2. 创建 Release
要发布新版本，只需创建并推送版本标签：

```bash
# 创建版本标签
git tag -a v1.2.3 -m "Release v1.2.3

新增功能：
- 功能描述1
- 功能描述2

修复问题：
- 问题修复1
- 问题修复2"

# 推送标签
git push origin v1.2.3
```

这会自动：
- 🔍 运行完整测试套件
- 🏗️ 构建多平台二进制文件
- 📦 创建 GitHub Release
- 📝 生成 Release Notes
- 🔐 生成校验和文件

## 📁 文件结构

```
.github/
└── workflows/
    ├── ci.yml      # 持续集成工作流
    └── release.yml # 自动发布工作流
```

## 🔧 配置说明

### CI 工作流 (ci.yml)
- **触发条件**：推送到 main/develop 分支，或向 main 分支提交 PR
- **Go 版本**：1.24
- **执行内容**：测试、格式检查、静态分析、构建验证

### Release 工作流 (release.yml)
- **触发条件**：推送 `v*` 格式的标签
- **构建平台**：Windows, Linux, macOS (amd64 + arm64)
- **输出文件**：二进制文件 + 校验和文件

## 📊 监控构建状态

1. 访问 GitHub 仓库
2. 点击 "Actions" 标签页
3. 查看工作流执行状态
4. 点击具体的工作流查看详细日志

## 🛠️ 故障排除

### 常见问题

**测试失败**：
```bash
# 本地运行测试
go test -v ./...
```

**格式检查失败**：
```bash
# 格式化代码
go fmt ./...
```

**构建失败**：
```bash
# 本地测试构建
go build -o atc .
```

## 📚 详细文档

- [Release操作手册](../docs/Release操作手册.md) - 完整的发布流程说明
- [项目需求文档](../docs/项目需求文档.md) - 项目功能和技术规范

---

**提示**：首次使用时，请确保仓库设置中启用了 Actions 功能。