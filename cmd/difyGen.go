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

	# 根据正例json报文生成8条测试用例（最大限制）
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --json --raw "xxxx" -n 8

	# 从XML文件读取正例报文生成测试用例
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --xml -f example.xml -n 6

	# 从JSON文件读取正例报文生成测试用例
	atc dify-gen -u http://localhost/v1 --api-key app-xxx --json -f example.json -n 3`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		baseURL, _ := cmd.Flags().GetString("url")
		apiKey, _ := cmd.Flags().GetString("api-key")
		raw, _ := cmd.Flags().GetString("raw")
		file, _ := cmd.Flags().GetString("file")
		num, _ := cmd.Flags().GetInt("num")
		output, _ := cmd.Flags().GetString("output")
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		debug, _ := cmd.Flags().GetBool("debug")

		// 验证生成数量限制
		if num <= 0 {
			fmt.Println("❌ 错误: 生成数量必须大于0")
			return
		}
		if num > 8 {
			fmt.Println("❌ 错误: dify-gen命令最多支持一次生成8条测试用例")
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

		// 准备请求参数
		inputs := map[string]any{
			"post_type": format, // 报文格式（json或xml）
			"test_num":  num,    // 生成的用例个数
			"text_only": "yes",  // 仅文本输出，默认值为yes
		}

		// 调用Dify API生成测试用例
		err := utils.GenerateTestCasesWithDify(apiKey, baseURL, inputContent, inputs, format, output, debug)
		if err != nil {
			fmt.Printf("❌ 生成测试用例失败: %v\n", err)
			return
		}

		fmt.Printf("✅ Dify调用已完成")
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

func init() {
	rootCmd.AddCommand(difyGenCmd)

	// 定义命令行参数
	difyGenCmd.Flags().StringP("url", "u", "", "Dify API Base URL（必需）")
	difyGenCmd.Flags().String("api-key", "", "Dify API Key（必需）")
	difyGenCmd.Flags().StringP("raw", "r", "", "请求参数（正例报文）")
	difyGenCmd.Flags().StringP("file", "f", "", "正例报文文件路径")
	difyGenCmd.Flags().IntP("num", "n", 8, "生成用例数量（默认8，最大8）")
	difyGenCmd.Flags().StringP("output", "o", "", "输出文件路径（可选，默认为当前目录下的test_cases.csv）")
	difyGenCmd.Flags().BoolP("xml", "x", false, "使用XML格式")
	difyGenCmd.Flags().BoolP("json", "j", false, "使用JSON格式")
	difyGenCmd.Flags().BoolP("debug", "d", false, "启用调试模式")

	// 标记必需的参数
	difyGenCmd.MarkFlagRequired("url")
	difyGenCmd.MarkFlagRequired("api-key")
	// 注意：raw和file参数互斥，在Run函数中进行验证
}
