/*
Copyright © 2025 API自动化测试命令行工具

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)



// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "atc",
	Short: "API自动化测试命令行工具",
	Long: `API自动化测试命令行工具(atc)是一个用于简化API测试流程的工具，支持以下功能：

1. 通过Dify Workflow API生成测试用例
2. 本地生成测试用例
3. 批量执行测试请求并保存结果

使用'atc [command] --help'获取更多信息。`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 这里定义全局标志和配置设置
	// Cobra支持持久性标志，如果在此处定义，将对应用程序全局有效

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-auto-test-cmd.yaml)")

	// 注意：子命令已在各自的init函数中添加到根命令
	// genCmd, localGenCmd, requestCmd已在各自的文件中通过init函数添加
}


