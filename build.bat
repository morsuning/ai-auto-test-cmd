@echo off
REM 构建脚本：用于交叉编译 API 自动化测试命令行工具 (atc)
REM 支持的平台：Windows x86, macOS ARM, Linux ARM, Linux x86

REM 设置版本号
set VERSION=1.0.0

REM 设置输出目录
set OUTPUT_DIR=.\dist

REM 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo 开始构建 API 自动化测试命令行工具 (atc) v%VERSION%
echo 目标平台: Windows x86, macOS ARM, Linux ARM, Linux x86
echo.

REM 构建 Windows x86 版本
echo 正在构建 Windows x86 版本...
set GOOS=windows
set GOARCH=386
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_windows_386.exe" -ldflags="-s -w" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ Windows x86 版本构建成功: %OUTPUT_DIR%\atc_windows_386.exe
) else (
    echo ✗ Windows x86 版本构建失败
)
echo.

REM 构建 macOS ARM 版本
echo 正在构建 macOS ARM 版本...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_darwin_arm64" -ldflags="-s -w" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ macOS ARM 版本构建成功: %OUTPUT_DIR%\atc_darwin_arm64
) else (
    echo ✗ macOS ARM 版本构建失败
)
echo.

REM 构建 Linux ARM 版本
echo 正在构建 Linux ARM 版本...
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_linux_arm64" -ldflags="-s -w" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ Linux ARM 版本构建成功: %OUTPUT_DIR%\atc_linux_arm64
) else (
    echo ✗ Linux ARM 版本构建失败
)
echo.

REM 构建 Linux x86 版本
echo 正在构建 Linux x86 版本...
set GOOS=linux
set GOARCH=386
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_linux_386" -ldflags="-s -w" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ Linux x86 版本构建成功: %OUTPUT_DIR%\atc_linux_386
) else (
    echo ✗ Linux x86 版本构建失败
)
echo.

echo 构建完成！所有版本已保存到 %OUTPUT_DIR% 目录
echo Windows x86: %OUTPUT_DIR%\atc_windows_386.exe
echo macOS ARM: %OUTPUT_DIR%\atc_darwin_arm64
echo Linux ARM: %OUTPUT_DIR%\atc_linux_arm64
echo Linux x86: %OUTPUT_DIR%\atc_linux_386

REM 提示用户如何压缩文件
echo.
echo 注意：Windows 批处理脚本不包含自动压缩功能
echo 您可以手动压缩这些文件或使用 build.sh 脚本在支持 zip 命令的环境中构建

pause