// Package cmd 提供命令行接口
package cmd

import (
	"fmt"
	"os"

	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/spf13/cobra"
)

// validateCmd 验证配置文件命令
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "验证配置文件的格式和内容",
	Long: `验证配置文件的格式和内容。

该命令会检查配置文件中的：
- LLM API配置是否正确
- 约束系统开关设置是否有效
- 约束类型是否有效
- 日期格式和范围是否正确
- 数值范围是否合理
- 内置数据是否完整
- 配置项是否符合规范

如果不指定配置文件，将验证默认的 config.toml 文件。`,
	Example: `  # 验证默认配置文件
  atc validate

  # 验证指定配置文件
  atc validate my-config.toml

  # 验证配置文件并显示详细信息
  atc validate --verbose config.toml`,
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
	configFile := "config.toml"
	if len(args) > 0 {
		configFile = args[0]
	}

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 配置文件 '%s' 不存在\n", configFile)
		os.Exit(1)
	}

	fmt.Printf("🔍 正在验证配置文件: %s\n", configFile)

	// 加载并验证配置
	config, err := utils.LoadConfigWithConstraints(configFile)
	if err != nil {
		fmt.Printf("❌ 验证失败:\n%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 配置文件验证通过！\n")

	// 如果启用详细模式，显示配置统计信息
	if verbose {
		showConfigStats(config)
	}
}

// showConfigStats 显示配置统计信息
func showConfigStats(config *utils.Config) {
	fmt.Println("\n📊 配置文件统计信息:")

	// 显示LLM配置信息
	fmt.Println("  • LLM API配置:")
	if config.LLM.URL != "" {
		fmt.Printf("    - URL: %s\n", config.LLM.URL)
	} else {
		fmt.Println("    - URL: 未配置")
	}
	if config.LLM.APIKey != "" {
		// 隐藏API Key的敏感信息
		maskedKey := config.LLM.APIKey
		if len(maskedKey) > 8 {
			maskedKey = maskedKey[:4] + "****" + maskedKey[len(maskedKey)-4:]
		}
		fmt.Printf("    - API Key: %s\n", maskedKey)
	} else {
		fmt.Println("    - API Key: 未配置")
	}
	if config.LLM.UserPrompt != "" {
		promptPreview := config.LLM.UserPrompt
		if len(promptPreview) > 50 {
			promptPreview = promptPreview[:50] + "..."
		}
		fmt.Printf("    - 自定义提示词: %s\n", promptPreview)
	} else {
		fmt.Println("    - 自定义提示词: 未配置")
	}

	// 显示约束系统配置
	fmt.Println("  • 约束系统配置:")
	constraintsEnabled := utils.IsConstraintsEnabled(config)
	if constraintsEnabled {
		fmt.Println("    - 状态: 已启用 ✅")
	} else {
		fmt.Println("    - 状态: 已禁用 ❌")
	}

	// 统计约束字段数量
	constraintCount := len(config.Constraints.Constraints)
	constraintTypes := make(map[string]int)

	for _, constraint := range config.Constraints.Constraints {
		constraintTypes[constraint.Type]++
	}

	fmt.Printf("    - 约束字段总数: %d\n", constraintCount)
	if constraintCount > 0 {
		fmt.Println("    - 约束类型分布:")
		for constraintType, count := range constraintTypes {
			fmt.Printf("      • %s: %d 个\n", constraintType, count)
		}
	}

	// 统计内置数据（优先使用constraints节点下的，向后兼容根节点下的）
	builtinData := config.Constraints.BuiltinData
	if len(builtinData.FirstNames) == 0 && len(config.BuiltinData.FirstNames) > 0 {
		builtinData = config.BuiltinData
	}

	if len(builtinData.FirstNames) > 0 || len(builtinData.LastNames) > 0 ||
		len(builtinData.Addresses) > 0 || len(builtinData.EmailDomains) > 0 ||
		len(builtinData.BankCards) > 0 || len(builtinData.PhoneNumbers) > 0 ||
		len(builtinData.IDCards) > 0 {
		fmt.Println("  • 内置数据集:")
		if len(builtinData.FirstNames) > 0 {
			fmt.Printf("    - 姓氏: %d 个\n", len(builtinData.FirstNames))
		}
		if len(builtinData.LastNames) > 0 {
			fmt.Printf("    - 名字: %d 个\n", len(builtinData.LastNames))
		}
		if len(builtinData.Addresses) > 0 {
			fmt.Printf("    - 地址: %d 个\n", len(builtinData.Addresses))
		}
		if len(builtinData.EmailDomains) > 0 {
			fmt.Printf("    - 邮箱域名: %d 个\n", len(builtinData.EmailDomains))
		}
		if len(builtinData.BankCards) > 0 {
			fmt.Printf("    - 银行卡号: %d 个\n", len(builtinData.BankCards))
		}
		if len(builtinData.PhoneNumbers) > 0 {
			fmt.Printf("    - 手机号: %d 个\n", len(builtinData.PhoneNumbers))
		}
		if len(builtinData.IDCards) > 0 {
			fmt.Printf("    - 身份证号: %d 个\n", len(builtinData.IDCards))
		}
	}

	fmt.Println("\n💡 提示:")
	fmt.Println("  - 使用 'atc llm-gen --help' 查看如何使用LLM测试用例生成功能")
	fmt.Println("  - 使用 'atc local-gen --help' 查看如何使用本地测试用例生成功能")
}
