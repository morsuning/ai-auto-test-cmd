// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"

	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/spf13/cobra"
)

// genCmd 表示通过Dify Workflow API生成测试用例的命令
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "通过Dify Workflow API生成测试用例",
	Long: `通过Dify Workflow API生成测试用例，并保存为本地CSV文件。

示例：
  # 根据正例xml报文生成10条测试用例
  atc gen -u https://xxx.dify.com/xxx/xxx -xml -raw "xxxx" -n 10

  # 根据正例json报文生成20条测试用例
  atc gen -u https://xxx.dify.com/xxx/xxx -json -raw "xxxx" -n 20

  # 从XML文件读取正例报文生成测试用例
  atc gen -u https://xxx.dify.com/xxx/xxx -xml -f example.xml -n 15

  # 从JSON文件读取正例报文生成测试用例
  atc gen -u https://xxx.dify.com/xxx/xxx -json -f example.json -n 25

  # （可选）根据正例xml报文和接口文档生成5条正例报文
  atc gen -u https://xxx.dify.com/xxx/xxx -xml -raw "xxxx" -n 5 -d xxx.xlsx -p`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		url, _ := cmd.Flags().GetString("url")
		raw, _ := cmd.Flags().GetString("raw")
		file, _ := cmd.Flags().GetString("file")
		num, _ := cmd.Flags().GetInt("num")
		doc, _ := cmd.Flags().GetString("doc")
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		pretty, _ := cmd.Flags().GetBool("pretty")

		// 检查输入方式：必须指定raw或file其中之一
		if raw == "" && file == "" {
			fmt.Println("错误: 必须指定正例输入方式（-raw 或 -f）")
			return
		}
		if raw != "" && file != "" {
			fmt.Println("错误: 不能同时指定 -raw 和 -f 参数")
			return
		}

		// 如果指定了文件输入，读取文件内容
		var inputContent string
		if file != "" {
			content, err := utils.ReadFileContent(file)
			if err != nil {
				fmt.Printf("读取文件失败: %v\n", err)
				return
			}
			inputContent = content
			fmt.Printf("从文件读取正例: %s\n", file)
		} else {
			inputContent = raw
		}

		// 打印参数信息（仅用于框架演示）
		fmt.Println("通过Dify API生成测试用例")
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("报文格式: %s\n", getFormatName(isXML, isJSON))
		if file != "" {
			fmt.Printf("输入文件: %s\n", file)
		} else {
			fmt.Printf("原始报文: %s\n", inputContent)
		}
		fmt.Printf("生成数量: %d\n", num)
		if doc != "" {
			fmt.Printf("接口文档: %s\n", doc)
		}
		fmt.Printf("美化输出: %v\n", pretty)

		// TODO: 实现通过Dify API生成测试用例的功能
		// 这里应该使用 inputContent 作为正例内容
		fmt.Println("功能尚未实现")
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
	rootCmd.AddCommand(genCmd)

	// 定义命令行参数
	genCmd.Flags().StringP("url", "u", "", "目标URL（必需）")
	genCmd.Flags().StringP("raw", "r", "", "请求参数（正例报文）")
	genCmd.Flags().StringP("file", "f", "", "正例报文文件路径")
	genCmd.Flags().IntP("num", "n", 10, "生成用例数量（默认10）")
	genCmd.Flags().StringP("doc", "d", "", "接口文档路径（可选）")
	genCmd.Flags().BoolP("xml", "x", false, "使用XML格式")
	genCmd.Flags().BoolP("json", "j", false, "使用JSON格式")
	genCmd.Flags().BoolP("pretty", "p", false, "美化输出")

	// 标记必需的参数
	genCmd.MarkFlagRequired("url")
	// 注意：raw和file参数互斥，在Run函数中进行验证
}