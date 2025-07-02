// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"

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
  atc local-gen -xml -raw "xxxx" -n 10

  # 本地根据正例json报文生成15条测试用例
  atc local-gen -json -raw "xxxx" -n 15`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		raw, _ := cmd.Flags().GetString("raw")
		num, _ := cmd.Flags().GetInt("num")
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		output, _ := cmd.Flags().GetString("output")

		// 设置默认输出文件
		if output == "" {
			output = "test_cases.csv"
		}

		// 打印参数信息
		fmt.Println("本地生成测试用例")
		fmt.Printf("报文格式: %s\n", getFormatName(isXML, isJSON))
		fmt.Printf("原始报文: %s\n", raw)
		fmt.Printf("生成数量: %d\n", num)
		fmt.Printf("输出文件: %s\n", output)

		// 检查报文格式
		if !isXML && !isJSON {
			fmt.Println("错误: 必须指定报文格式（XML或JSON）")
			return
		}

		// 解析报文并生成测试用例
		var data map[string]interface{}
		var err error

		if isXML {
			// 解析XML
			data, err = utils.ParseXML(raw)
			if err != nil {
				fmt.Printf("解析XML失败: %v\n", err)
				return
			}
		} else {
			// 解析JSON
			data, err = utils.ParseJSON(raw)
			if err != nil {
				fmt.Printf("解析JSON失败: %v\n", err)
				return
			}
		}

		// 生成测试用例
		fmt.Println("正在生成测试用例...")
		testCases := utils.GenerateTestCases(data, num)

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
	},
}

func init() {
	rootCmd.AddCommand(localGenCmd)

	// 定义命令行参数
	localGenCmd.Flags().StringP("raw", "r", "", "请求参数（正例报文）（必需）")
	localGenCmd.Flags().IntP("num", "n", 10, "生成用例数量（默认10）")
	localGenCmd.Flags().BoolP("xml", "x", false, "使用XML格式")
	localGenCmd.Flags().BoolP("json", "j", false, "使用JSON格式")
	localGenCmd.Flags().StringP("output", "o", "", "输出文件路径（默认为当前目录下的test_cases.csv）")

	// 标记必需的参数
	localGenCmd.MarkFlagRequired("raw")
}