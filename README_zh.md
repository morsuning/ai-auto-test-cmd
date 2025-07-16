# ATC - API自动化测试命令行工具

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)](#安装)

> **语言**: 中文文档 | [English](README.md)

ATC (API Test Command) 是一个功能强大的API自动化测试命令行工具，专为简化API测试流程而设计。支持智能测试用例生成、批量接口测试、多种数据格式处理等功能。

## ✨ 核心特性

### 🎯 智能测试用例生成
- **本地生成**：基于正例输入快速生成多样化测试用例
- **智能约束系统**：支持11种约束类型，生成真实有效的中文测试数据
- **多格式支持**：支持JSON和XML格式的输入输出
- **数据变化规则**：数值50%波动，字符串10%长度变化

### 🤖 AI智能生成
- **Dify API集成**：通过Dify Chatflow API生成智能测试用例
- **配置文件支持**：支持从config.toml文件读取API配置
- **多种输入方式**：支持命令行输入和文件输入
- **流式响应处理**：实时显示生成进度
- **智能解析**：自动解析API响应并生成测试用例

### 🚀 批量接口测试
- **多HTTP方法**：支持POST、GET等HTTP请求方法
- **多种鉴权**：Bearer Token、Basic Auth、API Key等
- **自定义请求头**：灵活添加HTTP头信息
- **并发执行**：提高测试执行效率
- **结果保存**：支持CSV格式结果导出

### 🛡️ 配置验证
- **格式验证**：约束配置文件完整性检查
- **内容验证**：数据类型和范围合理性验证
- **错误报告**：详细的错误信息和位置定位

## 📦 安装

### 系统要求

**重要提示**：本工具在输出中使用了emoji表情符号（✅、❌、🔍等）以提供更好的用户体验。为了正确显示这些字符，您的终端环境必须支持UTF-8编码。

- **Windows系统**：建议使用Windows Terminal、PowerShell Core，或在命令提示符中启用UTF-8支持
- **macOS/Linux系统**：大多数现代终端默认支持UTF-8编码
- **SSH/远程连接**：确保SSH客户端和服务器都支持UTF-8编码

如果emoji字符显示异常，请检查您的终端编码设置。

### 预编译二进制文件

从 [Releases](https://github.com/morsuning/ai-auto-test-cmd/releases) 页面下载适合您系统的预编译版本：

- **Windows (amd64)**: `atc-windows-amd64.exe`
- **macOS (ARM)**: `atc-darwin-arm64`
- **Linux (ARM)**: `atc-linux-arm64`
- **Linux (amd64)**: `atc-linux-amd64`

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/morsuning/ai-auto-test-cmd.git
cd ai-auto-test-cmd

# 编译
go build -o atc

# 或使用构建脚本
# Windows
build\build.bat

# macOS/Linux
bash build/build.sh
```

## 🚀 快速开始

### 1. 生成测试用例

```bash
# 从JSON正例生成测试用例
atc local-gen '{"name":"张三","age":25,"email":"test@example.com"}' --json --num 10

# 从XML正例生成测试用例
atc local-gen '<user><name>张三</name><age>25</age></user>' --xml --num 5

# 从文件读取正例并生成
atc local-gen -f examples/json_example.json --json --num 20

# 使用智能约束系统生成
atc local-gen -f examples/json_example.json --json --num 10 --constraints
```

### 2. 执行接口测试

```bash
# POST请求发送JSON数据
atc request -u https://api.example.com/users -m post -f testcases.csv --json

# GET请求
atc request -u https://api.example.com/users -m get -f testcases.csv --json

# 使用Bearer Token鉴权
atc request -u https://api.example.com/users -m post -f testcases.csv --json --auth-bearer "your_token"

# 添加自定义请求头
atc request -u https://api.example.com/users -m post -f testcases.csv --json --header "X-API-Key: key123"

# 保存测试结果
atc request -u https://api.example.com/users -m post -f testcases.csv --json -s results.csv
```

### 3. AI智能生成测试用例

```bash
# 使用默认配置文件生成测试用例
atc dify-gen --xml --raw "<user><name>张三</name></user>" -n 5

# 指定配置文件生成测试用例
atc dify-gen -c my-config.toml --json --raw '{"name":"test"}' -n 3

# 从文件读取输入并生成
atc dify-gen --json -f input.json -n 5 --debug

# 使用自定义提示词文件生成测试用例
atc dify-gen --xml --raw "<user><name>张三</name></user>" --prompt custom_prompt.txt -n 3

# 结合配置文件和提示词文件
atc dify-gen -c my-config.toml --json --raw '{"name":"test"}' --prompt prompt.txt -n 5

# 显式指定API参数（覆盖配置文件）
atc dify-gen -u https://api.dify.ai/v1 --api-key your_key --xml --raw "<test/>" -n 2
```

### 4. 验证约束配置

```bash
# 验证默认配置文件
atc validate

# 验证指定配置文件
atc validate my-constraints.toml

# 显示详细验证信息
atc validate --verbose
```

## 📋 命令详解

### `dify-gen` - AI智能生成测试用例

通过Dify Chatflow API生成智能测试用例。

```bash
atc dify-gen [flags]
```

**主要参数：**
- `--url, -u`: Dify API URL（可选，可从配置文件读取）
- `--api-key`: Dify API Key（可选，可从配置文件读取）
- `--config, -c`: 配置文件路径（默认：config.toml）
- `--json`: 指定JSON格式
- `--xml`: 指定XML格式
- `--raw`: 直接输入原始数据
- `--file, -f`: 从文件读取输入
- `--prompt`: 自定义提示词文件路径（可选，文件必须是UTF-8编码）
- `--num, -n`: 生成数量（默认5）
- `--output, -o`: 输出文件路径
- `--debug, -d`: 启用调试模式

**配置文件支持：**

创建 `config.toml` 文件：
```toml
[dify]
url = "https://api.dify.ai/v1/chatflows/xxx/run"
api_key = "app-xxxxxxxxxx"
```

**参数优先级：**
1. 命令行参数（最高优先级）
2. 配置文件参数
3. 如果都未指定则报错

**示例：**
```bash
# 使用默认配置文件
atc dify-gen --json --raw '{"name":"test"}' -n 3

# 指定配置文件
atc dify-gen -c my-config.toml --xml --raw "<test/>" -n 5

# 覆盖配置文件中的参数
atc dify-gen --api-key new_key --json -f input.json -n 2

# 从文件读取并启用调试
atc dify-gen -f input.xml --xml -n 3 --debug
```

### `local-gen` - 本地生成测试用例

基于正例输入生成多样化的测试用例。

```bash
atc local-gen [正例输入] [flags]
```

**主要参数：**
- `--json`: 指定JSON格式
- `--xml`: 指定XML格式
- `--num, -n`: 生成数量（默认10）
- `--file, -f`: 从文件读取正例
- `--output, -o`: 输出文件路径
- `--constraints`: 启用智能约束系统
- `--constraints-file`: 指定约束配置文件

**示例：**
```bash
# 生成10个JSON测试用例
atc local-gen '{"name":"张三","age":25}' --json -n 10

# 使用约束系统生成真实数据
atc local-gen '{"name":"张三","phone":"13800138000"}' --json --constraints -n 5

# 从文件生成并保存到指定位置
atc local-gen -f input.json --json -n 20 -o testcases.csv
```

### `request` - 批量接口测试

基于CSV测试用例文件批量执行HTTP请求。

```bash
atc request -u [URL] -m [METHOD] -f [CSV文件] [flags]
```

**主要参数：**
- `--url, -u`: 目标接口URL（必需）
  - **注意**：如果URL未包含协议（http://或https://），系统将自动添加http://前缀
  - 示例：`localhost:8080/user` 将被处理为 `http://localhost:8080/user`
- `--method, -m`: HTTP方法（post/get）
- `--file, -f`: CSV测试用例文件（必需）
- `--json`: JSON格式请求体
- `--xml`: XML格式请求体
- `--save, -s`: 保存结果到文件
- `--timeout`: 请求超时时间（默认30秒）
- `--debug`: 启用调试模式

**鉴权参数：**
- `--auth-bearer`: Bearer Token认证
- `--auth-basic`: Basic Auth认证（格式：username:password）
- `--header`: 自定义HTTP头（可多次使用）

**示例：**
```bash
# 基本POST请求
atc request -u https://api.example.com/users -m post -f users.csv --json

# 本地服务器（自动添加http://协议）
atc request -u localhost:8080/api/test -m post -f users.csv --json

# 使用鉴权和自定义头
atc request -u https://api.example.com/users -m post -f users.csv --json \
  --auth-bearer "eyJhbGciOiJIUzI1NiIs..." \
  --header "X-Request-ID: 12345" \
  --header "X-Client-Version: 1.0"

# GET请求（自动转换为查询参数）
atc request -u https://api.example.com/users -m get -f users.csv --json

# 启用调试模式并保存结果
atc request -u https://api.example.com/users -m post -f users.csv --json --debug -s results.csv
```

### `validate` - 配置验证

验证约束配置文件的格式和内容正确性。

```bash
atc validate [配置文件] [flags]
```

**主要参数：**
- `--verbose, -v`: 显示详细验证信息

**示例：**
```bash
# 验证默认配置
atc validate

# 验证指定配置文件
atc validate my-constraints.toml

# 显示详细信息
atc validate --verbose
```

## 🎯 智能约束系统

智能约束系统是ATC的核心特性，能够根据字段名自动识别并生成真实有效的测试数据。

### 支持的约束类型

| 约束类型 | 说明 | 示例字段名 | 生成示例 |
|---------|------|-----------|----------|
| `date` | 日期类型 | date, time, created_at | 20230101 |
| `chinese_name` | 中文姓名 | name, username, author | 周桂兰 |
| `phone` | 手机号码 | phone, mobile, tel | 17234495798 |
| `email` | 邮箱地址 | email, mail | test473@189.cn |
| `chinese_address` | 中文地址 | address, location | 武汉市武昌区中南路99号 |
| `id_card` | 身份证号 | id_card, identity | 500101198909148195 |
| `integer` | 整数类型 | age, count, quantity | 64 |
| `float` | 浮点数类型 | price, amount, rate | 161782.59 |

### 配置文件示例

创建 `constraints.toml` 文件：

```toml
# 日期字段约束
[date]
type = "date"
format = "20060102"  # Go时间格式
min_date = "20200101"
max_date = "20301231"
description = "日期字段，格式为YYYYMMDD"

# 姓名字段约束
[name]
type = "chinese_name"
description = "中文姓名"

# 年龄字段约束
[age]
type = "integer"
min = 1
max = 120
description = "年龄范围1-120"

# 价格字段约束
[price]
type = "float"
min = 0.01
max = 999999.99
precision = 2
description = "价格字段，保留2位小数"

# 内置数据集
[builtin_data]
first_names = ["张", "王", "李", "赵", "刘"]
last_names = ["伟", "芳", "娜", "敏", "静"]
addresses = ["北京市朝阳区建国门外大街1号", "上海市浦东新区陆家嘴环路1000号"]
email_domains = ["qq.com", "163.com", "126.com", "gmail.com"]
```

### 生成效果对比

**使用约束系统前（随机变化）：**
```json
{"date":"27388202","name":"p","age":18,"phone":"11684695289","email":"1haDgsai8xOmpyU.C0m","price":122,"address":"K京w区"}
```

**使用约束系统后（智能约束）：**
```json
{"date":"20230101","name":"周桂兰","age":64,"phone":"17234495798","email":"test473@189.cn","price":161782.59,"address":"武汉市武昌区中南路99号"}
```

## 📁 CSV文件格式

### 生成阶段格式

- **JSON格式**：单列CSV，列名为"JSON"，每行一个JSON字符串
- **XML格式**：单列CSV，列名为"XML"，每行一个XML字符串

### 测试阶段格式识别

- **单列JSON**：列名为"JSON"，直接使用JSON内容作为请求体
- **单列XML**：列名为"XML"，直接使用XML内容作为请求体
- **多列格式**：将各列数据组合为JSON对象
- **GET请求**：仅支持JSON格式，自动转换为查询参数

## 🔧 高级功能

### 调试模式

使用 `--debug` 参数启用详细的调试输出：

```bash
atc request -u https://api.example.com/users -m post -f users.csv --json --debug
```

调试模式会显示：
- 每个请求的详细信息（URL、方法、头部、请求体）
- 完整的响应信息（状态码、响应时间、响应体）
- 格式化的JSON响应内容

### 并发控制

系统自动根据测试用例数量调整并发数，提高执行效率的同时避免对目标服务器造成过大压力。

### 错误处理

- 详细的错误信息提示
- 多错误批量报告
- 错误位置精确定位
- 友好的用户提示

### XML编码支持

**重要说明**：Go标准库的XML处理包（`encoding/xml`）对XML文档编码有以下限制：

- **仅支持UTF-8编码**：标准库只能正确解析UTF-8编码的XML文档
- **不支持其他编码**：对于GBK、GB2312、ISO-8859-1等非UTF-8编码的XML文档，标准库无法直接处理
- **编码声明被忽略**：即使XML文档声明了`<?xml version="1.0" encoding="GBK"?>`，标准库也会按UTF-8处理

**ATC的解决方案**：

1. **自动编码检测**：使用`golang.org/x/text/encoding`包检测XML文档的实际编码
2. **编码转换**：将非UTF-8编码的XML文档自动转换为UTF-8编码后再进行解析
3. **支持的编码格式**：
   - UTF-8（原生支持）
   - GBK/GB2312（中文编码）
   - ISO-8859-1（西欧编码）
   - 其他常见编码格式

**使用建议**：

- **推荐使用UTF-8编码**：为获得最佳性能和兼容性，建议使用UTF-8编码的XML文档
- **非UTF-8编码处理**：工具会自动处理非UTF-8编码，但可能会有轻微的性能开销
- **编码声明一致性**：确保XML文档的编码声明与实际文件编码一致，避免解析错误

## 📊 示例项目

`examples/` 目录包含了完整的使用示例：

- `constraints.toml`: 约束配置示例
- `json_example.json`: JSON正例输入示例
- `xml_example.xml`: XML正例输入示例
- `input.xml`: 复杂XML结构示例

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [问题反馈](https://github.com/morsuning/ai-auto-test-cmd/issues)
- [功能请求](https://github.com/morsuning/ai-auto-test-cmd/issues/new?template=feature_request.md)
- [English Documentation](README.md)

---

**ATC** - 让API测试更简单、更智能、更高效！

## 许可证

[LICENSE](LICENSE)