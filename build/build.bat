@echo off
setlocal enabledelayedexpansion
REM 构建脚本：用于交叉编译 API 自动化测试命令行工具 (atc)
REM 支持的平台：Windows x86_64, macOS ARM64, Linux ARM64, Linux x86_64

REM 设置输出目录
set OUTPUT_DIR=.\bin

REM 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM 获取版本号
if "%~1"=="" (
    set /p VERSION="请输入版本号 (例如: v1.0.0): "
    if "!VERSION!"=="" (
        echo ❌ 版本号不能为空
        pause
        exit /b 1
    )
) else (
    set VERSION=%~1
)

REM 确保版本号以v开头
echo %VERSION% | findstr /r "^v" >nul
if errorlevel 1 (
    set VERSION=v%VERSION%
)

REM 获取构建时间 (Windows格式)
for /f "tokens=1-4 delims=/ " %%a in ('date /t') do set BUILD_DATE=%%d-%%b-%%c
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set BUILD_TIME_PART=%%a:%%b:00
set BUILD_TIME=%BUILD_DATE%T%BUILD_TIME_PART%Z

REM 尝试获取Git提交哈希
git rev-parse --short HEAD >nul 2>&1
if errorlevel 1 (
    set GIT_COMMIT=unknown
) else (
    for /f %%i in ('git rev-parse --short HEAD') do set GIT_COMMIT=%%i
)

REM 构建ldflags
set LDFLAGS=-s -w -X main.version=%VERSION% -X main.buildTime=%BUILD_TIME% -X main.gitCommit=%GIT_COMMIT%

echo 开始构建 API 自动化测试命令行工具 (atc) %VERSION%
echo 目标平台: Windows amd64, macOS arm64, Linux arm64, Linux amd64
echo 构建时间: %BUILD_TIME%
echo Git提交: %GIT_COMMIT%
echo.

REM 构建 Windows amd64 版本
echo 正在构建 Windows amd64 版本...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_windows_amd64_%VERSION%.exe" -ldflags="%LDFLAGS%" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ Windows amd64 版本构建成功: %OUTPUT_DIR%\atc_windows_amd64_%VERSION%.exe
) else (
    echo ✗ Windows amd64 版本构建失败
)
echo.

REM 构建 macOS ARM 版本
echo 正在构建 macOS ARM 版本...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_darwin_arm64_%VERSION%" -ldflags="%LDFLAGS%" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ macOS ARM 版本构建成功: %OUTPUT_DIR%\atc_darwin_arm64_%VERSION%
) else (
    echo ✗ macOS ARM 版本构建失败
)
echo.

REM 构建 Linux ARM 版本
echo 正在构建 Linux ARM 版本...
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_linux_arm64_%VERSION%" -ldflags="%LDFLAGS%" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ Linux ARM 版本构建成功: %OUTPUT_DIR%\atc_linux_arm64_%VERSION%
) else (
    echo ✗ Linux ARM 版本构建失败
)
echo.

REM 构建 Linux amd64 版本
echo 正在构建 Linux amd64 版本...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "%OUTPUT_DIR%\atc_linux_amd64_%VERSION%" -ldflags="%LDFLAGS%" .
if %ERRORLEVEL% EQU 0 (
    echo ✓ Linux amd64 版本构建成功: %OUTPUT_DIR%\atc_linux_amd64_%VERSION%
) else (
    echo ✗ Linux amd64 版本构建失败
)
echo.

echo 构建完成！所有版本已保存到 %OUTPUT_DIR% 目录
echo Windows amd64: %OUTPUT_DIR%\atc_windows_amd64_%VERSION%.exe
echo macOS ARM: %OUTPUT_DIR%\atc_darwin_arm64_%VERSION%
echo Linux ARM: %OUTPUT_DIR%\atc_linux_arm64_%VERSION%
echo Linux amd64: %OUTPUT_DIR%\atc_linux_amd64_%VERSION%

pause