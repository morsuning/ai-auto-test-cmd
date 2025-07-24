// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"

	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/spf13/cobra"
)

// difyGenCmd 表示通过Dify Chatflow API生成测试用例的命令
var difyGenCmd = &cobra.Command{
	Use:   "dify-gen",
	Short: "通过Dify Chatflow API生成测试用例",
	Long: `通过Dify Chatflow API生成测试用例，并保存为本地CSV文件。

示例：
  # 根据正例xml报文生成5条测试用例
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --xml --raw "xxxx" -n 5

	# 根据正例json报文生成10条测试用例
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --json --raw "xxxx" -n 10

	# 从XML文件读取正例报文生成测试用例
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --xml -f example.xml -n 6

	# 从JSON文件读取正例报文生成测试用例
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --json -f example.json -n 3

	# 使用默认配置文件(config.toml)中的URL和API Key
	atc dify-gen --xml --raw "xxxx" -n 5

	# 使用指定配置文件中的URL和API Key
	atc dify-gen -c my-config.toml --xml --raw "xxxx" -n 5

	# 使用自定义提示词文件生成测试用例
	atc dify-gen --xml --raw "xxxx" --prompt prompt.txt -n 3

	# 结合配置文件和提示词文件
	atc dify-gen -c my-config.toml --json --raw '{"test":"data"}' --prompt custom_prompt.txt -n 5`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		baseURL, _ := cmd.Flags().GetString("url")
		apiKey, _ := cmd.Flags().GetString("api-key")
		configFile, _ := cmd.Flags().GetString("config")
		raw, _ := cmd.Flags().GetString("raw")
		file, _ := cmd.Flags().GetString("file")
		promptFile, _ := cmd.Flags().GetString("prompt")
		num, _ := cmd.Flags().GetInt("num")
		output, _ := cmd.Flags().GetString("output")
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		debug, _ := cmd.Flags().GetBool("debug")
		exec, _ := cmd.Flags().GetBool("exec")

		// 如果使用exec参数，验证request相关参数
		var requestParams RequestParams
		if exec {
			var err error
			requestParams, err = getRequestParams(cmd)
			if err != nil {
				fmt.Printf("❌ 获取执行参数失败: %v\n", err)
				return
			}
			if err := validateRequestParams(requestParams); err != nil {
				fmt.Printf("❌ 执行参数验证失败: %v\n", err)
				return
			}
		}

		// 如果未显式指定URL或API Key，尝试从配置文件读取
		if baseURL == "" || apiKey == "" {
			// 确定配置文件路径
			if configFile == "" {
				configFile = "config.toml" // 默认配置文件
			}

			// 尝试加载配置文件
			config, err := utils.LoadConfig(configFile)
			if err != nil {
				// 如果配置文件加载失败且未显式指定URL和API Key，则报错
				if baseURL == "" && apiKey == "" {
					fmt.Printf("❌ 错误: 无法加载配置文件 %s: %v\n", configFile, err)
					fmt.Println("请通过 -u 和 --api-key 参数显式指定，或创建配置文件")
					return
				}
				// 如果只是部分参数缺失，给出提示但继续执行
				if debug {
					fmt.Printf("⚠️  警告: 配置文件加载失败: %v\n", err)
				}
			} else {
				// 从配置文件补充缺失的参数
				if baseURL == "" && config.Dify.URL != "" {
					baseURL = config.Dify.URL
					if debug {
						fmt.Printf("📄 从配置文件读取URL: %s\n", baseURL)
					}
				}
				if apiKey == "" && config.Dify.APIKey != "" {
					apiKey = config.Dify.APIKey
					if debug {
						fmt.Println("📄 从配置文件读取API Key")
					}
				}
			}
		}

		// 验证必需参数
		if baseURL == "" {
			fmt.Println("❌ 错误: 必须指定Dify API Base URL（通过 -u 参数或配置文件）")
			return
		}
		if apiKey == "" {
			fmt.Println("❌ 错误: 必须指定Dify API Key（通过 --api-key 参数或配置文件）")
			return
		}

		// 验证生成数量限制
		if num <= 0 {
			fmt.Println("❌ 错误: 生成数量必须大于0")
			return
		}

		// 检查输入方式：必须指定raw或file其中之一
		if raw == "" && file == "" {
			fmt.Println("❌ 错误: 必须指定正例输入方式（--raw 或 -f）")
			return
		}
		if raw != "" && file != "" {
			fmt.Println("❌ 错误: 不能同时指定 --raw 和 -f 参数")
			return
		}

		// 确定输入格式
		var format string
		if isXML && isJSON {
			fmt.Println("❌ 错误: 不能同时指定 --xml 和 --json 参数")
			return
		}
		if isXML {
			format = "xml"
		} else if isJSON {
			format = "json"
		} else {
			fmt.Println("❌ 错误: 必须指定报文格式（--xml 或 --json）")
			return
		}

		// 如果指定了文件输入，读取并验证文件内容
		var inputContent string
		if file != "" {
			content, err := utils.ReadAndValidateFileContent(file, format)
			if err != nil {
				fmt.Printf("❌ 文件读取或格式验证失败: %v\n", err)
				return
			}
			inputContent = content
			fmt.Printf("📁 从文件读取并验证正例: %s\n", file)
		} else {
			// 验证命令行输入的格式
			var err error
			if format == "xml" {
				err = utils.ValidateXMLFormat(raw)
			} else {
				err = utils.ValidateJSONFormat(raw)
			}
			if err != nil {
				fmt.Printf("❌ 输入格式验证失败: %v\n", err)
				return
			}
			inputContent = raw
		}

		// 设置默认输出文件
		if output == "" {
			if format == "xml" {
				output = "test_cases.csv"
			} else {
				output = "test_cases.csv"
			}
		}

		// 打印开始信息
		fmt.Println("🚀 通过Dify API生成测试用例")
		fmt.Printf("🌐 Base URL: %s\n", baseURL)
		fmt.Printf("📝 报文格式: %s\n", getFormatName(isXML, isJSON))
		fmt.Printf("🔢 生成数量: %d\n", num)
		fmt.Printf("💾 输出文件: %s\n", output)

		// 读取自定义提示词（如果指定）
		var userPrompt string
		if promptFile != "" {
			prompt, err := utils.ReadPromptFile(promptFile)
			if err != nil {
				fmt.Printf("❌ 读取提示词文件失败: %v\n", err)
				return
			}
			userPrompt = prompt
			if debug {
				fmt.Printf("📝 从文件读取自定义提示词: %s\n", promptFile)
				fmt.Printf("📄 提示词内容预览: %s...\n", truncateString(userPrompt, 100))
			}
		}

		// 准备请求参数
		inputs := map[string]any{
			"post_type": format, // 报文格式（json或xml）
			"test_num":  num,    // 生成的用例个数
		}

		// 如果有自定义提示词，添加到inputs中
		if userPrompt != "" {
			inputs["user_prompt"] = userPrompt
		}

		// 调用Dify API生成测试用例
		err := utils.GenerateTestCasesWithDify(apiKey, baseURL, inputContent, inputs, format, output, debug)
		if err != nil {
			fmt.Printf("❌ 生成测试用例失败: %v\n", err)
			return
		}

		fmt.Printf("✅ Dify调用已完成")

		// 如果使用exec参数，执行生成的测试用例
		if exec {
			if err := executeGeneratedTestCases(output, requestParams); err != nil {
				fmt.Printf("❌ 执行测试用例失败: %v\n", err)
				return
			}
		}
	},
}

