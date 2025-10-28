// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"

	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/spf13/cobra"
)

// llmGenCmd 表示通过LLM API生成测试用例的命令
var llmGenCmd = &cobra.Command{
	Use:   "llm-gen",
	Short: "通过LLM生成测试用例",
	Long: `通过LLM API生成测试用例，并保存为本地CSV文件。

示例：
  # 根据正例xml报文生成5条测试用例
	atc llm-gen -u http://localhost/v1 --api-key app-xxx --xml "<root><name>test</name></root>" -n 5

	# 根据正例json报文生成10条测试用例
	atc llm-gen -u http://localhost/v1 --api-key app-xxx --json '{"name":"test","age":25}' -n 10

	# 使用默认配置文件(config.toml)中的参数和正例报文
	atc llm-gen

	# 使用指定配置文件中的参数和正例报文
	atc llm-gen -c my-config.toml

	# 命令行参数覆盖配置文件中的正例报文
	atc llm-gen -c config.toml --xml "<root><name>test</name></root>"

	# 使用自定义提示词文件生成测试用例
	atc llm-gen -c config.toml --prompt prompt.txt -n 3

	# 生成测试用例并立即执行（从配置文件读取request参数）
	atc llm-gen -c config.toml -e`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		baseURL, _ := cmd.Flags().GetString("url")
		apiKey, _ := cmd.Flags().GetString("api-key")
		configFile, _ := cmd.Flags().GetString("config")
		xmlContent, _ := cmd.Flags().GetString("xml")
		jsonContent, _ := cmd.Flags().GetString("json")
		promptFile, _ := cmd.Flags().GetString("prompt")
		num, _ := cmd.Flags().GetInt("num")
		output, _ := cmd.Flags().GetString("output")
		debug, _ := cmd.Flags().GetBool("debug")
		exec, _ := cmd.Flags().GetBool("exec")

		// 从配置文件读取参数（如果指定了配置文件或使用默认配置文件）
		var config *utils.Config
		if configFile != "" || baseURL == "" || apiKey == "" || num == 5 || output == "" {
			var err error
			config, err = utils.LoadConfig(configFile)
			if err != nil && (baseURL == "" || apiKey == "") {
				fmt.Printf("❌ 错误: 无法加载配置文件 %s: %v\n", configFile, err)
				fmt.Println("请通过 -u 和 --api-key 参数显式指定，或创建配置文件")
				return
			}

			if config != nil {
				// 从配置文件补充缺失的参数
				if baseURL == "" && config.LLM.URL != "" {
					baseURL = config.LLM.URL
					if debug {
						fmt.Printf("📄 从配置文件读取URL: %s\n", baseURL)
					}
				}
				if apiKey == "" && config.LLM.APIKey != "" {
					apiKey = config.LLM.APIKey
					if debug {
						fmt.Println("📄 从配置文件读取API Key")
					}
				}
				if num == 5 && config.TestCase.Num != 0 { // 只有当num是默认值时才从配置文件读取
					num = config.TestCase.Num
				}
				if output == "" && config.TestCase.Output != "" {
					output = config.TestCase.Output
				}
			}
		}

		// 确定输入格式和内容
		var isXML, isJSON bool
		var inputContent string

		if xmlContent != "" && jsonContent != "" {
			fmt.Println("❌ 错误: 不能同时指定 --xml 和 --json 参数")
			return
		}

		if xmlContent != "" {
			isXML = true
			inputContent = xmlContent
			// 验证XML格式
			if err := utils.ValidateXMLFormat(xmlContent); err != nil {
				fmt.Printf("❌ XML格式验证失败: %v\n", err)
				return
			}
		} else if jsonContent != "" {
			isJSON = true
			inputContent = jsonContent
			// 验证JSON格式
			if err := utils.ValidateJSONFormat(jsonContent); err != nil {
				fmt.Printf("❌ JSON格式验证失败: %v\n", err)
				return
			}
		} else {
			// 从配置文件读取正例报文
			if config == nil {
				fmt.Println("❌ 错误: 必须指定报文内容（--xml 'content' 或 --json 'content'）或在配置文件中设置正例报文")
				return
			}

			// 根据配置文件中的报文类型和内容确定格式
			if config.TestCase.Type == "xml" && config.TestCase.PositiveExample != "" {
				isXML = true
				inputContent = config.TestCase.PositiveExample
				// 验证XML格式
				if err := utils.ValidateXMLFormat(inputContent); err != nil {
					fmt.Printf("❌ 配置文件中的XML格式验证失败: %v\n", err)
					return
				}
				if debug {
					fmt.Println("📄 从配置文件读取正例XML报文")
				}
			} else if config.TestCase.Type == "json" && config.TestCase.PositiveExample != "" {
				isJSON = true
				inputContent = config.TestCase.PositiveExample
				// 验证JSON格式
				if err := utils.ValidateJSONFormat(inputContent); err != nil {
					fmt.Printf("❌ 配置文件中的JSON格式验证失败: %v\n", err)
					return
				}
				if debug {
					fmt.Println("📄 从配置文件读取正例JSON报文")
				}
			} else {
				fmt.Println("❌ 错误: 必须指定报文内容（--xml 'content' 或 --json 'content'）或在配置文件中正确设置正例报文")
				fmt.Println("💡 提示: 在配置文件中设置 type=\"xml\" 和 positive_example，或设置 type=\"json\" 和 positive_example")
				return
			}
		}

		// 如果使用exec参数，从配置文件读取request相关参数
		var requestParams RequestParams
		if exec {
			requestParams = RequestParams{
				URL:           config.Request.URL,
				Method:        config.Request.Method,
				SavePath:      config.Request.SavePath,
				Timeout:       config.Request.Timeout,
				Concurrent:    config.Request.Concurrent,
				AuthBearer:    config.Request.AuthBearer,
				AuthBasic:     config.Request.AuthBasic,
				AuthAPIKey:    config.Request.AuthAPIKey,
				CustomHeaders: config.Request.Headers,
				QueryParams:   config.Request.Query,
				IsXML:         isXML,
				IsJSON:        isJSON,
				IgnoreTLS:     config.Request.IgnoreTLSErrors,
			}

			// 设置默认值
			if requestParams.Method == "" {
				requestParams.Method = "post"
			}
			if requestParams.Timeout == 0 {
				requestParams.Timeout = 30
			}
			if requestParams.Concurrent == 0 {
				requestParams.Concurrent = 1
			}

			if err := validateRequestParams(requestParams); err != nil {
				fmt.Printf("❌ 配置文件中的request参数验证失败: %v\n", err)
				return
			}
		}

		// 验证必需参数
		if baseURL == "" {
			fmt.Println("❌ 错误: 必须指定LLM API Base URL（通过 -u 参数或配置文件）")
			return
		}
		if apiKey == "" {
			fmt.Println("❌ 错误: 必须指定LLM API Key（通过 --api-key 参数或配置文件）")
			return
		}

		// 验证生成数量限制
		if num <= 0 {
			fmt.Println("❌ 错误: 生成数量必须大于0")
			return
		}

		// 设置默认输出文件
		if output == "" {
			output = "test_cases.csv"
		}

		// 打印开始信息
		fmt.Println("🚀 通过LLM API生成测试用例")
		fmt.Printf("🌐 Base URL: %s\n", baseURL)
		fmt.Printf("📝 报文格式: %s\n", getFormatName(isXML, isJSON))
		fmt.Printf("🔢 生成数量: %d\n", num)
		fmt.Printf("💾 输出文件: %s\n", output)

		// 读取自定义提示词（优先级：命令行文件 > 配置文件 > 无）
		var userPrompt string
		if promptFile != "" {
			// 从命令行指定的文件读取提示词
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
		} else if config != nil && config.LLM.UserPrompt != "" {
			// 尝试从配置文件读取提示词
			userPrompt = config.LLM.UserPrompt
			if debug {
				fmt.Printf("📝 从配置文件读取自定义提示词: %s\n", configFile)
				fmt.Printf("📄 提示词内容预览: %s...\n", truncateString(userPrompt, 100))
			}
		}

		// 准备请求参数
		var format string
		if isXML {
			format = "xml"
		} else {
			format = "json"
		}
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

		fmt.Printf("✅ LLM调用已完成")

		// 如果使用exec参数，执行生成的测试用例
		if exec {
			if err := executeGeneratedTestCases(output, requestParams); err != nil {
				fmt.Printf("❌ 执行测试用例失败: %v\n", err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(llmGenCmd)

	// 必填参数组 - 报文格式和内容（必须选择其一）
	llmGenCmd.Flags().StringP("xml", "x", "", "XML格式报文内容")
	llmGenCmd.Flags().StringP("json", "j", "", "JSON格式报文内容")

	// API连接参数组
	llmGenCmd.Flags().StringP("url", "u", "", "LLM API Base URL")
	llmGenCmd.Flags().String("api-key", "", "LLM API Key")
	llmGenCmd.Flags().StringP("config", "c", "config.toml", "配置文件路径（默认为config.toml）")

	// 生成控制参数组
	llmGenCmd.Flags().IntP("num", "n", 5, "生成用例数量（默认5）")
	llmGenCmd.Flags().StringP("prompt", "p", "", "自定义提示词文件路径（文件必须是UTF-8编码）")

	// 输出控制参数组
	llmGenCmd.Flags().StringP("output", "o", "", "输出文件路径（默认为当前目录下的test_cases.csv）")

	// 执行控制参数组
	llmGenCmd.Flags().BoolP("exec", "e", false, "生成测试用例后立即执行")

	// 调试参数组
	llmGenCmd.Flags().BoolP("debug", "d", false, "启用调试模式")

	// 注意：使用-e参数时，request相关参数从配置文件读取

	// 自定义参数显示顺序
	llmGenCmd.Flags().SortFlags = false

	// 注意：url和api-key参数不再是必需的，可以从配置文件读取
	// raw和file参数互斥，在Run函数中进行验证
}
