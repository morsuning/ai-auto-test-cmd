# API 自动化测试命令行工具 (atc)

这是一个用于 API 自动化测试的命令行工具，支持通过 Dify Workflow API 生成测试用例，也支持本地生成测试用例，并能批量执行测试请求。

## 功能特点

- **测试用例生成**
  - 通过 Dify API 生成测试用例
  - 本地生成测试用例（基于正例报文）
  - 支持 XML 和 JSON 格式

- **测试执行**
  - 批量请求目标系统接口
  - 支持 POST 和 GET 等 HTTP 方法
  - 结果可直接显示或保存到文件

## 安装

### 下载预编译版本

我们提供了多平台的预编译版本，您可以直接下载对应平台的可执行文件：

- Windows x86: `atc_windows_386.exe`
- macOS ARM: `atc_darwin_arm64`
- Linux ARM: `atc_linux_arm64`
- Linux x86: `atc_linux_386`

### 从源码构建

如果您想从源码构建，请确保已安装 Go 1.24 或更高版本，然后按照以下步骤操作：

#### 使用构建脚本（推荐）

我们提供了构建脚本，可以一键构建所有平台的可执行文件：

**在 macOS/Linux 上：**

```bash
# 赋予脚本执行权限
chmod +x build.sh

# 执行构建脚本
./build.sh
```

**在 Windows 上：**

```cmd
# 执行构建脚本
build.bat
```

构建完成后，可执行文件将保存在 `dist` 目录中。

#### 手动构建

如果您只需要构建特定平台的版本，可以使用以下命令：

```bash
# 构建当前平台版本
go build -o atc .

# 构建 Windows x86 版本
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o atc_windows_386.exe .

# 构建 macOS ARM 版本
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o atc_darwin_arm64 .

# 构建 Linux ARM 版本
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o atc_linux_arm64 .

# 构建 Linux x86 版本
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o atc_linux_386 .
```

## 使用方法

### 生成测试用例

#### 通过 Dify API 生成测试用例

```bash
# 根据正例 XML 报文生成 10 条测试用例
atc gen -u https://xxx.dify.com/xxx/xxx -xml -raw "xxxx" -n 10

# 根据正例 JSON 报文生成 20 条测试用例
atc gen -u https://xxx.dify.com/xxx/xxx -json -raw "xxxx" -n 20
```

#### 本地生成测试用例

```bash
# 本地根据正例 XML 报文生成 10 条测试用例
atc local-gen -xml -raw "xxxx" -n 10

# 本地根据正例 JSON 报文生成 15 条测试用例
atc local-gen -json -raw "xxxx" -n 15

# 本地根据正例 JSON 报文生成 5 条测试用例并保存到指定文件
atc local-gen -json -raw "xxxx" -n 5 -o test_data.csv
```

### 执行测试请求

```bash
# 根据测试用例文件 xxx.csv，批量使用 POST 方法请求目标系统 HTTP 接口
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv

# 根据测试用例文件 xxx.csv，批量使用 GET 方法请求目标系统 HTTP 接口，结果保存至当前目录
atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv -s

# 根据测试用例文件 xxx.csv，批量使用 GET 方法请求目标系统 HTTP 接口，结果保存至指定目录及文件
atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv -s /xxx/tool/result.csv
```

## 许可证

[LICENSE](LICENSE)