// 获取报文格式名称
func getFormatName(isXML, isJSON bool) string {
	if isXML {
		return "XML"
	} else if isJSON {
		return "JSON"
	}
	return "未指定"
}

// truncateString 截断字符串到指定长度，用于调试输出
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func init() {
	rootCmd.AddCommand(difyGenCmd)

	// 必填参数组 - 报文格式（必须选择其一）
	difyGenCmd.Flags().BoolP("xml", "x", false, "使用XML格式")
	difyGenCmd.Flags().BoolP("json", "j", false, "使用JSON格式")

	// 必填参数组 - 输入方式（必须选择其一）
	difyGenCmd.Flags().StringP("raw", "r", "", "请求参数（正例报文）")
	difyGenCmd.Flags().StringP("file", "f", "", "正例报文文件路径")

	// API连接参数组
	difyGenCmd.Flags().StringP("url", "u", "", "Dify API Base URL（可选，可从配置文件读取）")
	difyGenCmd.Flags().String("api-key", "", "Dify API Key（可选，可从配置文件读取）")
	difyGenCmd.Flags().StringP("config", "c", "", "配置文件路径（默认为config.toml）")

	// 生成控制参数组
	difyGenCmd.Flags().IntP("num", "n", 5, "生成用例数量（默认5）")
	difyGenCmd.Flags().StringP("prompt", "p", "", "自定义提示词文件路径（可选，文件必须是UTF-8编码）")

	// 输出控制参数组
	difyGenCmd.Flags().StringP("output", "o", "", "输出文件路径（可选，默认为当前目录下的test_cases.csv）")

	// 执行控制参数组
	difyGenCmd.Flags().BoolP("exec", "e", false, "生成测试用例后立即执行")

	// 调试参数组
	difyGenCmd.Flags().BoolP("debug", "d", false, "启用调试模式")

	// 添加request相关参数（当使用exec时需要）
	addRequestFlags(difyGenCmd)

	// 自定义参数显示顺序
	difyGenCmd.Flags().SortFlags = false
	
	// 注意：url和api-key参数不再是必需的，可以从配置文件读取
	// raw和file参数互斥，在Run函数中进行验证
}
