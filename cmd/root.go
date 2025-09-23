// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 版本信息变量
var (
	appVersion   = "dev"
	appBuildTime = "unknown"
	appGitCommit = "unknown"
)

// SetVersionInfo 设置版本信息
func SetVersionInfo(version, buildTime, gitCommit string) {
	appVersion = version
	appBuildTime = buildTime
	appGitCommit = gitCommit
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "atc",
	Short: "API自动化测试命令行工具",
	Long: `API自动化测试命令行工具(atc)是一个用于简化API测试流程的工具，支持以下功能：

1. 通过LLM生成测试用例
2. 本地生成测试用例
3. 批量执行测试请求并保存结果

使用'atc [command] --help'获取更多信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否使用了--version标志
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Printf("API自动化测试命令行工具 (atc) %s\n", appVersion)
			if appVersion != "dev" {
				if appBuildTime != "unknown" {
					fmt.Printf("构建时间: %s\n", appBuildTime)
				}
				if appGitCommit != "unknown" {
					fmt.Printf("Git提交: %s\n", appGitCommit)
				}
			} else {
				fmt.Println("开发版本 - 使用构建脚本设置版本号")
			}
			return
		}
		// 如果没有指定子命令，显示帮助信息
		cmd.Help()
	},
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

	// 添加--version标志到根命令
	rootCmd.Flags().BoolP("version", "v", false, "显示版本信息")

	// 这里定义全局标志和配置设置
	// Cobra支持持久性标志，如果在此处定义，将对应用程序全局有效

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-auto-test-cmd.yaml)")
}
