// Package cmd 提供命令行接口
package cmd

import (
	"fmt"
	"os"

	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

// validateCmd 验证约束配置文件命令
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "验证约束配置文件的格式和内容",
	Long: `验证约束配置文件的格式和内容。

该命令会检查约束配置文件中的：
- 约束类型是否有效
- 日期格式和范围是否正确
- 数值范围是否合理
- 内置数据是否完整
- 配置项是否符合规范

如果不指定配置文件，将验证默认的 constraints.toml 文件。`,
	Example: `  # 验证默认配置文件
  ai-auto-test-cmd validate

  # 验证指定配置文件
  ai-auto-test-cmd validate my-constraints.toml

  # 验证配置文件并显示详细信息
  ai-auto-test-cmd validate --verbose constraints.toml`,
	Args: cobra.MaximumNArgs(1),
	Run:  runValidate,
}

var (
	// verbose 是否显示详细验证信息
	verbose bool
)

// init 初始化验证命令
func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细验证信息")
}

// runValidate 执行验证命令
func runValidate(cmd *cobra.Command, args []string) {
	// 确定配置文件路径
	configFile := "constraints.toml"
	if len(args) > 0 {
		configFile = args[0]
	}

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 配置文件 '%s' 不存在\n", configFile)
		os.Exit(1)
	}

	fmt.Printf("🔍 正在验证约束配置文件: %s\n", configFile)

	// 加载并验证配置
	err := utils.LoadConstraintConfig(configFile)
	if err != nil {
		fmt.Printf("❌ 验证失败:\n%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 配置文件验证通过！\n")

	// 如果启用详细模式，显示配置统计信息
	if verbose {
		showConfigStats(configFile)
	}
}

// showConfigStats 显示配置统计信息
func showConfigStats(configFile string) {
	fmt.Println("\n📊 配置文件统计信息:")

	// 重新加载配置以获取统计信息
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	var rawConfig map[string]any
	if err := toml.Unmarshal(data, &rawConfig); err != nil {
		return
	}

	// 统计约束字段数量
	constraintCount := 0
	constraintTypes := make(map[string]int)

	for key, value := range rawConfig {
		if key != "builtin_data" {
			constraintCount++
			if valueMap, ok := value.(map[string]any); ok {
				if constraintType, exists := valueMap["type"]; exists {
					if typeStr, ok := constraintType.(string); ok {
						constraintTypes[typeStr]++
					}
				}
			}
		}
	}

	fmt.Printf("  • 约束字段总数: %d\n", constraintCount)
	fmt.Println("  • 约束类型分布:")
	for constraintType, count := range constraintTypes {
		fmt.Printf("    - %s: %d 个\n", constraintType, count)
	}

	// 统计内置数据
	if builtinData, exists := rawConfig["builtin_data"]; exists {
		if builtinMap, ok := builtinData.(map[string]any); ok {
			fmt.Println("  • 内置数据集:")
			if firstNames, exists := builtinMap["first_names"]; exists {
				if names, ok := firstNames.([]any); ok {
					fmt.Printf("    - 姓氏: %d 个\n", len(names))
				}
			}
			if lastNames, exists := builtinMap["last_names"]; exists {
				if names, ok := lastNames.([]any); ok {
					fmt.Printf("    - 名字: %d 个\n", len(names))
				}
			}
			if addresses, exists := builtinMap["addresses"]; exists {
				if addrs, ok := addresses.([]any); ok {
					fmt.Printf("    - 地址: %d 个\n", len(addrs))
				}
			}
			if emailDomains, exists := builtinMap["email_domains"]; exists {
				if domains, ok := emailDomains.([]any); ok {
					fmt.Printf("    - 邮箱域名: %d 个\n", len(domains))
				}
			}
		}
	}

	fmt.Println("\n💡 提示: 使用 'ai-auto-test-cmd local-gen --help' 查看如何使用约束功能")
}
