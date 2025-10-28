// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/morsuning/ai-auto-test-cmd/models"
	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/spf13/cobra"
)

// localGenCmd 表示本地生成测试用例的命令
var localGenCmd = &cobra.Command{
	Use:   "local-gen",
	Short: "本地生成测试用例",
	Long: `本地生成测试用例，根据正向用例自动生成随机测试数据。

支持两种生成模式：
1. 随机变化模式：对原始数据进行随机变化（默认模式）
2. 智能约束模式：根据字段名应用相应的约束规则，生成更真实的测试数据

随机化因子配置：
- 可通过配置文件中的 testcase.variation_rate 设置随机化程度（0.0-1.0）
- 默认值为 0.5（50%变化程度）
- 值越大，生成的数据变化越大；值越小，生成的数据越接近原始数据

约束系统开关：
- 可通过配置文件中的 constraints.enable 控制
- 如果未明确设置，有约束配置时默认启用

示例：
  # 本地根据正例xml报文生成10条测试用例（随机变化模式）
  atc local-gen --xml "<root><name>test</name></root>" -n 10

  # 本地根据正例json报文生成15条测试用例（随机变化模式）
  atc local-gen --json '{"name":"test","age":25}' -n 15

  # 使用默认配置文件(config.toml)中的参数和正例报文
  atc local-gen

  # 命令行参数覆盖配置文件中的正例报文
  atc local-gen -c config.toml --json '{"name":"test"}'

  # 使用配置文件中的约束配置和自定义随机化因子生成智能测试用例
  atc local-gen -c config.toml -n 20

  # 生成测试用例并立即执行（从配置文件读取request参数）
  atc local-gen -c config.toml -e`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		xmlContent, _ := cmd.Flags().GetString("xml")
		jsonContent, _ := cmd.Flags().GetString("json")
		num, _ := cmd.Flags().GetInt("num")
		output, _ := cmd.Flags().GetString("output")
		configFile, _ := cmd.Flags().GetString("config")
		exec, _ := cmd.Flags().GetBool("exec")

		// 加载配置文件
		var config *utils.Config
		if configFile != "" {
			var err error
			config, err = utils.LoadConfigWithConstraints(configFile)
			if err != nil {
				fmt.Printf("❌ 加载配置文件失败: %v\n", err)
				return
			}
			fmt.Printf("📄 加载配置文件: %s\n", configFile)

			// 从配置文件补充缺失的参数
			if num == 10 && config.TestCase.Num != 0 { // 只有当num是默认值时才从配置文件读取
				num = config.TestCase.Num
			}
			if output == "" && config.TestCase.Output != "" {
				output = config.TestCase.Output
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
				fmt.Println("📄 从配置文件读取正例XML报文")
			} else if config.TestCase.Type == "json" && config.TestCase.PositiveExample != "" {
				isJSON = true
				inputContent = config.TestCase.PositiveExample
				// 验证JSON格式
				if err := utils.ValidateJSONFormat(inputContent); err != nil {
					fmt.Printf("❌ 配置文件中的JSON格式验证失败: %v\n", err)
					return
				}
				fmt.Println("📄 从配置文件读取正例JSON报文")
			} else {
				fmt.Println("❌ 错误: 必须指定报文内容（--xml 'content' 或 --json 'content'）或在配置文件中正确设置正例报文")
				fmt.Println("💡 提示: 在配置文件中设置 type=\"xml\" 和 positive_example，或设置 type=\"json\" 和 positive_example")
				return
			}
		}

		// 如果使用exec参数，从已加载的配置文件读取request相关参数
		var requestParams RequestParams
		if exec {
			if config == nil {
				fmt.Println("❌ 错误: 使用 --exec 参数时必须指定配置文件")
				return
			}

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

		// 设置默认输出文件
		if output == "" {
			output = "test_cases.csv"
		}

		// 打印参数信息
		fmt.Println("🔧 本地生成测试用例")
		fmt.Printf("📝 报文格式: %s\n", getFormatName(isXML, isJSON))
		fmt.Printf("📄 原始报文: %s\n", inputContent)
		fmt.Printf("🔢 生成数量: %d\n", num)
		fmt.Printf("💾 输出文件: %s\n", output)

		// 检查约束系统是否启用
		var useConstraints bool
		if config != nil {
			useConstraints = utils.IsConstraintsEnabled(config)

			if useConstraints {
				// 检查是否有约束配置
				if len(config.Constraints.Constraints) > 0 || len(config.Constraints.BuiltinData.FirstNames) > 0 || len(config.BuiltinData.FirstNames) > 0 {
					fmt.Println("✅ 约束系统已启用，智能约束模式生效")
				} else {
					fmt.Println("⚠️  约束系统已启用，但未找到约束配置，将使用随机变化模式")
					useConstraints = false
				}
			} else {
				fmt.Println("📋 约束系统已禁用，使用随机变化模式")
			}
		}

		// 解析报文并生成测试用例
		var data map[string]any
		var err error

		if isXML {
			// 解析XML
			data, err = utils.ParseXML(inputContent)
			if err != nil {
				fmt.Printf("解析XML失败: %v\n", err)
				return
			}
		} else {
			// 解析JSON
			data, err = utils.ParseJSON(inputContent)
			if err != nil {
				fmt.Printf("解析JSON失败: %v\n", err)
				return
			}
		}

		// 生成测试用例
		fmt.Println("🔄 正在生成测试用例...")
		var testCases []map[string]any

		// 获取随机化因子，优先使用配置文件中的设置，否则使用默认值0.5
		variationRate := 0.5
		if config != nil && config.TestCase.VariationRate > 0 {
			variationRate = config.TestCase.VariationRate
			fmt.Printf("🎲 使用配置的随机化因子: %.2f\n", variationRate)
		} else {
			fmt.Printf("🎲 使用默认随机化因子: %.2f\n", variationRate)
		}

		if useConstraints {
			testCases = utils.GenerateTestCasesWithVariationRate(data, num, variationRate, true)
		} else {
			testCases = utils.GenerateTestCasesWithVariationRate(data, num, variationRate, false)
		}

		// 根据格式转换数据
		var csvData [][]string
		if isXML {
			// XML格式：每行一个完整的XML
			csvData = utils.ConvertToXMLRows(testCases)
		} else {
			// JSON格式：每行一个完整的JSON
			csvData = utils.ConvertToJSONRows(testCases)
		}

		// 在命令行输出生成的测试用例
		if len(csvData) > 0 {
			for i, row := range csvData {
				if i == 0 {
					continue
				}
				fmt.Printf("🧪 测试用例 %d: %s\n", i, row[0])
			}
		}

		// 保存到文件
		err = utils.SaveToCSV(csvData, output)
		if err != nil {
			fmt.Printf("保存CSV文件失败: %v\n", err)
			return
		}
		fmt.Printf("✅ 成功生成 %d 条测试用例并保存到 %s\n", num, output)

		// 如果使用exec参数，执行生成的测试用例
		if exec {
			// 将生成的测试用例转换为models.TestCase格式
			modelTestCases := make([]models.TestCase, len(testCases))
			for i, testCase := range testCases {
				var testData map[string]any
				if isXML {
					// XML格式：将测试用例数据序列化为JSON字符串，然后存储为XML内容
					jsonBytes, _ := json.Marshal(testCase)
					testData = map[string]any{
						"_xml_content": string(jsonBytes), // 临时使用JSON字符串作为XML内容
					}
				} else {
					// JSON格式：使用特殊键存储JSON内容
					jsonBytes, _ := json.Marshal(testCase)
					testData = map[string]any{
						"_json_content": string(jsonBytes),
					}
				}

				modelTestCases[i] = models.TestCase{
					ID:          fmt.Sprintf("test_%d", i+1),
					Name:        fmt.Sprintf("测试用例_%d", i+1),
					Description: fmt.Sprintf("本地生成的第%d个测试用例", i+1),
					Type:        "auto",
					Data:        testData,
				}
			}

			// 直接执行测试用例
			if err := executeTestCasesDirectly(modelTestCases, requestParams); err != nil {
				fmt.Printf("❌ 执行测试用例失败: %v\n", err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(localGenCmd)

	// 必填参数组 - 报文格式和内容（必须选择其一）
	localGenCmd.Flags().StringP("xml", "x", "", "XML格式报文内容")
	localGenCmd.Flags().StringP("json", "j", "", "JSON格式报文内容")

	// 生成控制参数组
	localGenCmd.Flags().IntP("num", "n", 10, "生成用例数量（默认10）")

	// 配置文件参数组
	localGenCmd.Flags().StringP("config", "c", "config.toml", "配置文件路径（包含约束配置和其他设置）")

	// 输出控制参数组
	localGenCmd.Flags().StringP("output", "o", "", "输出文件路径（默认为当前目录下的test_cases.csv）")

	// 执行控制参数组
	localGenCmd.Flags().BoolP("exec", "e", false, "生成测试用例后立即执行")

	// 注意：使用-e参数时，request相关参数从配置文件读取

	// 自定义参数显示顺序
	localGenCmd.Flags().SortFlags = false

	// 注意：raw和file参数互斥，在Run函数中进行验证
}
