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

示例：
  # 本地根据正例xml报文生成10条测试用例
  atc local-gen --xml --raw "xxxx" -n 10

  # 本地根据正例json报文生成15条测试用例
  atc local-gen --json --raw "xxxx" -n 15

  # 从XML文件读取正例报文生成测试用例
  atc local-gen --xml -f example.xml -n 20

  # 从JSON文件读取正例报文生成测试用例
  atc local-gen --json -f example.json -n 25

  # 使用默认约束配置生成智能测试用例
  atc local-gen --json -f example.json -n 10 --constraints

  # 使用自定义约束配置文件生成测试用例
  atc local-gen --json -f example.json -n 10 --constraints-file custom.toml`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		raw, _ := cmd.Flags().GetString("raw")
		file, _ := cmd.Flags().GetString("file")
		num, _ := cmd.Flags().GetInt("num")
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		output, _ := cmd.Flags().GetString("output")
		useConstraints, _ := cmd.Flags().GetBool("constraints")
		constraintsFile, _ := cmd.Flags().GetString("constraints-file")
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

		// 检查输入方式：必须指定raw或file其中之一
		if raw == "" && file == "" {
			fmt.Println("错误: 必须指定正例输入方式（--raw 或 -f）")
			return
		}
		if raw != "" && file != "" {
			fmt.Println("错误: 不能同时指定 --raw 和 -f 参数")
			return
		}

		// 确定输入格式
		var format string
		if isXML {
			format = "xml"
		} else if isJSON {
			format = "json"
		} else {
			fmt.Println("错误: 必须指定报文格式（--xml 或 --json）")
			return
		}

		// 如果指定了文件输入，读取并验证文件内容
		var inputContent string
		if file != "" {
			content, err := utils.ReadAndValidateFileContent(file, format)
			if err != nil {
				fmt.Printf("文件读取或格式验证失败: %v\n", err)
				return
			}
			inputContent = content
			fmt.Printf("从文件读取并验证正例: %s\n", file)
		} else {
			// 验证命令行输入的格式
			var err error
			if format == "xml" {
				err = utils.ValidateXMLFormat(raw)
			} else {
				err = utils.ValidateJSONFormat(raw)
			}
			if err != nil {
				fmt.Printf("输入格式验证失败: %v\n", err)
				return
			}
			inputContent = raw
		}

		// 设置默认输出文件
		if output == "" {
			output = "test_cases.csv"
		}

		// 打印参数信息
		fmt.Println("本地生成测试用例")
		fmt.Printf("报文格式: %s\n", getFormatName(isXML, isJSON))
		if file != "" {
			fmt.Printf("输入文件: %s\n", file)
		} else {
			fmt.Printf("原始报文: %s\n", inputContent)
		}
		fmt.Printf("生成数量: %d\n", num)
		fmt.Printf("输出文件: %s\n", output)

		// 加载约束配置
		if useConstraints || constraintsFile != "" {
			fmt.Println("启用约束模式")
			if constraintsFile != "" {
				fmt.Printf("加载约束配置文件: %s\n", constraintsFile)
				if err := utils.LoadConstraintConfig(constraintsFile); err != nil {
					fmt.Printf("加载约束配置失败: %v\n", err)
					return
				}
				fmt.Println("约束配置加载成功")
			} else {
				fmt.Println("加载默认约束配置")
				if err := utils.LoadDefaultConstraints(); err != nil {
					fmt.Printf("加载默认约束配置失败: %v\n", err)
					return
				}
				fmt.Println("默认约束配置加载成功")
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
		fmt.Println("正在生成测试用例...")
		var testCases []map[string]any
		if useConstraints || constraintsFile != "" {
			testCases = utils.GenerateTestCasesWithConstraints(data, num, true)
		} else {
			testCases = utils.GenerateTestCases(data, num)
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

		// 保存到文件
		err = utils.SaveToCSV(csvData, output)
		if err != nil {
			fmt.Printf("保存CSV文件失败: %v\n", err)
			return
		}
		fmt.Printf("成功生成 %d 条测试用例并保存到 %s\n", num, output)

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

	// 定义命令行参数
	localGenCmd.Flags().StringP("raw", "r", "", "请求参数（正例报文）")
	localGenCmd.Flags().StringP("file", "f", "", "正例报文文件路径")
	localGenCmd.Flags().IntP("num", "n", 10, "生成用例数量（默认10）")
	localGenCmd.Flags().BoolP("xml", "x", false, "使用XML格式")
	localGenCmd.Flags().BoolP("json", "j", false, "使用JSON格式")
	localGenCmd.Flags().StringP("output", "o", "", "输出文件路径（默认为当前目录下的test_cases.csv）")
	localGenCmd.Flags().BoolP("constraints", "c", false, "启用智能约束模式（使用默认配置）")
	localGenCmd.Flags().StringP("constraints-file", "C", "", "指定约束配置文件路径")

	// 添加exec参数
	localGenCmd.Flags().BoolP("exec", "e", false, "生成测试用例后立即执行")

	// 添加request相关参数
	addRequestFlags(localGenCmd)

	// 注意：raw和file参数互斥，在Run函数中进行验证
}
