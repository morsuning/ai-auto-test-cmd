package main

import (
	"fmt"

	"github.com/morsuning/ai-auto-test-cmd/cmd"
)

// 版本信息，通过 ldflags 在构建时注入
// 使用方法: go build -ldflags="-X main.version=v1.2.3" -o atc .
var (
	version   = "dev"     // 版本号，默认为开发版本
	buildTime = "unknown" // 构建时间
	gitCommit = "unknown" // Git提交哈希
)

func main() {
	// 显示欢迎信息
	fmt.Printf("API自动化测试命令行工具 (atc) %s\n", version)
	if version == "dev" {
		fmt.Println("开发版本 - 使用 'go build -ldflags=\"-X main.version=v1.0.0\"' 设置版本号")
	}
	fmt.Println("使用 'atc --help' 获取更多信息")
	fmt.Println()

	// 执行命令
	cmd.Execute()
}
