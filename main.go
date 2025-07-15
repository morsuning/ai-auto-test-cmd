package main

import (
	"fmt"

	"github.com/morsuning/ai-auto-test-cmd/cmd"
)

func main() {
	// 显示欢迎信息
	fmt.Println("API自动化测试命令行工具 (atc) v1.2.1")
	fmt.Println("使用 'atc --help' 获取更多信息")
	fmt.Println()

	// 执行命令
	cmd.Execute()
}
